package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSecretKey(t *testing.T) {
	assert.True(t, isSecretKey("my_fake_token"))
	assert.False(t, isSecretKey("my_fake_tokenz"))
	assert.True(t, isSecretKey("my_.fake.secret"))
	assert.True(t, isSecretKey("my.password"))
	assert.False(t, isSecretKey("my.passwords"))
	assert.False(t, isSecretKey("my.passwords"))
}

func TestGetEnvPrefix(t *testing.T) {
	var (
		prefix string
		err    error
	)

	prefix, err = getEnvSecretPrefix("AUTHELIA_TEST")
	assert.EqualError(t, err, "invalid prefix")
	assert.Equal(t, "AUTHELIA_", prefix)

	prefix, err = getEnvSecretPrefix("AUTHELIA_TEST_FILE")
	assert.NoError(t, err)
	assert.Equal(t, "AUTHELIA_", prefix)

	prefix, err = getEnvSecretPrefix("AUTHELIA__TEST_FILE")
	assert.NoError(t, err)
	assert.Equal(t, "AUTHELIA__", prefix)

	prefix, err = getEnvSecretPrefix("NOTAUTHELIA__TEST_FILE")
	assert.EqualError(t, err, "invalid prefix")
	assert.Equal(t, "", prefix)
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

	keys, ignoredKeys := getEnvConfigMap(input)

	key, ok = keys["AUTHELIA__MY_NON_SECRET_CONFIG_ITEM"]
	assert.True(t, ok)
	assert.Equal(t, key, "my.non_secret.config_item")

	key, ok = keys["AUTHELIA__MYSECRET_USER_PASSWORD"]
	assert.True(t, ok)
	assert.Equal(t, key, "mysecret.user_password")

	key, ok = keys["AUTHELIA__MYOTHER_CONFIGKEY"]
	assert.False(t, ok)
	assert.Equal(t, key, "")

	key, ok = keys["AUTHELIA__MYSECRET_PASSWORD"]
	assert.False(t, ok)
	assert.Equal(t, key, "")

	assert.Len(t, ignoredKeys, 6)
	assert.Contains(t, ignoredKeys, "AUTHELIA__MYOTHER_CONFIGKEY_FILE")
	assert.Contains(t, ignoredKeys, "AUTHELIA_MYOTHER_CONFIGKEY_FILE")
	assert.Contains(t, ignoredKeys, "AUTHELIA__MYSECRET_PASSWORD_FILE")
	assert.Contains(t, ignoredKeys, "AUTHELIA_MYSECRET_PASSWORD_FILE")
	assert.Contains(t, ignoredKeys, "AUTHELIA__MYSECRET_USER_PASSWORD_FILE")
	assert.Contains(t, ignoredKeys, "AUTHELIA_MYSECRET_USER_PASSWORD_FILE")
}

func TestGetSecretConfigMap(t *testing.T) {
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

	keys := getSecretConfigMap(input)

	key, ok = keys["AUTHELIA_MY_NON_SECRET_CONFIG_ITEM_FILE"]
	assert.False(t, ok)
	assert.Equal(t, key, "")

	key, ok = keys["AUTHELIA_MYOTHER_CONFIGKEY_FILE"]
	assert.True(t, ok)
	assert.Equal(t, key, "myother.configkey")

	key, ok = keys["AUTHELIA_MYSECRET_PASSWORD_FILE"]
	assert.True(t, ok)
	assert.Equal(t, key, "mysecret.password")

	key, ok = keys["AUTHELIA_MYSECRET_USER_PASSWORD_FILE"]
	assert.True(t, ok)
	assert.Equal(t, key, "mysecret.user_password")

	key, ok = keys["AUTHELIA__MY_NON_SECRET_CONFIG_ITEM_FILE"]
	assert.False(t, ok)
	assert.Equal(t, key, "")

	key, ok = keys["AUTHELIA__MYOTHER_CONFIGKEY_FILE"]
	assert.True(t, ok)
	assert.Equal(t, key, "myother.configkey")

	key, ok = keys["AUTHELIA__MYSECRET_PASSWORD_FILE"]
	assert.True(t, ok)
	assert.Equal(t, key, "mysecret.password")

	key, ok = keys["AUTHELIA__MYSECRET_USER_PASSWORD_FILE"]
	assert.True(t, ok)
	assert.Equal(t, key, "mysecret.user_password")
}
