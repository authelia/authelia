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

// StorageConfiguration represents the configuration of the storage backend.
type StorageConfiguration struct {
	Local *LocalStorageConfiguration `yaml:"local"`
	SQL   *SQLStorageConfiguration   `yaml:"sql"`
}
