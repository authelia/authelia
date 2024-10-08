package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// ConfigurationGET get the configuration accessible to authenticated users.
func ConfigurationGET(ctx *middlewares.AutheliaCtx) {
	body := configurationBody{
		AvailableMethods:       make(MethodList, 0, 3),
		PasswordChangeDisabled: false,
	}

	if ctx.Providers.Authorizer.IsSecondFactorEnabled() {
		body.AvailableMethods = ctx.AvailableSecondFactorMethods()
	}

	body.PasswordChangeDisabled = ctx.Configuration.AuthenticationBackend.PasswordChange.Disable

	var passwordChangeString string

	if body.PasswordChangeDisabled {
		passwordChangeString = "disabled."
	} else {
		passwordChangeString = "enabled."
	}

	ctx.Logger.Tracef("Available methods are %s", body.AvailableMethods)
	ctx.Logger.Tracef("Password change is %s", passwordChangeString)

	if err := ctx.SetJSONBody(body); err != nil {
		ctx.Logger.Errorf("Unable to set configuration response in body: %s", err)
	}
}
