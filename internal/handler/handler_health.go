package handler

import (
	"github.com/authelia/authelia/v4/internal/middleware"
)

// HealthGET can be used by health checks.
func HealthGET(ctx *middleware.AutheliaCtx) {
	ctx.ReplyOK()
}
