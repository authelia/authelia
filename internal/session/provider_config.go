package session

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"strings"

	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memory"
	"github.com/fasthttp/session/v2/providers/redis"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/storage"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewProviderConfig creates a configuration for creating the session provider.
func NewProviderConfig(config schema.SessionCookie, providerName string, serializer Serializer) ProviderConfig {
	c := session.NewDefaultConfig()

	c.SessionIDGeneratorFunc = func() []byte {
		bytes := make([]byte, 32)

		_, _ = rand.Read(bytes)

		for i, b := range bytes {
			bytes[i] = randomSessionChars[b%byte(len(randomSessionChars))]
		}

		return bytes
	}

	// Override the cookie name.
	c.CookieName = config.Name

	// Set the cookie to the given domain.
	c.Domain = config.Domain

	// Set the cookie SameSite option.
	switch config.SameSite {
	case "strict":
		c.CookieSameSite = fasthttp.CookieSameSiteStrictMode
	case "none":
		c.CookieSameSite = fasthttp.CookieSameSiteNoneMode
	case "lax":
		c.CookieSameSite = fasthttp.CookieSameSiteLaxMode
	default:
		c.CookieSameSite = fasthttp.CookieSameSiteLaxMode
	}

	// Only serve the header over HTTPS.
	c.Secure = true

	// Ignore the error as it will be handled by validator.
	c.Expiration = config.Expiration

	c.IsSecureFunc = func(*fasthttp.RequestCtx) bool {
		return true
	}

	if serializer != nil {
		c.EncodeFunc = serializer.Encode
		c.DecodeFunc = serializer.Decode
	}

	return ProviderConfig{
		c,
		providerName,
	}
}

func NewProviderSession(pconfig ProviderConfig, provider session.Provider) (p *session.Session, err error) {
	p = session.New(pconfig.config)

	if err = p.SetProvider(provider); err != nil {
		return nil, err
	}

	return p, nil
}

func NewProviderConfigAndSession(config schema.SessionCookie, providerName string, serializer Serializer, provider session.Provider) (c ProviderConfig, p *session.Session, err error) {
	c = NewProviderConfig(config, providerName, serializer)

	if p, err = NewProviderSession(c, provider); err != nil {
		return c, nil, err
	}

	return c, p, nil
}

func NewSessionProvider(config schema.Session, certPool *x509.CertPool, storageProvider storage.SessionProvider) (name string, provider session.Provider, serializer Serializer, err error) {
	// If redis configuration is provided, then use the redis provider.
	switch {
	case config.Redis != nil:
		serializer = NewEncryptingSerializer(config.Secret)

		var tlsConfig *tls.Config

		if config.Redis.TLS != nil {
			tlsConfig = utils.NewTLSConfig(config.Redis.TLS, certPool)
		}

		if config.Redis.HighAvailability != nil && config.Redis.HighAvailability.SentinelName != "" {
			addrs := make([]string, 0)

			if config.Redis.Host != "" {
				addrs = append(addrs, fmt.Sprintf("%s:%d", strings.ToLower(config.Redis.Host), config.Redis.Port))
			}

			for _, node := range config.Redis.HighAvailability.Nodes {
				addr := fmt.Sprintf("%s:%d", strings.ToLower(node.Host), node.Port)
				if !utils.IsStringInSlice(addr, addrs) {
					addrs = append(addrs, addr)
				}
			}

			name = "redis-sentinel"

			provider, err = redis.NewFailover(redis.FailoverConfig{
				Logger:           logging.LoggerCtxPrintf(logrus.TraceLevel),
				MasterName:       config.Redis.HighAvailability.SentinelName,
				SentinelAddrs:    addrs,
				DialTimeout:      config.Redis.Timeout,
				MaxRetries:       config.Redis.MaxRetries,
				SentinelUsername: config.Redis.HighAvailability.SentinelUsername,
				SentinelPassword: config.Redis.HighAvailability.SentinelPassword,
				RouteByLatency:   config.Redis.HighAvailability.RouteByLatency,
				RouteRandomly:    config.Redis.HighAvailability.RouteRandomly,
				Username:         config.Redis.Username,
				Password:         config.Redis.Password,
				DB:               config.Redis.DatabaseIndex, // DB is the fasthttp/session property for the Redis DB Index.
				PoolSize:         config.Redis.MaximumActiveConnections,
				MinIdleConns:     config.Redis.MinimumIdleConnections,
				ConnMaxIdleTime:  300,
				TLSConfig:        tlsConfig,
				KeyPrefix:        "authelia-session",
			})
		} else {
			name = "redis"
			network := "tcp"

			var addr string

			if config.Redis.Port == 0 {
				network = "unix"
				addr = config.Redis.Host
			} else {
				addr = fmt.Sprintf("%s:%d", config.Redis.Host, config.Redis.Port)
			}

			provider, err = redis.New(redis.Config{
				Logger:          logging.LoggerCtxPrintf(logrus.TraceLevel),
				Network:         network,
				Addr:            addr,
				DialTimeout:     config.Redis.Timeout,
				MaxRetries:      config.Redis.MaxRetries,
				Username:        config.Redis.Username,
				Password:        config.Redis.Password,
				DB:              config.Redis.DatabaseIndex, // DB is the fasthttp/session property for the Redis DB Index.
				PoolSize:        config.Redis.MaximumActiveConnections,
				MinIdleConns:    config.Redis.MinimumIdleConnections,
				ConnMaxIdleTime: 300,
				TLSConfig:       tlsConfig,
				KeyPrefix:       "authelia-session",
			})
		}
	case storageProvider != nil:
		name = "sql"
		serializer = NewEncryptingSerializer(config.Secret)
		provider = NewSQLSessionProvider(storageProvider)
	default:
		name = "memory"
		provider, err = memory.New(memory.Config{})
	}

	return name, provider, serializer, err
}
