package handlers

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/middlewares"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/utils"
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

// getOriginalURL extract the URL from the request headers (X-Original-URI or X-Forwarded-* headers).
func getOriginalURL(ctx *middlewares.AutheliaCtx) (*url.URL, error) {
	originalURL := ctx.XOriginalURL()
	if originalURL != nil {
		url, err := url.ParseRequestURI(string(originalURL))
		if err != nil {
			return nil, fmt.Errorf("Unable to parse URL extracted from X-Original-URL header: %v", err)
		}

		ctx.Logger.Trace("Using X-Original-URL header content as targeted site URL")

		return url, nil
	}

	forwardedProto := ctx.XForwardedProto()
	forwardedHost := ctx.XForwardedHost()
	forwardedURI := ctx.XForwardedURI()

	if forwardedProto == nil {
		return nil, errMissingXForwardedProto
	}

	if forwardedHost == nil {
		return nil, errMissingXForwardedHost
	}

	var requestURI string

	scheme := append(forwardedProto, protoHostSeparator...)
	requestURI = string(append(scheme,
		append(forwardedHost, forwardedURI...)...))

	url, err := url.ParseRequestURI(requestURI)
	if err != nil {
		return nil, fmt.Errorf("Unable to parse URL %s: %v", requestURI, err)
	}

	ctx.Logger.Tracef("Using X-Forwarded-Proto, X-Forwarded-Host and X-Forwarded-URI headers " +
		"to construct targeted site URL")

	return url, nil
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, err error) {
	if !strings.HasPrefix(auth, authPrefix) {
		return "", "", fmt.Errorf("%s prefix not found in %s header", strings.Trim(authPrefix, " "), AuthorizationHeader)
	}

	c, err := base64.StdEncoding.DecodeString(auth[len(authPrefix):])
	if err != nil {
		return "", "", err
	}

	cs := string(c)
	s := strings.IndexByte(cs, ':')

	if s < 0 {
		return "", "", fmt.Errorf("Format of %s header must be user:password", AuthorizationHeader)
	}

	return cs[:s], cs[s+1:], nil
}

// isTargetURLAuthorized check whether the given user is authorized to access the resource.
func isTargetURLAuthorized(authorizer *authorization.Authorizer, targetURL url.URL,
	username string, userGroups []string, clientIP net.IP, authLevel authentication.Level) authorizationMatching {
	level := authorizer.GetRequiredLevel(authorization.Subject{
		Username: username,
		Groups:   userGroups,
		IP:       clientIP,
	}, targetURL)

	switch {
	case level == authorization.Bypass:
		return Authorized
	case level == authorization.Denied && username != "":
		// If the user is not anonymous, it means that we went through
		// all the rules related to that user and knowing who he is we can
		// deduce the access is forbidden
		// For anonymous users though, we cannot be sure that she
		// could not be granted the rights to access the resource. Consequently
		// for anonymous users we send Unauthorized instead of Forbidden
		return Forbidden
	case level == authorization.OneFactor && authLevel >= authentication.OneFactor,
		level == authorization.TwoFactor && authLevel >= authentication.TwoFactor:
		return Authorized
	}

	return NotAuthorized
}

// verifyBasicAuth verify that the provided username and password are correct and
// that the user is authorized to target the resource.
func verifyBasicAuth(auth []byte, targetURL url.URL, ctx *middlewares.AutheliaCtx) (username string, groups []string, authLevel authentication.Level, err error) { //nolint:unparam
	username, password, err := parseBasicAuth(string(auth))

	if err != nil {
		return "", nil, authentication.NotAuthenticated, fmt.Errorf("Unable to parse content of %s header: %s", AuthorizationHeader, err)
	}

	authenticated, err := ctx.Providers.UserProvider.CheckUserPassword(username, password)

	if err != nil {
		return "", nil, authentication.NotAuthenticated, fmt.Errorf("Unable to check credentials extracted from %s header: %s", AuthorizationHeader, err)
	}

	// If the user is not correctly authenticated, send a 401.
	if !authenticated {
		// Request Basic Authentication otherwise
		return "", nil, authentication.NotAuthenticated, fmt.Errorf("User %s is not authenticated", username)
	}

	details, err := ctx.Providers.UserProvider.GetDetails(username)

	if err != nil {
		return "", nil, authentication.NotAuthenticated, fmt.Errorf("Unable to retrieve details of user %s: %s", username, err)
	}

	return username, details.Groups, authentication.OneFactor, nil
}

// setForwardedHeaders set the forwarded User and Groups headers.
func setForwardedHeaders(headers *fasthttp.ResponseHeader, username string, groups []string) {
	if username != "" {
		headers.Set(remoteUserHeader, username)
		headers.Set(remoteGroupsHeader, strings.Join(groups, ","))
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
	refreshProfileInterval time.Duration) (username string, groups []string, authLevel authentication.Level, err error) {
	// No username in the session means the user is anonymous.
	isUserAnonymous := userSession.Username == ""

	if isUserAnonymous && userSession.AuthenticationLevel != authentication.NotAuthenticated {
		return "", nil, authentication.NotAuthenticated, fmt.Errorf("An anonymous user cannot be authenticated. That might be the sign of a compromise")
	}

	if !userSession.KeepMeLoggedIn && !isUserAnonymous {
		inactiveLongEnough, err := hasUserBeenInactiveTooLong(ctx)
		if err != nil {
			return "", nil, authentication.NotAuthenticated, fmt.Errorf("Unable to check if user has been inactive for a long time: %s", err)
		}

		if inactiveLongEnough {
			// Destroy the session a new one will be regenerated on next request.
			err := ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx)
			if err != nil {
				return "", nil, authentication.NotAuthenticated, fmt.Errorf("Unable to destroy user session after long inactivity: %s", err)
			}

			return userSession.Username, userSession.Groups, authentication.NotAuthenticated, fmt.Errorf("User %s has been inactive for too long", userSession.Username)
		}
	}

	err = verifySessionHasUpToDateProfile(ctx, targetURL, userSession, refreshProfile, refreshProfileInterval)
	if err != nil {
		if err == authentication.ErrUserNotFound {
			err = ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx)
			if err != nil {
				ctx.Logger.Error(fmt.Errorf("Unable to destroy user session after provider refresh didn't find the user: %s", err))
			}

			return userSession.Username, userSession.Groups, authentication.NotAuthenticated, err
		}

		ctx.Logger.Warnf("Error occurred while attempting to update user details from LDAP: %s", err)
	}

	return userSession.Username, userSession.Groups, userSession.AuthenticationLevel, nil
}

func handleUnauthorized(ctx *middlewares.AutheliaCtx, targetURL fmt.Stringer, username string) {
	// Kubernetes ingress controller and Traefik use the rd parameter of the verify
	// endpoint to provide the URL of the login portal. The target URL of the user
	// is computed from X-Forwarded-* headers or X-Original-URL.
	rd := string(ctx.QueryArgs().Peek("rd"))
	if rd != "" {
		redirectionURL := fmt.Sprintf("%s?rd=%s", rd, url.QueryEscape(targetURL.String()))
		if strings.Contains(redirectionURL, "/%23/") {
			ctx.Logger.Warn("Characters /%23/ have been detected in redirection URL. This is not needed anymore, please strip it")
		}

		ctx.Logger.Infof("Access to %s is not authorized to user %s, redirecting to %s", targetURL.String(), username, redirectionURL)
		ctx.Redirect(redirectionURL, 302)
		ctx.SetBodyString(fmt.Sprintf("Found. Redirecting to %s", redirectionURL))
	} else {
		ctx.Logger.Infof("Access to %s is not authorized to user %s, sending 401 response", targetURL.String(), username)
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
		groupsDelta = append(groupsDelta, fmt.Sprintf("Added: %s.", strings.Join(groupsAdded, ", ")))
	}

	if len(groupsRemoved) != 0 {
		groupsDelta = append(groupsDelta, fmt.Sprintf("Removed: %s.", strings.Join(groupsRemoved, ", ")))
	}

	if len(groupsDelta) != 0 {
		ctx.Logger.Tracef("Updated groups detected for %s. %s", userSession.Username, strings.Join(groupsDelta, " "))
	} else {
		ctx.Logger.Tracef("No updated groups detected for %s", userSession.Username)
	}

	// Check Emails.
	var emailsDelta []string
	if len(emailsAdded) != 0 {
		emailsDelta = append(emailsDelta, fmt.Sprintf("Added: %s.", strings.Join(emailsAdded, ", ")))
	}

	if len(emailsRemoved) != 0 {
		emailsDelta = append(emailsDelta, fmt.Sprintf("Removed: %s.", strings.Join(emailsRemoved, ", ")))
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
	// See https://docs.authelia.com/security/threat-model.html#potential-future-guarantees
	ctx.Logger.Tracef("Checking if we need check the authentication backend for an updated profile for %s.", userSession.Username)

	if refreshProfile && userSession.Username != "" && targetURL != nil &&
		ctx.Providers.Authorizer.IsURLMatchingRuleWithGroupSubjects(*targetURL) &&
		(refreshProfileInterval == schema.RefreshIntervalAlways || userSession.RefreshTTL.Before(ctx.Clock.Now())) {
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
			// Only update TTL if the user has a interval set.
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
			if ctx.Configuration.LogLevel == "trace" {
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
	}

	// Return nil if disabled or if no changes and refresh interval set to always.
	return nil
}

func getProfileRefreshSettings(cfg schema.AuthenticationBackendConfiguration) (refresh bool, refreshInterval time.Duration) {
	if cfg.Ldap != nil {
		if cfg.RefreshInterval != schema.ProfileRefreshDisabled {
			refresh = true

			if cfg.RefreshInterval != schema.ProfileRefreshAlways {
				// Skip Error Check since validator checks it
				refreshInterval, _ = utils.ParseDurationString(cfg.RefreshInterval)
			} else {
				refreshInterval = schema.RefreshIntervalAlways
			}
		}
	}

	return refresh, refreshInterval
}

// VerifyGet returns the handler verifying if a request is allowed to go through.
func VerifyGet(cfg schema.AuthenticationBackendConfiguration) middlewares.RequestHandler {
	refreshProfile, refreshProfileInterval := getProfileRefreshSettings(cfg)

	return func(ctx *middlewares.AutheliaCtx) {
		ctx.Logger.Tracef("Headers=%s", ctx.Request.Header.String())
		targetURL, err := getOriginalURL(ctx)

		if err != nil {
			ctx.Error(fmt.Errorf("Unable to parse target URL: %s", err), operationFailedMessage)
			return
		}

		if !isSchemeHTTPS(targetURL) && !isSchemeWSS(targetURL) {
			ctx.Logger.Error(fmt.Errorf("Scheme of target URL %s must be secure since cookies are "+
				"only transported over a secure connection for security reasons", targetURL.String()))
			ctx.ReplyUnauthorized()

			return
		}

		if !isURLUnderProtectedDomain(targetURL, ctx.Configuration.Session.Domain) {
			ctx.Logger.Error(fmt.Errorf("The target URL %s is not under the protected domain %s",
				targetURL.String(), ctx.Configuration.Session.Domain))
			ctx.ReplyUnauthorized()

			return
		}

		var username string

		var groups []string

		var authLevel authentication.Level

		proxyAuthorization := ctx.Request.Header.Peek(AuthorizationHeader)
		isBasicAuth := proxyAuthorization != nil
		userSession := ctx.GetSession()

		if isBasicAuth {
			username, groups, authLevel, err = verifyBasicAuth(proxyAuthorization, *targetURL, ctx)
		} else {
			username, groups, authLevel, err = verifySessionCookie(ctx, targetURL, &userSession,
				refreshProfile, refreshProfileInterval)
		}

		if err != nil {
			ctx.Logger.Error(fmt.Sprintf("Error caught when verifying user authorization: %s", err))

			if err := updateActivityTimestamp(ctx, isBasicAuth, username); err != nil {
				ctx.Error(fmt.Errorf("Unable to update last activity: %s", err), operationFailedMessage)
				return
			}

			handleUnauthorized(ctx, targetURL, username)

			return
		}

		authorization := isTargetURLAuthorized(ctx.Providers.Authorizer, *targetURL, username,
			groups, ctx.RemoteIP(), authLevel)

		switch authorization {
		case Forbidden:
			ctx.Logger.Infof("Access to %s is forbidden to user %s", targetURL.String(), username)
			ctx.ReplyForbidden()
		case NotAuthorized:
			handleUnauthorized(ctx, targetURL, username)
		case Authorized:
			setForwardedHeaders(&ctx.Response.Header, username, groups)
		}

		if err := updateActivityTimestamp(ctx, isBasicAuth, username); err != nil {
			ctx.Error(fmt.Errorf("Unable to update last activity: %s", err), operationFailedMessage)
		}
	}
}
