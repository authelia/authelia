package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// HealthGet can be used by health checks.
func HealthGet(ctx *middlewares.AutheliaCtx) {
	ctx.ReplyOK()
}
