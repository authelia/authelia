package handlers

import (
	"net/url"

	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
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

type urlpair struct {
	TargetURI   *url.URL
	AutheliaURI *url.URL
}

func setRequestXHRValues(ctx *middlewares.AutheliaCtx, accept, xhr bool) {
	if accept {
		ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")
	}

	if xhr {
		ctx.Request.Header.Set(fasthttp.HeaderXRequestedWith, "XMLHttpRequest")
	}
}
