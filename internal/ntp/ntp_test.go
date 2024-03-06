package ntp

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
)

func TestShouldCheckNTPV4(t *testing.T) {
	config := &schema.Configuration{
		NTP: schema.NTP{
			Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, "time.cloudflare.com", 123)},
			Version:       4,
			MaximumDesync: time.Second * 3,
		},
	}

	svalidator := schema.NewStructValidator()
	validator.ValidateNTP(config, svalidator)

	ntp := NewProvider(&config.NTP)

	assert.NoError(t, ntp.StartupCheck())
}

func TestShouldCheckNTPV3(t *testing.T) {
	config := &schema.Configuration{
		NTP: schema.NTP{
			Address:       &schema.AddressUDP{Address: schema.NewAddressFromNetworkValues(schema.AddressSchemeUDP, "time.cloudflare.com", 123)},
			Version:       3,
			MaximumDesync: time.Second * 3,
		},
	}

	svalidator := schema.NewStructValidator()
	validator.ValidateNTP(config, svalidator)

	ntp := NewProvider(&config.NTP)

	assert.NoError(t, ntp.StartupCheck())
}
