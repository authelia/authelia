package handlers

import (
	"fmt"

	"github.com/authelia/authelia/internal/middlewares"
)

// SecondFactorTOTPPost validate the TOTP passcode provided by the user.
func SecondFactorTOTPPost(totpVerifier TOTPVerifier) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		requestBody := signTOTPRequestBody{}
		err := ctx.ParseBody(&requestBody)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, err, messageMFAValidationFailed)
			return
		}

		userSession := ctx.GetSession()

		secret, err := ctx.Providers.StorageProvider.LoadTOTPSecret(userSession.Username)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Unable to load TOTP secret: %s", err), messageMFAValidationFailed)
			return
		}

		isValid, err := totpVerifier.Verify(requestBody.Token, secret)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Error occurred during OTP validation for user %s: %s", userSession.Username, err), messageMFAValidationFailed)
			return
		}

		if !isValid {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Wrong passcode during TOTP validation for user %s", userSession.Username), messageMFAValidationFailed)
			return
		}

		err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Unable to regenerate session for user %s: %s", userSession.Username, err), messageMFAValidationFailed)
			return
		}

		userSession.SetTwoFactor(ctx.Clock.Now())

		err = ctx.SaveSession(userSession)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("Unable to update the authentication level with TOTP: %s", err), messageMFAValidationFailed)
			return
		}

		if userSession.OIDCWorkflowSession != nil {
			handleOIDCWorkflowResponse(ctx)
		} else {
			Handle2FAResponse(ctx, requestBody.TargetURL)
		}
	}
}
