package handlers

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/authelia/authelia/internal/mocks"
)

type OAuth2AuthSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *OAuth2AuthSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
}

func (s *OAuth2AuthSuite) TearDownTest() {
	s.mock.Close()
}

func (s *OAuth2AuthSuite) TestShouldReturn302() {
	// AuthEndpointGet(s.mock.Ctx)
	s.mock.Ctx.QueryArgs().Add("client_id", "authelia")
	s.mock.Ctx.QueryArgs().Add("response_type", "code")
	s.mock.Ctx.QueryArgs().Add("redirect_uri", "http://localhost:8080/oauth2/callback")
	s.mock.Ctx.QueryArgs().Add("scope", "openid")
	s.mock.Ctx.QueryArgs().Add("state", "random-string-here")
}

func TestRunOAuth2AuthSuite(t *testing.T) {
	suite.Run(t, new(OAuth2AuthSuite))
}
