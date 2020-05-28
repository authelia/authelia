package storage

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Load the SQLite Driver used in the connection string.

	"github.com/authelia/authelia/internal/logging"
)

// SQLiteProvider is a SQLite3 provider.
type SQLiteProvider struct {
	SQLProvider
}

// NewSQLiteProvider constructs a SQLite provider.
func NewSQLiteProvider(path string) *SQLiteProvider {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		logging.Logger().Fatalf("Unable to create SQLite database %s: %s", path, err)
	}

	provider := SQLiteProvider{
		SQLProvider{
			name: "sqlite",

			sqlCreateUserPreferencesTable:            SQLCreateUserPreferencesTable,
			sqlCreateIdentityVerificationTokensTable: SQLCreateIdentityVerificationTokensTable,
			sqlCreateTOTPSecretsTable:                SQLCreateTOTPSecretsTable,
			sqlCreateU2FDeviceHandlesTable:           SQLCreateU2FDeviceHandlesTable,
			sqlCreateAuthenticationLogsTable:         SQLCreateAuthenticationLogsTable,
			sqlCreateAuthenticationLogsUserTimeIndex: fmt.Sprintf("CREATE INDEX IF NOT EXISTS usr_time_idx ON %s (username, time)", authenticationLogsTableName),
			sqlCreateConfigTable:                     SQLCreateConfigTable,

			sqlGetPreferencesByUsername:     fmt.Sprintf("SELECT second_factor_method FROM %s WHERE username=?", preferencesTableName),
			sqlUpsertSecondFactorPreference: fmt.Sprintf("REPLACE INTO %s (username, second_factor_method) VALUES (?, ?)", preferencesTableName),

			sqlTestIdentityVerificationTokenExistence: fmt.Sprintf("SELECT EXISTS (SELECT * FROM %s WHERE token=?)", identityVerificationTokensTableName),
			sqlInsertIdentityVerificationToken:        fmt.Sprintf("INSERT INTO %s (token) VALUES (?)", identityVerificationTokensTableName),
			sqlDeleteIdentityVerificationToken:        fmt.Sprintf("DELETE FROM %s WHERE token=?", identityVerificationTokensTableName),

			sqlGetTOTPSecretByUsername: fmt.Sprintf("SELECT secret FROM %s WHERE username=?", totpSecretsTableName),
			sqlUpsertTOTPSecret:        fmt.Sprintf("REPLACE INTO %s (username, secret) VALUES (?, ?)", totpSecretsTableName),
			sqlDeleteTOTPSecret:        fmt.Sprintf("DELETE FROM %s WHERE username=?", totpSecretsTableName),

			sqlGetU2FDeviceHandleByUsername: fmt.Sprintf("SELECT keyHandle, publicKey FROM %s WHERE username=?", u2fDeviceHandlesTableName),
			sqlUpsertU2FDeviceHandle:        fmt.Sprintf("REPLACE INTO %s (username, keyHandle, publicKey) VALUES (?, ?, ?)", u2fDeviceHandlesTableName),

			sqlInsertAuthenticationLog:     fmt.Sprintf("INSERT INTO %s (username, successful, time) VALUES (?, ?, ?)", authenticationLogsTableName),
			sqlGetLatestAuthenticationLogs: fmt.Sprintf("SELECT successful, time FROM %s WHERE time>? AND username=? ORDER BY time DESC", authenticationLogsTableName),

			sqlGetExistingTables: "SELECT name FROM sqlite_master WHERE type='table'",
			sqlCheckTableExists:  "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?",

			sqlConfigSetValue: fmt.Sprintf("REPLACE INTO %s (category, key_name, value) VALUES(?, ?, ?)", configTableName),
			sqlConfigGetValue: fmt.Sprintf("SELECT value FROM %s WHERE category=? AND key_name=?", configTableName),
		},
	}
	if err := provider.initialize(db); err != nil {
		logging.Logger().Fatalf("Unable to initialize SQLite database %s: %s", path, err)
	}

	return &provider
}
