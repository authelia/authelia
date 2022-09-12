package handlers

import (
	"fmt"
	"net"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
)

var ConfigAuthz = schema.Configuration{
	Session: schema.SessionConfiguration{
		Domain: "example.com",
	},
	AuthenticationBackend: schema.AuthenticationBackendConfiguration{
		RefreshInterval: schema.RefreshIntervalDefault,
	},
}

func TestShouldRaiseWhenTargetUrlIsMalformed(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", schemeHTTPS)
	mock.Ctx.Request.Header.Set("X-Forwarded-Host", "home.example.com")
	mock.Ctx.Request.Header.Set("X-Forwarded-URI", "/abc")
	originalURL, err := mock.Ctx.GetOriginalURL()
	assert.NoError(t, err)

	expectedURL, err := url.ParseRequestURI("https://home.example.com/abc")
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)
}

func TestShouldRaiseWhenNoHeaderProvidedToDetectTargetURL(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	_, err := mock.Ctx.GetOriginalURL()
	assert.Error(t, err)
	assert.Equal(t, "Missing header X-Forwarded-Host", err.Error())
}

func TestShouldRaiseWhenNoXForwardedHostHeaderProvidedToDetectTargetURL(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", schemeHTTPS)
	_, err := mock.Ctx.GetOriginalURL()
	assert.Error(t, err)
	assert.Equal(t, "Missing header X-Forwarded-Host", err.Error())
}

func TestShouldRaiseWhenXForwardedProtoIsNotParsable(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "!:;;:,")
	mock.Ctx.Request.Header.Set("X-Forwarded-Host", "myhost.local")

	_, err := mock.Ctx.GetOriginalURL()
	assert.EqualError(t, err, "failed to parse URL '!:;;:,://myhost.local/': parse \"!:;;:,://myhost.local/\": invalid URI for request")
}

func TestShouldRaiseWhenXForwardedURIIsNotParsable(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", schemeHTTPS)
	mock.Ctx.Request.Header.Set("X-Forwarded-Host", "myhost.local")
	mock.Ctx.Request.Header.Set("X-Forwarded-URI", "!:;;:,")

	_, err := mock.Ctx.GetOriginalURL()
	assert.EqualError(t, err, "failed to parse URL 'https://myhost.local!:;;:,': parse \"https://myhost.local!:;;:,\": invalid port \":,\" after host")
}

// Test parseBasicAuth.
func TestShouldRaiseWhenHeaderDoesNotContainBasicPrefix(t *testing.T) {
	_, _, err := parseBasicAuth(headerProxyAuthorization, "alzefzlfzemjfej==")
	assert.Error(t, err)
	assert.Equal(t, "Basic prefix not found in Proxy-Authorization header", err.Error())
}

func TestShouldRaiseWhenCredentialsAreNotInBase64(t *testing.T) {
	_, _, err := parseBasicAuth(headerProxyAuthorization, "Basic alzefzlfzemjfej==")
	assert.Error(t, err)
	assert.Equal(t, "illegal base64 data at input byte 16", err.Error())
}

func TestShouldRaiseWhenCredentialsAreNotInCorrectForm(t *testing.T) {
	// The decoded format should be user:password.
	_, _, err := parseBasicAuth(headerProxyAuthorization, "Basic am9obiBwYXNzd29yZA==")
	assert.Error(t, err)
	assert.Equal(t, "format of Proxy-Authorization header must be user:password", err.Error())
}

func TestShouldUseProvidedHeaderName(t *testing.T) {
	// The decoded format should be user:password.
	_, _, err := parseBasicAuth([]byte("HeaderName"), "")
	assert.Error(t, err)
	assert.Equal(t, "Basic prefix not found in HeaderName header", err.Error())
}

func TestShouldReturnUsernameAndPassword(t *testing.T) {
	// the decoded format should be user:password.
	user, password, err := parseBasicAuth(headerProxyAuthorization, "Basic am9objpwYXNzd29yZA==")
	assert.NoError(t, err)
	assert.Equal(t, "john", user)
	assert.Equal(t, "password", password)
}

// Test isTargetURLAuthorized.
func TestShouldCheckAuthorizationMatching(t *testing.T) {
	type Rule struct {
		Policy           string
		AuthLevel        authentication.Level
		ExpectedMatching AuthzResult
	}

	rules := []Rule{
		{"bypass", authentication.NotAuthenticated, AuthzResultAuthorized},
		{"bypass", authentication.OneFactor, AuthzResultAuthorized},
		{"bypass", authentication.TwoFactor, AuthzResultAuthorized},

		{"one_factor", authentication.NotAuthenticated, AuthzResultUnauthorized},
		{"one_factor", authentication.OneFactor, AuthzResultAuthorized},
		{"one_factor", authentication.TwoFactor, AuthzResultAuthorized},

		{"two_factor", authentication.NotAuthenticated, AuthzResultUnauthorized},
		{"two_factor", authentication.OneFactor, AuthzResultUnauthorized},
		{"two_factor", authentication.TwoFactor, AuthzResultAuthorized},

		{"deny", authentication.NotAuthenticated, AuthzResultForbidden},
		{"deny", authentication.OneFactor, AuthzResultForbidden},
		{"deny", authentication.TwoFactor, AuthzResultForbidden},
	}

	u, _ := url.ParseRequestURI("https://test.example.com")

	for _, rule := range rules {
		authorizer := authorization.NewAuthorizer(&schema.Configuration{
			AccessControl: schema.AccessControlConfiguration{
				DefaultPolicy: "deny",
				Rules: []schema.ACLRule{{
					Domains: []string{"test.example.com"},
					Policy:  rule.Policy,
				}},
			}})

		username := ""
		if rule.AuthLevel > authentication.NotAuthenticated {
			username = testUsername
		}

		matching := isTargetURLAuthorized(authorizer, *u, username, []string{}, net.ParseIP("127.0.0.1"), []byte("GET"), rule.AuthLevel)
		assert.Equal(t, rule.ExpectedMatching, matching, "policy=%s, authLevel=%v, expected=%v, actual=%v",
			rule.Policy, rule.AuthLevel, rule.ExpectedMatching, matching)
	}
}

// Test verifyBasicAuth.
func TestShouldVerifyWrongCredentials(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(false, nil)

	_, _, _, _, _, err := verifyBasicAuth(mock.Ctx, headerProxyAuthorization, []byte("Basic am9objpwYXNzd29yZA=="))

	assert.Error(t, err)
}

type BasicAuthorizationSuite struct {
	suite.Suite
}

func NewBasicAuthorizationSuite() *BasicAuthorizationSuite {
	return &BasicAuthorizationSuite{}
}

func (s *BasicAuthorizationSuite) TestShouldNotBeAbleToParseBasicAuth() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objpaaaaaaaaaaaaaaaa")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://test.example.com")

	authz := NewAuthzBuilder().WithImplementationLegacy().Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
}

func (s *BasicAuthorizationSuite) TestShouldApplyDefaultPolicy() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objpwYXNzd29yZA==")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://test.example.com")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
}

func (s *BasicAuthorizationSuite) TestShouldApplyPolicyOfBypassDomain() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objpwYXNzd29yZA==")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://bypass.example.com")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
}

func (s *BasicAuthorizationSuite) TestShouldApplyPolicyOfOneFactorDomain() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objpwYXNzd29yZA==")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://one-factor.example.com")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
}

func (s *BasicAuthorizationSuite) TestShouldApplyPolicyOfTwoFactorDomain() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objpwYXNzd29yZA==")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
}

func (s *BasicAuthorizationSuite) TestShouldApplyPolicyOfDenyDomain() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objpwYXNzd29yZA==")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://deny.example.com")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
}

func (s *BasicAuthorizationSuite) TestShouldVerifyAuthBasicArgOk() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.QueryArgs().Add("auth", "basic")
	mock.Ctx.Request.Header.Set("Authorization", "Basic am9objpwYXNzd29yZA==")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://one-factor.example.com")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(s.T(), []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
}

func (s *BasicAuthorizationSuite) TestShouldVerifyAuthBasicArgFailingNoHeader() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.QueryArgs().Add("auth", "basic")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://one-factor.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "401 Unauthorized", string(mock.Ctx.Response.Body()))
	assert.NotEmpty(s.T(), mock.Ctx.Response.Header.Peek("WWW-Authenticate"))
	assert.Regexp(s.T(), regexp.MustCompile("^Basic realm="), string(mock.Ctx.Response.Header.Peek("WWW-Authenticate")))
}

func (s *BasicAuthorizationSuite) TestShouldVerifyAuthBasicArgFailingEmptyHeader() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.QueryArgs().Add("auth", "basic")
	mock.Ctx.Request.Header.Set("Authorization", "")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://one-factor.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "401 Unauthorized", string(mock.Ctx.Response.Body()))
	assert.NotEmpty(s.T(), mock.Ctx.Response.Header.Peek("WWW-Authenticate"))
	assert.Regexp(s.T(), regexp.MustCompile("^Basic realm="), string(mock.Ctx.Response.Header.Peek("WWW-Authenticate")))
}

func (s *BasicAuthorizationSuite) TestShouldVerifyAuthBasicArgFailingWrongPassword() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.QueryArgs().Add("auth", "basic")
	mock.Ctx.Request.Header.Set("Authorization", "Basic am9objpwYXNzd29yZA==")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://one-factor.example.com")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(false, nil)

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "401 Unauthorized", string(mock.Ctx.Response.Body()))
	assert.NotEmpty(s.T(), mock.Ctx.Response.Header.Peek("WWW-Authenticate"))
	assert.Regexp(s.T(), regexp.MustCompile("^Basic realm="), string(mock.Ctx.Response.Header.Peek("WWW-Authenticate")))
}

func (s *BasicAuthorizationSuite) TestShouldVerifyAuthBasicArgFailingWrongHeader() {
	mock := mocks.NewMockAutheliaCtx(s.T())
	defer mock.Close()

	mock.Ctx.QueryArgs().Add("auth", "basic")
	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objpwYXNzd29yZA==")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://one-factor.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "401 Unauthorized", string(mock.Ctx.Response.Body()))
	assert.NotEmpty(s.T(), mock.Ctx.Response.Header.Peek("WWW-Authenticate"))
	assert.Regexp(s.T(), regexp.MustCompile("^Basic realm="), string(mock.Ctx.Response.Header.Peek("WWW-Authenticate")))
}

func TestShouldVerifyAuthorizationsUsingBasicAuth(t *testing.T) {
	suite.Run(t, NewBasicAuthorizationSuite())
}

func TestShouldVerifyWrongCredentialsInBasicAuth(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("wrongpass")).
		Return(false, nil)

	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objp3cm9uZ3Bhc3M=")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://test.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
}

func TestShouldVerifyFailingPasswordCheckingInBasicAuth(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("wrongpass")).
		Return(false, fmt.Errorf("Failed"))

	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objp3cm9uZ3Bhc3M=")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://test.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
}

func TestShouldVerifyFailingDetailsFetchingInBasicAuth(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(nil, fmt.Errorf("Failed"))

	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objpwYXNzd29yZA==")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://test.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
}

func TestShouldNotCrashOnEmptyEmail(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Clock.Set(time.Now())

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.Emails = nil
	userSession.AuthenticationLevel = authentication.OneFactor
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://bypass.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerWWWAuthenticate))
	assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerProxyAuthenticate))
	assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerRemoteEmail))
}

type Pair struct {
	URL                 string
	Username            string
	Emails              []string
	AuthenticationLevel authentication.Level
	ExpectedStatusCode  int
}

func (p Pair) String() string {
	return fmt.Sprintf("url=%s, username=%s, auth_lvl=%d, exp_status=%d",
		p.URL, p.Username, p.AuthenticationLevel, p.ExpectedStatusCode)
}

func TestShouldVerifyAuthorizationsUsingOnlyAuthorizationHeader(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithStrategyAuthorization().
		WithConfig(&config).
		Build()

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://bypass.example.com")

	authz.Handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func TestShouldVerifyAuthorizationsUsingSessionCookie(t *testing.T) {
	testCases := []Pair{
		{"https://test.example.com", "", nil, authentication.NotAuthenticated, fasthttp.StatusForbidden},
		{"https://bypass.example.com", "", nil, authentication.NotAuthenticated, fasthttp.StatusOK},
		{"https://one-factor.example.com", "", nil, authentication.NotAuthenticated, fasthttp.StatusUnauthorized},
		{"https://two-factor.example.com", "", nil, authentication.NotAuthenticated, fasthttp.StatusUnauthorized},
		{"https://deny.example.com", "", nil, authentication.NotAuthenticated, fasthttp.StatusForbidden},

		{"https://test.example.com", "john", []string{"john.doe@example.com"}, authentication.OneFactor, fasthttp.StatusForbidden},
		{"https://bypass.example.com", "john", []string{"john.doe@example.com"}, authentication.OneFactor, fasthttp.StatusOK},
		{"https://one-factor.example.com", "john", []string{"john.doe@example.com"}, authentication.OneFactor, fasthttp.StatusOK},
		{"https://two-factor.example.com", "john", []string{"john.doe@example.com"}, authentication.OneFactor, fasthttp.StatusUnauthorized},
		{"https://deny.example.com", "john", []string{"john.doe@example.com"}, authentication.OneFactor, fasthttp.StatusForbidden},

		{"https://test.example.com", "john", []string{"john.doe@example.com"}, authentication.TwoFactor, fasthttp.StatusForbidden},
		{"https://bypass.example.com", "john", []string{"john.doe@example.com"}, authentication.TwoFactor, fasthttp.StatusOK},
		{"https://one-factor.example.com", "john", []string{"john.doe@example.com"}, authentication.TwoFactor, fasthttp.StatusOK},
		{"https://two-factor.example.com", "john", []string{"john.doe@example.com"}, authentication.TwoFactor, fasthttp.StatusOK},
		{"https://deny.example.com", "john", []string{"john.doe@example.com"}, authentication.TwoFactor, fasthttp.StatusForbidden},
	}

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.String(), func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Clock.Set(time.Now())

			userSession := mock.Ctx.GetSession()
			userSession.Username = testCase.Username
			userSession.Emails = testCase.Emails
			userSession.AuthenticationLevel = testCase.AuthenticationLevel
			userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

			err := mock.Ctx.SaveSession(userSession)
			require.NoError(t, err)

			mock.Ctx.Request.Header.Set("X-Original-URL", testCase.URL)

			authz.Handler(mock.Ctx)

			expStatus, actualStatus := testCase.ExpectedStatusCode, mock.Ctx.Response.StatusCode()
			assert.Equal(t, expStatus, actualStatus, "URL=%s -> AuthLevel=%d, StatusCode=%d != ExpectedStatusCode=%d",
				testCase.URL, testCase.AuthenticationLevel, actualStatus, expStatus)

			if testCase.ExpectedStatusCode == fasthttp.StatusOK && testCase.Username != "" {
				assert.Equal(t, []byte(testCase.Username), mock.Ctx.Response.Header.PeekBytes(headerRemoteUser))
				assert.Equal(t, []byte("john.doe@example.com"), mock.Ctx.Response.Header.PeekBytes(headerRemoteEmail))
			} else {
				assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerRemoteUser))
				assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.PeekBytes(headerRemoteEmail))
			}
		})
	}
}

func TestShouldDestroySessionWhenInactiveForTooLong(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	clock := mocks.TestingClock{}
	clock.Set(time.Now())
	past := clock.Now().Add(-1 * time.Hour)

	mock.Ctx.Configuration.Session.Inactivity = testInactivity
	// Reload the session provider since the configuration is indirect.
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)
	assert.Equal(t, time.Second*10, mock.Ctx.Providers.SessionProvider.Inactivity)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = past.Unix()

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	// The session has been destroyed.
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, "", newUserSession.Username)
	assert.Equal(t, authentication.NotAuthenticated, newUserSession.AuthenticationLevel)

	// Check the inactivity timestamp has been updated to current time in the new session.
	assert.Equal(t, clock.Now().Unix(), newUserSession.LastActivity)
}

func TestShouldDestroySessionWhenInactiveForTooLongUsingDurationNotation(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = time.Second * 10
	// Reload the session provider since the configuration is indirect.
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)
	assert.Equal(t, time.Second*10, mock.Ctx.Providers.SessionProvider.Inactivity)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = clock.Now().Add(-1 * time.Hour).Unix()

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	// The session has been destroyed.
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, "", newUserSession.Username)
	assert.Equal(t, authentication.NotAuthenticated, newUserSession.AuthenticationLevel)
}

func TestShouldKeepSessionWhenUserCheckedRememberMeAndIsInactiveForTooLong(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = testInactivity

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.Emails = []string{"john.doe@example.com"}
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = 0
	userSession.KeepMeLoggedIn = true
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	// Check the session is still active.
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, "john", newUserSession.Username)
	assert.Equal(t, authentication.TwoFactor, newUserSession.AuthenticationLevel)

	// Check the inactivity timestamp is set to 0 in case remember me is checked.
	assert.Equal(t, int64(0), newUserSession.LastActivity)
}

func TestShouldKeepSessionWhenInactivityTimeoutHasNotBeenExceeded(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = testInactivity

	past := mock.Clock.Now().Add(-1 * time.Hour)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.Emails = []string{"john.doe@example.com"}
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = past.Unix()
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	// The session has been destroyed.
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, "john", newUserSession.Username)
	assert.Equal(t, authentication.TwoFactor, newUserSession.AuthenticationLevel)

	// Check the inactivity timestamp has been updated to current time in the new session.
	assert.Equal(t, mock.Clock.Now().Unix(), newUserSession.LastActivity)
}

// In the case of Traefik and Nginx ingress controller in Kube, the response to an inactive
// session is 302 instead of 401.
func TestShouldRedirectWhenSessionInactiveForTooLongAndRDParamProvided(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = testInactivity
	// Reload the session provider since the configuration is indirect.
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)
	assert.Equal(t, time.Second*10, mock.Ctx.Providers.SessionProvider.Inactivity)

	past := clock.Now().Add(-1 * time.Hour)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = past.Unix()

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.QueryArgs().Add(queryStrArgumentRedirect, "https://login.example.com")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	mock.Ctx.Request.Header.Set("X-Forwarded-Method", "GET")
	mock.Ctx.Request.Header.Set("Accept", "text/html; charset=utf-8")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, "<a href=\"https://login.example.com/?rd=https%3A%2F%2Ftwo-factor.example.com&amp;rm=GET\">302 Found</a>",
		string(mock.Ctx.Response.Body()))
	assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())

	// Check the inactivity timestamp has been updated to current time in the new session.
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, clock.Now().Unix(), newUserSession.LastActivity)
}

func TestShouldRedirectWithCorrectStatusCodeBasedOnRequestMethod(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.QueryArgs().Add(queryStrArgumentRedirect, "https://login.example.com")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	mock.Ctx.Request.Header.Set("X-Forwarded-Method", "GET")
	mock.Ctx.Request.Header.Set("Accept", "text/html; charset=utf-8")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, "<a href=\"https://login.example.com/?rd=https%3A%2F%2Ftwo-factor.example.com&amp;rm=GET\">302 Found</a>",
		string(mock.Ctx.Response.Body()))
	assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())

	mock.Ctx.QueryArgs().Add(queryStrArgumentRedirect, "https://login.example.com")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	mock.Ctx.Request.Header.Set("X-Forwarded-Method", "POST")
	mock.Ctx.Request.Header.Set("Accept", "text/html; charset=utf-8")

	authz.Handler(mock.Ctx)

	assert.Equal(t, "<a href=\"https://login.example.com/?rd=https%3A%2F%2Ftwo-factor.example.com&amp;rm=POST\">303 See Other</a>",
		string(mock.Ctx.Response.Body()))
	assert.Equal(t, fasthttp.StatusSeeOther, mock.Ctx.Response.StatusCode())
}

func TestShouldUpdateInactivityTimestampEvenWhenHittingForbiddenResources(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = testInactivity

	past := mock.Clock.Now().Add(-1 * time.Hour)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = past.Unix()
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://deny.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	// The resource if forbidden.
	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())

	// Check the inactivity timestamp has been updated to current time in the new session.
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, mock.Clock.Now().Unix(), newUserSession.LastActivity)
}

func TestShouldURLEncodeRedirectionURLParameter(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Clock.Set(time.Now())

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.NotAuthenticated
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	mock.Ctx.Request.Header.Set("Accept", "text/html; charset=utf-8")
	mock.Ctx.Request.SetHost("mydomain.com")
	mock.Ctx.Request.SetRequestURI("/?rd=https://auth.mydomain.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, "<a href=\"https://auth.mydomain.com/?rd=https%3A%2F%2Ftwo-factor.example.com\">302 Found</a>",
		string(mock.Ctx.Response.Body()))
}

func TestIsDomainProtected(t *testing.T) {
	GetURL := func(u string) *url.URL {
		x, err := url.ParseRequestURI(u)
		require.NoError(t, err)

		return x
	}

	assert.True(t, isURLUnderProtectedDomain(
		GetURL("http://mytest.example.com/abc/?query=abc"), "example.com"))

	assert.True(t, isURLUnderProtectedDomain(
		GetURL("http://example.com/abc/?query=abc"), "example.com"))

	assert.True(t, isURLUnderProtectedDomain(
		GetURL("https://mytest.example.com/abc/?query=abc"), "example.com"))

	// Cookies readable by a service on a machine is also readable by a service on the same machine
	// with a different port as mentioned in https://tools.ietf.org/html/rfc6265#section-8.5.
	assert.True(t, isURLUnderProtectedDomain(
		GetURL("https://mytest.example.com:8080/abc/?query=abc"), "example.com"))
}

func TestSchemeIsHTTPS(t *testing.T) {
	GetURL := func(u string) *url.URL {
		x, err := url.ParseRequestURI(u)
		require.NoError(t, err)

		return x
	}

	assert.False(t, isSchemeHTTPS(
		GetURL("http://mytest.example.com/abc/?query=abc")))
	assert.False(t, isSchemeHTTPS(
		GetURL("ws://mytest.example.com/abc/?query=abc")))
	assert.False(t, isSchemeHTTPS(
		GetURL("wss://mytest.example.com/abc/?query=abc")))
	assert.True(t, isSchemeHTTPS(
		GetURL("https://mytest.example.com/abc/?query=abc")))
}

func TestSchemeIsWSS(t *testing.T) {
	GetURL := func(u string) *url.URL {
		x, err := url.ParseRequestURI(u)
		require.NoError(t, err)

		return x
	}

	assert.False(t, isSchemeWSS(
		GetURL("ws://mytest.example.com/abc/?query=abc")))
	assert.False(t, isSchemeWSS(
		GetURL("http://mytest.example.com/abc/?query=abc")))
	assert.False(t, isSchemeWSS(
		GetURL("https://mytest.example.com/abc/?query=abc")))
	assert.True(t, isSchemeWSS(
		GetURL("wss://mytest.example.com/abc/?query=abc")))
}

func TestShouldNotRefreshUserGroupsFromBackend(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	// Setup pointer to john so we can adjust it during the test.
	user := &authentication.UserDetails{
		Username: "john",
		Groups: []string{
			"admin",
			"users",
		},
		Emails: []string{
			"john@example.com",
		},
	}

	config := ConfigAuthz

	config.AuthenticationBackend.RefreshInterval = "disable"

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	mock.UserProviderMock.EXPECT().GetDetails("john").Times(0)

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	userSession := mock.Ctx.GetSession()
	userSession.Username = user.Username
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = clock.Now().Unix()
	userSession.Groups = user.Groups
	userSession.Emails = user.Emails
	userSession.KeepMeLoggedIn = true
	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://admin.example.com")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	// Check Refresh TTL has not been updated.
	userSession = mock.Ctx.GetSession()

	// Check user groups are correct.
	require.Len(t, userSession.Groups, len(user.Groups))
	assert.Equal(t, utils.RFC3339Zero, userSession.RefreshTTL.Unix())
	assert.Equal(t, "admin", userSession.Groups[0])
	assert.Equal(t, "users", userSession.Groups[1])

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://admin.example.com")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	// Check admin group is not removed from the session.
	userSession = mock.Ctx.GetSession()
	assert.Equal(t, utils.RFC3339Zero, userSession.RefreshTTL.Unix())
	require.Len(t, userSession.Groups, 2)
	assert.Equal(t, "admin", userSession.Groups[0])
	assert.Equal(t, "users", userSession.Groups[1])
}

func TestShouldNotRefreshUserGroupsFromBackendWhenDisabled(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	// Setup user john.
	user := &authentication.UserDetails{
		Username: "john",
		Groups: []string{
			"admin",
			"users",
		},
		Emails: []string{
			"john@example.com",
		},
	}

	mock.UserProviderMock.EXPECT().GetDetails("john").Times(0)

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	userSession := mock.Ctx.GetSession()
	userSession.Username = user.Username
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = clock.Now().Unix()
	userSession.RefreshTTL = clock.Now().Add(-1 * time.Minute)
	userSession.Groups = user.Groups
	userSession.Emails = user.Emails
	userSession.KeepMeLoggedIn = true
	err := mock.Ctx.SaveSession(userSession)

	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	config := ConfigAuthz
	config.AuthenticationBackend.RefreshInterval = schema.ProfileRefreshDisabled

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	// Session time should NOT have been updated, it should still have a refresh TTL 1 minute in the past.
	userSession = mock.Ctx.GetSession()
	assert.Equal(t, clock.Now().Add(-1*time.Minute).Unix(), userSession.RefreshTTL.Unix())
}

func TestShouldDestroySessionWhenUserNotExist(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	// Setup user john.
	user := &authentication.UserDetails{
		Username: "john",
		Groups: []string{
			"admin",
			"users",
		},
		Emails: []string{
			"john@example.com",
		},
	}

	mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(1)

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	userSession := mock.Ctx.GetSession()
	userSession.Username = user.Username
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = clock.Now().Unix()
	userSession.RefreshTTL = clock.Now().Add(-1 * time.Minute)
	userSession.Groups = user.Groups
	userSession.Emails = user.Emails
	userSession.KeepMeLoggedIn = true
	err := mock.Ctx.SaveSession(userSession)

	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	// Session time should NOT have been updated, it should still have a refresh TTL 1 minute in the past.
	userSession = mock.Ctx.GetSession()
	assert.Equal(t, clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())

	// Simulate a Deleted User.
	userSession.RefreshTTL = clock.Now().Add(-1 * time.Minute)
	err = mock.Ctx.SaveSession(userSession)

	require.NoError(t, err)

	mock.UserProviderMock.EXPECT().GetDetails("john").Return(nil, authentication.ErrUserNotFound).Times(1)

	authz.Handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())

	userSession = mock.Ctx.GetSession()
	assert.Equal(t, "", userSession.Username)
	assert.Equal(t, authentication.NotAuthenticated, userSession.AuthenticationLevel)
}

func TestShouldGetRemovedUserGroupsFromBackend(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	// Setup pointer to john so we can adjust it during the test.
	user := &authentication.UserDetails{
		Username: "john",
		Groups: []string{
			"admin",
			"users",
		},
		Emails: []string{
			"john@example.com",
		},
	}

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(2)

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	userSession := mock.Ctx.GetSession()
	userSession.Username = user.Username
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = clock.Now().Unix()
	userSession.RefreshTTL = clock.Now().Add(-1 * time.Minute)
	userSession.Groups = user.Groups
	userSession.Emails = user.Emails
	userSession.KeepMeLoggedIn = true
	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	// Request should get refresh settings and new user details.

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://admin.example.com")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	// Check Refresh TTL has been updated since admin.example.com has a group subject and refresh is enabled.
	userSession = mock.Ctx.GetSession()

	// Check user groups are correct.
	require.Len(t, userSession.Groups, len(user.Groups))
	assert.Equal(t, clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	assert.Equal(t, "admin", userSession.Groups[0])
	assert.Equal(t, "users", userSession.Groups[1])

	// Remove the admin group, and force the next request to refresh.
	user.Groups = []string{"users"}
	userSession.RefreshTTL = clock.Now().Add(-1 * time.Second)
	err = mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://admin.example.com")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())

	// Check admin group is removed from the session.
	userSession = mock.Ctx.GetSession()
	assert.Equal(t, clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	require.Len(t, userSession.Groups, 1)
	assert.Equal(t, "users", userSession.Groups[0])
}

func TestShouldGetAddedUserGroupsFromBackend(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	// Setup pointer to john so we can adjust it during the test.
	user := &authentication.UserDetails{
		Username: "john",
		Groups: []string{
			"admin",
			"users",
		},
		Emails: []string{
			"john@example.com",
		},
	}

	mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(1)

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	mock.Clock.Set(time.Now())

	userSession := mock.Ctx.GetSession()
	userSession.Username = user.Username
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = mock.Clock.Now().Unix()
	userSession.RefreshTTL = mock.Clock.Now().Add(-1 * time.Minute)
	userSession.Groups = user.Groups
	userSession.Emails = user.Emails
	userSession.KeepMeLoggedIn = true
	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://grafana.example.com")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())

	// Check Refresh TTL has been updated since grafana.example.com has a group subject and refresh is enabled.
	userSession = mock.Ctx.GetSession()

	// Check user groups are correct.
	require.Len(t, userSession.Groups, len(user.Groups))
	assert.Equal(t, mock.Clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	assert.Equal(t, "admin", userSession.Groups[0])
	assert.Equal(t, "users", userSession.Groups[1])

	// Add the grafana group, and force the next request to refresh.
	user.Groups = append(user.Groups, "grafana")
	userSession.RefreshTTL = mock.Clock.Now().Add(-1 * time.Second)
	err = mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	// Reset otherwise we get the last 403 when we check the Response. Is there a better way to do this?
	mock.Close()

	mock = mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	err = mock.Ctx.SaveSession(userSession)
	assert.NoError(t, err)

	mock.Clock.Set(time.Now())

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(1),
	)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://grafana.example.com")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	// Check admin group is removed from the session.
	userSession = mock.Ctx.GetSession()
	assert.Equal(t, true, userSession.KeepMeLoggedIn)
	assert.Equal(t, authentication.TwoFactor, userSession.AuthenticationLevel)
	assert.Equal(t, mock.Clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	require.Len(t, userSession.Groups, 3)
	assert.Equal(t, "admin", userSession.Groups[0])
	assert.Equal(t, "users", userSession.Groups[1])
	assert.Equal(t, "grafana", userSession.Groups[2])
}

func TestShouldCheckValidSessionUsernameHeaderAndReturn200(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Clock.Set(time.Now())

	expectedStatusCode := fasthttp.StatusOK

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.OneFactor
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://one-factor.example.com")
	mock.Ctx.Request.Header.SetBytesK(headerSessionUsername, testUsername)

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, expectedStatusCode, mock.Ctx.Response.StatusCode())
	assert.Equal(t, "200 OK", string(mock.Ctx.Response.Body()))
}

func TestShouldCheckInvalidSessionUsernameHeaderAndReturn401(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Clock.Set(time.Now())

	expectedStatusCode := fasthttp.StatusUnauthorized

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.OneFactor
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://one-factor.example.com")
	mock.Ctx.Request.Header.SetBytesK(headerSessionUsername, "root")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, expectedStatusCode, mock.Ctx.Response.StatusCode())
	assert.Equal(t, "401 Unauthorized", string(mock.Ctx.Response.Body()))
}

func TestShouldNotRedirectRequestsForBypassACLWhenInactiveForTooLong(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	clock := mocks.TestingClock{}
	clock.Set(time.Now())
	past := clock.Now().Add(-1 * time.Hour)

	mock.Ctx.Configuration.Session.Inactivity = testInactivity
	// Reload the session provider since the configuration is indirect.
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)
	assert.Equal(t, time.Second*10, mock.Ctx.Providers.SessionProvider.Inactivity)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = past.Unix()

	err := mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	// Should respond 200 OK.
	mock.Ctx.QueryArgs().Add(queryStrArgumentRedirect, "https://login.example.com")
	mock.Ctx.Request.Header.Set("X-Forwarded-Method", "GET")
	mock.Ctx.Request.Header.Set("Accept", "text/html; charset=utf-8")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://bypass.example.com")

	config := ConfigAuthz

	authz := NewAuthzBuilder().
		WithImplementationLegacy().
		WithConfig(&config).
		Build()

	authz.Handler(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Nil(t, mock.Ctx.Response.Header.Peek("Location"))

	// Should respond 302 Found.
	mock.Ctx.QueryArgs().Add(queryStrArgumentRedirect, "https://login.example.com")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	mock.Ctx.Request.Header.Set("X-Forwarded-Method", "GET")
	mock.Ctx.Request.Header.Set("Accept", "text/html; charset=utf-8")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
	assert.Equal(t, "https://login.example.com/?rd=https%3A%2F%2Ftwo-factor.example.com&rm=GET", string(mock.Ctx.Response.Header.Peek("Location")))

	// Should respond 401 Unauthorized.
	mock.Ctx.QueryArgs().Del(queryStrArgumentRedirect)
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	mock.Ctx.Request.Header.Set("X-Forwarded-Method", "GET")
	mock.Ctx.Request.Header.Set("Accept", "text/html; charset=utf-8")
	authz.Handler(mock.Ctx)
	assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	assert.Nil(t, mock.Ctx.Response.Header.Peek("Location"))
}

func TestIsSessionInactiveTooLong(t *testing.T) {
	testCases := []struct {
		name       string
		have       *session.UserSession
		now        time.Time
		inactivity time.Duration
		expected   bool
	}{
		{
			name:       "ShouldNotBeInactiveTooLong",
			have:       &session.UserSession{Username: "john", LastActivity: 1656994960},
			now:        time.Unix(1656994970, 0),
			inactivity: time.Second * 90,
			expected:   false,
		},
		{
			name:       "ShouldNotBeInactiveTooLongIfAnonymous",
			have:       &session.UserSession{Username: "", LastActivity: 1656994960},
			now:        time.Unix(1656994990, 0),
			inactivity: time.Second * 20,
			expected:   false,
		},
		{
			name:       "ShouldNotBeInactiveTooLongIfRemembered",
			have:       &session.UserSession{Username: "john", LastActivity: 1656994960, KeepMeLoggedIn: true},
			now:        time.Unix(1656994990, 0),
			inactivity: time.Second * 20,
			expected:   false,
		},
		{
			name:       "ShouldNotBeInactiveTooLongIfDisabled",
			have:       &session.UserSession{Username: "john", LastActivity: 1656994960},
			now:        time.Unix(1656994990, 0),
			inactivity: time.Second * 0,
			expected:   false,
		},
		{
			name:       "ShouldBeInactiveTooLong",
			have:       &session.UserSession{Username: "john", LastActivity: 1656994960},
			now:        time.Unix(4656994990, 0),
			inactivity: time.Second * 1,
			expected:   true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := mocks.NewMockAutheliaCtx(t)

			defer ctx.Close()

			ctx.Ctx.Configuration.Session.Inactivity = tc.inactivity
			ctx.Ctx.Providers.SessionProvider = session.NewProvider(ctx.Ctx.Configuration.Session, nil)

			ctx.Clock.Set(tc.now)
			ctx.Ctx.Clock = &ctx.Clock

			actual := isSessionInactiveTooLong(ctx.Ctx, tc.have, tc.have.Username == "")

			assert.Equal(t, tc.expected, actual)
		})
	}
}
