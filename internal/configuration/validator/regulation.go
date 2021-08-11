package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateRegulation validates and update regulator configuration.
func ValidateRegulation(configuration *schema.RegulationConfiguration, validator *schema.StructValidator) {
	if configuration.FindTime == "" {
		configuration.FindTime = schema.DefaultRegulationConfiguration.FindTime // 2 min
	}

	if configuration.BanTime == "" {
		configuration.BanTime = schema.DefaultRegulationConfiguration.BanTime // 5 min
	}

	findTime, err := utils.ParseDurationString(configuration.FindTime)
	if err != nil {
		validator.Push(fmt.Errorf("Error occurred parsing regulation find_time string: %s", err))
	}

	banTime, err := utils.ParseDurationString(configuration.BanTime)
	if err != nil {
		validator.Push(fmt.Errorf("Error occurred parsing regulation ban_time string: %s", err))
	}

	if findTime > banTime {
		validator.Push(fmt.Errorf("find_time cannot be greater than ban_time"))
	}
}
