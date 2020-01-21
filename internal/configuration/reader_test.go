package configuration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldParseConfigFile(t *testing.T) {
	err := os.Setenv("AUTHELIA_JWT_SECRET", "secret_from_env")
	require.NoError(t, err)

	err = os.Setenv("AUTHELIA_DUO_API_SECRET_KEY", "duo_secret_from_env")
	require.NoError(t, err)

	config, errors := Read("./test_resources/config.yml")

	require.Len(t, errors, 0)

	assert.Equal(t, 9091, config.Port)
	assert.Equal(t, "debug", config.LogsLevel)
	assert.Equal(t, "https://home.example.com:8080/", config.DefaultRedirectionURL)
	assert.Equal(t, "authelia.com", config.TOTP.Issuer)
	assert.Equal(t, "secret_from_env", config.JWTSecret)

	assert.Equal(t, "api-123456789.example.com", config.DuoAPI.Hostname)
	assert.Equal(t, "ABCDEF", config.DuoAPI.IntegrationKey)
	assert.Equal(t, "duo_secret_from_env", config.DuoAPI.SecretKey)

	assert.Equal(t, "deny", config.AccessControl.DefaultPolicy)
	assert.Len(t, config.AccessControl.Rules, 11)
}
