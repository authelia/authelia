package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	jwt "github.com/dgrijalva/jwt-go"

	"github.com/authelia/authelia/internal/templates"
)

// IdentityVerificationStart the handler for initiating the identity validation process.
func IdentityVerificationStart(args IdentityVerificationStartArgs) RequestHandler {
	if args.IdentityRetrieverFunc == nil {
		panic(fmt.Errorf("Identity verification requires an identity retriever"))
	}

	return func(ctx *AutheliaCtx) {
		identity, err := args.IdentityRetrieverFunc(ctx)

		if err != nil {
			// In that case we reply ok to avoid user enumeration.
			ctx.Logger.Error(err)
			ctx.ReplyOK()

			return
		}

		// Create the claim with the action to sign it.
		claims := &IdentityVerificationClaim{
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
				Issuer:    jwtIssuer,
			},
			args.ActionClaim,
			identity.Username,
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		ss, err := token.SignedString([]byte(ctx.Configuration.JWTSecret))

		if err != nil {
			ctx.Error(err, operationFailedMessage)
			return
		}

		err = ctx.Providers.StorageProvider.SaveIdentityVerificationToken(ss)
		if err != nil {
			ctx.Error(err, operationFailedMessage)
			return
		}

		if ctx.XForwardedProto() == nil {
			ctx.Error(errMissingXForwardedProto, operationFailedMessage)
			return
		}

		if ctx.XForwardedHost() == nil {
			ctx.Error(errMissingXForwardedHost, operationFailedMessage)
			return
		}

		link := fmt.Sprintf("%s://%s%s%s?token=%s", ctx.XForwardedProto(),
			ctx.XForwardedHost(), ctx.Configuration.Server.Path, args.TargetEndpoint, ss)

		bufHTML := new(bytes.Buffer)

		disableHTML := false
		if ctx.Configuration.Notifier != nil && ctx.Configuration.Notifier.SMTP != nil {
			disableHTML = ctx.Configuration.Notifier.SMTP.DisableHTMLEmails
		}

		if !disableHTML {
			htmlParams := map[string]interface{}{
				"title":  args.MailTitle,
				"url":    link,
				"button": args.MailButtonContent,
			}

			err = templates.HTMLEmailTemplate.Execute(bufHTML, htmlParams)

			if err != nil {
				ctx.Error(err, operationFailedMessage)
				return
			}
		}

		bufText := new(bytes.Buffer)
		textParams := map[string]interface{}{
			"url": link,
		}

		err = templates.PlainTextEmailTemplate.Execute(bufText, textParams)

		if err != nil {
			ctx.Error(err, operationFailedMessage)
			return
		}

		ctx.Logger.Debugf("Sending an email to user %s (%s) to confirm identity for registering a device.",
			identity.Username, identity.Email)

		err = ctx.Providers.Notifier.Send(identity.Email, args.MailTitle, bufText.String(), bufHTML.String())

		if err != nil {
			ctx.Error(err, operationFailedMessage)
			return
		}

		ctx.ReplyOK()
	}
}

// IdentityVerificationFinish the middleware for finishing the identity validation process.
func IdentityVerificationFinish(args IdentityVerificationFinishArgs, next func(ctx *AutheliaCtx, username string)) RequestHandler {
	return func(ctx *AutheliaCtx) {
		var finishBody IdentityVerificationFinishBody

		b := ctx.PostBody()

		err := json.Unmarshal(b, &finishBody)

		if err != nil {
			ctx.Error(err, operationFailedMessage)
			return
		}

		if finishBody.Token == "" {
			ctx.Error(fmt.Errorf("No token provided"), operationFailedMessage)
			return
		}

		found, err := ctx.Providers.StorageProvider.FindIdentityVerificationToken(finishBody.Token)

		if err != nil {
			ctx.Error(err, operationFailedMessage)
			return
		}

		if !found {
			ctx.Error(fmt.Errorf("Token is not in DB, it might have already been used"),
				identityVerificationTokenAlreadyUsedMessage)
			return
		}

		token, err := jwt.ParseWithClaims(finishBody.Token, &IdentityVerificationClaim{},
			func(token *jwt.Token) (interface{}, error) {
				return []byte(ctx.Configuration.JWTSecret), nil
			})

		if err != nil {
			if ve, ok := err.(*jwt.ValidationError); ok {
				switch {
				case ve.Errors&jwt.ValidationErrorMalformed != 0:
					ctx.Error(fmt.Errorf("Cannot parse token"), operationFailedMessage)
					return
				case ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0:
					// Token is either expired or not active yet
					ctx.Error(fmt.Errorf("Token expired"), identityVerificationTokenHasExpiredMessage)
					return
				default:
					ctx.Error(fmt.Errorf("Cannot handle this token: %s", ve), operationFailedMessage)
					return
				}
			}

			ctx.Error(err, operationFailedMessage)

			return
		}

		claims, ok := token.Claims.(*IdentityVerificationClaim)
		if !ok {
			ctx.Error(fmt.Errorf("Wrong type of claims (%T != *middlewares.IdentityVerificationClaim)", claims), operationFailedMessage)
			return
		}

		// Verify that the action claim in the token is the one expected for the given endpoint.
		if claims.Action != args.ActionClaim {
			ctx.Error(fmt.Errorf("This token has not been generated for this kind of action"), operationFailedMessage)
			return
		}

		if args.IsTokenUserValidFunc != nil && !args.IsTokenUserValidFunc(ctx, claims.Username) {
			ctx.Error(fmt.Errorf("This token has not been generated for this user"), operationFailedMessage)
			return
		}

		// TODO(c.michaud): find a way to garbage collect unused tokens.
		err = ctx.Providers.StorageProvider.RemoveIdentityVerificationToken(finishBody.Token)
		if err != nil {
			ctx.Error(err, operationFailedMessage)
			return
		}

		next(ctx, claims.Username)
	}
}
