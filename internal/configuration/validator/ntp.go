package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateNTP validates and update NTP configuration.
func ValidateNTP(configuration *schema.NTPConfiguration, validator *schema.StructValidator) {
	if configuration.Address == "" {
		configuration.Address = schema.DefaultNTPConfiguration.Address // time.cloudflare.com:123
	}

	if configuration.Version == 0 {
		configuration.Version = schema.DefaultNTPConfiguration.Version // 4
	} else if configuration.Version < 3 || configuration.Version > 4 {
		validator.Push(fmt.Errorf("ntp: version must be either 3 or 4"))
	}

	if configuration.MaximumDesync == "" {
		configuration.MaximumDesync = schema.DefaultNTPConfiguration.MaximumDesync // 3 sec
	}

	_, err := utils.ParseDurationString(configuration.MaximumDesync)
	if err != nil {
		validator.Push(fmt.Errorf("ntp: error occurred parsing NTP max_desync string: %s", err))
	}
}
