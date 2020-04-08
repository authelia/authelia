package configuration

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestShouldParseConfigFile(t *testing.T) {
	require.NoError(t, os.Setenv("AUTHELIA_JWT_SECRET", "secret_from_env"))
	require.NoError(t, os.Setenv("AUTHELIA_DUO_API_SECRET_KEY", "duo_secret_from_env"))
	require.NoError(t, os.Setenv("AUTHELIA_SESSION_SECRET", "session_secret_from_env"))
	require.NoError(t, os.Setenv("AUTHELIA_AUTHENTICATION_BACKEND_LDAP_PASSWORD", "ldap_secret_from_env"))
	require.NoError(t, os.Setenv("AUTHELIA_NOTIFIER_SMTP_PASSWORD", "smtp_secret_from_env"))
	require.NoError(t, os.Setenv("AUTHELIA_SESSION_REDIS_PASSWORD", "redis_secret_from_env"))
	require.NoError(t, os.Setenv("AUTHELIA_STORAGE_MYSQL_PASSWORD", "mysql_secret_from_env"))
	require.NoError(t, os.Setenv("AUTHELIA_STORAGE_POSTGRES_PASSWORD", "postgres_secret_from_env"))

	config, errors := Read("./test_resources/config.yml")

	require.Len(t, errors, 0)

	assert.Equal(t, 9091, config.Port)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, "https://home.example.com:8080/", config.DefaultRedirectionURL)
	assert.Equal(t, "authelia.com", config.TOTP.Issuer)
	assert.Equal(t, "secret_from_env", config.JWTSecret)

	assert.Equal(t, "api-123456789.example.com", config.DuoAPI.Hostname)
	assert.Equal(t, "ABCDEF", config.DuoAPI.IntegrationKey)
	assert.Equal(t, "duo_secret_from_env", config.DuoAPI.SecretKey)

	assert.Equal(t, "session_secret_from_env", config.Session.Secret)
	assert.Equal(t, "ldap_secret_from_env", config.AuthenticationBackend.Ldap.Password)
	assert.Equal(t, "smtp_secret_from_env", config.Notifier.SMTP.Password)
	assert.Equal(t, "redis_secret_from_env", config.Session.Redis.Password)
	assert.Equal(t, "mysql_secret_from_env", config.Storage.MySQL.Password)
	assert.Equal(t, "postgres_secret_from_env", config.Storage.PostgreSQL.Password)

	assert.Equal(t, "deny", config.AccessControl.DefaultPolicy)
	assert.Len(t, config.AccessControl.Rules, 11)
}
