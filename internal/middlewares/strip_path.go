package middlewares

import (
	"bytes"

	"github.com/valyala/fasthttp"
)

// StripPathMiddleware strips the first level of a path.
func StripPathMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		uri := ctx.Request.RequestURI()
		n := bytes.IndexByte(uri[1:], '/')

		if n >= 0 {
			uri = uri[n+1:]
			ctx.Request.SetRequestURI(string(uri))
		}

		next(ctx)
	}
}
