package handlers

import "github.com/authelia/authelia/internal/middlewares"

// ConfigurationBody configuration parameters exposed to the frontend.
type ConfigurationBody struct {
	GoogleAnalyticsTrackingID string `json:"ga_tracking_id,omitempty"`
	RememberMe                bool   `json:"remember_me"` // whether remember me is enabled or not
	ResetPassword             bool   `json:"reset_password"`
	Path                      string `json:"path"`
}

// ConfigurationGet fetches configuration parameters for frontend mutation.
func ConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := ConfigurationBody{
		GoogleAnalyticsTrackingID: ctx.Configuration.GoogleAnalyticsTrackingID,
		RememberMe:                ctx.Providers.SessionProvider.RememberMe != 0,
		ResetPassword:             !ctx.Configuration.AuthenticationBackend.DisableResetPassword,
		Path:                      ctx.Configuration.Path,
	}
	ctx.SetJSONBody(body) //nolint:errcheck // TODO: Legacy code, consider refactoring time permitting.
}
