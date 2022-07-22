package handler

import (
	"bytes"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middleware"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// VerifyGET returns the handler verifying if a request is allowed to go through.
func VerifyGET(config schema.AuthenticationBackendConfiguration) middleware.RequestHandler {
	profileRefreshEnabled, profileRefreshInterval := getProfileRefreshSettings(config)

	return func(ctx *middleware.AutheliaCtx) {
		var (
			targetURL *url.URL
			err       error
		)

		if targetURL = handleVerifyGETTargetURL(ctx); targetURL == nil {
			ctx.ReplyUnauthorized()

			return
		}

		var authn Authentication

		if authn, err = handleVerifyGETAuthn(ctx, profileRefreshEnabled, profileRefreshInterval); err != nil {
			switch {
			case authn.Type == AuthTypeAuthorization:
				break
			default:
				ctx.Logger.Errorf("Failed to validate the authentication state: %+v", err)
				ctx.ReplyUnauthorized()

				return
			}
		}

		method := ctx.XForwardedMethod()

		levelRequired := ctx.Providers.Authorizer.GetRequiredLevel(
			authorization.Subject{
				Username: authn.Details.Username,
				Groups:   authn.Details.Groups,
				IP:       ctx.RemoteIP(),
			},
			authorization.NewObjectRaw(targetURL, method),
		)

		switch isAuthorizationMatching(levelRequired, authn.Level) {
		case Forbidden:
			handleVerifyGETUnauthorized(ctx, method, targetURL, &authn, true)
		case NotAuthorized:
			handleVerifyGETUnauthorized(ctx, method, targetURL, &authn, false)
		case Authorized:
			setForwardedHeaders(&ctx.Response.Header, &authn)
		}
	}
}

// VerifyGETProxyNGINX returns the handler verifying if a request is allowed to go through.
func VerifyGETProxyNGINX(config schema.AuthenticationBackendConfiguration) middleware.RequestHandler {
	profileRefreshEnabled, profileRefreshInterval := getProfileRefreshSettings(config)

	return func(ctx *middleware.AutheliaCtx) {
		var (
			targetURL *url.URL
			err       error
		)

		if targetURL = handleVerifyGETTargetURL(ctx); targetURL == nil {
			ctx.ReplyUnauthorized()

			return
		}

		var authn Authentication

		if authn, err = handleVerifyGETAuthn(ctx, profileRefreshEnabled, profileRefreshInterval); err != nil {
			switch {
			case authn.Type == AuthTypeAuthorization:
				break
			default:
				ctx.Logger.Errorf("Failed to validate the authentication state: %+v", err)
				ctx.ReplyUnauthorized()

				return
			}
		}

		method := ctx.XForwardedMethod()

		levelRequired := ctx.Providers.Authorizer.GetRequiredLevel(
			authorization.Subject{
				Username: authn.Details.Username,
				Groups:   authn.Details.Groups,
				IP:       ctx.RemoteIP(),
			},
			authorization.NewObjectRaw(targetURL, method),
		)

		switch isAuthorizationMatching(levelRequired, authn.Level) {
		case Forbidden:
			handleVerifyGETProxyNGINXUnauthorized(ctx, method, targetURL, &authn, true)
		case NotAuthorized:
			handleVerifyGETProxyNGINXUnauthorized(ctx, method, targetURL, &authn, false)
		case Authorized:
			setForwardedHeaders(&ctx.Response.Header, &authn)
		}
	}
}

func handleVerifyGETTargetURL(ctx *middleware.AutheliaCtx) (targetURL *url.URL) {
	var err error

	if targetURL, err = ctx.GetOriginalURL(); err != nil {
		ctx.Logger.Errorf("Failed to parse Endpoint URL: %+v", err)

		return nil
	}

	if !isSchemeSecure(targetURL) {
		ctx.Logger.Errorf("Endpoint URL '%s' has an insecure scheme '%s', only the 'https' and 'wss' schemes are supported so session cookies can be transmitted securely", targetURL, targetURL.Scheme)

		return nil
	}

	if !isURLUnderProtectedDomain(targetURL, ctx.Configuration.Session.Domain) {
		ctx.Logger.Errorf("Endpoint URL '%s' is not on a domain which is a direct subdomain of the configured session domain '%s'", targetURL, ctx.Configuration.Session.Domain)

		return nil
	}

	return targetURL
}

func handleVerifyGETAuthn(ctx *middleware.AutheliaCtx, profileRefreshEnabled bool, profileRefreshInterval time.Duration) (authn Authentication, err error) {
	if authn, err = handleVerifyGETAuthnHeader(ctx); err != nil {
		return Authentication{Level: authentication.NotAuthenticated, Type: authn.Type}, err
	}

	if authn.Type != AuthTypeNone {
		return authn, nil
	}

	return handleVerifyGETAuthnCookie(ctx, profileRefreshEnabled, profileRefreshInterval), nil
}

func handleVerifyGETAuthnCookie(ctx *middleware.AutheliaCtx, profileRefreshEnabled bool, profileRefreshInterval time.Duration) (authn Authentication) {
	var err error

	authn.Type = AuthTypeCookie

	userSession := ctx.GetSession()

	if invalid := handleVerifyGETAuthnCookieValidate(ctx, &userSession, profileRefreshEnabled, profileRefreshInterval); invalid {
		if err = ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx); err != nil {
			ctx.Logger.Errorf("Unable to destroy user session: %+v", err)
		}

		return authn
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf("Unable to save updated user session: %+v", err)
	}

	return Authentication{
		Details: authentication.UserDetails{
			Username:    userSession.Username,
			DisplayName: userSession.DisplayName,
			Emails:      userSession.Emails,
			Groups:      userSession.Groups,
		},
		Level: userSession.AuthenticationLevel,
		Type:  AuthTypeCookie,
	}
}

func handleVerifyGETAuthnCookieValidate(ctx *middleware.AutheliaCtx, userSession *session.UserSession, profileRefreshEnabled bool, profileRefreshInterval time.Duration) (invalid bool) {
	isAnonymous := userSession.Username == ""

	if isAnonymous && userSession.AuthenticationLevel != authentication.NotAuthenticated {
		ctx.Logger.Errorf("Session for anonymous user has an authentication level of '%s': this may be a sign of a compromise", userSession.AuthenticationLevel)

		return true
	}

	if invalid = handleVerifyGETAuthnCookieValidateInactivity(ctx, userSession, isAnonymous); invalid {
		ctx.Logger.Infof("Session for user '%s' not marked as remembereded has exceeded configured session inactivity", userSession.Username)

		return true
	}

	if invalid = handleVerifyGETAuthnCookieValidateUpdate(ctx, userSession, isAnonymous, profileRefreshEnabled, profileRefreshInterval); invalid {
		return true
	}

	if username := ctx.Request.Header.PeekBytes(headerSessionUsername); username != nil && !strings.EqualFold(string(username), userSession.Username) {
		ctx.Logger.Warnf("Session for user '%s' does not match the Session-Username header with value '%s' which could be a sign of a cookie hijack", userSession.Username, username)

		return true
	}

	if !userSession.KeepMeLoggedIn {
		userSession.LastActivity = ctx.Clock.Now().Unix()
	}

	return false
}

func handleVerifyGETAuthnCookieValidateInactivity(ctx *middleware.AutheliaCtx, userSession *session.UserSession, isAnonymous bool) (invalid bool) {
	if isAnonymous || userSession.KeepMeLoggedIn || int64(ctx.Providers.SessionProvider.Inactivity.Seconds()) == 0 {
		return false
	}

	ctx.Logger.Tracef("Inactivity report for user '%s'. Current Time: %d, Last Activity: %d, Maximum Inactivity: %d.", userSession.Username, ctx.Clock.Now().Unix(), userSession.LastActivity, int(ctx.Providers.SessionProvider.Inactivity.Seconds()))

	return time.Unix(userSession.LastActivity, 0).Add(ctx.Providers.SessionProvider.Inactivity).Before(ctx.Clock.Now())
}

func handleVerifyGETAuthnCookieValidateUpdate(ctx *middleware.AutheliaCtx, userSession *session.UserSession, isAnonymous, enabled bool, interval time.Duration) (invalid bool) {
	if !enabled || isAnonymous {
		return false
	}

	ctx.Logger.Tracef("Checking if we need check the authentication backend for an updated profile for user '%s'", userSession.Username)

	if interval != schema.RefreshIntervalAlways && userSession.RefreshTTL.After(ctx.Clock.Now()) {
		return false
	}

	ctx.Logger.Debugf("Checking the authentication backend for an updated profile for user '%s'", userSession.Username)

	var (
		details *authentication.UserDetails
		err     error
	)

	if details, err = ctx.Providers.UserProvider.GetDetails(userSession.Username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			ctx.Logger.Errorf("Error occurred while attempting to update user details for user '%s': the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login", userSession.Username)

			return true
		}

		ctx.Logger.Errorf("Error occurred while attempting to update user details for user '%s': %v", userSession.Username, err)

		return false
	}

	var (
		diffEmails, diffGroups, diffDisplayName bool
	)

	diffEmails, diffGroups = utils.IsStringSlicesDifferent(userSession.Emails, details.Emails), utils.IsStringSlicesDifferent(userSession.Groups, details.Groups)
	diffDisplayName = userSession.DisplayName != details.DisplayName

	if interval != schema.RefreshIntervalAlways {
		userSession.RefreshTTL = ctx.Clock.Now().Add(interval)
	}

	if !diffEmails && !diffGroups && !diffDisplayName {
		ctx.Logger.Tracef("Updated profile not detected for user '%s'", userSession.Username)

		return false
	}

	ctx.Logger.Debugf("Updated profile detected for user '%s'", userSession.Username)

	if ctx.Configuration.Log.Level == "trace" {
		generateVerifySessionHasUpToDateProfileTraceLogs(ctx, userSession, details)
	}

	userSession.Emails, userSession.Groups, userSession.DisplayName = details.Emails, details.Groups, details.DisplayName

	return false
}

func handleVerifyGETAuthnHeader(ctx *middleware.AutheliaCtx) (authn Authentication, err error) {
	var header, value []byte

	forced := false

	switch {
	case bytes.Equal(ctx.QueryArgs().PeekBytes(queryArgAuth), valueBasic):
		forced = true
		header = headerAuthorization
	default:
		header = headerProxyAuthorization
	}

	value = ctx.Request.Header.PeekBytes(header)

	switch {
	case forced && value == nil:
		return Authentication{Level: authentication.NotAuthenticated, Type: AuthTypeAuthorization}, errMissingAuthorizationHeaderSchemeBasicForced
	case value != nil:
		return handleVerifyGETAuthnHeaderBasic(ctx, header, value, forced)
	default:
		return Authentication{Level: authentication.NotAuthenticated, Type: AuthTypeNone}, nil
	}
}

func handleVerifyGETAuthnHeaderBasic(ctx *middleware.AutheliaCtx, header, value []byte, forced bool) (authn Authentication, err error) {
	var (
		username, password string
		valid              bool
		details            *authentication.UserDetails
	)

	if forced {
		authn.Type = AuthTypeAuthorization
	} else {
		authn.Type = AuthTypeProxyAuthorization
	}

	if username, password, err = headerAuthorizationParseBasic(value); err != nil {
		return authn, fmt.Errorf("unable to parse content of %s header: %w", header, err)
	}

	if valid, err = ctx.Providers.UserProvider.CheckUserPassword(username, password); err != nil {
		return authn, fmt.Errorf("unable to check credentials extracted from %s header: %w", header, err)
	}

	if !valid {
		return authn, fmt.Errorf("credentials for username '%s' are invalid", username)
	}

	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			ctx.Logger.Errorf("Error occurred while attempting to get user details for user '%s': the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login", username)

			return authn, err
		}

		return authn, fmt.Errorf("unable to retrieve details for user '%s': %w", username, err)
	}

	authn.Details = *details
	authn.Level = authentication.OneFactor

	return authn, nil
}

func fmtFriendlyUsername(username string) string {
	switch username {
	case "":
		return "<anonymous>"
	default:
		return username
	}
}

func fmtFriendlyMethod(method string) string {
	switch method {
	case "":
		return "unknown"
	default:
		return method
	}
}

func handleVerifyGETUnauthorized(ctx *middleware.AutheliaCtx, method []byte, targetURL *url.URL, authn *Authentication, forbidden bool) {
	var (
		termAccess string
		statusCode int
	)

	switch {
	case forbidden:
		termAccess = "forbidden"
		statusCode = fasthttp.StatusForbidden
	default:
		termAccess = "not authorized"
		statusCode = fasthttp.StatusUnauthorized
	}

	strMethod := string(method)

	rd, rm, username := string(ctx.QueryArgs().PeekBytes(queryArgRD)), fmtFriendlyMethod(strMethod), fmtFriendlyUsername(authn.Details.Username)

	if authn.Type == AuthTypeAuthorization {
		ctx.Logger.Infof("Access to '%s' (method %s) is %s to user '%s', responding with 401 Unauthorized and a Basic scheme WWW-Authenticate header", targetURL, rm, termAccess, username)

		switch {
		case forbidden:
			ctx.ReplyForbidden()
		default:
			ctx.ReplyUnauthorized()
			ctx.Response.Header.SetBytesKV(headerWWWAuthenticate, headerWWWAuthenticateValueBasic)
		}

		return
	}

	var (
		redirectionURL *url.URL

		err error
	)

	if redirectionURL, err = handleVerifyGETRedirectionURL(rd, strMethod, targetURL, forbidden); err != nil {
		ctx.Logger.Errorf("Failed to determine redirect URL: %v", err)
		ctx.ReplyBadRequest()

		return
	}

	if redirectionURL == nil {
		ctx.Logger.Infof("Access to '%s' (method %s) is %s to user '%s', responding with 401 Unauthorized without a Location header", targetURL, rm, termAccess, username)
		ctx.ReplyStatusCode(statusCode)

		return
	}

	switch {
	case ctx.IsXHR() || !ctx.AcceptsMIME(headerAcceptsMIMETextHTML):
		break
	default:
		switch strMethod {
		case fasthttp.MethodGet, fasthttp.MethodOptions, "":
			statusCode = fasthttp.StatusFound
		default:
			statusCode = fasthttp.StatusSeeOther
		}
	}

	ctx.Logger.Infof("Access to '%s' (method %s) is %s to user '%s', responding with %d %s with Location header '%s'", targetURL, rm, termAccess, username, statusCode, fasthttp.StatusMessage(statusCode), redirectionURL)
	ctx.SpecialRedirect(redirectionURL.String(), statusCode)
}

func handleVerifyGETProxyNGINXUnauthorized(ctx *middleware.AutheliaCtx, method []byte, targetURL *url.URL, authn *Authentication, forbidden bool) {
	var (
		termAccess string
	)

	switch {
	case forbidden:
		termAccess = "forbidden"
	default:
		termAccess = "not authorized"
	}

	strMethod := string(method)

	rd, rm, username := string(ctx.QueryArgs().PeekBytes(queryArgRD)), fmtFriendlyMethod(strMethod), fmtFriendlyUsername(authn.Details.Username)

	if authn.Type == AuthTypeAuthorization {
		ctx.Logger.Infof("Access to '%s' (method %s) is %s to user '%s', responding with 401 Unauthorized and a Basic scheme WWW-Authenticate header", targetURL, rm, termAccess, username)

		ctx.ReplyUnauthorized()
		ctx.Response.Header.SetBytesKV(headerWWWAuthenticate, headerWWWAuthenticateValueBasic)

		return
	}

	var (
		redirectionURL *url.URL

		err error
	)

	if redirectionURL, err = handleVerifyGETRedirectionURL(rd, strMethod, targetURL, forbidden); err != nil {
		ctx.Logger.Errorf("Failed to determine redirect URL: %v", err)
		ctx.ReplyBadRequest()

		return
	}

	if redirectionURL == nil {
		ctx.Logger.Errorf("Redirection URL parameter was not specified and is required for this endpoint.")
		ctx.ReplyBadRequest()

		return
	}

	ctx.Logger.Infof("Access to '%s' (method %s) is %s to user '%s', responding with %d %s with Location header '%s'", targetURL, rm, termAccess, username, fasthttp.StatusUnauthorized, fasthttp.StatusMessage(fasthttp.StatusUnauthorized), redirectionURL)
	ctx.SpecialRedirect(redirectionURL.String(), fasthttp.StatusUnauthorized)
}

// generateVerifySessionHasUpToDateProfileTraceLogs is used to generate trace logs only when trace logging is enabled.
// The information calculated in this function is completely useless other than trace for now.
func generateVerifySessionHasUpToDateProfileTraceLogs(ctx *middleware.AutheliaCtx, userSession *session.UserSession,
	details *authentication.UserDetails) {
	groupsAdded, groupsRemoved := utils.StringSlicesDelta(userSession.Groups, details.Groups)
	emailsAdded, emailsRemoved := utils.StringSlicesDelta(userSession.Emails, details.Emails)
	nameDelta := userSession.DisplayName != details.DisplayName

	// Check Groups.
	var groupsDelta []string
	if len(groupsAdded) != 0 {
		groupsDelta = append(groupsDelta, fmt.Sprintf("added: %s.", strings.Join(groupsAdded, ", ")))
	}

	if len(groupsRemoved) != 0 {
		groupsDelta = append(groupsDelta, fmt.Sprintf("removed: %s.", strings.Join(groupsRemoved, ", ")))
	}

	if len(groupsDelta) != 0 {
		ctx.Logger.Tracef("Updated groups detected for %s. %s", userSession.Username, strings.Join(groupsDelta, " "))
	} else {
		ctx.Logger.Tracef("No updated groups detected for %s", userSession.Username)
	}

	// Check Emails.
	var emailsDelta []string
	if len(emailsAdded) != 0 {
		emailsDelta = append(emailsDelta, fmt.Sprintf("added: %s.", strings.Join(emailsAdded, ", ")))
	}

	if len(emailsRemoved) != 0 {
		emailsDelta = append(emailsDelta, fmt.Sprintf("removed: %s.", strings.Join(emailsRemoved, ", ")))
	}

	if len(emailsDelta) != 0 {
		ctx.Logger.Tracef("Updated emails detected for %s. %s", userSession.Username, strings.Join(emailsDelta, " "))
	} else {
		ctx.Logger.Tracef("No updated emails detected for %s", userSession.Username)
	}

	// Check Name.
	if nameDelta {
		ctx.Logger.Tracef("Updated display name detected for %s. Added: %s. Removed: %s.", userSession.Username, details.DisplayName, userSession.DisplayName)
	} else {
		ctx.Logger.Tracef("No updated display name detected for %s", userSession.Username)
	}
}
