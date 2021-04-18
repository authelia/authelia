package schema

// RedisNode Represents a Node.
type RedisNode struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// RedisHighAvailabilityConfiguration holds configuration variables for Redis Cluster/Sentinel.
type RedisHighAvailabilityConfiguration struct {
	SentinelName     string      `mapstructure:"sentinel_name"`
	SentinelPassword string      `mapstructure:"sentinel_password"`
	Nodes            []RedisNode `mapstructure:"nodes"`
	RouteByLatency   bool        `mapstructure:"route_by_latency"`
	RouteRandomly    bool        `mapstructure:"route_randomly"`
}

// RedisSessionConfiguration represents the configuration related to redis session store.
type RedisSessionConfiguration struct {
	Host                     string                              `mapstructure:"host"`
	Port                     int                                 `mapstructure:"port"`
	Username                 string                              `mapstructure:"username"`
	Password                 string                              `mapstructure:"password"`
	DatabaseIndex            int                                 `mapstructure:"database_index"`
	MaximumActiveConnections int                                 `mapstructure:"maximum_active_connections"`
	MinimumIdleConnections   int                                 `mapstructure:"minimum_idle_connections"`
	TLS                      *TLSConfig                          `mapstructure:"tls"`
	HighAvailability         *RedisHighAvailabilityConfiguration `mapstructure:"high_availability"`
}

// SessionConfiguration represents the configuration related to user sessions.
type SessionConfiguration struct {
	Name               string                     `mapstructure:"name"`
	Domain             string                     `mapstructure:"domain"`
	SameSite           string                     `mapstructure:"same_site"`
	Secret             string                     `mapstructure:"secret"`
	Expiration         string                     `mapstructure:"expiration"`
	Inactivity         string                     `mapstructure:"inactivity"`
	RememberMeDuration string                     `mapstructure:"remember_me_duration"`
	Redis              *RedisSessionConfiguration `mapstructure:"redis"`
}

// DefaultSessionConfiguration is the default session configuration.
var DefaultSessionConfiguration = SessionConfiguration{
	Name:               "authelia_session",
	Expiration:         "1h",
	Inactivity:         "5m",
	RememberMeDuration: "1M",
	SameSite:           "lax",
}
