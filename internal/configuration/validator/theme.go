package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

var rePortalTemplateName = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-_]*[a-z0-9])?$`)

// ValidateTheme validates and update Theme configuration.
func ValidateTheme(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Theme == "" {
		config.Theme = "light"
	}

	if !utils.IsStringInSlice(config.Theme, validThemeNames) {
		validator.Push(fmt.Errorf(errFmtThemeName, utils.StringJoinOr(validThemeNames), config.Theme))
	}

	if config.PortalTemplate == "" {
		return
	}

	config.PortalTemplate = strings.ToLower(config.PortalTemplate)

	if config.PortalTemplate != schema.PortalTemplateNone && !rePortalTemplateName.MatchString(config.PortalTemplate) {
		validator.Push(fmt.Errorf(errFmtPortalTemplateName, config.PortalTemplate))
	}
}
