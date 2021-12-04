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
func NewPostgreSQLProvider(config *schema.Configuration) (provider *PostgreSQLProvider) {
	provider = &PostgreSQLProvider{
		SQLProvider: NewSQLProvider(config, providerPostgres, "pgx", dataSourceNamePostgreSQL(*config.Storage.PostgreSQL)),
	}

	// All providers have differing SELECT existing table statements.
	provider.sqlSelectExistingTables = queryPostgreSelectExistingTables

	// Specific alterations to this provider.
	// PostgreSQL doesn't have a UPSERT statement but has an ON CONFLICT operation instead.
	provider.sqlUpsertU2FDevice = fmt.Sprintf(queryFmtPostgresUpsertU2FDevice, tableU2FDevices)
	provider.sqlUpsertDuoDevice = fmt.Sprintf(queryFmtPostgresUpsertDuoDevice, tableDuoDevices)
	provider.sqlUpsertTOTPConfig = fmt.Sprintf(queryFmtPostgresUpsertTOTPConfiguration, tableTOTPConfigurations)
	provider.sqlUpsertPreferred2FAMethod = fmt.Sprintf(queryFmtPostgresUpsertPreferred2FAMethod, tableUserPreferences)
	provider.sqlUpsertEncryptionValue = fmt.Sprintf(queryFmtPostgresUpsertEncryptionValue, tableEncryption)

	// PostgreSQL requires rebinding of any query that contains a '?' placeholder to use the '$#' notation placeholders.
	provider.sqlFmtRenameTable = provider.db.Rebind(provider.sqlFmtRenameTable)
	provider.sqlSelectPreferred2FAMethod = provider.db.Rebind(provider.sqlSelectPreferred2FAMethod)
	provider.sqlSelectUserInfo = provider.db.Rebind(provider.sqlSelectUserInfo)
	provider.sqlSelectIdentityVerification = provider.db.Rebind(provider.sqlSelectIdentityVerification)
	provider.sqlInsertIdentityVerification = provider.db.Rebind(provider.sqlInsertIdentityVerification)
	provider.sqlConsumeIdentityVerification = provider.db.Rebind(provider.sqlConsumeIdentityVerification)
	provider.sqlSelectTOTPConfig = provider.db.Rebind(provider.sqlSelectTOTPConfig)
	provider.sqlDeleteTOTPConfig = provider.db.Rebind(provider.sqlDeleteTOTPConfig)
	provider.sqlSelectTOTPConfigs = provider.db.Rebind(provider.sqlSelectTOTPConfigs)
	provider.sqlUpdateTOTPConfigSecret = provider.db.Rebind(provider.sqlUpdateTOTPConfigSecret)
	provider.sqlUpdateTOTPConfigSecretByUsername = provider.db.Rebind(provider.sqlUpdateTOTPConfigSecretByUsername)
	provider.sqlSelectU2FDevice = provider.db.Rebind(provider.sqlSelectU2FDevice)
	provider.sqlSelectU2FDevices = provider.db.Rebind(provider.sqlSelectU2FDevices)
	provider.sqlUpdateU2FDevicePublicKey = provider.db.Rebind(provider.sqlUpdateU2FDevicePublicKey)
	provider.sqlUpdateU2FDevicePublicKeyByUsername = provider.db.Rebind(provider.sqlUpdateU2FDevicePublicKeyByUsername)
	provider.sqlSelectDuoDevice = provider.db.Rebind(provider.sqlSelectDuoDevice)
	provider.sqlDeleteDuoDevice = provider.db.Rebind(provider.sqlDeleteDuoDevice)
	provider.sqlInsertAuthenticationAttempt = provider.db.Rebind(provider.sqlInsertAuthenticationAttempt)
	provider.sqlSelectAuthenticationAttemptsByUsername = provider.db.Rebind(provider.sqlSelectAuthenticationAttemptsByUsername)
	provider.sqlInsertMigration = provider.db.Rebind(provider.sqlInsertMigration)
	provider.sqlSelectMigrations = provider.db.Rebind(provider.sqlSelectMigrations)
	provider.sqlSelectLatestMigration = provider.db.Rebind(provider.sqlSelectLatestMigration)
	provider.sqlSelectEncryptionValue = provider.db.Rebind(provider.sqlSelectEncryptionValue)

	provider.schema = config.Storage.PostgreSQL.Schema

	return provider
}

func dataSourceNamePostgreSQL(config schema.PostgreSQLStorageConfiguration) (dataSourceName string) {
	args := []string{
		fmt.Sprintf("host=%s", config.Host),
		fmt.Sprintf("user='%s'", config.Username),
		fmt.Sprintf("password='%s'", config.Password),
		fmt.Sprintf("dbname=%s", config.Database),
		fmt.Sprintf("search_path=%s", config.Schema),
		fmt.Sprintf("sslmode=%s", config.SSL.Mode),
	}

	if config.Port > 0 {
		args = append(args, fmt.Sprintf("port=%d", config.Port))
	}

	if config.SSL.RootCertificate != "" {
		args = append(args, fmt.Sprintf("sslrootcert=%s", config.SSL.RootCertificate))
	}

	if config.SSL.Certificate != "" {
		args = append(args, fmt.Sprintf("sslcert=%s", config.SSL.Certificate))
	}

	if config.SSL.Key != "" {
		args = append(args, fmt.Sprintf("sslkey=%s", config.SSL.Key))
	}

	args = append(args, fmt.Sprintf("connect_timeout=%d", int32(config.Timeout/time.Second)))

	return strings.Join(args, " ")
}
