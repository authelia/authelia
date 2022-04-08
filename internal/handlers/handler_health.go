package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// HealthGET can be used by health checks.
func HealthGET(ctx *middlewares.AutheliaCtx) {
	ctx.ReplyOK()
}
