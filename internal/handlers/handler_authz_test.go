package handlers

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	fjwt "authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/regulation"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
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

func (s *AuthzSuite) ConfigureMockSessionProviderWithoutAutheliaURLs(mock *mocks.MockAutheliaCtx) {
	for i := range mock.Ctx.Configuration.Session.Cookies {
		mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = nil
	}

	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)
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

func (s *AuthzSuite) BuildWithDelayer() (authz *Authz) {
	authz = s.Builder().Build()

	s.ApplyTestDelayer(authz)

	return authz
}

func (s *AuthzSuite) ApplyTestDelayer(authz *Authz) {
	for i, v := range authz.strategies {
		switch strategy := v.(type) {
		case *HeaderAuthnStrategy:
			strategy.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

			authz.strategies[i] = strategy
		case *HeaderLegacyAuthnStrategy:
			strategy.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

			authz.strategies[i] = strategy
		}
	}
}

func (s *AuthzSuite) BuilderWithBearerScheme() (builder *AuthzBuilder) {
	proxyHeader := NewHeaderProxyAuthorizationAuthnStrategy(time.Duration(0), model.AuthorizationSchemeBasic.String(), model.AuthorizationSchemeBearer.String())
	proxyHeader.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	proxyAuthHeader := NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Duration(0), model.AuthorizationSchemeBasic.String(), model.AuthorizationSchemeBearer.String())
	proxyAuthHeader.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	legacyHeader := NewHeaderLegacyAuthnStrategy()
	legacyHeader.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	switch s.implementation {
	case AuthzImplExtAuthz:
		return NewAuthzBuilder().WithImplementationExtAuthz().WithStrategies(proxyHeader, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplForwardAuth:
		return NewAuthzBuilder().WithImplementationForwardAuth().WithStrategies(proxyHeader, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplAuthRequest:
		return NewAuthzBuilder().WithImplementationAuthRequest().WithStrategies(proxyAuthHeader, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplLegacy:
		return NewAuthzBuilder().WithImplementationLegacy().WithStrategies(legacyHeader, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	default:
		s.T().FailNow()
	}

	return nil
}

func (s *AuthzSuite) BuilderWithDPoPScheme() (builder *AuthzBuilder) {
	header := NewHeaderAuthorizationAuthnStrategy(time.Duration(0), model.AuthorizationSchemeBasic.String(), model.AuthorizationSchemeBearer.String(), model.AuthorizationSchemeDPoP.String())
	header.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	switch s.implementation {
	case AuthzImplExtAuthz:
		return NewAuthzBuilder().WithImplementationExtAuthz().WithStrategies(header, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplForwardAuth:
		return NewAuthzBuilder().WithImplementationForwardAuth().WithStrategies(header, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplAuthRequest:
		return NewAuthzBuilder().WithImplementationAuthRequest().WithStrategies(header, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	default:
		s.T().FailNow()
	}

	return nil
}

func (s *AuthzSuite) BuilderWithProxyAuthorizationBasicSchemeCached() (builder *AuthzBuilder) {
	proxyHeader := NewHeaderProxyAuthorizationAuthnStrategy(time.Minute, model.AuthorizationSchemeBasic.String(), model.AuthorizationSchemeBearer.String())
	proxyHeader.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	proxyAuthHeader := NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Minute, model.AuthorizationSchemeBasic.String(), model.AuthorizationSchemeBearer.String())
	proxyAuthHeader.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	legacyHeader := NewHeaderLegacyAuthnStrategy()
	legacyHeader.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	switch s.implementation {
	case AuthzImplExtAuthz:
		return NewAuthzBuilder().WithImplementationExtAuthz().WithStrategies(proxyHeader, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplForwardAuth:
		return NewAuthzBuilder().WithImplementationForwardAuth().WithStrategies(proxyHeader, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplAuthRequest:
		return NewAuthzBuilder().WithImplementationAuthRequest().WithStrategies(proxyAuthHeader, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	case AuthzImplLegacy:
		return NewAuthzBuilder().WithImplementationLegacy().WithStrategies(legacyHeader, NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways()))
	default:
		s.T().FailNow()
	}

	return nil
}

func (s *AuthzSuite) TestShouldNotBeAbleToParseBasicAuth() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

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

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://test.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

	mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
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

	mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).Return(true, nil)

	mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("john")).Return(&authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"dev", "admins"}}, nil)

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

	authz := s.BuildWithDelayer()

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

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

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

	mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
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

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Username: "john",
			Emails:   []string{"john@example.com"},
			Groups:   []string{"dev", "admins"},
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

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

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

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().
			GetDetails(gomock.Eq("john")).
			Return(nil, fmt.Errorf("generic failure")),
	)

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

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().
			GetDetails(gomock.Eq("john")).
			Return(nil, fmt.Errorf("generic failure")),
	)

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

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
		Successful:    false,
		Banned:        false,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
		RequestURI:    "https://one-factor.example.com",
		RequestMethod: fasthttp.MethodGet,
	}

	gomock.InOrder(
		mock.UserProviderMock.
			EXPECT().
			GetDetails(gomock.Eq("john")).Return(&authentication.UserDetails{Username: "john"}, nil),
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

	gomock.InOrder(
		mock.UserProviderMock.
			EXPECT().
			GetDetails(gomock.Eq("john")).Return(&authentication.UserDetails{Username: "john"}, nil),
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

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
		Successful:    false,
		Banned:        false,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
		RequestURI:    "https://one-factor.example.com",
		RequestMethod: fasthttp.MethodGet,
	}

	gomock.InOrder(
		mock.UserProviderMock.
			EXPECT().
			GetDetails(gomock.Eq("john")).Return(&authentication.UserDetails{Username: "john"}, nil),
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

	gomock.InOrder(
		mock.UserProviderMock.
			EXPECT().
			GetDetails(gomock.Eq("john")).Return(&authentication.UserDetails{Username: "john"}, nil),
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

func (s *AuthzSuite) TestShouldRejectBannedUserUsingBasicScheme() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	expires := mock.Ctx.Providers.Clock.Now().Add(time.Minute)

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
		Successful:    false,
		Banned:        true,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
		RequestURI:    "https://one-factor.example.com",
		RequestMethod: fasthttp.MethodGet,
	}

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().
			GetDetails(gomock.Eq("john")).
			Return(&authentication.UserDetails{Username: "john"}, nil),
		mock.StorageMock.EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
			Return(nil, nil),
		mock.StorageMock.EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).
			Return([]model.BannedUser{{ID: 1, Username: "john", Expires: sql.NullTime{Time: expires, Valid: true}}}, nil),
		mock.StorageMock.EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).
			Return(nil),
	)

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

func (s *AuthzSuite) TestShouldRejectBannedIPUsingBasicScheme() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	expires := mock.Ctx.Providers.Clock.Now().Add(time.Minute)

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
		Successful:    false,
		Banned:        true,
		Username:      model.NewIP(mock.Ctx.RemoteIP()).String(),
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
		RequestURI:    "https://one-factor.example.com",
		RequestMethod: fasthttp.MethodGet,
	}

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().
			GetDetails(gomock.Eq("john")).
			Return(&authentication.UserDetails{Username: "john"}, nil),
		mock.StorageMock.EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
			Return([]model.BannedIP{{ID: 1, IP: model.NewIP(mock.Ctx.RemoteIP()), Expires: sql.NullTime{Time: expires, Valid: true}}}, nil),
		mock.StorageMock.EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).
			Return(nil),
	)

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

func (s *AuthzSuite) TestShouldRejectBannedCanonicalUserUsingBasicScheme() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic Sk9ITjpwYXNzd29yZA==")

	expires := mock.Ctx.Providers.Clock.Now().Add(time.Minute)

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
		Successful:    false,
		Banned:        true,
		Username:      "john",
		Type:          regulation.AuthType1FA,
		RemoteIP:      model.NewNullIP(mock.Ctx.RemoteIP()),
		RequestURI:    "https://one-factor.example.com",
		RequestMethod: fasthttp.MethodGet,
	}

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().
			GetDetails(gomock.Eq("JOHN")).
			Return(&authentication.UserDetails{Username: "john"}, nil),
		mock.StorageMock.EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
			Return(nil, nil),
		mock.StorageMock.EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).
			Return([]model.BannedUser{{ID: 1, Username: "john", Expires: sql.NullTime{Time: expires, Valid: true}}}, nil),
		mock.StorageMock.EXPECT().
			AppendAuthenticationLog(gomock.Eq(mock.Ctx), gomock.Eq(attempt)).
			Return(nil),
	)

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

func (s *AuthzSuite) TestShouldHandleBanCheckStorageErrorUsingBasicScheme() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	gomock.InOrder(
		mock.UserProviderMock.EXPECT().
			GetDetails(gomock.Eq("john")).
			Return(&authentication.UserDetails{Username: "john"}, nil),
		mock.StorageMock.EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
			Return(nil, fmt.Errorf("database unreachable")),
	)

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

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://bypass.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq("john")).Return(nil, fmt.Errorf("generic failure"))

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func (s *AuthzSuite) TestShouldVerifyBypassWithErrorToGetDetailsUsingBearerScheme() {
	if s.setRequest == nil {
		s.T().Skip()
	}

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

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

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

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

func (s *AuthzSuite) TestShouldAuthenticateAsClientUsingBearerSchemeClientCredentials() {
	if s.setRequest == nil || s.implementation == AuthzImplLegacy {
		s.T().Skip()
	}

	authz := s.BuilderWithBearerScheme().Build()

	s.ApplyTestDelayer(authz)

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	audience := []string{"https://one-factor.example.com", "https://one-factor.example.com/"}

	mock.Ctx.Configuration.IdentityProviders = schema.IdentityProviders{
		OIDC: &schema.IdentityProvidersOpenIDConnect{
			HMACSecret: "abcdefghijklmnopqrstuvwxyz123456",
			Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
				BearerAuthorization: true,
			},
			Clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "test-ccs-client",
					Scopes:              []string{oidc.ScopeAutheliaBearerAuthz},
					Audience:            audience,
					GrantTypes:          []string{oidc.GrantTypeClientCredentials},
					AuthorizationPolicy: "one_factor",
				},
			},
		},
	}

	mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(&mock.Ctx.Configuration, mock.StorageMock, mock.Ctx.Providers.Templates)

	client, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, "test-ccs-client")
	s.Require().NoError(err)

	now := mock.Ctx.Providers.Clock.Now()

	session := &oidc.Session{
		ClientID:          "test-ccs-client",
		ClientCredentials: true,
		DefaultSession: &openid.DefaultSession{
			Headers: &fjwt.Headers{Extra: map[string]any{}},
			Claims: &fjwt.IDTokenClaims{
				Issuer:   "https://auth.example.com",
				Subject:  "test-ccs-client",
				IssuedAt: fjwt.NewNumericDate(now),
				Extra:    map[string]any{},
			},
			RequestedAt: now,
		},
	}

	requester := &oauthelia2.AccessRequest{
		GrantTypes: oauthelia2.Arguments{oidc.GrantTypeClientCredentials},
		Request: oauthelia2.Request{
			ID:                "request-ccs",
			RequestedAt:       now,
			Client:            client,
			RequestedScope:    oauthelia2.Arguments{oidc.ScopeAutheliaBearerAuthz},
			GrantedScope:      oauthelia2.Arguments{oidc.ScopeAutheliaBearerAuthz},
			RequestedAudience: oauthelia2.Arguments(audience),
			GrantedAudience:   oauthelia2.Arguments(audience),
			Session:           session,
			Form:              url.Values{},
		},
	}

	token, signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.GenerateAccessToken(mock.Ctx, requester)
	s.Require().NoError(err)

	oauthSession, err := model.NewOAuth2SessionFromRequest(signature, requester)
	s.Require().NoError(err)

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Bearer "+token)

	mock.StorageMock.EXPECT().
		LoadOAuth2Session(gomock.Eq(mock.Ctx), gomock.Eq(storage.OAuth2SessionTypeAccessToken), gomock.Eq(signature)).
		Return(oauthSession, nil)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
	s.Equal([]byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderProxyAuthenticate))
}

func (s *AuthzSuite) TestShouldValidateDPoPProofsUsingBearerAuthorization() {
	if s.setRequest == nil || s.implementation == AuthzImplLegacy {
		s.T().Skip()
	}

	const (
		schemeBearer = "Bearer"
		targetURI    = "https://one-factor.example.com/api"
		unusableJKT  = "D6Nq0uHi1xL9fbLBu6xVGvKtOsBqiOxfHy_hOZlLzHM"
	)

	testCases := []struct {
		Name              string
		Bound             bool
		Scheme            string
		Proof             bool
		Duplicate         bool
		NonceEnforced     bool
		ExpectedStatus    int
		ExpectedChallenge string
		ExpectedNonce     bool
	}{
		{
			Name:              "ShouldRejectBoundTokenPresentedAsBearer",
			Bound:             true,
			Scheme:            schemeBearer,
			ExpectedStatus:    fasthttp.StatusUnauthorized,
			ExpectedChallenge: oidc.SchemeDPoP + " ",
		},
		{
			Name:              "ShouldRejectBoundTokenPresentedAsBearerWithValidProof",
			Bound:             true,
			Scheme:            schemeBearer,
			Proof:             true,
			ExpectedStatus:    fasthttp.StatusUnauthorized,
			ExpectedChallenge: oidc.SchemeDPoP + " ",
		},
		{
			Name:              "ShouldRejectBoundTokenPresentedAsDPoPWithDuplicateProofHeaders",
			Bound:             true,
			Scheme:            oidc.SchemeDPoP,
			Proof:             true,
			Duplicate:         true,
			ExpectedStatus:    fasthttp.StatusUnauthorized,
			ExpectedChallenge: oidc.SchemeDPoP + " ",
		},
		{
			Name:              "ShouldRejectUnboundTokenPresentedAsDPoP",
			Scheme:            oidc.SchemeDPoP,
			ExpectedStatus:    fasthttp.StatusUnauthorized,
			ExpectedChallenge: oidc.SchemeDPoP + " ",
		},
		{
			Name:              "ShouldRejectBoundTokenPresentedAsDPoPWithoutProof",
			Bound:             true,
			Scheme:            oidc.SchemeDPoP,
			ExpectedStatus:    fasthttp.StatusUnauthorized,
			ExpectedChallenge: oidc.SchemeDPoP + " ",
		},
		{
			Name:              "ShouldChallengeBoundTokenPresentedAsDPoPWithProofMissingNonce",
			Bound:             true,
			Scheme:            oidc.SchemeDPoP,
			Proof:             true,
			NonceEnforced:     true,
			ExpectedStatus:    fasthttp.StatusUnauthorized,
			ExpectedChallenge: oidc.SchemeDPoP + " ",
			ExpectedNonce:     true,
		},
		{
			Name:           "ShouldAllowBoundTokenPresentedAsDPoPWithValidProof",
			Bound:          true,
			Scheme:         oidc.SchemeDPoP,
			Proof:          true,
			ExpectedStatus: fasthttp.StatusOK,
		},
		{
			Name:           "ShouldAllowUnboundTokenPresentedAsBearer",
			Scheme:         schemeBearer,
			ExpectedStatus: fasthttp.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.Name, func(t *testing.T) {
			authz := s.BuilderWithDPoPScheme().Build()

			s.ApplyTestDelayer(authz)

			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			setUpMockClock(mock)

			target := s.RequireParseRequestURI(targetURI)

			audience := []string{targetURI, targetURI + "/"}

			mock.Ctx.Configuration.IdentityProviders = schema.IdentityProviders{
				OIDC: &schema.IdentityProvidersOpenIDConnect{
					HMACSecret: "abcdefghijklmnopqrstuvwxyz123456",
					Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
						BearerAuthorization: true,
					},
					DPoP: schema.IdentityProvidersOpenIDConnectDPoP{
						Enabled:       true,
						ClockSkew:     time.Minute,
						NonceEnforced: tc.NonceEnforced,
						NonceLifespan: time.Minute,
					},
					Clients: []schema.IdentityProvidersOpenIDConnectClient{
						{
							ID:                  "test-dpop-client",
							Scopes:              []string{oidc.ScopeAutheliaBearerAuthz},
							Audience:            audience,
							GrantTypes:          []string{oidc.GrantTypeClientCredentials},
							AuthorizationPolicy: "one_factor",
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(&mock.Ctx.Configuration, mock.StorageMock, mock.Ctx.Providers.Templates)

			client, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, "test-dpop-client")
			require.NoError(t, err)

			var (
				key *ecdsa.PrivateKey
				jkt string
			)

			switch {
			case tc.Proof:
				key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
				require.NoError(t, err)

				jkt, err = fjwt.ThumbprintJWK(&jose.JSONWebKey{Key: key.Public()})
				require.NoError(t, err)
			case tc.Bound:
				jkt = unusableJKT
			}

			now := mock.Ctx.Providers.Clock.Now()

			oidcSession := &oidc.Session{
				ClientID:          "test-dpop-client",
				ClientCredentials: true,
				DPoPJWKThumbprint: jkt,
				DefaultSession: &openid.DefaultSession{
					Headers: &fjwt.Headers{Extra: map[string]any{}},
					Claims: &fjwt.IDTokenClaims{
						Issuer:   "https://auth.example.com",
						Subject:  "test-dpop-client",
						IssuedAt: fjwt.NewNumericDate(now),
						Extra:    map[string]any{},
					},
					RequestedAt: now,
				},
			}

			requester := &oauthelia2.AccessRequest{
				GrantTypes: oauthelia2.Arguments{oidc.GrantTypeClientCredentials},
				Request: oauthelia2.Request{
					ID:                "request-dpop",
					RequestedAt:       now,
					Client:            client,
					RequestedScope:    oauthelia2.Arguments{oidc.ScopeAutheliaBearerAuthz},
					GrantedScope:      oauthelia2.Arguments{oidc.ScopeAutheliaBearerAuthz},
					RequestedAudience: oauthelia2.Arguments(audience),
					GrantedAudience:   oauthelia2.Arguments(audience),
					Session:           oidcSession,
					Form:              url.Values{},
				},
			}

			token, signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.GenerateAccessToken(mock.Ctx, requester)
			require.NoError(t, err)

			oauthSession, err := model.NewOAuth2SessionFromRequest(signature, requester)
			require.NoError(t, err)

			s.setRequest(mock.Ctx, fasthttp.MethodGet, target, true, false)

			mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, tc.Scheme+" "+token)

			if tc.Proof {
				proof := newTestDPoPProof(t, key, fasthttp.MethodGet, targetURI, token)

				mock.Ctx.Request.Header.Set(oidc.HeaderDPoP, proof)

				if tc.Duplicate {
					mock.Ctx.Request.Header.Add(oidc.HeaderDPoP, proof)
				}
			}

			if tc.Proof && tc.Scheme == oidc.SchemeDPoP && !tc.Duplicate {
				if tc.NonceEnforced {
					mock.StorageMock.EXPECT().
						SaveOAuth2DPoPNonce(gomock.Any(), gomock.Any()).
						Return(nil)
				} else {
					mock.StorageMock.EXPECT().
						CheckAndSetOAuth2DPoPProofUsed(gomock.Any(), gomock.Any(), gomock.Eq(fasthttp.MethodGet), gomock.Eq(targetURI), gomock.Any(), gomock.Any()).
						Return(false, nil)
				}
			}

			mock.StorageMock.EXPECT().
				LoadOAuth2Session(gomock.Eq(mock.Ctx), gomock.Eq(storage.OAuth2SessionTypeAccessToken), gomock.Eq(signature)).
				Return(oauthSession, nil)

			authz.Handler(mock.Ctx)

			assert.Equal(t, tc.ExpectedStatus, mock.Ctx.Response.StatusCode())

			challenge := string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))

			if tc.ExpectedChallenge == "" {
				assert.Equal(t, "", challenge)
			} else {
				assert.True(t, strings.HasPrefix(challenge, tc.ExpectedChallenge), challenge)
			}

			if tc.ExpectedNonce {
				assert.NotEmpty(t, string(mock.Ctx.Response.Header.Peek(oidc.HeaderDPoPNonce)))
			} else {
				assert.Empty(t, string(mock.Ctx.Response.Header.Peek(oidc.HeaderDPoPNonce)))
			}
		})
	}
}

func (s *AuthzSuite) TestShouldRejectBearerSchemeClientCredentialsWithoutBearerAuthzScope() {
	if s.setRequest == nil || s.implementation == AuthzImplLegacy {
		s.T().Skip()
	}

	authz := s.BuilderWithBearerScheme().Build()

	s.ApplyTestDelayer(authz)

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	audience := []string{"https://one-factor.example.com", "https://one-factor.example.com/"}

	mock.Ctx.Configuration.IdentityProviders = schema.IdentityProviders{
		OIDC: &schema.IdentityProvidersOpenIDConnect{
			HMACSecret: "abcdefghijklmnopqrstuvwxyz123456",
			Discovery: schema.IdentityProvidersOpenIDConnectDiscovery{
				BearerAuthorization: true,
			},
			Clients: []schema.IdentityProvidersOpenIDConnectClient{
				{
					ID:                  "test-ccs-client",
					Scopes:              []string{"openid"},
					Audience:            audience,
					GrantTypes:          []string{oidc.GrantTypeClientCredentials},
					AuthorizationPolicy: "one_factor",
				},
			},
		},
	}

	mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(&mock.Ctx.Configuration, mock.StorageMock, mock.Ctx.Providers.Templates)

	client, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, "test-ccs-client")
	s.Require().NoError(err)

	now := mock.Ctx.Providers.Clock.Now()

	session := &oidc.Session{
		ClientID:          "test-ccs-client",
		ClientCredentials: true,
		DefaultSession: &openid.DefaultSession{
			Headers: &fjwt.Headers{Extra: map[string]any{}},
			Claims: &fjwt.IDTokenClaims{
				Issuer:   "https://auth.example.com",
				Subject:  "test-ccs-client",
				IssuedAt: fjwt.NewNumericDate(now),
				Extra:    map[string]any{},
			},
			RequestedAt: now,
		},
	}

	requester := &oauthelia2.AccessRequest{
		GrantTypes: oauthelia2.Arguments{oidc.GrantTypeClientCredentials},
		Request: oauthelia2.Request{
			ID:                "request-ccs-noauthz",
			RequestedAt:       now,
			Client:            client,
			RequestedScope:    oauthelia2.Arguments{"openid"},
			GrantedScope:      oauthelia2.Arguments{"openid"},
			RequestedAudience: oauthelia2.Arguments(audience),
			GrantedAudience:   oauthelia2.Arguments(audience),
			Session:           session,
			Form:              url.Values{},
		},
	}

	token, signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.GenerateAccessToken(mock.Ctx, requester)
	s.Require().NoError(err)

	oauthSession, err := model.NewOAuth2SessionFromRequest(signature, requester)
	s.Require().NoError(err)

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Bearer "+token)

	mock.StorageMock.EXPECT().
		LoadOAuth2Session(gomock.Eq(mock.Ctx), gomock.Eq(storage.OAuth2SessionTypeAccessToken), gomock.Eq(signature)).
		Return(oauthSession, nil)

	authz.Handler(mock.Ctx)

	switch s.implementation {
	case AuthzImplAuthRequest, AuthzImplLegacy:
		s.Equal(fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
	default:
		s.Equal(fasthttp.StatusProxyAuthRequired, mock.Ctx.Response.StatusCode())
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

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

	mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
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

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Username: "john",
			Emails:   []string{"john@example.com"},
			Groups:   []string{"dev", "admins"},
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

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
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
				GetDetails(gomock.Eq("john")).
				Return(&authentication.UserDetails{
					Username: "john",
					Emails:   []string{"john@example.com"},
					Groups:   []string{"dev", "admins"},
				}, nil),
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
					Username: "john",
					Emails:   []string{"john@example.com"},
					Groups:   []string{"dev", "admins"},
				}, nil),
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
		)
	} else {
		gomock.InOrder(
			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(&authentication.UserDetails{
					Username: "john",
					Emails:   []string{"john@example.com"},
					Groups:   []string{"dev", "admins"},
				}, nil),
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
					Username: "john",
					Emails:   []string{"john@example.com"},
					Groups:   []string{"dev", "admins"},
				}, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil),
			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil),
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

	authz := s.BuildWithDelayer()

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(s.T())

			defer mock.Close()

			setUpMockClock(mock)

			targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

			s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

			mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, fmt.Sprintf("%s am9objpwYXNzd29yZA==", tc.scheme))

			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

			attempt := model.AuthenticationAttempt{
				Time:          mock.Ctx.Providers.Clock.Now(),
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

			mock.UserProviderMock.EXPECT().
				CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
				Return(true, nil)

			mock.UserProviderMock.EXPECT().
				GetDetails(gomock.Eq("john")).
				Return(&authentication.UserDetails{
					Username: "john",
					Emails:   []string{"john@example.com"},
					Groups:   []string{"dev", "admins"},
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

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://two-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

	mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
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

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Username: "john",
			Emails:   []string{"john@example.com"},
			Groups:   []string{"dev", "admins"},
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

	authz := s.BuildWithDelayer()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://deny.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")

	mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

	mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

	attempt := model.AuthenticationAttempt{
		Time:          mock.Ctx.Providers.Clock.Now(),
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

	mock.UserProviderMock.EXPECT().
		CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
		Return(true, nil)

	mock.UserProviderMock.EXPECT().
		GetDetails(gomock.Eq("john")).
		Return(&authentication.UserDetails{
			Username: "john",
			Emails:   []string{"john@example.com"},
			Groups:   []string{"dev", "admins"},
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

	header := NewHeaderAuthorizationAuthnStrategy(time.Duration(0), "basic")
	header.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	headerProxy := NewHeaderProxyAuthorizationAuthnStrategy(time.Duration(0), "basic")
	headerProxy.delay = header.delay

	builder = builder.WithStrategies(
		header,
		headerProxy,
		NewCookieSessionAuthnStrategy(builder.config.RefreshInterval),
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

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
			Time:          mock.Ctx.Providers.Clock.Now(),
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
			Username: "john",
			Emails:   []string{"john@example.com"},
			Groups:   []string{"dev", "admins"},
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

	header := NewHeaderAuthorizationAuthnStrategy(time.Duration(0), "basic")
	header.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	headerAuth := NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Duration(0), "basic")
	headerAuth.delay = header.delay

	builder = builder.WithStrategies(
		header,
		headerAuth,
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

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

	header := NewHeaderAuthorizationAuthnStrategy(time.Duration(0), "basic")
	header.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	headerProxy := NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Duration(0), "basic")
	headerProxy.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	builder = builder.WithStrategies(
		header,
		headerProxy,
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

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

	header := NewHeaderAuthorizationAuthnStrategy(time.Duration(0), "basic")
	header.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	headerProxy := NewHeaderProxyAuthorizationAuthRequestAuthnStrategy(time.Duration(0), "basic")
	headerProxy.delay = header.delay

	builder = builder.WithStrategies(
		header,
		headerProxy,
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	setUpMockClock(mock)

	targetURI := s.RequireParseRequestURI("https://one-factor.example.com")

	s.setRequest(mock.Ctx, fasthttp.MethodGet, targetURI, true, false)

	mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "Basic am9objpwYXNzd29yZA==")

	switch s.implementation {
	case AuthzImplLegacy:
		break
	default:
		mock.UserProviderMock.
			EXPECT().
			GetDetails(gomock.Eq("john")).Return(&authentication.UserDetails{Username: "john"}, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).Return(nil, nil)

		mock.StorageMock.
			EXPECT().
			LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq("john")).Return(nil, nil)

		attempt := model.AuthenticationAttempt{
			Time:          mock.Ctx.Providers.Clock.Now(),
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

	header := NewHeaderAuthorizationAuthnStrategy(time.Duration(0), "basic")
	header.delay = middlewares.NewTimingAttackDelay(1, time.Millisecond).SetMinimumDelay(10).SetRecord(false)

	builder = builder.WithStrategies(
		header,
	)

	authz := builder.Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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

	mock.Ctx.Providers.Clock = &mock.Clock
	mock.Ctx.Configuration.AuthenticationBackend.RefreshInterval = schema.NewRefreshIntervalDurationNever()
	mock.Ctx.Configuration.Session.Cookies[0].Inactivity = testInactivity
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

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
	mock.Ctx.Providers.Clock = &mock.Clock
	mock.Clock.Set(time.Now())
}
