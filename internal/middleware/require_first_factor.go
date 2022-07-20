package middleware

import (
	"github.com/authelia/authelia/v4/internal/authentication"
)

// Require1FA check if user has enough permissions to execute the next handler.
func Require1FA(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		if ctx.GetSession().AuthenticationLevel < authentication.OneFactor {
			ctx.ReplyForbidden()
			return
		}

		next(ctx)
	}
}
