package middlewares

import (
	"strings"

	"github.com/valyala/fasthttp"
)

// StripPath strips the first level of a path.
func StripPath(path string) (middleware Middleware) {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		return func(ctx *fasthttp.RequestCtx) {
			uri := string(ctx.RequestURI())

			if strings.HasPrefix(uri, path) {
				ctx.SetUserValue(UserValueKeyBaseURL, path)
				ctx.SetUserValue(UserValueKeyRawURI, uri)

				newURI := strings.TrimPrefix(uri, path)
				ctx.Request.SetRequestURI(newURI)
			}

			next(ctx)
		}
	}
}
