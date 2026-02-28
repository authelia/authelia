package handlers

import (
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/utils"
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
		domain := mock.Ctx.GetCookieDomainFromTargetURI(targetURI)

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

func (s *AuthzSuite) ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock *mocks.MockAutheliaCtx) {
	for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil, nil)
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

func (s *AuthzSuite) BuilderWithBearerScheme() (builder *AuthzBuilder) {
	switch s.implementation {
	case AuthzImplExtAuthz:
		return NewAuthzBuilder().WithImplementationExtAuthz().WithStrategies(NewHeaderProxyAuthorizationAuthnStrategy(time.Duration(0), model.AuthorizationSchemeBasic.String(), model.AuthorizationSchemeBearer.String()), NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplForwardAuth:
		return NewAuthzBuilder().WithImplementationForwardAuth().WithStrategies(NewHeaderProxyAuthorizationAuthnStrategy(time.Duration(0), model.AuthorizationSchemeBasic.String(), model.AuthorizationSchemeBearer.String()), NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplAuthRequest:
		return NewAuthzBuilder().WithImplementationAuthRequest().WithStrategies(NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Duration(0), model.AuthorizationSchemeBasic.String(), model.AuthorizationSchemeBearer.String()), NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplLegacy:
		return NewAuthzBuilder().WithImplementationLegacy().WithStrategies(NewHeaderLegacyAuthnStrategy(), NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	default:
		s.T().FailNow()
	}

	return nil
}

func (s *AuthzSuite) BuilderWithProxyAuthorizationBasicSchemeCached() (builder *AuthzBuilder) {
	switch s.implementation {
	case AuthzImplExtAuthz:
		return NewAuthzBuilder().WithImplementationExtAuthz().WithStrategies(NewHeaderProxyAuthorizationAuthnStrategy(time.Minute, model.AuthorizationSchemeBasic.String()), NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplForwardAuth:
		return NewAuthzBuilder().WithImplementationForwardAuth().WithStrategies(NewHeaderProxyAuthorizationAuthnStrategy(time.Minute, model.AuthorizationSchemeBasic.String()), NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplAuthRequest:
		return NewAuthzBuilder().WithImplementationAuthRequest().WithStrategies(NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Minute, model.AuthorizationSchemeBasic.String()), NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplLegacy:
		return NewAuthzBuilder().WithImplementationLegacy().WithStrategies(NewHeaderLegacyAuthnStrategy(), NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	default:
		s.T().FailNow()
	}

	return nil
}

func (s *AuthzSuite) TestShouldNotBeAbleToParseBasicAuth() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://test.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpaaaaaaaaaaaaaaaa")

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}
}

func (s *AuthzSuite) TestShouldApplyDefaultPolicy() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://test.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	switch s.implementation {
	case AuthzImplLegacy:
		break
	default:
		mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Clock.Now(),
			Successful:    true,
			Banned:        false,
			Username:      "john",
			Type:          regulation.AuthType1FA,
			RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
			RequestURI:    "https://test.example.com",
			RequestMethod: fasthttp.MethodGet,
		}

		mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil)
	}

	mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).Return(true, nil)

	mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("john")).Return(&authentication.UserDetails{Emails: []string{"john@example.com"}, Groups: []string{"dev", "admins"}}, nil)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldDenyObject() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	testCases := []struct {
		name  string
		value string
	}{
		{
			"NotProtected",
			"https://test.not-a-protected-domain.com",
		},
		{
			"Insecure",
			"http://test.example.com",
		},
	}

	authz := s.Builder().Build()

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

			targetURI := s.RequireParseRequestURI(tc.value)

			s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

			authz.Handler(mock.Ctx)

			switch s.implementation {
			case AuthzImplLegacy:
				assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
			default:
				assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
			}
		})
	}
}

func (s *AuthzSuite) TestShouldApplyPolicyOfBypassDomain() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	switch s.implementation {
	case AuthzImplLegacy:
		break
	default:
		mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Clock.Now(),
			Successful:    true,
			Banned:        false,
			Username:      "john",
			Type:          regulation.AuthType1FA,
			RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
			RequestURI:    "https://bypass.example.com",
			RequestMethod: fasthttp.MethodGet,
		}

		mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil)
	}

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

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldVerifyFailureToGetDetailsUsingBasicScheme() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	switch s.implementation {
	case AuthzImplLegacy:
		break
	default:
		mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Clock.Now(),
			Successful:    true,
			Banned:        false,
			Username:      "john",
			Type:          regulation.AuthType1FA,
			RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
			RequestURI:    "https://one-factor.example.com",
			RequestMethod: fasthttp.MethodGet,
		}

		mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil)
	}

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(nil, fmt.Errorf("generic failure"))

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}
}

func (s *AuthzSuite) TestShouldVerifyFailureToGetDetailsUsingBasicSchemeCached() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuilderWithProxyAuthorizationBasicSchemeCached().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Clock.Now(),
		Successful:    true,
		Banned:        false,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
		RequestURI:    "https://one-factor.example.com",
		RequestMethod: fasthttp.MethodGet,
	}

	if s.implementation == AuthzImplLegacy {
		gomock.InOrder(
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(true, nil),
			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(nil, fmt.Errorf("generic failure")),
		)
	} else {
		gomock.InOrder(
			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil),
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(true, nil),
			mock.StorageMock.
				EXPECT().
				AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil),
			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(nil, fmt.Errorf("generic failure")),
		)
	}

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}

	mock.Ctx.Request.Reset()
	mock.Ctx.Response.Reset()

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	if s.implementation == AuthzImplLegacy {
		gomock.InOrder(
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(true, nil),
			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(nil, fmt.Errorf("generic failure")),
		)
	} else {
		gomock.InOrder(
			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil),
			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(nil, fmt.Errorf("generic failure")),
		)
	}

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}
}

func (s *AuthzSuite) TestShouldVerifyFailureToCheckPasswordUsingBasicSchemeCached() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuilderWithProxyAuthorizationBasicSchemeCached().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Clock.Now(),
		Successful:    false,
		Banned:        false,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
		RequestURI:    "https://one-factor.example.com",
		RequestMethod: fasthttp.MethodGet,
	}

	if s.implementation == AuthzImplLegacy {
		gomock.InOrder(
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(false, nil),
		)
	} else {
		gomock.InOrder(
			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil),
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(false, nil),
			mock.StorageMock.
				EXPECT().
				AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil),
		)
	}

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}

	mock.Ctx.Request.Reset()
	mock.Ctx.Response.Reset()

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	if s.implementation == AuthzImplLegacy {
		gomock.InOrder(
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(false, nil),
		)
	} else {
		gomock.InOrder(
			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil),
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(false, nil),
			mock.StorageMock.
				EXPECT().
				AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil),
		)
	}

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}
}

func (s *AuthzSuite) TestShouldVerifyErrorToCheckPasswordUsingBasicSchemeCached() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuilderWithProxyAuthorizationBasicSchemeCached().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Clock.Now(),
		Successful:    false,
		Banned:        false,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
		RequestURI:    "https://one-factor.example.com",
		RequestMethod: fasthttp.MethodGet,
	}

	if s.implementation == AuthzImplLegacy {
		gomock.InOrder(
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(false, fmt.Errorf("bad data")),
		)
	} else {
		gomock.InOrder(
			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil),
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(false, fmt.Errorf("bad data")),
			mock.StorageMock.
				EXPECT().
				AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil),
		)
	}

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}

	mock.Ctx.Request.Reset()
	mock.Ctx.Response.Reset()

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	if s.implementation == AuthzImplLegacy {
		gomock.InOrder(
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(false, fmt.Errorf("bad data")),
		)
	} else {
		gomock.InOrder(
			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil),
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(false, fmt.Errorf("bad data")),
			mock.StorageMock.
				EXPECT().
				AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil),
		)
	}

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}
}

func (s *AuthzSuite) TestShouldVerifyBypassWithErrorToGetDetailsUsingBasicScheme() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	switch s.implementation {
	case AuthzImplLegacy:
		break
	default:
		mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Clock.Now(),
			Successful:    true,
			Banned:        false,
			Username:      "john",
			Type:          regulation.AuthType1FA,
			RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
			RequestURI:    "https://bypass.example.com",
			RequestMethod: fasthttp.MethodGet,
		}

		mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil)
	}

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().
			CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
			Return(true, nil),

		mock.UserProviderMock.EXPECT().
			GetDetails(gomock.Eq("john")).
			Return(nil, fmt.Errorf("generic failure")),
	)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func (s *AuthzSuite) TestShouldVerifyBypassWithErrorToGetDetailsUsingBearerScheme() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Bearer am9objpwYXNzd29yZA==")

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func (s *AuthzSuite) TestShouldVerifyBypassWithErrorToGetDetailsUsingBearerSchemePossibleToken() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Bearer authelia_at_aaaa.aaaaaa")

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func (s *AuthzSuite) TestShouldVerifyOneFactorWithErrorToGetDetailsUsingBearerScheme() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuilderWithBearerScheme().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Bearer am9objpwYXNzd29yZA==")

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplExtAuthz, AuthzImplForwardAuth:
		s.Equal(fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
	default:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	}
}

func (s *AuthzSuite) TestShouldVerifyOneFactorWithErrorToGetDetailsUsingBearerSchemePossibleToken() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuilderWithBearerScheme().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Bearer authelia_at_aaaa.aaaaaa")

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplExtAuthz, AuthzImplForwardAuth:
		s.Equal(fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
	default:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	}
}

func (s *AuthzSuite) TestShouldNotFailOnMissingEmail() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	authz := s.Builder().WithConfig(&mock.Ctx.Configuration).Build()

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = testUsername
	userSession.DisplayName = "John Smith"
	userSession.Groups = []string{"abc,123"}
	userSession.Emails = nil
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Equal(testUsername, string(mock.Ctx.Response.Header.PeekBytes(headerRemoteUser)))
	s.Equal("John Smith", string(mock.Ctx.Response.Header.PeekBytes(headerRemoteName)))
	s.Equal("abc,123", string(mock.Ctx.Response.Header.PeekBytes(headerRemoteGroups)))
}

func (s *AuthzSuite) TestShouldApplyPolicyOfOneFactorDomain() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	switch s.implementation {
	case AuthzImplLegacy:
		break
	default:
		mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Clock.Now(),
			Successful:    true,
			Banned:        false,
			Username:      "john",
			Type:          regulation.AuthType1FA,
			RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
			RequestURI:    "https://one-factor.example.com",
			RequestMethod: fasthttp.MethodGet,
		}

		mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil)
	}

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

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldApplyPolicyOfOneFactorDomainCached() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuilderWithProxyAuthorizationBasicSchemeCached().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	if s.implementation == AuthzImplLegacy {
		gomock.InOrder(
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(true, nil),
			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(&authentication.UserDetails{
					Emails: []string{"john@example.com"},
					Groups: []string{"dev", "admins"},
				}, nil),
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(true, nil),
			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(&authentication.UserDetails{
					Emails: []string{"john@example.com"},
					Groups: []string{"dev", "admins"},
				}, nil),
		)
	} else {
		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Clock.Now(),
			Successful:    true,
			Banned:        false,
			Username:      "john",
			Type:          regulation.AuthType1FA,
			RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
			RequestURI:    "https://one-factor.example.com",
			RequestMethod: fasthttp.MethodGet,
		}

		gomock.InOrder(
			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil),
			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(true, nil),
			mock.StorageMock.
				EXPECT().
				AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil),
			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(&authentication.UserDetails{
					Emails: []string{"john@example.com"},
					Groups: []string{"dev", "admins"},
				}, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil),
			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(&authentication.UserDetails{
					Emails: []string{"john@example.com"},
					Groups: []string{"dev", "admins"},
				}, nil),
		)
	}

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))

	mock.Ctx.Request.Reset()
	mock.Ctx.Response.Reset()

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldHandleAnyCaseSchemeParameter() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	testCases := []struct {
		name, scheme string
	}{
		{"Standard", "Basic"},
		{"LowerCase", "basic"},
		{"UpperCase", "BASIC"},
		{"MixedCase", "BaSIc"},
	}

	authz := s.Builder().Build()

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(s.T())

			defer mock.Close()

			setUpMockClock(mock)

			s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

			targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

			s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

			mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, fmt.Sprintf("%s am9objpwYXNzd29yZA==", tc.scheme))

			switch s.implementation {
			case AuthzImplLegacy:
				break
			default:
				mock.StorageMock.
					EXPECT().
					LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

				mock.StorageMock.
					EXPECT().
					LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

				attempt := model.AuthenticationAttempt{
					Time:          mock.Ctx.Clock.Now(),
					Successful:    true,
					Banned:        false,
					Username:      "john",
					Type:          regulation.AuthType1FA,
					RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
					RequestURI:    "https://one-factor.example.com",
					RequestMethod: fasthttp.MethodGet,
				}

				mock.StorageMock.
					EXPECT().
					AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil)
			}

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

			s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
			s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
			s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
		})
	}
}

func (s *AuthzSuite) TestShouldApplyPolicyOfTwoFactorDomain() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://two-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	switch s.implementation {
	case AuthzImplLegacy:
		break
	default:
		mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Clock.Now(),
			Successful:    true,
			Banned:        false,
			Username:      "john",
			Type:          regulation.AuthType1FA,
			RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
			RequestURI:    "https://two-factor.example.com",
			RequestMethod: fasthttp.MethodGet,
		}

		mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil)
	}

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
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
	default:
		s.Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
		s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
		s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate)))
	}
}

func (s *AuthzSuite) TestShouldApplyPolicyOfDenyDomain() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://deny.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	switch s.implementation {
	case AuthzImplLegacy:
		break
	default:
		mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Clock.Now(),
			Successful:    true,
			Banned:        false,
			Username:      "john",
			Type:          regulation.AuthType1FA,
			RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
			RequestURI:    "https://deny.example.com",
			RequestMethod: fasthttp.MethodGet,
		}

		mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil)
	}

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

	s.Equal(fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldApplyPolicyOfOneFactorDomainWithAuthorizationHeader() {
	if s.setRequest == nil || s.implementation == AuthzImplLegacy {
		s.T().Skip()
	}

	builder := NewAuthzBuilder().WithImplementationLegacy()

	builder = builder.WithStrategies(
		NewHeaderAuthorizationAuthnStrategy(time.Duration(0), "basic"),
		NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Duration(0), "basic"),
		NewCookieSessionAuthnStrategy(builder.config.RefreshInterval),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "Basic am9objpwYXNzd29yZA==")

	switch s.implementation {
	case AuthzImplLegacy:
		break
	default:
		mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Clock.Now(),
			Successful:    true,
			Banned:        false,
			Username:      "john",
			Type:          regulation.AuthType1FA,
			RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
			RequestURI:    "https://one-factor.example.com",
			RequestMethod: fasthttp.MethodGet,
		}

		switch s.implementation {
		case AuthzImplExtAuthz, AuthzImplForwardAuth:
			attempt.RequestURI += "/"
		}

		mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil)
	}

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

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldHandleAuthzWithoutHeaderNoCookie() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	// Equivalent of TestShouldVerifyAuthBasicArgFailingNoHeader.

	builder := NewAuthzBuilder().WithImplementationLegacy()

	builder = builder.WithStrategies(
		NewHeaderAuthorizationAuthnStrategy(time.Duration(0), "basic"),
		NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Duration(0), "basic"),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldHandleAuthzWithEmptyAuthorizationHeader() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	// Equivalent of TestShouldVerifyAuthBasicArgFailingEmptyHeader.

	builder := NewAuthzBuilder().WithImplementationLegacy()

	builder = builder.WithStrategies(
		NewHeaderAuthorizationAuthnStrategy(time.Duration(0), "basic"),
		NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Duration(0), "basic"),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "")

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldHandleAuthzWithAuthorizationHeaderInvalidPassword() {
	if s.setRequest == nil || s.implementation == AuthzImplLegacy {
		s.T().Skip()
	}

	builder := NewAuthzBuilder().WithImplementationLegacy()

	builder = builder.WithStrategies(
		NewHeaderAuthorizationAuthnStrategy(time.Duration(0), "basic"),
		NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Duration(0), "basic"),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "Basic am9objpwYXNzd29yZA==")

	switch s.implementation {
	case AuthzImplLegacy:
		break
	default:
		mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Clock.Now(),
			Successful:    false,
			Banned:        false,
			Username:      "john",
			Type:          regulation.AuthType1FA,
			RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
			RequestURI:    "https://one-factor.example.com",
			RequestMethod: fasthttp.MethodGet,
		}

		switch s.implementation {
		case AuthzImplExtAuthz, AuthzImplForwardAuth:
			attempt.RequestURI += "/"
		}

		mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).Return(nil)
	}

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(false, nil)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldHandleAuthzWithIncorrectAuthHeader() { // TestShouldVerifyAuthBasicArgFailingWrongHeader.
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewHeaderAuthorizationAuthnStrategy(time.Duration(0), "basic"),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	s.Equal(`Basic realm="Authorization Required"`, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldDestroySessionWhenInactiveForTooLong() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(testInactivity)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	past := mock.Clock.Now().Add(-1 * time.Hour)

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://two-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = testUsername
	userSession.AuthenticationMethodRefs.WebAuthn = true
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.LastActivity = past.Unix()

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	authz.Handler(mock.Ctx)

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal("", userSession.Username)
	s.Equal(authentication.NotAuthenticated, userSession.AuthenticationLevel(false))
	s.Equal(mock.Clock.Now().Unix(), userSession.LastActivity)
}

func (s *AuthzSuite) TestShouldNotDestroySessionWhenInactiveForTooLongRememberMe() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(testInactivity)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://two-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = testUsername
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.AuthenticationMethodRefs.WebAuthn = true
	userSession.LastActivity = 0
	userSession.KeepMeLoggedIn = true
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	authz.Handler(mock.Ctx)

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(testUsername, userSession.Username)
	s.Equal(authentication.TwoFactor, userSession.AuthenticationLevel(false))
	s.Equal(int64(0), userSession.LastActivity)
}

func (s *AuthzSuite) TestShouldNotDestroySessionWhenNotInactiveForTooLong() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(testInactivity)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://two-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	last := mock.Clock.Now().Add(-1 * time.Second)

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = testUsername
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.AuthenticationMethodRefs.WebAuthn = true
	userSession.LastActivity = last.Unix()
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	authz.Handler(mock.Ctx)

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(testUsername, userSession.Username)
	s.Equal(authentication.TwoFactor, userSession.AuthenticationLevel(false))
	s.Equal(mock.Clock.Now().Unix(), userSession.LastActivity)
}

func (s *AuthzSuite) TestShouldUpdateInactivityTimestampEvenWhenHittingForbiddenResources() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(testInactivity)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://deny.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	last := mock.Clock.Now().Add(-3 * time.Second)

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = testUsername
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.AuthenticationMethodRefs.WebAuthn = true
	userSession.LastActivity = last.Unix()
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	authz.Handler(mock.Ctx)

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(testUsername, userSession.Username)
	s.Equal(authentication.TwoFactor, userSession.AuthenticationLevel(false))
	s.Equal(mock.Clock.Now().Unix(), userSession.LastActivity)
}

func (s *AuthzSuite) TestShouldNotRefreshUserDetailsFromBackendWhenRefreshDisabled() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationNever()),
	)

	authz := builder.Build()

	user := &authentication.UserDetails{
		Username: "john",
		Groups: []string{
			"admin",
			"users",
		},
		Emails: []string{
			"john@example.com",
		},
	}

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	mock.Clock.Set(time.Now())

	mock.Ctx.Clock = &mock.Clock
	mock.Ctx.Configuration.AuthenticationBackend.RefreshInterval = schema.NewRefreshIntervalDurationNever()
	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://two-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = user.Username
	userSession.Groups = user.Groups
	userSession.Emails = user.Emails
	userSession.KeepMeLoggedIn = true
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.AuthenticationMethodRefs.WebAuthn = true
	userSession.LastActivity = mock.Clock.Now().Unix()

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	mock.UserProviderMock.EXPECT().GetDetails("john").Times(0)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	targetURI = s.RequireParseRequestURI("https://admin.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(user.Username, userSession.Username)
	s.Equal(authentication.TwoFactor, userSession.AuthenticationLevel(false))
	s.Equal(mock.Clock.Now().Unix(), userSession.LastActivity)
	s.Require().Len(userSession.Groups, 2)
	s.Equal("admin", userSession.Groups[0])
	s.Equal("users", userSession.Groups[1])
	s.Equal(utils.RFC3339Zero, userSession.RefreshTTL.Unix())

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(user.Username, userSession.Username)
	s.Equal(authentication.TwoFactor, userSession.AuthenticationLevel(false))
	s.Equal(mock.Clock.Now().Unix(), userSession.LastActivity)
	s.Require().Len(userSession.Groups, 2)
	s.Equal("admin", userSession.Groups[0])
	s.Equal("users", userSession.Groups[1])
	s.Equal(utils.RFC3339Zero, userSession.RefreshTTL.Unix())
}

func (s *AuthzSuite) TestShouldDestroySessionWhenUserDoesNotExist() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(5 * time.Minute)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://two-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	user := &authentication.UserDetails{
		Username: "john",
		Groups: []string{
			"admin",
			"users",
		},
		Emails: []string{
			"john@example.com",
		},
	}

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = user.Username
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.AuthenticationMethodRefs.WebAuthn = true
	userSession.LastActivity = mock.Clock.Now().Unix()
	userSession.RefreshTTL = mock.Clock.Now().Add(-1 * time.Minute)
	userSession.Groups = user.Groups
	userSession.Emails = user.Emails
	userSession.KeepMeLoggedIn = true

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(1),
		mock.UserProviderMock.EXPECT().GetDetails("john").Return(nil, authentication.ErrUserNotFound).Times(1),
	)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(mock.Clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())

	userSession.RefreshTTL = mock.Clock.Now().Add(-1 * time.Minute)

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	default:
		s.Equal(fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
	}

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal("", userSession.Username)
	s.Equal(authentication.NotAuthenticated, userSession.AuthenticationLevel(false))
	s.True(userSession.IsAnonymous())
}

func (s *AuthzSuite) TestShouldUpdateRemovedUserGroupsFromBackendAndDeny() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(5 * time.Minute)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://admin.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	user := &authentication.UserDetails{
		Username: "john",
		Groups: []string{
			"admin",
			"users",
		},
		Emails: []string{
			"john@example.com",
		},
	}

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = user.Username
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.AuthenticationMethodRefs.WebAuthn = true
	userSession.LastActivity = mock.Clock.Now().Unix()
	userSession.RefreshTTL = mock.Clock.Now().Add(-1 * time.Minute)
	userSession.Groups = user.Groups
	userSession.Emails = user.Emails
	userSession.KeepMeLoggedIn = true

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(1),
		mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(1),
	)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(mock.Clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	s.Require().Len(userSession.Groups, 2)
	s.Require().Equal("admin", userSession.Groups[0])
	s.Require().Equal("users", userSession.Groups[1])

	user.Groups = []string{"users"}

	mock.Clock.Set(mock.Clock.Now().Add(6 * time.Minute))

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(mock.Clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	s.Require().Len(userSession.Groups, 1)
	s.Require().Equal("users", userSession.Groups[0])
}

func (s *AuthzSuite) TestShouldUpdateAddedUserGroupsFromBackendAndDeny() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(5 * time.Minute)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://admin.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	user := &authentication.UserDetails{
		Username: "john",
		Groups: []string{
			"users",
		},
		Emails: []string{
			"john@example.com",
		},
	}

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = user.Username
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.AuthenticationMethodRefs.WebAuthn = true
	userSession.LastActivity = mock.Clock.Now().Unix()
	userSession.RefreshTTL = mock.Clock.Now().Add(-1 * time.Minute)
	userSession.Groups = user.Groups
	userSession.Emails = user.Emails
	userSession.KeepMeLoggedIn = true

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(1),
		mock.UserProviderMock.EXPECT().GetDetails("john").Return(user, nil).Times(1),
	)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(mock.Clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	s.Require().Len(userSession.Groups, 1)
	s.Require().Equal("users", userSession.Groups[0])

	user.Groups = []string{"admin", "users"}

	mock.Clock.Set(mock.Clock.Now().Add(6 * time.Minute))

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(mock.Clock.Now().Add(5*time.Minute).Unix(), userSession.RefreshTTL.Unix())
	s.Require().Len(userSession.Groups, 2)
	s.Require().Equal("admin", userSession.Groups[0])
	s.Require().Equal("users", userSession.Groups[1])
}

func (s *AuthzSuite) TestShouldCheckValidSessionUsernameHeaderAndReturn200() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(testInactivity)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.SetBytesK(headerSessionUsername, testUsername)

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = testUsername
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.LastActivity = mock.Clock.Now().Unix()
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal(testUsername, userSession.Username)
	s.Equal(authentication.OneFactor, userSession.AuthenticationLevel(false))
	s.Equal(mock.Clock.Now().Unix(), userSession.LastActivity)
}

func (s *AuthzSuite) TestShouldCheckInvalidSessionUsernameHeaderAndReturn401AndDestroySession() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(5 * time.Minute)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.SetBytesK(headerSessionUsername, "root")

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = testUsername
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.LastActivity = mock.Clock.Now().Unix()
	userSession.RefreshTTL = mock.Clock.Now().Add(5 * time.Minute)

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	default:
		s.Equal(fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
		location := s.RequireParseRequestURI(mock.Ctx.Configuration.Session.Cookies[0].AutheliaURL.String())

		if location.Path == "" {
			location.Path = "/"
		}

		query := location.Query()
		query.Set(queryArgRD, targetURI.String())
		query.Set(queryArgRM, fasthttp.MethodGet)

		location.RawQuery = query.Encode()

		s.Equal(location.String(), string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
	}

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal("", userSession.Username)
	s.Equal(authentication.NotAuthenticated, userSession.AuthenticationLevel(false))
	s.Equal(mock.Clock.Now().Unix(), userSession.LastActivity)
}

func (s *AuthzSuite) TestShouldNotRedirectRequestsForBypassACLWhenInactiveForTooLong() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(testInactivity)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	past := mock.Clock.Now().Add(-24 * time.Hour)

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	userSession, err := mock.Ctx.GetSession()
	s.Require().NoError(err)

	userSession.Username = testUsername
	userSession.AuthenticationMethodRefs.UsernameAndPassword = true
	userSession.AuthenticationMethodRefs.WebAuthn = true
	userSession.LastActivity = past.Unix()

	s.Require().NoError(mock.Ctx.SaveSession(userSession))

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	userSession, err = mock.Ctx.GetSession()
	s.Require().NoError(err)

	s.Equal("", userSession.Username)
	s.Equal(authentication.NotAuthenticated, userSession.AuthenticationLevel(false))
	s.Equal(mock.Clock.Now().Unix(), userSession.LastActivity)

	targetURI = s.RequireParseRequestURI("https://two-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	default:
		s.Equal(fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
		location := s.RequireParseRequestURI(mock.Ctx.Configuration.Session.Cookies[0].AutheliaURL.String())

		if location.Path == "" {
			location.Path = "/"
		}

		query := location.Query()
		query.Set(queryArgRD, targetURI.String())
		query.Set(queryArgRM, fasthttp.MethodGet)

		location.RawQuery = query.Encode()

		s.Equal(location.String(), string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
	}
}

func (s *AuthzSuite) TestShouldFailToParsePortalURL() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	builder := s.Builder()

	builder = builder.WithStrategies(
		NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDuration(testInactivity)),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	expected := fasthttp.StatusBadRequest

	switch s.implementation {
	case AuthzImplLegacy:
		expected = fasthttp.StatusUnauthorized

		mock.Ctx.RequestCtx.QueryArgs().Set(queryArgRD, "JKL$#N%KJ#@$N")
	case AuthzImplForwardAuth, AuthzImplAuthRequest:
		mock.Ctx.RequestCtx.QueryArgs().Set("authelia_url", "JKL$#N%KJ#@$N")
	case AuthzImplExtAuthz:
		mock.Ctx.Request.Header.Set("X-Authelia-URL", "JKL$#N%KJ#@$N")
	}

	authz.Handler(mock.Ctx)

	s.Equal(expected, mock.Ctx.Response.StatusCode())
	s.Equal(fmt.Sprintf("%d %s", expected, fasthttp.StatusMessage(expected)), string(mock.Ctx.Response.Body()))
	s.Equal("", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
	s.Equal("text/plain; charset=utf-8", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderContentType)))
}

func setRequestXHRValues(ctx *middlewares.AutheliaCtx, accept, xhr bool) {
	if accept {
		ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")
	}

	if xhr {
		ctx.Request.Header.Set(fasthttp.HeaderXRequestedWith, "XMLHttpRequest")
	}
}

type urlpair struct {
	TargetURI   *url.URL
	AutheliaURI *url.URL
}

func setUpMockClock(mock *mocks.MockAutheliaCtx) {
	mock.Ctx.Clock = &mock.Clock
	mock.Clock.Set(time.Now())
}
