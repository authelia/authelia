// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"github.com/valyala/fasthttp"
)

// The SecurityHeadersBase middleware adds several modern recommended security headers with relaxed values.
func SecurityHeadersBase(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		SetBaseSecurityHeaders(ctx)

		next(ctx)
	}
}

// The SecurityHeadersCORPCrossOrigin middleware overrides the Cross-Origin-Resource-Policy header with the value
// 'cross-origin'. It must be applied after SecurityHeadersBase, which sets the restrictive value it overrides.
func SecurityHeadersCORPCrossOrigin(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		SetSecurityHeadersCORPCrossOrigin(ctx)

		next(ctx)
	}
}

// SetBaseSecurityHeaders sets the security headers applied to every response.
func SetBaseSecurityHeaders(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetBytesKV(headerXContentTypeOptions, headerValueNoSniff)
	ctx.Response.Header.SetBytesKV(headerReferrerPolicy, headerValueStrictOriginCrossOrigin)
	ctx.Response.Header.SetBytesKV(headerPermissionsPolicy, headerValuePermissionsPolicy)
	ctx.Response.Header.SetBytesKV(headerXFrameOptions, headerValueDENY)
	ctx.Response.Header.SetBytesKV(headerXDNSPrefetchControl, headerValueOff)
	ctx.Response.Header.SetBytesKV(headerCrossOriginResourcePolicy, headerValueSameOrigin)
}

// SetSecurityHeadersCORPCrossOrigin sets the Cross-Origin-Resource-Policy header with the value 'cross-origin', for the
// deliberately public responses which a Relying Party is expected to fetch from another origin.
func SetSecurityHeadersCORPCrossOrigin(ctx *fasthttp.RequestCtx) {
	ctx.Response.Header.SetBytesKV(headerCrossOriginResourcePolicy, headerValueCrossOrigin)
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
