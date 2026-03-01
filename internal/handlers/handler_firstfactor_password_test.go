package handlers

import (
	"database/sql"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
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
	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// No body.
	s.mock.AssertLastLogMessage(s.T(), "Failed to parse 1FA request body", "unable to parse body: unexpected end of JSON input")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfBodyIsInBadFormat() {
	// Missing password.
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test"
	}`)
	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.AssertLastLogMessage(s.T(), "Failed to parse 1FA request body", "unable to validate body: password: non zero value required")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfUserProviderCheckPasswordFail() {
	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{Username: "test"}, nil)

	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
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

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),

		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq("test")).Return(nil, nil),
	)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.AssertLastLogMessage(s.T(), "Unsuccessful 1FA authentication attempt by user 'test'", "failed")

	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldCheckAuthenticationIsNotMarkedWhenProviderCheckPasswordError() {
	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			GetDetails(gomock.Eq("test")).
			Return(&authentication.UserDetails{Username: testValue}, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
			Return(false, fmt.Errorf("invalid credentials")),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
				Username:   "test",
				Successful: false,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthType1FA,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})),
	)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.AssertLastLogMessage(s.T(), "Unsuccessful 1FA authentication attempt by user 'test'", "invalid credentials")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldCheckAuthenticationIsMarkedWhenInvalidCredentials() {
	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq(testValue)).
		Return(&authentication.UserDetails{
			Username: "test",
			Emails:   []string{"test@example.com"},
			Groups:   []string{"dev", "admins"},
		}, nil)

	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(false, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   testValue,
			Successful: false,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthType1FA,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq("test")).Return(nil, nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)
}

func (s *FirstFactorSuite) TestShouldFailIfUserProviderGetDetailsFail() {
	attempt := model.AuthenticationAttempt{Time: s.mock.Clock.Now(), Type: regulation.AuthType1FA, RemoteIP: model.NewNullIPFromString("0.0.0.0")}

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq(testValue)).
		Return(nil, fmt.Errorf("failed"))

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(attempt)).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.AssertLastLogMessage(s.T(), "Error occurred getting details for user with username input 'test' which usually indicates they do not exist", "failed")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldNotFailIfAuthenticationMarkFail() {
	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			GetDetails(gomock.Eq("test")).
			Return(&authentication.UserDetails{
				Username: "test",
				Emails:   []string{"test@example.com"},
				Groups:   []string{"dev", "admins"},
			}, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
			Return(true, nil),
		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
			Return(fmt.Errorf("failed")),
		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
	)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	s.mock.AssertLastLogMessage(s.T(), "Failed to record 1FA authentication attempt", "failed")
	s.mock.Assert200OK(s.T(), nil)
}

func (s *FirstFactorSuite) TestShouldAuthenticateUserWithRememberMeChecked() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq(testValue)).
		Return(&authentication.UserDetails{
			Username: testValue,
			Emails:   []string{"test@example.com"},
			Groups:   []string{"dev", "admins"},
		}, nil)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq("test")).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	assert.Equal(s.T(), testValue, userSession.Username)
	assert.Equal(s.T(), true, userSession.KeepMeLoggedIn)
	assert.Equal(s.T(), authentication.OneFactor, userSession.AuthenticationLevel(s.mock.Ctx.Configuration.WebAuthn.EnablePasskey2FA))
	assert.Equal(s.T(), []string{"test@example.com"}, userSession.Emails)
	assert.Equal(s.T(), []string{"dev", "admins"}, userSession.Groups)
}

func (s *FirstFactorSuite) TestShouldAuthenticateUserWithRememberMeUnchecked() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq(testValue)).
		Return(&authentication.UserDetails{
			Username: testValue,
			Emails:   []string{"test@example.com"},
			Groups:   []string{"dev", "admins"},
		}, nil)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false
	}`)
	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	assert.Equal(s.T(), testValue, userSession.Username)
	assert.Equal(s.T(), false, userSession.KeepMeLoggedIn)
	assert.Equal(s.T(), authentication.OneFactor, userSession.AuthenticationLevel(s.mock.Ctx.Configuration.WebAuthn.EnablePasskey2FA))
	assert.Equal(s.T(), []string{"test@example.com"}, userSession.Emails)
	assert.Equal(s.T(), []string{"dev", "admins"}, userSession.Groups)
}

func (s *FirstFactorSuite) TestShouldSaveUsernameFromAuthenticationBackendInSession() {
	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq(testValue)).
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
		LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq("Test")).Return(nil, nil)

	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("Test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("Test")).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": true
	}`)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	assert.Equal(s.T(), "Test", userSession.Username)
	assert.Equal(s.T(), true, userSession.KeepMeLoggedIn)
	assert.Equal(s.T(), authentication.OneFactor, userSession.AuthenticationLevel(s.mock.Ctx.Configuration.WebAuthn.EnablePasskey2FA))
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
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq(testValue)).
		Return(&authentication.UserDetails{
			Username: testValue,
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
	s.mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false
	}`)
	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

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
	s.mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"targetURL": "http://notsafe.local"
	}`)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

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
	s.mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
		Return(nil)

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

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

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
	s.mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil)

	s.mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil)

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

	s.mock.StorageMock.
		EXPECT().
		UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false
	}`)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyWhenBadTargetURL() {
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
		"keepMeLoggedIn": false,
		"targetURL": "#https://23kjnm412jk3"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),

		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),

		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyTwoFactorOK() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"targetURL": "https://two-factor.example.com"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),

		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),

		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyTwoTwoFactorUnsafe() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"targetURL": "https://unsafe-domain.com"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),

		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),

		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyTwoTwoFactorSafe() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"targetURL": "https://test.example.com"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),

		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),

		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), &redirectResponse{Redirect: "https://test.example.com"})
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyOpenIDConnectCantParseUUID() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"flow": "openid_connect",
		"flowID": "aaaaaaaaaaaaaaaaaaaaaaaaaaa-9107-4067-8d31-407ca59eb69c"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
	s.mock.AssertLogEntryAdvanced(s.T(), 0, logrus.ErrorLevel, "Error occurred parsing the consent session flow id", map[string]any{"error": "invalid UUID length: 55", "flow": "openid_connect", "flow_id": "aaaaaaaaaaaaaaaaaaaaaaaaaaa-9107-4067-8d31-407ca59eb69c", "subflow": ""})
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyOpenIDConnectCantGetConsentSession() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"flow": "openid_connect",
		"flowID": "d1ba0ad8-9107-4067-8d31-407ca59eb69c"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(s.mock.Ctx), gomock.Eq(uuid.Must(uuid.Parse("d1ba0ad8-9107-4067-8d31-407ca59eb69c")))).
			Return(nil, fmt.Errorf("failed to obtain")),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
	s.mock.AssertLogEntryAdvanced(s.T(), 0, logrus.ErrorLevel, "Error occurred loading the consent session", map[string]any{"error": "failed to obtain", "flow": "openid_connect", "flow_id": "d1ba0ad8-9107-4067-8d31-407ca59eb69c", "subflow": ""})
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyOpenIDConnectConsentSessionAlreadyResponded() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"flow": "openid_connect",
		"flowID": "d1ba0ad8-9107-4067-8d31-407ca59eb69c"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(s.mock.Ctx), gomock.Eq(uuid.Must(uuid.Parse("d1ba0ad8-9107-4067-8d31-407ca59eb69c")))).
			Return(&model.OAuth2ConsentSession{RespondedAt: sql.NullTime{Valid: true, Time: time.Now().Add(-time.Minute)}}, nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
	s.mock.AssertLogEntryAdvanced(s.T(), 0, logrus.ErrorLevel, "Failed to process consent session as it has already been responded to", map[string]any{"flow": "openid_connect", "flow_id": "d1ba0ad8-9107-4067-8d31-407ca59eb69c", "subflow": ""})
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyOpenIDConnectCantGetClient() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})

	s.mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(&schema.Configuration{IdentityProviders: schema.IdentityProviders{OIDC: &schema.IdentityProvidersOpenIDConnect{}}}, s.mock.StorageMock, s.mock.Ctx.Providers.Templates)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"flow": "openid_connect",
		"flowID": "d1ba0ad8-9107-4067-8d31-407ca59eb69c"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(s.mock.Ctx), gomock.Eq(uuid.Must(uuid.Parse("d1ba0ad8-9107-4067-8d31-407ca59eb69c")))).
			Return(&model.OAuth2ConsentSession{ClientID: "abc"}, nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
	s.mock.AssertLogEntryAdvanced(s.T(), 0, logrus.ErrorLevel, "Error occurred loading the client for the consent session", map[string]any{"error": "invalid_client", "client_id": "abc", "flow": "openid_connect", "flow_id": "d1ba0ad8-9107-4067-8d31-407ca59eb69c", "subflow": ""})
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyOpenIDConnectFormRequiresLogin() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})

	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID: "abc",
					},
				},
			},
		},
	}

	s.mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, s.mock.StorageMock, s.mock.Ctx.Providers.Templates)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"flow": "openid_connect",
		"flowID": "d1ba0ad8-9107-4067-8d31-407ca59eb69c"
	}`)

	form := url.Values{
		"max_age": []string{"0"},
	}

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(s.mock.Ctx), gomock.Eq(uuid.Must(uuid.Parse("d1ba0ad8-9107-4067-8d31-407ca59eb69c")))).
			Return(&model.OAuth2ConsentSession{ClientID: "abc", Form: form.Encode()}, nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), &redirectResponse{Redirect: "http://example.com/consent/openid/decision?flow=openid_connect&flow_id=d1ba0ad8-9107-4067-8d31-407ca59eb69c"})
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyOpenIDConnectFormRequiresLoginBadForm() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})

	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID: "abc",
					},
				},
			},
		},
	}

	s.mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, s.mock.StorageMock, s.mock.Ctx.Providers.Templates)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"flow": "openid_connect",
		"flowID": "d1ba0ad8-9107-4067-8d31-407ca59eb69c"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(s.mock.Ctx), gomock.Eq(uuid.Must(uuid.Parse("d1ba0ad8-9107-4067-8d31-407ca59eb69c")))).
			Return(&model.OAuth2ConsentSession{ClientID: "abc", Form: "1238y12978y189gb128g1287g12807g128702g38172%1"}, nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
	s.mock.AssertLogEntryAdvanced(s.T(), 0, logrus.ErrorLevel, "Error occurred getting the original form from the consent session", map[string]any{"error": "invalid URL escape \"%1\"", "client_id": "abc", "flow": "openid_connect", "flow_id": "d1ba0ad8-9107-4067-8d31-407ca59eb69c", "subflow": "", "username": "test"})
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyOpenIDConnectNeeds2FA() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})

	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  "abc",
						AuthorizationPolicy: "two_factor",
					},
				},
			},
		},
	}

	s.mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, s.mock.StorageMock, s.mock.Ctx.Providers.Templates)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"flow": "openid_connect",
		"flowID": "d1ba0ad8-9107-4067-8d31-407ca59eb69c"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(s.mock.Ctx), gomock.Eq(uuid.Must(uuid.Parse("d1ba0ad8-9107-4067-8d31-407ca59eb69c")))).
			Return(&model.OAuth2ConsentSession{ClientID: "abc", Form: ""}, nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
	s.mock.AssertLogEntryAdvanced(s.T(), 0, logrus.InfoLevel, "OpenID Connect 1.0 client requires 2FA", map[string]any{"client_id": "abc", "flow": "openid_connect", "flow_id": "d1ba0ad8-9107-4067-8d31-407ca59eb69c", "subflow": ""})
}

func (s *FirstFactorRedirectionSuite) TestShouldReplyOpenIDConnectNeeds1FA() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "one_factor",
			Rules: []schema.AccessControlRule{
				{
					Domains: []string{"test.example.com"},
					Policy:  "one_factor",
				},
				{
					Domains: []string{"two-factor.example.com"},
					Policy:  "two_factor",
				},
			},
		}})

	config := &schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				Clients: []schema.IdentityProvidersOpenIDConnectClient{
					{
						ID:                  "abc",
						AuthorizationPolicy: "one_factor",
					},
				},
			},
		},
	}

	s.mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, s.mock.StorageMock, s.mock.Ctx.Providers.Templates)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"flow": "openid_connect",
		"flowID": "d1ba0ad8-9107-4067-8d31-407ca59eb69c"
	}`)

	gomock.InOrder(
		s.mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(s.mock.Ctx), gomock.Eq(model.NewIP(s.mock.Ctx.RemoteIP()))).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(s.mock.Ctx), gomock.Eq(testValue)).Return(nil, nil),
		s.mock.StorageMock.
			EXPECT().
			UpdateUserSignInDateByUsername(s.mock.Ctx, gomock.Eq("test")).
			Return(nil),
		s.mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(s.mock.Ctx), gomock.Eq(uuid.Must(uuid.Parse("d1ba0ad8-9107-4067-8d31-407ca59eb69c")))).
			Return(&model.OAuth2ConsentSession{ClientID: "abc", Form: "grant_type=authorization_code"}, nil),
	)

	FirstFactorPasswordPOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), &redirectResponse{Redirect: "http://example.com/api/oidc/authorization?consent_id=d1ba0ad8-9107-4067-8d31-407ca59eb69c&grant_type=authorization_code"})
}

type FirstFactorReauthenticateSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *FirstFactorReauthenticateSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())

	session, err := s.mock.Ctx.GetSession()

	s.Require().NoError(err)

	session.SetOneFactorPasskey(s.mock.Ctx.Clock.Now(), &authentication.UserDetails{Username: testValue}, false, false, false, false)

	s.Require().NoError(s.mock.Ctx.SaveSession(session))
}

func (s *FirstFactorReauthenticateSuite) TearDownTest() {
	s.mock.Close()
}

func (s *FirstFactorReauthenticateSuite) TestShouldFailIfBodyIsNil() {
	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

	// No body.
	s.mock.AssertLastLogMessage(s.T(), "Failed to parse 1FA request body", "unable to parse body: unexpected end of JSON input")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorReauthenticateSuite) TestShouldFailIfBodyIsInBadFormat() {
	// Missing password.
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test"
	}`)
	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

	s.mock.AssertLastLogMessage(s.T(), "Failed to parse 1FA request body", "unable to validate body: password: non zero value required")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorReauthenticateSuite) TestShouldFailIfUserProviderCheckPasswordFail() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(false, fmt.Errorf("failed"))

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   testValue,
			Successful: false,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthType1FA,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.Ctx.Request.SetBodyString(`{
		"password": "hello"
	}`)
	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

	s.mock.AssertLastLogMessage(s.T(), "Unsuccessful 1FA authentication attempt by user 'test'", "failed")

	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorReauthenticateSuite) TestShouldCheckAuthenticationIsNotMarkedWhenProviderCheckPasswordError() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(false, fmt.Errorf("invalid credentials"))

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   testValue,
			Successful: false,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthType1FA,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.Ctx.Request.SetBodyString(`{
		"password": "hello"
	}`)

	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)
}

func (s *FirstFactorReauthenticateSuite) TestShouldCheckUserNotBanned() {
	s.mock.Ctx.Providers.Regulator = regulation.NewRegulator(schema.Regulation{MaxRetries: 2}, s.mock.StorageMock, &s.mock.Clock)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)

	gomock.InOrder(
		s.mock.UserProviderMock.
			EXPECT().
			CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
			Return(false, fmt.Errorf("invalid credentials")),

		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(s.mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
				Username:   testValue,
				Successful: false,
				Banned:     false,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthType1FA,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})))

	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)
}

func (s *FirstFactorReauthenticateSuite) TestShouldCheckBannedUser() {
	s.mock.Ctx.Providers.Regulator = regulation.NewRegulator(schema.Regulation{MaxRetries: 2, FindTime: time.Hour, BanTime: time.Hour}, s.mock.StorageMock, &s.mock.Clock)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)

	gomock.InOrder(
		/*
			s.mock.StorageMock.EXPECT().
				LoadAuthenticationLogs(gomock.Eq(s.mock.Ctx), testValue, gomock.Any(), gomock.Any(), gomock.Any()).
				Return([]model.AuthenticationAttempt{
					{Successful: false, Time: s.mock.Clock.Now().Add(-time.Second)},
					{Successful: false, Time: s.mock.Clock.Now().Add(-time.Second)},
					{Successful: false, Time: s.mock.Clock.Now().Add(-time.Second)},
					{Successful: false, Time: s.mock.Clock.Now().Add(-time.Second)},
				}, nil),
		*/

		s.mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(s.mock.Ctx), gomock.Eq(model.AuthenticationAttempt{
				Username:   testValue,
				Successful: false,
				Banned:     true,
				Time:       s.mock.Clock.Now(),
				Type:       regulation.AuthType1FA,
				RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
			})))

	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

	s.mock.AssertLastLogMessage(s.T(), "Unsuccessful 1FA authentication attempt by user 'test' and they are banned until 2013-02-03 00:59:59 +0000 UTC", "")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorReauthenticateSuite) TestShouldCheckAuthenticationIsMarkedWhenInvalidCredentials() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(false, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Eq(model.AuthenticationAttempt{
			Username:   testValue,
			Successful: false,
			Banned:     false,
			Time:       s.mock.Clock.Now(),
			Type:       regulation.AuthType1FA,
			RemoteIP:   model.NewNullIPFromString("0.0.0.0"),
		}))

	s.mock.Ctx.Request.SetBodyString(`{
		"password": "hello"
	}`)

	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)
}

func (s *FirstFactorReauthenticateSuite) TestShouldFailIfUserProviderGetDetailsFail() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq(testValue)).
		Return(nil, fmt.Errorf("failed"))

	s.mock.Ctx.Request.SetBodyString(`{
		"password": "hello"
	}`)
	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

	s.mock.AssertLastLogMessage(s.T(), "Could not obtain profile details during 1FA authentication for user 'test'", "failed")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorReauthenticateSuite) TestShouldFailIfAuthenticationMarkFail() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(fmt.Errorf("failed"))

	s.mock.Ctx.Request.SetBodyString(`{
		"password": "hello"
	}`)
	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

	s.mock.AssertLastLogMessage(s.T(), "Unable to mark 1FA authentication attempt by user 'test'", "failed")
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorReauthenticateSuite) TestShouldSaveUsernameFromAuthenticationBackendInSession() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq(testValue)).
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
		"password": "hello"
	}`)
	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	assert.Equal(s.T(), "test", userSession.Username)
	assert.Equal(s.T(), false, userSession.KeepMeLoggedIn)
	assert.Equal(s.T(), authentication.OneFactor, userSession.AuthenticationLevel(s.mock.Ctx.Configuration.WebAuthn.EnablePasskey2FA))
	assert.Equal(s.T(), []string{"test@example.com"}, userSession.Emails)
	assert.Equal(s.T(), []string{"dev", "admins"}, userSession.Groups)
}

type FirstFactorReauthenticateRedirectionSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *FirstFactorReauthenticateRedirectionSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())

	session, err := s.mock.Ctx.GetSession()

	s.Require().NoError(err)

	session.SetOneFactorPasskey(s.mock.Ctx.Clock.Now(), &authentication.UserDetails{Username: testValue}, false, false, false, false)

	s.Require().NoError(s.mock.Ctx.SaveSession(session))

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
		CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq(testValue)).
		Return(&authentication.UserDetails{
			Username: testValue,
			Emails:   []string{"test@example.com"},
			Groups:   []string{"dev", "admins"},
		}, nil)

	s.mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(s.mock.Ctx, gomock.Any()).
		Return(nil)
}

func (s *FirstFactorReauthenticateRedirectionSuite) TearDownTest() {
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
func (s *FirstFactorReauthenticateRedirectionSuite) TestShouldRedirectToDefaultURLWhenNoTargetURLProvidedAndTwoFactorDisabled() {
	s.mock.Ctx.Request.SetBodyString(`{
		"password": "hello"
	}`)
	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

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
func (s *FirstFactorReauthenticateRedirectionSuite) TestShouldRedirectToDefaultURLWhenURLIsUnsafeAndTwoFactorDisabled() {
	s.mock.Ctx.Request.SetBodyString(`{
		"password": "hello",
		"targetURL": "http://notsafe.local"
	}`)

	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

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
func (s *FirstFactorReauthenticateRedirectionSuite) TestShouldReply200WhenNoTargetURLProvidedAndTwoFactorEnabled() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControl{
			DefaultPolicy: "two_factor",
		},
	})
	s.mock.Ctx.Request.SetBodyString(`{
		"password": "hello"
	}`)

	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

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
func (s *FirstFactorReauthenticateRedirectionSuite) TestShouldReply200WhenUnsafeTargetURLProvidedAndTwoFactorEnabled() {
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
		"password": "hello"
	}`)

	FirstFactorReauthenticatePOST(nil)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
}

func TestFirstFactorSuite(t *testing.T) {
	suite.Run(t, new(FirstFactorSuite))
	suite.Run(t, new(FirstFactorRedirectionSuite))
}
