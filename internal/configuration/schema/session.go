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

// IsSentinel returns true if either SentinelName or SentinelPassword are configured in the Redis HA Config.
func (c RedisHighAvailabilityConfiguration) IsSentinel() bool {
	return c.SentinelName != "" || c.SentinelPassword != ""
}

// RedisTimeoutsConfiguration sets the timeouts for the redis connection in seconds.
type RedisTimeoutsConfiguration struct {
	Dial  int `mapstructure:"dial"`
	Idle  int `mapstructure:"idle"`
	Pool  int `mapstructure:"pool"`
	Read  int `mapstructure:"read"`
	Write int `mapstructure:"write"`
}

// RedisSessionConfiguration represents the configuration related to redis session store.
type RedisSessionConfiguration struct {
	Host             string                              `mapstructure:"host"`
	Port             int                                 `mapstructure:"port"`
	Username         string                              `mapstructure:"username"`
	Password         string                              `mapstructure:"password"`
	DatabaseIndex    int                                 `mapstructure:"database_index"`
	PoolSize         int                                 `mapstructure:"pool_size"`
	TLS              *TLSConfig                          `mapstructure:"tls"`
	Timeouts         RedisTimeoutsConfiguration          `mapstructure:"timeouts"`
	HighAvailability *RedisHighAvailabilityConfiguration `mapstructure:"high_availability"`
}

// SessionConfiguration represents the configuration related to user sessions.
type SessionConfiguration struct {
	Name               string                     `mapstructure:"name"`
	Secret             string                     `mapstructure:"secret"`
	Expiration         string                     `mapstructure:"expiration"`
	Inactivity         string                     `mapstructure:"inactivity"`
	RememberMeDuration string                     `mapstructure:"remember_me_duration"`
	Domain             string                     `mapstructure:"domain"`
	Redis              *RedisSessionConfiguration `mapstructure:"redis"`
}

// DefaultSessionConfiguration is the default session configuration.
var DefaultSessionConfiguration = SessionConfiguration{
	Name:               "authelia_session",
	Expiration:         "1h",
	Inactivity:         "5m",
	RememberMeDuration: "1M",
}
