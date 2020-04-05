package storage

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/lib/pq" // Load the PostgreSQL Driver used in the connection string.

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
)

// PostgreSQLProvider is a Postrgres provider
type PostgreSQLProvider struct {
	SQLProvider
}

// NewPostgreSQLProvider a SQL provider
func NewPostgreSQLProvider(configuration schema.PostgreSQLStorageConfiguration) *PostgreSQLProvider {
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

	connectionString := strings.Join(args, " ")

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		logging.Logger().Fatalf("Unable to connect to SQL database: %v", err)
	}

	provider := PostgreSQLProvider{
		SQLProvider{
			sqlCreateUserPreferencesTable:            SQLCreateUserPreferencesTable,
			sqlCreateIdentityVerificationTokensTable: SQLCreateIdentityVerificationTokensTable,
			sqlCreateTOTPSecretsTable:                SQLCreateTOTPSecretsTable,
			sqlCreateU2FDeviceHandlesTable:           SQLCreateU2FDeviceHandlesTable,
			sqlCreateAuthenticationLogsTable:         fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (username VARCHAR(100), successful BOOL, time INTEGER)", authenticationLogsTableName),
			sqlCreateAuthenticationLogsUserTimeIndex: fmt.Sprintf("CREATE INDEX IF NOT EXISTS usr_time_idx ON %s (username, time)", authenticationLogsTableName),

			sqlGetPreferencesByUsername:     fmt.Sprintf("SELECT second_factor_method FROM %s WHERE username=$1", preferencesTableName),
			sqlUpsertSecondFactorPreference: fmt.Sprintf("INSERT INTO %s (username, second_factor_method) VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET second_factor_method=$2", preferencesTableName),

			sqlTestIdentityVerificationTokenExistence: fmt.Sprintf("SELECT EXISTS (SELECT * FROM %s WHERE token=$1)", identityVerificationTokensTableName),
			sqlInsertIdentityVerificationToken:        fmt.Sprintf("INSERT INTO %s (token) VALUES ($1)", identityVerificationTokensTableName),
			sqlDeleteIdentityVerificationToken:        fmt.Sprintf("DELETE FROM %s WHERE token=$1", identityVerificationTokensTableName),

			sqlGetTOTPSecretByUsername: fmt.Sprintf("SELECT secret FROM %s WHERE username=$1", totpSecretsTableName),
			sqlUpsertTOTPSecret:        fmt.Sprintf("INSERT INTO %s (username, secret) VALUES ($1, $2) ON CONFLICT (username) DO UPDATE SET secret=$2", totpSecretsTableName),
			sqlDeleteTOTPSecret:        fmt.Sprintf("DELETE FROM %s WHERE username=$1", totpSecretsTableName),

			sqlGetU2FDeviceHandleByUsername: fmt.Sprintf("SELECT keyHandle, publicKey FROM %s WHERE username=$1", u2fDeviceHandlesTableName),
			sqlUpsertU2FDeviceHandle:        fmt.Sprintf("INSERT INTO %s (username, keyHandle, publicKey) VALUES ($1, $2, $3) ON CONFLICT (username) DO UPDATE SET keyHandle=$2, publicKey=$3", u2fDeviceHandlesTableName),

			sqlInsertAuthenticationLog:     fmt.Sprintf("INSERT INTO %s (username, successful, time) VALUES ($1, $2, $3)", authenticationLogsTableName),
			sqlGetLatestAuthenticationLogs: fmt.Sprintf("SELECT successful, time FROM %s WHERE time>$1 AND username=$2 ORDER BY time DESC", authenticationLogsTableName),
		},
	}
	if err := provider.initialize(db); err != nil {
		logging.Logger().Fatalf("Unable to initialize SQL database: %v", err)
	}
	return &provider
}
