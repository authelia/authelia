// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
)

type authzURIParserImpl struct {
	name    string
	builder func() *AuthzBuilder
	reject  int
	set     func(ctx *middlewares.AutheliaCtx, host, path string)
}

func authzURIParserImpls() []authzURIParserImpl {
	return []authzURIParserImpl{
		{
			name:    "ForwardAuth",
			builder: func() *AuthzBuilder { return NewAuthzBuilder().WithImplementationForwardAuth() },
			reject:  fasthttp.StatusBadRequest,
			set: func(ctx *middlewares.AutheliaCtx, host, path string) {
				ctx.Request.Header.Set("X-Forwarded-Method", fasthttp.MethodGet)
				ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
				ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, host)
				ctx.Request.Header.Set("X-Forwarded-URI", path)
			},
		},
		{
			name:    "AuthRequest",
			builder: func() *AuthzBuilder { return NewAuthzBuilder().WithImplementationAuthRequest() },
			reject:  fasthttp.StatusBadRequest,
			set: func(ctx *middlewares.AutheliaCtx, host, path string) {
				ctx.Request.Header.Set(testXOriginalMethod, fasthttp.MethodGet)
				ctx.Request.Header.Set(testXOriginalUrl, "https://"+host+path)
			},
		},
		{
			name:    "ExtAuthz",
			builder: func() *AuthzBuilder { return NewAuthzBuilder().WithImplementationExtAuthz() },
			reject:  fasthttp.StatusBadRequest,
			set: func(ctx *middlewares.AutheliaCtx, host, path string) {
				ctx.Request.Header.SetMethod(fasthttp.MethodGet)
				ctx.Request.SetHost(host)
				ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
				ctx.SetUserValue(middlewares.UserValueRouterKeyExtAuthzPath, path)
			},
		},
		{
			name:    "Legacy",
			builder: func() *AuthzBuilder { return NewAuthzBuilder().WithImplementationLegacy() },
			reject:  fasthttp.StatusUnauthorized,
			set: func(ctx *middlewares.AutheliaCtx, host, path string) {
				ctx.Request.Header.Set("X-Forwarded-Method", fasthttp.MethodGet)
				ctx.Request.Header.Set(testXOriginalUrl, "https://"+host+path)
			},
		},
	}
}

func TestAuthzRejectsAmbiguousURI(t *testing.T) {
	testCases := []struct {
		path     string
		expected string
	}{
		{"/public/..;/admin", "parameter"},
		{"/foo/%2e%2e/bar", "encoded dot segment"},
		{"/foo/..%2fbar", "encoded dot segment"},
		{"/a%5Cb", "backslash"},
	}

	for _, impl := range authzURIParserImpls() {
		for _, tc := range testCases {
			t.Run(fmt.Sprintf("%s/%s", impl.name, tc.path), func(t *testing.T) {
				authz := impl.builder().Build()

				mock := mocks.NewMockAutheliaCtx(t)

				defer mock.Close()

				impl.set(mock.Ctx, "bypass.example.com", tc.path)

				authz.Handler(mock.Ctx)

				assert.Equal(t, impl.reject, mock.Ctx.Response.StatusCode())
				assert.Empty(t, mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
				assert.Empty(t, mock.Ctx.Response.Header.Peek("Remote-User"))

				entries := mock.Hook.AllEntries()

				require.NotEmpty(t, entries)

				var found bool

				for _, entry := range entries {
					if err, ok := entry.Data["error"].(error); ok && strings.Contains(err.Error(), tc.expected) {
						found = true
					}
				}

				assert.True(t, found, "expected a logged error containing %q", tc.expected)
			})
		}
	}
}

func TestAuthzRejectsUserInfoByDefault(t *testing.T) {
	for _, impl := range authzURIParserImpls() {
		t.Run(impl.name, func(t *testing.T) {
			authz := impl.builder().Build()

			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			impl.set(mock.Ctx, "u:pw@bypass.example.com", "/")

			authz.Handler(mock.Ctx)

			assert.Equal(t, impl.reject, mock.Ctx.Response.StatusCode())
			assert.Empty(t, mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
			assert.Empty(t, mock.Ctx.Response.Header.Peek("Remote-User"))

			entries := mock.Hook.AllEntries()

			require.NotEmpty(t, entries)

			var found bool

			for _, entry := range entries {
				assert.NotContains(t, entry.Message, "pw")

				for _, v := range entry.Data {
					assert.NotContains(t, fmt.Sprint(v), "pw")

					if err, ok := v.(error); ok && strings.Contains(err.Error(), "url contains userinfo") {
						found = true
					}
				}
			}

			assert.True(t, found, "expected a logged error containing %q", "url contains userinfo")
		})
	}
}

func TestAuthzAllowsUnambiguousEncoded(t *testing.T) {
	for _, impl := range authzURIParserImpls() {
		for _, path := range []string{"/group%2Fproject", "/a%25b"} {
			t.Run(fmt.Sprintf("%s/%s", impl.name, path), func(t *testing.T) {
				authz := impl.builder().Build()

				mock := mocks.NewMockAutheliaCtx(t)

				defer mock.Close()

				impl.set(mock.Ctx, "bypass.example.com", path)

				authz.Handler(mock.Ctx)

				assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
			})
		}
	}
}

func TestAuthzBuilderURIParser(t *testing.T) {
	t.Run("ShouldAllowUserInfo", func(t *testing.T) {
		authz := NewAuthzBuilder().WithConfig(uriParserConfig(schema.URIParser{Allow: []string{"user_info"}})).WithEndpointConfig(schema.ServerEndpointsAuthz{
			Implementation: AuthzImplForwardAuth.String(),
			URIParser:      "custom",
		}).Build()

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		authzURIParserImpls()[0].set(mock.Ctx, "u:pw@bypass.example.com", "/")

		authz.Handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	})

	t.Run("ShouldRejectEncodedPercent", func(t *testing.T) {
		authz := NewAuthzBuilder().WithConfig(uriParserConfig(schema.URIParser{Reject: []string{"path_encoded_percent"}})).WithEndpointConfig(schema.ServerEndpointsAuthz{
			Implementation: AuthzImplForwardAuth.String(),
			URIParser:      "custom",
		}).Build()

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		authzURIParserImpls()[0].set(mock.Ctx, "bypass.example.com", "/a%25b")

		authz.Handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
	})

	t.Run("ShouldUseDefaultDefinitionWhenUnset", func(t *testing.T) {
		authz := NewAuthzBuilder().WithConfig(&schema.Configuration{
			Definitions: schema.Definitions{
				URIParsers: map[string]schema.URIParser{
					schema.DefinitionURIParserDefault: {Allow: []string{"user_info"}},
				},
			},
		}).WithEndpointConfig(schema.ServerEndpointsAuthz{
			Implementation: AuthzImplForwardAuth.String(),
		}).Build()

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		authzURIParserImpls()[0].set(mock.Ctx, "u:pw@bypass.example.com", "/")

		authz.Handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	})

	t.Run("ShouldUseDefaultRulesWhenDefinitionMissing", func(t *testing.T) {
		authz := NewAuthzBuilder().WithConfig(&schema.Configuration{}).WithEndpointConfig(schema.ServerEndpointsAuthz{
			Implementation: AuthzImplForwardAuth.String(),
			URIParser:      "missing",
		}).Build()

		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		authzURIParserImpls()[0].set(mock.Ctx, "u:pw@bypass.example.com", "/")

		authz.Handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
	})

	t.Run("ShouldDefaultWithoutEndpointConfig", func(t *testing.T) {
		authz := NewAuthzBuilder().Build()

		require.NotNil(t, authz.config.URIParser)
	})
}

func requireNoPasswordLogged(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	entries := mock.Hook.AllEntries()

	require.NotEmpty(t, entries)

	for _, entry := range entries {
		assert.NotContains(t, entry.Message, "pw")

		for _, v := range entry.Data {
			assert.NotContains(t, fmt.Sprint(v), "pw")
		}
	}
}

func TestAuthzDoesNotLogAllowedUserInfoPassword(t *testing.T) {
	newAuthz := func() *Authz {
		return NewAuthzBuilder().WithConfig(uriParserConfig(schema.URIParser{Allow: []string{"user_info"}})).WithEndpointConfig(schema.ServerEndpointsAuthz{
			Implementation: AuthzImplForwardAuth.String(),
			URIParser:      "custom",
		}).Build()
	}

	set := authzURIParserImpls()[0].set

	t.Run("ShouldRedactOnRedirect", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		set(mock.Ctx, "u:pw@one-factor.example.com", "/")

		mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html")

		newAuthz().Handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
		assert.NotEmpty(t, mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))

		requireNoPasswordLogged(t, mock)
	})

	t.Run("ShouldRedactOnPortalDomainMismatch", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		set(mock.Ctx, "bypass.example.com", "/")

		mock.Ctx.Request.URI().SetQueryString("authelia_url=https%3A%2F%2Fu%3Apw%40auth.other.net%2F")

		newAuthz().Handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())

		requireNoPasswordLogged(t, mock)
	})
}

func TestAuthzPortalURLUsesEndpointURIParser(t *testing.T) {
	const portal = "authelia_url=https%3A%2F%2Fauth.example.com%2F%252e%252e%2F"

	set := authzURIParserImpls()[0].set

	t.Run("ShouldRejectByDefault", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		set(mock.Ctx, "one-factor.example.com", "/")

		mock.Ctx.Request.URI().SetQueryString(portal)

		NewAuthzBuilder().WithImplementationForwardAuth().Build().Handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())
		assert.Empty(t, mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
	})

	t.Run("ShouldAllowWhenConfigured", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		set(mock.Ctx, "one-factor.example.com", "/")

		mock.Ctx.Request.URI().SetQueryString(portal)
		mock.Ctx.Request.Header.Set(fasthttp.HeaderAccept, "text/html")

		NewAuthzBuilder().WithConfig(uriParserConfig(schema.URIParser{Allow: []string{"path_encoded_dot_segment"}})).WithEndpointConfig(schema.ServerEndpointsAuthz{
			Implementation: AuthzImplForwardAuth.String(),
			URIParser:      "custom",
		}).Build().Handler(mock.Ctx)

		assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
		assert.NotEmpty(t, mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation))
	})
}

func uriParserConfig(parser schema.URIParser) *schema.Configuration {
	return &schema.Configuration{
		Definitions: schema.Definitions{
			URIParsers: map[string]schema.URIParser{"custom": parser},
		},
	}
}
