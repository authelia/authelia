package session

import (
	"time"

	"github.com/valyala/fasthttp"

	"github.com/clems4ever/authelia/configuration/schema"
	"github.com/fasthttp/session"
	"github.com/fasthttp/session/memory"
	"github.com/fasthttp/session/redis"
)

// NewProviderConfig creates a configuration for creating the session provider
func NewProviderConfig(configuration schema.SessionConfiguration) ProviderConfig {
	config := session.NewDefaultConfig()

	// Override the cookie name.
	config.CookieName = configuration.Name

	// Set the cookie to the given domain.
	config.Domain = configuration.Domain

	// Only serve the header over HTTPS.
	config.Secure = true

	if configuration.Expiration > 0 {
		config.Expires = time.Duration(configuration.Expiration) * time.Second
	} else {
		// If Expiration is 0 then cookie expiration is disabled.
		config.Expires = 0
	}

	// TODO(c.michaud): Make this configurable by giving the list of IPs that are trustable.
	config.IsSecureFunc = func(*fasthttp.RequestCtx) bool {
		return true
	}

	var providerConfig session.ProviderConfig
	var providerName string

	// If redis configuration is provided, then use the redis provider.
	if configuration.Redis != nil {
		providerName = "redis"
		providerConfig = &redis.Config{
			Host:        configuration.Redis.Host,
			Port:        configuration.Redis.Port,
			Password:    configuration.Redis.Password,
			PoolSize:    8,
			IdleTimeout: 300,
			KeyPrefix:   "authelia-session",
		}
	} else { // if no option is provided, use the memory provider.
		providerName = "memory"
		providerConfig = &memory.Config{}
	}
	return ProviderConfig{
		config:         config,
		providerName:   providerName,
		providerConfig: providerConfig,
	}
}
