package handlers

import (
	"strings"
	"testing"

	"github.com/clems4ever/authelia/mocks"

	"github.com/stretchr/testify/assert"

	"github.com/stretchr/testify/suite"
)

type LogoutSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *LogoutSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	userSession := s.mock.Ctx.GetSession()
	userSession.Username = "john"
	s.mock.Ctx.SaveSession(userSession)
}

func (s *LogoutSuite) TearDownTest() {
	s.mock.Close()
}

func (s *LogoutSuite) TestShouldDestroySession() {
	LogoutPost(s.mock.Ctx)
	b := s.mock.Ctx.Response.Header.PeekCookie("authelia_session")

	// Reset the cookie, meaning it resets the value and expires the cookie by setting
	// date to one minute in the past.
	assert.True(s.T(), strings.HasPrefix(string(b), "authelia_session=;"))
}

func TestRunLogoutSuite(t *testing.T) {
	s := new(LogoutSuite)
	suite.Run(t, s)
}
