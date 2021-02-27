package session

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"

	"github.com/authelia/session/v2"
	"github.com/authelia/session/v2/providers/redis"
	"github.com/authelia/session/v2/providers/redisfailover"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// NewProviderConfig creates a configuration for creating the session provider.
func NewProviderConfig(configuration schema.SessionConfiguration, certPool *x509.CertPool) ProviderConfig {
	config := session.NewDefaultConfig()

	// Override the cookie name.
	config.CookieName = configuration.Name

	// Set the cookie to the given domain.
	config.Domain = configuration.Domain

	// Only serve the header over HTTPS.
	config.Secure = true

	// Ignore the error as it will be handled by validator.
	config.Expiration, _ = utils.ParseDurationString(configuration.Expiration)

	// TODO(c.michaud): Make this configurable by giving the list of IPs that are trustable.
	config.IsSecureFunc = func(*fasthttp.RequestCtx) bool {
		return true
	}

	var redisConfig *redis.Config

	var redisSentinelConfig *redisfailover.Config

	var providerName string

	// If redis configuration is provided, then use the redis provider.
	switch {
	case configuration.Redis != nil:
		serializer := NewEncryptingSerializer(configuration.Secret)

		var tlsConfig *tls.Config

		if configuration.Redis.TLS != nil {
			tlsConfig = utils.NewTLSConfig(configuration.Redis.TLS, tls.VersionTLS12, certPool)
		}

		if configuration.Redis.HighAvailability != nil && configuration.Redis.HighAvailability.SentinelName != "" {
			nodes := make([]string, 0)

			if configuration.Redis.Host != "" {
				nodes = []string{fmt.Sprintf("%s:%d", configuration.Redis.Host, configuration.Redis.Port)}
			}

			for _, addr := range configuration.Redis.HighAvailability.Nodes {
				nodes = append(nodes, fmt.Sprintf("%s:%d", addr.Host, addr.Port))
			}

			providerName = "redis-sentinel"
			redisSentinelConfig = &redisfailover.Config{
				MasterName:       configuration.Redis.HighAvailability.SentinelName,
				SentinelAddrs:    nodes,
				SentinelPassword: configuration.Redis.HighAvailability.SentinelPassword,
				RouteByLatency:   configuration.Redis.HighAvailability.RouteByLatency,
				RouteRandomly:    configuration.Redis.HighAvailability.RouteRandomly,
				Username:         configuration.Redis.Username,
				Password:         configuration.Redis.Password,
				DB:               configuration.Redis.DatabaseIndex, // DB is the fasthttp/session property for the Redis DB Index.
				PoolSize:         configuration.Redis.MaximumActiveConnections,
				MinIdleConns:     configuration.Redis.MinimumIdleConnections,
				IdleTimeout:      300,
				TLSConfig:        tlsConfig,
				KeyPrefix:        "authelia-session",
			}
		} else {
			providerName = "redis"
			network := "tcp"

			var addr string

			if configuration.Redis.Port == 0 {
				network = "unix"
				addr = configuration.Redis.Host
			} else {
				addr = fmt.Sprintf("%s:%d", configuration.Redis.Host, configuration.Redis.Port)
			}

			redisConfig = &redis.Config{
				Network:      network,
				Addr:         addr,
				Username:     configuration.Redis.Username,
				Password:     configuration.Redis.Password,
				DB:           configuration.Redis.DatabaseIndex, // DB is the fasthttp/session property for the Redis DB Index.
				PoolSize:     configuration.Redis.MaximumActiveConnections,
				MinIdleConns: configuration.Redis.MinimumIdleConnections,
				IdleTimeout:  300,
				TLSConfig:    tlsConfig,
				KeyPrefix:    "authelia-session",
			}
		}

		config.EncodeFunc = serializer.Encode
		config.DecodeFunc = serializer.Decode
	default:
		providerName = "memory"
	}

	return ProviderConfig{
		config,
		redisConfig,
		redisSentinelConfig,
		providerName,
	}
}
