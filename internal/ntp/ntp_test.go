package ntp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
)

func TestShouldCheckNTP(t *testing.T) {
	config := schema.NTPConfiguration{
		Address:             "time.google.com:123",
		Version:             4,
		MaximumDesync:       "3s",
		DisableStartupCheck: false,
	}
	sv := schema.NewStructValidator()
	validator.ValidateNTP(&config, sv)

	NTP := NewProvider(&config)

	checkfailed, _ := NTP.StartupCheck()
	assert.Equal(t, false, checkfailed)
}
