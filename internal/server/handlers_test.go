package server

import (
	"fmt"
	"net"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestHandleError(t *testing.T) {
	handler := handleError("server")

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

// TestHandlerMainOpenIDConnectLinkRoutes drives the real registered router (not a hand-built stand-in) to prove two
// things about the OpenID Connect 1.0 relying party link management routes which a handler unit test cannot: that
// DELETE /api/user/openid-connect/link/pending resolves to the pending decline handler rather than the id delete
// handler with linkID == "pending", and that the accept (PUT) and id delete (DELETE) routes are actually wired
// behind the elevated session middleware rather than the plain first factor one. An anonymous request is enough to
// tell these apart: middleware1FA replies with a plain text "403 Forbidden" body while middlewareElevated1FA replies
// with a JSON body carrying "first_factor":true, so the two failure shapes reveal which middleware a route actually
// runs behind.
func TestHandlerMainOpenIDConnectLinkRoutes(t *testing.T) {
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
			OpenIDConnect: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{ID: "example", Name: "Example", Issuer: "https://op.example.com"},
				},
			},
		},
	}

	providers := middlewares.NewProvidersBasic()
	providers.Random = random.NewMathematical()
	providers.Templates = provider
	providers.SessionProvider = session.NewProvider(config.Session, nil)

	handler, err := handlerMain(config, providers)

	require.NoError(t, err)
	require.NotNil(t, handler)

	testCases := []struct {
		Name                 string
		Method               string
		Path                 string
		ExpectedStatusCode   int
		ExpectedContentType  string
		ExpectedBodyContains string
	}{
		{
			"ShouldRoutePendingDeclineToThePlainFirstFactorMiddleware",
			fasthttp.MethodDelete,
			"/api/user/openid-connect/link/pending",
			fasthttp.StatusForbidden,
			"text/plain; charset=utf-8",
			"403 Forbidden",
		},
		{
			"ShouldRouteIDDeleteToTheElevatedMiddleware",
			fasthttp.MethodDelete,
			"/api/user/openid-connect/link/123",
			fasthttp.StatusForbidden,
			"application/json; charset=utf-8",
			`"first_factor":true`,
		},
		{
			"ShouldRouteAcceptToTheElevatedMiddleware",
			fasthttp.MethodPut,
			"/api/user/openid-connect/link",
			fasthttp.StatusForbidden,
			"application/json; charset=utf-8",
			`"first_factor":true`,
		},
		{
			"ShouldRouteLinksListToThePlainFirstFactorMiddleware",
			fasthttp.MethodGet,
			"/api/user/openid-connect/links",
			fasthttp.StatusForbidden,
			"text/plain; charset=utf-8",
			"403 Forbidden",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := newAssetRequestCtx(tc.Method, tc.Path, "")
			ctx.Request.Header.SetHost("login.example.com")

			handler(ctx)

			assert.Equal(t, tc.ExpectedStatusCode, ctx.Response.StatusCode())
			assert.Equal(t, tc.ExpectedContentType, string(ctx.Response.Header.ContentType()))
			assert.Contains(t, string(ctx.Response.Body()), tc.ExpectedBodyContains)
		})
	}
}

type timeoutError struct{}

func (e *timeoutError) Error() string { return "i/o timeout" }

func (e *timeoutError) Timeout() bool { return true }

func (e *timeoutError) Temporary() bool { return true }
