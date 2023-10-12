package middlewares

import (
	"encoding/json"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/templates"
)

// IdentityVerificationStart the handler for initiating the identity validation process.
func IdentityVerificationStart(args IdentityVerificationStartArgs, delayFunc TimingAttackDelayFunc) RequestHandler {
	if args.IdentityRetrieverFunc == nil {
		panic(fmt.Errorf("Identity verification requires an identity retriever"))
	}

	return func(ctx *AutheliaCtx) {
		requestTime := time.Now()
		success := false

		if delayFunc != nil {
			defer delayFunc(ctx, requestTime, &success)
		}

		identity, err := args.IdentityRetrieverFunc(ctx)
		if err != nil {
			// In that case we reply ok to avoid user enumeration.
			ctx.Logger.Error(err)
			ctx.ReplyOK()

			return
		}

		var jti uuid.UUID

		if jti, err = uuid.NewRandom(); err != nil {
			ctx.Error(err, messageOperationFailed)
			return
		}

		verification := model.NewIdentityVerification(jti, identity.Username, args.ActionClaim, ctx.RemoteIP(), ctx.Configuration.IdentityValidation.ResetPassword.Expiration)

		// Create the claim with the action to sign it.
		claims := verification.ToIdentityVerificationClaim()

		var method *jwt.SigningMethodHMAC

		switch ctx.Configuration.IdentityValidation.ResetPassword.JWTAlgorithm {
		case "HS256":
			method = jwt.SigningMethodHS256
		case "HS384":
			method = jwt.SigningMethodHS384
		case "HS512":
			method = jwt.SigningMethodHS512
		default:
			method = jwt.SigningMethodHS256
		}

		token := jwt.NewWithClaims(method, claims)

		signedToken, err := token.SignedString([]byte(ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret))
		if err != nil {
			ctx.Error(err, messageOperationFailed)
			return
		}

		if err = ctx.Providers.StorageProvider.SaveIdentityVerification(ctx, verification); err != nil {
			ctx.Error(err, messageOperationFailed)
			return
		}

		linkURL := ctx.RootURL()

		query := linkURL.Query()

		query.Set(queryArgToken, signedToken)

		linkURL.Path = path.Join(linkURL.Path, args.TargetEndpoint)
		linkURL.RawQuery = query.Encode()

		data := templates.EmailIdentityVerificationJWTValues{
			Title:       args.MailTitle,
			LinkURL:     linkURL.String(),
			LinkText:    args.MailButtonContent,
			DisplayName: identity.DisplayName,
			RemoteIP:    ctx.RemoteIP().String(),
		}

		ctx.Logger.Debugf("Sending an email to user %s (%s) to confirm identity for registering a device.",
			identity.Username, identity.Email)

		if err = ctx.Providers.Notifier.Send(ctx, identity.Address(), args.MailTitle, ctx.Providers.Templates.GetIdentityVerificationJWTEmailTemplate(), data); err != nil {
			ctx.Error(err, messageOperationFailed)
			return
		}

		success = true

		ctx.ReplyOK()
	}
}

// IdentityVerificationFinish the middleware for finishing the identity validation process.
//
//nolint:gocyclo
func IdentityVerificationFinish(args IdentityVerificationFinishArgs, next func(ctx *AutheliaCtx, username string)) RequestHandler {
	return func(ctx *AutheliaCtx) {
		var finishBody IdentityVerificationFinishBody

		b := ctx.PostBody()

		err := json.Unmarshal(b, &finishBody)

		if err != nil {
			ctx.Error(err, messageOperationFailed)
			return
		}

		if finishBody.Token == "" {
			ctx.Error(fmt.Errorf("No token provided"), messageOperationFailed)
			return
		}

		token, err := jwt.ParseWithClaims(finishBody.Token, &model.IdentityVerificationClaim{},
			func(token *jwt.Token) (any, error) {
				return []byte(ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret), nil
			},
			jwt.WithIssuedAt(),
			jwt.WithIssuer("Authelia"),
			jwt.WithStrictDecoding(),
			ctx.GetJWTWithTimeFuncOption(),
		)

		switch {
		case err == nil:
			break
		case errors.Is(err, jwt.ErrTokenMalformed):
			ctx.Error(fmt.Errorf("Cannot parse token"), messageOperationFailed)
			return
		case errors.Is(err, jwt.ErrTokenExpired):
			ctx.Error(fmt.Errorf("Token expired"), messageIdentityVerificationTokenHasExpired)
			return
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			ctx.Error(fmt.Errorf("Token is only valid in the future"), messageIdentityVerificationTokenNotValidYet)
			return
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			ctx.Error(fmt.Errorf("Token signature does't match"), messageIdentityVerificationTokenSig)
			return
		default:
			ctx.Error(fmt.Errorf("Cannot handle this token: %w", err), messageOperationFailed)
			return
		}

		claims, ok := token.Claims.(*model.IdentityVerificationClaim)
		if !ok {
			ctx.Error(fmt.Errorf("Wrong type of claims (%T != *middlewares.IdentityVerificationClaim)", claims), messageOperationFailed)
			return
		}

		verification, err := claims.ToIdentityVerification()
		if err != nil {
			ctx.Error(fmt.Errorf("Token seems to be invalid: %w", err),
				messageOperationFailed)
			return
		}

		found, err := ctx.Providers.StorageProvider.FindIdentityVerification(ctx, verification.JTI.String())
		if err != nil {
			ctx.Error(err, messageOperationFailed)
			return
		}

		if !found {
			ctx.Error(fmt.Errorf("Token is not in DB, it might have already been used"),
				messageIdentityVerificationTokenAlreadyUsed)
			return
		}

		// Verify that the action claim in the token is the one expected for the given endpoint.
		if claims.Action != args.ActionClaim {
			ctx.Error(fmt.Errorf("This token has not been generated for this kind of action"), messageOperationFailed)
			return
		}

		if args.IsTokenUserValidFunc != nil && !args.IsTokenUserValidFunc(ctx, claims.Username) {
			ctx.Error(fmt.Errorf("This token has not been generated for this user"), messageOperationFailed)
			return
		}

		if err = ctx.Providers.StorageProvider.ConsumeIdentityVerification(ctx, claims.ID, model.NewNullIP(ctx.RemoteIP())); err != nil {
			ctx.Error(err, messageOperationFailed)
			return
		}

		next(ctx, claims.Username)
	}
}
