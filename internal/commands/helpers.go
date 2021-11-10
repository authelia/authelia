package commands

import (
	"errors"

	"github.com/authelia/authelia/v4/internal/storage"
)

func getStorageProvider() (provider storage.Provider, err error) {
	switch {
	case config.Storage.PostgreSQL != nil:
		provider = storage.NewPostgreSQLProvider(*config.Storage.PostgreSQL, config.Storage.EncryptionKey)
	case config.Storage.MySQL != nil:
		provider = storage.NewMySQLProvider(*config.Storage.MySQL, config.Storage.EncryptionKey)
	case config.Storage.Local != nil:
		provider = storage.NewSQLiteProvider(config.Storage.Local.Path, config.Storage.EncryptionKey)
	default:
		return nil, errors.New("no storage provider configured")
	}

	if (config.Storage.MySQL != nil && config.Storage.PostgreSQL != nil) ||
		(config.Storage.MySQL != nil && config.Storage.Local != nil) ||
		(config.Storage.PostgreSQL != nil && config.Storage.Local != nil) {
		return nil, errors.New("multiple storage providers are configured but should only configure one")
	}

	return provider, err
}
