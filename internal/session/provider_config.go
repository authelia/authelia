package session

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/redis"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewProviderConfig creates a configuration for creating the session provider.
func NewProviderConfig(configuration schema.SessionConfiguration, certPool *x509.CertPool) ProviderConfig {
	config := session.NewDefaultConfig()

	// Override the cookie name.
	config.CookieName = configuration.Name

	// Set the cookie to the given domain.
	config.Domain = configuration.Domain

	// Set the cookie SameSite option.
	switch configuration.SameSite {
	case "strict":
		config.CookieSameSite = fasthttp.CookieSameSiteStrictMode
	case "none":
		config.CookieSameSite = fasthttp.CookieSameSiteNoneMode
	case "lax":
		config.CookieSameSite = fasthttp.CookieSameSiteLaxMode
	default:
		config.CookieSameSite = fasthttp.CookieSameSiteLaxMode
	}

	// Only serve the header over HTTPS.
	config.Secure = true

	// Ignore the error as it will be handled by validator.
	config.Expiration, _ = utils.ParseDurationString(configuration.Expiration)

	// TODO(c.michaud): Make this configurable by giving the list of IPs that are trustable.
	config.IsSecureFunc = func(*fasthttp.RequestCtx) bool {
		return true
	}

	var redisConfig *redis.Config

	var redisSentinelConfig *redis.FailoverConfig

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
			addrs := make([]string, 0)

			if configuration.Redis.Host != "" {
				addrs = append(addrs, fmt.Sprintf("%s:%d", strings.ToLower(configuration.Redis.Host), configuration.Redis.Port))
			}

			for _, node := range configuration.Redis.HighAvailability.Nodes {
				addr := fmt.Sprintf("%s:%d", strings.ToLower(node.Host), node.Port)
				if !utils.IsStringInSlice(addr, addrs) {
					addrs = append(addrs, addr)
				}
			}

			providerName = "redis-sentinel"
			redisSentinelConfig = &redis.FailoverConfig{
				Logger:           &redisLogger{logger: logging.Logger()},
				MasterName:       configuration.Redis.HighAvailability.SentinelName,
				SentinelAddrs:    addrs,
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
				Logger:       newRedisLogger(),
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
