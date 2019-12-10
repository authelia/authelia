package handlers

import (
	"github.com/clems4ever/authelia/internal/authentication"
	"github.com/clems4ever/authelia/internal/middlewares"
)

type ExtendedConfigurationBody struct {
	AvailableMethods MethodList `json:"available_methods"`
}

// ExtendedConfigurationGet get the extended configuration accessbile to authenticated users.
func ExtendedConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := ExtendedConfigurationBody{}
	body.AvailableMethods = MethodList{authentication.TOTP, authentication.U2F}

	if ctx.Configuration.DuoAPI != nil {
		body.AvailableMethods = append(body.AvailableMethods, authentication.Push)
	}

	ctx.Logger.Debugf("Available methods are %s", body.AvailableMethods)
	ctx.SetJSONBody(body)
}
