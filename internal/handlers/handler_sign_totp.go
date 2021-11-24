package handlers

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
)

// SecondFactorTOTPPost validate the TOTP passcode provided by the user.
func SecondFactorTOTPPost(totpVerifier TOTPVerifier) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		requestBody := signTOTPRequestBody{}

		if err := ctx.ParseBody(&requestBody); err != nil {
			handleAuthenticationUnauthorized(ctx, err, messageMFAValidationFailed)
			return
		}

		userSession := ctx.GetSession()

		config, err := ctx.Providers.StorageProvider.LoadTOTPConfiguration(ctx, userSession.Username)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to load TOTP secret: %s", err), messageMFAValidationFailed)
			return
		}

		isValid, err := totpVerifier.Verify(config, requestBody.Token)
		if err != nil {
			_ = handleAuthenticationAttempt(ctx, err, false, nil, userSession.Username, regulation.AuthTypeTOTP)
			return
		}

		if !isValid {
			_ = handleAuthenticationAttempt(ctx, nil, false, nil, userSession.Username, regulation.AuthTypeTOTP)
			return
		}

		if err = handleAuthenticationAttempt(ctx, nil, true, nil, userSession.Username, regulation.AuthTypeTOTP); err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to mark authentication: %w", err), messageMFAValidationFailed)
			return
		}

		if err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx); err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to regenerate session for user %s: %s", userSession.Username, err), messageMFAValidationFailed)
			return
		}

		userSession.SetTwoFactor(ctx.Clock.Now())

		if err = ctx.SaveSession(userSession); err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to update the authentication level with TOTP: %s", err), messageMFAValidationFailed)
			return
		}

		if userSession.OIDCWorkflowSession != nil {
			handleOIDCWorkflowResponse(ctx)
		} else {
			Handle2FAResponse(ctx, requestBody.TargetURL)
		}
	}
}
