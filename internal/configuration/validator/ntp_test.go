package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDefaultNTPConfig() schema.NTPConfiguration {
	config := schema.NTPConfiguration{}
	return config
}

func TestShouldSetDefaultNtpAddress(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.Address, config.Address)
}

func TestShouldSetDefaultNtpVersion(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.Version, config.Version)
}

func TestShouldSetDefaultNtpMaximumDesync(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.MaximumDesync, config.MaximumDesync)
}

func TestShouldSetDefaultNtpDisableStartupCheck(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.DisableStartupCheck, config.DisableStartupCheck)
}

func TestShouldRaiseErrorOnMaximumDesyncString(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()
	config.MaximumDesync = "a second"

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "ntp: error occurred parsing NTP max_desync string: could not convert the input string of a second into a duration")
}
