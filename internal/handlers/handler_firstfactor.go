package handlers

import (
	"errors"
	"time"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
)

// FirstFactorPOST is the handler performing the first factory.
//
//nolint:gocyclo // TODO: Consider refactoring time permitting.
func FirstFactorPOST(delayFunc middlewares.TimingAttackDelayFunc) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var successful bool

		requestTime := time.Now()

		if delayFunc != nil {
			defer delayFunc(ctx, requestTime, &successful)
		}

		bodyJSON := bodyFirstFactorRequest{}

		if err := ctx.ParseBody(&bodyJSON); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrParseRequestBody, regulation.AuthType1FA)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if bannedUntil, err := ctx.Providers.Regulator.Regulate(ctx, bodyJSON.Username); err != nil {
			if errors.Is(err, regulation.ErrUserIsBanned) {
				_ = markAuthenticationAttempt(ctx, false, &bannedUntil, bodyJSON.Username, regulation.AuthType1FA, nil)

				respondUnauthorized(ctx, messageAuthenticationFailed)

				return
			}

			ctx.Logger.WithError(err).Errorf(logFmtErrRegulationFail, regulation.AuthType1FA, bodyJSON.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		userPasswordOk, err := ctx.Providers.UserProvider.CheckUserPassword(bodyJSON.Username, bodyJSON.Password)
		if err != nil {
			_ = markAuthenticationAttempt(ctx, false, nil, bodyJSON.Username, regulation.AuthType1FA, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if !userPasswordOk {
			_ = markAuthenticationAttempt(ctx, false, nil, bodyJSON.Username, regulation.AuthType1FA, nil)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if err = markAuthenticationAttempt(ctx, true, nil, bodyJSON.Username, regulation.AuthType1FA, nil); err != nil {
			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		provider, err := ctx.GetSessionProvider()
		if err != nil {
			ctx.Logger.WithError(err).Error("Failed to get session provider during 1FA attempt")

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		userSession, err := provider.GetSession(ctx.RequestCtx)
		if err != nil {
			ctx.Logger.Errorf("%s", err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		newSession := provider.NewDefaultUserSession()

		// Reset all values from previous session except OIDC workflow before regenerating the cookie.
		if err = ctx.SaveSession(newSession); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionReset, regulation.AuthType1FA, bodyJSON.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if err = ctx.RegenerateSession(); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionRegenerate, regulation.AuthType1FA, bodyJSON.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		// Check if bodyJSON.KeepMeLoggedIn can be deref'd and derive the value based on the configuration and JSON data.
		keepMeLoggedIn := !provider.Config.DisableRememberMe && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

		// Set the cookie to expire if remember me is enabled and the user has asked us to.
		if keepMeLoggedIn {
			err = provider.UpdateExpiration(ctx.RequestCtx, provider.Config.RememberMe)
			if err != nil {
				ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated expiration", regulation.AuthType1FA, logFmtActionAuthentication, bodyJSON.Username)

				respondUnauthorized(ctx, messageAuthenticationFailed)

				return
			}
		}

		// Get the details of the given user from the user provider.
		userDetails, err := ctx.Providers.UserProvider.GetDetails(bodyJSON.Username)
		if err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrObtainProfileDetails, regulation.AuthType1FA, bodyJSON.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		ctx.Logger.Tracef(logFmtTraceProfileDetails, bodyJSON.Username, userDetails.Groups, userDetails.Emails)

		userSession.SetOneFactor(ctx.Clock.Now(), userDetails, keepMeLoggedIn)

		if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
			userSession.RefreshTTL = ctx.Clock.Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
		}

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated profile", regulation.AuthType1FA, logFmtActionAuthentication, bodyJSON.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		successful = true

		if bodyJSON.Workflow == workflowOpenIDConnect {
			handleOIDCWorkflowResponse(ctx, &userSession, bodyJSON.TargetURL, bodyJSON.WorkflowID)
		} else {
			Handle1FAResponse(ctx, bodyJSON.TargetURL, bodyJSON.RequestMethod, userSession.Username, userSession.Groups)
		}
	}
}
