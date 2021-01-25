package session

import (
	"fmt"

	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/redis"

	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// NewProviderConfig creates a configuration for creating the session provider.
func NewProviderConfig(configuration schema.SessionConfiguration) ProviderConfig {
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

	var providerName string

	// If redis configuration is provided, then use the redis provider.
	if configuration.Redis != nil {
		providerName = "redis"
		serializer := NewEncryptingSerializer(configuration.Secret)
		network := "tcp"

		var addr string

		if configuration.Redis.Port == 0 {
			network = "unix"
			addr = configuration.Redis.Host
		} else {
			addr = fmt.Sprintf("%s:%d", configuration.Redis.Host, configuration.Redis.Port)
		}

		redisConfig = &redis.Config{
			Network:  network,
			Addr:     addr,
			Password: configuration.Redis.Password,
			// DB is the fasthttp/session property for the Redis DB Index.
			DB:          configuration.Redis.DatabaseIndex,
			PoolSize:    8,
			IdleTimeout: 300,
			KeyPrefix:   "authelia-session",
		}
		config.EncodeFunc = serializer.Encode
		config.DecodeFunc = serializer.Decode
	} else { // if no option is provided, use the memory provider.
		providerName = "memory"
	}

	return ProviderConfig{
		config:       config,
		redisConfig:  redisConfig,
		providerName: providerName,
	}
}
