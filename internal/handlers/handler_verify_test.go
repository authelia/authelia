package handlers

import (
	"fmt"
	"net"
	"net/url"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/mocks"
	"github.com/authelia/authelia/internal/session"
	"github.com/authelia/authelia/internal/utils"
)

var verifyGetCfg = schema.AuthenticationBackendConfiguration{
	RefreshInterval: schema.RefreshIntervalDefault,
	Ldap:            &schema.LDAPAuthenticationBackendConfiguration{},
}

// Test getOriginalURL.
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
	assert.Equal(t, "Unable to parse URL extracted from X-Original-URL header: parse \"htt-ps//home?-.example.com\": invalid URI for request", err.Error())
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
	assert.Equal(t, "Missing header X-Forwarded-Proto", err.Error())
}

func TestShouldRaiseWhenNoXForwardedHostHeaderProvidedToDetectTargetURL(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "https")
	_, err := getOriginalURL(mock.Ctx)
	assert.Error(t, err)
	assert.Equal(t, "Missing header X-Forwarded-Host", err.Error())
}

func TestShouldRaiseWhenXForwardedProtoIsNotParsable(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "!:;;:,")
	mock.Ctx.Request.Header.Set("X-Forwarded-Host", "myhost.local")

	_, err := getOriginalURL(mock.Ctx)
	assert.Error(t, err)
	assert.Equal(t, "Unable to parse URL !:;;:,://myhost.local: parse \"!:;;:,://myhost.local\": invalid URI for request", err.Error())
}

func TestShouldRaiseWhenXForwardedURIIsNotParsable(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "https")
	mock.Ctx.Request.Header.Set("X-Forwarded-Host", "myhost.local")
	mock.Ctx.Request.Header.Set("X-Forwarded-URI", "!:;;:,")

	_, err := getOriginalURL(mock.Ctx)
	require.Error(t, err)
	assert.Equal(t, "Unable to parse URL https://myhost.local!:;;:,: parse \"https://myhost.local!:;;:,\": invalid port \":,\" after host", err.Error())
}

// Test parseBasicAuth.
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
	// The decoded format should be user:password.
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

// Test isTargetURLAuthorized.
func TestShouldCheckAuthorizationMatching(t *testing.T) {
	type Rule struct {
		Policy           string
		AuthLevel        authentication.Level
		ExpectedMatching authorizationMatching
	}

	rules := []Rule{
		{"bypass", authentication.NotAuthenticated, Authorized},
		{"bypass", authentication.OneFactor, Authorized},
		{"bypass", authentication.TwoFactor, Authorized},

		{"one_factor", authentication.NotAuthenticated, NotAuthorized},
		{"one_factor", authentication.OneFactor, Authorized},
		{"one_factor", authentication.TwoFactor, Authorized},

		{"two_factor", authentication.NotAuthenticated, NotAuthorized},
		{"two_factor", authentication.OneFactor, NotAuthorized},
		{"two_factor", authentication.TwoFactor, Authorized},

		{"deny", authentication.NotAuthenticated, NotAuthorized},
		{"deny", authentication.OneFactor, Forbidden},
		{"deny", authentication.TwoFactor, Forbidden},
	}

	url, _ := url.ParseRequestURI("https://test.example.com")

	for _, rule := range rules {
		authorizer := authorization.NewAuthorizer(schema.AccessControlConfiguration{
			DefaultPolicy: "deny",
			Rules: []schema.ACLRule{{
				Domains: []string{"test.example.com"},
				Policy:  rule.Policy,
			}},
		})

		username := ""
		if rule.AuthLevel > authentication.NotAuthenticated {
			username = testUsername
		}

		matching := isTargetURLAuthorized(authorizer, *url, username, []string{}, net.ParseIP("127.0.0.1"), rule.AuthLevel)
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

	VerifyGet(verifyGetCfg)(mock.Ctx)

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

	VerifyGet(verifyGetCfg)(mock.Ctx)

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

	VerifyGet(verifyGetCfg)(mock.Ctx)

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

	VerifyGet(verifyGetCfg)(mock.Ctx)

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

	VerifyGet(verifyGetCfg)(mock.Ctx)

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

	VerifyGet(verifyGetCfg)(mock.Ctx)

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

	VerifyGet(verifyGetCfg)(mock.Ctx)
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

	VerifyGet(verifyGetCfg)(mock.Ctx)
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

	VerifyGet(verifyGetCfg)(mock.Ctx)
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
		{"https://test.example.com", "", authentication.NotAuthenticated, 401},
		{"https://bypass.example.com", "", authentication.NotAuthenticated, 200},
		{"https://one-factor.example.com", "", authentication.NotAuthenticated, 401},
		{"https://two-factor.example.com", "", authentication.NotAuthenticated, 401},
		{"https://deny.example.com", "", authentication.NotAuthenticated, 401},

		{"https://test.example.com", "john", authentication.OneFactor, 403},
		{"https://bypass.example.com", "john", authentication.OneFactor, 200},
		{"https://one-factor.example.com", "john", authentication.OneFactor, 200},
		{"https://two-factor.example.com", "john", authentication.OneFactor, 401},
		{"https://deny.example.com", "john", authentication.OneFactor, 403},

		{"https://test.example.com", "john", authentication.TwoFactor, 403},
		{"https://bypass.example.com", "john", authentication.TwoFactor, 200},
		{"https://one-factor.example.com", "john", authentication.TwoFactor, 200},
		{"https://two-factor.example.com", "john", authentication.TwoFactor, 200},
		{"https://deny.example.com", "john", authentication.TwoFactor, 403},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.String(), func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			userSession := mock.Ctx.GetSession()
			userSession.Username = testCase.Username
			userSession.AuthenticationLevel = testCase.AuthenticationLevel
			mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

			mock.Ctx.Request.Header.Set("X-Original-URL", testCase.URL)

			VerifyGet(verifyGetCfg)(mock.Ctx)
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
	past := clock.Now().Add(-1 * time.Hour)

	mock.Ctx.Configuration.Session.Inactivity = testInactivity
	// Reload the session provider since the configuration is indirect.
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session)
	assert.Equal(t, time.Second*10, mock.Ctx.Providers.SessionProvider.Inactivity)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = past.Unix()
	mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	VerifyGet(verifyGetCfg)(mock.Ctx)

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

	mock.Ctx.Configuration.Session.Inactivity = "10s"
	// Reload the session provider since the configuration is indirect.
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session)
	assert.Equal(t, time.Second*10, mock.Ctx.Providers.SessionProvider.Inactivity)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = clock.Now().Add(-1 * time.Hour).Unix()
	mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	VerifyGet(verifyGetCfg)(mock.Ctx)

	// The session has been destroyed.
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, "", newUserSession.Username)
	assert.Equal(t, authentication.NotAuthenticated, newUserSession.AuthenticationLevel)
}

func TestShouldKeepSessionWhenUserCheckedRememberMeAndIsInactiveForTooLong(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = testInactivity

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = 0
	userSession.KeepMeLoggedIn = true
	mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	VerifyGet(verifyGetCfg)(mock.Ctx)

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

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = testInactivity

	past := clock.Now().Add(-1 * time.Hour)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = past.Unix()
	mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	VerifyGet(verifyGetCfg)(mock.Ctx)

	// The session has been destroyed.
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, "john", newUserSession.Username)
	assert.Equal(t, authentication.TwoFactor, newUserSession.AuthenticationLevel)

	// Check the inactivity timestamp has been updated to current time in the new session.
	assert.Equal(t, clock.Now().Unix(), newUserSession.LastActivity)
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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session)
	assert.Equal(t, time.Second*10, mock.Ctx.Providers.SessionProvider.Inactivity)

	past := clock.Now().Add(-1 * time.Hour)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = past.Unix()
	mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	mock.Ctx.QueryArgs().Add("rd", "https://login.example.com")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")

	VerifyGet(verifyGetCfg)(mock.Ctx)

	assert.Equal(t, "Found. Redirecting to https://login.example.com?rd=https%3A%2F%2Ftwo-factor.example.com",
		string(mock.Ctx.Response.Body()))
	assert.Equal(t, 302, mock.Ctx.Response.StatusCode())

	// Check the inactivity timestamp has been updated to current time in the new session.
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, clock.Now().Unix(), newUserSession.LastActivity)
}

func TestShouldUpdateInactivityTimestampEvenWhenHittingForbiddenResources(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	clock := mocks.TestingClock{}
	clock.Set(time.Now())

	mock.Ctx.Configuration.Session.Inactivity = testInactivity

	past := clock.Now().Add(-1 * time.Hour)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.TwoFactor
	userSession.LastActivity = past.Unix()
	mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://deny.example.com")

	VerifyGet(verifyGetCfg)(mock.Ctx)

	// The resource if forbidden.
	assert.Equal(t, 403, mock.Ctx.Response.StatusCode())

	// Check the inactivity timestamp has been updated to current time in the new session.
	newUserSession := mock.Ctx.GetSession()
	assert.Equal(t, clock.Now().Unix(), newUserSession.LastActivity)
}

func TestShouldURLEncodeRedirectionURLParameter(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.AuthenticationLevel = authentication.NotAuthenticated
	mock.Ctx.SaveSession(userSession) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://two-factor.example.com")
	mock.Ctx.Request.SetHost("mydomain.com")
	mock.Ctx.Request.SetRequestURI("/?rd=https://auth.mydomain.com")

	VerifyGet(verifyGetCfg)(mock.Ctx)

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

	cfg := verifyGetCfg
	cfg.RefreshInterval = "disable"
	verifyGet := VerifyGet(cfg)

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
	verifyGet(mock.Ctx)
	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://admin.example.com")
	verifyGet(mock.Ctx)
	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

	// Check Refresh TTL has not been updated.
	userSession = mock.Ctx.GetSession()

	// Check user groups are correct.
	require.Len(t, userSession.Groups, len(user.Groups))
	assert.Equal(t, utils.RFC3339Zero, userSession.RefreshTTL.Unix())
	assert.Equal(t, "admin", userSession.Groups[0])
	assert.Equal(t, "users", userSession.Groups[1])

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://admin.example.com")
	verifyGet(mock.Ctx)
	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

	// Check admin group is not removed from the session.
	userSession = mock.Ctx.GetSession()
	assert.Equal(t, utils.RFC3339Zero, userSession.RefreshTTL.Unix())
	require.Len(t, userSession.Groups, 2)
	assert.Equal(t, "admin", userSession.Groups[0])
	assert.Equal(t, "users", userSession.Groups[1])
}

func TestShouldNotRefreshUserGroupsFromBackendWhenNoGroupSubject(t *testing.T) {
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
	VerifyGet(verifyGetCfg)(mock.Ctx)
	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

	// Session time should NOT have been updated, it should still have a refresh TTL 1 minute in the past.
	userSession = mock.Ctx.GetSession()
	assert.Equal(t, clock.Now().Add(-1*time.Minute).Unix(), userSession.RefreshTTL.Unix())
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

	verifyGet := VerifyGet(verifyGetCfg)

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
	verifyGet(mock.Ctx)
	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

	// Request should get refresh settings and new user details.

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://admin.example.com")
	verifyGet(mock.Ctx)
	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

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
	verifyGet(mock.Ctx)
	assert.Equal(t, 403, mock.Ctx.Response.StatusCode())

	// Check admin group is removed from the session.
	userSession = mock.Ctx.GetSession()
	assert.Equal(t, clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	require.Len(t, userSession.Groups, 1)
	assert.Equal(t, "users", userSession.Groups[0])
}

func TestShouldGetAddedUserGroupsFromBackend(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	//defer mock.Close()

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

	mock.UserProviderMock.EXPECT().GetDetails("john").Times(0)

	verifyGet := VerifyGet(verifyGetCfg)

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
	verifyGet(mock.Ctx)
	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

	// Request should get refresh user profile.
	mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(1)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://grafana.example.com")
	verifyGet(mock.Ctx)
	assert.Equal(t, 403, mock.Ctx.Response.StatusCode())

	// Check Refresh TTL has been updated since grafana.example.com has a group subject and refresh is enabled.
	userSession = mock.Ctx.GetSession()

	// Check user groups are correct.
	require.Len(t, userSession.Groups, len(user.Groups))
	assert.Equal(t, clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	assert.Equal(t, "admin", userSession.Groups[0])
	assert.Equal(t, "users", userSession.Groups[1])

	// Add the grafana group, and force the next request to refresh.
	user.Groups = append(user.Groups, "grafana")
	userSession.RefreshTTL = clock.Now().Add(-1 * time.Second)
	err = mock.Ctx.SaveSession(userSession)
	require.NoError(t, err)

	// Reset otherwise we get the last 403 when we check the Response. Is there a better way to do this?
	mock.Close()

	mock = mocks.NewMockAutheliaCtx(t)
	defer mock.Close()
	err = mock.Ctx.SaveSession(userSession)
	assert.NoError(t, err)

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(1),
	)

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://grafana.example.com")
	VerifyGet(verifyGetCfg)(mock.Ctx)
	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())

	// Check admin group is removed from the session.
	userSession = mock.Ctx.GetSession()
	assert.Equal(t, true, userSession.KeepMeLoggedIn)
	assert.Equal(t, authentication.TwoFactor, userSession.AuthenticationLevel)
	assert.Equal(t, clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	require.Len(t, userSession.Groups, 3)
	assert.Equal(t, "admin", userSession.Groups[0])
	assert.Equal(t, "users", userSession.Groups[1])
	assert.Equal(t, "grafana", userSession.Groups[2])
}
