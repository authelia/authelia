package handlers

import (
	"fmt"
	"net"
	"net/url"
	"testing"

	"github.com/clems4ever/authelia/authentication"
	"github.com/clems4ever/authelia/authorization"
	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/clems4ever/authelia/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, "parse htt-ps//home?-.example.com: invalid URI for request", err.Error())
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
	assert.Equal(t, "Missing headers for detecting target URL", err.Error())
}

// Test parseBasicAuth
func TestShouldRaiseWhenHeaderDoesNotContainBasicPrefix(t *testing.T) {
	_, _, err := parseBasicAuth("alzefzlfzemjfej==")
	assert.Error(t, err)
	assert.Equal(t, "Basic prefix not found in authorization header", err.Error())
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
	assert.Equal(t, "Format for basic auth must be user:password", err.Error())
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

func TestShouldVerifyAuthorizationsUsingBasicAuth(t *testing.T) {
	testCases := []TestCase{
		// Authorization has bad format.
		TestCase{"https://bypass.example.com", "Basic am9objpaaaaaaaaaaaaaaaa", 401},

		// Correct Authorization
		TestCase{"https://test.example.com", "Basic am9objpwYXNzd29yZA==", 403},
		TestCase{"https://bypass.example.com", "Basic am9objpwYXNzd29yZA==", 200},
		TestCase{"https://one-factor.example.com", "Basic am9objpwYXNzd29yZA==", 200},
		TestCase{"https://two-factor.example.com", "Basic am9objpwYXNzd29yZA==", 401},
		TestCase{"https://deny.example.com", "Basic am9objpwYXNzd29yZA==", 403},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.String(), func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(true, nil)

			details := authentication.UserDetails{
				Emails: []string{"john@example.com"},
				Groups: []string{"dev", "admin"},
			}
			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(&details, nil)

			mock.Ctx.Request.Header.Set("Proxy-Authorization", testCase.Authorization)
			mock.Ctx.Request.Header.Set("X-Original-URL", testCase.URL)

			VerifyGet(mock.Ctx)
			expStatus, actualStatus := testCase.ExpectedStatusCode, mock.Ctx.Response.StatusCode()
			assert.Equal(t, expStatus, actualStatus, "URL=%s -> StatusCode=%d != ExpectedStatusCode=%d",
				testCase.URL, actualStatus, expStatus)

			if testCase.ExpectedStatusCode == 200 {
				assert.Equal(t, []byte("john"), mock.Ctx.Response.Header.Peek("Remote-User"))
				assert.Equal(t, []byte("dev,admin"), mock.Ctx.Response.Header.Peek("Remote-Groups"))
			} else {
				assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek("Remote-User"))
				assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek("Remote-Groups"))
			}
		})
	}
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
