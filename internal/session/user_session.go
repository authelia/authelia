package session

import (
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/authentication"
)

// NewDefaultUserSession create a default user session.
func NewDefaultUserSession(ctx *fasthttp.RequestCtx) UserSession {
	return UserSession{
		KeepMeLoggedIn:      false,
		AuthenticationLevel: authentication.NotAuthenticated,
		LastActivity:        0,
		IP:                  ctx.RemoteIP().String(),
	}
}
