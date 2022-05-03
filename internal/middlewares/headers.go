package middlewares

import (
	"github.com/valyala/fasthttp"
)

// SecurityHeaders middleware adds several modern recommended security headers with safe values.
func SecurityHeaders(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetBytesKV(headerXContentTypeOptions, headerValueNoSniff)
		ctx.Response.Header.SetBytesKV(headerReferrerPolicy, headerValueStrictOriginCrossOrigin)
		ctx.Response.Header.SetBytesKV(headerPermissionsPolicy, headerValueCohort)
		ctx.Response.Header.SetBytesKV(headerXFrameOptions, headerValueSAMEORIGIN)
		ctx.Response.Header.SetBytesKV(headerXXSSProtection, headerValueXSSDisabled)

		ctx.Response.Header.SetBytesKV(headerCrossOriginResourcePolicy, headerValueSameOrigin)
		ctx.Response.Header.SetBytesKV(headerCrossOriginEmbedderPolicy, headerValueRequireCORP)
		ctx.Response.Header.SetBytesKV(headerCrossOriginOpenerPolicy, headerValueSameOrigin)

		next(ctx)
	}
}

// SecurityHeadersCSPNone middleware adds the Content-Security-Policy header with the value "default-src 'none';".
func SecurityHeadersCSPNone(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetBytesKV(headerContentSecurityPolicy, headerValueCSPNone)

		next(ctx)
	}
}

// SecurityHeadersNoStore middleware adds the Pragma no-cache and Cache-Control no-store headers.
func SecurityHeadersNoStore(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetBytesKV(headerPragma, headerValueNoCache)
		ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueNoStore)

		next(ctx)
	}
}
