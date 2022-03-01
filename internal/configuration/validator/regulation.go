package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateRegulation validates and update regulator configuration.
func ValidateRegulation(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Regulation == nil {
		config.Regulation = &schema.DefaultRegulationConfiguration

		return
	}

	if config.Regulation.FindTime == "" {
		config.Regulation.FindTime = schema.DefaultRegulationConfiguration.FindTime // 2 min.
	}

	if config.Regulation.BanTime == "" {
		config.Regulation.BanTime = schema.DefaultRegulationConfiguration.BanTime // 5 min.
	}

	findTime, err := utils.ParseDurationString(config.Regulation.FindTime)
	if err != nil {
		validator.Push(fmt.Errorf(errFmtRegulationParseDuration, "find_time", err))
	}

	banTime, err := utils.ParseDurationString(config.Regulation.BanTime)
	if err != nil {
		validator.Push(fmt.Errorf(errFmtRegulationParseDuration, "ban_time", err))
	}

	if findTime > banTime {
		validator.Push(fmt.Errorf(errFmtRegulationFindTimeGreaterThanBanTime))
	}
}
