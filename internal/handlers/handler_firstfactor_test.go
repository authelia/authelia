package handlers

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/models"
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
	FirstFactorPost(0, false)(s.mock.Ctx)

	// No body
	assert.Equal(s.T(), "Unable to parse body: unexpected end of JSON input", s.mock.Hook.LastEntry().Message)
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
}

func (s *FirstFactorSuite) TestShouldFailIfBodyIsInBadFormat() {
	// Missing password
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test"
	}`)
	FirstFactorPost(0, false)(s.mock.Ctx)

	assert.Equal(s.T(), "Unable to validate body: password: non zero value required", s.mock.Hook.LastEntry().Message)
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
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
	FirstFactorPost(0, false)(s.mock.Ctx)

	assert.Equal(s.T(), "error while checking password for user test: Failed", s.mock.Hook.LastEntry().Message)
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
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

	FirstFactorPost(0, false)(s.mock.Ctx)
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
	FirstFactorPost(0, false)(s.mock.Ctx)

	assert.Equal(s.T(), "error while retrieving details from user test: Failed", s.mock.Hook.LastEntry().Message)
	s.mock.Assert401KO(s.T(), "Authentication failed. Check your credentials.")
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
	FirstFactorPost(0, false)(s.mock.Ctx)

	assert.Equal(s.T(), "unable to mark authentication: failed", s.mock.Hook.LastEntry().Message)
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

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPost(0, false)(s.mock.Ctx)

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
			Username: "test",
			Emails:   []string{"test@example.com"},
			Groups:   []string{"dev", "admins"},
		}, nil)

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false
	}`)
	FirstFactorPost(0, false)(s.mock.Ctx)

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

func (s *FirstFactorSuite) TestShouldSaveUsernameFromAuthenticationBackendInSession() {
	s.mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("test"), gomock.Eq("hello")).
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

	s.mock.StorageProviderMock.
		EXPECT().
		AppendAuthenticationLog(gomock.Any()).
		Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": true
	}`)
	FirstFactorPost(0, false)(s.mock.Ctx)

	// Respond with 200.
	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("{\"status\":\"OK\"}"), s.mock.Ctx.Response.Body())

	// And store authentication in session.
	session := s.mock.Ctx.GetSession()
	assert.Equal(s.T(), "Test", session.Username)
	assert.Equal(s.T(), true, session.KeepMeLoggedIn)
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
		"requestMethod": "GET",
		"keepMeLoggedIn": false
	}`)
	FirstFactorPost(0, false)(s.mock.Ctx)

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
		"requestMethod": "GET",
		"keepMeLoggedIn": false,
		"targetURL": "http://notsafe.local"
	}`)

	FirstFactorPost(0, false)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), redirectResponse{Redirect: "https://default.local"})
}

// When:
//   1/ two_factor is enabled (default policy)
// Then:
//   the user should receive 200 without redirection URL.
func (s *FirstFactorRedirectionSuite) TestShouldReply200WhenNoTargetURLProvidedAndTwoFactorEnabled() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControlConfiguration{
			DefaultPolicy: "two_factor",
		},
	})
	s.mock.Ctx.Request.SetBodyString(`{
		"username": "test",
		"password": "hello",
		"requestMethod": "GET",
		"keepMeLoggedIn": false
	}`)

	FirstFactorPost(0, false)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
}

// When:
//   1/ two_factor is enabled (some rule)
// Then:
//   the user should receive 200 without redirection URL.
func (s *FirstFactorRedirectionSuite) TestShouldReply200WhenUnsafeTargetURLProvidedAndTwoFactorEnabled() {
	s.mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
		AccessControl: schema.AccessControlConfiguration{
			DefaultPolicy: "one_factor",
			Rules: []schema.ACLRule{
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

	FirstFactorPost(0, false)(s.mock.Ctx)

	// Respond with 200.
	s.mock.Assert200OK(s.T(), nil)
}

func TestFirstFactorSuite(t *testing.T) {
	suite.Run(t, new(FirstFactorSuite))
	suite.Run(t, new(FirstFactorRedirectionSuite))
}

func TestFirstFactorDelayAverages(t *testing.T) {
	execDuration := time.Millisecond * 500
	oneSecond := time.Millisecond * 1000
	durations := []time.Duration{oneSecond, oneSecond, oneSecond, oneSecond, oneSecond, oneSecond, oneSecond, oneSecond, oneSecond, oneSecond}
	cursor := 0
	mutex := &sync.Mutex{}
	avgExecDuration := movingAverageIteration(execDuration, false, &cursor, &durations, mutex)
	assert.Equal(t, avgExecDuration, float64(1000))

	execDurations := []time.Duration{
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
		time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500, time.Millisecond * 500,
	}

	current := float64(1000)

	// Execute at 500ms for 12 requests.
	for _, execDuration = range execDurations {
		// Should not dip below 500, and should decrease in value by 50 each iteration.
		if current > 500 {
			current -= 50
		}

		avgExecDuration := movingAverageIteration(execDuration, true, &cursor, &durations, mutex)
		assert.Equal(t, avgExecDuration, current)
	}
}

func TestFirstFactorDelayCalculations(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	successful := false

	execDuration := 500 * time.Millisecond
	avgExecDurationMs := 1000.0
	expectedMinimumDelayMs := avgExecDurationMs - float64(execDuration.Milliseconds())

	for i := 0; i < 100; i++ {
		delay := calculateActualDelay(mock.Ctx, execDuration, avgExecDurationMs, &successful)
		assert.True(t, delay >= expectedMinimumDelayMs)
		assert.True(t, delay <= expectedMinimumDelayMs+float64(loginDelayMaximumRandomDelayMilliseconds))
	}

	execDuration = 5 * time.Millisecond
	avgExecDurationMs = 5.0
	expectedMinimumDelayMs = loginDelayMinimumDelayMilliseconds - float64(execDuration.Milliseconds())

	for i := 0; i < 100; i++ {
		delay := calculateActualDelay(mock.Ctx, execDuration, avgExecDurationMs, &successful)
		assert.True(t, delay >= expectedMinimumDelayMs)
		assert.True(t, delay <= expectedMinimumDelayMs+float64(loginDelayMaximumRandomDelayMilliseconds))
	}
}
