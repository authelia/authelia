// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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
				assert.Equal(t, 0, have.Redis.PoolSize)
				assert.Equal(t, time.Minute*30, have.Redis.IdleTimeout)
				assert.Equal(t, time.Duration(0), have.Redis.ConnectionLifetime)
				assert.Equal(t, time.Second*4, have.Redis.PoolTimeout)
				assert.Equal(t, time.Millisecond*8, have.Redis.MinimumRetryBackoff)
				assert.Equal(t, time.Millisecond*512, have.Redis.MaximumRetryBackoff)
				assert.Equal(t, 0, have.Redis.PoolMinimumIdleConnections)
				assert.Equal(t, 10, have.Redis.PoolMaximumIdleConnections)
				assert.Equal(t, 10, have.Redis.PoolMaximumConnections)
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
			"ShouldNotOverrideConfiguredRedisPoolValues",
			schema.Cache{Redis: &schema.RedisCache{
				Address:                    mustAddressTCP("tcp://redis.example.com:6379"),
				ConnectionLifetime:         time.Minute,
				PoolTimeout:                time.Second * 2,
				MinimumRetryBackoff:        time.Millisecond * 16,
				MaximumRetryBackoff:        time.Second,
				PoolMinimumIdleConnections: 2,
				PoolMaximumIdleConnections: 4,
				PoolMaximumConnections:     20,
			}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, time.Minute, have.Redis.ConnectionLifetime)
				assert.Equal(t, time.Second*2, have.Redis.PoolTimeout)
				assert.Equal(t, time.Millisecond*16, have.Redis.MinimumRetryBackoff)
				assert.Equal(t, time.Second, have.Redis.MaximumRetryBackoff)
				assert.Equal(t, 2, have.Redis.PoolMinimumIdleConnections)
				assert.Equal(t, 4, have.Redis.PoolMaximumIdleConnections)
				assert.Equal(t, 20, have.Redis.PoolMaximumConnections)
			},
			nil,
		},
		{
			"ShouldDeriveRedisPoolTimeoutFromConfiguredReadTimeout",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com:6379"), ReadTimeout: time.Second * 10}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, time.Second*11, have.Redis.PoolTimeout)
			},
			nil,
		},
		{
			"ShouldResetNegativeRedisPoolSizeToClientDefault",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com:6379"), PoolSize: -1}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, 0, have.Redis.PoolSize)
			},
			nil,
		},
		{
			"ShouldDefaultRedisConnectionLifetimeJitterToZero",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com:6379")}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, time.Duration(0), have.Redis.ConnectionLifetimeJitter)
			},
			nil,
		},
		{
			"ShouldAllowRedisConnectionLifetimeJitterEqualToLifetime",
			schema.Cache{Redis: &schema.RedisCache{
				Address:                  mustAddressTCP("tcp://redis.example.com:6379"),
				ConnectionLifetime:       time.Minute,
				ConnectionLifetimeJitter: time.Minute,
			}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, time.Minute, have.Redis.ConnectionLifetimeJitter)
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnRedisConnectionLifetimeJitterGreaterThanLifetime",
			schema.Cache{Redis: &schema.RedisCache{
				Address:                  mustAddressTCP("tcp://redis.example.com:6379"),
				ConnectionLifetime:       time.Minute,
				ConnectionLifetimeJitter: time.Minute * 2,
			}},
			nil,
			[]string{"cache: redis: option 'connection_lifetime_jitter' must not be greater than option 'connection_lifetime' but they're configured as '2m0s' and '1m0s' respectively"},
		},
		{
			"ShouldRaiseErrorOnRedisSentinelConnectionLifetimeJitterWithoutLifetime",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName:               "mysentinel",
				Addresses:                []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
				ConnectionLifetimeJitter: time.Second * 30,
			}},
			nil,
			[]string{"cache: redis_sentinel: option 'connection_lifetime_jitter' must not be greater than option 'connection_lifetime' but they're configured as '30s' and '0s' respectively"},
		},
		{
			"ShouldRaiseErrorOnRedisClusterConnectionLifetimeJitterGreaterThanLifetime",
			schema.Cache{RedisCluster: &schema.RedisClusterCache{
				Addresses:                []*schema.AddressTCP{mustAddressTCP("tcp://node1:6379")},
				ConnectionLifetime:       time.Second,
				ConnectionLifetimeJitter: time.Second * 5,
			}},
			nil,
			[]string{"cache: redis_cluster: option 'connection_lifetime_jitter' must not be greater than option 'connection_lifetime' but they're configured as '5s' and '1s' respectively"},
		},
		{
			"ShouldNotOverrideConfiguredRedisIdleTimeout",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com:6379"), IdleTimeout: time.Minute * 5}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, time.Minute*5, have.Redis.IdleTimeout)
			},
			nil,
		},
		{
			"ShouldNotOverrideDisabledRedisIdleTimeout",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com:6379"), IdleTimeout: -1}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, time.Duration(-1), have.Redis.IdleTimeout)
			},
			nil,
		},
		{
			"ShouldNotOverrideDisabledRedisSentinelIdleTimeout",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName:  "mysentinel",
				Addresses:   []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
				IdleTimeout: -1,
			}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, time.Duration(-1), have.RedisSentinel.IdleTimeout)
			},
			nil,
		},
		{
			"ShouldNotOverrideDisabledRedisClusterIdleTimeout",
			schema.Cache{RedisCluster: &schema.RedisClusterCache{
				Addresses:   []*schema.AddressTCP{mustAddressTCP("tcp://node1:6379")},
				IdleTimeout: -1,
			}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, time.Duration(-1), have.RedisCluster.IdleTimeout)
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
			"ShouldRaiseErrorOnRedisDatabaseNegative",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com:6379"), Database: -1}},
			nil,
			[]string{"cache: redis: option 'database' must be 0 or greater but it's configured as '-1'"},
		},
		{
			"ShouldAllowRedisDatabaseAboveTheDefaultServerLimit",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://redis.example.com:6379"), Database: 20}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, 20, have.Redis.Database)
			},
			nil,
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
				assert.Equal(t, 0, have.RedisSentinel.PoolSize)
				assert.Equal(t, time.Minute*30, have.RedisSentinel.IdleTimeout)
				assert.Equal(t, schema.DefaultRedisSentinelCacheConfiguration.ConnectionLifetime, have.RedisSentinel.ConnectionLifetime)
				assert.Equal(t, time.Second*4, have.RedisSentinel.PoolTimeout)
				assert.Equal(t, schema.DefaultRedisSentinelCacheConfiguration.MinimumRetryBackoff, have.RedisSentinel.MinimumRetryBackoff)
				assert.Equal(t, schema.DefaultRedisSentinelCacheConfiguration.MaximumRetryBackoff, have.RedisSentinel.MaximumRetryBackoff)
				assert.Equal(t, schema.DefaultRedisSentinelCacheConfiguration.PoolMaximumIdleConnections, have.RedisSentinel.PoolMaximumIdleConnections)
				assert.Equal(t, schema.DefaultRedisSentinelCacheConfiguration.PoolMaximumConnections, have.RedisSentinel.PoolMaximumConnections)
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
			"ShouldRaiseErrorOnRedisAddressWithoutHostname",
			schema.Cache{Redis: &schema.RedisCache{Address: mustAddressTCP("tcp://:6379")}},
			nil,
			[]string{"cache: redis: option 'address' must have a hostname but it's configured as 'tcp://:6379'"},
		},
		{
			"ShouldRaiseErrorOnRedisSentinelEmptyAddress",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName: "mysentinel",
				Addresses:  []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379"), nil},
			}},
			nil,
			[]string{"cache: redis_sentinel: option 'addresses' index 2 is invalid: the address is empty"},
		},
		{
			"ShouldRaiseErrorOnRedisSentinelDatabaseNegative",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName: "mysentinel",
				Addresses:  []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
				Database:   -1,
			}},
			nil,
			[]string{"cache: redis_sentinel: option 'database' must be 0 or greater but it's configured as '-1'"},
		},
		{
			"ShouldAllowRedisSentinelDatabaseAboveTheDefaultServerLimit",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName: "mysentinel",
				Addresses:  []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
				Database:   20,
			}},
			func(t *testing.T, have schema.Cache) {
				assert.Equal(t, 20, have.RedisSentinel.Database)
			},
			nil,
		},
		{
			"ShouldRaiseErrorOnRedisClusterEmptyAddress",
			schema.Cache{RedisCluster: &schema.RedisClusterCache{
				Addresses: []*schema.AddressTCP{mustAddressTCP("tcp://node1:6379"), nil},
			}},
			nil,
			[]string{"cache: redis_cluster: option 'addresses' index 2 is invalid: the address is empty"},
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
			"ShouldRaiseErrorOnRedisSentinelRouteByLatencyDefaultMode",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName:     "mysentinel",
				Addresses:      []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
				RouteByLatency: true,
			}},
			nil,
			[]string{"cache: redis_sentinel: option 'route_by_latency' requires option 'sentinel_mode' to be configured as 'cluster' but it's configured as 'failover'"},
		},
		{
			"ShouldRaiseErrorOnRedisSentinelRouteByLatencyFailoverMode",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName:     "mysentinel",
				Addresses:      []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
				SentinelMode:   "failover",
				RouteByLatency: true,
			}},
			nil,
			[]string{"cache: redis_sentinel: option 'route_by_latency' requires option 'sentinel_mode' to be configured as 'cluster' but it's configured as 'failover'"},
		},
		{
			"ShouldRaiseErrorOnRedisSentinelRouteRandomlyDefaultMode",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName:    "mysentinel",
				Addresses:     []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
				RouteRandomly: true,
			}},
			nil,
			[]string{"cache: redis_sentinel: option 'route_randomly' requires option 'sentinel_mode' to be configured as 'cluster' but it's configured as 'failover'"},
		},
		{
			"ShouldRaiseErrorOnRedisSentinelRouteRandomlyFailoverMode",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName:    "mysentinel",
				Addresses:     []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
				SentinelMode:  "failover",
				RouteRandomly: true,
			}},
			nil,
			[]string{"cache: redis_sentinel: option 'route_randomly' requires option 'sentinel_mode' to be configured as 'cluster' but it's configured as 'failover'"},
		},
		{
			"ShouldAllowRedisSentinelRoutingInClusterMode",
			schema.Cache{RedisSentinel: &schema.RedisSentinelCache{
				MasterName:     "mysentinel",
				Addresses:      []*schema.AddressTCP{mustAddressTCP("tcp://sentinel1:26379")},
				SentinelMode:   "cluster",
				RouteByLatency: true,
				RouteRandomly:  true,
			}},
			nil,
			nil,
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
				assert.Equal(t, time.Minute*30, have.RedisCluster.IdleTimeout)
				assert.Equal(t, schema.DefaultRedisClusterCacheConfiguration.ConnectionLifetime, have.RedisCluster.ConnectionLifetime)
				assert.Equal(t, time.Second*4, have.RedisCluster.PoolTimeout)
				assert.Equal(t, schema.DefaultRedisClusterCacheConfiguration.MinimumRetryBackoff, have.RedisCluster.MinimumRetryBackoff)
				assert.Equal(t, schema.DefaultRedisClusterCacheConfiguration.MaximumRetryBackoff, have.RedisCluster.MaximumRetryBackoff)
				assert.Equal(t, schema.DefaultRedisClusterCacheConfiguration.PoolMaximumIdleConnections, have.RedisCluster.PoolMaximumIdleConnections)
				assert.Equal(t, schema.DefaultRedisClusterCacheConfiguration.PoolMaximumConnections, have.RedisCluster.PoolMaximumConnections)
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
