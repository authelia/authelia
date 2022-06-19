package handlers

import (
	"github.com/valyala/fasthttp"
)

// Status handles basic status responses.
func Status(statusCode int) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		SetStatusCodeResponse(ctx, statusCode)
	}
}
