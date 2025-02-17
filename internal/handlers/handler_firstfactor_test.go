package handlers

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
)

type FirstFactorSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *FirstFactorSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
}

func (s *FirstFactorSuite) TearDownTest() {
	s.mock.Close()
}

func (s *FirstFactorSuite) TestShouldFailIfBodyIsNil() {
	FirstFactorPOST(nil)(s.mock.Ctx)

	// No body.
	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Failed to parse 1FA request body", "unable to parse body: unexpected end of JSON input")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfBodyIsInBadFormat() {
	// Missing password.
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test"
	}`)
	FirstFactorPOST(nil)(s.mock.Ctx)

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Failed to parse 1FA request body", "unable to validate body: password: non zero value required")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfUserProviderCheckPasswordFail() {
	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{Username: "test"}, nil)

	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(false, fmt.Errorf("failed"))

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "test",
			Successful: false,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthType1FA,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPOST(nil)(s.mock.Ctx)

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Unsuccessful 1FA authentication attempt by user 'test'", "failed")

	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldCheckAuthenticationIsNotMarkedWhenProviderCheckPasswordError() {
	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{Username: "test"}, nil)

	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(false, fmt.Errorf("invalid credentials"))

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "test",
			Successful: false,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthType1FA,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)

	FirstFactorPOST(nil)(s.mock.Ctx)
}

func (s *FirstFactorSuite) TestShouldCheckAuthenticationIsMarkedWhenInvalidCredentials() {
	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{Username: "test"}, nil)

	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(false, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   "test",
			Successful: false,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthType1FA,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)

	FirstFactorPOST(nil)(s.mock.Ctx)
}

func (s *FirstFactorSuite) TestShouldFailIfUserProviderGetDetailsFail() {
	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(nil, fmt.Errorf("failed"))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPOST(nil)(s.mock.Ctx)

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Error occurred getting details for user with username input 'test' which usually indicates they do not exist", "failed")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfAuthenticationMarkFail() {
	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{Username: "test"}, nil)

	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(fmt.Errorf("failed"))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPOST(nil)(s.mock.Ctx)

	AssertLogEntryMessageAndError(s.T(), s.mock.Hook.LastEntry(), "Unable to mark 1FA authentication attempt by user 'test'", "failed")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldAuthenticateUserWithRememberMeChecked() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{
			Username: "test",
			Emails:   []string{"test@example.com"},
			Groups:   []string{"dev", "admins"},
		}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	assert.Equal(s.T(), "test", userSession.Username)
	assert.Equal(s.T(), true, userSession.KeepMeLoggedIn)
	assert.Equal(s.T(), authentication.OneFactor, userSession.AuthenticationLevel)
	assert.Equal(s.T(), []string{"test@example.com"}, userSession.Emails)
	assert.Equal(s.T(), []string{"dev", "admins"}, userSession.Groups)
}

func (s *FirstFactorSuite) TestShouldAuthenticateUserWithRememberMeUnchecked() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{
			Username: "test",
			Emails:   []string{"test@example.com"},
			Groups:   []string{"dev", "admins"},
		}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false
	}`)
	FirstFactorPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	assert.Equal(s.T(), "test", userSession.Username)
	assert.Equal(s.T(), false, userSession.KeepMeLoggedIn)
	assert.Equal(s.T(), authentication.OneFactor, userSession.AuthenticationLevel)
	assert.Equal(s.T(), []string{"test@example.com"}, userSession.Emails)
	assert.Equal(s.T(), []string{"dev", "admins"}, userSession.Groups)
}

func (s *FirstFactorSuite) TestShouldAuthenticateUserWithEmailAsUsernameInput() {
	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			GetDetails(gomock.Eq("test@example.com")).
			Return(&authentication.UserDetails{
				Username: "test",
				Emails:   []string{"test@example.com"},
				Groups:   []string{"dev", "admins"},
			}, nil),
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
			Return(true, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{Time: s.mock.Clock.Now(), Successful: true, Username: "test", Type: regulation.AuthType1FA, RemoteIP: model.NewNullIP(s.mock.Ctx.RemoteIP())})).
			Return(nil),
	)

	s.mock.Ctx.Request.SetBodyString(`{"username":"test@example.com","password":"hello","requestMethod":"GET","keepMeLoggedIn":false}`)
	FirstFactorPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.Equal(fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	s.Equal([]byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	s.Equal("test", userSession.Username)
	s.Equal(false, userSession.KeepMeLoggedIn)
	s.Equal(authentication.OneFactor, userSession.AuthenticationLevel)
	s.Equal([]string{"test@example.com"}, userSession.Emails)
	s.Equal([]string{"dev", "admins"}, userSession.Groups)
}

func (s *FirstFactorSuite) TestShouldSaveUsernameFromAuthenticationBackendInSession() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("Test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{
			// This is the name in authentication backend, in some setups the binding is
			// case insensitive but the user ID in session must match the user in LDAP
			// for the other modules of Authelia to be coherent.
			Username: "Test",
			Emails:   []string{"test@example.com"},
			Groups:   []string{"dev", "admins"},
		}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	assert.Equal(s.T(), "Test", userSession.Username)
	assert.Equal(s.T(), true, userSession.KeepMeLoggedIn)
	assert.Equal(s.T(), authentication.OneFactor, userSession.AuthenticationLevel)
	assert.Equal(s.T(), []string{"test@example.com"}, userSession.Emails)
	assert.Equal(s.T(), []string{"dev", "admins"}, userSession.Groups)
}

type FirstFactorRedirectionSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *FirstFactorRedirectionSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	s.mock.Ctx.Configuration.Session.Cookies[0].DefaultRedirectionURL = &url.URL{Scheme: "https", Host: "default.local"}
	s.mock.Ctx.Configuration.AccessControl.DefaultPolicy = testBypass
	s.mock.Ctx.Configuration.AccessControl.Rules = []schema.AccessControlRule{
		{
			Domains: []string{"default.local"},
			Policy:  "one_factor",
		},
	}
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&s.mock.Ctx.Configuration)

	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{
			Username: "test",
			Emails:   []string{"test@example.com"},
			Groups:   []string{"dev", "admins"},
		}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(nil)
}

func (s *FirstFactorRedirectionSuite) TearDownTest() {
	s.mock.Close()
}

// When:
//
//	1/ the target url is unknown
//	2/ two_factor is disabled (no policy is set to two_factor)
//	3/ default_redirect_url is provided
//
// Then:
//
//	the user should be redirected to the default url.
func (s *FirstFactorRedirectionSuite) TestShouldRedirectToDefaultURLWhenNoTargetURLProvidedAndTwoFactorDisabled() {
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false
	}`)
	FirstFactorPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), &redirectResponse{Redirect: "https://www.example.com"})
}

// When:
//
//	1/ the target url is unsafe
//	2/ two_factor is disabled (no policy is set to two_factor)
//	3/ default_redirect_url is provided
//
// Then:
//
//	the user should be redirected to the default url.
func (s *FirstFactorRedirectionSuite) TestShouldRedirectToDefaultURLWhenURLIsUnsafeAndTwoFactorDisabled() {
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"targetURL": "http://notsafe.local"
	}`)

	FirstFactorPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), &redirectResponse{Redirect: "https://www.example.com"})
}

// When:
//
//	1/ two_factor is enabled (default policy)
//
// Then:
//
//	the user should receive 200 without redirection URL.
func (s *FirstFactorRedirectionSuite) TestShouldReply200WhenNoTargetURLProvidedAndTwoFactorEnabled() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "two_factor",
		},
	})
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false
	}`)

	FirstFactorPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
}

// When:
//
//	1/ two_factor is enabled (some rule)
//
// Then:
//
//	the user should receive 200 without redirection URL.
func (s *FirstFactorRedirectionSuite) TestShouldReply200WhenUnsafeTargetURLProvidedAndTwoFactorEnabled() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"example.com"},
					Policy:  "two_factor",
				},
			},
		}})
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false
	}`)

	FirstFactorPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
}

func TestFirstFactorSuite(t *testing.T) {
	suite.Run(t, new(FirstFactorSuite))
	suite.Run(t, new(FirstFactorRedirectionSuite))
}
