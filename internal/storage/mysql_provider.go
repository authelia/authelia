package storage

import (
	"database/sql"
	"fmt"

	// Register the mysql SQL provider.
	_ "github.com/go-sql-driver/mysql"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	// Load the MySQL Driver used in the connection string.
)

// MySQLProvider is a MySQL provider.
type MySQLProvider struct {
	SQLProvider
}

// NewMySQLProvider a MySQL provider.
func NewMySQLProvider(configuration schema.MySQLStorageConfiguration) *MySQLProvider {
	connectionString := configuration.Username

	if configuration.Password != "" {
		connectionString += fmt.Sprintf(":%s", configuration.Password)
	}

	if connectionString != "" {
		connectionString += "@"
	}

	address := configuration.Host
	if configuration.Port > 0 {
		address += fmt.Sprintf(":%d", configuration.Port)
	}
	connectionString += fmt.Sprintf("tcp(%s)", address)

	if configuration.Database != "" {
		connectionString += fmt.Sprintf("/%s", configuration.Database)
	}

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		logging.Logger().Fatalf("Unable to connect to SQL database: %v", err)
	}

	provider := MySQLProvider{
		SQLProvider{
			sqlCreateUserPreferencesTable:            SQLCreateUserPreferencesTable,
			sqlCreateIdentityVerificationTokensTable: SQLCreateIdentityVerificationTokensTable,
			sqlCreateTOTPSecretsTable:                SQLCreateTOTPSecretsTable,
			sqlCreateU2FDeviceHandlesTable:           SQLCreateU2FDeviceHandlesTable,
			sqlCreateAuthenticationLogsTable:         SQLCreateAuthenticationLogsTable,

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
		logging.Logger().Fatalf("Unable to initialize SQL database: %v", err)
	}
	return &provider
}
