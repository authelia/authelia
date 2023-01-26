package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDefaultNTPConfig() schema.Configuration {
	return schema.Configuration{
		NTP: schema.NTP{},
	}
}

func TestShouldSetDefaultNTPValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.Address, config.NTP.Address)
	assert.Equal(t, schema.DefaultNTPConfiguration.Version, config.NTP.Version)
	assert.Equal(t, schema.DefaultNTPConfiguration.MaximumDesync, config.NTP.MaximumDesync)
	assert.Equal(t, schema.DefaultNTPConfiguration.DisableStartupCheck, config.NTP.DisableStartupCheck)
}

func TestShouldSetDefaultNtpVersion(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()

	config.NTP.MaximumDesync = -1

	ValidateNTP(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultNTPConfiguration.MaximumDesync, config.NTP.MaximumDesync)
}

func TestShouldRaiseErrorOnInvalidNTPVersion(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()
	config.NTP.Version = 1

	ValidateNTP(&config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "ntp: option 'version' must be either 3 or 4 but it's configured as '1'")
}

func TestShouldRaiseErrorOnInvalidNTPScheme(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultNTPConfig()
	config.NTP.Address = &schema.AddressUDP{Address: MustParseAddress("tcp://abc:123")}

	ValidateNTP(&config, validator)

	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "ntp: option 'address' with value 'tcp://abc:123' is invalid: scheme must be one of 'udp', 'udp4', or 'udp6' but is configured as 'tcp'")
}
