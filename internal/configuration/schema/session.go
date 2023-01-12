package schema

import (
	"crypto/tls"
	"net/url"
	"time"
)

// RedisNode Represents a Node.
type RedisNode struct {
	Host string `koanf:"host"`
	Port int    `koanf:"port"`
}

// RedisHighAvailabilityConfiguration holds configuration variables for Redis Cluster/Sentinel.
type RedisHighAvailabilityConfiguration struct {
	SentinelName     string      `koanf:"sentinel_name"`
	SentinelUsername string      `koanf:"sentinel_username"`
	SentinelPassword string      `koanf:"sentinel_password"`
	Nodes            []RedisNode `koanf:"nodes"`
	RouteByLatency   bool        `koanf:"route_by_latency"`
	RouteRandomly    bool        `koanf:"route_randomly"`
}

// RedisSessionConfiguration represents the configuration related to redis session store.
type RedisSessionConfiguration struct {
	Host                     string                              `koanf:"host"`
	Port                     int                                 `koanf:"port"`
	Username                 string                              `koanf:"username"`
	Password                 string                              `koanf:"password"`
	DatabaseIndex            int                                 `koanf:"database_index"`
	MaximumActiveConnections int                                 `koanf:"maximum_active_connections"`
	MinimumIdleConnections   int                                 `koanf:"minimum_idle_connections"`
	TLS                      *TLSConfig                          `koanf:"tls"`
	HighAvailability         *RedisHighAvailabilityConfiguration `koanf:"high_availability"`
}

// SessionConfiguration represents the configuration related to user sessions.
type SessionConfiguration struct {
	Secret string `koanf:"secret"`

	SessionCookieCommonConfiguration `koanf:",squash"`

	Cookies []SessionCookieConfiguration `koanf:"cookies"`

	Redis *RedisSessionConfiguration `koanf:"redis"`
}

type SessionCookieCommonConfiguration struct {
	Name       string        `koanf:"name"`
	Domain     string        `koanf:"domain"`
	SameSite   string        `koanf:"same_site"`
	Expiration time.Duration `koanf:"expiration"`
	Inactivity time.Duration `koanf:"inactivity"`
	RememberMe time.Duration `koanf:"remember_me"`

	DisableRememberMe bool
}

// SessionCookieConfiguration represents the configuration for a cookie domain.
type SessionCookieConfiguration struct {
	SessionCookieCommonConfiguration `koanf:",squash"`

	AutheliaURL *url.URL `koanf:"authelia_url"`
}

// DefaultSessionConfiguration is the default session configuration.
var DefaultSessionConfiguration = SessionConfiguration{
	SessionCookieCommonConfiguration: SessionCookieCommonConfiguration{
		Name:       "authelia_session",
		Expiration: time.Hour,
		Inactivity: time.Minute * 5,
		RememberMe: time.Hour * 24 * 30,
		SameSite:   "lax",
	},
}

// DefaultRedisConfiguration is the default redis configuration.
var DefaultRedisConfiguration = RedisSessionConfiguration{
	TLS: &TLSConfig{
		MinimumVersion: TLSVersion{Value: tls.VersionTLS12},
	},
}
