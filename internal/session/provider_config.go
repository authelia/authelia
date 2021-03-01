package session

import (
	"fmt"

	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/memcache"
	"github.com/fasthttp/session/v2/providers/mysql"
	"github.com/fasthttp/session/v2/providers/postgre"
	"github.com/fasthttp/session/v2/providers/redis"
	"github.com/fasthttp/session/v2/providers/sqlite3"
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

	var (
		redisConfig    *redis.Config
		memcacheConfig *memcache.Config
		mysqlConfig    *mysql.Config
		postgreConfig  *postgre.Config
		sqlite3Config  *sqlite3.Config
	)

	var providerName string

	switch {
	case configuration.Redis != nil: // If redis configuration is provided, then use the redis provider.
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
			Username: configuration.Redis.Username,
			Password: configuration.Redis.Password,
			// DB is the fasthttp/session property for the Redis DB Index.
			DB:          configuration.Redis.DatabaseIndex,
			PoolSize:    8,
			IdleTimeout: 300,
			KeyPrefix:   "authelia-session",
		}
		config.EncodeFunc = serializer.Encode
		config.DecodeFunc = serializer.Decode
	case len(configuration.Memcache) > 0: // If memcache configuration is provided, then use the memcache provider.
		providerName = "memcache"
		serializer := NewEncryptingSerializer(configuration.Secret)

		var serverList []string

		for _, server := range configuration.Memcache {
			if server.Port == 0 {
				serverList = append(serverList, server.Host)
			} else {
				serverList = append(serverList, fmt.Sprintf("%s:%d", server.Host, server.Port))
			}
		}

		memcacheConfig = &memcache.Config{
			ServerList:   serverList,
			MaxIdleConns: 8,
			KeyPrefix:    "authelia-session",
		}
		config.EncodeFunc = serializer.Encode
		config.DecodeFunc = serializer.Decode
	case configuration.MySQL != nil: // If mysql configuration is provided, then use the mysql provider.
		providerName = "mysql"
		serializer := NewEncryptingSerializer(configuration.Secret)

		cf := mysql.NewConfigWith(
			configuration.MySQL.Host,
			configuration.MySQL.Port,
			configuration.MySQL.Username,
			configuration.MySQL.Password,
			configuration.MySQL.Database,
			"authelia_session",
		)
		mysqlConfig = &cf
		config.EncodeFunc = serializer.Encode
		config.DecodeFunc = serializer.Decode
	case configuration.PostgreSQL != nil: // If postgres configuration is provided, then use the postgre provider.
		providerName = "postgre"
		serializer := NewEncryptingSerializer(configuration.Secret)

		cf := postgre.NewConfigWith(
			configuration.PostgreSQL.Host,
			int64(configuration.PostgreSQL.Port),
			configuration.PostgreSQL.Username,
			configuration.PostgreSQL.Password,
			configuration.PostgreSQL.Database,
			"authelia_session",
		)
		postgreConfig = &cf
		config.EncodeFunc = serializer.Encode
		config.DecodeFunc = serializer.Decode
	case configuration.Local != nil: // If local configuration is provided, then use the sqlite3 provider.
		providerName = "sqlite3"
		serializer := NewEncryptingSerializer(configuration.Secret)

		cf := sqlite3.NewConfigWith(
			configuration.Local.Path,
			"authelia_session",
		)
		sqlite3Config = &cf
		config.EncodeFunc = serializer.Encode
		config.DecodeFunc = serializer.Decode
	default: // if no option is provided, use the memory provider.
		providerName = "memory"
	}

	return ProviderConfig{
		config:         config,
		redisConfig:    redisConfig,
		memcacheConfig: memcacheConfig,
		mysqlConfig:    mysqlConfig,
		postgreConfig:  postgreConfig,
		sqlite3Config:  sqlite3Config,
		providerName:   providerName,
	}
}
