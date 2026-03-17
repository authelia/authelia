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
func FirstFactorPasswordPOST(delayer middlewares.Delayer) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var successful bool

		requestTime := time.Now()

		if delayer != nil {
			defer delayer.Delay(ctx, requestTime, &successful)
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
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeUnknown, "", nil), regulation.AuthType1FA, err)

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
			if isRegulatorSkippedErr(err) {
				ctx.Logger.WithError(err).Errorf("Unsuccessful %s authentication attempt by user '%s'", regulation.AuthType1FA, details.Username)
			} else {
				doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthType1FA, err)
			}

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if !userPasswordOk {
			doMarkAuthenticationAttempt(ctx, false, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthType1FA, nil)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		doMarkAuthenticationAttempt(ctx, true, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthType1FA, nil)

		var provider session.Strategy

		if provider, err = ctx.GetSessionProvider(); err != nil {
			ctx.Logger.WithError(err).Error("Failed to get session provider during 1FA attempt")

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if err = provider.Destroy(ctx); err != nil {
			// This failure is not likely to be critical as we ensure to regenerate the session below.
			ctx.Logger.WithError(err).Trace("Failed to destroy session during 1FA attempt")
		}

		userSession := provider.NewDefault()

		// Reset all values from previous session except OIDC workflow before regenerating the cookie.
		if err = provider.Save(ctx, &userSession); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionReset, regulation.AuthType1FA, details.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		if err = provider.Regenerate(ctx); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionRegenerate, regulation.AuthType1FA, details.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		// Check if bodyJSON.KeepMeLoggedIn can be deref'd and derive the value based on the configuration and JSON data.
		keepMeLoggedIn := !provider.GetConfig().DisableRememberMe && bodyJSON.KeepMeLoggedIn != nil && *bodyJSON.KeepMeLoggedIn

		ctx.Logger.Tracef(logFmtTraceProfileDetails, details.Username, details.Groups, details.Emails)

		userSession.SetOneFactorPassword(ctx.GetClock().Now(), details.Username, keepMeLoggedIn)

		if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
			userSession.RefreshTTL = ctx.GetClock().Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
		}

		if err = provider.Save(ctx, &userSession); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated profile", regulation.AuthType1FA, logFmtActionAuthentication, details.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		successful = true

		if len(bodyJSON.Flow) > 0 {
			handleFlowResponse(ctx, &userSession, details, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
		} else {
			Handle1FAResponse(ctx, bodyJSON.TargetURL, bodyJSON.RequestMethod, userSession.Username, details.Groups)
		}
	}
}

// FirstFactorReauthenticatePOST is a specialized handler which checks the currently logged in users current password
// and updates their last authenticated time.
func FirstFactorReauthenticatePOST(delayer middlewares.Delayer) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		var successful bool

		requestTime := time.Now()

		if delayer != nil {
			defer delayer.Delay(ctx, requestTime, &successful)
		}

		bodyJSON := bodyFirstFactorReauthenticateRequest{}

		var err error
		if err = ctx.ParseBody(&bodyJSON); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrParseRequestBody, regulation.AuthType1FA)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		var (
			provider    session.Strategy
			userSession session.UserSession
		)

		if provider, err = ctx.GetSessionProvider(); err != nil {
			ctx.Logger.WithError(err).Error("Failed to get session provider during 1FA attempt")

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		var current *session.UserSession

		if current, err = provider.Get(ctx); err != nil {
			ctx.Logger.WithError(err).Errorf("Error occurred attempting to load session.")

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		userSession = *current

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
			details *authentication.UserDetails
		)

		if details, err = ctx.Providers.UserProvider.GetDetails(userSession.Username); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrObtainProfileDetails, regulation.AuthType1FA, userSession.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		ctx.Logger.Tracef(logFmtTraceProfileDetails, userSession.Username, details.Groups, details.Emails)

		userSession.SetOneFactorReauthenticate(ctx.GetClock().Now())

		if ctx.Configuration.AuthenticationBackend.RefreshInterval.Update() {
			userSession.RefreshTTL = ctx.GetClock().Now().Add(ctx.Configuration.AuthenticationBackend.RefreshInterval.Value())
		}

		if err = ctx.SaveSession(&userSession); err != nil {
			ctx.Logger.WithError(err).Errorf(logFmtErrSessionSave, "updated profile", regulation.AuthType1FA, logFmtActionAuthentication, userSession.Username)

			respondUnauthorized(ctx, messageAuthenticationFailed)

			return
		}

		successful = true

		if len(bodyJSON.Flow) > 0 {
			handleFlowResponse(ctx, &userSession, details, bodyJSON.FlowID, bodyJSON.Flow, bodyJSON.SubFlow, bodyJSON.UserCode)
		} else {
			Handle1FAResponse(ctx, bodyJSON.TargetURL, bodyJSON.RequestMethod, userSession.Username, details.Groups)
		}
	}
}
