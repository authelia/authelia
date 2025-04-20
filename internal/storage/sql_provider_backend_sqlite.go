package storage

import (
	"database/sql"
	"encoding/base64"

	"github.com/mattn/go-sqlite3"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// SQLiteProvider is a SQLite3 provider.
type SQLiteProvider struct {
	SQLProvider
}

// NewSQLiteProvider constructs a SQLite provider.
func NewSQLiteProvider(config *schema.Configuration) (provider *SQLiteProvider) {
	provider = &SQLiteProvider{
		SQLProvider: NewSQLProvider(config, providerSQLite, "sqlite3e", config.Storage.Local.Path),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = querySQLiteSelectExistingTables

	return provider
}

func sqlite3BLOBToTEXTBase64(data []byte) (b64 string) {
	return base64.StdEncoding.EncodeToString(data)
}

func sqlite3TEXTBase64ToBLOB(b64 string) (data []byte, err error) {
	return base64.StdEncoding.DecodeString(b64)
}

func init() {
	sql.Register("sqlite3e", &sqlite3.SQLiteDriver{
		ConnectHook: func(conn *sqlite3.SQLiteConn) (err error) {
			if err = conn.RegisterFunc("BIN2B64", sqlite3BLOBToTEXTBase64, true); err != nil {
				return err
			}

			if err = conn.RegisterFunc("B642BIN", sqlite3TEXTBase64ToBLOB, true); err != nil {
				return err
			}

			return nil
		},
	})
}
