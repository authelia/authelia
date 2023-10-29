package handlers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/ory/fosite"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewCookieSessionAuthnStrategy creates a new CookieSessionAuthnStrategy.
func NewCookieSessionAuthnStrategy(refresh schema.RefreshIntervalDuration) *CookieSessionAuthnStrategy {
	return &CookieSessionAuthnStrategy{
		refresh: refresh,
	}
}

// NewHeaderAuthorizationAuthnStrategy creates a new HeaderAuthnStrategy using the Authorization and WWW-Authenticate
// headers, and the 407 Proxy Auth Required response.
func NewHeaderAuthorizationAuthnStrategy(schemes ...string) *HeaderAuthnStrategy {
	return &HeaderAuthnStrategy{
		authn:              AuthnTypeAuthorization,
		headerAuthorize:    headerAuthorization,
		headerAuthenticate: headerWWWAuthenticate,
		handleAuthenticate: true,
		statusAuthenticate: fasthttp.StatusUnauthorized,
		schemes:            model.NewAuthorizationSchemes(schemes...),
	}
}

// NewHeaderProxyAuthorizationAuthnStrategy creates a new HeaderAuthnStrategy using the Proxy-Authorization and
// Proxy-Authenticate headers, and the 407 Proxy Auth Required response.
func NewHeaderProxyAuthorizationAuthnStrategy(schemes ...string) *HeaderAuthnStrategy {
	return &HeaderAuthnStrategy{
		authn:              AuthnTypeProxyAuthorization,
		headerAuthorize:    headerProxyAuthorization,
		headerAuthenticate: headerProxyAuthenticate,
		handleAuthenticate: true,
		statusAuthenticate: fasthttp.StatusProxyAuthRequired,
		schemes:            model.NewAuthorizationSchemes(schemes...),
	}
}

// NewHeaderProxyAuthorizationAuthRequestAuthnStrategy creates a new HeaderAuthnStrategy using the Proxy-Authorization
// and WWW-Authenticate headers, and the 401 Proxy Auth Required response. This is a special AuthnStrategy for the
// AuthRequest implementation.
func NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(schemes ...string) *HeaderAuthnStrategy {
	return &HeaderAuthnStrategy{
		authn:              AuthnTypeProxyAuthorization,
		headerAuthorize:    headerProxyAuthorization,
		headerAuthenticate: headerWWWAuthenticate,
		handleAuthenticate: true,
		statusAuthenticate: fasthttp.StatusUnauthorized,
		schemes:            model.NewAuthorizationSchemes(schemes...),
	}
}

// NewHeaderLegacyAuthnStrategy creates a new HeaderLegacyAuthnStrategy.
func NewHeaderLegacyAuthnStrategy() *HeaderLegacyAuthnStrategy {
	return &HeaderLegacyAuthnStrategy{}
}

// CookieSessionAuthnStrategy is a session cookie AuthnStrategy.
type CookieSessionAuthnStrategy struct {
	refresh schema.RefreshIntervalDuration
}

// Get returns the Authn information for this AuthnStrategy.
func (s *CookieSessionAuthnStrategy) Get(ctx *middlewares.AutheliaCtx, provider *session.Session, _ *authorization.Object) (authn Authn, err error) {
	var userSession session.UserSession

	authn = Authn{
		Type:     AuthnTypeCookie,
		Level:    authentication.NotAuthenticated,
		Username: anonymous,
	}

	if userSession, err = provider.GetSession(ctx.RequestCtx); err != nil {
		return authn, fmt.Errorf("failed to retrieve user session: %w", err)
	}

	if userSession.CookieDomain != provider.Config.Domain {
		ctx.Logger.Warnf("Destroying session cookie as the cookie domain '%s' does not match the requests detected cookie domain '%s' which may be a sign a user tried to move this cookie from one domain to another", userSession.CookieDomain, provider.Config.Domain)

		if err = provider.DestroySession(ctx.RequestCtx); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred trying to destroy the session cookie")
		}

		userSession = provider.NewDefaultUserSession()

		if err = provider.SaveSession(ctx.RequestCtx, userSession); err != nil {
			ctx.Logger.WithError(err).Error("Error occurred trying to save the new session cookie")
		}
	}

	if invalid := handleVerifyGETAuthnCookieValidate(ctx, provider, &userSession, s.refresh); invalid {
		if err = ctx.DestroySession(); err != nil {
			ctx.Logger.WithError(err).Errorf("Unable to destroy user session")
		}

		userSession = provider.NewDefaultUserSession()
		userSession.LastActivity = ctx.Clock.Now().Unix()

		if err = provider.SaveSession(ctx.RequestCtx, userSession); err != nil {
			ctx.Logger.WithError(err).Error("Unable to save updated user session")
		}

		return authn, nil
	}

	if err = provider.SaveSession(ctx.RequestCtx, userSession); err != nil {
		ctx.Logger.WithError(err).Error("Unable to save updated user session")
	}

	return Authn{
		Username: friendlyUsername(userSession.Username),
		Details: authentication.UserDetails{
			Username:    userSession.Username,
			DisplayName: userSession.DisplayName,
			Emails:      userSession.Emails,
			Groups:      userSession.Groups,
		},
		Level: userSession.AuthenticationLevel,
		Type:  AuthnTypeCookie,
	}, nil
}

// CanHandleUnauthorized returns true if this AuthnStrategy should handle Unauthorized requests.
func (s *CookieSessionAuthnStrategy) CanHandleUnauthorized() (handle bool) {
	return false
}

// HandleUnauthorized is the Unauthorized handler for the cookie AuthnStrategy.
func (s *CookieSessionAuthnStrategy) HandleUnauthorized(_ *middlewares.AutheliaCtx, _ *Authn, _ *url.URL) {
}

// HeaderAuthnStrategy is a header AuthnStrategy.
type HeaderAuthnStrategy struct {
	authn              AuthnType
	headerAuthorize    []byte
	headerAuthenticate []byte
	handleAuthenticate bool
	statusAuthenticate int
	schemes            model.AuthorizationSchemes
}

// Get returns the Authn information for this AuthnStrategy.
func (s *HeaderAuthnStrategy) Get(ctx *middlewares.AutheliaCtx, _ *session.Session, object *authorization.Object) (authn Authn, err error) {
	var value []byte

	authn = Authn{
		Type:     s.authn,
		Level:    authentication.NotAuthenticated,
		Username: anonymous,
	}

	if value = ctx.Request.Header.PeekBytes(s.headerAuthorize); value == nil {
		return authn, nil
	}

	authz := model.NewAuthorization()

	if err = authz.ParseBytes(value); err != nil {
		return authn, fmt.Errorf("failed to parse content of %s header: %w", s.headerAuthorize, err)
	}

	var (
		username string
		level    authentication.Level
	)

	scheme := authz.Scheme()

	if !s.schemes.Has(scheme) {
		return authn, fmt.Errorf("invalid scheme: scheme with name '%s' isn't available on this endpoint", scheme.String())
	}

	switch scheme {
	case model.AuthorizationSchemeBasic:
		username, level, err = s.handleGetBasic(ctx, authz, object)
	case model.AuthorizationSchemeBearer:
		username, level, err = s.handleGetBearer(ctx, authz, object)
	default:
		err = fmt.Errorf("failed to parse content of %s header: the scheme '%s' is not known", s.headerAuthorize, authz.SchemeRaw())
	}

	if err != nil {
		return authn, err
	}

	if len(username) == 0 {
		return authn, fmt.Errorf("failed to determine username from the %s header", s.headerAuthorize)
	}

	var details *authentication.UserDetails

	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			ctx.Logger.WithField("username", username).Error("Error occurred while attempting to get user details for user: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login")

			return authn, err
		}

		return authn, fmt.Errorf("unable to retrieve details for user '%s': %w", username, err)
	}

	authn.Username = friendlyUsername(details.Username)
	authn.Details = *details
	authn.Level = level

	return authn, nil
}

func (s *HeaderAuthnStrategy) handleGetBasic(ctx *middlewares.AutheliaCtx, authz *model.Authorization, _ *authorization.Object) (username string, level authentication.Level, err error) {
	var (
		valid bool
	)

	if valid, err = ctx.Providers.UserProvider.CheckUserPassword(authz.Basic()); err != nil {
		return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate parsed credentials of %s header for user '%s': %w", s.headerAuthorize, authz.BasicUsername(), err)
	}

	if !valid {
		return "", authentication.NotAuthenticated, fmt.Errorf("validated parsed credentials of %s header but they are not valid for user '%s': %w", s.headerAuthorize, authz.BasicUsername(), err)
	}

	return authz.BasicUsername(), authentication.OneFactor, nil
}

func (s *HeaderAuthnStrategy) handleGetBearer(ctx *middlewares.AutheliaCtx, authz *model.Authorization, object *authorization.Object) (username string, level authentication.Level, err error) {
	if ctx.Providers.OpenIDConnect == nil || ctx.Configuration.IdentityProviders.OIDC == nil || !ctx.Configuration.IdentityProviders.OIDC.Discovery.BearerAuthorization {
		return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate %s header with bearer scheme: the bearer scheme requires an OpenID Connect 1.0 configuration but it's absent", s.headerAuthorize)
	}

	if !ctx.Configuration.IdentityProviders.OIDC.Discovery.BearerAuthorization {
		return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate %s header with bearer scheme: the bearer bearer scheme requires an OpenID Connect 1.0 client configured with the '%s' scope but there are none", s.headerAuthorize, oidc.ScopeAutheliaBearerAuthz)
	}

	use, ar, err := ctx.Providers.OpenIDConnect.IntrospectToken(ctx, authz.Value(), fosite.AccessToken, oidc.NewSession(), oidc.ScopeAutheliaBearerAuthz)
	if err != nil {
		ctx.Logger.WithError(oidc.ErrorToDebugRFC6749Error(err)).Error("Error occurred while introspecting the bearer token for authorization")

		return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate %s header with bearer scheme: token introspection failed", s.headerAuthorize)
	}

	if use != fosite.AccessToken {
		return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate %s header with bearer scheme: the token is not an access token", s.headerAuthorize)
	}

	if err = ctx.Providers.OpenIDConnect.GetAudienceStrategy(ctx)(ar.GetGrantedAudience(), []string{object.URL.String()}); err != nil {
		return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate %s header with bearer scheme: the token does not contain a valid audience for the url '%s' with the error: %w", s.headerAuthorize, object.URL, err)
	}

	fsession := ar.GetSession()

	var (
		session *oidc.Session
		ok      bool
	)

	if session, ok = fsession.(*oidc.Session); !ok {
		return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate %s header with bearer scheme: the introspection returned an invalid session type", s.headerAuthorize)
	}

	if session.DefaultSession == nil || session.DefaultSession.Claims == nil {
		return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate %s header with bearer scheme: the introspection returned a session missing required values", s.headerAuthorize)
	}

	if oidc.NewAuthenticationMethodsReferencesFromClaim(session.DefaultSession.Claims.AuthenticationMethodsReferences).MultiFactorAuthentication() {
		level = authentication.TwoFactor
	} else {
		level = authentication.OneFactor
	}

	return session.Username, level, nil
}

// CanHandleUnauthorized returns true if this AuthnStrategy should handle Unauthorized requests.
func (s *HeaderAuthnStrategy) CanHandleUnauthorized() (handle bool) {
	return s.handleAuthenticate
}

// HandleUnauthorized is the Unauthorized handler for the header AuthnStrategy.
func (s *HeaderAuthnStrategy) HandleUnauthorized(ctx *middlewares.AutheliaCtx, _ *Authn, _ *url.URL) {
	ctx.Logger.Debugf("Responding %d %s", s.statusAuthenticate, s.headerAuthenticate)

	ctx.ReplyStatusCode(s.statusAuthenticate)

	if s.headerAuthenticate != nil {
		ctx.Response.Header.SetBytesKV(s.headerAuthenticate, headerValueAuthenticateBasic)
	}
}

// HeaderLegacyAuthnStrategy is a legacy header AuthnStrategy which can be switched based on the query parameters.
type HeaderLegacyAuthnStrategy struct{}

// Get returns the Authn information for this AuthnStrategy.
func (s *HeaderLegacyAuthnStrategy) Get(ctx *middlewares.AutheliaCtx, _ *session.Session, _ *authorization.Object) (authn Authn, err error) {
	var (
		username, password string
		value, header      []byte
	)

	authn = Authn{
		Level:    authentication.NotAuthenticated,
		Username: anonymous,
	}

	if qryValueAuth := ctx.QueryArgs().PeekBytes(qryArgAuth); bytes.Equal(qryValueAuth, qryValueBasic) {
		authn.Type = AuthnTypeAuthorization
		header = headerAuthorization
	} else {
		authn.Type = AuthnTypeProxyAuthorization
		header = headerProxyAuthorization
	}

	value = ctx.Request.Header.PeekBytes(header)

	switch {
	case value == nil && authn.Type == AuthnTypeAuthorization:
		return authn, fmt.Errorf("header %s expected", headerAuthorization)
	case value == nil:
		return authn, nil
	}

	if username, password, err = headerAuthorizationParse(value); err != nil {
		return authn, fmt.Errorf("failed to parse content of %s header: %w", header, err)
	}

	if username == "" || password == "" {
		return authn, fmt.Errorf("failed to validate parsed credentials of %s header for user '%s': %w", header, username, err)
	}

	var (
		valid   bool
		details *authentication.UserDetails
	)

	if valid, err = ctx.Providers.UserProvider.CheckUserPassword(username, password); err != nil {
		return authn, fmt.Errorf("failed to validate parsed credentials of %s header for user '%s': %w", header, username, err)
	}

	if !valid {
		return authn, fmt.Errorf("validated parsed credentials of %s header but they are not valid for user '%s': %w", header, username, err)
	}

	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			ctx.Logger.WithField("username", username).Error("Error occurred while attempting to get user details for user: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login")

			return authn, err
		}

		return authn, fmt.Errorf("unable to retrieve details for user '%s': %w", username, err)
	}

	authn.Username = friendlyUsername(details.Username)
	authn.Details = *details
	authn.Level = authentication.OneFactor

	return authn, nil
}

// CanHandleUnauthorized returns true if this AuthnStrategy should handle Unauthorized requests.
func (s *HeaderLegacyAuthnStrategy) CanHandleUnauthorized() (handle bool) {
	return true
}

// HandleUnauthorized is the Unauthorized handler for the Legacy header AuthnStrategy.
func (s *HeaderLegacyAuthnStrategy) HandleUnauthorized(ctx *middlewares.AutheliaCtx, authn *Authn, _ *url.URL) {
	handleAuthzUnauthorizedAuthorizationBasic(ctx, authn)
}

func handleVerifyGETAuthnCookieValidate(ctx *middlewares.AutheliaCtx, provider *session.Session, userSession *session.UserSession, refresh schema.RefreshIntervalDuration) (invalid bool) {
	isAnonymous := userSession.Username == ""

	if isAnonymous && userSession.AuthenticationLevel != authentication.NotAuthenticated {
		ctx.Logger.WithFields(map[string]any{"username": anonymous, "level": userSession.AuthenticationLevel.String()}).Errorf("Session for user has an invalid authentication level: this may be a sign of a compromise")

		return true
	}

	if invalid = handleVerifyGETAuthnCookieValidateInactivity(ctx, provider, userSession, isAnonymous); invalid {
		ctx.Logger.WithField("username", userSession.Username).Info("Session for user not marked as remembered has exceeded configured session inactivity")

		return true
	}

	if invalid = handleVerifyGETAuthnCookieValidateRefresh(ctx, userSession, isAnonymous, refresh); invalid {
		return true
	}

	if username := ctx.Request.Header.PeekBytes(headerSessionUsername); username != nil && !strings.EqualFold(string(username), userSession.Username) {
		ctx.Logger.WithField("username", userSession.Username).Warnf("Session for user does not match the Session-Username header with value '%s' which could be a sign of a cookie hijack", username)

		return true
	}

	if !userSession.KeepMeLoggedIn {
		userSession.LastActivity = ctx.Clock.Now().Unix()
	}

	return false
}

func handleVerifyGETAuthnCookieValidateInactivity(ctx *middlewares.AutheliaCtx, provider *session.Session, userSession *session.UserSession, isAnonymous bool) (invalid bool) {
	if isAnonymous || userSession.KeepMeLoggedIn || int64(provider.Config.Inactivity.Seconds()) == 0 {
		return false
	}

	ctx.Logger.WithField("username", userSession.Username).Tracef("Inactivity report for user. Current Time: %d, Last Activity: %d, Maximum Inactivity: %d.", ctx.Clock.Now().Unix(), userSession.LastActivity, int(provider.Config.Inactivity.Seconds()))

	return time.Unix(userSession.LastActivity, 0).Add(provider.Config.Inactivity).Before(ctx.Clock.Now())
}

func handleVerifyGETAuthnCookieValidateRefresh(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, isAnonymous bool, refresh schema.RefreshIntervalDuration) (invalid bool) {
	if refresh.Never() || isAnonymous {
		return false
	}

	ctx.Logger.WithField("username", userSession.Username).Trace("Checking if we need check the authentication backend for an updated profile for user")

	if !refresh.Always() && userSession.RefreshTTL.After(ctx.Clock.Now()) {
		return false
	}

	ctx.Logger.WithField("username", userSession.Username).Debug("Checking the authentication backend for an updated profile for user")

	var (
		details *authentication.UserDetails
		err     error
	)

	if details, err = ctx.Providers.UserProvider.GetDetails(userSession.Username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			ctx.Logger.WithField("username", userSession.Username).Error("Error occurred while attempting to update user details for user: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login")

			return true
		}

		ctx.Logger.WithError(err).WithField("username", userSession.Username).Error("Error occurred while attempting to update user details for user")

		return false
	}

	var (
		diffEmails, diffGroups, diffDisplayName bool
	)

	diffEmails, diffGroups = utils.IsStringSlicesDifferent(userSession.Emails, details.Emails), utils.IsStringSlicesDifferent(userSession.Groups, details.Groups)
	diffDisplayName = userSession.DisplayName != details.DisplayName

	if !refresh.Always() {
		userSession.RefreshTTL = ctx.Clock.Now().Add(refresh.Value())
	}

	if !diffEmails && !diffGroups && !diffDisplayName {
		ctx.Logger.WithField("username", userSession.Username).Trace("Updated profile not detected for user")

		return false
	}

	ctx.Logger.WithField("username", userSession.Username).Debug("Updated profile detected for user")

	if ctx.Logger.Level >= logrus.TraceLevel {
		generateVerifySessionHasUpToDateProfileTraceLogs(ctx, userSession, details)
	}

	userSession.Emails, userSession.Groups, userSession.DisplayName = details.Emails, details.Groups, details.DisplayName

	return false
}

func headerAuthorizationParse(value []byte) (username, password string, err error) {
	if bytes.Equal(value, qryValueEmpty) {
		return "", "", fmt.Errorf("header is malformed: empty value")
	}

	parts := strings.SplitN(string(value), " ", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("header is malformed: does not appear to have a scheme")
	}

	scheme := strings.ToLower(parts[0])

	switch scheme {
	case headerAuthorizationSchemeBasic:
		if username, password, err = headerAuthorizationParseBasic(parts[1]); err != nil {
			return username, password, fmt.Errorf("header is malformed: %w", err)
		}

		return username, password, nil
	default:
		return "", "", fmt.Errorf("header is malformed: unsupported scheme '%s': supported schemes '%s'", parts[0], strings.ToTitle(headerAuthorizationSchemeBasic))
	}
}

func headerAuthorizationParseBasic(value string) (username, password string, err error) {
	var content []byte

	if content, err = base64.StdEncoding.DecodeString(value); err != nil {
		return "", "", fmt.Errorf("could not decode credentials: %w", err)
	}

	strContent := string(content)
	s := strings.IndexByte(strContent, ':')

	if s < 1 {
		return "", "", fmt.Errorf("format of header must be <user>:<password> but either doesn't have a colon or username")
	}

	return strContent[:s], strContent[s+1:], nil
}
