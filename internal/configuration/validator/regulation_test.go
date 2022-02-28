package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDefaultRegulationConfig() schema.Configuration {
	config := schema.Configuration{
		Regulation: &schema.RegulationConfiguration{},
	}

	return config
}

func TestShouldSetDefaultRegulationBanTime(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultRegulationConfiguration.BanTime, config.Regulation.BanTime)
}

func TestShouldSetDefaultRegulationFindTime(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultRegulationConfiguration.FindTime, config.Regulation.FindTime)
}

func TestShouldRaiseErrorWhenFindTimeLessThanBanTime(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()
	config.Regulation.FindTime = "1m"
	config.Regulation.BanTime = "10s"

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "regulation: option 'find_time' must be less than or equal to option 'ban_time'")
}

func TestShouldRaiseErrorOnBadDurationStrings(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()
	config.Regulation.FindTime = "a year"
	config.Regulation.BanTime = "forever"

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 2)
	assert.EqualError(t, validator.Errors()[0], "regulation: option 'find_time' could not be parsed: could not parse 'a year' as a duration")
	assert.EqualError(t, validator.Errors()[1], "regulation: option 'ban_time' could not be parsed: could not parse 'forever' as a duration")
}
