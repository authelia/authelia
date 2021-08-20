package session

import (
	"crypto/sha256"
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
	providerConfig := NewSessionConfig(configuration)

	assert.Equal(t, "my_session", providerConfig.CookieName)
	assert.Equal(t, testDomain, providerConfig.Domain)
	assert.Equal(t, true, providerConfig.Secure)
	assert.Equal(t, time.Duration(40)*time.Second, providerConfig.Expiration)
	assert.True(t, providerConfig.IsSecureFunc(nil))
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
		providerConfig := NewSessionConfig(configuration)

		assert.Equal(t, expectedValue, providerConfig.CookieSameSite)
	}
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
	providerConfig := NewSessionConfig(configuration)

	payload := session.Dict{}
	payload.Set("key", "value")

	encoded, err := providerConfig.EncodeFunc(payload)
	require.NoError(t, err)

	// Now we try to decrypt what has been serialized
	key := sha256.Sum256([]byte("abc"))
	decrypted, err := utils.Decrypt(encoded, &key)
	require.NoError(t, err)

	decoded := session.Dict{}
	_, _ = decoded.UnmarshalMsg(decrypted)
	assert.Equal(t, "value", decoded.Get("key"))
}
