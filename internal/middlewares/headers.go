// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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
		ctx.Response.Header.SetBytesKV(headerXFrameOptions, headerValueSameOrigin)
		ctx.Response.Header.SetBytesKV(headerXXSSProtection, headerValueXSSModeBlock)

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

// SecurityHeadersCSPNoneOpenIDConnect middleware adds the Content-Security-Policy header with the value
// "default-src 'none'" except in special circumstances.
func SecurityHeadersCSPNoneOpenIDConnect(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.SetUserValueBytes(UserValueKeyFormPost, false)

		next(ctx)

		if modeFormPost, ok := ctx.UserValueBytes(UserValueKeyFormPost).(bool); ok && modeFormPost {
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
