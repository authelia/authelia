package session

import (
	"crypto/sha256"
	"crypto/tls"
	"testing"
	"time"

	"github.com/fasthttp/session/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldCreateInMemorySessionProvider(t *testing.T) {
	// The redis configuration is not provided so we create a in-memory provider.
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration
	providerConfig := NewProviderConfig(configuration, nil)

	assert.Equal(t, "my_session", providerConfig.config.CookieName)
	assert.Equal(t, testDomain, providerConfig.config.Domain)
	assert.Equal(t, true, providerConfig.config.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.config.Expiration)
	assert.True(t, providerConfig.config.IsSecureFunc(nil))
	assert.Equal(t, "memory", providerConfig.providerName)
}

func TestShouldCreateRedisSessionProviderTLS(t *testing.T) {
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:     "redis.example.com",
		Port:     6379,
		Password: "pass",
		TLS: &schema.TLSConfig{
			ServerName:     "redis.fqdn.example.com",
			MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
		},
	}
	providerConfig := NewProviderConfig(configuration, nil)

	assert.Nil(t, providerConfig.redisSentinelConfig)
	assert.Equal(t, "my_session", providerConfig.config.CookieName)
	assert.Equal(t, testDomain, providerConfig.config.Domain)
	assert.Equal(t, true, providerConfig.config.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.config.Expiration)
	assert.True(t, providerConfig.config.IsSecureFunc(nil))

	assert.Equal(t, "redis", providerConfig.providerName)

	pConfig := providerConfig.redisConfig
	assert.Equal(t, "redis.example.com:6379", pConfig.Addr)
	assert.Equal(t, "pass", pConfig.Password)
	// DbNumber is the fasthttp/session property for the Redis DB Index.
	assert.Equal(t, 0, pConfig.DB)
	assert.Equal(t, 0, pConfig.PoolSize)
	assert.Equal(t, 0, pConfig.MinIdleConns)

	require.NotNil(t, pConfig.TLSConfig)
	require.Equal(t, uint16(tls.VersionTLS13), pConfig.TLSConfig.MinVersion)
	require.Equal(t, "redis.fqdn.example.com", pConfig.TLSConfig.ServerName)
	require.False(t, pConfig.TLSConfig.InsecureSkipVerify)
}

func TestShouldCreateRedisSessionProvider(t *testing.T) {
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:     "redis.example.com",
		Port:     6379,
		Password: "pass",
	}
	providerConfig := NewProviderConfig(configuration, nil)

	assert.Nil(t, providerConfig.redisSentinelConfig)
	assert.Equal(t, "my_session", providerConfig.config.CookieName)
	assert.Equal(t, testDomain, providerConfig.config.Domain)
	assert.Equal(t, true, providerConfig.config.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.config.Expiration)
	assert.True(t, providerConfig.config.IsSecureFunc(nil))

	assert.Equal(t, "redis", providerConfig.providerName)

	pConfig := providerConfig.redisConfig
	assert.Equal(t, "redis.example.com:6379", pConfig.Addr)
	assert.Equal(t, "pass", pConfig.Password)
	// DbNumber is the fasthttp/session property for the Redis DB Index.
	assert.Equal(t, 0, pConfig.DB)
	assert.Equal(t, 0, pConfig.PoolSize)
	assert.Equal(t, 0, pConfig.MinIdleConns)

	assert.Nil(t, pConfig.TLSConfig)
}

func TestShouldCreateRedisSentinelSessionProviderWithoutDuplicateHosts(t *testing.T) {
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:                     "REDIS.example.com",
		Port:                     26379,
		Password:                 "pass",
		MaximumActiveConnections: 8,
		MinimumIdleConnections:   2,
		HighAvailability: &schema.RedisHighAvailabilityConfiguration{
			SentinelName:     "mysent",
			SentinelPassword: "mypass",
			Nodes: []schema.RedisNode{
				{
					Host: "redis2.example.com",
					Port: 26379,
				},
				{
					Host: "redis.example.com",
					Port: 26379,
				},
			},
		},
	}

	providerConfig := NewProviderConfig(configuration, nil)

	assert.Len(t, providerConfig.redisSentinelConfig.SentinelAddrs, 2)
	assert.Equal(t, providerConfig.redisSentinelConfig.SentinelAddrs[0], "redis.example.com:26379")
	assert.Equal(t, providerConfig.redisSentinelConfig.SentinelAddrs[1], "redis2.example.com:26379")
}

func TestShouldCreateRedisSentinelSessionProvider(t *testing.T) {
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:                     "redis.example.com",
		Port:                     26379,
		Password:                 "pass",
		MaximumActiveConnections: 8,
		MinimumIdleConnections:   2,
		HighAvailability: &schema.RedisHighAvailabilityConfiguration{
			SentinelName:     "mysent",
			SentinelPassword: "mypass",
			Nodes: []schema.RedisNode{
				{
					Host: "redis2.example.com",
					Port: 26379,
				},
			},
		},
	}
	providerConfig := NewProviderConfig(configuration, nil)

	assert.Nil(t, providerConfig.redisConfig)
	assert.Equal(t, "my_session", providerConfig.config.CookieName)
	assert.Equal(t, testDomain, providerConfig.config.Domain)
	assert.Equal(t, true, providerConfig.config.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.config.Expiration)
	assert.True(t, providerConfig.config.IsSecureFunc(nil))

	assert.Equal(t, "redis-sentinel", providerConfig.providerName)

	pConfig := providerConfig.redisSentinelConfig
	assert.Equal(t, "redis.example.com:26379", pConfig.SentinelAddrs[0])
	assert.Equal(t, "redis2.example.com:26379", pConfig.SentinelAddrs[1])
	assert.Equal(t, "pass", pConfig.Password)
	assert.Equal(t, "mysent", pConfig.MasterName)
	assert.Equal(t, "mypass", pConfig.SentinelPassword)
	assert.False(t, pConfig.RouteRandomly)
	assert.False(t, pConfig.RouteByLatency)
	assert.Equal(t, 8, pConfig.PoolSize)
	assert.Equal(t, 2, pConfig.MinIdleConns)

	// DbNumber is the fasthttp/session property for the Redis DB Index.
	assert.Equal(t, 0, pConfig.DB)
	assert.Nil(t, pConfig.TLSConfig)
}

func TestShouldSetCookieSameSite(t *testing.T) {
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration

	configValueExpectedValue := map[string]fasthttp.CookieSameSite{
		"":        fasthttp.CookieSameSiteLaxMode,
		"lax":     fasthttp.CookieSameSiteLaxMode,
		"strict":  fasthttp.CookieSameSiteStrictMode,
		"none":    fasthttp.CookieSameSiteNoneMode,
		"invalid": fasthttp.CookieSameSiteLaxMode,
	}

	for configValue, expectedValue := range configValueExpectedValue {
		configuration.SameSite = configValue
		providerConfig := NewProviderConfig(configuration, nil)

		assert.Equal(t, expectedValue, providerConfig.config.CookieSameSite)
	}
}

func TestShouldCreateRedisSessionProviderWithUnixSocket(t *testing.T) {
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:     "/var/run/redis/redis.sock",
		Port:     0,
		Password: "pass",
	}

	providerConfig := NewProviderConfig(configuration, nil)

	assert.Nil(t, providerConfig.redisSentinelConfig)

	assert.Equal(t, "my_session", providerConfig.config.CookieName)
	assert.Equal(t, testDomain, providerConfig.config.Domain)
	assert.Equal(t, true, providerConfig.config.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.config.Expiration)
	assert.True(t, providerConfig.config.IsSecureFunc(nil))

	assert.Equal(t, "redis", providerConfig.providerName)

	pConfig := providerConfig.redisConfig
	assert.Equal(t, "/var/run/redis/redis.sock", pConfig.Addr)
	assert.Equal(t, "pass", pConfig.Password)
	// DbNumber is the fasthttp/session property for the Redis DB Index.
	assert.Equal(t, 0, pConfig.DB)
	assert.Nil(t, pConfig.TLSConfig)
}

func TestShouldSetDbNumber(t *testing.T) {
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:          "redis.example.com",
		Port:          6379,
		Password:      "pass",
		DatabaseIndex: 5,
	}

	providerConfig := NewProviderConfig(configuration, nil)

	assert.Nil(t, providerConfig.redisSentinelConfig)

	assert.Equal(t, "redis", providerConfig.providerName)
	pConfig := providerConfig.redisConfig
	// DbNumber is the fasthttp/session property for the Redis DB Index.
	assert.Equal(t, 5, pConfig.DB)
}

func TestShouldUseEncryptingSerializerWithRedis(t *testing.T) {
	configuration := schema.SessionConfiguration{}
	configuration.Secret = "abc"
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:          "redis.example.com",
		Port:          6379,
		Password:      "pass",
		DatabaseIndex: 5,
	}
	providerConfig := NewProviderConfig(configuration, nil)

	payload := session.Dict{}
	payload.Set("key", "value")

	encoded, err := providerConfig.config.EncodeFunc(payload)
	require.NoError(t, err)

	// Now we try to decrypt what has been serialized.
	key := sha256.Sum256([]byte("abc"))
	decrypted, err := utils.Decrypt(encoded, &key)
	require.NoError(t, err)

	decoded := session.Dict{}
	_, _ = decoded.UnmarshalMsg(decrypted)
	assert.Equal(t, "value", decoded.Get("key"))
}
