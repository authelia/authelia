package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateNTP validates and update NTP configuration.
func ValidateNTP(config *schema.Configuration, validator *schema.StructValidator) {
	if config.NTP.Address == nil {
		config.NTP.Address = schema.DefaultNTPConfiguration.Address
	}

	if !config.NTP.Address.IsUDP() {
		validator.Push(fmt.Errorf(errFmtNTPAddressScheme, config.NTP.Address.String(), fmt.Errorf("scheme must be one of 'udp', 'udp4', or 'udp6' but is configured as '%s'", config.NTP.Address.Scheme())))
	}

	if config.NTP.Version == 0 {
		config.NTP.Version = schema.DefaultNTPConfiguration.Version
	} else if config.NTP.Version < 3 || config.NTP.Version > 4 {
		validator.Push(fmt.Errorf(errFmtNTPVersion, config.NTP.Version))
	}

	if config.NTP.MaximumDesync <= 0 {
		config.NTP.MaximumDesync = schema.DefaultNTPConfiguration.MaximumDesync
	}
}
