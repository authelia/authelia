package session

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/fasthttp/session/v2"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

func TestShouldCreateInMemorySessionProvider(t *testing.T) {
	// The redis configuration is not provided so we create a in-memory provider.
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration
	providerConfig := NewProviderConfig(configuration)

	assert.Equal(t, "my_session", providerConfig.config.CookieName)
	assert.Equal(t, testDomain, providerConfig.config.Domain)
	assert.Equal(t, true, providerConfig.config.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.config.Expiration)
	assert.True(t, providerConfig.config.IsSecureFunc(nil))

	assert.Equal(t, "memory", providerConfig.providerName)
}

func TestShouldCreateRedisSessionProvider(t *testing.T) {
	// The redis configuration is not provided so we create a in-memory provider.
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:     "redis.example.com",
		Port:     6379,
		Password: "pass",
	}
	providerConfig := NewProviderConfig(configuration)

	assert.Equal(t, "my_session", providerConfig.config.CookieName)
	assert.Equal(t, testDomain, providerConfig.config.Domain)
	assert.Equal(t, true, providerConfig.config.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.config.Expiration)
	assert.True(t, providerConfig.config.IsSecureFunc(nil))

	assert.Equal(t, "redis", providerConfig.providerName)

	pConfig := providerConfig.redisConfig
	assert.Equal(t, "redis.example.com:6379", pConfig.Addr)
	assert.Equal(t, "pass", pConfig.Password)
	// DbNumber is the fasthttp/session property for the Redis DB Index
	assert.Equal(t, 0, pConfig.DB)
}

func TestShouldCreateRedisSessionProviderWithUnixSocket(t *testing.T) {
	// The redis configuration is not provided so we create a in-memory provider.
	configuration := schema.SessionConfiguration{}
	configuration.Domain = testDomain
	configuration.Name = testName
	configuration.Expiration = testExpiration
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:     "/var/run/redis/redis.sock",
		Port:     0,
		Password: "pass",
	}
	providerConfig := NewProviderConfig(configuration)

	assert.Equal(t, "my_session", providerConfig.config.CookieName)
	assert.Equal(t, testDomain, providerConfig.config.Domain)
	assert.Equal(t, true, providerConfig.config.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.config.Expiration)
	assert.True(t, providerConfig.config.IsSecureFunc(nil))

	assert.Equal(t, "redis", providerConfig.providerName)

	pConfig := providerConfig.redisConfig
	assert.Equal(t, "/var/run/redis/redis.sock", pConfig.Addr)
	assert.Equal(t, "pass", pConfig.Password)
	// DbNumber is the fasthttp/session property for the Redis DB Index
	assert.Equal(t, 0, pConfig.DB)
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
	providerConfig := NewProviderConfig(configuration)
	assert.Equal(t, "redis", providerConfig.providerName)
	pConfig := providerConfig.redisConfig
	// DbNumber is the fasthttp/session property for the Redis DB Index
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
	providerConfig := NewProviderConfig(configuration)

	payload := session.Dict{}
	payload.Set("key", "value")

	encoded, err := providerConfig.config.EncodeFunc(payload)
	require.NoError(t, err)

	// Now we try to decrypt what has been serialized
	key := sha256.Sum256([]byte("abc"))
	decrypted, err := utils.Decrypt(encoded, &key)
	require.NoError(t, err)

	decoded := session.Dict{}
	_, _ = decoded.UnmarshalMsg(decrypted)
	assert.Equal(t, "value", decoded.Get("key"))
}
