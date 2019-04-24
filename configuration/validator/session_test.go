package validator

import (
	"testing"

	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/stretchr/testify/assert"
)

func newDefaultSessionConfig() schema.SessionConfiguration {
	config := schema.SessionConfiguration{}
	config.Secret = "a_secret"
	config.Domain = "example.com"
	return config
}

func TestShouldSetDefaultSessionName(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, "authelia_session", config.Name)
}

func TestShouldRaiseErrorWhenPasswordNotSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Secret = ""

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Set secret of the session object")
}

func TestShouldRaiseErrorWhenDomainNotSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Domain = ""

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Set domain of the session object")
}
