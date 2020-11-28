package handlers

import (
	"github.com/authelia/authelia/internal/middlewares"
)

// HealthGet can be used by health checks.
func HealthGet(ctx *middlewares.AutheliaCtx) {
	ctx.ReplyOK()
}
