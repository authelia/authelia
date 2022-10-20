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
	provider.sqlUpsertWebauthnDevice = fmt.Sprintf(queryFmtUpsertWebauthnDevicePostgreSQL, tableWebauthnDevices)
	provider.sqlUpsertDuoDevice = fmt.Sprintf(queryFmtUpsertDuoDevicePostgreSQL, tableDuoDevices)
	provider.sqlUpsertTOTPConfig = fmt.Sprintf(queryFmtUpsertTOTPConfigurationPostgreSQL, tableTOTPConfigurations)
	provider.sqlUpsertPreferred2FAMethod = fmt.Sprintf(queryFmtUpsertPreferred2FAMethodPostgreSQL, tableUserPreferences)
	provider.sqlUpsertEncryptionValue = fmt.Sprintf(queryFmtUpsertEncryptionValuePostgreSQL, tableEncryption)
	provider.sqlUpsertOAuth2BlacklistedJTI = fmt.Sprintf(queryFmtUpsertOAuth2BlacklistedJTIPostgreSQL, tableOAuth2BlacklistedJTI)
	provider.sqlInsertOAuth2ConsentPreConfiguration = fmt.Sprintf(queryFmtInsertOAuth2ConsentPreConfigurationPostgreSQL, tableOAuth2ConsentPreConfiguration)

	// PostgreSQL requires rebinding of any query that contains a '?' placeholder to use the '$#' notation placeholders.
	provider.sqlFmtRenameTable = provider.db.Rebind(provider.sqlFmtRenameTable)

	provider.sqlSelectPreferred2FAMethod = provider.db.Rebind(provider.sqlSelectPreferred2FAMethod)
	provider.sqlSelectUserInfo = provider.db.Rebind(provider.sqlSelectUserInfo)

	provider.sqlInsertUserOpaqueIdentifier = provider.db.Rebind(provider.sqlInsertUserOpaqueIdentifier)
	provider.sqlSelectUserOpaqueIdentifier = provider.db.Rebind(provider.sqlSelectUserOpaqueIdentifier)
	provider.sqlSelectUserOpaqueIdentifierBySignature = provider.db.Rebind(provider.sqlSelectUserOpaqueIdentifierBySignature)

	provider.sqlSelectIdentityVerification = provider.db.Rebind(provider.sqlSelectIdentityVerification)
	provider.sqlInsertIdentityVerification = provider.db.Rebind(provider.sqlInsertIdentityVerification)
	provider.sqlConsumeIdentityVerification = provider.db.Rebind(provider.sqlConsumeIdentityVerification)

	provider.sqlSelectTOTPConfig = provider.db.Rebind(provider.sqlSelectTOTPConfig)
	provider.sqlUpdateTOTPConfigRecordSignIn = provider.db.Rebind(provider.sqlUpdateTOTPConfigRecordSignIn)
	provider.sqlUpdateTOTPConfigRecordSignInByUsername = provider.db.Rebind(provider.sqlUpdateTOTPConfigRecordSignInByUsername)
	provider.sqlDeleteTOTPConfig = provider.db.Rebind(provider.sqlDeleteTOTPConfig)
	provider.sqlSelectTOTPConfigs = provider.db.Rebind(provider.sqlSelectTOTPConfigs)
	provider.sqlUpdateTOTPConfigSecret = provider.db.Rebind(provider.sqlUpdateTOTPConfigSecret)
	provider.sqlUpdateTOTPConfigSecretByUsername = provider.db.Rebind(provider.sqlUpdateTOTPConfigSecretByUsername)

	provider.sqlSelectWebauthnDevices = provider.db.Rebind(provider.sqlSelectWebauthnDevices)
	provider.sqlSelectWebauthnDevicesByUsername = provider.db.Rebind(provider.sqlSelectWebauthnDevicesByUsername)
	provider.sqlUpdateWebauthnDevicePublicKey = provider.db.Rebind(provider.sqlUpdateWebauthnDevicePublicKey)
	provider.sqlUpdateWebauthnDevicePublicKeyByUsername = provider.db.Rebind(provider.sqlUpdateWebauthnDevicePublicKeyByUsername)
	provider.sqlUpdateWebauthnDeviceRecordSignIn = provider.db.Rebind(provider.sqlUpdateWebauthnDeviceRecordSignIn)
	provider.sqlUpdateWebauthnDeviceRecordSignInByUsername = provider.db.Rebind(provider.sqlUpdateWebauthnDeviceRecordSignInByUsername)
	provider.sqlDeleteWebauthnDevice = provider.db.Rebind(provider.sqlDeleteWebauthnDevice)
	provider.sqlDeleteWebauthnDeviceByUsername = provider.db.Rebind(provider.sqlDeleteWebauthnDeviceByUsername)
	provider.sqlDeleteWebauthnDeviceByUsernameAndDescription = provider.db.Rebind(provider.sqlDeleteWebauthnDeviceByUsernameAndDescription)

	provider.sqlSelectDuoDevice = provider.db.Rebind(provider.sqlSelectDuoDevice)
	provider.sqlDeleteDuoDevice = provider.db.Rebind(provider.sqlDeleteDuoDevice)

	provider.sqlInsertAuthenticationAttempt = provider.db.Rebind(provider.sqlInsertAuthenticationAttempt)
	provider.sqlSelectAuthenticationAttemptsByUsername = provider.db.Rebind(provider.sqlSelectAuthenticationAttemptsByUsername)

	provider.sqlInsertMigration = provider.db.Rebind(provider.sqlInsertMigration)
	provider.sqlSelectMigrations = provider.db.Rebind(provider.sqlSelectMigrations)
	provider.sqlSelectLatestMigration = provider.db.Rebind(provider.sqlSelectLatestMigration)

	provider.sqlSelectEncryptionValue = provider.db.Rebind(provider.sqlSelectEncryptionValue)

	provider.sqlSelectOAuth2ConsentPreConfigurations = provider.db.Rebind(provider.sqlSelectOAuth2ConsentPreConfigurations)

	provider.sqlInsertOAuth2ConsentSession = provider.db.Rebind(provider.sqlInsertOAuth2ConsentSession)
	provider.sqlUpdateOAuth2ConsentSessionSubject = provider.db.Rebind(provider.sqlUpdateOAuth2ConsentSessionSubject)
	provider.sqlUpdateOAuth2ConsentSessionResponse = provider.db.Rebind(provider.sqlUpdateOAuth2ConsentSessionResponse)
	provider.sqlUpdateOAuth2ConsentSessionGranted = provider.db.Rebind(provider.sqlUpdateOAuth2ConsentSessionGranted)
	provider.sqlSelectOAuth2ConsentSessionByChallengeID = provider.db.Rebind(provider.sqlSelectOAuth2ConsentSessionByChallengeID)

	provider.sqlInsertOAuth2AuthorizeCodeSession = provider.db.Rebind(provider.sqlInsertOAuth2AuthorizeCodeSession)
	provider.sqlRevokeOAuth2AuthorizeCodeSession = provider.db.Rebind(provider.sqlRevokeOAuth2AuthorizeCodeSession)
	provider.sqlRevokeOAuth2AuthorizeCodeSessionByRequestID = provider.db.Rebind(provider.sqlRevokeOAuth2AuthorizeCodeSessionByRequestID)
	provider.sqlDeactivateOAuth2AuthorizeCodeSession = provider.db.Rebind(provider.sqlDeactivateOAuth2AuthorizeCodeSession)
	provider.sqlDeactivateOAuth2AuthorizeCodeSessionByRequestID = provider.db.Rebind(provider.sqlDeactivateOAuth2AuthorizeCodeSessionByRequestID)
	provider.sqlSelectOAuth2AuthorizeCodeSession = provider.db.Rebind(provider.sqlSelectOAuth2AuthorizeCodeSession)

	provider.sqlInsertOAuth2AccessTokenSession = provider.db.Rebind(provider.sqlInsertOAuth2AccessTokenSession)
	provider.sqlRevokeOAuth2AccessTokenSession = provider.db.Rebind(provider.sqlRevokeOAuth2AccessTokenSession)
	provider.sqlRevokeOAuth2AccessTokenSessionByRequestID = provider.db.Rebind(provider.sqlRevokeOAuth2AccessTokenSessionByRequestID)
	provider.sqlDeactivateOAuth2AccessTokenSession = provider.db.Rebind(provider.sqlDeactivateOAuth2AccessTokenSession)
	provider.sqlDeactivateOAuth2AccessTokenSessionByRequestID = provider.db.Rebind(provider.sqlDeactivateOAuth2AccessTokenSessionByRequestID)
	provider.sqlSelectOAuth2AccessTokenSession = provider.db.Rebind(provider.sqlSelectOAuth2AccessTokenSession)

	provider.sqlInsertOAuth2RefreshTokenSession = provider.db.Rebind(provider.sqlInsertOAuth2RefreshTokenSession)
	provider.sqlRevokeOAuth2RefreshTokenSession = provider.db.Rebind(provider.sqlRevokeOAuth2RefreshTokenSession)
	provider.sqlRevokeOAuth2RefreshTokenSessionByRequestID = provider.db.Rebind(provider.sqlRevokeOAuth2RefreshTokenSessionByRequestID)
	provider.sqlDeactivateOAuth2RefreshTokenSession = provider.db.Rebind(provider.sqlDeactivateOAuth2RefreshTokenSession)
	provider.sqlDeactivateOAuth2RefreshTokenSessionByRequestID = provider.db.Rebind(provider.sqlDeactivateOAuth2RefreshTokenSessionByRequestID)
	provider.sqlSelectOAuth2RefreshTokenSession = provider.db.Rebind(provider.sqlSelectOAuth2RefreshTokenSession)

	provider.sqlInsertOAuth2PKCERequestSession = provider.db.Rebind(provider.sqlInsertOAuth2PKCERequestSession)
	provider.sqlRevokeOAuth2PKCERequestSession = provider.db.Rebind(provider.sqlRevokeOAuth2PKCERequestSession)
	provider.sqlRevokeOAuth2PKCERequestSessionByRequestID = provider.db.Rebind(provider.sqlRevokeOAuth2PKCERequestSessionByRequestID)
	provider.sqlDeactivateOAuth2PKCERequestSession = provider.db.Rebind(provider.sqlDeactivateOAuth2PKCERequestSession)
	provider.sqlDeactivateOAuth2PKCERequestSessionByRequestID = provider.db.Rebind(provider.sqlDeactivateOAuth2PKCERequestSessionByRequestID)
	provider.sqlSelectOAuth2PKCERequestSession = provider.db.Rebind(provider.sqlSelectOAuth2PKCERequestSession)

	provider.sqlInsertOAuth2OpenIDConnectSession = provider.db.Rebind(provider.sqlInsertOAuth2OpenIDConnectSession)
	provider.sqlRevokeOAuth2OpenIDConnectSession = provider.db.Rebind(provider.sqlRevokeOAuth2OpenIDConnectSession)
	provider.sqlRevokeOAuth2OpenIDConnectSessionByRequestID = provider.db.Rebind(provider.sqlRevokeOAuth2OpenIDConnectSessionByRequestID)
	provider.sqlDeactivateOAuth2OpenIDConnectSession = provider.db.Rebind(provider.sqlDeactivateOAuth2OpenIDConnectSession)
	provider.sqlDeactivateOAuth2OpenIDConnectSessionByRequestID = provider.db.Rebind(provider.sqlDeactivateOAuth2OpenIDConnectSessionByRequestID)
	provider.sqlSelectOAuth2OpenIDConnectSession = provider.db.Rebind(provider.sqlSelectOAuth2OpenIDConnectSession)

	provider.sqlSelectOAuth2BlacklistedJTI = provider.db.Rebind(provider.sqlSelectOAuth2BlacklistedJTI)

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
