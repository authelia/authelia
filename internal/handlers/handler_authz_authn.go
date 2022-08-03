package handlers

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewCookieAuthnStrategy creates a new CookieAuthnStrategy using the Authorization and WWW-Authenticate
// headers, and the 407 Proxy Auth Required response.
func NewCookieAuthnStrategy(refreshInterval time.Duration) *CookieAuthnStrategy {
	if refreshInterval < time.Second*0 {
		return &CookieAuthnStrategy{}
	}

	return &CookieAuthnStrategy{
		refreshEnabled:  true,
		refreshInterval: refreshInterval,
	}
}

// NewAuthorizationHeaderAuthnStrategy creates a new HeaderAuthnStrategy using the Authorization and WWW-Authenticate
// headers, and the 407 Proxy Auth Required response.
func NewAuthorizationHeaderAuthnStrategy() *HeaderAuthnStrategy {
	return &HeaderAuthnStrategy{
		authn:              AuthnTypeAuthorization,
		headerAuthorize:    headerAuthorization,
		headerAuthenticate: headerWWWAuthenticate,
		handleAuthenticate: true,
		statusAuthenticate: fasthttp.StatusUnauthorized,
	}
}

// NewProxyAuthorizationHeaderAuthnStrategy creates a new HeaderAuthnStrategy using the Proxy-Authorization and
// Proxy-Authenticate headers, and the 407 Proxy Auth Required response.
func NewProxyAuthorizationHeaderAuthnStrategy() *HeaderAuthnStrategy {
	return &HeaderAuthnStrategy{
		authn:              AuthnTypeProxyAuthorization,
		headerAuthorize:    headerProxyAuthorization,
		headerAuthenticate: headerProxyAuthenticate,
		handleAuthenticate: true,
		statusAuthenticate: fasthttp.StatusProxyAuthRequired,
	}
}

// NewLegacyHeaderAuthnStrategy creates a new LegacyHeaderAuthnStrategy.
func NewLegacyHeaderAuthnStrategy() *LegacyHeaderAuthnStrategy {
	return &LegacyHeaderAuthnStrategy{}
}

// CookieAuthnStrategy is a session cookie AuthnStrategy.
type CookieAuthnStrategy struct {
	refreshEnabled  bool
	refreshInterval time.Duration
}

// Get returns the Authn information for this AuthnStrategy.
func (s *CookieAuthnStrategy) Get(ctx *middlewares.AutheliaCtx) (authn Authn, err error) {
	authn = Authn{
		Type:  AuthnTypeCookie,
		Level: authentication.NotAuthenticated,
	}

	userSession := ctx.GetSession()

	if invalid := handleVerifyGETAuthnCookieValidate(ctx, &userSession, s.refreshEnabled, s.refreshInterval); invalid {
		if err = ctx.Providers.SessionProvider.DestroySession(ctx.RequestCtx); err != nil {
			ctx.Logger.Errorf("Unable to destroy user session: %+v", err)
		}

		return authn, nil
	}

	if err = ctx.SaveSession(userSession); err != nil {
		ctx.Logger.Errorf("Unable to save updated user session: %+v", err)
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
func (s *CookieAuthnStrategy) CanHandleUnauthorized() (handle bool) {
	return false
}

// HandleUnauthorized is the Unauthorized handler for the cookie AuthnStrategy.
func (s *CookieAuthnStrategy) HandleUnauthorized(_ *middlewares.AutheliaCtx, _ *Authn, _ *url.URL) {
}

// HeaderAuthnStrategy is a header AuthnStrategy.
type HeaderAuthnStrategy struct {
	authn              AuthnType
	headerAuthorize    []byte
	headerAuthenticate []byte
	handleAuthenticate bool
	statusAuthenticate int
}

// Get returns the Authn information for this AuthnStrategy.
func (s *HeaderAuthnStrategy) Get(ctx *middlewares.AutheliaCtx) (authn Authn, err error) {
	var (
		username, password string
		value              []byte
	)

	authn = Authn{
		Type:  s.authn,
		Level: authentication.NotAuthenticated,
	}

	if value = ctx.Request.Header.PeekBytes(s.headerAuthorize); value == nil {
		return authn, nil
	}

	if username, password, err = headerAuthorizationParseBasic(value); err != nil {
		return authn, fmt.Errorf("failed to parse content of %s header: %w", s.headerAuthorize, err)
	}

	if username == "" || password == "" {
		return authn, fmt.Errorf("failed to validate parsed credentials of %s header for user '%s': %w", s.headerAuthorize, username, err)
	}

	var (
		valid   bool
		details *authentication.UserDetails
	)

	if valid, err = ctx.Providers.UserProvider.CheckUserPassword(username, password); err != nil {
		return authn, fmt.Errorf("failed to validate parsed credentials of %s header for user '%s': %w", s.headerAuthorize, username, err)
	}

	if !valid {
		return authn, fmt.Errorf("validated parsed credentials of %s header but they are not valid for user '%s': %w", s.headerAuthorize, username, err)
	}

	if details, err = ctx.Providers.UserProvider.GetDetails(username); err != nil {
		if errors.Is(err, authentication.ErrUserNotFound) {
			ctx.Logger.Errorf("Error occurred while attempting to get user details for user '%s': the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login", username)

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
func (s *HeaderAuthnStrategy) CanHandleUnauthorized() (handle bool) {
	return s.handleAuthenticate
}

// HandleUnauthorized is the Unauthorized handler for the header AuthnStrategy.
func (s *HeaderAuthnStrategy) HandleUnauthorized(ctx *middlewares.AutheliaCtx, _ *Authn, _ *url.URL) {
	ctx.ReplyStatusCode(s.statusAuthenticate)

	if s.headerAuthenticate != nil {
		ctx.Response.Header.SetBytesKV(s.headerAuthenticate, headerValueAuthenticateBasic)
	}
}

// LegacyHeaderAuthnStrategy is a legacy header AuthnStrategy which can be switched based on the query parameters.
type LegacyHeaderAuthnStrategy struct{}

// Get returns the Authn information for this AuthnStrategy.
func (s *LegacyHeaderAuthnStrategy) Get(ctx *middlewares.AutheliaCtx) (authn Authn, err error) {
	var (
		username, password string
		value, header      []byte
	)

	authn = Authn{
		Level: authentication.NotAuthenticated,
	}

	if auth := ctx.QueryArgs().PeekBytes(queryArgumentAuth); bytes.Equal(auth, valueBasic) {
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

	if username, password, err = headerAuthorizationParseBasic(value); err != nil {
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
			ctx.Logger.Errorf("Error occurred while attempting to get user details for user '%s': the user was not found indicating they were deleted, disabled, or otherwise no longer authorized to login", username)

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
func (s *LegacyHeaderAuthnStrategy) CanHandleUnauthorized() (handle bool) {
	return true
}

// HandleUnauthorized is the Unauthorized handler for the Legacy header AuthnStrategy.
func (s *LegacyHeaderAuthnStrategy) HandleUnauthorized(ctx *middlewares.AutheliaCtx, authn *Authn, _ *url.URL) {
	switch authn.Type {
	case AuthnTypeProxyAuthorization:
		ctx.ReplyStatusCode(fasthttp.StatusUnauthorized)
	case AuthnTypeAuthorization:
		ctx.ReplyStatusCode(fasthttp.StatusUnauthorized)
		ctx.Response.Header.SetBytesKV(headerWWWAuthenticate, headerValueAuthenticateBasic)
	}
}

func handleVerifyGETAuthnCookieValidate(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, profileRefreshEnabled bool, profileRefreshInterval time.Duration) (invalid bool) {
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

func handleVerifyGETAuthnCookieValidateInactivity(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, isAnonymous bool) (invalid bool) {
	if isAnonymous || userSession.KeepMeLoggedIn || int64(ctx.Providers.SessionProvider.Inactivity.Seconds()) == 0 {
		return false
	}

	ctx.Logger.Tracef("Inactivity report for user '%s'. Current Time: %d, Last Activity: %d, Maximum Inactivity: %d.", userSession.Username, ctx.Clock.Now().Unix(), userSession.LastActivity, int(ctx.Providers.SessionProvider.Inactivity.Seconds()))

	return time.Unix(userSession.LastActivity, 0).Add(ctx.Providers.SessionProvider.Inactivity).Before(ctx.Clock.Now())
}

func handleVerifyGETAuthnCookieValidateUpdate(ctx *middlewares.AutheliaCtx, userSession *session.UserSession, isAnonymous, enabled bool, interval time.Duration) (invalid bool) {
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

func headerAuthorizationParseBasic(value []byte) (username, password string, err error) {
	if bytes.Equal(value, valueEmpty) {
		return "", "", fmt.Errorf("header is malformed: empty value")
	}

	parts := strings.SplitN(string(value), " ", 2)

	if len(parts) != 2 {
		return "", "", fmt.Errorf("header is malformed: does not appear to have a scheme")
	}

	if parts[0] != headerAuthorizationSchemeBasic {
		return "", "", fmt.Errorf("header is malformed: unexpected scheme '%s': expected scheme '%s'", parts[0], headerAuthorizationSchemeBasic)
	}

	var content []byte

	if content, err = base64.StdEncoding.DecodeString(parts[1]); err != nil {
		return "", "", fmt.Errorf("header is malformed: could not decode credentials: %w", err)
	}

	strContent := string(content)
	s := strings.IndexByte(strContent, ':')

	if s < 1 {
		return "", "", fmt.Errorf("header is malformed: format of header must be <user>:<password> but either doesn't have a colon or username")
	}

	return strContent[:s], strContent[s+1:], nil
}
