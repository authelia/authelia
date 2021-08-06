package middlewares

import (
	"strings"

	"github.com/valyala/fasthttp"
)

// StripPathMiddleware strips the first level of a path.
func StripPathMiddleware(path string, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	path = "/" + path + "/"

	return func(ctx *fasthttp.RequestCtx) {
		ctx.Request.SetRequestURI(strings.TrimPrefix(string(ctx.RequestURI()), path))

		next(ctx)
	}
}
