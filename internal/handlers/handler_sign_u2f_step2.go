package handlers

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// SecondFactorU2FSignPost handler for completing a signing request.
func SecondFactorU2FSignPost(u2fVerifier U2FVerifier) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var requestBody signU2FRequestBody
		err := ctx.ParseBody(&requestBody)

		if err != nil {
			ctx.Error(err, messageMFAValidationFailed)
			return
		}

		userSession := ctx.GetSession()
		if userSession.U2FChallenge == nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("U2F signing has not been initiated yet (no challenge)"), messageMFAValidationFailed)
			return
		}

		if userSession.U2FRegistration == nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("U2F signing has not been initiated yet (no registration)"), messageMFAValidationFailed)
			return
		}

		err = u2fVerifier.Verify(
			userSession.U2FRegistration.KeyHandle,
			userSession.U2FRegistration.PublicKey,
			requestBody.SignResponse,
			*userSession.U2FChallenge)

		if err != nil {
			ctx.Error(err, messageMFAValidationFailed)
			return
		}

		err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx)

		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to regenerate session for user %s: %s", userSession.Username, err), messageMFAValidationFailed)
			return
		}

		userSession.SetTwoFactor(ctx.Clock.Now())

		err = ctx.SaveSession(userSession)
		if err != nil {
			handleAuthenticationUnauthorized(ctx, fmt.Errorf("unable to update authentication level with U2F: %s", err), messageMFAValidationFailed)
			return
		}

		if userSession.OIDCWorkflowSession != nil {
			handleOIDCWorkflowResponse(ctx)
		} else {
			Handle2FAResponse(ctx, requestBody.TargetURL)
		}
	}
}
