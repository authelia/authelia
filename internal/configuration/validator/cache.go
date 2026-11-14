// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidateCache validates and updates the cache configuration.
func ValidateCache(config *schema.Configuration, validator *schema.StructValidator) {
	var configured []string

	if config.Cache.Redis != nil {
		configured = append(configured, "redis")
	}

	if config.Cache.RedisSentinel != nil {
		configured = append(configured, "redis_sentinel")
	}

	if config.Cache.RedisCluster != nil {
		configured = append(configured, "redis_cluster")
	}

	if len(configured) > 1 {
		validator.Push(fmt.Errorf(errFmtCacheMultipleProviders, utils.StringJoinAnd(configured)))

		return
	}

	switch {
	case config.Cache.Redis != nil:
		validateCacheRedis(config.Cache.Redis, validator)
	case config.Cache.RedisSentinel != nil:
		validateCacheRedisSentinel(config.Cache.RedisSentinel, validator)
	case config.Cache.RedisCluster != nil:
		validateCacheRedisCluster(config.Cache.RedisCluster, validator)
	}
}

func validateCacheRedis(config *schema.RedisCache, validator *schema.StructValidator) {
	if config.Address == nil {
		validator.Push(fmt.Errorf(errFmtCacheOptionRequired, "redis", "address"))
	} else {
		validateCacheRedisAddress("redis", "address", config.Address, validator)
	}

	if config.Database < 0 {
		validator.Push(fmt.Errorf(errFmtCacheRedisDatabase, "redis", config.Database))
	}

	validateCacheRedisTLS("redis", config.TLS, cacheRedisServerName(config.Address), validator)

	validateCacheRedisCommon("redis", &config.RedisCacheCommon, validator)
}

func validateCacheRedisSentinelRouting(config *schema.RedisSentinelCache, validator *schema.StructValidator) {
	if config.SentinelMode != schema.RedisSentinelModeFailover {
		return
	}

	if config.RouteByLatency {
		validator.Push(fmt.Errorf(errFmtCacheRedisSentinelModeRouting, "route_by_latency", config.SentinelMode))
	}

	if config.RouteRandomly {
		validator.Push(fmt.Errorf(errFmtCacheRedisSentinelModeRouting, "route_randomly", config.SentinelMode))
	}
}

func validateCacheRedisSentinel(config *schema.RedisSentinelCache, validator *schema.StructValidator) {
	if config.MasterName == "" {
		validator.Push(fmt.Errorf(errFmtCacheOptionRequired, "redis_sentinel", "master_name"))
	}

	if len(config.Addresses) == 0 {
		validator.Push(fmt.Errorf(errFmtCacheOptionRequired, "redis_sentinel", "addresses"))
	}

	for i, address := range config.Addresses {
		if address == nil {
			validator.Push(fmt.Errorf(errFmtCacheRedisSentinelAddressEmpty, i+1))

			continue
		}

		validateCacheRedisAddress("redis_sentinel", fmt.Sprintf("addresses[%d]", i), address, validator)
	}

	switch config.SentinelMode {
	case "":
		config.SentinelMode = schema.DefaultRedisSentinelCacheConfiguration.SentinelMode
	case schema.RedisSentinelModeFailover, schema.RedisSentinelModeCluster:
		break
	default:
		validator.Push(fmt.Errorf(errFmtCacheRedisSentinelMode, utils.StringJoinOr(validCacheRedisSentinelModes), config.SentinelMode))
	}

	validateCacheRedisSentinelRouting(config, validator)

	if config.Database < 0 {
		validator.Push(fmt.Errorf(errFmtCacheRedisDatabase, "redis_sentinel", config.Database))
	}

	validateCacheRedisTLS("redis_sentinel", config.TLS, "", validator)

	validateCacheRedisCommon("redis_sentinel", &config.RedisCacheCommon, validator)
}

func validateCacheRedisCluster(config *schema.RedisClusterCache, validator *schema.StructValidator) {
	if len(config.Addresses) == 0 {
		validator.Push(fmt.Errorf(errFmtCacheOptionRequired, "redis_cluster", "addresses"))
	}

	for i, address := range config.Addresses {
		if address == nil {
			validator.Push(fmt.Errorf(errFmtCacheRedisClusterAddressEmpty, i+1))

			continue
		}

		validateCacheRedisAddress("redis_cluster", fmt.Sprintf("addresses[%d]", i), address, validator)
	}

	validateCacheRedisTLS("redis_cluster", config.TLS, "", validator)

	setDefault(&config.MaximumRedirects, schema.DefaultRedisClusterCacheConfiguration.MaximumRedirects)

	validateCacheRedisCommon("redis_cluster", &config.RedisCacheCommon, validator)
}

func validateCacheRedisCommon(provider string, config *schema.RedisCacheCommon, validator *schema.StructValidator) {
	defaults := schema.DefaultRedisCacheCommonConfiguration

	setDefault(&config.DialTimeout, defaults.DialTimeout)
	setDefault(&config.ReadTimeout, defaults.ReadTimeout)
	setDefault(&config.WriteTimeout, defaults.WriteTimeout)
	setDefault(&config.PoolTimeout, config.ReadTimeout+time.Second)
	setDefault(&config.MinimumRetryBackoff, defaults.MinimumRetryBackoff)
	setDefault(&config.MaximumRetryBackoff, defaults.MaximumRetryBackoff)
	setDefault(&config.MaximumRetries, defaults.MaximumRetries)
	setDefault(&config.PoolMaximumIdleConnections, defaults.PoolMaximumIdleConnections)
	setDefault(&config.DialerRetries, defaults.DialerRetries)
	setDefault(&config.DialerRetryTimeout, defaults.DialerRetryTimeout)
	setDefault(&config.FailingTimeout, defaults.FailingTimeout)
	setDefault(&config.ReadBufferSize, defaults.ReadBufferSize)
	setDefault(&config.WriteBufferSize, defaults.WriteBufferSize)

	setDefaultUnset(&config.IdleTimeout, defaults.IdleTimeout)

	config.PoolSize = max(config.PoolSize, 0)
	config.PoolMaximumConnections = max(config.PoolMaximumConnections, 0)

	validateCacheRedisConnectionLifetimeJitter(provider, config.ConnectionLifetime, config.ConnectionLifetimeJitter, validator)
}

func validateCacheRedisAddress(provider, option string, address *schema.AddressTCP, validator *schema.StructValidator) {
	switch {
	case address.IsUnixDomainSocket():
		return
	case address.Hostname() == "":
		validator.Push(fmt.Errorf(errFmtCacheRedisAddressNoHost, provider, option, address.String()))
	case address.Port() == 0:
		validator.Push(fmt.Errorf(errFmtCacheRedisAddressNoPort, provider, option, address.String()))
	}
}

func validateCacheRedisTLS(provider string, config *schema.TLS, serverName string, validator *schema.StructValidator) {
	if config == nil {
		return
	}

	defaults := &schema.TLS{
		ServerName:     serverName,
		MinimumVersion: schema.DefaultRedisCacheConfiguration.TLS.MinimumVersion,
		MaximumVersion: schema.DefaultRedisCacheConfiguration.TLS.MaximumVersion,
	}

	if err := ValidateTLSConfig(config, defaults); err != nil {
		validator.Push(fmt.Errorf(errFmtCacheRedisTLSConfigInvalid, provider, err))
	}
}

func validateCacheRedisConnectionLifetimeJitter(provider string, lifetime, jitter time.Duration, validator *schema.StructValidator) {
	if jitter > lifetime {
		validator.Push(fmt.Errorf(errFmtCacheRedisConnectionLifetimeJitter, provider, jitter, lifetime))
	}
}

func cacheRedisServerName(address *schema.AddressTCP) (serverName string) {
	if address == nil {
		return ""
	}

	return address.Hostname()
}

func setDefault[T int | time.Duration](value *T, def T) {
	if *value <= 0 {
		*value = def
	}
}

func setDefaultUnset(value *time.Duration, def time.Duration) {
	if *value == 0 {
		*value = def
	}
}
