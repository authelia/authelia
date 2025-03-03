package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// ConfigurationGET get the configuration accessible to authenticated users.
func ConfigurationGET(ctx *middlewares.AutheliaCtx) {
	body := configurationBody{
		AvailableMethods:       make(MethodList, 0, 3),
		PasswordChangeDisabled: false,
		PasswordResetDisabled:  false,
	}

	if ctx.Providers.Authorizer.IsSecondFactorEnabled() {
		body.AvailableMethods = ctx.AvailableSecondFactorMethods()
	}

	body.PasswordChangeDisabled = ctx.Configuration.AuthenticationBackend.PasswordChange.Disable
	body.PasswordResetDisabled = ctx.Configuration.AuthenticationBackend.PasswordReset.Disable

	ctx.Logger.WithFields(
		map[string]any{
			"available_methods":        body.AvailableMethods,
			"password_change_disabled": body.PasswordChangeDisabled,
			"password_reset_disabled":  body.PasswordResetDisabled,
		}).Trace("Authelia configuration requested")

	if err := ctx.SetJSONBody(body); err != nil {
		ctx.Logger.Errorf("Unable to set configuration response in body: %s", err)
	}
}
