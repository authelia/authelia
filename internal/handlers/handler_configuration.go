package handlers

import (
	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// ConfigurationGet get the configuration accessible to authenticated users.
func ConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := configurationBody{
		AvailableMethods: MethodList{},
	}

	if ctx.Providers.Authorizer.IsSecondFactorEnabled() {
		if ctx.Configuration.TOTP == nil || !ctx.Configuration.TOTP.Disable {
			body.AvailableMethods = append(body.AvailableMethods, authentication.TOTP)
		}

		if !ctx.Configuration.Webauthn.Disable {
			body.AvailableMethods = append(body.AvailableMethods, authentication.Webauthn)
		}

		if ctx.Configuration.DuoAPI != nil {
			body.AvailableMethods = append(body.AvailableMethods, authentication.Push)
		}
	}

	ctx.Logger.Tracef("Available methods are %s", body.AvailableMethods)

	err := ctx.SetJSONBody(body)
	if err != nil {
		ctx.Logger.Errorf("Unable to set configuration response in body: %s", err)
	}
}
