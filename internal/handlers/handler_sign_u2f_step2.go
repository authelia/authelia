package handlers

import (
	"errors"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
)

// SecondFactorU2FSignPost handler for completing a signing request.
func SecondFactorU2FSignPost(u2fVerifier U2FVerifier) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var (
			requestBody signU2FRequestBody
			err         error
		)

		if err := ctx.ParseBody(&requestBody); err != nil {
			ctx.Logger.Errorf(logFmtErrParseRequestBody, regulation.AuthTypeU2F, err)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		userSession := ctx.GetSession()
		if userSession.U2FChallenge == nil {
			_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypeU2F, errors.New("session did not contain a challenge"))

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		if userSession.U2FRegistration == nil {
			_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypeU2F, errors.New("session did not contain a registration"))

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		if err = u2fVerifier.Verify(userSession.U2FRegistration.KeyHandle, userSession.U2FRegistration.PublicKey,
			requestBody.SignResponse, *userSession.U2FChallenge); err != nil {
			_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypeU2F, err)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		if err = ctx.Providers.SessionProvider.RegenerateSession(ctx.RequestCtx); err != nil {
			ctx.Logger.Errorf(logFmtErrSessionRegenerate, regulation.AuthTypeU2F, userSession.Username, err)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		if err = markAuthenticationAttempt(ctx, true, nil, userSession.Username, regulation.AuthTypeU2F, nil); err != nil {
			respondUnauthorized(ctx, messageMFAValidationFailed)
			return
		}

		userSession.SetTwoFactor(ctx.Clock.Now())

		err = ctx.SaveSession(userSession)
		if err != nil {
			ctx.Logger.Errorf(logFmtErrSessionSave, "authentication time", regulation.AuthTypeU2F, userSession.Username, err)

			respondUnauthorized(ctx, messageMFAValidationFailed)

			return
		}

		if userSession.OIDCWorkflowSession != nil {
			handleOIDCWorkflowResponse(ctx)
		} else {
			Handle2FAResponse(ctx, requestBody.TargetURL)
		}
	}
}
