package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func newDefaultNtpConfig() schema.NtpConfiguration {
	config := schema.NtpConfiguration{}
	return config
}

func TestShouldSetDefaultNtpAddress(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNtpConfig()

	ValidateNtp(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNtpConfiguration.Address, config.Address)
}

func TestShouldSetDefaultNtpVersion(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNtpConfig()

	ValidateNtp(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNtpConfiguration.Version, config.Version)
}

func TestShouldSetDefaultNtpMaximumDesync(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNtpConfig()

	ValidateNtp(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNtpConfiguration.MaximumDesync, config.MaximumDesync)
}

func TestShouldSetDefaultNtpDisableStartupCheck(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNtpConfig()

	ValidateNtp(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNtpConfiguration.DisableStartupCheck, config.DisableStartupCheck)
}

func TestShouldRaiseErrorOnMaximumDesyncString(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNtpConfig()
	config.MaximumDesync = "a second"

	ValidateNtp(&config, validator)

	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Error occurred parsing ntp max_desync string: could not convert the input string of a second into a duration")
}
