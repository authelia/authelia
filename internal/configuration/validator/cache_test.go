package validator

import (
	"crypto/tls"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidateCache(t *testing.T) {
	testCases := []struct {
		name     string
		have     schema.Cache
		expected func(t *testing.T, have schema.Cache)
		errs     []string
	}{
		{
			"ShouldAllowNoProvider",
			schema.Cache{},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultsForRedis",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com:6379")}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, schema.DefaultRedisCacheConfiguration.DialerRetries, have.Redis.DialerRetries)
				assert.Equal(t, schema.DefaultRedisCacheConfiguration.DialerRetryTimeout, have.Redis.DialerRetryTimeout)
				assert.Equal(t, schema.DefaultRedisCacheConfiguration.ReadBufferSize, have.Redis.ReadBufferSize)
				assert.Equal(t, schema.DefaultRedisCacheConfiguration.WriteBufferSize, have.Redis.WriteBufferSize)
				assert.Equal(t, schema.DefaultRedisCacheConfiguration.FailingTimeout, have.Redis.FailingTimeout)
				assert.Equal(t, schema.DefaultRedisCacheConfiguration.DialTimeout, have.Redis.DialTimeout)
				assert.Equal(t, schema.DefaultRedisCacheConfiguration.ReadTimeout, have.Redis.ReadTimeout)
				assert.Equal(t, schema.DefaultRedisCacheConfiguration.WriteTimeout, have.Redis.WriteTimeout)
				assert.Equal(t, schema.DefaultRedisCacheConfiguration.MaximumRetries, have.Redis.MaximumRetries)
				assert.Equal(t, schema.DefaultRedisCacheConfiguration.PoolSize, have.Redis.PoolSize)
			},
			nil,
		},
		{
			"ShouldNotOverrideConfiguredRedisValues",
			schema.Cache{Redis: &schema.RedisCache{
				Address:        mustAddressTCP("tcp://redis.example.com:6379"),
				DialTimeout:    time.Second * 30,
				MaximumRetries: 9,
				PoolSize:       32,
				DialerRetries:  11,
				ReadBufferSize: 4096,
				FailingTimeout: time.Second * 45,
			}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, 11, have.Redis.DialerRetries)
				assert.Equal(t, 4096, have.Redis.ReadBufferSize)
				assert.Equal(t, time.Second*45, have.Redis.FailingTimeout)
				assert.Equal(t, time.Second*30, have.Redis.DialTimeout)
				assert.Equal(t, 9, have.Redis.MaximumRetries)
				assert.Equal(t, 32, have.Redis.PoolSize)
			},
			nil,
		},
		{
			"ShouldAllowRedisUnixSocket",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("unix:///var/run/redis/redis.sock")}},
			nil,
			nil,
		},
		{
			"ShouldRaiseErrorOnRedisMissingAddress",
			schema.Cache{Redis: &schema.RedisCache{}},
			nil,
			[]string{"cache: redis: option 'address' is required"},
		},
		{
			"ShouldRaiseErrorOnRedisAddressWithoutPort",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com")}},
			nil,
			[]string{"cache: redis: option 'address' must have a port but it's configured as 'tcp://redis.example.com:0'"},
		},
		{
			"ShouldRaiseErrorOnRedisDatabaseOutOfRange",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com:6379"), Database: 16}},
			nil,
			[]string{"cache: redis: option 'database' must be between 0 and 15 but it's configured as '16'"},
		},
		{
			"ShouldRaiseErrorOnRedisBadTLSVersions",
			schema.Cache{Redis: &schema.RedisCache{
				Address: mustAddressTCP("tcp://redis.example.com:6379"),
				TLS: &schema.TLS{
					MinimumVersion: schema.TLSVersion{Value: tls.VersionTLS13},
					MaximumVersion: schema.TLSVersion{Value: tls.VersionTLS10},
				},
			}},
			nil,
			[]string{"cache: redis: tls: option combination of 'minimum_version' and 'maximum_version' is invalid: minimum version TLS 1.3 is greater than the maximum version TLS 1.0"},
		},
		{
			"ShouldSetDefaultTLSServerNameFromAddress",
			schema.Cache{Redis: &schema.RedisCache{
				Address: mustAddressTCP("tcp://redis.example.com:6379"),
				TLS:     &schema.TLS{},
			}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, "redis.example.com", have.Redis.TLS.ServerName)
				assert.Equal(t, uint16(tls.VersionTLS12), have.Redis.TLS.MinimumVersion.Value)
			},
			nil,
		},
		{
			"ShouldSetDefaultsForRedisSentinel",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName: "mysentinel",
				Addresses:  []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
			}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, "failover", have.RedisSentinel.SentinelMode)
				assert.Equal(t, schema.DefaultRedisSentinelCacheConfiguration.PoolSize, have.RedisSentinel.PoolSize)
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnRedisSentinelMissingMasterName",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				Addresses: []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
			}},
			nil,
			[]string{"cache: redis_sentinel: option 'master_name' is required"},
		},
		{
			"ShouldRaiseErrorOnRedisSentinelMissingAddresses",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{MasterName: "mysentinel"}},
			nil,
			[]string{"cache: redis_sentinel: option 'addresses' is required"},
		},
		{
			"ShouldRaiseErrorOnRedisSentinelBadMode",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName:   "mysentinel",
				Addresses:    []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
				SentinelMode: "bad",
			}},
			nil,
			[]string{"cache: redis_sentinel: option 'sentinel_mode' must be one of 'cluster' or 'failover' but it's configured as 'bad'"},
		},
		{
			"ShouldRaiseErrorOnRedisClusterMissingAddresses",
			schema.Cache{RedisCluster: &schema.RedisClusterCache{}},
			nil,
			[]string{"cache: redis_cluster: option 'addresses' is required"},
		},
		{
			"ShouldSetDefaultsForRedisCluster",
			schema.Cache{RedisCluster: &schema.RedisClusterCache{
				Addresses: []*schema.AddressTCP{mustAddressTCP("tcp://node1:6379")},
			}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, schema.DefaultRedisClusterCacheConfiguration.MaximumRedirects, have.RedisCluster.MaximumRedirects)
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnMultipleProviders",
			schema.Cache{
				Redis:         &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com:6379")},
				RedisSentinel: &schema.RedisSentinelCache{MasterName: "mysentinel"},
			},
			nil,
			[]string{"cache: only one provider can be configured at a time but 'redis' and 'redis_sentinel' are configured"},
		},
		{
			"ShouldRaiseErrorOnAllThreeProviders",
			schema.Cache{
				Redis:         &schema.RedisCache{},
				RedisSentinel: &schema.RedisSentinelCache{},
				RedisCluster:  &schema.RedisClusterCache{},
			},
			nil,
			[]string{"cache: only one provider can be configured at a time but 'redis', 'redis_sentinel', and 'redis_cluster' are configured"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.Configuration{Cache: tc.have}

			ValidateCache(config, validator)

			assert.Len(t, validator.Warnings(), 0)

			require.Len(t, validator.Errors(), len(tc.errs))

			for i, expected := range tc.errs {
				assert.EqualError(t, validator.Errors()[i], expected)
			}

			if tc.expected != nil {
				tc.expected(t, config.Cache)
			}
		})
	}
}

func mustAddressTCP(value string) *schema.AddressTCP {
	address, err := schema.NewAddressFromNetworkValuesDefault(value, 0, schema.AddressSchemeTCP, schema.AddressSchemeUnix)
	if err != nil {
		panic(err)
	}

	return &schema.AddressTCP{Address: *address}
}
