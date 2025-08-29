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
		panic(fmt.Errorf("identity verification requires an identity retriever"))
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

		verification := model.NewIdentityVerification(jti, identity.Username, args.ActionClaim, ctx.RemoteIP(), ctx.Configuration.IdentityValidation.ResetPassword.JWTExpiration)

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

		revocationLinkURL := ctx.RootURL()

		query = revocationLinkURL.Query()

		query.Set(queryArgToken, signedToken)

		revocationLinkURL.Path = path.Join(revocationLinkURL.Path, args.RevokeEndpoint)
		revocationLinkURL.RawQuery = query.Encode()

		domain, _ := ctx.GetCookieDomain()

		data := templates.EmailIdentityVerificationJWTValues{
			Title:              args.MailTitle,
			LinkURL:            linkURL.String(),
			LinkText:           args.MailButtonContent,
			RevocationLinkURL:  revocationLinkURL.String(),
			RevocationLinkText: args.MailButtonRevokeContent,
			DisplayName:        identity.DisplayName,
			Domain:             domain,
			RemoteIP:           ctx.RemoteIP().String(),
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
			ctx.Error(fmt.Errorf("no token provided"), messageOperationFailed)
			return
		}

		token, err := jwt.ParseWithClaims(finishBody.Token, &model.IdentityVerificationClaim{},
			func(token *jwt.Token) (any, error) {
				return []byte(ctx.Configuration.IdentityValidation.ResetPassword.JWTSecret), nil
			},
			jwt.WithIssuedAt(),
			jwt.WithIssuer("Authelia"),
			jwt.WithStrictDecoding(),
			ctx.GetClock().GetJWTWithTimeFuncOption(),
		)

		switch {
		case err == nil:
			break
		case errors.Is(err, jwt.ErrTokenMalformed):
			ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token as it appears to be malformed, this potentially can occur if you've not copied the full link")
			ctx.SetJSONError(messageOperationFailed)

			return
		case errors.Is(err, jwt.ErrTokenExpired):
			ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token validity period as it appears to be expired")
			ctx.SetJSONError(messageIdentityVerificationTokenHasExpired)

			return
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token validity period as it appears to only be valid in the future")
			ctx.SetJSONError(messageIdentityVerificationTokenNotValidYet)

			return
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token signature")
			ctx.SetJSONError(messageIdentityVerificationTokenSig)

			return
		default:
			ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token")
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		claims, ok := token.Claims.(*model.IdentityVerificationClaim)
		if !ok {
			ctx.Logger.WithError(fmt.Errorf("failed to map the %T claims to a *model.IdentityVerificationClaim", claims)).Error("Error occurred validating the identity verification token claims")
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		verification, err := claims.ToIdentityVerification()
		if err != nil {
			ctx.Logger.WithError(err).Error("Error occurred validating the identity verification token claims as they appear to be malformed")
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		found, err := ctx.Providers.StorageProvider.FindIdentityVerification(ctx, verification.JTI.String())
		if err != nil {
			ctx.Logger.WithError(err).Error("Error occurred looking up identity verification during the validation phase")
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		if !found {
			ctx.Logger.Error("Error occurred looking up identity verification during the validation phase, the token was not found in the database which could indicate it was never generated or was already used")
			ctx.SetJSONError(messageIdentityVerificationTokenAlreadyUsed)

			return
		}

		// Verify that the action claim in the token is the one expected for the given endpoint.
		if claims.Action != args.ActionClaim {
			ctx.Logger.Errorf("Error occurred handling the identity verification token, the token action '%s' does not match the endpoint action '%s' which is not allowed", claims.Action, args.ActionClaim)
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		if args.IsTokenUserValidFunc != nil && !args.IsTokenUserValidFunc(ctx, claims.Username) {
			ctx.Logger.Errorf("Error occurred handling the identity verification token, the user is not allowed to use this token")
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		if err = ctx.Providers.StorageProvider.ConsumeIdentityVerification(ctx, claims.ID, model.NewNullIP(ctx.RemoteIP())); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred consuming the identity verification during the validation phase")
			ctx.SetJSONError(messageOperationFailed)

			return
		}

		next(ctx, claims.Username)
	}
}
