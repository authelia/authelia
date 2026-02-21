package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/regulation"
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
func NewHeaderAuthorizationAuthnStrategy(schemaBasicCacheLifeSpan time.Duration, schemes ...string) *HeaderAuthnStrategy {
	return &HeaderAuthnStrategy{
		authn:              AuthnTypeAuthorization,
		headerAuthorize:    headerAuthorization,
		headerAuthenticate: headerWWWAuthenticate,
		handleAuthenticate: true,
		statusAuthenticate: fasthttp.StatusUnauthorized,
		schemes:            model.NewAuthorizationSchemes(schemes...),
		basic:              NewBasicAuthHandler(schemaBasicCacheLifeSpan),
	}
}

// NewHeaderProxyAuthorizationAuthnStrategy creates a new HeaderAuthnStrategy using the Proxy-Authorization and
// Proxy-Authenticate headers, and the 407 Proxy Auth Required response.
func NewHeaderProxyAuthorizationAuthnStrategy(schemaBasicCacheLifeSpan time.Duration, schemes ...string) *HeaderAuthnStrategy {
	return &HeaderAuthnStrategy{
		authn:              AuthnTypeProxyAuthorization,
		headerAuthorize:    headerProxyAuthorization,
		headerAuthenticate: headerProxyAuthenticate,
		handleAuthenticate: true,
		statusAuthenticate: fasthttp.StatusProxyAuthRequired,
		schemes:            model.NewAuthorizationSchemes(schemes...),
		basic:              NewBasicAuthHandler(schemaBasicCacheLifeSpan),
	}
}

// NewHeaderProxyAuthorizationAuthRequestAuthnStrategy creates a new HeaderAuthnStrategy using the Proxy-Authorization
// and WWW-Authenticate headers, and the 401 Proxy Auth Required response. This is a special AuthnStrategy for the
// AuthRequest implementation.
func NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(schemaBasicCacheLifeSpan time.Duration, schemes ...string) *HeaderAuthnStrategy {
	return &HeaderAuthnStrategy{
		authn:              AuthnTypeProxyAuthorization,
		headerAuthorize:    headerProxyAuthorization,
		headerAuthenticate: headerWWWAuthenticate,
		handleAuthenticate: true,
		statusAuthenticate: fasthttp.StatusUnauthorized,
		schemes:            model.NewAuthorizationSchemes(schemes...),
		basic:              NewBasicAuthHandler(schemaBasicCacheLifeSpan),
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
func (s *CookieSessionAuthnStrategy) Get(ctx AuthzContext, _ *authorization.Object) (authn *Authn, err error) {
	var userSession session.UserSession

	authn = &Authn{
		Type:     AuthnTypeCookie,
		Level:    authentication.NotAuthenticated,
		Username: anonymous,
	}

	if userSession, err = ctx.GetSession(); err != nil {
		return authn, fmt.Errorf("failed to retrieve user session: %w", err)
	}

	if userSession.CookieDomain != ctx.GetSessionConfig().Domain {
		ctx.GetLogger().Warnf("Destroying session cookie as the cookie domain '%s' does not match the requests detected cookie domain '%s' which may be a sign a user tried to move this cookie from one domain to another", userSession.CookieDomain, ctx.GetSessionConfig().Domain)

		if err = ctx.DestroySession(); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred trying to destroy the session cookie")
		}

		userSession = ctx.NewSession()

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred trying to save the new session cookie")
		}
	}

	if modified, invalid := handleAuthnCookieValidate(ctx, &userSession, s.refresh); invalid {
		if err = ctx.DestroySession(); err != nil {
			ctx.GetLogger().WithError(err).Errorf("Unable to destroy user session")
		}

		userSession = ctx.NewSession()
		userSession.LastActivity = ctx.GetClock().Now().Unix()

		if err = ctx.SaveSession(userSession); err != nil {
			ctx.GetLogger().WithError(err).Error("Unable to save updated user session")
		}

		return authn, nil
	} else if modified {
		if err = ctx.SaveSession(userSession); err != nil {
			ctx.GetLogger().WithError(err).Error("Unable to save updated user session")
		}
	}

	return &Authn{
		Username: friendlyUsername(userSession.Username),
		Details: authentication.UserDetails{
			Username:    userSession.Username,
			DisplayName: userSession.DisplayName,
			Emails:      userSession.Emails,
			Groups:      userSession.Groups,
		},
		Level: userSession.AuthenticationLevel(ctx.GetConfiguration().WebAuthn.EnablePasskey2FA),
		Type:  AuthnTypeCookie,
	}, nil
}

// CanHandleUnauthorized returns true if this AuthnStrategy should handle Unauthorized requests.
func (s *CookieSessionAuthnStrategy) CanHandleUnauthorized() (handle bool) {
	return false
}

// HeaderStrategy returns true if this AuthnStrategy is header based.
func (s *CookieSessionAuthnStrategy) HeaderStrategy() (header bool) {
	return false
}

// HandleUnauthorized is the Unauthorized handler for the cookie AuthnStrategy.
func (s *CookieSessionAuthnStrategy) HandleUnauthorized(_ AuthzContext, _ *Authn, _ *url.URL) {
}

// HeaderAuthnStrategy is a header AuthnStrategy.
type HeaderAuthnStrategy struct {
	authn              AuthnType
	headerAuthorize    []byte
	headerAuthenticate []byte
	handleAuthenticate bool
	statusAuthenticate int
	schemes            model.AuthorizationSchemes

	basic BasicAuthHandler
}

// BasicAuthHandler is a function signature that handles basic authentication. This is used to implement caching.
type BasicAuthHandler func(ctx AuthzContext, authorization *model.Authorization) (valid, cached bool, err error)

// NewBasicAuthHandler creates a new BasicAuthHandler depending on the lifespan.
func NewBasicAuthHandler(lifespan time.Duration) BasicAuthHandler {
	if lifespan == 0 {
		return DefaultBasicAuthHandler
	}

	return NewCachedBasicAuthHandler(lifespan)
}

// DefaultBasicAuthHandler is a BasicAuthHandler that just checks the username and password directly.
func DefaultBasicAuthHandler(ctx AuthzContext, authorization *model.Authorization) (valid, cached bool, err error) {
	valid, err = ctx.GetProviders().UserProvider.CheckUserPassword(authorization.Basic())

	return valid, false, err
}

// NewCachedBasicAuthHandler creates a new BasicAuthHandler which uses the authentication.NewCredentialCacheHMAC using
// the sha256 checksum functions.
func NewCachedBasicAuthHandler(lifespan time.Duration) BasicAuthHandler {
	cache := authentication.NewCredentialCacheHMAC(sha256.New, lifespan)

	return func(ctx AuthzContext, authorization *model.Authorization) (valid, cached bool, err error) {
		username, password := authorization.Basic()

		return cache.Check(ctx, username, password)
	}
}

// Get returns the Authn information for this AuthnStrategy.
func (s *HeaderAuthnStrategy) Get(ctx AuthzContext, object *authorization.Object) (authn *Authn, err error) {
	var value []byte

	authn = &Authn{
		Type:     s.authn,
		Level:    authentication.NotAuthenticated,
		Username: anonymous,
	}

	if value = ctx.GetRequestHeaderValue(s.headerAuthorize); len(value) == 0 {
		return authn, nil
	}

	authz := model.NewAuthorization()

	if err = authz.ParseBytes(value); err != nil {
		return authn, fmt.Errorf("failed to parse content of %s header: %w", s.headerAuthorize, err)
	}

	authn.Header.Authorization = authz

	var (
		username, clientID string

		ccs   bool
		level authentication.Level
	)

	scheme := authn.Header.Authorization.Scheme()

	if !s.schemes.Has(scheme) {
		ctx.GetLogger().
			WithFields(map[string]any{"scheme": authn.Header.Authorization.SchemeRaw(), "header": string(s.headerAuthorize)}).
			Debug("Skipping header authorization as the scheme and header combination is unknown to this endpoint configuration")

		return authn, nil
	}

	switch scheme {
	case model.AuthorizationSchemeBasic:
		username, level, err = s.handleGetBasic(ctx, authn, object)
	case model.AuthorizationSchemeBearer:
		username, clientID, ccs, level, err = handleVerifyGETAuthorizationBearer(ctx, authn, object)
	default:
		ctx.GetLogger().
			WithFields(map[string]any{"scheme": authn.Header.Authorization.SchemeRaw(), "header": string(s.headerAuthorize)}).
			Debug("Skipping header authorization as the scheme is unknown to this endpoint configuration")

		return authn, nil
	}

	if err != nil {
		if errors.Is(err, errTokenIntent) {
			return authn, nil
		}

		return authn, fmt.Errorf("failed to validate %s header with %s scheme: %w", s.headerAuthorize, scheme, err)
	}

	switch {
	case ccs:
		if len(clientID) == 0 {
			return authn, fmt.Errorf("failed to determine client id from the %s header", s.headerAuthorize)
		}

		authn.ClientID = clientID
	case len(username) == 0:
		return authn, fmt.Errorf("failed to determine username from the %s header", s.headerAuthorize)
	default:
		var details *authentication.UserDetails

		if details, err = ctx.GetProviders().UserProvider.GetDetails(username); err != nil {
			if errors.Is(err, authentication.ErrUserNotFound) {
				ctx.GetLogger().WithField("username", username).Error("Error occurred while attempting to get user details for user: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login")

				return authn, err
			}

			return authn, fmt.Errorf("unable to retrieve details for user '%s': %w", username, err)
		}

		authn.Username = friendlyUsername(details.Username)
		authn.Details = *details
	}

	authn.Level = level

	return authn, nil
}

func (s *HeaderAuthnStrategy) handleGetBasic(ctx AuthzContext, authn *Authn, object *authorization.Object) (username string, level authentication.Level, err error) {
	var (
		ban     regulation.BanType
		value   string
		expires *time.Time
	)

	username = authn.Header.Authorization.BasicUsername()

	if ban, value, expires, err = ctx.GetProviders().Regulator.BanCheck(ctx, username); err != nil {
		if errors.Is(err, regulation.ErrUserIsBanned) {
			doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(ban, value, expires), regulation.AuthType1FA, object.String(), object.Method, nil)

			return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate the credentials of user '%s' parsed from the %s header: %w", username, s.headerAuthorize, err)
		}

		ctx.GetLogger().WithError(err).Errorf(logFmtErrRegulationFail, regulation.AuthType1FA, username)

		return "", authentication.NotAuthenticated, fmt.Errorf("failed to check the regulation status of user '%s' during an attempt to authenticate using the %s header: %w", username, s.headerAuthorize, err)
	}

	var valid, cached bool

	if valid, cached, err = s.basic(ctx, authn.Header.Authorization); err != nil {
		if isRegulatorSkippedErr(err) {
			ctx.GetLogger().WithError(err).Errorf("Unsuccessful %s authentication attempt by user '%s'", regulation.AuthType1FA, authn.Header.Authorization.BasicUsername())
		} else {
			doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(regulation.BanTypeNone, username, nil), regulation.AuthType1FA, object.String(), object.Method, err)
		}

		return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate the credentials of user '%s' parsed from the %s header: %w", username, s.headerAuthorize, err)
	}

	if !valid {
		doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(regulation.BanTypeNone, username, nil), regulation.AuthType1FA, object.String(), object.Method, nil)

		return "", authentication.NotAuthenticated, fmt.Errorf("failed to validate parsed credentials of %s header valid for user '%s': the username and password do not match", s.headerAuthorize, username)
	}

	if !cached {
		doMarkAuthenticationAttemptWithRequest(ctx, true, regulation.NewBan(regulation.BanTypeNone, username, nil), regulation.AuthType1FA, object.String(), object.Method, nil)
	}

	return username, authentication.OneFactor, nil
}

// CanHandleUnauthorized returns true if this AuthnStrategy should handle Unauthorized requests.
func (s *HeaderAuthnStrategy) CanHandleUnauthorized() (handle bool) {
	return s.handleAuthenticate
}

// HeaderStrategy returns true if this AuthnStrategy is header based.
func (s *HeaderAuthnStrategy) HeaderStrategy() (header bool) {
	return true
}

// HandleUnauthorized is the Unauthorized handler for the header AuthnStrategy.
func (s *HeaderAuthnStrategy) HandleUnauthorized(ctx AuthzContext, authn *Authn, _ *url.URL) {
	ctx.GetLogger().Debugf("Responding %d %s", s.statusAuthenticate, s.headerAuthenticate)

	ctx.ReplyStatusCode(s.statusAuthenticate)

	if authn.Header.Authorization != nil && authn.Header.Authorization.Scheme() == model.AuthorizationSchemeBearer && authn.Header.Error != nil {
		ctx.SetResponseHeaderValue(s.headerAuthenticate, fmt.Sprintf(`Bearer %s`, oidc.RFC6750Header(authn.Header.Realm, authn.Header.Scope, authn.Header.Error)))
	} else if s.headerAuthenticate != nil {
		ctx.SetResponseHeaderValueBytes(s.headerAuthenticate, headerValueAuthenticateBasic)
	}
}

// HeaderLegacyAuthnStrategy is a legacy header AuthnStrategy which can be switched based on the query parameters.
type HeaderLegacyAuthnStrategy struct{}

// Get returns the Authn information for this AuthnStrategy.
func (s *HeaderLegacyAuthnStrategy) Get(ctx AuthzContext, _ *authorization.Object) (authn *Authn, err error) {
	var (
		username, password string
		value, header      []byte
	)

	authn = &Authn{
		Level:    authentication.NotAuthenticated,
		Username: anonymous,
	}

	if qryValueAuth := ctx.GetRequestQueryArgValue(qryArgAuth); bytes.Equal(qryValueAuth, qryValueBasic) {
		authn.Type = AuthnTypeAuthorization
		header = headerAuthorization
	} else {
		authn.Type = AuthnTypeProxyAuthorization
		header = headerProxyAuthorization
	}

	value = ctx.GetRequestHeaderValue(header)

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

	if valid, err = ctx.GetProviders().UserProvider.CheckUserPassword(username, password); err != nil {
		return authn, fmt.Errorf("failed to validate parsed credentials of %s header for user '%s': %w", header, username, err)
	}

	if !valid {
		return authn, fmt.Errorf("validated parsed credentials of %s header but they are not valid for user '%s': %w", header, username, err)
	}

	if details, err = ctx.GetProviders().UserProvider.GetDetails(username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			ctx.GetLogger().WithField("username", username).Error("Error occurred while attempting to get user details for user: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login")

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

// HeaderStrategy returns true if this AuthnStrategy is header based.
func (s *HeaderLegacyAuthnStrategy) HeaderStrategy() (header bool) {
	return true
}

// HandleUnauthorized is the Unauthorized handler for the Legacy header AuthnStrategy.
func (s *HeaderLegacyAuthnStrategy) HandleUnauthorized(ctx AuthzContext, authn *Authn, _ *url.URL) {
	handleAuthzUnauthorizedAuthorizationBasic(ctx, authn)
}

func handleAuthnCookieValidate(ctx AuthzContext, userSession *session.UserSession, refresh schema.RefreshIntervalDuration) (modified, invalid bool) {
	// TODO: Remove this check as it's no longer possible i.e. ineffectual.
	isAnonymous := userSession.Username == ""

	if isAnonymous && userSession.AuthenticationLevel(ctx.GetConfiguration().WebAuthn.EnablePasskey2FA) != authentication.NotAuthenticated {
		ctx.GetLogger().WithFields(map[string]any{"username": anonymous, "level": userSession.AuthenticationLevel(ctx.GetConfiguration().WebAuthn.EnablePasskey2FA).String()}).Errorf("Session for user has an invalid authentication level: this may be a sign of a compromise")

		return modified, true
	}

	if invalid = handleAuthnCookieValidateInactivity(ctx, userSession, isAnonymous); invalid {
		ctx.GetLogger().WithField("username", userSession.Username).Info("Session for user not marked as remembered has exceeded configured session inactivity")

		return modified, true
	}

	if modified, invalid = handleSessionValidateRefresh(ctx, userSession, refresh); invalid {
		return modified, true
	}

	if username := ctx.GetRequestHeaderValue(headerSessionUsername); username != nil && !strings.EqualFold(string(username), userSession.Username) {
		ctx.GetLogger().WithField("username", userSession.Username).Warnf("Session for user does not match the Session-Username header with value '%s' which could be a sign of a cookie hijack", username)

		return modified, true
	}

	if !userSession.KeepMeLoggedIn {
		modified = true

		userSession.LastActivity = ctx.GetClock().Now().Unix()
	}

	return modified, false
}

func handleAuthnCookieValidateInactivity(ctx AuthzContext, userSession *session.UserSession, isAnonymous bool) (invalid bool) {
	config := ctx.GetSessionConfig()

	if isAnonymous || userSession.KeepMeLoggedIn || int64(config.Inactivity.Seconds()) == 0 {
		return false
	}

	ctx.GetLogger().WithField("username", userSession.Username).Tracef("Inactivity report for user. Current Time: %d, Last Activity: %d, Maximum Inactivity: %d.", ctx.GetClock().Now().Unix(), userSession.LastActivity, int(config.Inactivity.Seconds()))

	return time.Unix(userSession.LastActivity, 0).Add(config.Inactivity).Before(ctx.GetClock().Now())
}

func handleSessionValidateRefresh(ctx AuthzContext, userSession *session.UserSession, refresh schema.RefreshIntervalDuration) (modified, invalid bool) {
	if refresh.Never() || userSession.IsAnonymous() {
		return false, false
	}

	ctx.GetLogger().WithField("username", userSession.Username).Trace("Checking if we need check the authentication backend for an updated profile for user")

	if !refresh.Always() && userSession.RefreshTTL.After(ctx.GetClock().Now()) {
		return false, false
	}

	ctx.GetLogger().WithField("username", userSession.Username).Debug("Checking the authentication backend for an updated profile for user")

	var (
		details *authentication.UserDetails
		err     error
	)
	if details, err = ctx.GetProviders().UserProvider.GetDetails(userSession.Username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			ctx.GetLogger().WithField("username", userSession.Username).Error("Error occurred while attempting to update user details for user: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login")

			return false, true
		}

		ctx.GetLogger().WithError(err).WithField("username", userSession.Username).Error("Error occurred while attempting to update user details for user")

		return false, false
	}

	var (
		diffEmails, diffGroups, diffDisplayName bool
	)

	diffEmails, diffGroups = utils.IsStringSlicesDifferent(userSession.Emails, details.Emails), utils.IsStringSlicesDifferent(userSession.Groups, details.Groups)
	diffDisplayName = userSession.DisplayName != details.DisplayName

	if !refresh.Always() {
		modified = true

		userSession.RefreshTTL = ctx.GetClock().Now().Add(refresh.Value())
	}

	if !diffEmails && !diffGroups && !diffDisplayName {
		ctx.GetLogger().WithField("username", userSession.Username).Trace("Updated profile not detected for user")

		return modified, false
	}

	ctx.GetLogger().WithField("username", userSession.Username).Debug("Updated profile detected for user")

	if ctx.GetLogger().Level >= logrus.TraceLevel {
		generateVerifySessionHasUpToDateProfileTraceLogs(ctx, userSession, details)
	}

	userSession.Emails, userSession.Groups, userSession.DisplayName = details.Emails, details.Groups, details.DisplayName

	return true, false
}

func handleVerifyGETAuthorizationBearer(ctx AuthzContext, authn *Authn, object *authorization.Object) (username, clientID string, ccs bool, level authentication.Level, err error) {
	var at bool

	if at, err = oidc.IsAccessToken(ctx, authn.Header.Authorization.Value()); !at {
		if err != nil {
			ctx.GetLogger().WithError(err).Debug("The bearer token does not appear to be a relevant access token")
		} else {
			ctx.GetLogger().Debug("The bearer token does not appear to be a relevant access token")
		}

		return "", "", false, authentication.NotAuthenticated, errTokenIntent
	}

	return handleVerifyGETAuthorizationBearerIntrospection(ctx, ctx.GetProviders().OpenIDConnect, authn, object)
}

func handleVerifyGETAuthorizationBearerIntrospection(ctx context.Context, provider AuthzBearerIntrospectionProvider, authn *Authn, object *authorization.Object) (username, clientID string, ccs bool, level authentication.Level, err error) {
	var (
		use       oauthelia2.TokenUse
		requester oauthelia2.AccessRequester
	)

	authn.Header.Error = &oauthelia2.RFC6749Error{
		ErrorField:       "invalid_token",
		DescriptionField: "The access token is expired, revoked, malformed, or invalid for other reasons. The client can obtain a new access token and try again.",
	}
	if use, requester, err = provider.IntrospectToken(ctx, authn.Header.Authorization.Value(), oauthelia2.AccessToken, oidc.NewSession(), oidc.ScopeAutheliaBearerAuthz); err != nil {
		return "", "", false, authentication.NotAuthenticated, fmt.Errorf("error performing token introspection: %w", oauthelia2.ErrorToDebugRFC6749Error(err))
	}

	if use != oauthelia2.AccessToken {
		authn.Header.Error = oauthelia2.ErrInvalidRequest

		return "", "", false, authentication.NotAuthenticated, fmt.Errorf("token is not an access token")
	}

	audience := []string{object.URL.String()}
	strategy := provider.GetAudienceStrategy(ctx)

	if err = strategy(requester.GetGrantedAudience(), audience); err != nil {
		return "", "", false, authentication.NotAuthenticated, fmt.Errorf("token does not contain a valid audience for the url '%s' with the error: %w", audience[0], err)
	}

	fsession := requester.GetSession()

	var (
		client   oidc.Client
		osession *oidc.Session
		ok       bool
	)

	if osession, ok = fsession.(*oidc.Session); !ok {
		return "", "", false, authentication.NotAuthenticated, fmt.Errorf("introspection returned an invalid session type")
	}

	if client, err = provider.GetRegisteredClient(ctx, osession.ClientID); err != nil || client == nil {
		return "", "", false, authentication.NotAuthenticated, fmt.Errorf("client id '%s' is not registered", osession.ClientID)
	}

	if !client.GetScopes().Has(oidc.ScopeAutheliaBearerAuthz) {
		return "", "", false, authentication.NotAuthenticated, fmt.Errorf("client id '%s' is registered but does not permit the '%s' scope", osession.ClientID, oidc.ScopeAutheliaBearerAuthz)
	}

	if err = strategy(client.GetAudience(), audience); err != nil {
		return "", "", false, authentication.NotAuthenticated, fmt.Errorf("client id '%s' is registered but does not permit an audience for the url '%s' with the error: %w", osession.ClientID, audience[0], err)
	}

	if osession.DefaultSession == nil || osession.Claims == nil {
		return "", "", false, authentication.NotAuthenticated, fmt.Errorf("introspection returned a session missing required values")
	}

	authn.Header.Error = nil

	if osession.ClientCredentials {
		return "", osession.ClientID, true, authentication.OneFactor, nil
	}

	if authorization.NewAuthenticationMethodsReferencesFromClaim(osession.DefaultSession.Claims.AuthenticationMethodsReferences).MultiFactorAuthentication() {
		level = authentication.TwoFactor
	} else {
		level = authentication.OneFactor
	}

	return osession.Username, "", false, level, nil
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
