// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"net"
	"testing"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestHandleError(t *testing.T) {
	handler := handleError("server", schema.ServerHeadersTrustedProxies{})

	testCases := []struct {
		name               string
		err                error
		xff                string
		isTLS              bool
		expectedStatusCode int
	}{
		{
			"ShouldHandleSmallBufferError",
			&fasthttp.ErrSmallBuffer{},
			"",
			false,
			fasthttp.StatusRequestHeaderFieldsTooLarge,
		},
		{
			"ShouldHandleNetOpErrorTimeout",
			&net.OpError{Op: "read", Err: &timeoutError{}},
			"",
			false,
			fasthttp.StatusRequestTimeout,
		},
		{
			"ShouldHandleNetOpErrorNonTimeout",
			&net.OpError{Op: "read", Err: fmt.Errorf("connection reset")},
			"",
			false,
			fasthttp.StatusBadRequest,
		},
		{
			"ShouldHandleGenericError",
			fmt.Errorf("some unknown error"),
			"",
			false,
			fasthttp.StatusBadRequest,
		},
		{
			"ShouldHandleTLSHandshakeOnPlainText",
			fmt.Errorf(`error when reading request headers: contents: \x16\x03\x03`),
			"",
			false,
			fasthttp.StatusBadRequest,
		},
		{
			"ShouldHandleGenericErrorWithXFF",
			fmt.Errorf("some error"),
			"10.0.0.1, 192.168.1.1",
			false,
			fasthttp.StatusBadRequest,
		},
		{
			"ShouldHandleGenericErrorOnTLS",
			fmt.Errorf("some error"),
			"",
			true,
			fasthttp.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				ctx fasthttp.RequestCtx
				req fasthttp.Request
			)

			if tc.xff != "" {
				req.Header.Set(fasthttp.HeaderXForwardedFor, tc.xff)
			}

			ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

			if tc.isTLS {
				ctx.Request.SetRequestURI("https://example.com/")
			}

			handler(&ctx, tc.err)

			assert.Equal(t, tc.expectedStatusCode, ctx.Response.StatusCode())
		})
	}
}

func TestHandleErrorShouldRedactAppAuthenticatorDevicePaths(t *testing.T) {
	logger := logging.Logger()

	hooks := logrus.LevelHooks{}

	for level, h := range logger.Hooks {
		hooks[level] = append(hooks[level], h...)
	}

	hook := logtest.NewLocal(logger)

	t.Cleanup(func() { logger.ReplaceHooks(hooks) })

	handler := handleError("server", schema.ServerHeadersTrustedProxies{})

	testCases := []struct {
		name     string
		path     string
		expected string
	}{
		{"ShouldRedactChallenge", "/api/secondfactor/app/challenge/secretpullid", "/api/secondfactor/app/challenge/{redacted}"},
		{"ShouldRedactUnderBasePath", "/auth/api/register/app/secretcode", "/auth/api/register/app/{redacted}"},
		{"ShouldNotRedactOtherPaths", "/api/health", "/api/health"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hook.Reset()

			var (
				ctx fasthttp.RequestCtx
				req fasthttp.Request
			)

			req.SetRequestURI(tc.path)

			ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

			handler(&ctx, fmt.Errorf("some error"))

			entry := hook.LastEntry()
			require.NotNil(t, entry)
			assert.Equal(t, tc.expected, entry.Data[logging.FieldPath])
		})
	}
}

func TestHandleNotFound(t *testing.T) {
	next := func(ctx *fasthttp.RequestCtx) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBodyString("next called")
	}

	handler := handleNotFound(next)

	testCases := []struct {
		name               string
		path               string
		expectedStatusCode int
		expectNext         bool
	}{
		{"ShouldReturn404ForAPIPath", "/api", fasthttp.StatusNotFound, false},
		{"ShouldReturn404ForAPISubpath", "/api/something", fasthttp.StatusNotFound, false},
		{"ShouldReturn404ForWellKnown", "/.well-known", fasthttp.StatusNotFound, false},
		{"ShouldReturn404ForWellKnownSubpath", "/.well-known/openid-configuration", fasthttp.StatusNotFound, false},
		{"ShouldReturn404ForStatic", "/static", fasthttp.StatusNotFound, false},
		{"ShouldReturn404ForStaticSubpath", "/static/js/app.js", fasthttp.StatusNotFound, false},
		{"ShouldReturn404ForLocales", "/locales", fasthttp.StatusNotFound, false},
		{"ShouldReturn404ForLocalesSubpath", "/locales/en/translation.json", fasthttp.StatusNotFound, false},
		{"ShouldCallNextForRootPath", "/", fasthttp.StatusOK, true},
		{"ShouldCallNextForUnknownPath", "/other", fasthttp.StatusOK, true},
		{"ShouldReturn404ForUppercaseAPI", "/API", fasthttp.StatusNotFound, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				ctx fasthttp.RequestCtx
				req fasthttp.Request
			)

			req.SetRequestURI(tc.path)
			ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

			handler(&ctx)

			assert.Equal(t, tc.expectedStatusCode, ctx.Response.StatusCode())

			if tc.expectNext {
				assert.Equal(t, "next called", string(ctx.Response.Body()))
			}
		})
	}
}

func TestHandleMethodNotAllowed(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{"ShouldReturn405"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var (
				ctx fasthttp.RequestCtx
				req fasthttp.Request
			)

			ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

			handleMethodNotAllowed(&ctx)

			assert.Equal(t, fasthttp.StatusMethodNotAllowed, ctx.Response.StatusCode())
			assert.Contains(t, string(ctx.Response.Body()), "405")
			assert.Contains(t, string(ctx.Response.Body()), "Method Not Allowed")
		})
	}
}

func TestHandlerMainWithAuthzEndpoints(t *testing.T) {
	provider, err := templates.New(templates.Config{})
	require.NoError(t, err)

	require.NoError(t, provider.LoadTemplatedAssets(assets))

	testCases := []struct {
		name  string
		authz map[string]schema.ServerEndpointsAuthz
	}{
		{
			"ShouldSucceedWithDefaultEndpoints",
			schema.DefaultServerConfiguration.Endpoints.Authz,
		},
		{
			"ShouldSucceedWithLegacyEndpoint",
			map[string]schema.ServerEndpointsAuthz{
				"legacy": {Implementation: schema.AuthzImplementationLegacy},
			},
		},
		{
			"ShouldSucceedWithExtAuthzEndpoint",
			map[string]schema.ServerEndpointsAuthz{
				"ext-authz": {Implementation: schema.AuthzImplementationExtAuthz},
			},
		},
		{
			"ShouldSucceedWithForwardAuthEndpoint",
			map[string]schema.ServerEndpointsAuthz{
				"forward-auth": {Implementation: schema.AuthzImplementationForwardAuth},
			},
		},
		{
			"ShouldSucceedWithAuthRequestEndpoint",
			map[string]schema.ServerEndpointsAuthz{
				"auth-request": {Implementation: schema.AuthzImplementationAuthRequest},
			},
		},
		{
			"ShouldSucceedWithNoAuthzEndpoints",
			nil,
		},
		{
			"ShouldSucceedWithMultipleEndpoints",
			map[string]schema.ServerEndpointsAuthz{
				"legacy":       {Implementation: schema.AuthzImplementationLegacy},
				"forward-auth": {Implementation: schema.AuthzImplementationForwardAuth},
				"ext-authz":    {Implementation: schema.AuthzImplementationExtAuthz},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.Configuration{
				Server: schema.Server{
					Address: schema.DefaultServerConfiguration.Address,
					Endpoints: schema.ServerEndpoints{
						Authz: tc.authz,
					},
				},
			}

			providers := middlewares.NewProvidersBasic()
			providers.Random = random.NewMathematical()
			providers.Templates = provider

			handler, err := handlerMain(config, providers)

			require.NoError(t, err)
			assert.NotNil(t, handler)
		})
	}
}

func TestHandlerMainWithOptionalFeatures(t *testing.T) {
	provider, err := templates.New(templates.Config{})
	require.NoError(t, err)

	require.NoError(t, provider.LoadTemplatedAssets(assets))

	testCases := []struct {
		name    string
		config  func() *schema.Configuration
		setOIDC bool
	}{
		{
			"ShouldSucceedWithPasskeyLogin",
			func() *schema.Configuration {
				return &schema.Configuration{
					Server: schema.Server{
						Address:   schema.DefaultServerConfiguration.Address,
						Endpoints: schema.DefaultServerConfiguration.Endpoints,
					},
					WebAuthn: schema.WebAuthn{
						EnablePasskeyLogin: true,
					},
				}
			},
			false,
		},
		{
			"ShouldSucceedWithPprof",
			func() *schema.Configuration {
				return &schema.Configuration{
					Server: schema.Server{
						Address: schema.DefaultServerConfiguration.Address,
						Endpoints: schema.ServerEndpoints{
							EnablePprof: true,
							Authz:       schema.DefaultServerConfiguration.Endpoints.Authz,
						},
					},
				}
			},
			false,
		},
		{
			"ShouldSucceedWithExpvars",
			func() *schema.Configuration {
				return &schema.Configuration{
					Server: schema.Server{
						Address: schema.DefaultServerConfiguration.Address,
						Endpoints: schema.ServerEndpoints{
							EnableExpvars: true,
							Authz:         schema.DefaultServerConfiguration.Endpoints.Authz,
						},
					},
				}
			},
			false,
		},
		{
			"ShouldSucceedWithOpenIDConnect",
			func() *schema.Configuration {
				return &schema.Configuration{
					Server: schema.Server{
						Address:   schema.DefaultServerConfiguration.Address,
						Endpoints: schema.DefaultServerConfiguration.Endpoints,
					},
					IdentityProviders: schema.IdentityProviders{
						OIDC: &schema.IdentityProvidersOpenIDConnect{},
					},
				}
			},
			true,
		},
		{
			"ShouldSucceedWithAllFeatures",
			func() *schema.Configuration {
				return &schema.Configuration{
					Server: schema.Server{
						Address: schema.DefaultServerConfiguration.Address,
						Endpoints: schema.ServerEndpoints{
							EnablePprof:   true,
							EnableExpvars: true,
							Authz:         schema.DefaultServerConfiguration.Endpoints.Authz,
						},
					},
					WebAuthn: schema.WebAuthn{
						EnablePasskeyLogin: true,
					},
					IdentityProviders: schema.IdentityProviders{
						OIDC: &schema.IdentityProvidersOpenIDConnect{},
					},
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := tc.config()

			providers := middlewares.NewProvidersBasic()
			providers.Random = random.NewMathematical()
			providers.Templates = provider

			if tc.setOIDC {
				providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, nil, provider)
			}

			handler, err := handlerMain(config, providers)

			require.NoError(t, err)
			assert.NotNil(t, handler)
		})
	}
}

type timeoutError struct{}

func (e *timeoutError) Error() string { return "i/o timeout" }

func (e *timeoutError) Timeout() bool { return true }

func (e *timeoutError) Temporary() bool { return true }

func TestIsAppAuthenticatorDevicePathShouldMatchDevicePaths(t *testing.T) {
	testCases := []struct {
		path     string
		expected bool
	}{
		{"/api/register/app", true},
		{"/api/register/app/abc", true},
		{"/api/register/app/confirm/abc/extra", true},
		{"/api/secondfactor/app/challenge/abc", true},
		{"/api/secondfactor/app/confirmation/abc", true},
		{"/api/secondfactor/app/device", true},
		{"/api/secondfactor/app/device/extra", true},
		{"/api/secondfactor/app", false},
		{"/api/secondfactor/app/status", false},
		{"/api/secondfactor/app/registration", false},
		{"/api/register/application", false},
		{"/api/secondfactor/totp", false},
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			assert.Equal(t, tc.expected, middlewares.IsAppAuthenticatorDevicePath(tc.path))
		})
	}
}

func TestHandlerMainShouldRouteAppAuthenticatorDevicePaths(t *testing.T) {
	provider, err := templates.New(templates.Config{})
	require.NoError(t, err)

	require.NoError(t, provider.LoadTemplatedAssets(assets))

	testCases := []struct {
		name         string
		disable      bool
		method       string
		path         string
		expectedCode int
		expectedBody string
	}{
		{"ShouldAnswerUnknownRoute", false, fasthttp.MethodPost, "/api/register/app/confirm/a/b", fasthttp.StatusNotFound, `"error":"unknown"`},
		{"ShouldAnswerUnknownMethod", false, fasthttp.MethodGet, "/api/secondfactor/app/device", fasthttp.StatusNotFound, `"error":"unknown"`},
		{"ShouldAnswerDefaultWhenDisabled", true, fasthttp.MethodPost, "/api/register/app/abc", fasthttp.StatusNotFound, "404 Not Found"},
		{"ShouldAnswerDefaultForOtherRoutes", false, fasthttp.MethodPost, "/api/secondfactor/unknown", fasthttp.StatusNotFound, "404 Not Found"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.Configuration{
				Server: schema.Server{
					Address:   schema.DefaultServerConfiguration.Address,
					Endpoints: schema.DefaultServerConfiguration.Endpoints,
				},
				AppAuthenticator: schema.AppAuthenticator{Disable: tc.disable},
				DuoAPI:           schema.DuoAPI{Disable: true},
			}

			providers := middlewares.NewProvidersBasic()
			providers.Random = random.NewMathematical()
			providers.Templates = provider

			handler, err := handlerMain(config, providers)
			require.NoError(t, err)

			var (
				ctx fasthttp.RequestCtx
				req fasthttp.Request
			)

			req.Header.SetMethod(tc.method)
			req.SetRequestURI(tc.path)
			req.Header.SetHost("auth.example.com")
			ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

			handler(&ctx)

			assert.Equal(t, tc.expectedCode, ctx.Response.StatusCode())
			assert.Contains(t, string(ctx.Response.Body()), tc.expectedBody)
		})
	}
}

func TestHandlerMainShouldRouteAppAuthenticatorPortalPaths(t *testing.T) {
	provider, err := templates.New(templates.Config{})
	require.NoError(t, err)

	require.NoError(t, provider.LoadTemplatedAssets(assets))

	routes := []struct {
		method string
		path   string
	}{
		{fasthttp.MethodPut, "/api/secondfactor/app/registration"},
		{fasthttp.MethodGet, "/api/secondfactor/app/registration"},
		{fasthttp.MethodDelete, "/api/secondfactor/app/registration"},
		{fasthttp.MethodPost, "/api/secondfactor/app"},
		{fasthttp.MethodPost, "/api/secondfactor/app/status"},
		{fasthttp.MethodGet, "/api/secondfactor/app/devices"},
		{fasthttp.MethodPut, "/api/secondfactor/app/devices/1"},
		{fasthttp.MethodDelete, "/api/secondfactor/app/devices/1"},
	}

	for _, disable := range []bool{false, true} {
		config := &schema.Configuration{
			Server: schema.Server{
				Address:   schema.DefaultServerConfiguration.Address,
				Endpoints: schema.DefaultServerConfiguration.Endpoints,
			},
			AppAuthenticator: schema.AppAuthenticator{Disable: disable},
			DuoAPI:           schema.DuoAPI{Disable: true},
		}

		providers := middlewares.NewProvidersBasic()
		providers.Random = random.NewMathematical()
		providers.Templates = provider

		handler, err := handlerMain(config, providers)
		require.NoError(t, err)

		for _, route := range routes {
			t.Run(fmt.Sprintf("Disable%t/%s%s", disable, route.method, route.path), func(t *testing.T) {
				var (
					ctx fasthttp.RequestCtx
					req fasthttp.Request
				)

				req.Header.SetMethod(route.method)
				req.SetRequestURI(route.path)
				req.Header.SetHost("auth.example.com")
				ctx.Init(&req, &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 8080}, nil)

				handler(&ctx)

				if disable {
					assert.Equal(t, fasthttp.StatusNotFound, ctx.Response.StatusCode())
				} else {
					assert.Equal(t, fasthttp.StatusForbidden, ctx.Response.StatusCode())
				}
			})
		}
	}
}
