package storage

import (
	_ "github.com/mattn/go-sqlite3" // Load the SQLite Driver used in the connection string.

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// SQLiteProvider is a SQLite3 provider.
type SQLiteProvider struct {
	SQLProvider
}

// NewSQLiteProvider constructs a SQLite provider.
func NewSQLiteProvider(config *schema.Configuration) (provider *SQLiteProvider) {
	provider = &SQLiteProvider{
		SQLProvider: NewSQLProvider(config, providerSQLite, "sqlite3", config.Storage.Local.Path),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = querySQLiteSelectExistingTables

	return provider
}
