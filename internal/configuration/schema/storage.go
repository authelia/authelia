package schema

import (
	"crypto/tls"
	"net/url"
	"time"
)

// Storage represents the configuration of the storage backend.
type Storage struct {
	Local      *StorageLocal      `koanf:"local" yaml:"local,omitempty" toml:"local,omitempty" json:"local,omitempty" jsonschema:"title=Local" jsonschema_description:"The Local SQLite3 Storage configuration settings."`
	MySQL      *StorageMySQL      `koanf:"mysql" yaml:"mysql,omitempty" toml:"mysql,omitempty" json:"mysql,omitempty" jsonschema:"title=MySQL" jsonschema_description:"The MySQL/MariaDB Storage configuration settings."`
	PostgreSQL *StoragePostgreSQL `koanf:"postgres" yaml:"postgres,omitempty" toml:"postgres,omitempty" json:"postgres,omitempty" jsonschema:"title=PostgreSQL" jsonschema_description:"The PostgreSQL Storage configuration settings."`

	EncryptionKey string `koanf:"encryption_key" yaml:"encryption_key,omitempty" toml:"encryption_key,omitempty" json:"encryption_key,omitempty" jsonschema:"title=Encryption Key" jsonschema_description:"The Storage Encryption Key used to secure security sensitive values in the storage engine."`
}

// StorageLocal represents the configuration when using local storage.
type StorageLocal struct {
	Path string `koanf:"path" yaml:"path,omitempty" toml:"path,omitempty" json:"path,omitempty" jsonschema:"title=Path" jsonschema_description:"The Path for the SQLite3 database file."`
}

// StorageSQL represents the configuration of the SQL database.
type StorageSQL struct {
	Address  *AddressTCP   `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty" jsonschema:"title=Address" jsonschema_description:"The address of the SQL Server."`
	Database string        `koanf:"database" yaml:"database,omitempty" toml:"database,omitempty" json:"database,omitempty" jsonschema:"title=Database" jsonschema_description:"The database name to use upon a successful connection."`
	Username string        `koanf:"username" yaml:"username,omitempty" toml:"username,omitempty" json:"username,omitempty" jsonschema:"title=Username" jsonschema_description:"The username to use to authenticate."`
	Password string        `koanf:"password" yaml:"password,omitempty" toml:"password,omitempty" json:"password,omitempty" jsonschema:"title=Password" jsonschema_description:"The password to use to authenticate."` //nolint:gosec
	Timeout  time.Duration `koanf:"timeout" yaml:"timeout,omitempty" toml:"timeout,omitempty" json:"timeout,omitempty" jsonschema:"default=5 seconds,title=Timeout" jsonschema_description:"The timeout for the database connection."`
	TLS      *TLS          `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty"`
}

// StorageMySQL represents the configuration of a MySQL database.
type StorageMySQL struct {
	StorageSQL `koanf:",squash"`
}

// StoragePostgreSQL represents the configuration of a PostgreSQL database.
type StoragePostgreSQL struct {
	StorageSQL `koanf:",squash"`

	Schema string `koanf:"schema" yaml:"schema,omitempty" toml:"schema,omitempty" json:"schema,omitempty" jsonschema:"default=public,title=Schema" jsonschema_description:"The default schema name to use."`

	Servers []StoragePostgreSQLServer `koanf:"servers" yaml:"servers,omitempty" toml:"servers,omitempty" json:"servers,omitempty" jsonschema:"title=Servers" jsonschema_description:"The fallback PostgreSQL severs to connect to in addition to the one defined by the address."`

	// Deprecated: Use the TLS configuration instead.
	SSL *StoragePostgreSQLSSL `koanf:"ssl" yaml:"ssl,omitempty" toml:"ssl,omitempty" json:"ssl,omitempty" jsonschema:"deprecated,title=SSL"`
}

type StoragePostgreSQLServer struct {
	Address *AddressTCP `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty" jsonschema:"title=Address" jsonschema_description:"The address of the PostgreSQL Server."`
	TLS     *TLS        `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty"`
}

// StoragePostgreSQLSSL represents the SSL configuration of a PostgreSQL database.
type StoragePostgreSQLSSL struct {
	Mode            string `koanf:"mode" yaml:"mode,omitempty" toml:"mode,omitempty" json:"mode,omitempty" jsonschema:"deprecated,enum=disable,enum=verify-ca,enum=require,enum=verify-full,title=Mode" jsonschema_description:"The SSL mode to use, deprecated and replaced with the TLS options."`
	RootCertificate string `koanf:"root_certificate" yaml:"root_certificate,omitempty" toml:"root_certificate,omitempty" json:"root_certificate,omitempty" jsonschema:"deprecated,title=Root Certificate" jsonschema_description:"Path to the Root Certificate to use, deprecated and replaced with the TLS options."`
	Certificate     string `koanf:"certificate" yaml:"certificate,omitempty" toml:"certificate,omitempty" json:"certificate,omitempty" jsonschema:"deprecated,title=Certificate" jsonschema_description:"Path to the Certificate to use, deprecated and replaced with the TLS options."`
	Key             string `koanf:"key" yaml:"key,omitempty" toml:"key,omitempty" json:"key,omitempty" jsonschema:"deprecated,title=Key" jsonschema_description:"Path to the Private Key to use, deprecated and replaced with the TLS options."`
}

// DefaultSQLStorageConfiguration represents the default SQL configuration.
var DefaultSQLStorageConfiguration = StorageSQL{
	Timeout: 5 * time.Second,
}

// DefaultMySQLStorageConfiguration represents the default MySQL configuration.
var DefaultMySQLStorageConfiguration = StorageMySQL{
	StorageSQL: StorageSQL{
		Address: &AddressTCP{Address{true, false, -1, 3306, nil, &url.URL{Scheme: AddressSchemeTCP, Host: "localhost:3306"}}},
		TLS: &TLS{
			MinimumVersion: TLSVersion{tls.VersionTLS12},
		},
	},
}

// DefaultPostgreSQLStorageConfiguration represents the default PostgreSQL configuration.
var DefaultPostgreSQLStorageConfiguration = StoragePostgreSQL{
	StorageSQL: StorageSQL{
		Address: &AddressTCP{Address{true, false, -1, 5432, nil, &url.URL{Scheme: AddressSchemeTCP, Host: "localhost:5432"}}},
		TLS: &TLS{
			MinimumVersion: TLSVersion{tls.VersionTLS12},
		},
	},
	Servers: []StoragePostgreSQLServer{
		{
			Address: &AddressTCP{Address{true, false, -1, 5432, nil, &url.URL{Scheme: AddressSchemeTCP, Host: "localhost:5432"}}},
			TLS: &TLS{
				MinimumVersion: TLSVersion{tls.VersionTLS12},
			},
		},
	},
	Schema: "public",
	SSL: &StoragePostgreSQLSSL{
		Mode: "disable",
	},
}
