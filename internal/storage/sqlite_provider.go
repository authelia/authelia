package storage

import (
	"database/sql"
	"fmt"

	// Register the sqlite SQL provider.
	_ "github.com/mattn/go-sqlite3"

	"github.com/authelia/authelia/internal/logging"
	// Load the SQLite Driver used in the connection string.
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
			sqlCreateUserPreferencesTable:            SQLCreateUserPreferencesTable,
			sqlCreateIdentityVerificationTokensTable: SQLCreateIdentityVerificationTokensTable,
			sqlCreateTOTPSecretsTable:                SQLCreateTOTPSecretsTable,
			sqlCreateU2FDeviceHandlesTable:           SQLCreateU2FDeviceHandlesTable,
			sqlCreateAuthenticationLogsTable:         fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (username VARCHAR(100), successful BOOL, time INTEGER)", authenticationLogsTableName), //nolint:gosec // Isn't actually a G201 issue.
			sqlCreateAuthenticationLogsUserTimeIndex: fmt.Sprintf("CREATE INDEX IF NOT EXISTS usr_time_idx ON %s (username, time)", authenticationLogsTableName),                       //nolint:gosec // Isn't actually a G201 issue.

			sqlGetPreferencesByUsername:     fmt.Sprintf("SELECT second_factor_method FROM %s WHERE username=?", preferencesTableName),           //nolint:gosec // Isn't actually a G201 issue.
			sqlUpsertSecondFactorPreference: fmt.Sprintf("REPLACE INTO %s (username, second_factor_method) VALUES (?, ?)", preferencesTableName), //nolint:gosec // Isn't actually a G201 issue.

			sqlTestIdentityVerificationTokenExistence: fmt.Sprintf("SELECT EXISTS (SELECT * FROM %s WHERE token=?)", identityVerificationTokensTableName), //nolint:gosec // Isn't actually a G201 issue.
			sqlInsertIdentityVerificationToken:        fmt.Sprintf("INSERT INTO %s (token) VALUES (?)", identityVerificationTokensTableName),              //nolint:gosec // Isn't actually a G201 issue.
			sqlDeleteIdentityVerificationToken:        fmt.Sprintf("DELETE FROM %s WHERE token=?", identityVerificationTokensTableName),                   //nolint:gosec // Isn't actually a G201 issue.

			sqlGetTOTPSecretByUsername: fmt.Sprintf("SELECT secret FROM %s WHERE username=?", totpSecretsTableName),           //nolint:gosec // Isn't actually a G201 issue.
			sqlUpsertTOTPSecret:        fmt.Sprintf("REPLACE INTO %s (username, secret) VALUES (?, ?)", totpSecretsTableName), //nolint:gosec // Isn't actually a G201 issue.
			sqlDeleteTOTPSecret:        fmt.Sprintf("DELETE FROM %s WHERE username=?", totpSecretsTableName),                  //nolint:gosec // Isn't actually a G201 issue.

			sqlGetU2FDeviceHandleByUsername: fmt.Sprintf("SELECT keyHandle, publicKey FROM %s WHERE username=?", u2fDeviceHandlesTableName),              //nolint:gosec // Isn't actually a G201 issue.
			sqlUpsertU2FDeviceHandle:        fmt.Sprintf("REPLACE INTO %s (username, keyHandle, publicKey) VALUES (?, ?, ?)", u2fDeviceHandlesTableName), //nolint:gosec // Isn't actually a G201 issue.

			sqlInsertAuthenticationLog:     fmt.Sprintf("INSERT INTO %s (username, successful, time) VALUES (?, ?, ?)", authenticationLogsTableName),                   //nolint:gosec // Isn't actually a G201 issue.
			sqlGetLatestAuthenticationLogs: fmt.Sprintf("SELECT successful, time FROM %s WHERE time>? AND username=? ORDER BY time DESC", authenticationLogsTableName), //nolint:gosec // Isn't actually a G201 issue.
		},
	}
	if err := provider.initialize(db); err != nil {
		logging.Logger().Fatalf("Unable to initialize SQLite database %s: %s", path, err)
	}
	return &provider
}
