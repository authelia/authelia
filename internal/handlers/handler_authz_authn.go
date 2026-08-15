package handlers

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
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
		delay:              middlewares.NewTimingAttackDelay(50, time.Second*2).SetSuccessDelay(false).SetRecord(true).SetMinimumDelayDuration(time.Second * 2),
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
		delay:              middlewares.NewTimingAttackDelay(50, time.Second*2).SetSuccessDelay(false).SetRecord(true).SetMinimumDelayDuration(time.Second * 2),
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
		delay:              middlewares.NewTimingAttackDelay(50, time.Second*2).SetSuccessDelay(false).SetRecord(true).SetMinimumDelayDuration(time.Second * 2),
		basic:              NewBasicAuthHandler(schemaBasicCacheLifeSpan),
	}
}

// NewHeaderLegacyAuthnStrategy creates a new HeaderLegacyAuthnStrategy.
func NewHeaderLegacyAuthnStrategy() *HeaderLegacyAuthnStrategy {
	return &HeaderLegacyAuthnStrategy{
		delay: middlewares.NewTimingAttackDelay(50, time.Second*2).SetSuccessDelay(false).SetRecord(true).SetMinimumDelayDuration(time.Second * 2),
	}
}

// CookieSessionAuthnStrategy is a session cookie AuthnStrategy.
type CookieSessionAuthnStrategy struct {
	refresh schema.RefreshIntervalDuration
}

// Get returns the Authn information for this AuthnStrategy.
func (s *CookieSessionAuthnStrategy) Get(ctx AuthzContext, manager session.Manager, _ *authorization.Object) (authn *Authn, err error) {
	var userSession session.UserSession

	authn = &Authn{
		Type:     AuthnTypeCookie,
		Level:    authentication.NotAuthenticated,
		Username: anonymous,
		Details:  newAnonymousUserDetails(),
	}

	if userSession, err = manager.GetSession(); err != nil {
		return authn, fmt.Errorf("failed to retrieve user session: %w", err)
	}

	if userSession.CookieDomain != manager.GetSessionConfig().Domain {
		ctx.GetLogger().Warnf("Destroying session cookie as the cookie domain '%s' does not match the requests detected cookie domain '%s' which may be a sign a user tried to move this cookie from one domain to another", userSession.CookieDomain, manager.GetSessionConfig().Domain)

		if err = manager.DestroySession(); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred trying to destroy the session cookie")
		}

		userSession = manager.NewDefaultUserSession()

		if err = manager.SaveSession(&userSession); err != nil {
			ctx.GetLogger().WithError(err).Error("Error occurred trying to save the new session cookie")
		}
	}

	modified, invalid := handleAuthnCookieValidate(ctx, manager, &userSession)

	var details authentication.UserDetailsExtended

	if !invalid {
		if details, err = authentication.MustGetUserDetailsExtendedCachedSafe(userSession.Username, ctx.GetUserProvider()); err != nil {
			// Failing closed here is intentional as a defense-in-depth measure.
			if !errors.Is(err, authentication.ErrUserNotFound) {
				return authn, fmt.Errorf("failed to retrieve user details: %w", err)
			}

			ctx.GetLogger().WithField("username", userSession.Username).Error("Error occurred while attempting to update user details for user: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login")

			invalid = true
		} else if handleAuthnCookieRefreshTTL(ctx, &userSession, s.refresh) {
			modified = true
		}
	}

	if invalid {
		if err = manager.DestroySession(); err != nil {
			ctx.GetLogger().WithError(err).Errorf("Unable to destroy user session")
		}

		userSession = manager.NewDefaultUserSession()
		userSession.LastActivity = ctx.GetClock().Now().Unix()

		if err = manager.SaveSession(&userSession); err != nil {
			ctx.GetLogger().WithError(err).Error("Unable to save updated user session")
		}

		return authn, nil
	}

	if modified {
		if err = manager.SaveSession(&userSession); err != nil {
			ctx.GetLogger().WithError(err).Error("Unable to save updated user session")
		}
	}

	return &Authn{
		Username: friendlyUsername(userSession.Username),
		Details:  details,
		Level:    userSession.AuthenticationLevel(ctx.GetConfiguration().WebAuthn.EnablePasskey2FA),
		Type:     AuthnTypeCookie,
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

	delay middlewares.Delayer
	basic BasicAuthHandler
}

// BasicAuthHandler is a function signature that handles basic authentication. This is used to implement caching. The
// username must be the canonical username as resolved by the authentication backend rather than the raw value parsed
// from the header, otherwise multiple representations of the same user occupy distinct cache entries.
type BasicAuthHandler func(ctx AuthzContext, username, password string) (valid, cached bool, err error)

// NewBasicAuthHandler creates a new BasicAuthHandler depending on the lifespan.
func NewBasicAuthHandler(lifespan time.Duration) BasicAuthHandler {
	if lifespan == 0 {
		return DefaultBasicAuthHandler
	}

	return NewCachedBasicAuthHandler(lifespan)
}

// DefaultBasicAuthHandler is a BasicAuthHandler that just checks the username and password directly.
func DefaultBasicAuthHandler(ctx AuthzContext, username, password string) (valid, cached bool, err error) {
	valid, err = ctx.GetUserProvider().CheckUserPassword(username, password)

	return valid, false, err
}

// NewCachedBasicAuthHandler creates a new BasicAuthHandler which uses the authentication.NewCredentialCacheHMAC using
// the sha256 checksum functions.
func NewCachedBasicAuthHandler(lifespan time.Duration) BasicAuthHandler {
	cache := authentication.NewCredentialCacheHMAC(sha256.New, lifespan)

	return func(ctx AuthzContext, username, password string) (valid, cached bool, err error) {
		return cache.Check(ctx, username, password)
	}
}

// Get returns the Authn information for this AuthnStrategy.
func (s *HeaderAuthnStrategy) Get(ctx AuthzContext, _ session.Manager, object *authorization.Object) (authn *Authn, err error) {
	var value []byte

	authn = &Authn{
		Type:     s.authn,
		Level:    authentication.NotAuthenticated,
		Username: anonymous,
		Details:  newAnonymousUserDetails(),
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
		clientID string

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

	var details *authentication.UserDetailsExtended

	switch scheme {
	case model.AuthorizationSchemeBasic:
		details, level, err = handleGetBasic(ctx, s.delay, authn, object, s.headerAuthorize, s.basic)
	case model.AuthorizationSchemeBearer:
		details, clientID, ccs, level, err = handleVerifyGETAuthorizationBearer(ctx, authn, object)
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
	case details == nil:
		return authn, fmt.Errorf("failed to determine user identity from the %s header", s.headerAuthorize)
	case len(details.Username) == 0:
		return authn, fmt.Errorf("failed to determine username from the %s header", s.headerAuthorize)
	default:
		authn.Username = friendlyUsername(details.Username)
		authn.Details = *details
	}

	authn.Level = level

	return authn, nil
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
type HeaderLegacyAuthnStrategy struct {
	delay middlewares.Delayer
}

// Get returns the Authn information for this AuthnStrategy.
func (s *HeaderLegacyAuthnStrategy) Get(ctx AuthzContext, _ session.Manager, object *authorization.Object) (authn *Authn, err error) {
	var (
		value, header []byte
	)

	authn = &Authn{
		Level:    authentication.NotAuthenticated,
		Username: anonymous,
		Details:  newAnonymousUserDetails(),
	}

	if qryValueAuth := ctx.GetRequestQueryArgValue(qryArgAuth); bytes.Equal(qryValueAuth, qryValueBasic) {
		authn.Type = AuthnTypeAuthorization
		header = headerAuthorization
	} else {
		authn.Type = AuthnTypeProxyAuthorization
		header = headerProxyAuthorization
	}

	if value = ctx.GetRequestHeaderValue(header); len(value) == 0 {
		if authn.Type == AuthnTypeAuthorization {
			return authn, fmt.Errorf("header %s expected", headerAuthorization)
		}

		return authn, nil
	}

	authz := model.NewAuthorization()

	if err = authz.ParseBytes(value); err != nil {
		return authn, fmt.Errorf("failed to parse content of %s header: %w", header, err)
	}

	authn.Header.Authorization = authz

	scheme := authn.Header.Authorization.Scheme()

	switch scheme {
	case model.AuthorizationSchemeBasic:
		break
	default:
		ctx.GetLogger().
			WithFields(map[string]any{"scheme": authn.Header.Authorization.SchemeRaw(), "header": string(header)}).
			Debug("Skipping header authorization as the scheme is unknown to this endpoint configuration")

		return authn, fmt.Errorf("header is malformed: unsupported scheme '%s': supported schemes '%s'", scheme, strings.ToTitle(headerAuthorizationSchemeBasic))
	}

	var (
		details *authentication.UserDetailsExtended
		level   authentication.Level
	)

	if details, level, err = handleGetBasic(ctx, s.delay, authn, object, header, DefaultBasicAuthHandler); err != nil {
		return authn, fmt.Errorf("failed to validate %s header with %s scheme: %w", header, scheme, err)
	}

	authn.Username = friendlyUsername(details.Username)
	authn.Details = *details
	authn.Level = level

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

func handleGetBasic(ctx AuthzContext, delayer middlewares.Delayer, authn *Authn, object *authorization.Object, header []byte, validate BasicAuthHandler) (details *authentication.UserDetailsExtended, level authentication.Level, err error) {
	var (
		ban           regulation.BanType
		value         string
		expires       *time.Time
		valid, cached bool
	)

	started := ctx.GetClock().Now()

	defer delayer.CachedDelay(ctx, started, &cached, &valid)

	username, password := authn.Header.Authorization.Basic()

	if len(username) == 0 || len(password) == 0 {
		return nil, authentication.NotAuthenticated, fmt.Errorf("failed to validate parsed credentials of %s header: the username or password was empty", header)
	}

	if details, err = ctx.GetUserProvider().GetDetailsExtendedCached(username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(regulation.BanTypeUnknown, "", nil), regulation.AuthType1FA, object.String(), object.Method, err)

			ctx.GetLogger().WithField("username", username).Error("Error occurred while attempting to get user details for user: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login")
		}

		return nil, authentication.NotAuthenticated, fmt.Errorf("failed to retrieve user details for user %s: %w", username, err)
	} else if details == nil {
		doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(regulation.BanTypeUnknown, "", nil), regulation.AuthType1FA, object.String(), object.Method, err)

		ctx.GetLogger().WithField("username", username).Error("Error occurred while attempting to get user details for user: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login")

		return nil, authentication.NotAuthenticated, fmt.Errorf("failed to retrieve user details for user %s: no user details were returned", username)
	}

	if ban, value, expires, err = ctx.GetProviders().Regulator.BanCheck(ctx, details.Username); err != nil {
		if errors.Is(err, regulation.ErrUserIsBanned) {
			doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(ban, value, expires), regulation.AuthType1FA, object.String(), object.Method, nil)

			return nil, authentication.NotAuthenticated, fmt.Errorf("failed to validate the credentials of user '%s' parsed from the %s header: %w", details.Username, header, err)
		}

		ctx.GetLogger().WithError(err).Errorf(logFmtErrRegulationFail, regulation.AuthType1FA, details.Username)

		return nil, authentication.NotAuthenticated, fmt.Errorf("failed to check the regulation status of user '%s' during an attempt to authenticate using the %s header: %w", details.Username, header, err)
	}

	if valid, cached, err = validate(ctx, details.Username, password); err != nil {
		if isRegulatorSkippedErr(err) {
			ctx.GetLogger().WithError(err).Errorf("Unsuccessful %s authentication attempt by user '%s'", regulation.AuthType1FA, details.Username)
		} else {
			doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthType1FA, object.String(), object.Method, err)
		}

		return nil, authentication.NotAuthenticated, fmt.Errorf("failed to validate the credentials of user '%s' parsed from the %s header: %w", details.Username, header, err)
	}

	if !valid {
		doMarkAuthenticationAttemptWithRequest(ctx, false, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthType1FA, object.String(), object.Method, nil)

		return nil, authentication.NotAuthenticated, fmt.Errorf("failed to validate parsed credentials of %s header valid for user '%s': the username and password do not match", header, details.Username)
	}

	if !cached {
		doMarkAuthenticationAttemptWithRequest(ctx, true, regulation.NewBan(regulation.BanTypeNone, details.Username, nil), regulation.AuthType1FA, object.String(), object.Method, nil)
	}

	return details, authentication.OneFactor, nil
}

// handleAuthnCookieRefreshTTL records when the authentication backend should next be consulted for this user. The user
// details themselves are cached by the user provider, so this only maintains the bookkeeping on the session.
func handleAuthnCookieRefreshTTL(ctx AuthzContext, userSession *session.UserSession, refresh schema.RefreshIntervalDuration) (modified bool) {
	if refresh.Never() || refresh.Always() || userSession.IsAnonymous() {
		return false
	}

	if userSession.RefreshTTL.After(ctx.GetClock().Now()) {
		return false
	}

	userSession.RefreshTTL = ctx.GetClock().Now().Add(refresh.Value())

	return true
}

func handleAuthnCookieValidate(ctx AuthzContext, manager session.Manager, userSession *session.UserSession) (modified, invalid bool) {
	// TODO: Remove this check as it's no longer possible i.e. ineffectual.
	isAnonymous := userSession.Username == ""

	if isAnonymous && userSession.AuthenticationLevel(ctx.GetConfiguration().WebAuthn.EnablePasskey2FA) != authentication.NotAuthenticated {
		ctx.GetLogger().WithFields(map[string]any{"username": anonymous, "level": userSession.AuthenticationLevel(ctx.GetConfiguration().WebAuthn.EnablePasskey2FA).String()}).Errorf("Session for user has an invalid authentication level: this may be a sign of a compromise")

		return modified, true
	}

	if invalid = handleAuthnCookieValidateInactivity(ctx, manager, userSession, isAnonymous); invalid {
		ctx.GetLogger().WithField("username", userSession.Username).Info("Session for user not marked as remembered has exceeded configured session inactivity")

		return modified, true
	}

	if username := ctx.GetRequestHeaderValue(headerSessionUsername); username != nil && !strings.EqualFold(string(username), userSession.Username) {
		ctx.GetLogger().WithField("username", userSession.Username).Warnf("Session for user does not match the Session-Username header with value '%s' which could be a sign of a cookie hijack", username)

		return modified, true
	}

	if handleAuthnCookieValidateActivityRefresh(ctx, manager, userSession, isAnonymous) {
		modified = true

		userSession.LastActivity = ctx.GetClock().Now().Unix()
	}

	return modified, false
}

func handleAuthnCookieValidateActivityRefresh(ctx AuthzContext, manager session.Manager, userSession *session.UserSession, isAnonymous bool) (refresh bool) {
	config := manager.GetSessionConfig()

	if isAnonymous || userSession.KeepMeLoggedIn || config.Inactivity <= 0 {
		return false
	}

	interval := config.Inactivity / sessionActivityRefreshDivisor

	if interval < time.Second {
		interval = time.Second
	}

	return !time.Unix(userSession.LastActivity, 0).Add(interval).After(ctx.GetClock().Now())
}

func handleAuthnCookieValidateInactivity(ctx AuthzContext, manager session.Manager, userSession *session.UserSession, isAnonymous bool) (invalid bool) {
	config := manager.GetSessionConfig()

	if isAnonymous || userSession.KeepMeLoggedIn || int64(config.Inactivity.Seconds()) == 0 {
		return false
	}

	ctx.GetLogger().WithField("username", userSession.Username).Tracef("Inactivity report for user. Current Time: %d, Last Activity: %d, Maximum Inactivity: %d.", ctx.GetClock().Now().Unix(), userSession.LastActivity, int(config.Inactivity.Seconds()))

	return time.Unix(userSession.LastActivity, 0).Add(config.Inactivity).Before(ctx.GetClock().Now())
}

func handleVerifyGETAuthorizationBearer(ctx AuthzContext, authn *Authn, object *authorization.Object) (details *authentication.UserDetailsExtended, clientID string, ccs bool, level authentication.Level, err error) {
	var at bool

	if at, err = oidc.IsAccessToken(ctx, authn.Header.Authorization.Value()); !at {
		if err != nil {
			ctx.GetLogger().WithError(err).Debug("The bearer token does not appear to be a relevant access token")
		} else {
			ctx.GetLogger().Debug("The bearer token does not appear to be a relevant access token")
		}

		return nil, "", false, authentication.NotAuthenticated, errTokenIntent
	}

	var username string

	if username, clientID, ccs, level, err = handleVerifyGETAuthorizationBearerIntrospection(ctx, ctx.GetProviders().OpenIDConnect, authn, object); err != nil {
		return nil, "", false, authentication.NotAuthenticated, err
	}

	return handleVerifyGETAuthorizationBearerResolveUser(ctx, username, clientID, ccs, level)
}

// handleVerifyGETAuthorizationBearerResolveUser turns the result of bearer-token introspection into the final return
// values for handleVerifyGETAuthorizationBearer. For client-credentials grants (ccs=true) there is no associated user
// so GetDetails is skipped and the clientID is propagated; for user-bound tokens GetDetails canonicalises the username.
func handleVerifyGETAuthorizationBearerResolveUser(ctx AuthzContext, username, clientID string, ccs bool, level authentication.Level) (details *authentication.UserDetailsExtended, clientIDOut string, ccsOut bool, levelOut authentication.Level, err error) {
	if ccs {
		return nil, clientID, ccs, level, nil
	}

	if details, err = ctx.GetUserProvider().GetDetailsExtendedCached(username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			ctx.GetLogger().WithField("username", username).Error("Error occurred while attempting to get user details for user: the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login")
		}

		return nil, "", false, authentication.NotAuthenticated, fmt.Errorf("failed to retrieve user details for user %s: %w", username, err)
	}

	return details, clientID, ccs, level, nil
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
