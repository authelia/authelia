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
func NewPostgreSQLProvider(config schema.PostgreSQLStorageConfiguration, encryptionKey string) (provider *PostgreSQLProvider) {
	provider = &PostgreSQLProvider{
		SQLProvider: NewSQLProvider(providerPostgres, "pgx", dataSourceNamePostgreSQL(config), encryptionKey),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryPostgreSelectExistingTables

	// Specific alterations to this provider.
	// PostgreSQL doesn't have a UPSERT statement but has an ON CONFLICT operation instead.
	provider.sqlUpsertU2FDevice = fmt.Sprintf(queryFmtPostgresUpsertU2FDevice, tableU2FDevices)
	provider.sqlUpsertTOTPConfig = fmt.Sprintf(queryFmtPostgresUpsertTOTPConfiguration, tableTOTPConfigurations)
	provider.sqlUpsertPreferred2FAMethod = fmt.Sprintf(queryFmtPostgresUpsertPreferred2FAMethod, tableUserPreferences)
	provider.sqlUpsertEncryptionValue = fmt.Sprintf(queryFmtPostgresUpsertEncryptionValue, tableEncryption)

	// PostgreSQL requires rebinding of any query that contains a '?' placeholder to use the '$#' notation placeholders.
	provider.sqlFmtRenameTable = provider.db.Rebind(provider.sqlFmtRenameTable)
	provider.sqlSelectPreferred2FAMethod = provider.db.Rebind(provider.sqlSelectPreferred2FAMethod)
	provider.sqlSelectUserInfo = provider.db.Rebind(provider.sqlSelectUserInfo)
	provider.sqlSelectExistsIdentityVerification = provider.db.Rebind(provider.sqlSelectExistsIdentityVerification)
	provider.sqlInsertIdentityVerification = provider.db.Rebind(provider.sqlInsertIdentityVerification)
	provider.sqlDeleteIdentityVerification = provider.db.Rebind(provider.sqlDeleteIdentityVerification)
	provider.sqlSelectTOTPConfig = provider.db.Rebind(provider.sqlSelectTOTPConfig)
	provider.sqlUpsertTOTPConfig = provider.db.Rebind(provider.sqlUpsertTOTPConfig)
	provider.sqlDeleteTOTPConfig = provider.db.Rebind(provider.sqlDeleteTOTPConfig)
	provider.sqlSelectTOTPConfigs = provider.db.Rebind(provider.sqlSelectTOTPConfigs)
	provider.sqlUpdateTOTPConfigSecret = provider.db.Rebind(provider.sqlUpdateTOTPConfigSecret)
	provider.sqlSelectU2FDevice = provider.db.Rebind(provider.sqlSelectU2FDevice)
	provider.sqlInsertAuthenticationAttempt = provider.db.Rebind(provider.sqlInsertAuthenticationAttempt)
	provider.sqlSelectAuthenticationAttemptsByUsername = provider.db.Rebind(provider.sqlSelectAuthenticationAttemptsByUsername)
	provider.sqlInsertMigration = provider.db.Rebind(provider.sqlInsertMigration)
	provider.sqlSelectEncryptionValue = provider.db.Rebind(provider.sqlSelectEncryptionValue)

	return provider
}

func dataSourceNamePostgreSQL(config schema.PostgreSQLStorageConfiguration) (dataSourceName string) {
	args := []string{
		fmt.Sprintf("user='%s'", config.Username),
		fmt.Sprintf("password='%s'", config.Password),
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

	args = append(args, fmt.Sprintf("connect_timeout=%d", int32(config.Timeout/time.Second)))

	return strings.Join(args, " ")
}
