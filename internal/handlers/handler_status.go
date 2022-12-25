package handlers

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// Status handles basic status responses.
func Status(statusCode int) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		middlewares.SetStatusCodeResponse(ctx, statusCode)
	}
}
