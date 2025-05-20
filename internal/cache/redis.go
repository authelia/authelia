package cache

import (
	"context"
	"crypto/x509"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func NewRedis(config *schema.RedisCache, rootCAs *x509.CertPool) *Redis {
	return &Redis{
		config:  config,
		rootCAs: rootCAs,
	}
}

type Redis struct {
	client  redis.Cmdable
	config  *schema.RedisCache
	rootCAs *x509.CertPool
}

func (r *Redis) StartupCheck() (err error) {
	options := &redis.Options{
		Network:    r.config.Address.Network(),
		Addr:       r.config.Address.NetworkAddress(),
		ClientName: fmt.Sprintf(driverParameterFmtAppName, utils.Version()),
		Protocol:   3,
		Username:   r.config.Username,
		Password:   r.config.Password,
		CredentialsProviderContext: func(ctx context.Context) (username string, password string, err error) {
			return r.config.Username, r.config.Password, nil
		},
		DB:              r.config.Database,
		MaxRetries:      r.config.MaximumRetries,
		MinRetryBackoff: r.config.MinimumRetryBackoff,
		MaxRetryBackoff: r.config.MaximumRetryBackoff,
		DialTimeout:     r.config.DialTimeout,
		ReadTimeout:     r.config.ReadTimeout,
		WriteTimeout:    r.config.WriteTimeout,
		ConnMaxIdleTime: r.config.IdleTimeout,
		ConnMaxLifetime: r.config.ConnectionTimeout,
		PoolTimeout:     r.config.PoolTimeout,
		PoolSize:        r.config.PoolSize,
		MinIdleConns:    r.config.PoolMinimumIdleConnections,
		MaxIdleConns:    r.config.PoolMaximumIdleConnections,
		MaxActiveConns:  r.config.PoolMaximumConnections,
		TLSConfig:       utils.NewTLSConfig(r.config.TLS, r.rootCAs),
	}

	r.client = redis.NewClient(options)

	return r.client.Ping(context.Background()).Err()
}
