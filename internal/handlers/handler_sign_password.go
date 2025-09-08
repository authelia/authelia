package handlers

import (
	"time"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

// SecondFactorPasswordPOST is the handler performing the knowledge based authentication factor after a user utilizes a
// alternative to usernames and passwords like passkeys.
func SecondFactorPasswordPOST(delayFunc middlewares.TimingAttackDelayFunc) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var successful bool

		requestTime := time.Now()

		if delayFunc != nil {
			defer delayFunc(ctx, requestTime, &successful)
		}

		bodyJSON := bodySecondFactorPasswordRequest{}

		var err error
		if err = ctx.ParseBody(&bodyJSON); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrParseRequestBody, regulation.AuthType1FA)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		var (
			provider    *session.Session
			userSession session.UserSession
		)

		if provider, err = ctx.GetSessionProvider(); err != nil {
			ctx.Logger.WithError(err).Error("Failed to get session provider during 2FA attempt")

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if userSession, err = provider.GetSession(ctx.RequestCtx); err != nil {
			ctx.Logger.Errorf("%s", err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		var (
			userPasswordOk bool
		)

		if userPasswordOk, err = ctx.Providers.UserProvider.CheckUserPassword(userSession.Username, bodyJSON.Password); err != nil {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypePassword, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if !userPasswordOk {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypePassword, nil)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthTypePassword, nil)

		userSession.SetTwoFactorPassword(ctx.GetClock().Now())

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

		if len(bodyJSON.Flow) > 0 {
			handleFlowResponse(ctx, &userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
		} else {
			Handle2FAResponse(ctx, bodyJSON.TargetURL)
		}
	}
}
