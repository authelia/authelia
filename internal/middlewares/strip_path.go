package middlewares

import (
	"strings"

	"github.com/valyala/fasthttp"
)

// StripPathMiddleware strips the first level of a path.
func StripPathMiddleware(path string, next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		uri := ctx.RequestURI()

		if strings.HasPrefix(string(uri), path) {
			ctx.SetUserValueBytes(UserValueKeyBaseURL, path)

			newURI := strings.TrimPrefix(string(uri), path)
			ctx.Request.SetRequestURI(newURI)
		}

		next(ctx)
	}
}
