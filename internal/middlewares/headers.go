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

		next(ctx)
	}
}
