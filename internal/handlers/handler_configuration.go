package handlers

import "github.com/authelia/authelia/internal/middlewares"

// ConfigurationBody configuration parameters exposed to the frontend.
type ConfigurationBody struct {
	RememberMe    bool `json:"remember_me"` // whether remember me is enabled or not
	ResetPassword bool `json:"reset_password"`
}

// ConfigurationGet fetches configuration parameters for frontend mutation.
func ConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := ConfigurationBody{
		RememberMe:    ctx.Providers.SessionProvider.RememberMe != 0,
		ResetPassword: !ctx.Configuration.AuthenticationBackend.DisableResetPassword,
	}
	ctx.SetJSONBody(body) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
}
