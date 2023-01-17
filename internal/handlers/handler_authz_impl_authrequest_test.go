package handlers

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
)

func TestRunAuthRequestAuthzSuite(t *testing.T) {
	suite.Run(t, NewAuthRequestAuthzSuite())
}

func NewAuthRequestAuthzSuite() *AuthRequestAuthzSuite {
	return &AuthRequestAuthzSuite{
		AuthzSuite: &AuthzSuite{
			builder: NewAuthzBuilder().WithImplementationAuthRequest(),
		},
	}
}

type AuthRequestAuthzSuite struct {
	*AuthzSuite
}

func (s *AuthRequestAuthzSuite) TestShouldHandleAllMethodsDeny() {
	for _, methodOriginal := range methods {
		s.T().Run(fmt.Sprintf("OriginalMethod%s", methodOriginal), func(t *testing.T) {
			for _, targetURI := range []*url.URL{
				s.RequireParseRequestURI("https://one-factor.example.com"),
				s.RequireParseRequestURI("https://one-factor.example.com/subpath"),
				s.RequireParseRequestURI("https://one-factor.example2.com"),
				s.RequireParseRequestURI("https://one-factor.example2.com/subpath"),
			} {
				t.Run(targetURI.String(), func(t *testing.T) {
					authz := s.builder.Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					mock.Ctx.Request.Header.Set("X-Original-Method", methodOriginal)
					mock.Ctx.Request.Header.Set("X-Original-Url", targetURI.String())

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}

		})
	}
}

func (s *AuthRequestAuthzSuite) TestShouldHandleInvalidMethodCharsDeny() {
	for _, methodOriginal := range methods {
		methodOriginal += "z"

		s.T().Run(fmt.Sprintf("OriginalMethod%s", methodOriginal), func(t *testing.T) {
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

					mock.Ctx.Request.Header.Set("X-Original-Method", methodOriginal)
					mock.Ctx.Request.Header.Set("X-Original-Url", targetURI.String())

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}

		})
	}
}

func (s *AuthRequestAuthzSuite) TestShouldHandleMissingXOriginalMethodDeny() {
	for _, targetURI := range []*url.URL{
		s.RequireParseRequestURI("https://bypass.example.com"),
		s.RequireParseRequestURI("https://bypass.example.com/subpath"),
		s.RequireParseRequestURI("https://bypass.example2.com"),
		s.RequireParseRequestURI("https://bypass.example2.com/subpath"),
	} {
		s.T().Run(targetURI.String(), func(t *testing.T) {
			authz := s.builder.Build()

			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Request.Header.Set("X-Original-Url", targetURI.String())

			authz.Handler(mock.Ctx)

			assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
			assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
		})
	}
}

func (s *AuthRequestAuthzSuite) TestShouldHandleMissingXOriginalURLDeny() {
	for _, methodOriginal := range methods {
		s.T().Run(fmt.Sprintf("OriginalMethod%s", methodOriginal), func(t *testing.T) {

			authz := s.builder.Build()

			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Request.Header.Set("X-Original-Method", methodOriginal)

			authz.Handler(mock.Ctx)

			assert.Equal(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
			assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
		})
	}
}

func (s *AuthRequestAuthzSuite) TestShouldHandleAllMethodsAllow() {
	for _, methodOriginal := range methods {
		s.T().Run(fmt.Sprintf("OriginalMethod%s", methodOriginal), func(t *testing.T) {
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

					mock.Ctx.Request.Header.Set("X-Original-Method", methodOriginal)
					mock.Ctx.Request.Header.Set("X-Original-Url", targetURI.String())

					authz.Handler(mock.Ctx)

					assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}

		})
	}
}

func (s *AuthRequestAuthzSuite) TestShouldHandleInvalidURL() {
	testCases := []struct {
		name     string
		uri      []byte
		expected int
	}{
		{"Should401UnauthorizedWithNullByte",
			[]byte{104, 116, 116, 112, 115, 58, 47, 47, 0, 110, 111, 116, 45, 111, 110, 101, 45, 102, 97, 99, 116, 111, 114, 46, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109},
			fasthttp.StatusUnauthorized,
		},
		{"Should200OkWithoutNullByte",
			[]byte{104, 116, 116, 112, 115, 58, 47, 47, 110, 111, 116, 45, 111, 110, 101, 45, 102, 97, 99, 116, 111, 114, 46, 101, 120, 97, 109, 112, 108, 101, 46, 99, 111, 109},
			fasthttp.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			for _, methodOriginal := range methods {
				t.Run(fmt.Sprintf("OriginalMethod%s", methodOriginal), func(t *testing.T) {
					authz := s.builder.Build()

					mock := mocks.NewMockAutheliaCtx(t)

					defer mock.Close()

					mock.Ctx.Configuration.AccessControl.DefaultPolicy = "bypass"
					mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&mock.Ctx.Configuration)

					mock.Ctx.Request.Header.Set("X-Original-Method", methodOriginal)
					mock.Ctx.Request.Header.SetBytesKV([]byte("X-Original-Url"), tc.uri)

					authz.Handler(mock.Ctx)

					assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())
					assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				})
			}
		})
	}
}

type namebytes struct {
	name string
	data []byte
}

var methods = []string{fasthttp.MethodOptions, fasthttp.MethodHead, fasthttp.MethodGet, fasthttp.MethodDelete, fasthttp.MethodPatch, fasthttp.MethodPost, fasthttp.MethodPut, fasthttp.MethodConnect, fasthttp.MethodTrace}
