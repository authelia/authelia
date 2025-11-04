package handlers

import (
	"strings"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

type portalTemplateConfigurationBody struct {
	Template               string `json:"template"`
	EnableTemplateSwitcher bool   `json:"enableTemplateSwitcher"`
}

// PortalTemplateGET exposes the configured portal template settings.
func PortalTemplateGET(ctx *middlewares.AutheliaCtx) {
	template := ctx.Configuration.PortalTemplate
	if template == "" {
		template = "none"
	}

	response := portalTemplateConfigurationBody{
		Template:               strings.ToLower(template),
		EnableTemplateSwitcher: ctx.Configuration.PortalTemplateSwitcher,
	}

	if err := ctx.SetJSONBody(response); err != nil {
		ctx.Logger.Errorf("Unable to set portal template response in body: %s", err)
	}
}
