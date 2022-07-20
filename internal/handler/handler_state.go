package handler

import (
	"github.com/authelia/authelia/v4/internal/middleware"
)

// StateGET is the handler serving the user state.
func StateGET(ctx *middleware.AutheliaCtx) {
	userSession := ctx.GetSession()
	stateResponse := StateResponse{
		Username:              userSession.Username,
		AuthenticationLevel:   userSession.AuthenticationLevel,
		DefaultRedirectionURL: ctx.Configuration.DefaultRedirectionURL,
	}

	err := ctx.SetJSONBody(stateResponse)
	if err != nil {
		ctx.Logger.Errorf("Unable to set state response in body: %s", err)
	}
}
