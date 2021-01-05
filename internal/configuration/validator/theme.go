package validator

import (
	"fmt"
	"regexp"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// ValidateTheme validates and update Theme configuration.
func ValidateTheme(configuration *schema.ThemeConfiguration, validator *schema.StructValidator) {
	validThemes := regexp.MustCompile("light|dark|grey|custom")
	validHexColor := regexp.MustCompile("#[0-9A-Fa-f]{6}")

	if !validThemes.MatchString(configuration.Name) {
		validator.Push(fmt.Errorf("Theme: %s is not valid, valid themes are: \"light\", \"dark\", \"grey\" or \"custom\"", configuration.Name))
	}

	if configuration.PrimaryColor == "" {
		configuration.PrimaryColor = schema.DefaultThemeConfiguration.PrimaryColor
		if configuration.Name == themeCustom {
			validator.PushWarning(fmt.Errorf("Theme primary color has not been specified, defaulting to: %s", schema.DefaultThemeConfiguration.PrimaryColor))
		}
	}

	if configuration.SecondaryColor == "" {
		configuration.SecondaryColor = schema.DefaultThemeConfiguration.SecondaryColor
		if configuration.Name == themeCustom {
			validator.PushWarning(fmt.Errorf("Theme secondary color has not been specified, defaulting to: %s", schema.DefaultThemeConfiguration.SecondaryColor))
		}
	}

	if configuration.Name == themeCustom {
		if !validHexColor.MatchString(configuration.PrimaryColor) {
			validator.Push(fmt.Errorf("Theme primary color: %s is not valid, valid color values are hex from: \"#000000\" to \"#FFFFFF\"", configuration.PrimaryColor))
		}

		if !validHexColor.MatchString(configuration.SecondaryColor) {
			validator.Push(fmt.Errorf("Theme secondary color: %s is not valid, valid color values are hex from: \"#000000\" to \"#FFFFFF\"", configuration.SecondaryColor))
		}
	}
}
