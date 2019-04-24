package schema

// MongoStorageConfiguration represents the configuration related to mongo connection.
type MongoStorageConfiguration struct {
	URL      string `yaml:"url"`
	Database string `yaml:"database"`
	Auth     struct {
		Username string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"auth"`
}

// LocalStorageConfiguration represents the configuration when using local storage.
type LocalStorageConfiguration struct {
	Path string `yaml:"path"`
}

// StorageConfiguration represents the configuration of the storage backend.
type StorageConfiguration struct {
	Mongo *MongoStorageConfiguration `yaml:"mongo"`
	Local *LocalStorageConfiguration `yaml:"local"`
}
