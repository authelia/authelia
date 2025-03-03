package provider

import (
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
)

// NewStoragePostgreSQL creates a new storage.Provider using the *storage.PostgreSQLProvider given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewStoragePostgreSQL(config *schema.Configuration, caCertPool *x509.CertPool) storage.Provider {
	return storage.NewPostgreSQLProvider(config, caCertPool)
}

// NewStorageMySQL creates a new storage.Provider using the *storage.MySQLProvider given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewStorageMySQL(config *schema.Configuration, caCertPool *x509.CertPool) storage.Provider {
	return storage.NewMySQLProvider(config, caCertPool)
}

// NewStorageSQLite creates a new storage.Provider using the *storage.SQLiteProvider given a valid configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewStorageSQLite(config *schema.Configuration) storage.Provider {
	return storage.NewSQLiteProvider(config)
}
