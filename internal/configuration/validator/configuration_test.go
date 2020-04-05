package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func newDefaultConfig() schema.Configuration {
	config := schema.Configuration{}
	config.Host = "127.0.0.1"
	config.Port = 9090
	config.LogLevel = "info"
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
	assert.Equal(t, "info", config.LogLevel)
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
	config.LogLevel = ""

	Validate(&config, validator)

	require.Len(t, validator.Errors(), 0)
	assert.Equal(t, "info", config.LogLevel)
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

func TestShouldRaiseErrorWhenTLSCertWithoutKeyIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.TLSCert = "/tmp/cert.pem"

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "No TLS key provided, please check the \"tls_key\" which has been configured")
}

func TestShouldRaiseErrorWhenTLSKeyWithoutCertIsProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.TLSKey = "/tmp/key.pem"

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "No TLS certificate provided, please check the \"tls_cert\" which has been configured")
}

func TestShouldNotRaiseErrorWhenBothTLSCertificateAndKeyAreProvided(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.TLSCert = "/tmp/cert.pem"
	config.TLSKey = "/tmp/key.pem"

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 0)
}

func TestShouldRaiseErrorWithUndefinedJWTSecretKey(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.JWTSecret = ""

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Provide a JWT secret using \"jwt_secret\" key")
}

func TestShouldRaiseErrorWithBadDefaultRedirectionURL(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.DefaultRedirectionURL = "abc"

	Validate(&config, validator)
	require.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Unable to parse default redirection url")
}
