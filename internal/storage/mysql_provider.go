package storage

import (
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	_ "github.com/go-sql-driver/mysql" // Load the MySQL Driver used in the connection string.
)

// MySQLProvider is a MySQL provider.
type MySQLProvider struct {
	SQLProvider
}

// NewMySQLProvider a MySQL provider.
func NewMySQLProvider(configuration schema.MySQLStorageConfiguration) (provider *MySQLProvider) {
	provider = &MySQLProvider{
		SQLProvider{
			name:       "mysql",
			driverName: "mysql",

			sqlUpgradesCreateTableStatements: sqlUpgradeCreateTableStatements,

			sqlGetPreferencesByUsername:     fmt.Sprintf("SELECT second_factor_method FROM %s WHERE username=?", userPreferencesTableName),
			sqlUpsertSecondFactorPreference: fmt.Sprintf("REPLACE INTO %s (username, second_factor_method) VALUES (?, ?)", userPreferencesTableName),

			sqlTestIdentityVerificationTokenExistence: fmt.Sprintf("SELECT EXISTS (SELECT * FROM %s WHERE token=?)", identityVerificationTokensTableName),
			sqlInsertIdentityVerificationToken:        fmt.Sprintf("INSERT INTO %s (token) VALUES (?)", identityVerificationTokensTableName),
			sqlDeleteIdentityVerificationToken:        fmt.Sprintf("DELETE FROM %s WHERE token=?", identityVerificationTokensTableName),

			sqlGetTOTPSecretByUsername: fmt.Sprintf("SELECT secret FROM %s WHERE username=?", totpSecretsTableName),
			sqlUpsertTOTPSecret:        fmt.Sprintf("REPLACE INTO %s (username, secret) VALUES (?, ?)", totpSecretsTableName),
			sqlDeleteTOTPSecret:        fmt.Sprintf("DELETE FROM %s WHERE username=?", totpSecretsTableName),

			sqlGetU2FDeviceHandleByUsername: fmt.Sprintf("SELECT keyHandle AS key_handle, publicKey AS public_key FROM %s WHERE username=?", u2fDeviceHandlesTableName),
			sqlUpsertU2FDeviceHandle:        fmt.Sprintf("REPLACE INTO %s (username, keyHandle, publicKey) VALUES (?, ?, ?)", u2fDeviceHandlesTableName),

			sqlInsertAuthenticationLog:         fmt.Sprintf("INSERT INTO %s (username, successful, time) VALUES (?, ?, ?)", authenticationLogsTableName),
			sqlGetFailedAuthenticationAttempts: fmt.Sprintf("SELECT username, successful, time FROM %s WHERE time>? AND username=? AND ORDER BY time DESC LIMIT ? OFFSET ?", authenticationLogsTableName),

			sqlGetExistingTables: "SELECT table_name FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema=database()",

			sqlConfigSetValue: fmt.Sprintf("REPLACE INTO %s (category, key_name, value) VALUES (?, ?, ?)", configTableName),
			sqlConfigGetValue: fmt.Sprintf("SELECT value FROM %s WHERE category=? AND key_name=?", configTableName),
		},
	}

	provider.sqlUpgradesCreateTableStatements[SchemaVersion(1)][authenticationLogsTableName] = "CREATE TABLE %s (username VARCHAR(100), successful BOOL, time INTEGER, INDEX usr_time_idx (username, time))"

	provider.connectionString = configuration.Username

	if configuration.Password != "" {
		provider.connectionString += fmt.Sprintf(":%s", configuration.Password)
	}

	if provider.connectionString != "" {
		provider.connectionString += "@"
	}

	address := configuration.Host
	if configuration.Port > 0 {
		address += fmt.Sprintf(":%d", configuration.Port)
	}

	provider.connectionString += fmt.Sprintf("tcp(%s)", address)
	if configuration.Database != "" {
		provider.connectionString += fmt.Sprintf("/%s", configuration.Database)
	}

	provider.connectionString += "?"
	provider.connectionString += fmt.Sprintf("timeout=%ds", int32(configuration.Timeout/time.Second))

	return provider
}
