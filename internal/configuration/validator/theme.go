package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateTheme validates and update Theme configuration.
func ValidateTheme(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Theme == "" {
		config.Theme = "light"
	}

	if !utils.IsStringInSlice(config.Theme, validThemeNames) {
		validator.Push(fmt.Errorf(errFmtThemeName, utils.StringJoinOr(validThemeNames), config.Theme))
	}
}
