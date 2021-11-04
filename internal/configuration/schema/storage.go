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
	Schema                  string `koanf:"schema"`
	SSLMode                 string `koanf:"sslmode"`
	SSL                     struct {
		Mode            string `koanf:"mode"`
		RootCertificate string `koanf:"root_certificate"`
		Certificate     string `koanf:"certificate"`
		Key             string `koanf:"key"`
	} `koanf:"SSL"`
}

// StorageConfiguration represents the configuration of the storage backend.
type StorageConfiguration struct {
	EncryptionKey string `koanf:"encryption_key"`

	Local      *LocalStorageConfiguration      `koanf:"local"`
	MySQL      *MySQLStorageConfiguration      `koanf:"mysql"`
	PostgreSQL *PostgreSQLStorageConfiguration `koanf:"postgres"`
}

// DefaultPostgreSQLStorageConfiguration represents the default PostgreSQL configuration.
var DefaultPostgreSQLStorageConfiguration = PostgreSQLStorageConfiguration{
	SQLStorageConfiguration: SQLStorageConfiguration{
		Timeout: 5 * time.Second,
	},
	Schema: "public",
}

// DefaultMySQLStorageConfiguration represents the default MySQL configuration.
var DefaultMySQLStorageConfiguration = MySQLStorageConfiguration{
	SQLStorageConfiguration: SQLStorageConfiguration{
		Timeout: 5 * time.Second,
	},
}
