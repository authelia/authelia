package handlers

import (
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// ConfigurationLanguagesGet get the configuration language information.
func ConfigurationLanguagesGet(ctx *middlewares.AutheliaCtx) {
	body := configurationLanguageBody{
		SupportedLanguages: []string{"en", "es"},
	}

	if err := ctx.SetJSONBody(body); err != nil {
		ctx.Logger.Errorf("Unable to set configuration language response in body: %s", err)
	}
}
