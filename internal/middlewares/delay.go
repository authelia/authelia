package middlewares

import (
	"time"

	"github.com/valyala/fasthttp"
)

// ArbitraryDelay returns a middleware which ensures the request takes at least the given duration.
func ArbitraryDelay(delay time.Duration) Basic {
	return func(next fasthttp.RequestHandler) (handler fasthttp.RequestHandler) {
		return func(ctx *fasthttp.RequestCtx) {
			started := time.Now()

			next(ctx)

			delta := delay - time.Since(started)

			if delta < 0 {
				return
			}

			time.Sleep(delta)
		}
	}
}
