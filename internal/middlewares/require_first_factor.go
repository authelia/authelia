package middlewares

import (
	"github.com/authelia/authelia/v4/internal/authentication"
)

// Require1FA check if user has enough permissions to execute the next handler.
func Require1FA(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		if s, err := ctx.GetSession(); err != nil || s.AuthenticationLevel < authentication.OneFactor {
			ctx.ReplyForbidden()
			return
		}

		next(ctx)
	}
}

func RequireElevated1FA(next RequestHandler) RequestHandler {
	return func(ctx *AutheliaCtx) {
		s, err := ctx.GetSession()

		if err != nil || s.AuthenticationLevel < authentication.OneFactor || s.Elevations.User == nil {
			ctx.ReplyForbidden()
			return
		}

		invalid := false

		if ctx.GetClock().Now().After(s.Elevations.User.Expires) {
			invalid = true

			ctx.Logger.WithFields(map[string]any{"user": s.Username, "expired": s.Elevations.User.Expires.Unix()}).Info("The user session elevation was expired. It will be destroyed and the users access will be forbidden.")
		}

		if !ctx.RemoteIP().Equal(s.Elevations.User.RemoteIP) {
			invalid = true

			ctx.Logger.WithFields(map[string]any{"user": s.Username, "expected_ip": s.Elevations.User.RemoteIP.String()}).Warn("The user session elevation did not have a matching IP. It will be destroyed and the users access will be forbidden.")
		}

		if invalid {
			s.Elevations.User = nil

			if err = ctx.SaveSession(s); err != nil {
				ctx.Logger.WithError(err).Error("Error occurred trying to save the user session after a policy constraint violation occurred.")
			}

			ctx.ReplyForbidden()
			return
		}

		next(ctx)
	}
}
