package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateRegulation validates and update regulator configuration.
func ValidateRegulation(configuration *schema.RegulationConfiguration, validator *schema.StructValidator) {
	if configuration.FindTime == 0 {
		configuration.FindTime = schema.DefaultRegulationConfiguration.FindTime // 2 min.
	}

	if configuration.BanTime == 0 {
		configuration.BanTime = schema.DefaultRegulationConfiguration.BanTime // 5 min.
	}

	if configuration.FindTime > configuration.BanTime {
		validator.Push(fmt.Errorf("regulation: invalid find_time '%s' and ban_time '%s': find_time cannot be greater than ban_time", configuration.FindTime.String(), configuration.BanTime.String()))
	}
}
