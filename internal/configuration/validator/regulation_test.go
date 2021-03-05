package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func newDefaultRegulationConfig() schema.RegulationConfiguration {
	config := schema.RegulationConfiguration{}
	return config
}

func TestShouldSetDefaultRegulationBanTime(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultRegulationConfiguration.BanTime, config.BanTime)
}

func TestShouldSetDefaultRegulationFindTime(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultRegulationConfiguration.FindTime, config.FindTime)
}

func TestShouldRaiseErrorWhenFindTimeLessThanBanTime(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()
	config.FindTime = "1m"
	config.BanTime = "10s"

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "find_time cannot be greater than ban_time")
}

func TestShouldRaiseErrorOnBadDurationStrings(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()
	config.FindTime = "a year"
	config.BanTime = "forever"

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 2)
	assert.EqualError(t, validator.Errors()[0], "Error occurred parsing regulation find_time string: Could not convert the input string of a year into a duration")
	assert.EqualError(t, validator.Errors()[1], "Error occurred parsing regulation ban_time string: Could not convert the input string of forever into a duration")
}
