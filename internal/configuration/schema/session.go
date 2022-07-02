package schema

import (
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
	Name               string                       `koanf:"name"`
	Domain             string                       `koanf:"domain"`
	Domains            []DomainSessionConfiguration `koanf:"domains"`
	DomainList         []string                     `koanf:"domain_list"`
	SameSite           string                       `koanf:"same_site"`
	Secret             string                       `koanf:"secret"`
	Expiration         time.Duration                `koanf:"expiration"`
	Inactivity         time.Duration                `koanf:"inactivity"`
	RememberMeDuration time.Duration                `koanf:"remember_me_duration"`

	Redis *RedisSessionConfiguration `koanf:"redis"`
}

// DomainSessionConfiguration represents the configuration related to a domain configuration in a multi domain environment.
type DomainSessionConfiguration struct {
	CookieDomain       string        `koanf:"cookie_domain"`
	Domains            []string      `koanf:"domains"`
	Expiration         time.Duration `koanf:"expiration"`
	Inactivity         time.Duration `koanf:"inactivity"`
	RememberMeDuration time.Duration `koanf:"remember_me_duration"`
}

// DefaultSessionConfiguration is the default session configuration.
var DefaultSessionConfiguration = SessionConfiguration{
	Name:               "authelia_session",
	Expiration:         time.Hour,
	Inactivity:         time.Minute * 5,
	RememberMeDuration: time.Hour * 24 * 30,
	SameSite:           "lax",
}
