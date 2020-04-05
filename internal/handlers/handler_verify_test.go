package handlers

import (
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/authelia/authelia/internal/session"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/mocks"
)

// Test getOriginalURL
func TestShouldGetOriginalURLFromOriginalURLHeader(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://home.example.com")
	originalURL, err := getOriginalURL(mock.Ctx)
	assert.NoError(t, err)

	expectedURL, err := url.ParseRequestURI("https://home.example.com")
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)
}

func TestShouldGetOriginalURLFromForwardedHeadersWithoutURI(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "https")
	mock.Ctx.Request.Header.Set("X-Forwarded-Host", "home.example.com")
	originalURL, err := getOriginalURL(mock.Ctx)
	assert.NoError(t, err)

	expectedURL, err := url.ParseRequestURI("https://home.example.com")
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)
}

func TestShouldGetOriginalURLFromForwardedHeadersWithURI(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	mock.Ctx.Request.Header.Set("X-Original-URL", "htt-ps//home?-.example.com")
	_, err := getOriginalURL(mock.Ctx)
	assert.Error(t, err)
	assert.Equal(t, "Unable to parse URL extracted from X-Original-URL header: parse htt-ps//home?-.example.com: invalid URI for request", err.Error())
}

func TestShouldRaiseWhenTargetUrlIsMalformed(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "https")
	mock.Ctx.Request.Header.Set("X-Forwarded-Host", "home.example.com")
	mock.Ctx.Request.Header.Set("X-Forwarded-URI", "/abc")
	originalURL, err := getOriginalURL(mock.Ctx)
	assert.NoError(t, err)

	expectedURL, err := url.ParseRequestURI("https://home.example.com/abc")
	assert.NoError(t, err)
	assert.Equal(t, expectedURL, originalURL)
}

func TestShouldRaiseWhenNoHeaderProvidedToDetectTargetURL(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	_, err := getOriginalURL(mock.Ctx)
	assert.Error(t, err)
	assert.Equal(t, "Missing header X-Fowarded-Proto", err.Error())
}

func TestShouldRaiseWhenNoXForwardedHostHeaderProvidedToDetectTargetURL(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "https")
	_, err := getOriginalURL(mock.Ctx)
	assert.Error(t, err)
	assert.Equal(t, "Missing header X-Fowarded-Host", err.Error())
}

func TestShouldRaiseWhenXForwardedProtoIsNotParseable(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "!:;;:,")
	mock.Ctx.Request.Header.Set("X-Forwarded-Host", "myhost.local")
	_, err := getOriginalURL(mock.Ctx)
	assert.Error(t, err)
	assert.Equal(t, "Unable to parse URL !:;;:,://myhost.local: parse !:;;:,://myhost.local: invalid URI for request", err.Error())
}

func TestShouldRaiseWhenXForwardedURIIsNotParseable(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "https")
	mock.Ctx.Request.Header.Set("X-Forwarded-Host", "myhost.local")
	mock.Ctx.Request.Header.Set("X-Forwarded-URI", "!:;;:,")
	_, err := getOriginalURL(mock.Ctx)
	require.Error(t, err)
	assert.Equal(t, "Unable to parse URL https://myhost.local!:;;:,: parse https://myhost.local!:;;:,: invalid port \":,\" after host", err.Error())
}

// Test parseBasicAuth
func TestShouldRaiseWhenHeaderDoesNotContainBasicPrefix(t *testing.T) {
	_, _, err := parseBasicAuth("alzefzlfzemjfej==")
	assert.Error(t, err)
	assert.Equal(t, "Basic prefix not found in Proxy-Authorization header", err.Error())
}

func TestShouldRaiseWhenCredentialsAreNotInBase64(t *testing.T) {
	_, _, err := parseBasicAuth("Basic alzefzlfzemjfej==")
	assert.Error(t, err)
	assert.Equal(t, "illegal base64 data at input byte 16", err.Error())
}

func TestShouldRaiseWhenCredentialsAreNotInCorrectForm(t *testing.T) {
	// the decoded format should be user:password.
	_, _, err := parseBasicAuth("Basic am9obiBwYXNzd29yZA==")
	assert.Error(t, err)
	assert.Equal(t, "Format of Proxy-Authorization header must be user:password", err.Error())
}

func TestShouldReturnUsernameAndPassword(t *testing.T) {
	// the decoded format should be user:password.
	user, password, err := parseBasicAuth("Basic am9objpwYXNzd29yZA==")
	assert.NoError(t, err)
	assert.Equal(t, "john", user)
	assert.Equal(t, "password", password)
}

// Test isTargetURLAuthorized
func TestShouldCheckAuthorizationMatching(t *testing.T) {
	type Rule struct {
		Policy           string
		AuthLevel        authentication.Level
		ExpectedMatching authorizationMatching
	}
	rules := []Rule{
		Rule{"bypass", authentication.NotAuthenticated, Authorized},
		Rule{"bypass", authentication.OneFactor, Authorized},
		Rule{"bypass", authentication.TwoFactor, Authorized},

		Rule{"one_factor", authentication.NotAuthenticated, NotAuthorized},
		Rule{"one_factor", authentication.OneFactor, Authorized},
		Rule{"one_factor", authentication.TwoFactor, Authorized},

		Rule{"two_factor", authentication.NotAuthenticated, NotAuthorized},
		Rule{"two_factor", authentication.OneFactor, NotAuthorized},
		Rule{"two_factor", authentication.TwoFactor, Authorized},

		Rule{"deny", authentication.NotAuthenticated, NotAuthorized},
		Rule{"deny", authentication.OneFactor, Forbidden},
		Rule{"deny", authentication.TwoFactor, Forbidden},
	}

	url, _ := url.ParseRequestURI("https://test.example.com")

	for _, rule := range rules {
		authorizer := authorization.NewAuthorizer(schema.AccessControlConfiguration{
			DefaultPolicy: "deny",
			Rules: []schema.ACLRule{schema.ACLRule{
				Domain: "test.example.com",
				Policy: rule.Policy,
			}},
		})

		username := ""
		if rule.AuthLevel > authentication.NotAuthenticated {
			username = "john"
		}

		matching := isTargetURLAuthorized(authorizer, *url, username, []string{}, net.ParseIP("127.0.0.1"), rule.AuthLevel)
		assert.Equal(t, rule.ExpectedMatching, matching, "policy=%s, authLevel=%v, expected=%v, actual=%v",
			rule.Policy, rule.AuthLevel, rule.ExpectedMatching, matching)
	}
}

// Test verifyBasicAuth
func TestShouldVerifyWrongCredentials(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(false, nil)

	url, _ := url.ParseRequestURI("https://test.example.com")
	_, _, _, err := verifyBasicAuth([]byte("Basic am9objpwYXNzd29yZA=="), *url, mock.Ctx)

	assert.Error(t, err)
}

type TestCase struct {
	URL                string
	Authorization      string
	ExpectedStatusCode int
}

func (tc TestCase) String() string {
	return fmt.Sprintf("url=%s, auth=%s, exp_status=%d", tc.URL, tc.Authorization, tc.ExpectedStatusCode)
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

	VerifyGet(mock.Ctx)

	assert.Equal(s.T(), 401, mock.Ctx.Response.StatusCode())
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

	VerifyGet(mock.Ctx)

	assert.Equal(s.T(), 403, mock.Ctx.Response.StatusCode())
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

	VerifyGet(mock.Ctx)

	assert.Equal(s.T(), 200, mock.Ctx.Response.StatusCode())
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

	VerifyGet(mock.Ctx)

	assert.Equal(s.T(), 200, mock.Ctx.Response.StatusCode())
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

	VerifyGet(mock.Ctx)

	assert.Equal(s.T(), 401, mock.Ctx.Response.StatusCode())
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

	VerifyGet(mock.Ctx)

	assert.Equal(s.T(), 403, mock.Ctx.Response.StatusCode())
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

	VerifyGet(mock.Ctx)
	expStatus, actualStatus := 401, mock.Ctx.Response.StatusCode()
	assert.Equal(t, expStatus, actualStatus, "URL=%s -> StatusCode=%d != ExpectedStatusCode=%d",
		"https://test.example.com", actualStatus, expStatus)
}

func TestShouldVerifyFailingPasswordCheckingInBasicAuth(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("wrongpass")).
		Return(false, fmt.Errorf("Failed"))

	mock.Ctx.Request.Header.Set("Proxy-Authorization", "Basic am9objp3cm9uZ3Bhc3M=")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://test.example.com")

	VerifyGet(mock.Ctx)
	expStatus, actualStatus := 401, mock.Ctx.Response.StatusCode()
	assert.Equal(t, expStatus, actualStatus, "URL=%s -> StatusCode=%d != ExpectedStatusCode=%d",
		"https://test.example.com", actualStatus, expStatus)
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

	VerifyGet(mock.Ctx)
	expStatus, actualStatus := 401, mock.Ctx.Response.StatusCode()
	assert.Equal(t, expStatus, actualStatus, "URL=%s -> StatusCode=%d != ExpectedStatusCode=%d",
		"https://test.example.com", actualStatus, expStatus)
}

type Pair struct {
	URL                 string
	Username            string
	AuthenticationLevel authentication.Level
	ExpectedStatusCode  int
}

func (p Pair) String() string {
	return fmt.Sprintf("url=%s, username=%s, auth_lvl=%d, exp_status=%d",
		p.URL, p.Username, p.AuthenticationLevel, p.ExpectedStatusCode)
}

func TestShouldVerifyAuthorizationsUsingSessionCookie(t *testing.T) {
	testCases := []Pair{
		Pair{"https://test.example.com", "", authentication.NotAuthenticated, 401},
		Pair{"https://bypass.example.com", "", authentication.NotAuthenticated, 200},
		Pair{"https://one-factor.example.com", "", authentication.NotAuthenticated, 401},
		Pair{"https://two-factor.example.com", "", authentication.NotAuthenticated, 401},
		Pair{"https://deny.example.com", "", authentication.NotAuthenticated, 401},

		Pair{"https://test.example.com", "john", authentication.OneFactor, 403},
		Pair{"https://bypass.example.com", "john", authentication.OneFactor, 200},
		Pair{"https://one-factor.example.com", "john", authentication.OneFactor, 200},
		Pair{"https://two-factor.example.com", "john", authentication.OneFactor, 401},
		Pair{"https://deny.example.com", "john", authentication.OneFactor, 403},

		Pair{"https://test.example.com", "john", authentication.TwoFactor, 403},
		Pair{"https://bypass.example.com", "john", authentication.TwoFactor, 200},
		Pair{"https://one-factor.example.com", "john", authentication.TwoFactor, 200},
		Pair{"https://two-factor.example.com", "john", authentication.TwoFactor, 200},
		Pair{"https://deny.example.com", "john", authentication.TwoFactor, 403},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.String(), func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			userSession := mock.Ctx.GetSession()
			userSession.Username = testCase.Username
			userSession.AuthenticationLevel = testCase.AuthenticationLevel
			mock.Ctx.SaveSession(userSession)

			mock.Ctx.Request.Header.Set("X-Original-URL", testCase.URL)

			VerifyGet(mock.Ctx)
			expStatus, actualStatus := testCase.ExpectedStatusCode, mock.Ctx.Response.StatusCode()
			assert.Equal(t, expStatus, actualStatus, "URL=%s -> AuthLevel=%d, StatusCode=%d != ExpectedStatusCode=%d",
				testCase.URL, testCase.AuthenticationLevel, actualStatus, expStatus)

			if testCase.ExpectedStatusCode == 200 && testCase.Username != "" {
				assert.Equal(t, []byte(testCase.Username), mock.Ctx.Response.Header.Peek("Remote-User"))
			} else {
				assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek("Remote-User"))
			}
		})
	}
}

func TestShouldDestroySessionWhenInactiveForTooLong(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = "10"
	// Reload the session provider since the configuration is indirect
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session)
	assert.Equal(t, time.Second*10, mock.Ctx.Providers.SessionProvider.Inactivity)

	userSession := mock.Ctx.GetSession()
	userSession.Username = "john"
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = clock.Now().Add(-1 * time.Hour).Unix()
	mock.Ctx.SaveSession(userSession)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	VerifyGet(mock.Ctx)

	// The session has been destroyed
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, "", newUserSession.Username)
	assert.Equal(t, authentication.NotAuthenticated, newUserSession.AuthenticationLevel)
}

func TestShouldDestroySessionWhenInactiveForTooLongUsingDurationNotation(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = "10s"
	// Reload the session provider since the configuration is indirect
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session)
	assert.Equal(t, time.Second*10, mock.Ctx.Providers.SessionProvider.Inactivity)

	userSession := mock.Ctx.GetSession()
	userSession.Username = "john"
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = clock.Now().Add(-1 * time.Hour).Unix()
	mock.Ctx.SaveSession(userSession)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	VerifyGet(mock.Ctx)

	// The session has been destroyed
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, "", newUserSession.Username)
	assert.Equal(t, authentication.NotAuthenticated, newUserSession.AuthenticationLevel)
}

func TestShouldKeepSessionWhenUserCheckedRememberMeAndIsInactiveForTooLong(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = "10"

	userSession := mock.Ctx.GetSession()
	userSession.Username = "john"
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = clock.Now().Add(-1 * time.Hour).Unix()
	userSession.KeepMeLoggedIn = true
	mock.Ctx.SaveSession(userSession)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	VerifyGet(mock.Ctx)

	// The session has been destroyed
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, "john", newUserSession.Username)
	assert.Equal(t, authentication.TwoFactor, newUserSession.AuthenticationLevel)
}

func TestShouldKeepSessionWhenInactivityTimeoutHasNotBeenExceeded(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = "10"

	userSession := mock.Ctx.GetSession()
	userSession.Username = "john"
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = clock.Now().Add(-1 * time.Second).Unix()
	mock.Ctx.SaveSession(userSession)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	VerifyGet(mock.Ctx)

	// The session has been destroyed
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, "john", newUserSession.Username)
	assert.Equal(t, authentication.TwoFactor, newUserSession.AuthenticationLevel)
}

func TestShouldURLEncodeRedirectionURLParameter(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	userSession := mock.Ctx.GetSession()
	userSession.Username = "john"
	userSession.AuthenticationLevel = authentication.NotAuthenticated
	mock.Ctx.SaveSession(userSession)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	mock.Ctx.Request.SetHost("mydomain.com")
	mock.Ctx.Request.SetRequestURI("/?rd=https://auth.mydomain.com")

	VerifyGet(mock.Ctx)

	assert.Equal(t, "Found. Redirecting to https://auth.mydomain.com?rd=https%3A%2F%2Ftwo-factor.example.com",
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

	// cookies readable by a service on a machine is also readable by a service on the same machine
	// with a different port as mentioned in https://tools.ietf.org/html/rfc6265#section-8.5
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
