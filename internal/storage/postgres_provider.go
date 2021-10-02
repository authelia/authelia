package storage

import (
	"fmt"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib" // Load the PostgreSQL Driver used in the connection string.

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// PostgreSQLProvider is a PostgreSQL provider.
type PostgreSQLProvider struct {
	SQLProvider
}

// NewPostgreSQLProvider a PostgreSQL provider.
func NewPostgreSQLProvider(config schema.PostgreSQLStorageConfiguration) (provider *PostgreSQLProvider) {
	provider = &PostgreSQLProvider{
		SQLProvider: NewSQLProvider("postgres", "pgx", buildPostgreSQLConnectionString(config)),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryPostgreSelectExistingTables

	// Specific alterations to this provider.
	// PostgreSQL doesn't have a UPSERT statement but has an ON CONFLICT operation instead.
	provider.sqlUpsertU2FDevice = fmt.Sprintf(queryFmtPostgresUpsertU2FDevice, tableU2FDevices)
	provider.sqlUpsertTOTPSecret = fmt.Sprintf(queryFmtPostgresUpsertTOTPSecret, tableTOTPSecrets)
	provider.sqlUpsertPreferred2FAMethod = fmt.Sprintf(queryFmtPostgresUpsertPreferred2FAMethod, tableUserPreferences)

	// TODO: Remove this as part of the migrations change.
	// Replace BLOB with BYTEA for Postgres.
	for version, tables := range provider.sqlUpgradesCreateTableStatements {
		for tableName, stmt := range tables {
			provider.sqlUpgradesCreateTableStatements[version][tableName] = strings.ReplaceAll(stmt, "BLOB", "BYTEA")
		}
	}

	provider.sqlConfigSetValue = fmt.Sprintf("INSERT INTO %s (category, key_name, value) VALUES ($1, $2, $3) ON CONFLICT (category, key_name) DO UPDATE SET value=$3", tableConfig)
	provider.sqlConfigGetValue = fmt.Sprintf("SELECT value FROM %s WHERE category=$1 AND key_name=$2", tableConfig)

	return provider
}

func buildPostgreSQLConnectionString(config schema.PostgreSQLStorageConfiguration) (connectionString string) {
	args := make([]string, 0)
	if config.Username != "" {
		args = append(args, fmt.Sprintf("user='%s'", config.Username))
	}

	if config.Password != "" {
		args = append(args, fmt.Sprintf("password='%s'", config.Password))
	}

	if config.Host != "" {
		args = append(args, fmt.Sprintf("host=%s", config.Host))
	}

	if config.Port > 0 {
		args = append(args, fmt.Sprintf("port=%d", config.Port))
	}

	if config.Database != "" {
		args = append(args, fmt.Sprintf("dbname=%s", config.Database))
	}

	if config.SSLMode != "" {
		args = append(args, fmt.Sprintf("sslmode=%s", config.SSLMode))
	}

	args = append(args, fmt.Sprintf("connect_timeout=%d", int32(config.Timeout/time.Second)))

	return strings.Join(args, " ")
}
