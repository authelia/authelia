package server

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

func handleOPTIONS(ctx *middlewares.AutheliaCtx) {
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}
