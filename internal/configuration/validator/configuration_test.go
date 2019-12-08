package validator

import (
	"testing"

	"github.com/clems4ever/authelia/internal/configuration/schema"
	"github.com/stretchr/testify/assert"
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
	config.Storage = &schema.StorageConfiguration{
		Local: &schema.LocalStorageConfiguration{
			Path: "abc",
		},
	}
	return config
}

func TestShouldNotUpdateConfig(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()

	Validate(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, 9090, config.Port)
	assert.Equal(t, "info", config.LogsLevel)
}

func TestShouldValidateAndUpdatePort(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Port = 0

	Validate(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, 8080, config.Port)
}

func TestShouldValidateAndUpdateHost(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.Host = ""

	Validate(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, "0.0.0.0", config.Host)
}

func TestShouldValidateAndUpdateLogsLevel(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultConfig()
	config.LogsLevel = ""

	Validate(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, "info", config.LogsLevel)
}
