package handlers

import (
	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// ConfigurationGet get the configuration accessible to authenticated users.
func ConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := configurationBody{}
	body.AvailableMethods = MethodList{authentication.TOTP}

	if ctx.Configuration.DuoAPI != nil {
		body.AvailableMethods = append(body.AvailableMethods, authentication.Push)
	}

	if !ctx.Configuration.Webauthn.Disable {
		body.AvailableMethods = append(body.AvailableMethods, authentication.Webauthn)
	}

	body.SecondFactorEnabled = ctx.Providers.Authorizer.IsSecondFactorEnabled()

	ctx.Logger.Tracef("Second factor enabled: %v", body.SecondFactorEnabled)
	ctx.Logger.Tracef("Available methods are %s", body.AvailableMethods)

	err := ctx.SetJSONBody(body)
	if err != nil {
		ctx.Logger.Errorf("Unable to set configuration response in body: %s", err)
	}
}
