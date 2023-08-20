package middlewares

import (
	"strconv"
	"strings"
	"time"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/metrics"
)

// NewMetricsRequest returns a middleware if provided with a metrics.Recorder, otherwise it returns nil.
func NewMetricsRequest(metrics metrics.Recorder) (middleware Basic) {
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

// NewMetricsRequestOpenIDConnect returns a middleware if provided with a metrics.Recorder, otherwise it returns nil.
func NewMetricsRequestOpenIDConnect(metrics metrics.Recorder, endpoint string) (middleware Basic) {
	if metrics == nil {
		return nil
	}

	endpoint = strings.ReplaceAll(endpoint, "-", "_")

	return func(next fasthttp.RequestHandler) (handler fasthttp.RequestHandler) {
		return func(ctx *fasthttp.RequestCtx) {
			started := time.Now()

			next(ctx)

			statusCode := strconv.Itoa(ctx.Response.StatusCode())

			metrics.RecordRequestOpenIDConnect(endpoint, statusCode, time.Since(started))
		}
	}
}

// NewMetricsAuthzRequest returns a middleware if provided with a metrics.Recorder, otherwise it returns nil.
func NewMetricsAuthzRequest(metrics metrics.Recorder) (middleware Basic) {
	if metrics == nil {
		return nil
	}

	return func(next fasthttp.RequestHandler) (handler fasthttp.RequestHandler) {
		return func(ctx *fasthttp.RequestCtx) {
			next(ctx)

			statusCode := strconv.Itoa(ctx.Response.StatusCode())

			metrics.RecordAuthz(statusCode)
		}
	}
}
