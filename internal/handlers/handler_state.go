package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// StateGet is the handler serving the user state.
func StateGet(ctx *middlewares.AutheliaCtx) {
	userSession := ctx.GetSession()
	stateResponse := StateResponse{
		Username:              userSession.Username,
		AuthenticationLevel:   userSession.AuthenticationLevel,
		DefaultRedirectionURL: ctx.Configuration.DefaultRedirectionURL,
	}

	err := ctx.SetJSONBody(stateResponse)
	if err != nil {
		ctx.Logger.Errorf("unable to set state response in body: %s", err)
	}
}
