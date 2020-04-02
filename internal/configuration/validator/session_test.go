package validator

import (
	"testing"

	"github.com/authelia/authelia/internal/configuration/schema"
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

func TestShouldRaiseErrorWhenRedisIsUsedAndPasswordNotSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Secret = ""

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	validator.Clear()

	// Set redis config because password must be set only when redis is used.
	config.Redis = &schema.RedisSessionConfiguration{}

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

func TestShouldRaiseErrorWhenBadInactivityAndExpirationSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Inactivity = -1
	config.Expiration = -1

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 2)
	assert.EqualError(t, validator.Errors()[0], "Set expiration of the session above 0")
	assert.EqualError(t, validator.Errors()[1], "Set inactivity of the session 0 or above")
}

func TestShouldRaiseErrorWhenBadRememberMeDurationSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.RememberMeDuration = "1 year"

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Error occurred parsing remember_me_duration string: could not convert the input string of 1 year into a duration")
}

func TestShouldSetDefaultRememberMeDuration(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, config.RememberMeDuration, schema.DefaultSessionConfiguration.RememberMeDuration)
}
