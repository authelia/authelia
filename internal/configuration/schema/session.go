package schema

// RedisSessionConfiguration represents the configuration related to redis session store.
type RedisSessionConfiguration struct {
	Host     string `mapstructure:"host"`
	Port     int64  `mapstructure:"port"`
	Password string `mapstructure:"password"`
}

// SessionConfiguration represents the configuration related to user sessions.
type SessionConfiguration struct {
	Name   string `mapstructure:"name"`
	Secret string `mapstructure:"secret"`
	// Expiration in seconds
	Expiration int64 `mapstructure:"expiration"`
	// Inactivity in seconds
	Inactivity int64                      `mapstructure:"inactivity"`
	Domain     string                     `mapstructure:"domain"`
	Redis      *RedisSessionConfiguration `mapstructure:"redis"`
}

// DefaultSessionConfiguration is the default session configuration
var DefaultSessionConfiguration = SessionConfiguration{
	Name:       "authelia_session",
	Expiration: 3600,
}
