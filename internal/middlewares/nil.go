package middlewares

import (
	"github.com/valyala/fasthttp"
)

// Nil just passes the request to the next middleware.
func Nil(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		next(ctx)
	}
}
