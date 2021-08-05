package ntp

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/configuration/validator"
)

func TestShouldCheckNTP(t *testing.T) {
	config := schema.NtpConfiguration{
		Address:             "time.google.com:123",
		Version:             4,
		MaximumDesync:       "3s",
		DisableStartupCheck: false,
	}
	sv := schema.NewStructValidator()
	validator.ValidateNtp(&config, sv)

	Ntp := NewProvider(&config)

	checkfailed, _ := Ntp.StartupCheck()
	assert.Equal(t, false, checkfailed)
}
