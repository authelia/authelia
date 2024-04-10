package provider

import (
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/storage"
)

func NewStoragePostgreSQL(config *schema.Configuration, caCertPool *x509.CertPool) storage.Provider {
	return storage.NewPostgreSQLProvider(config, caCertPool)
}

func NewStorageMySQL(config *schema.Configuration, caCertPool *x509.CertPool) storage.Provider {
	return storage.NewMySQLProvider(config, caCertPool)
}

func NewStorageSQLite(config *schema.Configuration) storage.Provider {
	return storage.NewSQLiteProvider(config)
}
