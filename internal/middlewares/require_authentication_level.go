package middlewares

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
)

// Require1FA requires the user to have authenticated with at least one-factor authentication (i.e. password).
func Require1FA(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		if session, err := ctx.GetSession(); err != nil || session.AuthenticationLevel < authentication.OneFactor {
			ctx.ReplyForbidden()
			return
		}

		next(ctx)
	}
}

// Require2FA requires the user to have authenticated with two-factor authentication.
func Require2FA(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		if session, err := ctx.GetSession(); err != nil || session.AuthenticationLevel < authentication.TwoFactor {
			ctx.ReplyForbidden()
			return
		}

		next(ctx)
	}
}

// Require2FAWithAPIResponse requires the user to have authenticated with two-factor authentication.
func Require2FAWithAPIResponse(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		if session, err := ctx.GetSession(); err != nil || session.AuthenticationLevel < authentication.TwoFactor {
			ctx.SetAuthenticationErrorJSON(fasthttp.StatusForbidden, "Authentication Required.", true, false)
			return
		}

		next(ctx)
	}
}
