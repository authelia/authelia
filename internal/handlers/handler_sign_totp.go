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
			ctx.Error(err, mfaValidationFailedMessage)
			return
		}

		userSession := ctx.GetSession()
		secret, err := ctx.Providers.StorageProvider.LoadTOTPSecret(userSession.Username)
		if err != nil {
			ctx.Error(fmt.Errorf("Unable to load TOTP secret: %s", err), mfaValidationFailedMessage)
			return
		}

		isValid := totpVerifier.Verify(bodyJSON.Token, secret)

		if !isValid {
			ctx.Error(fmt.Errorf("Wrong passcode during TOTP validation for user %s", userSession.Username), mfaValidationFailedMessage)
			return
		}

		userSession.AuthenticationLevel = authentication.TwoFactor
		err = ctx.SaveSession(userSession)

		if err != nil {
			ctx.Error(fmt.Errorf("Unable to update the authentication level with TOTP: %s", err), mfaValidationFailedMessage)
			return
		}

		HandleAuthResponse(ctx, bodyJSON.TargetURL)
	}
}
