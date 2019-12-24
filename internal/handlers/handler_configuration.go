package handlers

import "github.com/authelia/authelia/internal/middlewares"

type ConfigurationBody struct {
	GoogleAnalyticsTrackingID string `json:"ga_tracking_id,omitempty"`
}

func ConfigurationGet(ctx *middlewares.AutheliaCtx) {
	body := ConfigurationBody{
		GoogleAnalyticsTrackingID: ctx.Configuration.GoogleAnalyticsTrackingID,
	}
	ctx.SetJSONBody(body)
}
