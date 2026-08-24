package handlers

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/v4/internal/mocks"
)

type LogoutSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *LogoutSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	provider, err := s.mock.Ctx.GetSessionProvider()
	s.Assert().NoError(err)

	userSession, err := provider.Get(s.mock.Ctx)
	s.Assert().NoError(err)

	userSession.Username = testUsername
	s.Assert().NoError(provider.Save(s.mock.Ctx, userSession))
}

func (s *LogoutSuite) TearDownTest() {
	s.mock.Close()
}

func (s *LogoutSuite) TestShouldDestroySession() {
	LogoutPOST(s.mock.Ctx)
	b := s.mock.Ctx.Response.Header.PeekCookie("authelia_session")

	// Reset the cookie, meaning it resets the value and expires the cookie by setting
	// date to one minute in the past.
	assert.True(s.T(), strings.HasPrefix(string(b), "authelia_session=;"))
}

func TestRunLogoutSuite(t *testing.T) {
	s := new(LogoutSuite)
	suite.Run(t, s)
}
