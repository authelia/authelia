package handlers

import (
	"errors"
	"time"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
)

// FirstFactorPasswordPOST is the handler performing the first factor authn with a password.
//
//nolint:gocyclo // TODO: Consider refactoring time permitting.
func FirstFactorPasswordPOST(delayFunc middlewares.TimingAttackDelayFunc) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var successful bool

		requestTime := time.Now()

		if delayFunc != nil {
			defer delayFunc(ctx, requestTime, &successful)
		}

		bodyJSON := bodyFirstFactorRequest{}

		var (
			details *authentication.UserDetails
			err     error
		)
		if err = ctx.ParseBody(&bodyJSON); err != nil {
			respondUnauthorized(ctx, messageAuthenticationFailed)

			ctx.Logger.WithError(err).Errorf(logFmtErrParseRequestBody, regulation.AuthType1FA)

			return
		}

		if details, err = ctx.Providers.UserProvider.GetDetails(bodyJSON.Username); err != nil || details == nil {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, "", nil), regulation.AuthType1FA, err)

			ctx.Logger.WithError(err).Errorf("Error occurred getting details for user with username input '%s' which usually indicates they do not exist", bodyJSON.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if ban, _, expires, err := ctx.Providers.Regulator.BanCheck(ctx, details.Username); err != nil {
			if errors.Is(err, regulation.ErrUserIsBanned) {
				doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(ban, details.Username, expires), regulation.AuthType1FA, nil)

				respondUnauthorized(ctx, messageAuthenticationFailed)

				return
			}

			ctx.Logger.WithError(err).Errorf(logFmtErrRegulationFail, regulation.AuthType1FA, details.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		userPasswordOk, err := ctx.Providers.UserProvider.CheckUserPassword(details.Username, bodyJSON.Password)
		if err != nil {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthType1FA, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if !userPasswordOk {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthType1FA, nil)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthType1FA, nil)

		var provider *session.Session

		if provider, err = ctx.GetSessionProvider(); err != nil {
			ctx.Logger.WithError(err).Error("Failed to get session provider during 1FA attempt")

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if err = provider.DestroySession(ctx.RequestCtx); err != nil {
			// This failure is not likely to be critical as we ensure to regenerate the session below.
			ctx.Logger.WithError(err).Trace("Failed to destroy session during 1FA attempt")
		}

		userSession := provider.NewDefaultUserSession()

		// Reset all values from previous session except OIDC workflow before regenerating the cookie.
		if err = provider.SaveSession(ctx.RequestCtx, userSession); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionReset, regulation.AuthType1FA, details.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if err = provider.RegenerateSession(ctx.RequestCtx); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionRegenerate, regulation.AuthType1FA, details.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		// Check if bodyJSON.KeepMeLoggedIn can be deref'd and derive the value based on the configuration and JSON data.
		keepMeLoggedIn := !provider.Config.DisableRememberMe && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

		// Set the cookie to expire if remember me is enabled and the user has asked us to.
		if keepMeLoggedIn {
			err = provider.UpdateExpiration(ctx.RequestCtx, provider.Config.RememberMe)
			if err != nil {
				ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated expiration", regulation.AuthType1FA, logFmtActionAuthentication, details.Username)

				respondUnauthorized(ctx, messageAuthenticationFailed)

				return
			}
		}

		ctx.Logger.Tracef(logFmtTraceProfileDetails, details.Username, details.Groups, details.Emails)

		userSession.SetOneFactorPassword(ctx.GetClock().Now(), details, keepMeLoggedIn)

		if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
			userSession.RefreshTTL = ctx.GetClock().Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
		}

		if err = provider.SaveSession(ctx.RequestCtx, userSession); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated profile", regulation.AuthType1FA, logFmtActionAuthentication, details.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		successful = true

		if len(bodyJSON.Flow) > 0 {
			handleFlowResponse(ctx, &userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
		} else {
			Handle1FAResponse(ctx, bodyJSON.TargetURL, bodyJSON.RequestMethod, userSession.Username, userSession.Groups)
		}
	}
}

// FirstFactorReauthenticatePOST is a specialized handler which checks the currently logged in users current password
// and updates their last authenticated time.
func FirstFactorReauthenticatePOST(delayFunc middlewares.TimingAttackDelayFunc) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var successful bool

		requestTime := time.Now()

		if delayFunc != nil {
			defer delayFunc(ctx, requestTime, &successful)
		}

		bodyJSON := bodyFirstFactorReauthenticateRequest{}

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
			ctx.Logger.WithError(err).Error("Failed to get session provider during 1FA attempt")

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if userSession, err = provider.GetSession(ctx.RequestCtx); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred attempting to load session.")

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		var (
			ban     regulation.BanType
			value   string
			expires *time.Time
		)

		if ban, value, expires, err = ctx.Providers.Regulator.BanCheck(ctx, userSession.Username); err != nil {
			if errors.Is(err, regulation.ErrUserIsBanned) {
				doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(ban, value, expires), regulation.AuthType1FA, nil)

				respondUnauthorized(ctx, messageAuthenticationFailed)

				return
			}

			ctx.Logger.WithError(err).Errorf(logFmtErrRegulationFail, regulation.AuthType1FA, userSession.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		userPasswordOk, err := ctx.Providers.UserProvider.CheckUserPassword(userSession.Username, bodyJSON.Password)
		if err != nil {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthType1FA, err)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if !userPasswordOk {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthType1FA, nil)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, userSession.Username, nil), regulation.AuthType1FA, nil)

		if err = ctx.RegenerateSession(); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionRegenerate, regulation.AuthType1FA, userSession.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		var (
			userDetails *authentication.UserDetails
		)

		if userDetails, err = ctx.Providers.UserProvider.GetDetails(userSession.Username); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrObtainProfileDetails, regulation.AuthType1FA, userSession.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		ctx.Logger.Tracef(logFmtTraceProfileDetails, userSession.Username, userDetails.Groups, userDetails.Emails)

		userSession.SetOneFactorReauthenticate(ctx.GetClock().Now(), userDetails)

		if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
			userSession.RefreshTTL = ctx.GetClock().Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
		}

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated profile", regulation.AuthType1FA, logFmtActionAuthentication, userSession.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		successful = true

		if len(bodyJSON.Flow) > 0 {
			handleFlowResponse(ctx, &userSession, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
		} else {
			Handle1FAResponse(ctx, bodyJSON.TargetURL, bodyJSON.RequestMethod, userSession.Username, userSession.Groups)
		}
	}
}
