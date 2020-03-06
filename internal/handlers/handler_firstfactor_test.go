package handlers

import (
	"fmt"
	"testing"

	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/mocks"
	"github.com/authelia/authelia/internal/models"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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

func (s *FirstFactorSuite) assertError500(err string) {
	assert.Equal(s.T(), 500, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte(InternalError), s.mock.Ctx.Response.Body())
	assert.Equal(s.T(), err, s.mock.Hook.LastEntry().Message)
	assert.Equal(s.T(), logrus.ErrorLevel, s.mock.Hook.LastEntry().Level)
}

func (s *FirstFactorSuite) TestShouldFailIfBodyIsNil() {
	FirstFactorPost(s.mock.Ctx)

	// No body
	assert.Equal(s.T(), "Unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfBodyIsInBadFormat() {
	// Missing password
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test"
	}`)
	FirstFactorPost(s.mock.Ctx)

	assert.Equal(s.T(), "Unable to validate body: password: non zero value required", s.mock.Hook.LastEntry().Message)
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfUserProviderCheckPasswordFail() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(false, fmt.Errorf("Failed"))

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Eq(models.AuthenticationAttempt{
			Username:   "test",
			Successful: false,
			Time:       s.mock.Clock.Now(),
		}))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPost(s.mock.Ctx)

	assert.Equal(s.T(), "Error while checking password for user test: Failed", s.mock.Hook.LastEntry().Message)
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldCheckAuthenticationIsMarkedWhenInvalidCredentials() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(false, fmt.Errorf("Invalid credentials"))

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Eq(models.AuthenticationAttempt{
			Username:   "test",
			Successful: false,
			Time:       s.mock.Clock.Now(),
		}))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)

	FirstFactorPost(s.mock.Ctx)
}

func (s *FirstFactorSuite) TestShouldFailIfUserProviderGetDetailsFail() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(nil, fmt.Errorf("Failed"))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPost(s.mock.Ctx)

	assert.Equal(s.T(), "Error while retrieving details from user test: Failed", s.mock.Hook.LastEntry().Message)
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfAuthenticationMarkFail() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(fmt.Errorf("failed"))

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPost(s.mock.Ctx)

	assert.Equal(s.T(), "Unable to mark authentication: failed", s.mock.Hook.LastEntry().Message)
	s.mock.Assert200KO(s.T(), "Authentication failed. Check your credentials.")
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
			Emails: []string{"test@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPost(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	// And store authentication in session.
	session := s.mock.Ctx.GetSession()
	assert.Equal(s.T(), "test", session.Username)
	assert.Equal(s.T(), true, session.KeepMeLoggedIn)
	assert.Equal(s.T(), authentication.OneFactor, session.AuthenticationLevel)
	assert.Equal(s.T(), []string{"test@example.com"}, session.Emails)
	assert.Equal(s.T(), []string{"dev", "admins"}, session.Groups)
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
			Emails: []string{"test@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": false
	}`)
	FirstFactorPost(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	// And store authentication in session.
	session := s.mock.Ctx.GetSession()
	assert.Equal(s.T(), "test", session.Username)
	assert.Equal(s.T(), false, session.KeepMeLoggedIn)
	assert.Equal(s.T(), authentication.OneFactor, session.AuthenticationLevel)
	assert.Equal(s.T(), []string{"test@example.com"}, session.Emails)
	assert.Equal(s.T(), []string{"dev", "admins"}, session.Groups)
}

type FirstFactorRedirectionSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *FirstFactorRedirectionSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	s.mock.Ctx.Configuration.DefaultRedirectionURL = "https://default.local"
	s.mock.Ctx.Configuration.AccessControl.DefaultPolicy = "bypass"
	s.mock.Ctx.Configuration.AccessControl.Rules = []schema.ACLRule{
		schema.ACLRule{
			Domain: "default.local",
			Policy: "one_factor",
		},
	}
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(
		s.mock.Ctx.Configuration.AccessControl)

	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
		Return(true, nil)

	s.mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("test")).
		Return(&authentication.UserDetails{
			Emails: []string{"test@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(nil)
}

func (s *FirstFactorRedirectionSuite) TearDownTest() {
	s.mock.Close()
}

// When:
//   1/ the target url is unknown
//   2/ two_factor is disabled (no policy is set to two_factor)
//   3/ default_redirect_url is provided
// Then:
//   the user should be redirected to the default url.
func (s *FirstFactorRedirectionSuite) TestShouldRedirectToDefaultURLWhenNoTargetURLProvidedAndTwoFactorDisabled() {
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": false
	}`)
	FirstFactorPost(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), redirectResponse{Redirect: "https://default.local"})
}

// When:
//   1/ the target url is unsafe
//   2/ two_factor is disabled (no policy is set to two_factor)
//   3/ default_redirect_url is provided
// Then:
//   the user should be redirected to the default url.
func (s *FirstFactorRedirectionSuite) TestShouldRedirectToDefaultURLWhenURLIsUnsafeAndTwoFactorDisabled() {
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": false,
		"targetURL": "http://notsafe.local"
	}`)
	FirstFactorPost(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), redirectResponse{Redirect: "https://default.local"})
}

// When:
//   1/ two_factor is enabled (default policy)
// Then:
//   the user should receive 200 without redirection URL.
func (s *FirstFactorRedirectionSuite) TestShouldReply200WhenNoTargetURLProvidedAndTwoFactorEnabled() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(schema.AccessControlConfiguration{
		DefaultPolicy: "two_factor",
	})
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": false
	}`)
	FirstFactorPost(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
}

// When:
//   1/ two_factor is enabled (some rule)
// Then:
//   the user should receive 200 without redirection URL.
func (s *FirstFactorRedirectionSuite) TestShouldReply200WhenUnsafeTargetURLProvidedAndTwoFactorEnabled() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(schema.AccessControlConfiguration{
		DefaultPolicy: "one_factor",
		Rules: []schema.ACLRule{
			schema.ACLRule{
				Domain: "test.example.com",
				Policy: "one_factor",
			},
			schema.ACLRule{
				Domain: "example.com",
				Policy: "two_factor",
			},
		},
	})
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": false
	}`)
	FirstFactorPost(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
}

func TestFirstFactorSuite(t *testing.T) {
	suite.Run(t, new(FirstFactorSuite))
	suite.Run(t, new(FirstFactorRedirectionSuite))
}
