package handlers

import "github.com/authelia/authelia/internal/middlewares"

type ConfigurationBody struct {
	GoogleAnalyticsTrackingID string `json:"ga_tracking_id,omitempty"`
	RememberMe                bool   `json:"remember_me"` // whether remember me is enabled or not
	ResetPassword             bool   `json:"reset_password"`
}

func ConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := ConfigurationBody{
		GoogleAnalyticsTrackingID: ctx.Configuration.GoogleAnalyticsTrackingID,
		RememberMe:                ctx.Providers.SessionProvider.RememberMe != 0,
		ResetPassword:             !ctx.Configuration.AuthenticationBackend.DisableResetPassword,
	}
	ctx.SetJSONBody(body)
}
