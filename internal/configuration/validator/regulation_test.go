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

func TestShouldValidateRegulationModes(t *testing.T) {
	t.Run("ShouldAcceptValidModeIP", func(t *testing.T) {
		validator := schema.NewStructValidator()
		config := newDefaultRegulationConfig()
		config.Regulation.Modes = []string{"ip"}

		ValidateRegulation(&config, validator)

		assert.Len(t, validator.Errors(), 0)
		assert.Equal(t, []string{"ip"}, config.Regulation.Modes)
	})

	t.Run("ShouldAcceptValidModeUser", func(t *testing.T) {
		validator := schema.NewStructValidator()
		config := newDefaultRegulationConfig()
		config.Regulation.Modes = []string{"user"}

		ValidateRegulation(&config, validator)

		assert.Len(t, validator.Errors(), 0)
		assert.Equal(t, []string{"user"}, config.Regulation.Modes)
	})

	t.Run("ShouldAcceptBothValidModes", func(t *testing.T) {
		validator := schema.NewStructValidator()
		config := newDefaultRegulationConfig()
		config.Regulation.Modes = []string{"ip", "user"}

		ValidateRegulation(&config, validator)

		assert.Len(t, validator.Errors(), 0)
		assert.Equal(t, []string{"ip", "user"}, config.Regulation.Modes)
	})

	t.Run("ShouldRejectInvalidMode", func(t *testing.T) {
		validator := schema.NewStructValidator()
		config := newDefaultRegulationConfig()
		config.Regulation.Modes = []string{"invalid"}

		ValidateRegulation(&config, validator)

		assert.Len(t, validator.Errors(), 1)
		assert.EqualError(t, validator.Errors()[0], "regulation: option 'modes' must only contain the values 'user' and 'ip' but contains the value 'invalid'")
	})

	t.Run("ShouldRejectMixedValidAndInvalidModes", func(t *testing.T) {
		validator := schema.NewStructValidator()
		config := newDefaultRegulationConfig()
		config.Regulation.Modes = []string{"ip", "bad", "user", "worse"}

		ValidateRegulation(&config, validator)

		assert.Len(t, validator.Errors(), 2)
		assert.EqualError(t, validator.Errors()[0], "regulation: option 'modes' must only contain the values 'user' and 'ip' but contains the value 'bad'")
		assert.EqualError(t, validator.Errors()[1], "regulation: option 'modes' must only contain the values 'user' and 'ip' but contains the value 'worse'")
	})

	t.Run("ShouldSetDefaultModesWhenEmpty", func(t *testing.T) {
		validator := schema.NewStructValidator()
		config := newDefaultRegulationConfig()
		config.Regulation.Modes = []string{}

		ValidateRegulation(&config, validator)

		assert.Len(t, validator.Errors(), 0)
		assert.Equal(t, schema.DefaultRegulationConfiguration.Modes, config.Regulation.Modes)
	})
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
