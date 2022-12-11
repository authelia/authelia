package validator

import (
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newDefaultSessionConfig() schema.SessionConfiguration {
	config := schema.SessionConfiguration{}
	config.Secret = testJWTSecret
	config.Domain = examplecom

	return config
}

func TestShouldSetDefaultSessionValues(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())
	assert.Equal(t, schema.DefaultSessionConfiguration.Name, config.Name)
	assert.Equal(t, schema.DefaultSessionConfiguration.Inactivity, config.Inactivity)
	assert.Equal(t, schema.DefaultSessionConfiguration.Expiration, config.Expiration)
	assert.Equal(t, schema.DefaultSessionConfiguration.RememberMeDuration, config.RememberMeDuration)
	assert.Equal(t, schema.DefaultSessionConfiguration.SameSite, config.SameSite)
}

func TestShouldSetDefaultSessionValuesWhenNegative(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Expiration, config.Inactivity, config.RememberMeDuration = -1, -1, -2

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)
	assert.Equal(t, schema.DefaultSessionConfiguration.Inactivity, config.Inactivity)
	assert.Equal(t, schema.DefaultSessionConfiguration.Expiration, config.Expiration)
	assert.Equal(t, schema.DefaultSessionConfiguration.RememberMeDuration, config.RememberMeDuration)
}

func TestShouldWarnSessionValuesWhenPotentiallyInvalid(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Domain = ".example.com"

	ValidateSession(&config, validator)

	require.Len(t, validator.Warnings(), 1)
	assert.Len(t, validator.Errors(), 0)

	assert.EqualError(t, validator.Warnings()[0], "session: option 'domain' has a prefix of '.' which is not supported or intended behaviour: you can use this at your own risk but we recommend removing it")
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

	assert.Equal(t, 8, config.Redis.MaximumActiveConnections)
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

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionRedisPortRange, -1))
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

	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionRedisPortRange, 65536))
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
	assert.EqualError(t, validator.Errors()[0], fmt.Sprintf(errFmtSessionSecretRequired, "redis"))
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
	assert.EqualError(t, validator.Errors()[0], "session: redis: option 'port' must be between 1 and 65535 but is configured as '0'")
}

func TestShouldRaiseOneErrorWhenRedisHighAvailabilityHasNodesWithNoHost(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Host: "redis",
		Port: 6379,
		HighAvailability: &schema.RedisHighAvailabilityConfiguration{
			SentinelName:     "authelia-sentinel",
			SentinelPassword: "abc123",
			Nodes: []schema.RedisNode{
				{
					Port: 26379,
				},
				{
					Port: 26379,
				},
			},
		},
	}

	ValidateSession(&config, validator)

	errors := validator.Errors()

	assert.False(t, validator.HasWarnings())
	require.Len(t, errors, 1)

	assert.EqualError(t, errors[0], "session: redis: high_availability: option 'nodes': option 'host' is required for each node but one or more nodes are missing this")
}

func TestShouldRaiseOneErrorWhenRedisHighAvailabilityDoesNotHaveSentinelName(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Host: "redis",
		Port: 6379,
		HighAvailability: &schema.RedisHighAvailabilityConfiguration{
			SentinelPassword: "abc123",
		},
	}

	ValidateSession(&config, validator)

	errors := validator.Errors()

	assert.False(t, validator.HasWarnings())
	require.Len(t, errors, 1)

	assert.EqualError(t, errors[0], "session: redis: high_availability: option 'sentinel_name' is required")
}

func TestShouldUpdateDefaultPortWhenRedisSentinelHasNodes(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Host: "redis",
		Port: 6379,
		HighAvailability: &schema.RedisHighAvailabilityConfiguration{
			SentinelName:     "authelia-sentinel",
			SentinelPassword: "abc123",
			Nodes: []schema.RedisNode{
				{
					Host: "node-1",
					Port: 333,
				},
				{
					Host: "node-2",
				},
				{
					Host: "node-3",
				},
			},
		},
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())

	assert.Equal(t, 333, config.Redis.HighAvailability.Nodes[0].Port)
	assert.Equal(t, 26379, config.Redis.HighAvailability.Nodes[1].Port)
	assert.Equal(t, 26379, config.Redis.HighAvailability.Nodes[2].Port)
}

func TestShouldRaiseErrorsWhenRedisSentinelOptionsIncorrectlyConfigured(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Secret = ""
	config.Redis = &schema.RedisSessionConfiguration{
		Port: 65536,
		HighAvailability: &schema.RedisHighAvailabilityConfiguration{
			SentinelName:     "sentinel",
			SentinelPassword: "abc123",
			Nodes: []schema.RedisNode{
				{
					Host: "node1",
					Port: 26379,
				},
			},
			RouteByLatency: true,
			RouteRandomly:  true,
		},
	}

	ValidateSession(&config, validator)

	errors := validator.Errors()

	assert.False(t, validator.HasWarnings())
	require.Len(t, errors, 2)

	assert.EqualError(t, errors[0], fmt.Sprintf(errFmtSessionRedisPortRange, 65536))
	assert.EqualError(t, errors[1], fmt.Sprintf(errFmtSessionSecretRequired, "redis"))

	validator.Clear()

	config.Redis.Port = -1

	ValidateSession(&config, validator)

	errors = validator.Errors()

	assert.False(t, validator.HasWarnings())
	require.Len(t, errors, 2)

	assert.EqualError(t, errors[0], fmt.Sprintf(errFmtSessionRedisPortRange, -1))
	assert.EqualError(t, errors[1], fmt.Sprintf(errFmtSessionSecretRequired, "redis"))
}

func TestShouldNotRaiseErrorsAndSetDefaultPortWhenRedisSentinelPortBlank(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Host: "mysentinelHost",
		Port: 0,
		HighAvailability: &schema.RedisHighAvailabilityConfiguration{
			SentinelName:     "sentinel",
			SentinelPassword: "abc123",
			Nodes: []schema.RedisNode{
				{
					Host: "node1",
					Port: 26379,
				},
			},
			RouteByLatency: true,
			RouteRandomly:  true,
		},
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())

	assert.Equal(t, 26379, config.Redis.Port)
}

func TestShouldRaiseErrorWhenRedisHostAndHighAvailabilityNodesEmpty(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Port: 26379,
		HighAvailability: &schema.RedisHighAvailabilityConfiguration{
			SentinelName:     "sentinel",
			SentinelPassword: "abc123",
			RouteByLatency:   true,
			RouteRandomly:    true,
		},
	}

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], errFmtSessionRedisHostOrNodesRequired)
}

func TestShouldRaiseErrorsWhenRedisHostNotSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Port: 6379,
	}

	ValidateSession(&config, validator)

	errors := validator.Errors()

	assert.False(t, validator.HasWarnings())
	require.Len(t, errors, 1)

	assert.EqualError(t, errors[0], errFmtSessionRedisHostRequired)
}

func TestShouldSetDefaultRedisTLSOptions(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Host: "redis.local",
		Port: 6379,
		TLS:  &schema.TLSConfig{},
	}

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	assert.Equal(t, uint16(tls.VersionTLS12), config.Redis.TLS.MinimumVersion.Value)
	assert.Equal(t, uint16(0), config.Redis.TLS.MaximumVersion.Value)
	assert.Equal(t, "redis.local", config.Redis.TLS.ServerName)
}

func TestShouldRaiseErrorOnBadRedisTLSOptionsSSL30(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Host: "redis.local",
		Port: 6379,
		TLS: &schema.TLSConfig{
			MinimumVersion: schema.TLSVersion{Value: tls.VersionSSL30}, //nolint:staticcheck
		},
	}

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "session: redis: tls: option 'minimum_version' is invalid: minimum version is TLS1.0 but SSL3.0 was configured")
}

func TestShouldRaiseErrorOnBadRedisTLSOptionsMinVerGreaterThanMax(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	config.Redis = &schema.RedisSessionConfiguration{
		Host: "redis.local",
		Port: 6379,
		TLS: &schema.TLSConfig{
			MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
			MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS10},
		},
	}

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	require.Len(t, validator.Errors(), 1)

	assert.EqualError(t, validator.Errors()[0], "session: redis: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS1.3 is greater than the maximum version TLS1.0")
}

func TestShouldRaiseErrorWhenDomainNotSet(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Domain = ""

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "session: option 'domain' is required")
}

func TestShouldRaiseErrorWhenDomainIsWildcard(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Domain = "*.example.com"

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "session: option 'domain' must be the domain you wish to protect not a wildcard domain but it is configured as '*.example.com'")
}

func TestShouldRaiseErrorWhenSameSiteSetIncorrectly(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.SameSite = "NOne"

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.Len(t, validator.Errors(), 1)
	assert.EqualError(t, validator.Errors()[0], "session: option 'same_site' must be one of 'none', 'lax', 'strict' but is configured as 'NOne'")
}

func TestShouldNotRaiseErrorWhenSameSiteSetCorrectly(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	validOptions := []string{"none", "lax", "strict"}

	for _, opt := range validOptions {
		config.SameSite = opt

		ValidateSession(&config, validator)

		assert.False(t, validator.HasWarnings())
		assert.Len(t, validator.Errors(), 0)
	}
}

func TestShouldSetDefaultWhenNegativeAndNotOverrideDisabledRememberMe(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()
	config.Inactivity = -1
	config.Expiration = -1
	config.RememberMeDuration = schema.RememberMeDisabled

	ValidateSession(&config, validator)

	assert.Len(t, validator.Warnings(), 0)
	assert.Len(t, validator.Errors(), 0)

	assert.Equal(t, schema.DefaultSessionConfiguration.Inactivity, config.Inactivity)
	assert.Equal(t, schema.DefaultSessionConfiguration.Expiration, config.Expiration)
	assert.Equal(t, schema.RememberMeDisabled, config.RememberMeDuration)
}

func TestShouldSetDefaultRememberMeDuration(t *testing.T) {
	validator := schema.NewStructValidator()
	config := newDefaultSessionConfig()

	ValidateSession(&config, validator)

	assert.False(t, validator.HasWarnings())
	assert.False(t, validator.HasErrors())
	assert.Equal(t, config.RememberMeDuration, schema.DefaultSessionConfiguration.RememberMeDuration)
}
