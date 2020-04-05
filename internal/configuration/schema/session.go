package schema

// RedisSessionConfiguration represents the configuration related to redis session store.
type RedisSessionConfiguration struct {
	Host          string `mapstructure:"host"`
	Port          int64  `mapstructure:"port"`
	Password      string `mapstructure:"password"`
	DatabaseIndex int    `mapstructure:"database_index"`
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

// DefaultSessionConfiguration is the default session configuration
var DefaultSessionConfiguration = SessionConfiguration{
	Name:               "authelia_session",
	Expiration:         "1h",
	Inactivity:         "5m",
	RememberMeDuration: "1M",
}
