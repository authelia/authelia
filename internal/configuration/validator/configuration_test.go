package validator

import (
	"testing"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDefaultConfig() schema.Configuration {
	config := schema.Configuration{}
	config.Host = "127.0.0.1"
	config.Port = 9090
	config.LogsLevel = "info"
	config.JWTSecret = "a_secret"
	config.AuthenticationBackend.File = new(schema.FileAuthenticationBackendConfiguration)
	config.AuthenticationBackend.File.Path = "/a/path"
	config.Session = schema.SessionConfiguration{
		Domain: "example.com",
		Name:   "authelia_session",
		Secret: "secret",
	}
	config.Storage.Local = &schema.LocalStorageConfiguration{
		Path: "abc",
	}
	config.Notifier = &schema.NotifierConfiguration{
		FileSystem: &schema.FileSystemNotifierConfiguration{
			Filename: "/tmp/file",
		},
	}
	return config
}

func TestShouldNotUpdateConfig(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	Validate(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, 9090, config.Port)
	assert.Equal(t, "info", config.LogsLevel)
}

func TestShouldValidateAndUpdatePort(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Port = 0

	Validate(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, 8080, config.Port)
}

func TestShouldValidateAndUpdateHost(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Host = ""

	Validate(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "0.0.0.0", config.Host)
}

func TestShouldValidateAndUpdateLogsLevel(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.LogsLevel = ""

	Validate(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "info", config.LogsLevel)
}

func TestShouldEnsureNotifierConfigIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 0)

	config.Notifier = nil

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "A notifier configuration must be provided")
}

func TestShouldAddDefaultAccessControl(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 0)
	assert.NotNil(t, config.AccessControl)
	assert.Equal(t, "deny", config.AccessControl.DefaultPolicy)
}

func TestShouldRaiseErrorWhenSSLCertWithoutKeyIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.SSLCert = "/tmp/cert.pem"

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "No SSL key provided, please check the "ssl_key" which has been configured")
}

func TestShouldRaiseErrorWhenSSLKeyWithoutCertIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.SSLKey = "/tmp/key.pem"

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "No SSL certificate provided, please check the "ssl_cert" which has been configured")
}

func TestShouldNotRaiseErrorWhenBothSSLCertificateAndKeyAreProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.SSLCert = "/tmp/cert.pem"
	config.SSLKey = "/tmp/key.pem"

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 0)
}
