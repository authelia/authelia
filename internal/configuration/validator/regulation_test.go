package validator

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
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
	config.FindTime = time.Minute
	config.BanTime = time.Second * 10

	ValidateRegulation(&config, validator)

	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf("regulation: invalid find_time '%s' and ban_time '%s': find_time cannot be greater than ban_time", config.FindTime.String(), config.BanTime.String()))
}
