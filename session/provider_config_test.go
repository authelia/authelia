package session

import (
	"testing"
	"time"

	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/fasthttp/session/memory"
	"github.com/fasthttp/session/redis"
	"github.com/stretchr/testify/assert"
)

func TestShouldCreateInMemorySessionProvider(t *testing.T) {
	// The redis configuration is not provided so we create a in-memory provider.
	configuration := schema.SessionConfiguration{}
	configuration.Domain = "example.com"
	configuration.Name = "my_session"
	configuration.Expiration = 40
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
	configuration.Expiration = 40
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
}
