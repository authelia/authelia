package schema

import "time"

// LocalStorageConfiguration represents the configuration when using local storage.
type LocalStorageConfiguration struct {
	Path string `koanf:"path"`
}

// SQLStorageConfiguration represents the configuration of the SQL database.
type SQLStorageConfiguration struct {
	Host     string        `koanf:"host"`
	Port     int           `koanf:"port"`
	Database string        `koanf:"database"`
	Username string        `koanf:"username"`
	Password string        `koanf:"password"`
	Timeout  time.Duration `koanf:"timeout"`
}

// MySQLStorageConfiguration represents the configuration of a MySQL database.
type MySQLStorageConfiguration struct {
	SQLStorageConfiguration `koanf:",squash"`
}

// PostgreSQLStorageConfiguration represents the configuration of a Postgres database.
type PostgreSQLStorageConfiguration struct {
	SQLStorageConfiguration `koanf:",squash"`
	SSLMode                 string `koanf:"sslmode"`
}

// StorageConfiguration represents the configuration of the storage backend.
type StorageConfiguration struct {
	Local      *LocalStorageConfiguration      `koanf:"local"`
	MySQL      *MySQLStorageConfiguration      `koanf:"mysql"`
	PostgreSQL *PostgreSQLStorageConfiguration `koanf:"postgres"`

	EncryptionKey string `koanf:"encryption_key"`
}

// DefaultPostgreSQLStorageConfiguration represents the default PostgreSQL configuration.
var DefaultPostgreSQLStorageConfiguration = PostgreSQLStorageConfiguration{
	SQLStorageConfiguration: SQLStorageConfiguration{
		Timeout: 5 * time.Second,
	},
}

// DefaultMySQLStorageConfiguration represents the default MySQL configuration.
var DefaultMySQLStorageConfiguration = MySQLStorageConfiguration{
	SQLStorageConfiguration: SQLStorageConfiguration{
		Timeout: 5 * time.Second,
	},
}
