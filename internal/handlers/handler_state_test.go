package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

type StateGetSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *StateGetSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
}

func (s *StateGetSuite) TearDownTest() {
	s.mock.Close()
}

func (s *StateGetSuite) TestShouldReturnUsernameFromSession() {
	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	userSession.Username = "username"
	s.Assert().NoError(s.mock.Ctx.SaveSession(userSession))

	StateGET(s.mock.Ctx)

	type Response struct {
		Status string
		Data   StateResponse
	}

	expectedBody := Response{
		Status: "OK",
		Data: StateResponse{
			Username:              "username",
			DefaultRedirectionURL: "https://www.example.com",
			AuthenticationLevel:   authentication.NotAuthenticated,
		},
	}
	actualBody := Response{}

	err = json.Unmarshal(s.mock.Ctx.Response.Body(), &actualBody)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("application/json; charset=utf-8"), s.mock.Ctx.Response.Header.ContentType())
	assert.Equal(s.T(), expectedBody, actualBody)
}

func (s *StateGetSuite) TestShouldReturnAuthenticationLevelFromSession() {
	userSession, err := s.mock.Ctx.GetSession()
	s.Assert().NoError(err)

	userSession.Username = "john"
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	s.Assert().NoError(s.mock.Ctx.SaveSession(userSession))
	require.NoError(s.T(), err)

	StateGET(s.mock.Ctx)

	type Response struct {
		Status string
		Data   StateResponse
	}

	expectedBody := Response{
		Status: "OK",
		Data: StateResponse{
			Username:              "john",
			DefaultRedirectionURL: "https://www.example.com",
			AuthenticationLevel:   authentication.OneFactor,
			FactorKnowledge:       true,
		},
	}
	actualBody := Response{}

	err = json.Unmarshal(s.mock.Ctx.Response.Body(), &actualBody)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("application/json; charset=utf-8"), s.mock.Ctx.Response.Header.ContentType())
	assert.Equal(s.T(), expectedBody, actualBody)
}

func (s *StateGetSuite) TestShouldReturnForbiddenWhenSessionProviderUnavailable() {
	s.mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "unknown.example.org")

	StateGET(s.mock.Ctx)

	assert.Equal(s.T(), fasthttp.StatusForbidden, s.mock.Ctx.Response.StatusCode())

	require.NotNil(s.T(), s.mock.Hook.LastEntry())
	assert.Equal(s.T(), "Error occurred retrieving user session", s.mock.Hook.LastEntry().Message)
}

func (s *StateGetSuite) TestShouldOmitDefaultRedirectionURLWhenNotConfigured() {
	config := s.mock.Ctx.Configuration.Session

	config.Cookies[0].DefaultRedirectionURL = nil

	s.mock.Ctx.Providers.SessionProvider = session.NewProvider(config, nil)

	StateGET(s.mock.Ctx)

	type Response struct {
		Status string
		Data   StateResponse
	}

	actualBody := Response{}

	require.NoError(s.T(), json.Unmarshal(s.mock.Ctx.Response.Body(), &actualBody))

	assert.Equal(s.T(), fasthttp.StatusOK, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), "", actualBody.Data.DefaultRedirectionURL)
}

func TestRunStateGetSuite(t *testing.T) {
	s := new(StateGetSuite)
	suite.Run(t, s)
}
