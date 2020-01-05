package handlers

import (
	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/middlewares"
)

type ExtendedConfigurationBody struct {
	AvailableMethods MethodList `json:"available_methods"`
}

// ExtendedConfigurationGet get the extended configuration accessible to authenticated users.
func ExtendedConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := ExtendedConfigurationBody{}
	body.AvailableMethods = MethodList{authentication.TOTP, authentication.U2F}

	if ctx.Configuration.DuoAPI != nil {
		body.AvailableMethods = append(body.AvailableMethods, authentication.Push)
	}

	ctx.Logger.Debugf("Available methods are %s", body.AvailableMethods)
	ctx.SetJSONBody(body)
}
