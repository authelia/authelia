package storage

import (
	"fmt"
	"strings"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	_ "github.com/jackc/pgx/v4/stdlib" // Load the PostgreSQL Driver used in the connection string.
)

// PostgreSQLProvider is a PostgreSQL provider.
type PostgreSQLProvider struct {
	SQLProvider
}

// NewPostgreSQLProvider a PostgreSQL provider.
func NewPostgreSQLProvider(configuration schema.PostgreSQLStorageConfiguration) (provider *PostgreSQLProvider) {
	provider = &PostgreSQLProvider{
		SQLProvider{
			name:       "postgres",
			driverName: "pgx",

			sqlUpgradesCreateTableStatements:        sqlUpgradeCreateTableStatements,
			sqlUpgradesCreateTableIndexesStatements: sqlUpgradesCreateTableIndexesStatements,

			sqlGetPreferencesByUsername:     fmt.Sprintf("SELECT second_factor_method FROM %s WHERE username=$1", userPreferencesTableName),
			sqlUpsertSecondFactorPreference: fmt.Sprintf("INSERT INTO %s (username, second_factor_method) VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET second_factor_method=$2", userPreferencesTableName),

			sqlTestIdentityVerificationTokenExistence: fmt.Sprintf("SELECT EXISTS (SELECT * FROM %s WHERE token=$1)", identityVerificationTokensTableName),
			sqlInsertIdentityVerificationToken:        fmt.Sprintf("INSERT INTO %s (token) VALUES ($1)", identityVerificationTokensTableName),
			sqlDeleteIdentityVerificationToken:        fmt.Sprintf("DELETE FROM %s WHERE token=$1", identityVerificationTokensTableName),

			sqlGetTOTPSecretByUsername: fmt.Sprintf("SELECT secret FROM %s WHERE username=$1", totpSecretsTableName),
			sqlUpsertTOTPSecret:        fmt.Sprintf("INSERT INTO %s (username, secret) VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET secret=$2", totpSecretsTableName),
			sqlDeleteTOTPSecret:        fmt.Sprintf("DELETE FROM %s WHERE username=$1", totpSecretsTableName),

			sqlGetU2FDeviceHandleByUsername: fmt.Sprintf("SELECT keyHandle AS key_handle, publicKey AS public_key FROM %s WHERE username=$1", u2fDeviceHandlesTableName),
			sqlUpsertU2FDeviceHandle:        fmt.Sprintf("INSERT INTO %s (username, keyHandle, publicKey) VALUES ($1, $2, $3) ON CONFLICT (username) DO UPDATE SET keyHandle=$2, publicKey=$3", u2fDeviceHandlesTableName),

			sqlInsertAuthenticationLog:         fmt.Sprintf("INSERT INTO %s (username, successful, time) VALUES ($1, $2, $3)", authenticationLogsTableName),
			sqlGetFailedAuthenticationAttempts: fmt.Sprintf("SELECT username, successful, time FROM %s WHERE time>$1 AND username=$2 AND ORDER BY time DESC LIMIT $3 OFFSET $4", authenticationLogsTableName),

			sqlGetExistingTables: "SELECT table_name FROM information_schema.tables WHERE table_type='BASE TABLE' AND table_schema='public'",

			sqlConfigSetValue: fmt.Sprintf("INSERT INTO %s (category, key_name, value) VALUES ($1, $2, $3) ON CONFLICT (category, key_name) DO UPDATE SET value=$3", configTableName),
			sqlConfigGetValue: fmt.Sprintf("SELECT value FROM %s WHERE category=$1 AND key_name=$2", configTableName),
		},
	}

	// Replace BLOB with BYTEA for Postgres.
	for version, tables := range provider.sqlUpgradesCreateTableStatements {
		for tableName, stmt := range tables {
			provider.sqlUpgradesCreateTableStatements[version][tableName] = strings.ReplaceAll(stmt, "BLOB", "BYTEA")
		}
	}

	args := make([]string, 0)
	if configuration.Username != "" {
		args = append(args, fmt.Sprintf("user='%s'", configuration.Username))
	}

	if configuration.Password != "" {
		args = append(args, fmt.Sprintf("password='%s'", configuration.Password))
	}

	if configuration.Host != "" {
		args = append(args, fmt.Sprintf("host=%s", configuration.Host))
	}

	if configuration.Port > 0 {
		args = append(args, fmt.Sprintf("port=%d", configuration.Port))
	}

	if configuration.Database != "" {
		args = append(args, fmt.Sprintf("dbname=%s", configuration.Database))
	}

	if configuration.SSLMode != "" {
		args = append(args, fmt.Sprintf("sslmode=%s", configuration.SSLMode))
	}

	args = append(args, fmt.Sprintf("connect_timeout=%d", int32(configuration.Timeout/time.Second)))
	provider.connectionString = strings.Join(args, " ")

	return provider
}
