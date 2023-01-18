package handlers

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestRunLegacyAuthzSuite(t *testing.T) {
	suite.Run(t, NewLegacyAuthzSuite())
}

func NewLegacyAuthzSuite() *LegacyAuthzSuite {
	return &LegacyAuthzSuite{
		AuthzSuite: &AuthzSuite{
			builder: NewAuthzBuilder().WithImplementationLegacy(),
		},
	}
}

type LegacyAuthzSuite struct {
	*AuthzSuite
}

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, pairURI := range []urlpair{
				{s.RequireParseRequestURI("https://one-factor.example.com/"), s.RequireParseRequestURI("https://auth.example.com/")},
				{s.RequireParseRequestURI("https://one-factor.example.com/subpath"), s.RequireParseRequestURI("https://auth.example.com/")},
				{s.RequireParseRequestURI("https://one-factor.example2.com/"), s.RequireParseRequestURI("https://auth.example2.com/")},
				{s.RequireParseRequestURI("https://one-factor.example2.com/subpath"), s.RequireParseRequestURI("https://auth.example2.com/")},
			} {
				t.Run(pairURI.TargetURI.String(), func(t *testing.T) {
					expected := s.RequireParseRequestURI(pairURI.AutheliaURI.String())

					authz := s.builder.Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
						mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
					}

					mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, pairURI.TargetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, pairURI.TargetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-Uri", pairURI.TargetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					switch method {
					case fasthttp.MethodGet, fasthttp.MethodOptions:
						assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
					default:
						assert.Equal(t, fasthttp.StatusSeeOther, mock.Ctx.Response.StatusCode())
					}

					query := expected.Query()
					query.Set(queryArgRD, pairURI.TargetURI.String())
					query.Set(queryArgRM, method)
					expected.RawQuery = query.Encode()

					assert.Equal(t, expected.String(), string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsOverrideAutheliaURLDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, pairURI := range []urlpair{
				{s.RequireParseRequestURI("https://one-factor.example.com/"), s.RequireParseRequestURI("https://auth-from-override.example.com/")},
				{s.RequireParseRequestURI("https://one-factor.example.com/subpath"), s.RequireParseRequestURI("https://auth-from-override.example.com/")},
				{s.RequireParseRequestURI("https://one-factor.example2.com/"), s.RequireParseRequestURI("https://auth-from-override.example2.com/")},
				{s.RequireParseRequestURI("https://one-factor.example2.com/subpath"), s.RequireParseRequestURI("https://auth-from-override.example2.com/")},
			} {
				t.Run(pairURI.TargetURI.String(), func(t *testing.T) {
					expected := s.RequireParseRequestURI(pairURI.AutheliaURI.String())

					authz := s.builder.Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
						mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
					}

					mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

					mock.Ctx.RequestCtx.QueryArgs().Set(queryArgRD, pairURI.AutheliaURI.String())
					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, pairURI.TargetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, pairURI.TargetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-Uri", pairURI.TargetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					switch method {
					case fasthttp.MethodGet, fasthttp.MethodOptions:
						assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
					default:
						assert.Equal(t, fasthttp.StatusSeeOther, mock.Ctx.Response.StatusCode())
					}

					query := expected.Query()
					query.Set(queryArgRD, pairURI.TargetURI.String())
					query.Set(queryArgRM, method)
					expected.RawQuery = query.Encode()

					assert.Equal(t, expected.String(), string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsMissingAutheliaURLDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://bypass.example.com"),
				s.RequireParseRequestURI("https://bypass.example.com/subpath"),
				s.RequireParseRequestURI("https://bypass.example2.com"),
				s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.builder.Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, targetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-Uri", targetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
					assert.Equal(t, "", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsXHRDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, x := range []bool{true, false} {
				xname := testWithoutAccept

				if x {
					xname = testWithXHRHeader
				}

				t.Run(xname, func(t *testing.T) {
					for _, pairURI := range []urlpair{
						{s.RequireParseRequestURI("https://one-factor.example.com/"), s.RequireParseRequestURI("https://auth.example.com/")},
						{s.RequireParseRequestURI("https://one-factor.example.com/subpath"), s.RequireParseRequestURI("https://auth.example.com/")},
						{s.RequireParseRequestURI("https://one-factor.example2.com/"), s.RequireParseRequestURI("https://auth.example2.com/")},
						{s.RequireParseRequestURI("https://one-factor.example2.com/subpath"), s.RequireParseRequestURI("https://auth.example2.com/")},
					} {
						t.Run(pairURI.TargetURI.String(), func(t *testing.T) {
							expected := s.RequireParseRequestURI(pairURI.AutheliaURI.String())

							authz := s.builder.Build()

							mock := mocks.NewMockAutheliaCtx(t)

							defer mock.Close()

							for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
								mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
							}

							mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

							mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
							mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, pairURI.TargetURI.Scheme)
							mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, pairURI.TargetURI.Host)
							mock.Ctx.Request.Header.Set("X-Forwarded-Uri", pairURI.TargetURI.Path)

							if x {
								mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")
								mock.Ctx.Request.Header.Set(fasthttp.HeaderXRequestedWith, "XMLHttpRequest")
							}

							authz.Handler(mock.Ctx)

							assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())

							query := expected.Query()
							query.Set(queryArgRD, pairURI.TargetURI.String())
							query.Set(queryArgRM, method)
							expected.RawQuery = query.Encode()

							assert.Equal(t, expected.String(), string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
						})
					}
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleInvalidMethodCharsDeny() {
	for _, method := range testRequestMethods {
		method += "z"

		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://bypass.example.com"),
				s.RequireParseRequestURI("https://bypass.example.com/subpath"),
				s.RequireParseRequestURI("https://bypass.example2.com"),
				s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.builder.Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
						mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
					}

					mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, targetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-Uri", targetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleMissingHostDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			authz := s.builder.Build()

			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
				mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
			}

			mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

			mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			mock.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)
			mock.Ctx.Request.Header.Set("X-Forwarded-Uri", "/")
			mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

			authz.Handler(mock.Ctx)

			assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
			assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsAllow() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://bypass.example.com"),
				s.RequireParseRequestURI("https://bypass.example.com/subpath"),
				s.RequireParseRequestURI("https://bypass.example2.com"),
				s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.builder.Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
						mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
					}

					mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, targetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-Uri", targetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsAllowXHR() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://bypass.example.com"),
				s.RequireParseRequestURI("https://bypass.example.com/subpath"),
				s.RequireParseRequestURI("https://bypass.example2.com"),
				s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.builder.Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
						mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
					}

					mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, targetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-Uri", targetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleInvalidURL() {
	testCases := []struct {
		name         string
		scheme, host []byte
		path         string
		expected     int
	}{
		{"Should401UnauthorizedWithNullByte",
			[]byte("https"), []byte{0, 110, 111, 116, 45, 111, 110, 101, 45, 102, 97, 99, 116, 111, 114, 46, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109}, "/path-example",
			fasthttp.StatusUnauthorized,
		},
		{"Should200OkWithoutNullByte",
			[]byte("https"), []byte{110, 111, 116, 45, 111, 110, 101, 45, 102, 97, 99, 116, 111, 114, 46, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109}, "/path-example",
			fasthttp.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			for _, method := range testRequestMethods {
				t.Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
					authz := s.builder.Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					mock.Ctx.Configuration.AccessControl.DefaultPolicy = testBypass
					mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&mock.Ctx.Configuration)

					for i, cookie := range mock.Ctx.Configuration.Session.Cookies {
						mock.Ctx.Configuration.Session.Cookies[i].AutheliaURL = s.RequireParseRequestURI(fmt.Sprintf("https://auth.%s", cookie.Domain))
					}

					mock.Ctx.Providers.SessionProvider = session.NewProvider(mock.Ctx.Configuration.Session, nil)

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.SetBytesKV([]byte(fasthttp.HeaderXForwardedProto), tc.scheme)
					mock.Ctx.Request.Header.SetBytesKV([]byte(fasthttp.HeaderXForwardedHost), tc.host)
					mock.Ctx.Request.Header.Set("X-Forwarded-Uri", tc.path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}
