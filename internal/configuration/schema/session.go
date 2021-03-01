package schema

// RedisSessionConfiguration represents the configuration related to redis session store.
type RedisSessionConfiguration struct {
	Host          string `mapstructure:"host"`
	Port          int64  `mapstructure:"port"`
	Username      string `mapstructure:"username"`
	Password      string `mapstructure:"password"`
	DatabaseIndex int    `mapstructure:"database_index"`
}

// MemcacheSessionConfiguration represents the configuration related to memcache session store.
type MemcacheSessionConfiguration struct {
	Host string `mapstructure:"host"`
	Port int64  `mapstructure:"port"`
}

// SQLSessionConfiguration represents the configuration of the SQL database.
type SQLSessionConfiguration struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// MySQLSessionConfiguration represents the configuration related to mysql session store.
type MySQLSessionConfiguration struct {
	SQLSessionConfiguration `mapstructure:",squash"`
}

// PostgreSQLSessionConfiguration represents the configuration related to postgre session store.
type PostgreSQLSessionConfiguration struct {
	SQLSessionConfiguration `mapstructure:",squash"`
}

// LocalSessionConfiguration represents the configuration related to sqlite session store.
type LocalSessionConfiguration struct {
	Path string `mapstructure:"path"`
}

// SessionConfiguration represents the configuration related to user sessions.
type SessionConfiguration struct {
	Name               string                          `mapstructure:"name"`
	Secret             string                          `mapstructure:"secret"`
	Expiration         string                          `mapstructure:"expiration"`
	Inactivity         string                          `mapstructure:"inactivity"`
	RememberMeDuration string                          `mapstructure:"remember_me_duration"`
	Domain             string                          `mapstructure:"domain"`
	Redis              *RedisSessionConfiguration      `mapstructure:"redis"`
	Memcache           []MemcacheSessionConfiguration  `mapstructure:"memcache,weak"`
	MySQL              *MySQLSessionConfiguration      `mapstructure:"mysql"`
	PostgreSQL         *PostgreSQLSessionConfiguration `mapstructure:"postgres"`
	Local              *LocalSessionConfiguration      `mapstructure:"local"`
}

// DefaultSessionConfiguration is the default session configuration.
var DefaultSessionConfiguration = SessionConfiguration{
	Name:               "authelia_session",
	Expiration:         "1h",
	Inactivity:         "5m",
	RememberMeDuration: "1M",
}
