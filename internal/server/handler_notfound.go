package server

import (
	"strings"

	"github.com/valyala/fasthttp"
)

func handleNotFound(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		path := strings.ToLower(string(ctx.Path()))

		for i := 0; i < len(httpServerDirs); i++ {
			if path == httpServerDirs[i].name || strings.HasPrefix(path, httpServerDirs[i].prefix) {
				ctx.SetStatusCode(fasthttp.StatusNotFound)
				ctx.SetBodyString(fasthttp.StatusMessage(fasthttp.StatusNotFound))

				return
			}
		}

		next(ctx)
	}
}
