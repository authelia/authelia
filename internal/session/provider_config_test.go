package session

import (
	"crypto/sha256"
	"testing"
	"time"

	"github.com/fasthttp/session"
	"github.com/fasthttp/session/memory"
	"github.com/fasthttp/session/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

func TestShouldCreateInMemorySessionProvider(t *testing.T) {
	// The redis configuration is not provided so we create a in-memory provider.
	configuration := schema.SessionConfiguration{}
	configuration.Domain = "example.com"
	configuration.Name = "my_session"
	configuration.Expiration = "40"
	providerConfig := NewProviderConfig(configuration)

	assert.Equal(t, "my_session", providerConfig.config.CookieName)
	assert.Equal(t, "example.com", providerConfig.config.Domain)
	assert.Equal(t, true, providerConfig.config.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.config.Expires)
	assert.True(t, providerConfig.config.IsSecureFunc(nil))

	assert.Equal(t, "memory", providerConfig.providerName)
	assert.IsType(t, &memory.Config{}, providerConfig.providerConfig)
}

func TestShouldCreateRedisSessionProvider(t *testing.T) {
	// The redis configuration is not provided so we create a in-memory provider.
	configuration := schema.SessionConfiguration{}
	configuration.Domain = "example.com"
	configuration.Name = "my_session"
	configuration.Expiration = "40"
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:     "redis.example.com",
		Port:     6379,
		Password: "pass",
	}
	providerConfig := NewProviderConfig(configuration)

	assert.Equal(t, "my_session", providerConfig.config.CookieName)
	assert.Equal(t, "example.com", providerConfig.config.Domain)
	assert.Equal(t, true, providerConfig.config.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.config.Expires)
	assert.True(t, providerConfig.config.IsSecureFunc(nil))

	assert.Equal(t, "redis", providerConfig.providerName)
	assert.IsType(t, &redis.Config{}, providerConfig.providerConfig)

	pConfig := providerConfig.providerConfig.(*redis.Config)
	assert.Equal(t, "redis.example.com", pConfig.Host)
	assert.Equal(t, int64(6379), pConfig.Port)
	assert.Equal(t, "pass", pConfig.Password)
	// DbNumber is the fasthttp/session property for the Redis DB Index
	assert.Equal(t, 0, pConfig.DbNumber)
}

func TestShouldSetDbNumber(t *testing.T) {
	configuration := schema.SessionConfiguration{}
	configuration.Domain = "example.com"
	configuration.Name = "my_session"
	configuration.Expiration = "40"
	configuration.Redis = &schema.RedisSessionConfiguration{
		Host:          "redis.example.com",
		Port:          6379,
		Password:      "pass",
		DatabaseIndex: 5,
	}
	providerConfig := NewProviderConfig(configuration)
	assert.Equal(t, "redis", providerConfig.providerName)
	assert.IsType(t, &redis.Config{}, providerConfig.providerConfig)
	pConfig := providerConfig.providerConfig.(*redis.Config)
	// DbNumber is the fasthttp/session property for the Redis DB Index
	assert.Equal(t, 5, pConfig.DbNumber)
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
	pConfig := providerConfig.providerConfig.(*redis.Config)

	payload := session.Dict{}
	payload.Set("key", "value")

	encoded, err := pConfig.SerializeFunc(payload)
	require.NoError(t, err)

	// Now we try to decrypt what has been serialized
	key := sha256.Sum256([]byte("abc"))
	decrypted, err := utils.Decrypt(encoded, &key)
	require.NoError(t, err)

	decoded := session.Dict{}
	_, err = decoded.UnmarshalMsg(decrypted)
	assert.Equal(t, "value", decoded.Get("key"))
}
