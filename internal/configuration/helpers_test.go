package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestIsSecretKey(t *testing.T) {
	assert.True(t, IsSecretKey("my_fake_token"))
	assert.False(t, IsSecretKey("my_fake_tokenz"))
	assert.True(t, IsSecretKey("my_.fake.secret"))
	assert.True(t, IsSecretKey("my.password"))
	assert.False(t, IsSecretKey("my.passwords"))
	assert.False(t, IsSecretKey("my.passwords"))
}

func TestGetEnvConfigMaps(t *testing.T) {
	var (
		key string
		ok  bool
	)

	input := []string{
		"my.non_secret.config_item",
		"myother.configkey",
		"mysecret.password",
		"mysecret.user_password",
	}

	keys, ignoredKeys := getEnvConfigMap(input, DefaultEnvPrefix, DefaultEnvDelimiter, deprecations, deprecationsMKM)

	key, ok = keys[DefaultEnvPrefix+"MY_NON_SECRET_CONFIG_ITEM"]
	assert.True(t, ok)
	assert.Equal(t, key, "my.non_secret.config_item")

	key, ok = keys[DefaultEnvPrefix+"MYSECRET_USER_PASSWORD"]
	assert.True(t, ok)
	assert.Equal(t, key, "mysecret.user_password")

	key, ok = keys[DefaultEnvPrefix+"MYOTHER_CONFIGKEY"]
	assert.True(t, ok)
	assert.Equal(t, "myother.configkey", key)

	key, ok = keys[DefaultEnvPrefix+"MYSECRET_PASSWORD"]
	assert.True(t, ok)
	assert.Equal(t, "mysecret.password", key)

	assert.Len(t, ignoredKeys, 6)
	assert.Contains(t, ignoredKeys, DefaultEnvPrefix+MYOTHER_CONFIGKEY_FILE)
	assert.Contains(t, ignoredKeys, DefaultEnvPrefix+MYSECRET_PASSWORD_FILE)
	assert.Contains(t, ignoredKeys, DefaultEnvPrefix+MYSECRET_USER_PASSWORD_FILE)
	assert.Contains(t, ignoredKeys, DefaultEnvPrefix+"IDENTITY_PROVIDERS_OIDC_ISSUER_PRIVATE_KEY_FILE")
	assert.Contains(t, ignoredKeys, DefaultEnvPrefix+"IDENTITY_PROVIDERS_OIDC_ISSUER_CERTIFICATE_CHAIN_FILE")
}

func TestGetSecretConfigMapMockInput(t *testing.T) {
	var (
		key string
		ok  bool
	)

	input := []string{
		"my.non_secret.config_item",
		"myother.configkey",
		"mysecret.password",
		"mysecret.user_password",
	}

	keys := getSecretConfigMap(input, DefaultEnvPrefix, DefaultEnvDelimiter, deprecations)

	key, ok = keys[DefaultEnvPrefix+"MY_NON_SECRET_CONFIG_ITEM_FILE"]
	assert.False(t, ok)
	assert.Equal(t, key, "")

	key, ok = keys[DefaultEnvPrefix+MYOTHER_CONFIGKEY_FILE]
	assert.True(t, ok)
	assert.Equal(t, "myother.configkey", key)

	key, ok = keys[DefaultEnvPrefix+MYSECRET_PASSWORD_FILE]
	assert.True(t, ok)
	assert.Equal(t, "mysecret.password", key)

	key, ok = keys[DefaultEnvPrefix+MYSECRET_USER_PASSWORD_FILE]
	assert.True(t, ok)
	assert.Equal(t, "mysecret.user_password", key)
}

func TestGetSecretConfigMap(t *testing.T) {
	keys := getSecretConfigMap(schema.Keys, DefaultEnvPrefix, DefaultEnvDelimiter, deprecations)

	var (
		key string
		ok  bool
	)

	key, ok = keys[DefaultEnvPrefix+JWT_SECRET_FILE]

	assert.True(t, ok)
	assert.Equal(t, "jwt_secret", key)
}
