package cache

import (
	"context"
	"crypto/x509"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func NewRedisSentinel(config *schema.RedisSentinelCache, rootCAs *x509.CertPool) *RedisSentinel {
	return &RedisSentinel{
		config:  config,
		rootCAs: rootCAs,
	}
}

type RedisSentinel struct {
	client  redis.Cmdable
	config  *schema.RedisSentinelCache
	rootCAs *x509.CertPool
}

func (r *RedisSentinel) StartupCheck() (err error) {
	addreses := make([]string, len(r.config.Addresses))

	for i, address := range r.config.Addresses {
		addreses[i] = address.NetworkAddress()
	}

	options := &redis.FailoverOptions{
		MasterName:       r.config.MasterName,
		SentinelAddrs:    addreses,
		ClientName:       fmt.Sprintf(driverParameterFmtAppName, utils.Version()),
		SentinelUsername: r.config.SentinelUsername,
		SentinelPassword: r.config.SentinelPassword,
		RouteByLatency:   r.config.RouteByLatency,
		RouteRandomly:    r.config.RouteRandomly,
		Protocol:         3,
		Username:         r.config.Username,
		Password:         r.config.Password,
		DB:               r.config.Database,
		MaxRetries:       r.config.MaximumRetries,
		MinRetryBackoff:  r.config.MinimumRetryBackoff,
		MaxRetryBackoff:  r.config.MaximumRetryBackoff,
		DialTimeout:      r.config.DialTimeout,
		ReadTimeout:      r.config.ReadTimeout,
		WriteTimeout:     r.config.WriteTimeout,
		PoolSize:         r.config.PoolSize,
		PoolTimeout:      r.config.PoolTimeout,
		MinIdleConns:     r.config.PoolMinimumIdleConnections,
		MaxIdleConns:     r.config.PoolMaximumIdleConnections,
		MaxActiveConns:   r.config.PoolMaximumConnections,
		ConnMaxIdleTime:  r.config.IdleTimeout,
		ConnMaxLifetime:  r.config.ConnectionTimeout,
		TLSConfig:        utils.NewTLSConfig(r.config.TLS, r.rootCAs),
	}

	if r.config.RouteByReplica {
		r.client = redis.NewFailoverClusterClient(options)
	} else {
		r.client = redis.NewFailoverClient(options)
	}

	return r.client.Ping(context.Background()).Err()
}
