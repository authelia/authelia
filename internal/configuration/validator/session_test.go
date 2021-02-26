package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
)

func newDefaultSessionConfig() schema.SessionConfiguration {
	config := schema.SessionConfiguration{}
	config.Secret = testJWTSecret
	config.Domain = "example.com"

	return config
}

func TestShouldSetDefaultSessionName(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())
	assert.Equal(t, schema.DefaultSessionConfiguration.Name, config.Name)
}

func TestShouldSetDefaultSessionInactivity(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())
	assert.Equal(t, schema.DefaultSessionConfiguration.Inactivity, config.Inactivity)
}

func TestShouldSetDefaultSessionExpiration(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())
	assert.Equal(t, schema.DefaultSessionConfiguration.Expiration, config.Expiration)
}

func TestShouldHandleRedisConfigSuccessfully(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	validator.Clear()

	// Set redis config because password must be set only when redis is used.
	config.Redis = &schema.RedisSessionConfiguration{
		Host:     "redis.localhost",
		Port:     6379,
		Password: "password",
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())
}

func TestShouldRaiseErrorWithInvalidRedisPortLow(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Host: "authelia-port-1",
		Port: -1,
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionRedisPortRange, "redis"))
}

func TestShouldRaiseErrorWithInvalidRedisPortHigh(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Host: "authelia-port-1",
		Port: 65536,
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionRedisPortRange, "redis"))
}

func TestShouldNotAllowBothRedisAndRedisSentinel(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Redis = &schema.RedisSessionConfiguration{
		Host:     "redis.localhost",
		Port:     6379,
		Password: "password",
	}

	config.RedisSentinel = &schema.RedisSentinelSessionConfiguration{}
	config.RedisSentinel.Host = "authelia"
	config.RedisSentinel.Port = 6379
	config.RedisSentinel.Password = "zpass"

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "Must only specify only one session provider (redis or redis_sentinel)")
}

func TestShouldRequireSentinelHostAndPort(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.RedisSentinel = &schema.RedisSentinelSessionConfiguration{}
	config.RedisSentinel.Password = "pazzword"

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "The host and port must be specified when using the redis sentinel session provider")
}

func TestShouldRequireSentinelNodeHost(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.RedisSentinel = &schema.RedisSentinelSessionConfiguration{}
	config.RedisSentinel.Password = "password"

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "The host and port must be specified when using the redis sentinel session provider")
}

func TestShouldSetDefaultRedisSentinelPort(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.RedisSentinel = &schema.RedisSentinelSessionConfiguration{}
	config.RedisSentinel.Host = "authelia-sentinel"

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())

	assert.Equal(t, 26379, config.RedisSentinel.Port)
}

func TestShouldNotRaiseErrorWithInvalidRedisSentinelPortMax(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.RedisSentinel = &schema.RedisSentinelSessionConfiguration{}
	config.RedisSentinel.Host = "authelia-sentinel-port-65535"
	config.RedisSentinel.Port = 65535

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())
}

func TestShouldRaiseErrorWithInvalidRedisSentinelPortLow(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.RedisSentinel = &schema.RedisSentinelSessionConfiguration{}
	config.RedisSentinel.Host = "authelia-sentinel-port-1"
	config.RedisSentinel.Port = -1

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionRedisPortRange, "redis sentinel"))
}

func TestShouldRaiseErrorWithInvalidRedisSentinelPortHigh(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.RedisSentinel = &schema.RedisSentinelSessionConfiguration{}
	config.RedisSentinel.Host = "authelia-sentinel-port-65536"
	config.RedisSentinel.Port = 65536

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionRedisPortRange, "redis sentinel"))
}

func TestShouldRaiseErrorWhenRedisIsUsedAndSecretNotSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Secret = ""

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	validator.Clear()

	// Set redis config because password must be set only when redis is used.
	config.Redis = &schema.RedisSessionConfiguration{
		Host: "redis.localhost",
		Port: 6379,
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionSecretRedisProvider, "redis"))
}

func TestShouldRaiseErrorWhenRedisSentinelIsUsedAndSecretNotSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Secret = ""

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	validator.Clear()

	// Set redis config because password must be set only when redis is used.
	config.RedisSentinel = &schema.RedisSentinelSessionConfiguration{}

	config.RedisSentinel.Host = "redis.localhost"
	config.RedisSentinel.Port = 26379

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionSecretRedisProvider, "redis sentinel"))
}

func TestShouldRaiseErrorWhenRedisHasHostnameButNoPort(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.Len(t, validator.Errors(), 0)
	validator.Clear()

	// Set redis config because password must be set only when redis is used.
	config.Redis = &schema.RedisSessionConfiguration{
		Host: "redis.localhost",
		Port: 0,
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "A redis port different than 0 must be provided")
}

func TestShouldRaiseErrorWhenDomainNotSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Domain = ""

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Set domain of the session object")
}

func TestShouldRaiseErrorWhenDomainIsWildcard(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Domain = "*.example.com"

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "The domain of the session must be the root domain you're protecting instead of a wildcard domain")
}

func TestShouldRaiseErrorWhenBadInactivityAndExpirationSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Inactivity = testBadTimer
	config.Expiration = testBadTimer

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 2)
	assert.EqualError(t, validator.Errors()[0], "Error occurred parsing session expiration string: Could not convert the input string of -1 into a duration")
	assert.EqualError(t, validator.Errors()[1], "Error occurred parsing session inactivity string: Could not convert the input string of -1 into a duration")
}

func TestShouldRaiseErrorWhenBadRememberMeDurationSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.RememberMeDuration = "1 year"

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "Error occurred parsing session remember_me_duration string: Could not convert the input string of 1 year into a duration")
}

func TestShouldSetDefaultRememberMeDuration(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())
	assert.Equal(t, config.RememberMeDuration, schema.DefaultSessionConfiguration.RememberMeDuration)
}
