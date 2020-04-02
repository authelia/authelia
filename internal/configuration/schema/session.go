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
	// TODO(james-d-elliott): Convert to duration notation (Both Expiration and Activity need to be strings, and default needs to be changed)
	Name               string                     `mapstructure:"name"`
	Secret             string                     `mapstructure:"secret"`
	Expiration         int64                      `mapstructure:"expiration"` // Expiration in seconds
	Inactivity         int64                      `mapstructure:"inactivity"` // Inactivity in seconds
	RememberMeDuration string                     `mapstructure:"remember_me_duration"`
	Domain             string                     `mapstructure:"domain"`
	Redis              *RedisSessionConfiguration `mapstructure:"redis"`
}

// DefaultSessionConfiguration is the default session configuration
var DefaultSessionConfiguration = SessionConfiguration{
	Name:               "authelia_session",
	Expiration:         3600,
	RememberMeDuration: "1M",
}
