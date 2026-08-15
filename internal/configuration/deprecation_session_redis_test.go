package configuration

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestMapSessionRedisToCache(t *testing.T) {
	testCases := []struct {
		name        string
		keys        map[string]any
		expected    map[string]any
		expectedErr string
		expectedWrn string
	}{
		{
			"ShouldMapStandalone",
			map[string]any{
				"session.redis.host":                       "redis.example.com",
				"session.redis.port":                       6379,
				"session.redis.username":                   "authelia",
				"session.redis.password":                   "password",
				"session.redis.database_index":             3,
				"session.redis.timeout":                    "5 seconds",
				"session.redis.max_retries":                5,
				"session.redis.maximum_active_connections": 16,
				"session.redis.minimum_idle_connections":   2,
				"session.redis.tls.server_name":            "redis.example.com",
			},
			map[string]any{
				"cache.redis.address":                       "tcp://redis.example.com:6379",
				"cache.redis.username":                      "authelia",
				"cache.redis.password":                      "password",
				"cache.redis.database":                      3,
				"cache.redis.dial_timeout":                  "5 seconds",
				"cache.redis.maximum_retries":               5,
				"cache.redis.pool_size":                     16,
				"cache.redis.pool_minimum_idle_connections": 2,
				"cache.redis.tls.server_name":               "redis.example.com",
				"session.storage":                           "cache",
			},
			"",
			"configuration keys prefixed with 'session.redis' are deprecated in 4.40.0 and have been replaced by the 'cache.redis' keys: you are not required to make any changes as this has been automatically mapped for you, but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0",
		},
		{
			"ShouldMapStandaloneUnixSocket",
			map[string]any{
				"session.redis.host": "/var/run/redis/redis.sock",
			},
			map[string]any{
				"cache.redis.address": "unix:///var/run/redis/redis.sock",
				"session.storage":     "cache",
			},
			"",
			"configuration keys prefixed with 'session.redis' are deprecated in 4.40.0 and have been replaced by the 'cache.redis' keys: you are not required to make any changes as this has been automatically mapped for you, but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0",
		},
		{
			"ShouldMapStandaloneDefaultPort",
			map[string]any{
				"session.redis.host": "redis.example.com",
			},
			map[string]any{
				"cache.redis.address": "tcp://redis.example.com:6379",
				"session.storage":     "cache",
			},
			"",
			"configuration keys prefixed with 'session.redis' are deprecated in 4.40.0 and have been replaced by the 'cache.redis' keys: you are not required to make any changes as this has been automatically mapped for you, but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0",
		},
		{
			"ShouldMapSentinelWithNodes",
			map[string]any{
				"session.redis.host":                                "sentinel-primary",
				"session.redis.port":                                26379,
				"session.redis.password":                            "password",
				"session.redis.high_availability.sentinel_name":     "mysentinel",
				"session.redis.high_availability.sentinel_username": "sentinel-user",
				"session.redis.high_availability.sentinel_password": "sentinel-pass",
				"session.redis.high_availability.route_by_latency":  true,
				"session.redis.high_availability.route_randomly":    false,
				"session.redis.high_availability.nodes": []any{
					map[string]any{"host": "sentinel-node1", "port": 26379},
					map[string]any{"host": "sentinel-node2"},
					map[string]any{"host": "sentinel-primary", "port": 26379},
				},
			},
			map[string]any{
				"cache.redis_sentinel.master_name":       "mysentinel",
				"cache.redis_sentinel.password":          "password",
				"cache.redis_sentinel.sentinel_username": "sentinel-user",
				"cache.redis_sentinel.sentinel_password": "sentinel-pass",
				"cache.redis_sentinel.route_by_latency":  true,
				"cache.redis_sentinel.route_randomly":    false,
				"cache.redis_sentinel.addresses": []any{
					"tcp://sentinel-primary:26379",
					"tcp://sentinel-node1:26379",
					"tcp://sentinel-node2:26379",
				},
				"session.storage": "cache",
			},
			"",
			"configuration keys prefixed with 'session.redis' are deprecated in 4.40.0 and have been replaced by the 'cache.redis_sentinel' keys: you are not required to make any changes as this has been automatically mapped for you, but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0",
		},
		{
			"ShouldNotOverrideExplicitSessionStorage",
			map[string]any{
				"session.redis.host": "redis.example.com",
				"session.storage":    "internal",
			},
			map[string]any{
				"cache.redis.address": "tcp://redis.example.com:6379",
				"session.storage":     "internal",
			},
			"",
			"configuration keys prefixed with 'session.redis' are deprecated in 4.40.0 and have been replaced by the 'cache.redis' keys: you are not required to make any changes as this has been automatically mapped for you, but to stop this warning being logged you will need to adjust your configuration, and this configuration key and auto-mapping is likely to be removed in 5.0.0",
		},
		{
			"ShouldErrorWhenCacheRedisAlsoConfigured",
			map[string]any{
				"session.redis.host":  "redis.example.com",
				"cache.redis.address": "tcp://other.example.com:6379",
			},
			map[string]any{
				"cache.redis.address": "tcp://other.example.com:6379",
			},
			"configuration keys prefixed with 'session.redis' are deprecated in 4.40.0 and have been replaced by the 'cache.redis' keys: this has not been automatically mapped for you because the replacement keys also exist and you will need to adjust your configuration to remove this message",
			"",
		},
		{
			"ShouldErrorWhenCacheRedisSentinelAlsoConfigured",
			map[string]any{
				"session.redis.host":                            "redis.example.com",
				"session.redis.high_availability.sentinel_name": "mysentinel",
				"cache.redis_sentinel.master_name":              "other",
			},
			map[string]any{
				"cache.redis_sentinel.master_name": "other",
			},
			"configuration keys prefixed with 'session.redis' are deprecated in 4.40.0 and have been replaced by the 'cache.redis_sentinel' keys: this has not been automatically mapped for you because the replacement keys also exist and you will need to adjust your configuration to remove this message",
			"",
		},
		{
			"ShouldErrorOnInvalidPort",
			map[string]any{
				"session.redis.host": "redis.example.com",
				"session.redis.port": "notaport",
			},
			map[string]any{},
			"error occurred performing deprecation mapping for the 'session.redis' keys to the 'cache.redis' keys: error occurred converting the port from a string: strconv.ParseUint: parsing \"notaport\": invalid syntax",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			val := schema.NewStructValidator()

			d := MultiKeyMappedDeprecation{
				Version: model.SemanticVersion{Major: 4, Minor: 40},
				Keys:    deprecationSessionRedisKeys,
				NewKey:  keyCacheRedis,
			}

			mapSessionRedisToCache(d, tc.keys, val)

			assert.Equal(t, tc.expected, tc.keys)

			if tc.expectedErr == "" {
				assert.Len(t, val.Errors(), 0)
			} else {
				require.Len(t, val.Errors(), 1)
				assert.EqualError(t, val.Errors()[0], tc.expectedErr)
			}

			if tc.expectedWrn == "" {
				assert.Len(t, val.Warnings(), 0)
			} else {
				require.Len(t, val.Warnings(), 1)
				assert.EqualError(t, val.Warnings()[0], tc.expectedWrn)
			}
		})
	}
}

func TestSessionRedisDeprecationKeysAreValidConfigurationKeys(t *testing.T) {
	remapped := GetMultiKeyMappedDeprecationKeys()

	for _, key := range deprecationSessionRedisKeys {
		assert.Contains(t, remapped, key)
		assert.NotContains(t, schema.Keys, key)
	}
}
