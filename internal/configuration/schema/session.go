package schema

// RedisSessionConfiguration represents the configuration related to redis session store.
type RedisSessionConfiguration struct {
	Host          string `mapstructure:"host"`
	Port          int64  `mapstructure:"port"`
	Password      string `mapstructure:"password"`
	DatabaseIndex int    `mapstructure:"database_index"`
}

type RememberMeConfiguration struct {
	Duration     int64  `mapstructure:"duration"`
	DurationUnit string `mapstructure:"duration_unit"`
	Refresh      bool   `mapstructure:"refresh"`
}

// SessionConfiguration represents the configuration related to user sessions.
type SessionConfiguration struct {
	Name       string                     `mapstructure:"name"`
	Secret     string                     `mapstructure:"secret"`
	Expiration int64                      `mapstructure:"expiration"`  // Expiration in seconds
	Inactivity int64                      `mapstructure:"inactivity"`  // Inactivity in seconds
	RememberMe *RememberMeConfiguration   `mapstructure:"remember_me"` // Remember Me Expiration in seconds
	Domain     string                     `mapstructure:"domain"`
	Redis      *RedisSessionConfiguration `mapstructure:"redis"`
}

// DefaultSessionConfiguration is the default session configuration
var DefaultSessionConfiguration = SessionConfiguration{
	Name:       "authelia_session",
	Expiration: 3600,
	RememberMe: &DefaultSessionRememberMeConfiguration,
}
var DefaultSessionRememberMeConfiguration = RememberMeConfiguration{
	Duration:     1,
	DurationUnit: "y",
}
