package validator

import (
	"errors"
	"fmt"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// ValidateSession validates and update session configuration.
func ValidateRegulation(configuration *schema.RegulationConfiguration, validator *schema.StructValidator) {
	if configuration.FindTime == "" {
		configuration.FindTime = schema.DefaultRegulationConfiguration.FindTime // 2 min
	}
	if configuration.BanTime == "" {
		configuration.BanTime = schema.DefaultRegulationConfiguration.BanTime // 5 min
	}
	findTime, err := utils.ParseDurationString(configuration.FindTime)
	if err != nil {
		validator.Push(errors.New(fmt.Sprintf("Error occurred parsing regulation find_time string: %s", err)))
	}
	banTime, err := utils.ParseDurationString(configuration.BanTime)
	if err != nil {
		validator.Push(errors.New(fmt.Sprintf("Error occurred parsing regulation ban_time string: %s", err)))
	}
	if findTime > banTime {
		validator.Push(errors.New(fmt.Sprintf("find_time cannot be greater than ban_time")))
	}
}
