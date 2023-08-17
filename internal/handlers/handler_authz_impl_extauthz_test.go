package handlers

import (
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestRunExtAuthzAuthzSuite(t *testing.T) {
	suite.Run(t, NewExtAuthzAuthzSuite())
}

func NewExtAuthzAuthzSuite() *ExtAuthzAuthzSuite {
	return &ExtAuthzAuthzSuite{
		AuthzSuite: &AuthzSuite{
			implementation: AuthzImplExtAuthz,
			setRequest:     setRequestExtAuthz,
		},
	}
}

type ExtAuthzAuthzSuite struct {
	*AuthzSuite
}

func (s *ExtAuthzAuthzSuite) TestShouldHandleAllMethodsDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, pairURI := range []urlpair{
				{s.RequireParseRequestURI("https://one-factor.example.com"), s.RequireParseRequestURI("https://auth.example.com/")},
				{s.RequireParseRequestURI("https://one-factor.example.com/subpath"), s.RequireParseRequestURI("https://auth.example.com/")},
				{s.RequireParseRequestURI("https://one-factor.example2.com"), s.RequireParseRequestURI("https://auth.example2.com/")},
				{s.RequireParseRequestURI("https://one-factor.example2.com/subpath"), s.RequireParseRequestURI("https://auth.example2.com/")},
			} {
				t.Run(pairURI.TargetURI.String(), func(t *testing.T) {
					expected := s.RequireParseRequestURI(pairURI.AutheliaURI.String())

					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					s.setRequest(mock.Ctx, method, pairURI.TargetURI, true, false)

					authz.Handler(mock.Ctx)

					switch method {
					case fasthttp.MethodGet, fasthttp.MethodOptions, fasthttp.MethodHead:
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

func (s *ExtAuthzAuthzSuite) TestShouldHandleAllMethodsOverrideAutheliaURLDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, pairURI := range []urlpair{
				{s.RequireParseRequestURI("https://one-factor.example.com"), s.RequireParseRequestURI("https://auth-from-override.example.com/")},
				{s.RequireParseRequestURI("https://one-factor.example.com/subpath"), s.RequireParseRequestURI("https://auth-from-override.example.com/")},
				{s.RequireParseRequestURI("https://one-factor.example2.com"), s.RequireParseRequestURI("https://auth-from-override.example2.com/")},
				{s.RequireParseRequestURI("https://one-factor.example2.com/subpath"), s.RequireParseRequestURI("https://auth-from-override.example2.com/")},
			} {
				t.Run(pairURI.TargetURI.String(), func(t *testing.T) {
					expected := s.RequireParseRequestURI(pairURI.AutheliaURI.String())

					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					mock.Ctx.Request.Header.Set("X-Authelia-Url", pairURI.AutheliaURI.String())
					s.setRequest(mock.Ctx, method, pairURI.TargetURI, true, false)

					authz.Handler(mock.Ctx)

					switch method {
					case fasthttp.MethodGet, fasthttp.MethodOptions, fasthttp.MethodHead:
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

func (s *ExtAuthzAuthzSuite) TestShouldHandleAllMethodsMissingAutheliaURLDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://bypass.example.com"),
				s.RequireParseRequestURI("https://bypass.example.com/subpath"),
				s.RequireParseRequestURI("https://bypass.example2.com"),
				s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.setRequest(mock.Ctx, method, targetURI, true, false)

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
					assert.Equal(t, fmt.Sprintf("%d %s", fasthttp.StatusBadRequest, fasthttp.StatusMessage(fasthttp.StatusBadRequest)), string(mock.Ctx.Response.Body()))
					assert.Equal(t, "", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
					assert.Equal(t, "text/plain; charset=utf-8", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderContentType)))
				})
			}
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldHandleAllMethodsXHRDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for xname, x := range testXHR {
				t.Run(xname, func(t *testing.T) {
					for _, pairURI := range []urlpair{
						{s.RequireParseRequestURI("https://one-factor.example.com"), s.RequireParseRequestURI("https://auth.example.com/")},
						{s.RequireParseRequestURI("https://one-factor.example.com/subpath"), s.RequireParseRequestURI("https://auth.example.com/")},
						{s.RequireParseRequestURI("https://one-factor.example2.com"), s.RequireParseRequestURI("https://auth.example2.com/")},
						{s.RequireParseRequestURI("https://one-factor.example2.com/subpath"), s.RequireParseRequestURI("https://auth.example2.com/")},
					} {
						t.Run(pairURI.TargetURI.String(), func(t *testing.T) {
							expected := s.RequireParseRequestURI(pairURI.AutheliaURI.String())

							authz := s.Builder().Build()

							mock := mocks.NewMockAutheliaCtx(t)

							defer mock.Close()

							s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

							s.setRequest(mock.Ctx, method, pairURI.TargetURI, x, x)

							mock.Ctx.SetUserValue(middlewares.UserValueRouterKeyExtAuthzPath, pairURI.TargetURI.Path)

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

func (s *ExtAuthzAuthzSuite) TestShouldHandleInvalidMethodCharsDeny() {
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
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					s.setRequest(mock.Ctx, method, targetURI, true, false)

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldHandleMissingHostDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			authz := s.Builder().Build()

			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

			s.setRequest(mock.Ctx, method, nil, true, false)

			authz.Handler(mock.Ctx)

			assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
			assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldHandleAllMethodsAllow() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://bypass.example.com"),
				s.RequireParseRequestURI("https://bypass.example.com/subpath"),
				s.RequireParseRequestURI("https://bypass.example2.com"),
				s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					s.setRequest(mock.Ctx, method, targetURI, true, false)

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldHandleAllMethodsAllowXHR() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for xname, x := range testXHR {
				t.Run(xname, func(t *testing.T) {
					for _, targetURI := range []*url.URL{
						s.RequireParseRequestURI("https://bypass.example.com"),
						s.RequireParseRequestURI("https://bypass.example.com/subpath"),
						s.RequireParseRequestURI("https://bypass.example2.com"),
						s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
					} {
						t.Run(targetURI.String(), func(t *testing.T) {
							authz := s.Builder().Build()

							mock := mocks.NewMockAutheliaCtx(t)

							defer mock.Close()

							s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

							s.setRequest(mock.Ctx, method, targetURI, x, x)

							authz.Handler(mock.Ctx)

							assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
							assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
						})
					}
				})
			}
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldHandleAllMethodsWithMethodsACL() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, methodACL := range testRequestMethods {
				targetURI := s.RequireParseRequestURI(fmt.Sprintf("https://bypass-%s.example.com", strings.ToLower(methodACL)))
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					s.setRequest(mock.Ctx, method, targetURI, true, false)

					authz.Handler(mock.Ctx)

					if method == methodACL {
						assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
						assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
					} else {
						expected := s.RequireParseRequestURI("https://auth.example.com/")

						query := expected.Query()
						query.Set(queryArgRD, targetURI.String())
						query.Set(queryArgRM, method)
						expected.RawQuery = query.Encode()

						switch method {
						case fasthttp.MethodHead:
							assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
							assert.Nil(t, mock.Ctx.Response.Body())
						case fasthttp.MethodGet, fasthttp.MethodOptions:
							assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
							assert.Equal(t, fmt.Sprintf(`<a href="%s">%d %s</a>`, utils.StringHTMLEscape(expected.String()), fasthttp.StatusFound, fasthttp.StatusMessage(fasthttp.StatusFound)), string(mock.Ctx.Response.Body()))
						default:
							assert.Equal(t, fasthttp.StatusSeeOther, mock.Ctx.Response.StatusCode())
							assert.Equal(t, fmt.Sprintf(`<a href="%s">%d %s</a>`, utils.StringHTMLEscape(expected.String()), fasthttp.StatusSeeOther, fasthttp.StatusMessage(fasthttp.StatusSeeOther)), string(mock.Ctx.Response.Body()))
						}

						assert.Equal(t, expected.String(), string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
					}
				})
			}
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldHandleInvalidURLForCVE202132637() {
	testCases := []struct {
		name         string
		scheme, host []byte
		path         string
		expected     int
	}{
		{"Should401UnauthorizedWithNullByte",
			[]byte("https"), []byte{0, 110, 111, 116, 45, 111, 110, 101, 45, 102, 97, 99, 116, 111, 114, 46, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109}, "/path-example",
			fasthttp.StatusBadRequest,
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
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					mock.Ctx.Configuration.AccessControl.DefaultPolicy = testBypass
					mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&mock.Ctx.Configuration)

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					mock.Ctx.Request.SetHostBytes(tc.host)
					mock.Ctx.Request.Header.SetMethodBytes([]byte(method))
					mock.Ctx.Request.Header.SetBytesKV([]byte(fasthttp.HeaderXForwardedProto), tc.scheme)
					mock.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")
					mock.Ctx.SetUserValue(middlewares.UserValueRouterKeyExtAuthzPath, tc.path)

					authz.Handler(mock.Ctx)

					assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldNotHandleAuthRequestAllMethodsAllow() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("OriginalMethod%s", method), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://bypass.example.com"),
				s.RequireParseRequestURI("https://bypass.example.com/subpath"),
				s.RequireParseRequestURI("https://bypass.example2.com"),
				s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					setRequestAuthRequest(mock.Ctx, method, targetURI, true, false)

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldNotHandleAuthRequestAllMethodsWithMethodsACL() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, methodACL := range testRequestMethods {
				targetURI := s.RequireParseRequestURI(fmt.Sprintf("https://bypass-%s.example.com", strings.ToLower(methodACL)))
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					setRequestAuthRequest(mock.Ctx, method, targetURI, true, false)

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldNotHandleForwardAuthAllMethodsAllow() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://bypass.example.com"),
				s.RequireParseRequestURI("https://bypass.example.com/subpath"),
				s.RequireParseRequestURI("https://bypass.example2.com"),
				s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					setRequestForwardAuth(mock.Ctx, method, targetURI, true, false)

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldNotHandleForwardAuthAllMethodsAllowXHR() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for xname, x := range testXHR {
				t.Run(xname, func(t *testing.T) {
					for _, targetURI := range []*url.URL{
						s.RequireParseRequestURI("https://bypass.example.com"),
						s.RequireParseRequestURI("https://bypass.example.com/subpath"),
						s.RequireParseRequestURI("https://bypass.example2.com"),
						s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
					} {
						t.Run(targetURI.String(), func(t *testing.T) {
							authz := s.Builder().Build()

							mock := mocks.NewMockAutheliaCtx(t)

							defer mock.Close()

							s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

							setRequestForwardAuth(mock.Ctx, method, targetURI, x, x)

							authz.Handler(mock.Ctx)

							assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
							assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
						})
					}
				})
			}
		})
	}
}

func (s *ExtAuthzAuthzSuite) TestShouldNotHandleForwardAuthAllMethodsWithMethodsACL() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, methodACL := range testRequestMethods {
				targetURI := s.RequireParseRequestURI(fmt.Sprintf("https://bypass-%s.example.com", strings.ToLower(methodACL)))
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					setRequestForwardAuth(mock.Ctx, method, targetURI, true, false)

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func setRequestExtAuthz(ctx *middlewares.AutheliaCtx, method string, targetURI *url.URL, accept, xhr bool) {
	ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)

	if method != "" {
		ctx.Request.Header.SetMethodBytes([]byte(method))
	}

	if targetURI != nil {
		ctx.Request.SetHost(targetURI.Host)
		ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
		ctx.SetUserValue(middlewares.UserValueRouterKeyExtAuthzPath, targetURI.Path)
	}

	setRequestXHRValues(ctx, accept, xhr)
}
