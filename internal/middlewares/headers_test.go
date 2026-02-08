package middlewares

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestHeaders(t *testing.T) {
	testCases := []struct {
		name       string
		middleware Middleware
		setup      func(ctx *fasthttp.RequestCtx)
		expect     func(t *testing.T, ctx *fasthttp.RequestCtx)
	}{
		{
			"ShouldHandleSecurityHeaders",
			SecurityHeaders,
			nil,
			func(t *testing.T, ctx *fasthttp.RequestCtx) {
				assert.Equal(t, "nosniff", string(ctx.Response.Header.Peek(fasthttp.HeaderXContentTypeOptions)))
				assert.Equal(t, "strict-origin-when-cross-origin", string(ctx.Response.Header.Peek(fasthttp.HeaderReferrerPolicy)))
				assert.Equal(t, "accelerometer=(), autoplay=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), payment=(), picture-in-picture=(), screen-wake-lock=(), sync-xhr=(), xr-spatial-tracking=(), interest-cohort=()", string(ctx.Response.Header.Peek("Permissions-Policy")))
				assert.Equal(t, "DENY", string(ctx.Response.Header.Peek(fasthttp.HeaderXFrameOptions)))
				assert.Equal(t, "off", string(ctx.Response.Header.Peek("X-DNS-Prefetch-Control")))

				assert.Equal(t, "same-origin", string(ctx.Response.Header.Peek("Cross-Origin-Opener-Policy")))
				assert.Equal(t, "require-corp", string(ctx.Response.Header.Peek("Cross-Origin-Embedder-Policy")))
				assert.Equal(t, "same-site", string(ctx.Response.Header.Peek("Cross-Origin-Resource-Policy")))
			},
		},
		{
			"ShouldHandleSecurityHeadersRelaxed",
			SecurityHeadersRelaxed,
			nil,
			func(t *testing.T, ctx *fasthttp.RequestCtx) {
				assert.Equal(t, "nosniff", string(ctx.Response.Header.Peek(fasthttp.HeaderXContentTypeOptions)))
				assert.Equal(t, "strict-origin-when-cross-origin", string(ctx.Response.Header.Peek(fasthttp.HeaderReferrerPolicy)))
				assert.Equal(t, "accelerometer=(), autoplay=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), payment=(), picture-in-picture=(), screen-wake-lock=(), sync-xhr=(), xr-spatial-tracking=(), interest-cohort=()", string(ctx.Response.Header.Peek("Permissions-Policy")))
				assert.Equal(t, "DENY", string(ctx.Response.Header.Peek(fasthttp.HeaderXFrameOptions)))
				assert.Equal(t, "off", string(ctx.Response.Header.Peek("X-DNS-Prefetch-Control")))

				assert.Equal(t, "same-origin", string(ctx.Response.Header.Peek("Cross-Origin-Opener-Policy")))
				assert.Equal(t, "unsafe-none", string(ctx.Response.Header.Peek("Cross-Origin-Embedder-Policy")))
				assert.Equal(t, "cross-origin", string(ctx.Response.Header.Peek("Cross-Origin-Resource-Policy")))
			},
		},
		{
			"ShouldHandleSecurityHeadersBase",
			SecurityHeadersBase,
			nil,
			func(t *testing.T, ctx *fasthttp.RequestCtx) {
				assert.Equal(t, "nosniff", string(ctx.Response.Header.Peek(fasthttp.HeaderXContentTypeOptions)))
				assert.Equal(t, "strict-origin-when-cross-origin", string(ctx.Response.Header.Peek(fasthttp.HeaderReferrerPolicy)))
				assert.Equal(t, "accelerometer=(), autoplay=(), camera=(), display-capture=(), geolocation=(), gyroscope=(), keyboard-map=(), magnetometer=(), microphone=(), midi=(), payment=(), picture-in-picture=(), screen-wake-lock=(), sync-xhr=(), xr-spatial-tracking=(), interest-cohort=()", string(ctx.Response.Header.Peek("Permissions-Policy")))
				assert.Equal(t, "DENY", string(ctx.Response.Header.Peek(fasthttp.HeaderXFrameOptions)))
				assert.Equal(t, "off", string(ctx.Response.Header.Peek("X-DNS-Prefetch-Control")))
			},
		},
		{
			"ShouldHandleSecurityHeadersCSPNone",
			SecurityHeadersCSPNone,
			nil,
			func(t *testing.T, ctx *fasthttp.RequestCtx) {
				assert.Equal(t, "default-src 'none'", string(ctx.Response.Header.Peek(fasthttp.HeaderContentSecurityPolicy)))
			},
		},
		{
			"ShouldHandleSecurityHeadersCSPNoneOpenIDConnect",
			SecurityHeadersCSPNoneOpenIDConnect,
			nil,
			func(t *testing.T, ctx *fasthttp.RequestCtx) {
				assert.Equal(t, "default-src 'none'", string(ctx.Response.Header.Peek(fasthttp.HeaderContentSecurityPolicy)))
			},
		},
		{
			"ShouldHandleSecurityHeadersCSPNoneOpenIDConnectFormPost",
			SecurityHeadersCSPNoneOpenIDConnect,
			func(ctx *fasthttp.RequestCtx) {
				ctx.SetUserValue(UserValueKeyOpenIDConnectResponseModeFormPost, true)
			},
			func(t *testing.T, ctx *fasthttp.RequestCtx) {
				assert.Equal(t, "default-src 'none'; script-src 'sha256-skflBqA90WuHvoczvimLdj49ExKdizFjX2Itd6xKZdU='", string(ctx.Response.Header.Peek(fasthttp.HeaderContentSecurityPolicy)))
			},
		},
		{
			"ShouldHandleSecurityHeadersCSPSelf",
			SecurityHeadersCSPSelf,
			nil,
			func(t *testing.T, ctx *fasthttp.RequestCtx) {
				assert.Equal(t, "default-src 'self'", string(ctx.Response.Header.Peek(fasthttp.HeaderContentSecurityPolicy)))
			},
		},
		{
			"ShouldHandleSecurityHeadersNoStore",
			SecurityHeadersNoStore,
			nil,
			func(t *testing.T, ctx *fasthttp.RequestCtx) {
				assert.Equal(t, "no-cache", string(ctx.Response.Header.Peek(fasthttp.HeaderPragma)))
				assert.Equal(t, "no-store", string(ctx.Response.Header.Peek(fasthttp.HeaderCacheControl)))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			next := func(ctx *fasthttp.RequestCtx) {
				if tc.setup != nil {
					tc.setup(ctx)
				}
			}

			handler := tc.middleware(next)

			ctx := &fasthttp.RequestCtx{
				Request: fasthttp.Request{},
			}

			handler(ctx)

			tc.expect(t, ctx)
		})
	}
}
