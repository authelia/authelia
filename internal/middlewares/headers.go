package middlewares

import (
	"github.com/valyala/fasthttp"
)

// The SecurityHeaders middleware adds several modern recommended security headers with safe values.
func SecurityHeaders(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		SetSecurityHeaders(ctx)

		next(ctx)
	}
}

// The SetSecurityHeaders function adds several modern recommended security headers with safe values.
func SetSecurityHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetBytesKV(headerXContentTypeOptions, headerValueNoSniff)
	ctx.Response.Header.SetBytesKV(headerReferrerPolicy, headerValueStrictOriginCrossOrigin)
	ctx.Response.Header.SetBytesKV(headerPermissionsPolicy, headerValuePermissionsPolicy)
	ctx.Response.Header.SetBytesKV(headerXFrameOptions, headerValueDENY)
	ctx.Response.Header.SetBytesKV(headerXDNSPrefetchControl, headerValueOff)
	ctx.Response.Header.SetBytesKV(headerCrossOriginOpenerPolicy, headerValueSameOrigin)
	ctx.Response.Header.SetBytesKV(headerCrossOriginEmbedderPolicy, headerValueRequireCORP)
	ctx.Response.Header.SetBytesKV(headerCrossOriginResourcePolicy, headerValueSameSite)
}

// SecurityHeadersCSPNone middleware adds the Content-Security-Policy header with the value "default-src 'none';".
func SecurityHeadersCSPNone(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetBytesKV(headerContentSecurityPolicy, headerValueCSPNone)

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

// SecurityHeadersNoStore middleware adds the Pragma no-cache and Cache-Control no-store headers.
func SecurityHeadersNoStore(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.SetBytesKV(headerPragma, headerValueNoCache)
		ctx.Response.Header.SetBytesKV(headerCacheControl, headerValueNoStore)

		next(ctx)
	}
}
