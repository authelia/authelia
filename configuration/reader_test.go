package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldParseConfigFile(t *testing.T) {
	config, errors := Read("../test-resources/config.yml")

	assert.Len(t, errors, 0)

	assert.Equal(t, 9091, config.Port)
	assert.Equal(t, "debug", config.LogsLevel)
	assert.Equal(t, "https://home.example.com:8080/", config.DefaultRedirectionURL)
	assert.Equal(t, "authelia.com", config.TOTP.Issuer)

	assert.Equal(t, "api-123456789.example.com", config.DuoAPI.Hostname)
	assert.Equal(t, "ABCDEF", config.DuoAPI.IntegrationKey)
	assert.Equal(t, "1234567890abcdefghifjkl", config.DuoAPI.SecretKey)
}
