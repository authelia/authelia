package middlewares

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/valyala/fasthttp"
)

// StripPath strips the first level of a path.
func StripPath(path string) (middleware Middleware) {
	return func(next fasthttp.RequestHandler) fasthttp.RequestHandler {
		if len(path) == 0 || (len(path) == 1 && path[0] == '/') {
			return next
		}

		if path[0] != '/' {
			path = "/" + path
		}

		pattern := regexp.MustCompile(fmt.Sprintf("^%s(/.*?)?$", path))

		return func(ctx *fasthttp.RequestCtx) {
			uri := string(ctx.RequestURI())

			if pattern.MatchString(uri) {
				ctx.SetUserValue(UserValueKeyBaseURL, path)
				ctx.SetUserValue(UserValueKeyRawURI, uri)

				ctx.Request.SetRequestURI(strings.TrimPrefix(uri, path))
			}

			next(ctx)
		}
	}
}
