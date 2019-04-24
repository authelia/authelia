package handlers

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/clems4ever/authelia/authentication"
	"github.com/clems4ever/authelia/authorization"
	"github.com/clems4ever/authelia/middlewares"
	"github.com/valyala/fasthttp"
)

// getOriginalURL extract the URL from the request headers (X-Original-URI or X-Forwarded-* headers).
func getOriginalURL(ctx *middlewares.AutheliaCtx) (*url.URL, error) {
	originalURL := ctx.XOriginalURL()
	if originalURL != nil {
		url, err := url.ParseRequestURI(string(originalURL))
		if err != nil {
			return nil, err
		}
		return url, nil
	}

	forwardedProto := ctx.XForwardedProto()
	forwardedHost := ctx.XForwardedHost()
	forwardedURI := ctx.XForwardedURI()

	if forwardedProto == nil || forwardedHost == nil {
		return nil, errMissingHeadersForTargetURL
	}

	var requestURI string
	scheme := append(forwardedProto, protoHostSeparator...)
	if forwardedURI == nil {
		requestURI = string(append(scheme, forwardedHost...))
	}

	requestURI = string(append(scheme,
		append(forwardedHost, forwardedURI...)...))

	url, err := url.ParseRequestURI(string(requestURI))
	if err != nil {
		return nil, err
	}
	return url, nil
}

// parseBasicAuth parses an HTTP Basic Authentication string.
// "Basic QWxhZGRpbjpvcGVuIHNlc2FtZQ==" returns ("Aladdin", "open sesame", true).
func parseBasicAuth(auth string) (username, password string, err error) {
	if !strings.HasPrefix(auth, authPrefix) {
		return "", "", fmt.Errorf("%s prefix not found in authorization header", strings.Trim(authPrefix, " "))
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(authPrefix):])
	if err != nil {
		return "", "", err
	}
	cs := string(c)
	s := strings.IndexByte(cs, ':')
	if s < 0 {
		return "", "", fmt.Errorf("Format for basic auth must be user:password")
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

	if level == authorization.Bypass {
		return Authorized
	} else if username != "" && level == authorization.Denied {
		// If the user is not anonymous, it means that we went through
		// all the rules related to that user and knowing who he is we can
		// deduce the access is forbidden.
		// For anonymous users though, we cannot be sure that she
		// could not be granted the rights to access the resource. Consequently
		// for anonymous users we send Unauthorized instead of Forbidden.
		return Forbidden
	} else {
		if level == authorization.OneFactor &&
			authLevel >= authentication.OneFactor {
			return Authorized
		} else if level == authorization.TwoFactor &&
			authLevel >= authentication.TwoFactor {
			return Authorized
		}
	}
	return NotAuthorized
}

// verifyBasicAuth verify that the provided username and password are correct and
// that the user is authorized to target the resource.
func verifyBasicAuth(auth []byte, targetURL url.URL, ctx *middlewares.AutheliaCtx) (username string, groups []string, authLevel authentication.Level, err error) {
	username, password, err := parseBasicAuth(string(auth))

	if err != nil {
		return "", nil, authentication.NotAuthenticated, fmt.Errorf("Unable to parse basic auth: %s", err)
	}

	authenticated, err := ctx.Providers.UserProvider.CheckUserPassword(username, password)

	if err != nil {
		return "", nil, authentication.NotAuthenticated, fmt.Errorf("Unable to check password in basic auth mode: %s", err)
	}

	// If the user is not correctly authenticated, send a 401.
	if !authenticated {
		// Request Basic Authentication otherwise
		return "", nil, authentication.NotAuthenticated, fmt.Errorf("User %s is not authenticated", username)
	}

	details, err := ctx.Providers.UserProvider.GetDetails(username)

	if err != nil {
		return "", nil, authentication.NotAuthenticated, fmt.Errorf("Unable to retrieve user details in basic auth mode: %s", err)
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

// hasUserBeenInactiveLongEnough check whether the user has been inactive for too long.
func hasUserBeenInactiveLongEnough(ctx *middlewares.AutheliaCtx) (bool, error) {
	expiration, err := ctx.Providers.SessionProvider.GetExpiration(ctx.RequestCtx)

	if err != nil {
		return false, err
	}

	// If the cookie has no expiration.
	if expiration == 0 {
		return false, nil
	}

	maxInactivityPeriod := ctx.Configuration.Session.Inactivity
	if maxInactivityPeriod == 0 {
		return false, nil
	}

	lastActivity := ctx.GetSession().LastActivity
	inactivityPeriod := time.Now().Unix() - lastActivity

	ctx.Logger.Debugf("Inactivity report: Inactivity=%d, MaxInactivity=%d",
		inactivityPeriod, maxInactivityPeriod)

	if inactivityPeriod > maxInactivityPeriod {
		return true, nil
	}

	return false, nil
}

// verifyFromSessionCookie verify if a user identified by a cookie is allowed to access target URL.
func verifyFromSessionCookie(targetURL url.URL, ctx *middlewares.AutheliaCtx) (username string, groups []string, authLevel authentication.Level, err error) {
	userSession := ctx.GetSession()
	// No username in the session means the user is anonymous.
	isUserAnonymous := userSession.Username == ""

	if isUserAnonymous && userSession.AuthenticationLevel != authentication.NotAuthenticated {
		return "", nil, authentication.NotAuthenticated, fmt.Errorf("An anonymous user cannot be authenticated. That might be the sign of a compromise")
	}

	if !isUserAnonymous {
		inactiveLongEnough, err := hasUserBeenInactiveLongEnough(ctx)
		if err != nil {
			return "", nil, authentication.NotAuthenticated, fmt.Errorf("Unable to check if user has been inactive for a long time: %s", err)
		}

		if inactiveLongEnough {
			// Destroy the session a new one will be regenerated on next request.
			err := ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx)
			if err != nil {
				return "", nil, authentication.NotAuthenticated, fmt.Errorf("Unable to destroy user session after long inactivity: %s", err)
			}

			return "", nil, authentication.NotAuthenticated, fmt.Errorf("User %s has been inactive for too long", userSession.Username)
		}
	}
	return userSession.Username, userSession.Groups, userSession.AuthenticationLevel, nil
}

// VerifyGet is the handler verifying if a request is allowed to go through.
func VerifyGet(ctx *middlewares.AutheliaCtx) {
	ctx.Logger.Tracef("Headers=%s", ctx.Request.Header.String())
	targetURL, err := getOriginalURL(ctx)

	if err != nil {
		ctx.Error(fmt.Errorf("Unable to parse target URL: %s", err), operationFailedMessage)
		return
	}

	var username string
	var groups []string
	var authLevel authentication.Level

	proxyAuthorization := ctx.Request.Header.Peek(authorizationHeader)
	hasBasicAuth := proxyAuthorization != nil

	if hasBasicAuth {
		username, groups, authLevel, err = verifyBasicAuth(proxyAuthorization, *targetURL, ctx)
	} else {
		username, groups, authLevel, err = verifyFromSessionCookie(*targetURL, ctx)
	}

	if err != nil {
		ctx.Logger.Error(fmt.Sprintf("Error caught when verifying user authorization: %s", err))
		ctx.ReplyUnauthorized()
		return
	}

	authorization := isTargetURLAuthorized(ctx.Providers.Authorizer, *targetURL, username,
		groups, ctx.RemoteIP(), authLevel)

	if authorization == Forbidden {
		ctx.ReplyForbidden()
		ctx.Logger.Errorf("Access to %s is forbidden to user %s", targetURL.String(), username)
		return
	} else if authorization == NotAuthorized {
		// Kubernetes ingress controller and Traefik use the rd parameter of the verify
		// endpoint to provide the URL of the login portal. The target URL of the user
		// is computed from X-Fowarded-* headers or X-Original-URL.
		rd := string(ctx.QueryArgs().Peek("rd"))
		if rd != "" {
			redirectionURL := fmt.Sprintf("%s?rd=%s", rd, targetURL.String())
			ctx.Redirect(redirectionURL, 302)
			ctx.SetBodyString(fmt.Sprintf("Found. Redirecting to %s", redirectionURL))
		} else {
			ctx.ReplyUnauthorized()
			ctx.Logger.Errorf("Access to %s is not authorized to user %s", targetURL.String(), username)
		}
	} else if authorization == Authorized {
		setForwardedHeaders(&ctx.Response.Header, username, groups)
	}

	// We mark activity of the current user if he comes with a session cookie.
	if !hasBasicAuth && username != "" {
		// Mark current activity
		userSession := ctx.GetSession()
		userSession.LastActivity = time.Now().Unix()
		err = ctx.SaveSession(userSession)

		if err != nil {
			ctx.Error(fmt.Errorf("Unable to update last activity: %s", err), operationFailedMessage)
			return
		}
	}
}
