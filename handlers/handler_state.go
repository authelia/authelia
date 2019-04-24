package handlers

import (
	"github.com/clems4ever/authelia/middlewares"
)

// StateGet is the handler serving the user state.
func StateGet(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()
	stateResponse := StateResponse{
		Username:              userSession.Username,
		AuthenticationLevel:   userSession.AuthenticationLevel,
		DefaultRedirectionURL: ctx.Configuration.DefaultRedirectionURL,
	}
	ctx.SetJSONBody(stateResponse)
}
