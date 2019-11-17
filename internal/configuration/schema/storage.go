package schema

// LocalStorageConfiguration represents the configuration when using local storage.
type LocalStorageConfiguration struct {
	Path string `yaml:"path"`
}

// SQLStorageConfiguration represents the configuration of the SQL database
type SQLStorageConfiguration struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// MySQLStorageConfiguration represents the configuration of a MySQL database
type MySQLStorageConfiguration struct {
	SQLStorageConfiguration `yaml:",inline"`
}

// PostgreSQLStorageConfiguration represents the configuration of a Postgres database
type PostgreSQLStorageConfiguration struct {
	SQLStorageConfiguration `yaml:",inline"`
	SSLMode                 string `yaml:"sslmode"`
}

// StorageConfiguration represents the configuration of the storage backend.
type StorageConfiguration struct {
	Local      *LocalStorageConfiguration      `yaml:"local"`
	MySQL      *MySQLStorageConfiguration      `yaml:"mysql"`
	PostgreSQL *PostgreSQLStorageConfiguration `yaml:"postgres"`
}
