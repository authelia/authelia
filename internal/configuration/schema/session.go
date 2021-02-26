package schema

// RedisNode Represents a Node.
type RedisNode struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}

// RedisSessionConfiguration represents the configuration related to redis session store.
type RedisSessionConfiguration struct {
	Host             string      `mapstructure:"host"`
	Port             int         `mapstructure:"port"`
	Username         string      `mapstructure:"username"`
	Password         string      `mapstructure:"password"`
	DatabaseIndex    int         `mapstructure:"database_index"`
	Sentinel         string      `mapstructure:"sentinel"`
	SentinelPassword string      `mapstructure:"sentinel_password"`
	Nodes            []RedisNode `mapstructure:"nodes"`
	RouteByLatency   bool        `mapstructure:"route_by_latency"`
	RouteRandomly    bool        `mapstructure:"route_randomly"`
	SlaveOnly        bool        `mapstructure:"slave_only"`
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
