package middlewares

import (
	"github.com/valyala/fasthttp"
)

// The SecurityHeaders middleware adds several modern recommended security headers with safe values.
func SecurityHeaders(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		SetStandardSecurityHeaders(ctx)

		next(ctx)
	}
}

// The SecurityHeadersRelaxed middleware adds several modern recommended security headers with relaxed values.
func SecurityHeadersRelaxed(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		SetRelaxedSecurityHeaders(ctx)

		next(ctx)
	}
}

// The SecurityHeadersBase middleware adds several modern recommended security headers with relaxed values.
func SecurityHeadersBase(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		SetBaseSecurityHeaders(ctx)

		next(ctx)
	}
}

// The SetStandardSecurityHeaders function adds several modern recommended security headers with safe values.
func SetStandardSecurityHeaders(ctx *fasthttp.RequestCtx) {
	SetBaseSecurityHeaders(ctx)

	ctx.Response.Header.SetBytesKV(headerCrossOriginOpenerPolicy, headerValueSameOrigin)
	ctx.Response.Header.SetBytesKV(headerCrossOriginEmbedderPolicy, headerValueRequireCORP)
	ctx.Response.Header.SetBytesKV(headerCrossOriginResourcePolicy, headerValueSameSite)
}

// The SetRelaxedSecurityHeaders function adds several modern recommended security headers with relaxed values.
func SetRelaxedSecurityHeaders(ctx *fasthttp.RequestCtx) {
	SetBaseSecurityHeaders(ctx)

	ctx.Response.Header.SetBytesKV(headerCrossOriginOpenerPolicy, headerValueSameOrigin)
	ctx.Response.Header.SetBytesKV(headerCrossOriginEmbedderPolicy, headerValueUnsafeNone)
	ctx.Response.Header.SetBytesKV(headerCrossOriginResourcePolicy, headerValueCrossOrigin)
}

func SetBaseSecurityHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetBytesKV(headerXContentTypeOptions, headerValueNoSniff)
	ctx.Response.Header.SetBytesKV(headerReferrerPolicy, headerValueStrictOriginCrossOrigin)
	ctx.Response.Header.SetBytesKV(headerPermissionsPolicy, headerValuePermissionsPolicy)
	ctx.Response.Header.SetBytesKV(headerXFrameOptions, headerValueDENY)
	ctx.Response.Header.SetBytesKV(headerXDNSPrefetchControl, headerValueOff)
}

// SecurityHeadersCSPNone middleware adds the Content-Security-Policy header with the value "default-src 'none';".
func SecurityHeadersCSPNone(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		SetSecurityHeadersCSPNone(ctx)

		next(ctx)
	}
}

// SecurityHeadersCSPNoneOpenIDConnect middleware adds the Content-Security-Policy header with the value
// "default-src 'none'" except in special circumstances.
func SecurityHeadersCSPNoneOpenIDConnect(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetUserValue(UserValueKeyOpenIDConnectResponseModeFormPost, false)

		next(ctx)

		if modeFormPost, ok := ctx.UserValue(UserValueKeyOpenIDConnectResponseModeFormPost).(bool); ok && modeFormPost {
			ctx.Response.Header.SetBytesKV(headerContentSecurityPolicy, headerValueCSPNoneFormPost)
		} else {
			ctx.Response.Header.SetBytesKV(headerContentSecurityPolicy, headerValueCSPNone)
		}
	}
}

// SecurityHeadersCSPSelf middleware adds the Content-Security-Policy header with the value "default-src 'self';".
func SecurityHeadersCSPSelf(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetBytesKV(headerContentSecurityPolicy, headerValueCSPSelf)

		next(ctx)
	}
}

// SetSecurityHeadersCSPNone function adds the Content-Security-Policy header with the value "default-src 'none';".
func SetSecurityHeadersCSPNone(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetBytesKV(headerContentSecurityPolicy, headerValueCSPNone)
}

// SecurityHeadersNoStore middleware adds the Pragma no-cache and Cache-Control no-store headers.
func SecurityHeadersNoStore(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetBytesKV(headerPragma, headerValueNoCache)
		ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueNoStore)

		next(ctx)
	}
}
