package middlewares

import (
	"github.com/authelia/authelia/v4/internal/authentication"
)

// RequireFirstFactor check if user has enough permissions to execute the next handler.
func RequireFirstFactor(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		if ctx.GetSession().AuthenticationLevel < authentication.OneFactor {
			ctx.ReplyForbidden()
			return
		}

		next(ctx)
	}
}
