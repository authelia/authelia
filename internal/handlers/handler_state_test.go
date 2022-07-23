package handlers

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/mocks"
)

type StateGetSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *StateGetSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	s.mock.Ctx.Request.Header.Set("X-Forwarded-Proto", "https")
	s.mock.Ctx.Request.Header.Set("X-Forwarded-Host", "home.example.com")
}

func (s *StateGetSuite) TearDownTest() {
	s.mock.Close()
}

func (s *StateGetSuite) TestShouldReturnUsernameFromSession() {
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = "username"
	err := s.mock.Ctx.SaveSession(userSession)
	require.NoError(s.T(), err)

	StateGET(s.mock.Ctx)

	type Response struct {
		Status string
		Data   StateResponse
	}

	expectedBody := Response{
		Status: "OK",
		Data: StateResponse{
			Username:              "username",
			DefaultRedirectionURL: "",
			AuthenticationLevel:   authentication.NotAuthenticated,
		},
	}
	actualBody := Response{}

	err = json.Unmarshal(s.mock.Ctx.Response.Body(), &actualBody)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("application/json; charset=utf-8"), s.mock.Ctx.Response.Header.ContentType())
	assert.Equal(s.T(), expectedBody, actualBody)
}

func (s *StateGetSuite) TestShouldReturnAuthenticationLevelFromSession() {
	userSession := s.mock.Ctx.GetSession()
	userSession.AuthenticationLevel = authentication.OneFactor
	err := s.mock.Ctx.SaveSession(userSession)
	require.NoError(s.T(), err)

	StateGET(s.mock.Ctx)

	type Response struct {
		Status string
		Data   StateResponse
	}

	expectedBody := Response{
		Status: "OK",
		Data: StateResponse{
			Username:              "",
			DefaultRedirectionURL: "",
			AuthenticationLevel:   authentication.OneFactor,
		},
	}
	actualBody := Response{}

	err = json.Unmarshal(s.mock.Ctx.Response.Body(), &actualBody)
	require.NoError(s.T(), err)
	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("application/json; charset=utf-8"), s.mock.Ctx.Response.Header.ContentType())
	assert.Equal(s.T(), expectedBody, actualBody)
}

func TestRunStateGetSuite(t *testing.T) {
	s := new(StateGetSuite)
	suite.Run(t, s)
}
