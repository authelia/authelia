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

	if config.Database < 0 || config.Database > 15 {
		validator.Push(fmt.Errorf(errFmtCacheRedisDatabase, "redis", config.Database))
	}

	validateCacheRedisTLS("redis", config.TLS, cacheRedisServerName(config.Address), validator)

	setDefault(&config.DialTimeout, schema.DefaultRedisCacheConfiguration.DialTimeout)
	setDefault(&config.ReadTimeout, schema.DefaultRedisCacheConfiguration.ReadTimeout)
	setDefault(&config.WriteTimeout, schema.DefaultRedisCacheConfiguration.WriteTimeout)

	if config.MaximumRetries <= 0 {
		config.MaximumRetries = schema.DefaultRedisCacheConfiguration.MaximumRetries
	}

	if config.PoolSize <= 0 {
		config.PoolSize = schema.DefaultRedisCacheConfiguration.PoolSize
	}

	setDefault(&config.DialerRetryTimeout, schema.DefaultRedisCacheConfiguration.DialerRetryTimeout)
	setDefault(&config.FailingTimeout, schema.DefaultRedisCacheConfiguration.FailingTimeout)

	if config.DialerRetries <= 0 {
		config.DialerRetries = schema.DefaultRedisCacheConfiguration.DialerRetries
	}

	if config.ReadBufferSize <= 0 {
		config.ReadBufferSize = schema.DefaultRedisCacheConfiguration.ReadBufferSize
	}

	if config.WriteBufferSize <= 0 {
		config.WriteBufferSize = schema.DefaultRedisCacheConfiguration.WriteBufferSize
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
	case "failover", "cluster":
		break
	default:
		validator.Push(fmt.Errorf(errFmtCacheRedisSentinelMode, utils.StringJoinOr(validCacheRedisSentinelModes), config.SentinelMode))
	}

	if config.Database < 0 || config.Database > 15 {
		validator.Push(fmt.Errorf(errFmtCacheRedisDatabase, "redis_sentinel", config.Database))
	}

	var serverName string

	if len(config.Addresses) != 0 {
		serverName = cacheRedisServerName(config.Addresses[0])
	}

	validateCacheRedisTLS("redis_sentinel", config.TLS, serverName, validator)

	setDefault(&config.DialTimeout, schema.DefaultRedisSentinelCacheConfiguration.DialTimeout)
	setDefault(&config.ReadTimeout, schema.DefaultRedisSentinelCacheConfiguration.ReadTimeout)
	setDefault(&config.WriteTimeout, schema.DefaultRedisSentinelCacheConfiguration.WriteTimeout)

	if config.MaximumRetries <= 0 {
		config.MaximumRetries = schema.DefaultRedisSentinelCacheConfiguration.MaximumRetries
	}

	if config.PoolSize <= 0 {
		config.PoolSize = schema.DefaultRedisSentinelCacheConfiguration.PoolSize
	}

	setDefault(&config.DialerRetryTimeout, schema.DefaultRedisSentinelCacheConfiguration.DialerRetryTimeout)
	setDefault(&config.FailingTimeout, schema.DefaultRedisSentinelCacheConfiguration.FailingTimeout)

	if config.DialerRetries <= 0 {
		config.DialerRetries = schema.DefaultRedisSentinelCacheConfiguration.DialerRetries
	}

	if config.ReadBufferSize <= 0 {
		config.ReadBufferSize = schema.DefaultRedisSentinelCacheConfiguration.ReadBufferSize
	}

	if config.WriteBufferSize <= 0 {
		config.WriteBufferSize = schema.DefaultRedisSentinelCacheConfiguration.WriteBufferSize
	}
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

	var serverName string

	if len(config.Addresses) != 0 {
		serverName = cacheRedisServerName(config.Addresses[0])
	}

	validateCacheRedisTLS("redis_cluster", config.TLS, serverName, validator)

	setDefault(&config.DialTimeout, schema.DefaultRedisClusterCacheConfiguration.DialTimeout)
	setDefault(&config.ReadTimeout, schema.DefaultRedisClusterCacheConfiguration.ReadTimeout)
	setDefault(&config.WriteTimeout, schema.DefaultRedisClusterCacheConfiguration.WriteTimeout)

	if config.MaximumRetries <= 0 {
		config.MaximumRetries = schema.DefaultRedisClusterCacheConfiguration.MaximumRetries
	}

	if config.MaximumRedirects <= 0 {
		config.MaximumRedirects = schema.DefaultRedisClusterCacheConfiguration.MaximumRedirects
	}

	if config.PoolSize <= 0 {
		config.PoolSize = schema.DefaultRedisClusterCacheConfiguration.PoolSize
	}

	setDefault(&config.DialerRetryTimeout, schema.DefaultRedisClusterCacheConfiguration.DialerRetryTimeout)
	setDefault(&config.FailingTimeout, schema.DefaultRedisClusterCacheConfiguration.FailingTimeout)

	if config.DialerRetries <= 0 {
		config.DialerRetries = schema.DefaultRedisClusterCacheConfiguration.DialerRetries
	}

	if config.ReadBufferSize <= 0 {
		config.ReadBufferSize = schema.DefaultRedisClusterCacheConfiguration.ReadBufferSize
	}

	if config.WriteBufferSize <= 0 {
		config.WriteBufferSize = schema.DefaultRedisClusterCacheConfiguration.WriteBufferSize
	}
}

// validateCacheRedisAddress ensures the address is usable. A TCP address requires a port as there is no port which can
// be assumed on a users behalf once they've specified the host themselves.
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

func cacheRedisServerName(address *schema.AddressTCP) (serverName string) {
	if address == nil {
		return ""
	}

	return address.Hostname()
}

func setDefault(value *time.Duration, def time.Duration) {
	if *value <= 0 {
		*value = def
	}
}
