package validator

import (
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateNTP validates and update NTP configuration.
func ValidateNTP(configuration *schema.NTPConfiguration, validator *schema.StructValidator) {
	if configuration.Address == "" {
		configuration.Address = schema.DefaultNTPConfiguration.Address // time.cloudflare.com:123
	}

	if configuration.Version == 0 || configuration.Version > 4 {
		configuration.Version = schema.DefaultNTPConfiguration.Version // 4
	}

	if configuration.MaximumDesync == "" {
		configuration.MaximumDesync = schema.DefaultNTPConfiguration.MaximumDesync // 3 sec
	}

	_, err := utils.ParseDurationString(configuration.MaximumDesync)
	if err != nil {
		validator.Push(fmt.Errorf("Error occurred parsing NTP max_desync string: %s", err))
	}
}
