package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
)

type locales struct {
	Supported []string `json:"supported"`
}

// ConfigurationLocalesGET get the configuration accessible to authenticated users.
func ConfigurationLocalesGET(ctx *middlewares.AutheliaCtx) {
	body := locales{
		Supported: []string{"en", "es", "de"},
	}

	if err := ctx.SetJSONBody(body); err != nil {
		ctx.Logger.Errorf("Unable to set locales configuration response in body: %s", err)
	}
}
