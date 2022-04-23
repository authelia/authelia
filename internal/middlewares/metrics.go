package middlewares

import (
	"strconv"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/metrics"
)

// NewMetricsRequestMiddleware returns a middleware if provided with a metrics.Recorder, otherwise it returns nil.
func NewMetricsRequestMiddleware(metrics metrics.Recorder) (middleware Basic) {
	if metrics == nil {
		return nil
	}

	return func(next fasthttp.RequestHandler) (handler fasthttp.RequestHandler) {
		return func(ctx *fasthttp.RequestCtx) {
			started := time.Now()

			next(ctx)

			statusCode := strconv.Itoa(ctx.Response.StatusCode())
			requestMethod := string(ctx.Method())

			metrics.RecordRequest(statusCode, requestMethod, time.Since(started))
		}
	}
}

// NewMetricsVerifyRequestMiddleware returns a middleware if provided with a metrics.Recorder, otherwise it returns nil.
func NewMetricsVerifyRequestMiddleware(metrics metrics.Recorder) (middleware Basic) {
	if metrics == nil {
		return nil
	}

	return func(next fasthttp.RequestHandler) (handler fasthttp.RequestHandler) {
		return func(ctx *fasthttp.RequestCtx) {
			next(ctx)

			statusCode := strconv.Itoa(ctx.Response.StatusCode())

			metrics.RecordVerifyRequest(statusCode)
		}
	}
}
