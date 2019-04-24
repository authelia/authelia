package handlers

import (
	"encoding/json"
	"testing"

	"github.com/clems4ever/authelia/mocks"

	"github.com/clems4ever/authelia/authentication"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
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
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = "username"
	s.mock.Ctx.SaveSession(userSession)

	StateGet(s.mock.Ctx)

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

	json.Unmarshal(s.mock.Ctx.Response.Body(), &actualBody)
	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("application/json"), s.mock.Ctx.Response.Header.ContentType())
	assert.Equal(s.T(), expectedBody, actualBody)
}

func (s *StateGetSuite) TestShouldReturnAuthenticationLevelFromSession() {
	userSession := s.mock.Ctx.GetSession()
	userSession.AuthenticationLevel = authentication.OneFactor
	s.mock.Ctx.SaveSession(userSession)

	StateGet(s.mock.Ctx)

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

	json.Unmarshal(s.mock.Ctx.Response.Body(), &actualBody)
	assert.Equal(s.T(), 200, s.mock.Ctx.Response.StatusCode())
	assert.Equal(s.T(), []byte("application/json"), s.mock.Ctx.Response.Header.ContentType())
	assert.Equal(s.T(), expectedBody, actualBody)
}

func TestRunStateGetSuite(t *testing.T) {
	s := new(StateGetSuite)
	suite.Run(t, s)
}
