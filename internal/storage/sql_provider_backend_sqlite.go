package storage

import (
	_ "github.com/mattn/go-sqlite3" // Load the SQLite Driver used in the connection string.
)

// SQLiteProvider is a SQLite3 provider.
type SQLiteProvider struct {
	SQLProvider
}

// NewSQLiteProvider constructs a SQLite provider.
func NewSQLiteProvider(path string) (provider *SQLiteProvider) {
	provider = &SQLiteProvider{
		SQLProvider: NewSQLProvider(providerSQLite, "sqlite3", path),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = querySQLiteSelectExistingTables

	return provider
}
