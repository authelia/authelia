package schema

// RedisSessionConfiguration represents the configuration related to redis session store.
type RedisSessionConfiguration struct {
	Host     string `yaml:"host"`
	Port     int64  `yaml:"port"`
	Password string `yaml:"password"`
}

// SessionConfiguration represents the configuration related to user sessions.
type SessionConfiguration struct {
	Name   string `yaml:"name"`
	Secret string `yaml:"secret"`
	// Expiration in seconds
	Expiration int64 `yaml:"expiration"`
	// Inactivity in seconds
	Inactivity int64                      `yaml:"inactivity"`
	Domain     string                     `yaml:"domain"`
	Redis      *RedisSessionConfiguration `yaml:"redis"`
}

// DefaultSessionConfiguration is the default session configuration
var DefaultSessionConfiguration = SessionConfiguration{
	Name:       "authelia_session",
	Expiration: 3600,
}
