package handlers

import (
	"time"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
)

// SecondFactorPasswordPOST is the handler performing the first factory.
func SecondFactorPasswordPOST(delayFunc middlewares.TimingAttackDelayFunc) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var successful bool

		requestTime := time.Now()

		if delayFunc != nil {
			defer delayFunc(ctx, requestTime, &successful)
		}

		bodyJSON := bodySecondFactorPasswordRequest{}

		if err := ctx.ParseBody(&bodyJSON); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrParseRequestBody, regulation.AuthType1FA)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		provider, err := ctx.GetSessionProvider()
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to get session provider during 2FA attempt")

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		userSession, err := provider.GetSession(ctx.RequestCtx)
		if err != nil {
			ctx.Logger.Errorf("%s", err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		userPasswordOk, err := ctx.Providers.UserProvider.CheckUserPassword(userSession.Username, bodyJSON.Password)
		if err != nil {
			_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypePassword, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if !userPasswordOk {
			_ = markAuthenticationAttempt(ctx, false, nil, userSession.Username, regulation.AuthTypePassword, nil)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if err = markAuthenticationAttempt(ctx, true, nil, userSession.Username, regulation.AuthTypePassword, nil); err != nil {
			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		userSession.SetTwoFactorPassword(ctx.Clock.Now())

		if err = ctx.RegenerateSession(); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionRegenerate, regulation.AuthTypePassword, userSession.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated profile", regulation.AuthTypePassword, logFmtActionAuthentication, userSession.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		successful = true

		if bodyJSON.Workflow == workflowOpenIDConnect {
			handleOIDCWorkflowResponse(ctx, &userSession, bodyJSON.TargetURL, bodyJSON.WorkflowID)
		} else {
			Handle2FAResponse(ctx, bodyJSON.TargetURL)
		}
	}
}
