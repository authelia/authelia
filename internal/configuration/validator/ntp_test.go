package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDefaultNTPConfig() schema.Configuration {
	return schema.Configuration{
		NTP: &schema.NTPConfiguration{},
	}
}

func TestShouldSetDefaultNtpAddress(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.Address, config.NTP.Address)
}

func TestShouldSetDefaultNtpVersion(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.Version, config.NTP.Version)
}

func TestShouldSetDefaultNtpMaximumDesync(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.MaximumDesync, config.NTP.MaximumDesync)
}

func TestShouldSetDefaultNtpDisableStartupCheck(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.DisableStartupCheck, config.NTP.DisableStartupCheck)
}

func TestShouldRaiseErrorOnMaximumDesyncString(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()
	config.NTP.MaximumDesync = "a second"

	ValidateNTP(&config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "ntp: option 'max_desync' can't be parsed: could not convert the input string of a second into a duration")
}

func TestShouldRaiseErrorOnInvalidNTPVersion(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()
	config.NTP.Version = 1

	ValidateNTP(&config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "ntp: option 'version' must be either 3 or 4 but it is configured as '1'")
}
