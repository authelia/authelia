package validator

import (
	"fmt"
	"regexp"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// ValidateTheme validates and update Theme configuration.
func ValidateTheme(configuration *schema.Configuration, validator *schema.StructValidator) {
	validThemes := regexp.MustCompile("light|dark|grey")
	if !validThemes.MatchString(configuration.Theme) {
		validator.Push(fmt.Errorf("Theme: %s is not valid, valid themes are: \"light\", \"dark\" or \"grey\"", configuration.Theme))
	}
}
