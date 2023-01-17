package handlers

import (
	"net/url"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/stretchr/testify/suite"
)

type AuthzSuite struct {
	suite.Suite

	builder *AuthzBuilder
}

func (s *AuthzSuite) GetMock(config *schema.Configuration, targetURI *url.URL, session *session.UserSession) *mocks.MockAutheliaCtx {
	mock := mocks.NewMockAutheliaCtx(s.T())

	if session != nil {
		domain := mock.Ctx.GetTargetURICookieDomain(targetURI)

		provider, err := mock.Ctx.GetCookieDomainSessionProvider(domain)
		s.Require().NoError(err)

		s.Require().NoError(provider.SaveSession(mock.Ctx.RequestCtx, *session))
	}

	return mock
}

func (s *AuthzSuite) RequireParseRequestURI(rawURL string) *url.URL {
	u, err := url.ParseRequestURI(rawURL)

	s.Require().NoError(err)

	return u
}
