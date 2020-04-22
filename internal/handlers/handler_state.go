package handlers

import (
	"github.com/authelia/authelia/internal/middlewares"
)

// StateGet is the handler serving the user state.
func StateGet(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()
	stateResponse := StateResponse{
		Username:              userSession.Username,
		AuthenticationLevel:   userSession.AuthenticationLevel,
		DefaultRedirectionURL: ctx.Configuration.DefaultRedirectionURL,
	}
	ctx.SetJSONBody(stateResponse) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
}
