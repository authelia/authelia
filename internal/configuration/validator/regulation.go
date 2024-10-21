package validator

import (
	"errors"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateRegulation validates and update regulator configuration.
func ValidateRegulation(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Regulation.Mode == "" {
		config.Regulation.Mode = schema.DefaultRegulationConfiguration.Mode
	}

	if config.Regulation.FindTime <= 0 {
		config.Regulation.FindTime = schema.DefaultRegulationConfiguration.FindTime // 2 min.
	}

	if config.Regulation.BanTime <= 0 {
		config.Regulation.BanTime = schema.DefaultRegulationConfiguration.BanTime // 5 min.
	}

	if config.Regulation.FindTime > config.Regulation.BanTime {
		validator.Push(errors.New(errFmtRegulationFindTimeGreaterThanBanTime))
	}
}
