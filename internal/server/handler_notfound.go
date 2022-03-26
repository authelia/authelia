package server

import (
	"strings"

	"github.com/valyala/fasthttp"
)

func handleNotFound(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := string(ctx.Path())

		if strings.EqualFold(path, "/api") || strings.HasPrefix(path, "/api/") {
			ctx.SetStatusCode(fasthttp.StatusNotFound)
			ctx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusNotFound))

			return
		}

		next(ctx)
	}
}
