package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewCORSMiddleware(t *testing.T) {
	cors := NewCORSPolicyBuilder()

	assert.Equal(t, 100, cors.maxAge)
	assert.Equal(t, false, cors.credentials)

	assert.Nil(t, cors.methods)
	assert.Nil(t, cors.origins)
	assert.Nil(t, cors.headers)
	assert.Nil(t, cors.vary)
	assert.False(t, cors.varyOnly)
	assert.False(t, cors.varySet)
}

func TestCORSPolicyBuilder_WithEnabled(t *testing.T) {
	cors := NewCORSPolicyBuilder()

	assert.True(t, cors.enabled)

	cors.WithEnabled(false)
	assert.False(t, cors.enabled)
}

func TestCORSPolicyBuilder_WithVary(t *testing.T) {
	cors := NewCORSPolicyBuilder()

	assert.Nil(t, cors.vary)
	assert.False(t, cors.varyOnly)
	assert.False(t, cors.varySet)

	cors.WithVary()
	assert.Nil(t, cors.vary)
	assert.False(t, cors.varyOnly)
	assert.True(t, cors.varySet)

	cors.WithVary("Origin", "Example", "Test")

	assert.Equal(t, []string{"Origin", "Example", "Test"}, cors.vary)
	assert.False(t, cors.varyOnly)
	assert.True(t, cors.varySet)
}

func TestCORSPolicyBuilder_WithAllowedMethods(t *testing.T) {
	cors := NewCORSPolicyBuilder()

	assert.Nil(t, cors.methods)

	cors.WithAllowedMethods(fasthttp.MethodGet)

	assert.Equal(t, []string{fasthttp.MethodGet}, cors.methods)

	cors.WithAllowedMethods(fasthttp.MethodPost, fasthttp.MethodPatch)

	assert.Equal(t, []string{fasthttp.MethodPost, fasthttp.MethodPatch}, cors.methods)

	cors.WithAllowedMethods()

	assert.Nil(t, cors.methods)
}

func TestCORSPolicyBuilder_WithAllowedOrigins(t *testing.T) {
	cors := NewCORSPolicyBuilder()

	assert.Nil(t, cors.origins)

	cors.WithAllowedOrigins("https://google.com", "http://localhost")

	assert.Equal(t, []string{"https://google.com", "http://localhost"}, cors.origins)

	cors.WithAllowedOrigins()

	assert.Nil(t, cors.origins)
}

func TestCORSPolicyBuilder_WithAllowedHeaders(t *testing.T) {
	cors := NewCORSPolicyBuilder()

	assert.Nil(t, cors.headers)

	cors.WithAllowedHeaders("Example", "Another")

	assert.Equal(t, []string{"Example", "Another"}, cors.headers)

	cors.WithAllowedHeaders()

	assert.Nil(t, cors.headers)
}

func TestCORSPolicyBuilder_WithAllowCredentials(t *testing.T) {
	cors := NewCORSPolicyBuilder()

	assert.Equal(t, false, cors.credentials)

	cors.WithAllowCredentials(false)

	assert.Equal(t, false, cors.credentials)

	cors.WithAllowCredentials(true)

	assert.Equal(t, true, cors.credentials)
}

func TestCORSPolicyBuilder_WithVaryOnly(t *testing.T) {
	cors := NewCORSPolicyBuilder()

	assert.False(t, cors.varyOnly)

	cors.WithVaryOnly(false)

	assert.False(t, cors.varyOnly)

	cors.WithVaryOnly(true)

	cors.WithVaryOnly(true)
}

func TestCORSPolicyBuilder_WithMaxAge(t *testing.T) {
	cors := NewCORSPolicyBuilder()

	assert.Equal(t, 100, cors.maxAge)

	cors.WithMaxAge(20)

	assert.Equal(t, 20, cors.maxAge)

	cors.WithMaxAge(0)

	assert.Equal(t, 0, cors.maxAge)
}

func TestCORSPolicyBuilder_HandleOPTIONS(t *testing.T) {
	testCases := []struct {
		name                        string
		enabled                     bool
		methods                     []string
		headers                     []string
		origins                     []string
		vary                        []string
		varyOnly                    bool
		maxage                      int
		requestMethod               string
		requestHeaderOrigin         string
		requestHeaderACRM           string
		requestHeaderACRH           string
		expectedStatus              int
		expectedHeaderContentLength any
		expectedHeaderVary          any
		expectedHeaderACAO          any
		expectedHeaderACAM          any
		expectedHeaderACAH          any
		expectedHeaderACAC          any
		expectedHeaderACMA          any
	}{
		{
			name:                        "ShouldHandleWildcardPreflight",
			enabled:                     true,
			methods:                     []string{fasthttp.MethodGet, fasthttp.MethodOptions},
			headers:                     nil,
			origins:                     []string{"*"},
			vary:                        nil,
			varyOnly:                    false,
			maxage:                      0,
			requestMethod:               fasthttp.MethodOptions,
			requestHeaderOrigin:         "https://myapp.example.com",
			requestHeaderACRM:           fasthttp.MethodGet,
			requestHeaderACRH:           "X-Example-Header",
			expectedStatus:              fasthttp.StatusOK,
			expectedHeaderContentLength: []byte("0"),
			expectedHeaderVary:          []byte(nil),
			expectedHeaderACAO:          []byte("*"),
			expectedHeaderACAM:          []byte("GET, OPTIONS"),
			expectedHeaderACAH:          []byte("X-Example-Header"),
			expectedHeaderACAC:          []byte("false"),
			expectedHeaderACMA:          []byte("100"),
		},
		{
			name:                        "ShouldHandleWildcard",
			enabled:                     true,
			methods:                     []string{fasthttp.MethodGet, fasthttp.MethodOptions},
			headers:                     nil,
			origins:                     []string{"*"},
			vary:                        nil,
			varyOnly:                    false,
			maxage:                      0,
			requestMethod:               fasthttp.MethodGet,
			requestHeaderOrigin:         "https://myapp.example.com",
			requestHeaderACRM:           "",
			requestHeaderACRH:           "",
			expectedStatus:              fasthttp.StatusOK,
			expectedHeaderContentLength: []byte("0"),
			expectedHeaderVary:          []byte(nil),
			expectedHeaderACAO:          []byte("*"),
			expectedHeaderACAM:          []byte("GET, OPTIONS"),
			expectedHeaderACAH:          []byte(nil),
			expectedHeaderACAC:          []byte("false"),
			expectedHeaderACMA:          []byte("100"),
		},
		{
			name:                        "ShouldHandleNonPreflight",
			enabled:                     true,
			methods:                     []string{fasthttp.MethodGet, fasthttp.MethodOptions},
			headers:                     nil,
			origins:                     []string{"*"},
			vary:                        nil,
			varyOnly:                    false,
			maxage:                      0,
			requestMethod:               fasthttp.MethodOptions,
			requestHeaderOrigin:         "",
			requestHeaderACRM:           "",
			requestHeaderACRH:           "X-Example-Header",
			expectedStatus:              fasthttp.StatusOK,
			expectedHeaderContentLength: []byte("0"),
			expectedHeaderVary:          []byte(nil),
			expectedHeaderACAO:          []byte(nil),
			expectedHeaderACAM:          []byte(nil),
			expectedHeaderACAH:          []byte(nil),
			expectedHeaderACAC:          []byte(nil),
			expectedHeaderACMA:          []byte(nil),
		},
		{
			name:                        "ShouldHandleCORSRequest",
			enabled:                     true,
			methods:                     []string{fasthttp.MethodGet, fasthttp.MethodOptions},
			headers:                     nil,
			origins:                     []string{"*"},
			vary:                        nil,
			varyOnly:                    false,
			maxage:                      0,
			requestMethod:               fasthttp.MethodGet,
			requestHeaderOrigin:         "https://app.example.com",
			requestHeaderACRM:           "",
			requestHeaderACRH:           "X-Example-Header",
			expectedStatus:              fasthttp.StatusOK,
			expectedHeaderContentLength: []byte("0"),
			expectedHeaderVary:          []byte(nil),
			expectedHeaderACAO:          []byte("*"),
			expectedHeaderACAM:          []byte("GET, OPTIONS"),
			expectedHeaderACAH:          []byte("X-Example-Header"),
			expectedHeaderACAC:          []byte("false"),
			expectedHeaderACMA:          []byte("100"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newFastHTTPRequestCtx()

			if len(tc.requestMethod) != 0 {
				ctx.Request.Header.SetMethod(tc.requestMethod)
			}

			if len(tc.requestHeaderOrigin) != 0 {
				ctx.Request.Header.SetBytesK(headerOrigin, tc.requestHeaderOrigin)
			}

			if len(tc.requestHeaderACRM) != 0 {
				ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, tc.requestHeaderACRM)
			}

			if len(tc.requestHeaderACRH) != 0 {
				ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, tc.requestHeaderACRH)
			}

			NewCORSPolicyBuilder().
				WithEnabled(tc.enabled).
				WithAllowedMethods(tc.methods...).
				WithAllowedHeaders(tc.headers...).
				WithAllowedOrigins(tc.origins...).
				WithVary(tc.vary...).
				WithVaryOnly(tc.varyOnly).
				WithMaxAge(tc.maxage).
				Build().
				HandleOPTIONS(ctx)

			assert.Equal(t, tc.expectedStatus, ctx.Response.StatusCode())
			assert.Equal(t, tc.expectedHeaderContentLength, ctx.Response.Header.PeekBytes(headerContentLength))
			assert.Equal(t, tc.expectedHeaderVary, ctx.Response.Header.PeekBytes(headerVary))
			assert.Equal(t, tc.expectedHeaderACAO, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
			assert.Equal(t, tc.expectedHeaderACAC, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
			assert.Equal(t, tc.expectedHeaderACAH, ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
			assert.Equal(t, tc.expectedHeaderACAM, ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
			assert.Equal(t, tc.expectedHeaderACMA, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
		})
	}
}

func TestCORSPolicyBuilder_HandleOPTIONS_TODO_Legacy_Replace(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesKV(headerOrigin, origin)
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)
	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")

	cors := NewCORSPolicyBuilder()
	policy := cors.Build()

	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte(fasthttp.MethodGet), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(fasthttp.MethodGet), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	ctx = newFastHTTPRequestCtx()

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowedMethods(fasthttp.MethodGet, fasthttp.MethodOptions)

	policy = cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	ctx = newFastHTTPRequestCtx()

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	policy = cors.Build()
	policy.HandleOnlyOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	ctx = newFastHTTPRequestCtx()

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithEnabled(false)

	policy = cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSPolicyBuilder_HandleOPTIONS_WithoutOrigin(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")

	cors := NewCORSPolicyBuilder()

	policy := cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	ctx = newFastHTTPRequestCtx()

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")

	cors.WithAllowedMethods(fasthttp.MethodGet, fasthttp.MethodOptions)

	policy = cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSPolicyBuilder_HandleOPTIONSWithAllowedOrigins(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors := NewCORSPolicyBuilder()
	cors.WithAllowedOrigins("https://myapp.example.com")

	policy := cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(fasthttp.MethodGet), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	ctx = newFastHTTPRequestCtx()
	ctx.Request.Header.SetMethod(fasthttp.MethodOptions)
	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowedOrigins("https://anotherapp.example.com")

	policy = cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	ctx = newFastHTTPRequestCtx()

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowedOrigins("*")
	cors.WithAllowedMethods(fasthttp.MethodGet, fasthttp.MethodOptions)

	policy = cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
	assert.Equal(t, []byte("Accept-Encoding"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, headerValueOriginWildcard, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSPolicyBuilder_WithAllowedOrigins_DoesntOverrideVary(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors := NewCORSPolicyBuilder()
	cors.WithVary("Accept-Encoding", "Origin", "Test")
	cors.WithAllowedOrigins("*")

	policy := cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte("Accept-Encoding, Origin, Test"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, headerValueOriginWildcard, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(fasthttp.MethodGet), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSPolicyBuilder_HandleOPTIONSWithVaryOnly(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors := NewCORSPolicyBuilder()

	cors.WithVaryOnly(true)

	policy := cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	ctx = newFastHTTPRequestCtx()

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowedMethods(fasthttp.MethodGet, fasthttp.MethodOptions)

	policy = cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSPolicyBuilder_HandleOPTIONSWithAllowedHeaders(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors := NewCORSPolicyBuilder()

	cors.WithAllowedHeaders("Example", "Test")

	policy := cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("Example, Test"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(fasthttp.MethodGet), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	ctx = newFastHTTPRequestCtx()

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowedMethods(fasthttp.MethodGet, fasthttp.MethodOptions)

	policy = cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("Example, Test"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))

	ctx = newFastHTTPRequestCtx()

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors.WithAllowCredentials(true)

	policy = cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueTrue, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("Example, Test, Cookie, Authorization, Proxy-Authorization"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func TestCORSPolicyBuilder_HandleOPTIONS_ShouldNotAllowWildcardInRequestedHeaders(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "*")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)

	cors := NewCORSPolicyBuilder()

	policy := cors.Build()
	policy.HandleOPTIONS(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, headerValueZero, ctx.Response.Header.PeekBytes(headerContentLength))
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(fasthttp.MethodGet), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func Test_CORSApplyAutomaticAllowAllPolicy_WithoutRequestMethod(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	origin := []byte("https://myapp.example.com")
	ctx.Request.Header.SetBytesKV(headerOrigin, origin)
	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)

	cors := NewCORSPolicyBuilder()

	policy := cors.Build()
	policy.handle(ctx)

	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(fasthttp.MethodGet), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func Test_CORSApplyAutomaticAllowAllPolicy_WithRequestMethod(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesKV(headerOrigin, origin)
	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)

	cors := NewCORSPolicyBuilder()

	policy := cors.Build()
	policy.handle(ctx)

	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(fasthttp.MethodGet), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func Test_CORSApplyAutomaticAllowAllPolicy_ShouldNotModifyFotNonHTTPSRequests(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	origin := []byte("http://myapp.example.com")

	ctx.Request.Header.SetBytesKV(headerOrigin, origin)
	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)

	cors := NewCORSPolicyBuilder().WithVary()

	policy := cors.Build()
	policy.handle(ctx)

	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte(nil), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func Test_CORSMiddleware_AsMiddleware(t *testing.T) {
	ctx := newFastHTTPRequestCtx()

	origin := []byte("https://myapp.example.com")

	ctx.Request.Header.SetBytesKV(headerOrigin, origin)
	ctx.Request.Header.SetBytesK(headerAccessControlRequestHeaders, "X-Example-Header")
	ctx.Request.Header.SetBytesK(headerAccessControlRequestMethod, fasthttp.MethodGet)

	middleware := NewBridgeBuilder(schema.Configuration{}, Providers{}).Build()

	cors := NewCORSPolicyBuilder().WithAllowedMethods(fasthttp.MethodGet, fasthttp.MethodOptions)

	policy := cors.Build()

	route := policy.Middleware(middleware(testNilHandler))

	route(ctx)

	assert.Equal(t, fasthttp.StatusOK, ctx.Response.StatusCode())
	assert.Equal(t, []byte("Accept-Encoding, Origin"), ctx.Response.Header.PeekBytes(headerVary))
	assert.Equal(t, origin, ctx.Response.Header.PeekBytes(headerAccessControlAllowOrigin))
	assert.Equal(t, headerValueFalse, ctx.Response.Header.PeekBytes(headerAccessControlAllowCredentials))
	assert.Equal(t, headerValueMaxAge, ctx.Response.Header.PeekBytes(headerAccessControlMaxAge))
	assert.Equal(t, []byte("X-Example-Header"), ctx.Response.Header.PeekBytes(headerAccessControlAllowHeaders))
	assert.Equal(t, []byte("GET, OPTIONS"), ctx.Response.Header.PeekBytes(headerAccessControlAllowMethods))
}

func testNilHandler(_ *AutheliaCtx) {}

func newFastHTTPRequestCtx() (ctx *fasthttp.RequestCtx) {
	return &fasthttp.RequestCtx{
		Request:  fasthttp.Request{},
		Response: fasthttp.Response{},
	}
}
