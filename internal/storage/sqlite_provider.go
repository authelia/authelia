package storage

import (
	"fmt"

	_ "github.com/mattn/go-sqlite3" // Load the SQLite Driver used in the connection string.
)

// SQLiteProvider is a SQLite3 provider.
type SQLiteProvider struct {
	SQLProvider
}

// NewSQLiteProvider constructs a SQLite provider.
func NewSQLiteProvider(path string) (provider *SQLiteProvider) {
	provider = &SQLiteProvider{
		SQLProvider{
			name:             "sqlite",
			driverName:       "sqlite3",
			connectionString: path,

			sqlUpgradesCreateTableStatements:        sqlUpgradeCreateTableStatements,
			sqlUpgradesCreateTableIndexesStatements: sqlUpgradesCreateTableIndexesStatements,

			sqlSelectPreferred2FAMethodByUsername: fmt.Sprintf("SELECT second_factor_method FROM %s WHERE username=?", userPreferencesTableName),
			sqlUpsertPreferred2FAMethod:           fmt.Sprintf("REPLACE INTO %s (username, second_factor_method) VALUES (?, ?)", userPreferencesTableName),

			sqlTestIdentityVerificationTokenExistence: fmt.Sprintf("SELECT EXISTS (SELECT * FROM %s WHERE token=?)", identityVerificationTokensTableName),
			sqlInsertIdentityVerificationToken:        fmt.Sprintf("INSERT INTO %s (token) VALUES (?)", identityVerificationTokensTableName),
			sqlDeleteIdentityVerificationToken:        fmt.Sprintf("DELETE FROM %s WHERE token=?", identityVerificationTokensTableName),

			sqlGetTOTPSecretByUsername: fmt.Sprintf("SELECT secret FROM %s WHERE username=?", totpSecretsTableName),
			sqlUpsertTOTPSecret:        fmt.Sprintf("REPLACE INTO %s (username, secret) VALUES (?, ?)", totpSecretsTableName),
			sqlDeleteTOTPSecret:        fmt.Sprintf("DELETE FROM %s WHERE username=?", totpSecretsTableName),

			sqlSelectU2FDeviceHandleByUsername: fmt.Sprintf("SELECT keyHandle AS key_handle, publicKey AS public_key FROM %s WHERE username=?", u2fDeviceHandlesTableName),
			sqlUpsertU2FDeviceHandle:           fmt.Sprintf("REPLACE INTO %s (username, keyHandle, publicKey) VALUES (?, ?, ?)", u2fDeviceHandlesTableName),

			sqlInsertAuthenticationAttempt:            fmt.Sprintf("INSERT INTO %s (username, successful, time) VALUES (?, ?, ?)", authenticationLogsTableName),
			sqlSelectAuthenticationAttemptsByUsername: fmt.Sprintf("SELECT username, successful, time FROM %s WHERE time>? AND username=? AND ORDER BY time DESC LIMIT ? OFFSET ?", authenticationLogsTableName),

			sqlSelectExistingTables: "SELECT name FROM sqlite_master WHERE type='table'",

			sqlConfigSetValue: fmt.Sprintf("REPLACE INTO %s (category, key_name, value) VALUES (?, ?, ?)", configTableName),
			sqlConfigGetValue: fmt.Sprintf("SELECT value FROM %s WHERE category=? AND key_name=?", configTableName),
		},
	}

	return provider
}
