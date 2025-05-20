package cache

import (
	"context"
	"crypto/x509"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func NewRedisCluster(config *schema.RedisClusterCache, rootCAs *x509.CertPool) *RedisCluster {
	return &RedisCluster{
		config:  config,
		rootCAs: rootCAs,
	}
}

type RedisCluster struct {
	client  redis.Cmdable
	config  *schema.RedisClusterCache
	rootCAs *x509.CertPool
}

func (r *RedisCluster) StartupCheck() (err error) {
	addreses := make([]string, len(r.config.Addresses))

	for i, address := range r.config.Addresses {
		addreses[i] = address.NetworkAddress()
	}

	options := &redis.ClusterOptions{
		Addrs:          addreses,
		ClientName:     fmt.Sprintf(driverParameterFmtAppName, utils.Version()),
		MaxRedirects:   r.config.MaximumRedirects,
		ReadOnly:       r.config.RouteByReplica,
		RouteByLatency: r.config.RouteByLatency,
		RouteRandomly:  r.config.RouteRandomly,
		Protocol:       3,
		Username:       r.config.Username,
		Password:       r.config.Password,
		CredentialsProviderContext: func(ctx context.Context) (username string, password string, err error) {
			return r.config.Username, r.config.Password, nil
		},
		MaxRetries:      r.config.MaximumRedirects,
		MinRetryBackoff: r.config.MinimumRetryBackoff,
		MaxRetryBackoff: r.config.MaximumRetryBackoff,
		DialTimeout:     r.config.DialTimeout,
		ReadTimeout:     r.config.ReadTimeout,
		WriteTimeout:    r.config.WriteTimeout,
		PoolSize:        r.config.PoolSize,
		PoolTimeout:     r.config.PoolTimeout,
		MinIdleConns:    r.config.PoolMinimumIdleConnections,
		MaxIdleConns:    r.config.PoolMaximumIdleConnections,
		MaxActiveConns:  r.config.PoolMaximumConnections,
		ConnMaxIdleTime: r.config.IdleTimeout,
		ConnMaxLifetime: r.config.ConnectionTimeout,
		TLSConfig:       utils.NewTLSConfig(r.config.TLS, r.rootCAs),
	}

	r.client = redis.NewClusterClient(options)

	return r.client.Ping(context.Background()).Err()
}
