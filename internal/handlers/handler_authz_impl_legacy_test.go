package handlers

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestRunLegacyAuthzSuite(t *testing.T) {
	suite.Run(t, NewLegacyAuthzSuite())
}

func NewLegacyAuthzSuite() *LegacyAuthzSuite {
	return &LegacyAuthzSuite{
		AuthzSuite: &AuthzSuite{
			implementation: AuthzImplLegacy,
			setRequest:     setRequestLegacy,
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

					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					mock.Ctx.RequestCtx.QueryArgs().Set(queryArgRD, pairURI.AutheliaURI.String())
					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, pairURI.TargetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, pairURI.TargetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-URI", pairURI.TargetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

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

					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					mock.Ctx.RequestCtx.QueryArgs().Set(queryArgRD, pairURI.AutheliaURI.String())
					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, pairURI.TargetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, pairURI.TargetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-URI", pairURI.TargetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

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

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsMissingAutheliaURLBypassStatus200() {
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

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, targetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-URI", targetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
					assert.Equal(t, "", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsMissingAutheliaURLOneFactorStatus401() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://one-factor.example.com"),
				s.RequireParseRequestURI("https://one-factor.example.com/subpath"),
				s.RequireParseRequestURI("https://one-factor.example2.com"),
				s.RequireParseRequestURI("https://one-factor.example2.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, targetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-URI", targetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
					assert.Equal(t, "", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsRDAutheliaURLOneFactorStatus302Or303() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://one-factor.example.com/"),
				s.RequireParseRequestURI("https://one-factor.example.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, targetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-URI", targetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")
					mock.Ctx.Request.SetRequestURI("/api/verify?rd=https%3A%2F%2Fauth.example.com")

					authz.Handler(mock.Ctx)

					switch method {
					case fasthttp.MethodGet, fasthttp.MethodOptions, fasthttp.MethodHead:
						assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
					default:
						assert.Equal(t, fasthttp.StatusSeeOther, mock.Ctx.Response.StatusCode())
					}

					query := &url.Values{}
					query.Set("rd", targetURI.String())
					query.Set("rm", method)

					assert.Equal(t, fmt.Sprintf("https://auth.example.com/?%s", query.Encode()), string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsXHRDeny() {
	for _, method := range testRequestMethods {
		s.T().Run(fmt.Sprintf("Method%s", method), func(t *testing.T) {
			for xname, x := range testXHR {
				t.Run(xname, func(t *testing.T) {
					for _, pairURI := range []urlpair{
						{s.RequireParseRequestURI("https://one-factor.example.com/"), s.RequireParseRequestURI("https://auth.example.com/")},
						{s.RequireParseRequestURI("https://one-factor.example.com/subpath"), s.RequireParseRequestURI("https://auth.example.com/")},
						{s.RequireParseRequestURI("https://one-factor.example2.com/"), s.RequireParseRequestURI("https://auth.example2.com/")},
						{s.RequireParseRequestURI("https://one-factor.example2.com/subpath"), s.RequireParseRequestURI("https://auth.example2.com/")},
					} {
						t.Run(pairURI.TargetURI.String(), func(t *testing.T) {
							expected := s.RequireParseRequestURI(pairURI.AutheliaURI.String())

							authz := s.Builder().Build()

							mock := mocks.NewMockAutheliaCtx(t)

							defer mock.Close()

							s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

							mock.Ctx.RequestCtx.QueryArgs().Set(queryArgRD, pairURI.AutheliaURI.String())
							mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
							mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, pairURI.TargetURI.Scheme)
							mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, pairURI.TargetURI.Host)
							mock.Ctx.Request.Header.Set("X-Forwarded-URI", pairURI.TargetURI.Path)

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
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, targetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-URI", targetURI.Path)
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
			authz := s.Builder().Build()

			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

			mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			mock.Ctx.Request.Header.Del(fasthttp.HeaderXForwardedHost)
			mock.Ctx.Request.Header.Set("X-Forwarded-URI", "/")
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
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, targetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-URI", targetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleAllMethodsWithMethodsACL() {
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
					mock.Ctx.RequestCtx.QueryArgs().Set(queryArgRD, "https://auth.example.com")

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
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, targetURI.Scheme)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, targetURI.Host)
					mock.Ctx.Request.Header.Set("X-Forwarded-URI", targetURI.Path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleLegacyBasicAuth() { // TestShouldVerifyAuthBasicArgOk.
	authz := s.Builder().Build()

	mock := mocks.NewMockAutheliaCtx(s.T())

	defer mock.Close()

	s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

	mock.Ctx.QueryArgs().Add("auth", "basic")
	mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "Basic am9objpwYXNzd29yZA==")
	mock.Ctx.Request.Header.Set("X-Original-URL", "https://one-factor.example.com")

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
	)

	authz.Handler(mock.Ctx)

	s.Equal(fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
}

func (s *LegacyAuthzSuite) TestShouldHandleLegacyBasicAuthFailures() {
	testCases := []struct {
		name  string
		setup func(mock *mocks.MockAutheliaCtx)
	}{
		{
			"HeaderAbsent", // TestShouldVerifyAuthBasicArgFailingNoHeader.
			nil,
		},
		{
			"HeaderEmpty", // TestShouldVerifyAuthBasicArgFailingEmptyHeader.
			func(mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "")
			},
		},
		{
			"HeaderIncorrect", // TestShouldVerifyAuthBasicArgFailingWrongHeader.
			func(mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set(fasthttp.HeaderProxyAuthorization, "Basic am9objpwYXNzd29yZA==")
			},
		},
		{
			"IncorrectPassword", // TestShouldVerifyAuthBasicArgFailingWrongPassword.
			func(mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "Basic am9objpwYXNzd29yZA==")

				mock.UserProviderMock.EXPECT().
					CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
					Return(false, fmt.Errorf("generic error"))
			},
		},
		{
			"NoAccess", // TestShouldVerifyAuthBasicArgFailingWrongPassword.
			func(mock *mocks.MockAutheliaCtx) {
				mock.Ctx.Request.Header.Set(fasthttp.HeaderAuthorization, "Basic am9objpwYXNzd29yZA==")
				mock.Ctx.Request.Header.Set("X-Original-URL", "https://admin.example.com/")

				gomock.InOrder(
					mock.UserProviderMock.EXPECT().
						CheckUserPassword(gomock.Eq("john"), gomock.Eq("password")).
						Return(true, nil),

					mock.UserProviderMock.EXPECT().
						GetDetails(gomock.Eq("john")).
						Return(&authentication.UserDetails{
							Emails: []string{"john@example.com"},
							Groups: []string{"dev", "admin"},
						}, nil),
				)
			},
		},
	}

	authz := s.Builder().Build()

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

			mock.Ctx.QueryArgs().Add("auth", "basic")
			mock.Ctx.Request.Header.Set("X-Original-URL", "https://one-factor.example.com")

			if tc.setup != nil {
				tc.setup(mock)
			}

			authz.Handler(mock.Ctx)

			assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
			assert.Equal(t, "401 Unauthorized", string(mock.Ctx.Response.Body()))
			assert.Regexp(t, regexp.MustCompile("^Basic realm="), string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate)))
		})
	}
}

func (s *LegacyAuthzSuite) TestShouldHandleInvalidURLForCVE202132637() {
	testCases := []struct {
		name         string
		scheme, host []byte
		path         string
		expected     int
	}{
		// The first byte in the host sequence is the null byte. This should never respond with 200 OK.
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
					authz := s.Builder().Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					mock.Ctx.Configuration.AccessControl.DefaultPolicy = testBypass
					mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&mock.Ctx.Configuration)

					s.ConfigureMockSessionProviderWithAutomaticAutheliaURLs(mock)

					mock.Ctx.Request.Header.Set("X-Forwarded-Method", method)
					mock.Ctx.Request.Header.SetBytesKV([]byte(fasthttp.HeaderXForwardedProto), tc.scheme)
					mock.Ctx.Request.Header.SetBytesKV([]byte(fasthttp.HeaderXForwardedHost), tc.host)
					mock.Ctx.Request.Header.Set("X-Forwarded-URI", tc.path)
					mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html; charset=utf-8")

					authz.Handler(mock.Ctx)

					assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

func setRequestLegacy(ctx *middlewares.AutheliaCtx, method string, targetURI *url.URL, accept, xhr bool) {
	if method != "" {
		ctx.Request.Header.Set("X-Forwarded-Method", method)
	}

	if targetURI != nil {
		ctx.Request.Header.Set(testXOriginalUrl, targetURI.String())
	}

	setRequestXHRValues(ctx, accept, xhr)
}
