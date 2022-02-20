package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateNTP validates and update NTP configuration.
func ValidateNTP(configuration *schema.NTPConfiguration, validator *schema.StructValidator) {
	if configuration.Address == "" {
		configuration.Address = schema.DefaultNTPConfiguration.Address
	}

	if configuration.Version == 0 {
		configuration.Version = schema.DefaultNTPConfiguration.Version
	} else if configuration.Version < 3 || configuration.Version > 4 {
		validator.Push(fmt.Errorf(errFmtNTPVersion, configuration.Version))
	}

	if configuration.MaximumDesync == 0 {
		configuration.MaximumDesync = schema.DefaultNTPConfiguration.MaximumDesync
	}
}
