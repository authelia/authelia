package handlers

import (
	"fmt"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/middlewares"
)

// SecondFactorTOTPPost validate the TOTP passcode provided by the user.
func SecondFactorTOTPPost(totpVerifier TOTPVerifier) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		bodyJSON := signTOTPRequestBody{}
		err := ctx.ParseBody(&bodyJSON)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, err, mfaValidationFailedMessage)
			return
		}

		userSession := ctx.GetSession()

		secret, algorithm, err := ctx.Providers.StorageProvider.LoadTOTPSecret(userSession.Username)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Unable to load TOTP secret: %s", err), mfaValidationFailedMessage)
			return
		}

		isValid, err := totpVerifier.Verify(bodyJSON.Token, secret, algorithm)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Error occurred during OTP validation for user %s: %s", userSession.Username, err), mfaValidationFailedMessage)
			return
		}

		if !isValid {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Wrong passcode during TOTP validation for user %s", userSession.Username), mfaValidationFailedMessage)
			return
		}

		err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Unable to regenerate session for user %s: %s", userSession.Username, err), mfaValidationFailedMessage)
			return
		}

		userSession.AuthenticationLevel = authentication.TwoFactor
		err = ctx.SaveSession(userSession)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Unable to update the authentication level with TOTP: %s", err), mfaValidationFailedMessage)
			return
		}

		Handle2FAResponse(ctx, bodyJSON.TargetURL)
	}
}
