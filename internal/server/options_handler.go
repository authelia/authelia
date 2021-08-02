package server

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/middlewares"
)

func handleOPTIONS(ctx *middlewares.AutheliaCtx) {
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}
