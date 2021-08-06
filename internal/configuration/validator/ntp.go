package validator

import (
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateNtp validates and update NTP configuration.
func ValidateNtp(configuration *schema.NtpConfiguration, validator *schema.StructValidator) {
	if configuration.Address == "" {
		configuration.Address = schema.DefaultNtpConfiguration.Address // time.cloudflare.com:123
	}

	if configuration.Version == 0 || configuration.Version > 4 {
		configuration.Version = schema.DefaultNtpConfiguration.Version // 4
	}

	if configuration.MaximumDesync == "" {
		configuration.MaximumDesync = schema.DefaultNtpConfiguration.MaximumDesync // 3 sec
	}

	_, err := utils.ParseDurationString(configuration.MaximumDesync)
	if err != nil {
		validator.Push(fmt.Errorf("Error occurred parsing ntp max_desync string: %s", err))
	}
}
