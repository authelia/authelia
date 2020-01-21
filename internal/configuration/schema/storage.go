package schema

// LocalStorageConfiguration represents the configuration when using local storage.
type LocalStorageConfiguration struct {
	Path string `mapstructure:"path"`
}

// SQLStorageConfiguration represents the configuration of the SQL database
type SQLStorageConfiguration struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

// MySQLStorageConfiguration represents the configuration of a MySQL database
type MySQLStorageConfiguration struct {
	SQLStorageConfiguration `mapstructure:",squash"`
}

// PostgreSQLStorageConfiguration represents the configuration of a Postgres database
type PostgreSQLStorageConfiguration struct {
	SQLStorageConfiguration `mapstructure:",squash"`
	SSLMode                 string `mapstructure:"sslmode"`
}

// StorageConfiguration represents the configuration of the storage backend.
type StorageConfiguration struct {
	Local      *LocalStorageConfiguration      `mapstructure:"local"`
	MySQL      *MySQLStorageConfiguration      `mapstructure:"mysql"`
	PostgreSQL *PostgreSQLStorageConfiguration `mapstructure:"postgres"`
}
