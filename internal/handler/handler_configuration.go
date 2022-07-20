package handler

import (
	"github.com/authelia/authelia/v4/internal/middleware"
)

// ConfigurationGET get the configuration accessible to authenticated users.
func ConfigurationGET(ctx *middleware.AutheliaCtx) {
	body := configurationBody{
		AvailableMethods: make(MethodList, 0, 3),
	}

	if ctx.Providers.Authorizer.IsSecondFactorEnabled() {
		body.AvailableMethods = ctx.AvailableSecondFactorMethods()
	}

	ctx.Logger.Tracef("Available methods are %s", body.AvailableMethods)

	if err := ctx.SetJSONBody(body); err != nil {
		ctx.Logger.Errorf("Unable to set configuration response in body: %s", err)
	}
}
