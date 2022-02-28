package validator

import (
	"fmt"
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

func TestShouldSetDefaultNTPAddress(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.Address, config.NTP.Address)
}

func TestShouldSetDefaultNTPVersion(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.Version, config.NTP.Version)
}

func TestShouldRaiseErrorOnInvalidNTPVersion(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	config.NTP.Version = 1

	ValidateNTP(&config, validator)

	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtNTPVersion, config.NTP.Version))
}

func TestShouldSetDefaultNTPMaximumDesync(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.MaximumDesync, config.NTP.MaximumDesync)
}

func TestShouldSetDefaultNTPDisableStartupCheck(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.DisableStartupCheck, config.NTP.DisableStartupCheck)
}
