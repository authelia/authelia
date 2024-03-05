package validator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDefaultRegulationConfig() schema.Configuration {
	config := schema.Configuration{
		Regulation: schema.Regulation{},
	}

	return config
}

func TestShouldSetDefaultRegulationTimeDurationsWhenUnset(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultRegulationConfiguration.BanTime, config.Regulation.BanTime)
	assert.Equal(t, schema.DefaultRegulationConfiguration.FindTime, config.Regulation.FindTime)
}

func TestShouldSetDefaultRegulationTimeDurationsWhenNegative(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()

	config.Regulation.BanTime = -1
	config.Regulation.FindTime = -1

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultRegulationConfiguration.FindTime, config.Regulation.FindTime)
}

func TestShouldRaiseErrorWhenFindTimeLessThanBanTime(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultRegulationConfig()
	config.Regulation.FindTime = time.Minute
	config.Regulation.BanTime = time.Second * 10

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "regulation: option 'find_time' must be less than or equal to option 'ban_time'")
}
