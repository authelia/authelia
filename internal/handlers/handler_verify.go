package handlers

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

func isURLUnderProtectedDomain(url *url.URL, domain string) bool {
	return strings.HasSuffix(url.Hostname(), domain)
}

func isSchemeHTTPS(url *url.URL) bool {
	return url.Scheme == "https"
}

func isSchemeWSS(url *url.URL) bool {
	return url.Scheme == "wss"
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(header []byte, auth string) (username, password string, err error) {
	if !strings.HasPrefix(auth, authPrefix) {
		return "", "", fmt.Errorf("%s prefix not found in %s header", strings.Trim(authPrefix, " "), header)
	}

	c, err := base64.StdEncoding.DecodeString(auth[len(authPrefix):])
	if err != nil {
		return "", "", err
	}

	cs := string(c)
	s := strings.IndexByte(cs, ':')

	if s < 0 {
		return "", "", fmt.Errorf("format of %s header must be user:password", header)
	}

	return cs[:s], cs[s+1:], nil
}

// isTargetURLAuthorized check whether the given user is authorized to access the resource.
func isTargetURLAuthorized(authorizer *authorization.Authorizer, targetURL url.URL,
	username string, userGroups []string, clientIP net.IP, method []byte, authLevel authentication.Level) authorizationMatching {
	level := authorizer.GetRequiredLevel(
		authorization.Subject{
			Username: username,
			Groups:   userGroups,
			IP:       clientIP,
		},
		authorization.NewObjectRaw(&targetURL, method))

	switch {
	case level == authorization.Bypass:
		return Authorized
	case level == authorization.Denied && username != "":
		// If the user is not anonymous, it means that we went through
		// all the rules related to that user and knowing who he is we can
		// deduce the access is forbidden
		// For anonymous users though, we cannot be sure that she
		// could not be granted the rights to access the resource. Consequently
		// for anonymous users we send Unauthorized instead of Forbidden.
		return Forbidden
	case level == authorization.OneFactor && authLevel >= authentication.OneFactor,
		level == authorization.TwoFactor && authLevel >= authentication.TwoFactor:
		return Authorized
	}

	return NotAuthorized
}

// verifyBasicAuth verify that the provided username and password are correct and
// that the user is authorized to target the resource.
func verifyBasicAuth(ctx *middlewares.AutheliaCtx, header, auth []byte) (username, name string, groups, emails []string, authLevel authentication.Level, err error) {
	username, password, err := parseBasicAuth(header, string(auth))

	if err != nil {
		return "", "", nil, nil, authentication.NotAuthenticated, fmt.Errorf("unable to parse content of %s header: %s", header, err)
	}

	authenticated, err := ctx.Providers.UserProvider.CheckUserPassword(username, password)

	if err != nil {
		return "", "", nil, nil, authentication.NotAuthenticated, fmt.Errorf("unable to check credentials extracted from %s header: %w", header, err)
	}

	// If the user is not correctly authenticated, send a 401.
	if !authenticated {
		// Request Basic Authentication otherwise.
		return "", "", nil, nil, authentication.NotAuthenticated, fmt.Errorf("user %s is not authenticated", username)
	}

	details, err := ctx.Providers.UserProvider.GetDetails(username)

	if err != nil {
		return "", "", nil, nil, authentication.NotAuthenticated, fmt.Errorf("unable to retrieve details of user %s: %s", username, err)
	}

	return username, details.DisplayName, details.Groups, details.Emails, authentication.OneFactor, nil
}

// setForwardedHeaders set the forwarded User, Groups, Name and Email headers.
func setForwardedHeaders(headers *fasthttp.ResponseHeader, username, name string, groups, emails []string) {
	if username != "" {
		headers.SetBytesK(headerRemoteUser, username)
		headers.SetBytesK(headerRemoteGroups, strings.Join(groups, ","))
		headers.SetBytesK(headerRemoteName, name)

		if emails != nil {
			headers.SetBytesK(headerRemoteEmail, emails[0])
		} else {
			headers.SetBytesK(headerRemoteEmail, "")
		}
	}
}

// hasUserBeenInactiveTooLong checks whether the user has been inactive for too long.
func hasUserBeenInactiveTooLong(ctx *middlewares.AutheliaCtx) (bool, error) { //nolint:unparam
	maxInactivityPeriod := int64(ctx.Providers.SessionProvider.Inactivity.Seconds())
	if maxInactivityPeriod == 0 {
		return false, nil
	}

	lastActivity := ctx.GetSession().LastActivity
	inactivityPeriod := ctx.Clock.Now().Unix() - lastActivity

	ctx.Logger.Tracef("Inactivity report: Inactivity=%d, MaxInactivity=%d",
		inactivityPeriod, maxInactivityPeriod)

	if inactivityPeriod > maxInactivityPeriod {
		return true, nil
	}

	return false, nil
}

// verifySessionCookie verifies if a user is identified by a cookie.
func verifySessionCookie(ctx *middlewares.AutheliaCtx, targetURL *url.URL, userSession *session.UserSession, refreshProfile bool,
	refreshProfileInterval time.Duration) (username, name string, groups, emails []string, authLevel authentication.Level, err error) {
	// No username in the session means the user is anonymous.
	isUserAnonymous := userSession.Username == ""

	if isUserAnonymous && userSession.AuthenticationLevel != authentication.NotAuthenticated {
		return "", "", nil, nil, authentication.NotAuthenticated, fmt.Errorf("an anonymous user cannot be authenticated. That might be the sign of a compromise")
	}

	if !userSession.KeepMeLoggedIn && !isUserAnonymous {
		inactiveLongEnough, err := hasUserBeenInactiveTooLong(ctx)
		if err != nil {
			return "", "", nil, nil, authentication.NotAuthenticated, fmt.Errorf("unable to check if user has been inactive for a long time: %s", err)
		}

		if inactiveLongEnough {
			// Destroy the session a new one will be regenerated on next request.
			err := ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx)
			if err != nil {
				return "", "", nil, nil, authentication.NotAuthenticated, fmt.Errorf("unable to destroy user session after long inactivity: %s", err)
			}

			ctx.Logger.Warnf("User %s has been inactive for too long", userSession.Username)

			// Fixes #3587.
			return "", "", nil, nil, authentication.NotAuthenticated, nil
		}
	}

	err = verifySessionHasUpToDateProfile(ctx, targetURL, userSession, refreshProfile, refreshProfileInterval)
	if err != nil {
		if err == authentication.ErrUserNotFound {
			err = ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx)
			if err != nil {
				ctx.Logger.Errorf("Unable to destroy user session after provider refresh didn't find the user: %s", err)
			}

			return userSession.Username, userSession.DisplayName, userSession.Groups, userSession.Emails, authentication.NotAuthenticated, err
		}

		ctx.Logger.Errorf("Error occurred while attempting to update user details from LDAP: %s", err)

		return "", "", nil, nil, authentication.NotAuthenticated, err
	}

	return userSession.Username, userSession.DisplayName, userSession.Groups, userSession.Emails, userSession.AuthenticationLevel, nil
}

func handleUnauthorized(ctx *middlewares.AutheliaCtx, targetURL fmt.Stringer, isBasicAuth bool, username string, method []byte) {
	var (
		statusCode            int
		redirectionURL        string
		friendlyUsername      string
		friendlyRequestMethod string
	)

	switch username {
	case "":
		friendlyUsername = "<anonymous>"
	default:
		friendlyUsername = username
	}

	if isBasicAuth {
		ctx.Logger.Infof("Access to %s is not authorized to user %s, sending 401 response with basic auth header", targetURL.String(), friendlyUsername)
		ctx.ReplyUnauthorized()
		ctx.Response.Header.Add("WWW-Authenticate", "Basic realm=\"Authentication required\"")

		return
	}

	// Kubernetes ingress controller and Traefik use the rd parameter of the verify
	// endpoint to provide the URL of the login portal. The target URL of the user
	// is computed from X-Forwarded-* headers or X-Original-URL.
	rd := string(ctx.QueryArgs().Peek("rd"))
	rm := string(method)

	switch rm {
	case "":
		friendlyRequestMethod = "unknown"
	default:
		friendlyRequestMethod = rm
	}

	if rd != "" {
		switch rm {
		case "":
			redirectionURL = fmt.Sprintf("%s?rd=%s", rd, url.QueryEscape(targetURL.String()))
		default:
			redirectionURL = fmt.Sprintf("%s?rd=%s&rm=%s", rd, url.QueryEscape(targetURL.String()), rm)
		}
	}

	switch {
	case ctx.IsXHR() || !ctx.AcceptsMIME("text/html") || rd == "":
		statusCode = fasthttp.StatusUnauthorized
	default:
		switch rm {
		case fasthttp.MethodGet, fasthttp.MethodOptions, "":
			statusCode = fasthttp.StatusFound
		default:
			statusCode = fasthttp.StatusSeeOther
		}
	}

	if redirectionURL != "" {
		ctx.Logger.Infof("Access to %s (method %s) is not authorized to user %s, responding with status code %d with location redirect to %s", targetURL.String(), friendlyRequestMethod, friendlyUsername, statusCode, redirectionURL)
		ctx.SpecialRedirect(redirectionURL, statusCode)
	} else {
		ctx.Logger.Infof("Access to %s (method %s) is not authorized to user %s, responding with status code %d", targetURL.String(), friendlyRequestMethod, friendlyUsername, statusCode)
		ctx.ReplyUnauthorized()
	}
}

func updateActivityTimestamp(ctx *middlewares.AutheliaCtx, isBasicAuth bool, username string) error {
	if isBasicAuth || username == "" {
		return nil
	}

	userSession := ctx.GetSession()
	// We don't need to update the activity timestamp when user checked keep me logged in.
	if userSession.KeepMeLoggedIn {
		return nil
	}

	// Mark current activity.
	userSession.LastActivity = ctx.Clock.Now().Unix()

	return ctx.SaveSession(userSession)
}

// generateVerifySessionHasUpToDateProfileTraceLogs is used to generate trace logs only when trace logging is enabled.
// The information calculated in this function is completely useless other than trace for now.
func generateVerifySessionHasUpToDateProfileTraceLogs(ctx *middlewares.AutheliaCtx, userSession *session.UserSession,
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

func verifySessionHasUpToDateProfile(ctx *middlewares.AutheliaCtx, targetURL *url.URL, userSession *session.UserSession,
	refreshProfile bool, refreshProfileInterval time.Duration) error {
	// TODO: Add a check for LDAP password changes based on a time format attribute.
	// See https://www.authelia.com/o/threatmodel#potential-future-guarantees
	ctx.Logger.Tracef("Checking if we need check the authentication backend for an updated profile for %s.", userSession.Username)

	if !refreshProfile || userSession.Username == "" || targetURL == nil {
		return nil
	}

	if refreshProfileInterval != schema.RefreshIntervalAlways && userSession.RefreshTTL.After(ctx.Clock.Now()) {
		return nil
	}

	ctx.Logger.Debugf("Checking the authentication backend for an updated profile for user %s", userSession.Username)
	details, err := ctx.Providers.UserProvider.GetDetails(userSession.Username)
	// Only update the session if we could get the new details.
	if err != nil {
		return err
	}

	emailsDiff := utils.IsStringSlicesDifferent(userSession.Emails, details.Emails)
	groupsDiff := utils.IsStringSlicesDifferent(userSession.Groups, details.Groups)
	nameDiff := userSession.DisplayName != details.DisplayName

	if !groupsDiff && !emailsDiff && !nameDiff {
		ctx.Logger.Tracef("Updated profile not detected for %s.", userSession.Username)
		// Only update TTL if the user has an interval set.
		// We get to this check when there were no changes.
		// Also make sure to update the session even if no difference was found.
		// This is so that we don't check every subsequent request after this one.
		if refreshProfileInterval != schema.RefreshIntervalAlways {
			// Update RefreshTTL and save session if refresh is not set to always.
			userSession.RefreshTTL = ctx.Clock.Now().Add(refreshProfileInterval)
			return ctx.SaveSession(*userSession)
		}
	} else {
		ctx.Logger.Debugf("Updated profile detected for %s.", userSession.Username)
		if ctx.Configuration.Log.Level == "trace" {
			generateVerifySessionHasUpToDateProfileTraceLogs(ctx, userSession, details)
		}
		userSession.Emails = details.Emails
		userSession.Groups = details.Groups
		userSession.DisplayName = details.DisplayName

		// Only update TTL if the user has a interval set.
		if refreshProfileInterval != schema.RefreshIntervalAlways {
			userSession.RefreshTTL = ctx.Clock.Now().Add(refreshProfileInterval)
		}
		// Return the result of save session if there were changes.
		return ctx.SaveSession(*userSession)
	}

	// Return nil if disabled or if no changes and refresh interval set to always.
	return nil
}

func getProfileRefreshSettings(cfg schema.AuthenticationBackendConfiguration) (refresh bool, refreshInterval time.Duration) {
	if cfg.LDAP != nil {
		if cfg.RefreshInterval == schema.ProfileRefreshDisabled {
			refresh = false
			refreshInterval = 0
		} else {
			refresh = true

			if cfg.RefreshInterval != schema.ProfileRefreshAlways {
				// Skip Error Check since validator checks it.
				refreshInterval, _ = utils.ParseDurationString(cfg.RefreshInterval)
			} else {
				refreshInterval = schema.RefreshIntervalAlways
			}
		}
	}

	return refresh, refreshInterval
}

func verifyAuth(ctx *middlewares.AutheliaCtx, targetURL *url.URL, refreshProfile bool, refreshProfileInterval time.Duration) (isBasicAuth bool, username, name string, groups, emails []string, authLevel authentication.Level, err error) {
	authHeader := headerProxyAuthorization
	if bytes.Equal(ctx.QueryArgs().Peek("auth"), []byte("basic")) {
		authHeader = headerAuthorization
		isBasicAuth = true
	}

	authValue := ctx.Request.Header.PeekBytes(authHeader)
	if authValue != nil {
		isBasicAuth = true
	} else if isBasicAuth {
		err = fmt.Errorf("basic auth requested via query arg, but no value provided via %s header", authHeader)
		return
	}

	if isBasicAuth {
		username, name, groups, emails, authLevel, err = verifyBasicAuth(ctx, authHeader, authValue)
		return
	}

	userSession := ctx.GetSession()
	username, name, groups, emails, authLevel, err = verifySessionCookie(ctx, targetURL, &userSession, refreshProfile, refreshProfileInterval)

	sessionUsername := ctx.Request.Header.PeekBytes(headerSessionUsername)
	if sessionUsername != nil && !strings.EqualFold(string(sessionUsername), username) {
		ctx.Logger.Warnf("Possible cookie hijack or attempt to bypass security detected destroying the session and sending 401 response")

		err = ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx)
		if err != nil {
			ctx.Logger.Errorf("Unable to destroy user session after handler could not match them to their %s header: %s", headerSessionUsername, err)
		}

		err = fmt.Errorf("could not match user %s to their %s header with a value of %s when visiting %s", username, headerSessionUsername, sessionUsername, targetURL.String())
	}

	return
}

// VerifyGET returns the handler verifying if a request is allowed to go through.
func VerifyGET(cfg schema.AuthenticationBackendConfiguration) middlewares.RequestHandler {
	refreshProfile, refreshProfileInterval := getProfileRefreshSettings(cfg)

	return func(ctx *middlewares.AutheliaCtx) {
		ctx.Logger.Tracef("Headers=%s", ctx.Request.Header.String())
		targetURL, err := ctx.GetOriginalURL()

		if err != nil {
			ctx.Logger.Errorf("Unable to parse target URL: %s", err)
			ctx.ReplyUnauthorized()

			return
		}

		if !isSchemeHTTPS(targetURL) && !isSchemeWSS(targetURL) {
			ctx.Logger.Errorf("Scheme of target URL %s must be secure since cookies are "+
				"only transported over a secure connection for security reasons", targetURL.String())
			ctx.ReplyUnauthorized()

			return
		}

		if !isURLUnderProtectedDomain(targetURL, ctx.Configuration.Session.Domain) {
			ctx.Logger.Errorf("Target URL %s is not under the protected domain %s",
				targetURL.String(), ctx.Configuration.Session.Domain)
			ctx.ReplyUnauthorized()

			return
		}

		method := ctx.XForwardedMethod()
		isBasicAuth, username, name, groups, emails, authLevel, err := verifyAuth(ctx, targetURL, refreshProfile, refreshProfileInterval)

		if err != nil {
			ctx.Logger.Errorf("Error caught when verifying user authorization: %s", err)

			if err := updateActivityTimestamp(ctx, isBasicAuth, username); err != nil {
				ctx.Error(fmt.Errorf("unable to update last activity: %s", err), messageOperationFailed)
				return
			}

			handleUnauthorized(ctx, targetURL, isBasicAuth, username, method)

			return
		}

		authorized := isTargetURLAuthorized(ctx.Providers.Authorizer, *targetURL, username,
			groups, ctx.RemoteIP(), method, authLevel)

		switch authorized {
		case Forbidden:
			ctx.Logger.Infof("Access to %s is forbidden to user %s", targetURL.String(), username)
			ctx.ReplyForbidden()
		case NotAuthorized:
			handleUnauthorized(ctx, targetURL, isBasicAuth, username, method)
		case Authorized:
			setForwardedHeaders(&ctx.Response.Header, username, name, groups, emails)
		}

		if err := updateActivityTimestamp(ctx, isBasicAuth, username); err != nil {
			ctx.Error(fmt.Errorf("unable to update last activity: %s", err), messageOperationFailed)
		}
	}
}
