package middlewares

import "github.com/valyala/fasthttp"

// ResponseHeadersMiddleware adds/removes/transforms response headers.
func ResponseHeadersMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		ctx.Response.Header.Add("Permissions-Policy", "interest-cohort=()")
		next(ctx)
	}
}
