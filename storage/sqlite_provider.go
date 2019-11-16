package storage

import (
	"database/sql"

	"github.com/clems4ever/authelia/logging"
	_ "github.com/mattn/go-sqlite3" // Load the SQLite Driver used in the connection string.
)

// SQLiteProvider is a sqlite3 provider
type SQLiteProvider struct {
	SQLProvider
}

// NewSQLiteProvider construct a sqlite provider.
func NewSQLiteProvider(path string) *SQLiteProvider {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		logging.Logger().Fatalf("Unable to create SQLite database %s: %s", path, err)
	}

	provider := SQLiteProvider{}
	if err := provider.initialize(db); err != nil {
		logging.Logger().Fatalf("Unable to initialize SQLite database %s: %s", path, err)
	}
	return &provider
}
