package schema

import (
	"crypto/tls"
	"time"
)

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

	TLS *TLSConfig `koanf:"tls"`
}

// PostgreSQLStorageConfiguration represents the configuration of a PostgreSQL database.
type PostgreSQLStorageConfiguration struct {
	SQLStorageConfiguration `koanf:",squash"`
	Schema                  string `koanf:"schema"`

	TLS *TLSConfig `koanf:"tls"`

	SSL *PostgreSQLSSLStorageConfiguration `koanf:"ssl"`
}

// PostgreSQLSSLStorageConfiguration represents the SSL configuration of a PostgreSQL database.
type PostgreSQLSSLStorageConfiguration struct {
	Mode            string `koanf:"mode"`
	RootCertificate string `koanf:"root_certificate"`
	Certificate     string `koanf:"certificate"`
	Key             string `koanf:"key"`
}

// StorageConfiguration represents the configuration of the storage backend.
type StorageConfiguration struct {
	Local      *LocalStorageConfiguration      `koanf:"local"`
	MySQL      *MySQLStorageConfiguration      `koanf:"mysql"`
	PostgreSQL *PostgreSQLStorageConfiguration `koanf:"postgres"`

	EncryptionKey string `koanf:"encryption_key"`
}

// DefaultSQLStorageConfiguration represents the default SQL configuration.
var DefaultSQLStorageConfiguration = SQLStorageConfiguration{
	Timeout: 5 * time.Second,
}

// DefaultPostgreSQLStorageConfiguration represents the default PostgreSQL configuration.
var DefaultPostgreSQLStorageConfiguration = PostgreSQLStorageConfiguration{
	Schema: "public",
	TLS: &TLSConfig{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
	SSL: &PostgreSQLSSLStorageConfiguration{
		Mode: "disable",
	},
}
