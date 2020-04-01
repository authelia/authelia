package handlers

import "github.com/authelia/authelia/internal/middlewares"

type ConfigurationBody struct {
	GoogleAnalyticsTrackingID string `json:"ga_tracking_id,omitempty"`
	RememberMeEnabled         bool   `json:"remember_me_enabled"` // whether remember me is enabled or not
}

func ConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := ConfigurationBody{
		GoogleAnalyticsTrackingID: ctx.Configuration.GoogleAnalyticsTrackingID,
		RememberMeEnabled:         ctx.Configuration.Session.RememberMe.Duration != 0,
	}
	ctx.SetJSONBody(body)
}
