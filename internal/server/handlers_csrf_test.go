// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"net/url"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestHandlerRouterShouldRequireCSRFTokenForEveryStateChangingAPIRoute(t *testing.T) {
	provider, err := templates.New(templates.Config{})
	require.NoError(t, err)

	require.NoError(t, provider.LoadTemplatedAssets(assets))

	config := &schema.Configuration{
		Server: schema.Server{
			Address:   schema.DefaultServerConfiguration.Address,
			Endpoints: schema.DefaultServerConfiguration.Endpoints,
		},
		Session: schema.Session{
			Cookies: []schema.SessionCookie{
				{
					SessionCookieCommon: schema.SessionCookieCommon{
						Name:       "authelia_session",
						RememberMe: schema.DefaultSessionConfiguration.RememberMe,
						Expiration: schema.DefaultSessionConfiguration.Expiration,
					},
					Domain:                "example.com",
					DefaultRedirectionURL: &url.URL{Scheme: "https", Host: "www.example.com"},
					AutheliaURL:           &url.URL{Scheme: "https", Host: "login.example.com"},
				},
			},
		},
		AuthenticationBackend: schema.AuthenticationBackend{
			ExternalIdentity: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "example", Name: "Example", Issuer: "https://op.example.com"},
				},
			},
		},
		WebAuthn: schema.WebAuthn{
			EnablePasskeyLogin: true,
		},
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{},
		},
	}

	providers := middlewares.NewProvidersBasic()
	providers.Random = random.NewMathematical()
	providers.Templates = provider
	providers.Clock = clock.New()
	providers.OpenIDConnect, _ = oidc.NewOpenIDConnectProvider(config, nil, provider)

	providers.Session, err = mocks.NewSessionProvider(config, providers.Clock, providers.Random)
	require.NoError(t, err)

	r, err := handlerRouter(config, providers)
	require.NoError(t, err)

	handler, err := handlerMain(config, providers)
	require.NoError(t, err)

	exempt := map[string]bool{
		"POST /api/checks/safe-redirection":           true,
		"POST /api/identity/callback":                 true,
		"POST /api/identity/{provider}/callback":      true,
		"POST /api/oidc/authorization":                true,
		"POST /api/oidc/device-authorization":         true,
		"POST /api/oidc/introspection":                true,
		"POST /api/oidc/pushed-authorization-request": true,
		"POST /api/oidc/revocation":                   true,
		"POST /api/oidc/token":                        true,
		"POST /api/oidc/userinfo":                     true,
	}

	hooks := logrus.StandardLogger().ReplaceHooks(make(logrus.LevelHooks))

	t.Cleanup(func() {
		logrus.StandardLogger().ReplaceHooks(hooks)
	})

	hook := test.NewGlobal()

	var routes []string

	for method, paths := range r.List() {
		switch method {
		case fasthttp.MethodGet, fasthttp.MethodHead, fasthttp.MethodOptions, "*":
			continue
		}

		for _, path := range paths {
			if !strings.HasPrefix(path, "/api/") {
				continue
			}

			route := method + " " + path

			if exempt[route] {
				delete(exempt, route)

				continue
			}

			routes = append(routes, route)
		}
	}

	assert.Empty(t, exempt)

	sort.Strings(routes)

	require.NotEmpty(t, routes)

	for _, route := range routes {
		t.Run(route, func(t *testing.T) {
			method, path, _ := strings.Cut(route, " ")

			cookies := newTestCSRFRouteSessionCookies(t, config, providers)

			ctx := newAssetRequestCtx(method, regexpTestCSRFRouteParam.ReplaceAllString(path, "example"), "")
			ctx.Request.Header.SetHost("login.example.com")
			ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
			ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "login.example.com")
			ctx.Request.Header.Set("Sec-Fetch-Site", "same-origin")

			for name, value := range cookies {
				ctx.Request.Header.SetCookie(name, value)
			}

			hook.Reset()

			handler(ctx)

			assert.Equal(t, fasthttp.StatusForbidden, ctx.Response.StatusCode())

			rejected := false

			for _, entry := range hook.AllEntries() {
				if entry.Level == logrus.WarnLevel && entry.Message == "Request rejected as it did not include a valid CSRF token" {
					rejected = true
				}
			}

			assert.True(t, rejected, "the route was not rejected by the CSRF token check")
		})
	}
}

var regexpTestCSRFRouteParam = regexp.MustCompile(`\{[^}]+}`)

func newTestCSRFRouteSessionCookies(t *testing.T, config *schema.Configuration, providers middlewares.Providers) (cookies map[string]string) {
	t.Helper()

	ctx := middlewares.NewAutheliaCtx(newAssetRequestCtx(fasthttp.MethodGet, "/api/state", ""), *config, providers)
	ctx.Request.Header.SetHost("login.example.com")
	ctx.Request.Header.Set(fasthttp.HeaderXForwardedProto, "https")
	ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "login.example.com")

	userSession, err := ctx.GetSession()
	require.NoError(t, err)

	userSession.Username = "john"
	userSession.SetOneFactorPassword(providers.Clock.Now(), false)
	userSession.SetTwoFactorTOTP(providers.Clock.Now())

	require.NoError(t, ctx.SaveSession(&userSession))

	cookies = map[string]string{}

	for key, value := range ctx.Response.Header.Cookies() {
		c := &fasthttp.Cookie{}

		require.NoError(t, c.ParseBytes(value))

		if string(key) != session.CSRFCookieName("authelia_session") {
			cookies[string(key)] = string(c.Value())
		}
	}

	require.NotEmpty(t, cookies["__Host-Http-authelia_session"])

	return cookies
}
