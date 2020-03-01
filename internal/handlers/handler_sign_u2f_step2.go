package handlers

import (
	"fmt"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/middlewares"
)

// SecondFactorU2FSignPost handler for completing a signing request.
func SecondFactorU2FSignPost(u2fVerifier U2FVerifier) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var requestBody signU2FRequestBody
		err := ctx.ParseBody(&requestBody)

		if err != nil {
			ctx.Error(err, mfaValidationFailedMessage)
			return
		}

		userSession := ctx.GetSession()
		if userSession.U2FChallenge == nil {
			ctx.Error(fmt.Errorf("U2F signing has not been initiated yet (no challenge)"), mfaValidationFailedMessage)
			return
		}

		if userSession.U2FRegistration == nil {
			ctx.Error(fmt.Errorf("U2F signing has not been initiated yet (no registration)"), mfaValidationFailedMessage)
			return
		}

		err = u2fVerifier.Verify(
			userSession.U2FRegistration.KeyHandle,
			userSession.U2FRegistration.PublicKey,
			requestBody.SignResponse,
			*userSession.U2FChallenge)

		if err != nil {
			ctx.Error(err, mfaValidationFailedMessage)
			return
		}

		err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx)

		if err != nil {
			ctx.Error(fmt.Errorf("Unable to regenerate session for user %s: %s", userSession.Username, err), authenticationFailedMessage)
			return
		}

		userSession.AuthenticationLevel = authentication.TwoFactor
		err = ctx.SaveSession(userSession)

		if err != nil {
			ctx.Error(fmt.Errorf("Unable to update authentication level with U2F: %s", err), mfaValidationFailedMessage)
			return
		}

		Handle2FAResponse(ctx, requestBody.TargetURL)
	}
}
