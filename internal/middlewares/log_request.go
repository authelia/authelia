package middlewares

import (
	"github.com/valyala/fasthttp"
)

// LogRequest provides trace logging for all requests.
func LogRequest(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		log := NewRequestLogger(ctx)

		log.Trace("Request hit")

		next(ctx)

		log.Tracef("Replied (status=%d)", ctx.Response.StatusCode())
	}
}
