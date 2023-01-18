package handlers

import (
	"fmt"
	"net/url"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

type AuthzSuite struct {
	suite.Suite

	implementation AuthzImplementation
	builder        *AuthzBuilder
	setRequest     func(ctx *middlewares.AutheliaCtx, method string, targetURI *url.URL, accept bool, xhr bool)
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

func (s *AuthzSuite) Builder() (builder *AuthzBuilder) {
	if s.builder != nil {
		return s.builder
	}

	switch s.implementation {
	case AuthzImplExtAuthz:
		return NewAuthzBuilder().WithImplementationExtAuthz()
	case AuthzImplForwardAuth:
		return NewAuthzBuilder().WithImplementationForwardAuth()
	case AuthzImplAuthRequest:
		return NewAuthzBuilder().WithImplementationAuthRequest()
	case AuthzImplLegacy:
		return NewAuthzBuilder().WithImplementationLegacy()
	}

	s.T().FailNow()

	return
}

func (s *AuthzSuite) TestShouldNotBeAbleToParseBasicAuth() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://test.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpaaaaaaaaaaaaaaaa")

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Assert().Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Assert().Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Assert().Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Assert().Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}
}

func (s *AuthzSuite) TestShouldApplyDefaultPolicy() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://test.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	authz.Handler(mock.Ctx)

	s.Assert().Equal(fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldApplyPolicyOfBypassDomain() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	authz.Handler(mock.Ctx)

	s.Assert().Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldVerifyFailureToGetDetailsUsingBasicScheme() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(nil, fmt.Errorf("generic failure"))

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Assert().Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Assert().Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Assert().Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Assert().Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}
}

func (s *AuthzSuite) TestShouldNotFailOnMissingEmail() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	mock.Clock.Set(time.Now())

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	userSession := mock.Ctx.GetSession()
	userSession.Username = testUsername
	userSession.DisplayName = "John Smith"
	userSession.Groups = []string{"abc,123"}
	userSession.Emails = nil
	userSession.AuthenticationLevel = authentication.OneFactor
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	authz.Handler(mock.Ctx)

	s.Assert().Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Assert().Equal(testUsername, string(mock.Ctx.Response.Header.PeekBytes(headerRemoteUser)))
	s.Assert().Equal("John Smith", string(mock.Ctx.Response.Header.PeekBytes(headerRemoteName)))
	s.Assert().Equal("abc,123", string(mock.Ctx.Response.Header.PeekBytes(headerRemoteGroups)))
}

func (s *AuthzSuite) TestShouldApplyPolicyOfOneFactorDomain() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	authz.Handler(mock.Ctx)

	s.Assert().Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldApplyPolicyOfTwoFactorDomain() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://two-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Assert().Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Assert().Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Assert().Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Assert().Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}
}

func (s *AuthzSuite) TestShouldApplyPolicyOfDenyDomain() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://deny.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	authz.Handler(mock.Ctx)

	s.Assert().Equal(fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldApplyPolicyOfOneFactorDomainWithAuthorizationHeader() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	// Equivalent of TestShouldVerifyAuthBasicArgOk.

	builder := NewAuthzBuilder().WithImplementationLegacy()

	builder = builder.WithStrategies(
		NewHeaderAuthorizationAuthnStrategy(),
		NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(),
		NewCookieSessionAuthnStrategy(builder.config.RefreshInterval),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Emails: []string{"john@example.com"},
			Groups: []string{"dev", "admins"},
		}, nil)

	authz.Handler(mock.Ctx)

	s.Assert().Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldHandleAuthzWithoutHeaderNoCookie() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	// Equivalent of TestShouldVerifyAuthBasicArgFailingNoHeader.

	builder := NewAuthzBuilder().WithImplementationLegacy()

	builder = builder.WithStrategies(
		NewHeaderAuthorizationAuthnStrategy(),
		NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	authz.Handler(mock.Ctx)

	s.Assert().Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	s.Assert().Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldHandleAuthzWithEmptyAuthorizationHeader() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	// Equivalent of TestShouldVerifyAuthBasicArgFailingEmptyHeader.

	builder := NewAuthzBuilder().WithImplementationLegacy()

	builder = builder.WithStrategies(
		NewHeaderAuthorizationAuthnStrategy(),
		NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "")

	authz.Handler(mock.Ctx)

	s.Assert().Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	s.Assert().Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldHandleAuthzWithAuthorizationHeaderInvalidPassword() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	// Equivalent of TestShouldVerifyAuthBasicArgFailingWrongPassword.

	builder := NewAuthzBuilder().WithImplementationLegacy()

	builder = builder.WithStrategies(
		NewHeaderAuthorizationAuthnStrategy(),
		NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(false, nil)

	authz.Handler(mock.Ctx)

	s.Assert().Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	s.Assert().Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldHandleAuthzWithIncorrectAuthHeader() { // TestShouldVerifyAuthBasicArgFailingWrongHeader.
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewHeaderAuthorizationAuthnStrategy(),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	authz.Handler(mock.Ctx)

	s.Assert().Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	s.Assert().Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
	s.Assert().Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func setRequestXHRValues(ctx *middlewares.AutheliaCtx, accept, xhr bool) {
	if accept {
		ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")
	}

	if xhr {
		ctx.Request.Header.Set(fasthttp.HeaderXRequestedWith, "XMLHttpRequest")
	}
}
