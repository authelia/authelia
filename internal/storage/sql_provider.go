package storage

import (
	"context"
	"crypto/sha256"
	"crypto/x509"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
)

// NewProvider dynamically initializes a storage.Provider given a *schema.Configuration and *x509.CertPool.
func NewProvider(config *schema.Configuration, caCertPool *x509.CertPool) (provider Provider) {
	switch {
	case config.Storage.PostgreSQL != nil:
		return NewPostgreSQLProvider(config, caCertPool)
	case config.Storage.MySQL != nil:
		return NewMySQLProvider(config, caCertPool)
	case config.Storage.Local != nil:
		return NewSQLiteProvider(config)
	default:
		return nil
	}
}

// NewSQLProvider generates a generic SQLProvider to be used with other SQL provider NewUp's.
func NewSQLProvider(config *schema.Configuration, name, driverName, dataSourceName string) (provider SQLProvider) {
	db, err := sqlx.Open(driverName, dataSourceName)

	provider = SQLProvider{
		db:         db,
		name:       name,
		driverName: driverName,
		config:     config,
		errOpen:    err,

		keys: SQLProviderKeys{
			encryption: sha256.Sum256([]byte(config.Storage.EncryptionKey)),
		},

		log: logging.Logger(),

		sqlInsertAuthenticationAttempt:                         fmt.Sprintf(queryFmtInsertAuthenticationLogEntry, tableAuthenticationLogs),
		sqlSelectAuthenticationLogsRegulationRecordsByUsername: fmt.Sprintf(queryFmtSelectAuthenticationLogsRegulationRecordsByUsername, tableAuthenticationLogs),
		sqlSelectAuthenticationLogsRegulationRecordsByRemoteIP: fmt.Sprintf(queryFmtSelectAuthenticationLogsRegulationRecordsByRemoteIP, tableAuthenticationLogs),

		sqlInsertBannedUser:         fmt.Sprintf(queryFmtInsertBannedUser, tableBannedUser),
		sqlSelectBannedUser:         fmt.Sprintf(queryFmtSelectBannedUser, tableBannedUser),
		sqlSelectBannedUserByID:     fmt.Sprintf(queryFmtSelectBannedUserByID, tableBannedUser),
		sqlSelectBannedUsers:        fmt.Sprintf(queryFmtSelectBannedUsers, tableBannedUser),
		sqlSelectBannedUserLastTime: fmt.Sprintf(queryFmtSelectBannedUserLastExpires, tableBannedUser),
		sqlRevokeBannedUser:         fmt.Sprintf(queryFmtRevokeBannedEntry, tableBannedUser),

		sqlInsertBannedIP:         fmt.Sprintf(queryFmtInsertBannedIP, tableBannedIP),
		sqlSelectBannedIP:         fmt.Sprintf(queryFmtSelectBannedIP, tableBannedIP),
		sqlSelectBannedIPByID:     fmt.Sprintf(queryFmtSelectBannedIPByID, tableBannedIP),
		sqlSelectBannedIPs:        fmt.Sprintf(queryFmtSelectBannedIPs, tableBannedIP),
		sqlSelectBannedIPLastTime: fmt.Sprintf(queryFmtSelectBannedIPLastExpires, tableBannedIP),
		sqlRevokeBannedIP:         fmt.Sprintf(queryFmtRevokeBannedEntry, tableBannedIP),

		sqlUpsertCachedData: fmt.Sprintf(queryFmtUpsertCachedData, tableCachedData),
		sqlSelectCachedData: fmt.Sprintf(queryFmtSelectCachedData, tableCachedData),
		sqlDeleteCachedData: fmt.Sprintf(queryFmtDeleteCachedData, tableCachedData),

		sqlInsertUserMetadata:                          fmt.Sprintf(queryFmtInsertUserMetadata, tableUserMetadata),
		sqlInsertNewUserMetadata:                       fmt.Sprintf(queryFmtInsertNewUserMetadata, tableUserMetadata),
		sqlInsertExistingUserAtLoginMetadata:           fmt.Sprintf(queryFmtInsertNewUserAtLoginMetadata, tableUserMetadata),
		sqlInsertExistingUserAtPasswordChangeMetadata:  fmt.Sprintf(queryFmtInsertExistingUserAtPasswordChangeMetadata, tableUserMetadata),
		sqlSelectUserMetadata:                          fmt.Sprintf(queryFmtSelectUserMetadata, tableUserMetadata),
		sqlSelectMultipleUserMetadataByUsername:        fmt.Sprintf(queryFmtSelectMultipleUserMetadataByUsername, tableUserMetadata),
		sqlSelectUserMetadataByUsername:                fmt.Sprintf(queryFmtSelectUserByUsername, tableUserMetadata),
		sqlSelectAllUserInfoAndMetadata:                fmt.Sprintf(queryFmtSelectAllUserInfoAndMetadata, tableUserMetadata, tableUserPreferences, tableTOTPConfigurations, tableWebAuthnCredentials, tableDuoDevices),
		sqlUpdateUserRecordSignInByUsername:            fmt.Sprintf(queryFmtUpdateUserRecordSignInByUsername, tableUserMetadata),
		sqlDeleteUserMetadataByUsername:                fmt.Sprintf(queryFmtDeleteUserMetadataByUsername, tableUserMetadata),
		sqlUpdateUserRecordPasswordChangedAtByUsername: fmt.Sprintf(queryFmtUpdateUserRecordPasswordChangedAtByUsername, tableUserMetadata),

		sqlInsertIdentityVerification:  fmt.Sprintf(queryFmtInsertIdentityVerification, tableIdentityVerification),
		sqlConsumeIdentityVerification: fmt.Sprintf(queryFmtConsumeIdentityVerification, tableIdentityVerification),
		sqlRevokeIdentityVerification:  fmt.Sprintf(queryFmtRevokeIdentityVerification, tableIdentityVerification),
		sqlSelectIdentityVerification:  fmt.Sprintf(queryFmtSelectIdentityVerification, tableIdentityVerification),

		sqlInsertOneTimeCode:            fmt.Sprintf(queryFmtInsertOTC, tableOneTimeCode),
		sqlConsumeOneTimeCode:           fmt.Sprintf(queryFmtConsumeOTC, tableOneTimeCode),
		sqlRevokeOneTimeCode:            fmt.Sprintf(queryFmtRevokeOTC, tableOneTimeCode),
		sqlSelectOneTimeCode:            fmt.Sprintf(queryFmtSelectOTCBySignatureAndUsername, tableOneTimeCode),
		sqlSelectOneTimeCodeBySignature: fmt.Sprintf(queryFmtSelectOTCBySignature, tableOneTimeCode),
		sqlSelectOneTimeCodeByID:        fmt.Sprintf(queryFmtSelectOTCByID, tableOneTimeCode),
		sqlSelectOneTimeCodeByPublicID:  fmt.Sprintf(queryFmtSelectOTCByPublicID, tableOneTimeCode),

		sqlUpsertTOTPConfig:  fmt.Sprintf(queryFmtUpsertTOTPConfiguration, tableTOTPConfigurations),
		sqlDeleteTOTPConfig:  fmt.Sprintf(queryFmtDeleteTOTPConfiguration, tableTOTPConfigurations),
		sqlSelectTOTPConfig:  fmt.Sprintf(queryFmtSelectTOTPConfiguration, tableTOTPConfigurations),
		sqlSelectTOTPConfigs: fmt.Sprintf(queryFmtSelectTOTPConfigurations, tableTOTPConfigurations),

		sqlUpdateTOTPConfigRecordSignIn:           fmt.Sprintf(queryFmtUpdateTOTPConfigRecordSignIn, tableTOTPConfigurations),
		sqlUpdateTOTPConfigRecordSignInByUsername: fmt.Sprintf(queryFmtUpdateTOTPConfigRecordSignInByUsername, tableTOTPConfigurations),

		sqlInsertTOTPHistory: fmt.Sprintf(queryFmtInsertTOTPHistory, tableTOTPHistory),
		sqlSelectTOTPHistory: fmt.Sprintf(queryFmtSelectTOTPHistory, tableTOTPHistory),

		sqlInsertWebAuthnUser:         fmt.Sprintf(queryFmtInsertWebAuthnUser, tableWebAuthnUsers),
		sqlSelectWebAuthnUser:         fmt.Sprintf(queryFmtSelectWebAuthnUser, tableWebAuthnUsers),
		sqlSelectWebAuthnUserByUserID: fmt.Sprintf(queryFmtSelectWebAuthnUserByUserID, tableWebAuthnUsers),

		sqlInsertWebAuthnCredential:                           fmt.Sprintf(queryFmtInsertWebAuthnCredential, tableWebAuthnCredentials),
		sqlSelectWebAuthnCredentials:                          fmt.Sprintf(queryFmtSelectWebAuthnCredentials, tableWebAuthnCredentials),
		sqlSelectWebAuthnCredentialsByUsername:                fmt.Sprintf(queryFmtSelectWebAuthnCredentialsByUsername, tableWebAuthnCredentials),
		sqlSelectWebAuthnCredentialsByRPIDByUsername:          fmt.Sprintf(queryFmtSelectWebAuthnCredentialsByRPIDByUsername, tableWebAuthnCredentials),
		sqlSelectWebAuthnCredentialByID:                       fmt.Sprintf(queryFmtSelectWebAuthnCredentialByID, tableWebAuthnCredentials),
		sqlUpdateWebAuthnCredentialDescriptionByUsernameAndID: fmt.Sprintf(queryFmtUpdateUpdateWebAuthnCredentialDescriptionByUsernameAndID, tableWebAuthnCredentials),
		sqlUpdateWebAuthnCredentialRecordSignIn:               fmt.Sprintf(queryFmtUpdateWebAuthnCredentialRecordSignIn, tableWebAuthnCredentials),
		sqlDeleteWebAuthnCredential:                           fmt.Sprintf(queryFmtDeleteWebAuthnCredential, tableWebAuthnCredentials),
		sqlDeleteWebAuthnCredentialByUsername:                 fmt.Sprintf(queryFmtDeleteWebAuthnCredentialByUsername, tableWebAuthnCredentials),
		sqlDeleteWebAuthnCredentialByUsernameAndDisplayName:   fmt.Sprintf(queryFmtDeleteWebAuthnCredentialByUsernameAndDescription, tableWebAuthnCredentials),

		sqlUpsertDuoDevice: fmt.Sprintf(queryFmtUpsertDuoDevice, tableDuoDevices),
		sqlDeleteDuoDevice: fmt.Sprintf(queryFmtDeleteDuoDevice, tableDuoDevices),
		sqlSelectDuoDevice: fmt.Sprintf(queryFmtSelectDuoDevice, tableDuoDevices),

		sqlUpsertPreferred2FAMethod: fmt.Sprintf(queryFmtUpsertPreferred2FAMethod, tableUserPreferences),
		sqlSelectPreferred2FAMethod: fmt.Sprintf(queryFmtSelectPreferred2FAMethod, tableUserPreferences),
		sqlSelectUserInfo:           fmt.Sprintf(queryFmtSelectUserInfoByUsername, tableTOTPConfigurations, tableWebAuthnCredentials, tableDuoDevices, tableUserPreferences),
		sqlSelectAllUserInfo:        fmt.Sprintf(queryFmtSelectAllUserInfo, tableUserPreferences, tableTOTPConfigurations, tableWebAuthnCredentials, tableDuoDevices),

		sqlInsertUserOpaqueIdentifier:            fmt.Sprintf(queryFmtInsertUserOpaqueIdentifier, tableUserOpaqueIdentifier),
		sqlSelectUserOpaqueIdentifier:            fmt.Sprintf(queryFmtSelectUserOpaqueIdentifier, tableUserOpaqueIdentifier),
		sqlSelectUserOpaqueIdentifiers:           fmt.Sprintf(queryFmtSelectUserOpaqueIdentifiers, tableUserOpaqueIdentifier),
		sqlSelectUserOpaqueIdentifierBySignature: fmt.Sprintf(queryFmtSelectUserOpaqueIdentifierBySignature, tableUserOpaqueIdentifier),

		sqlUpsertOAuth2BlacklistedJTI: fmt.Sprintf(queryFmtUpsertOAuth2BlacklistedJTI, tableOAuth2BlacklistedJTI),
		sqlSelectOAuth2BlacklistedJTI: fmt.Sprintf(queryFmtSelectOAuth2BlacklistedJTI, tableOAuth2BlacklistedJTI),

		sqlInsertOAuth2PARContext: fmt.Sprintf(queryFmtInsertOAuth2PARContext, tableOAuth2PARContext),
		sqlUpdateOAuth2PARContext: fmt.Sprintf(queryFmtUpdateOAuth2PARContext, tableOAuth2PARContext),
		sqlSelectOAuth2PARContext: fmt.Sprintf(queryFmtSelectOAuth2PARContext, tableOAuth2PARContext),
		sqlRevokeOAuth2PARContext: fmt.Sprintf(queryFmtRevokeOAuth2Session, tableOAuth2PARContext),

		sqlInsertOAuth2ConsentPreConfiguration:  fmt.Sprintf(queryFmtInsertOAuth2ConsentPreConfiguration, tableOAuth2ConsentPreConfiguration),
		sqlSelectOAuth2ConsentPreConfigurations: fmt.Sprintf(queryFmtSelectOAuth2ConsentPreConfigurations, tableOAuth2ConsentPreConfiguration),

		sqlInsertOAuth2ConsentSession:              fmt.Sprintf(queryFmtInsertOAuth2ConsentSession, tableOAuth2ConsentSession),
		sqlUpdateOAuth2ConsentSessionResponse:      fmt.Sprintf(queryFmtUpdateOAuth2ConsentSessionResponse, tableOAuth2ConsentSession),
		sqlUpdateOAuth2ConsentSessionGranted:       fmt.Sprintf(queryFmtUpdateOAuth2ConsentSessionGranted, tableOAuth2ConsentSession),
		sqlSelectOAuth2ConsentSessionByChallengeID: fmt.Sprintf(queryFmtSelectOAuth2ConsentSessionByChallengeID, tableOAuth2ConsentSession),

		sqlInsertOAuth2AccessTokenSession:                fmt.Sprintf(queryFmtInsertOAuth2Session, tableOAuth2AccessTokenSession),
		sqlSelectOAuth2AccessTokenSession:                fmt.Sprintf(queryFmtSelectOAuth2Session, tableOAuth2AccessTokenSession),
		sqlRevokeOAuth2AccessTokenSession:                fmt.Sprintf(queryFmtRevokeOAuth2Session, tableOAuth2AccessTokenSession),
		sqlRevokeOAuth2AccessTokenSessionByRequestID:     fmt.Sprintf(queryFmtRevokeOAuth2SessionByRequestID, tableOAuth2AccessTokenSession),
		sqlDeactivateOAuth2AccessTokenSession:            fmt.Sprintf(queryFmtDeactivateOAuth2Session, tableOAuth2AccessTokenSession),
		sqlDeactivateOAuth2AccessTokenSessionByRequestID: fmt.Sprintf(queryFmtDeactivateOAuth2SessionByRequestID, tableOAuth2AccessTokenSession),

		sqlInsertOAuth2AuthorizeCodeSession:                fmt.Sprintf(queryFmtInsertOAuth2Session, tableOAuth2AuthorizeCodeSession),
		sqlSelectOAuth2AuthorizeCodeSession:                fmt.Sprintf(queryFmtSelectOAuth2Session, tableOAuth2AuthorizeCodeSession),
		sqlRevokeOAuth2AuthorizeCodeSession:                fmt.Sprintf(queryFmtRevokeOAuth2Session, tableOAuth2AuthorizeCodeSession),
		sqlRevokeOAuth2AuthorizeCodeSessionByRequestID:     fmt.Sprintf(queryFmtRevokeOAuth2SessionByRequestID, tableOAuth2AuthorizeCodeSession),
		sqlDeactivateOAuth2AuthorizeCodeSession:            fmt.Sprintf(queryFmtDeactivateOAuth2Session, tableOAuth2AuthorizeCodeSession),
		sqlDeactivateOAuth2AuthorizeCodeSessionByRequestID: fmt.Sprintf(queryFmtDeactivateOAuth2SessionByRequestID, tableOAuth2AuthorizeCodeSession),

		sqlInsertOAuth2DeviceCodeSession:           fmt.Sprintf(queryFmtInsertOAuth2DeviceCodeSession, tableOAuth2DeviceCodeSession),
		sqlSelectOAuth2DeviceCodeSession:           fmt.Sprintf(queryFmtSelectOAuth2DeviceCodeSession, tableOAuth2DeviceCodeSession),
		sqlUpdateOAuth2DeviceCodeSession:           fmt.Sprintf(queryFmtUpdateOAuth2DeviceCodeSession, tableOAuth2DeviceCodeSession),
		sqlUpdateOAuth2DeviceCodeSessionData:       fmt.Sprintf(queryFmtUpdateOAuth2DeviceCodeSessionData, tableOAuth2DeviceCodeSession),
		sqlDeactivateOAuth2DeviceCodeSession:       fmt.Sprintf(queryFmtDeactivateOAuth2Session, tableOAuth2DeviceCodeSession),
		sqlSelectOAuth2DeviceCodeSessionByUserCode: fmt.Sprintf(queryFmtSelectOAuth2DeviceCodeSessionByUserCode, tableOAuth2DeviceCodeSession),

		sqlInsertOAuth2OpenIDConnectSession:                fmt.Sprintf(queryFmtInsertOAuth2Session, tableOAuth2OpenIDConnectSession),
		sqlSelectOAuth2OpenIDConnectSession:                fmt.Sprintf(queryFmtSelectOAuth2Session, tableOAuth2OpenIDConnectSession),
		sqlRevokeOAuth2OpenIDConnectSession:                fmt.Sprintf(queryFmtRevokeOAuth2Session, tableOAuth2OpenIDConnectSession),
		sqlRevokeOAuth2OpenIDConnectSessionByRequestID:     fmt.Sprintf(queryFmtRevokeOAuth2SessionByRequestID, tableOAuth2OpenIDConnectSession),
		sqlDeactivateOAuth2OpenIDConnectSession:            fmt.Sprintf(queryFmtDeactivateOAuth2Session, tableOAuth2OpenIDConnectSession),
		sqlDeactivateOAuth2OpenIDConnectSessionByRequestID: fmt.Sprintf(queryFmtDeactivateOAuth2SessionByRequestID, tableOAuth2OpenIDConnectSession),

		sqlInsertOAuth2PKCERequestSession:                fmt.Sprintf(queryFmtInsertOAuth2Session, tableOAuth2PKCERequestSession),
		sqlSelectOAuth2PKCERequestSession:                fmt.Sprintf(queryFmtSelectOAuth2Session, tableOAuth2PKCERequestSession),
		sqlRevokeOAuth2PKCERequestSession:                fmt.Sprintf(queryFmtRevokeOAuth2Session, tableOAuth2PKCERequestSession),
		sqlRevokeOAuth2PKCERequestSessionByRequestID:     fmt.Sprintf(queryFmtRevokeOAuth2SessionByRequestID, tableOAuth2PKCERequestSession),
		sqlDeactivateOAuth2PKCERequestSession:            fmt.Sprintf(queryFmtDeactivateOAuth2Session, tableOAuth2PKCERequestSession),
		sqlDeactivateOAuth2PKCERequestSessionByRequestID: fmt.Sprintf(queryFmtDeactivateOAuth2SessionByRequestID, tableOAuth2PKCERequestSession),

		sqlInsertOAuth2RefreshTokenSession:                fmt.Sprintf(queryFmtInsertOAuth2Session, tableOAuth2RefreshTokenSession),
		sqlSelectOAuth2RefreshTokenSession:                fmt.Sprintf(queryFmtSelectOAuth2Session, tableOAuth2RefreshTokenSession),
		sqlRevokeOAuth2RefreshTokenSession:                fmt.Sprintf(queryFmtRevokeOAuth2Session, tableOAuth2RefreshTokenSession),
		sqlRevokeOAuth2RefreshTokenSessionByRequestID:     fmt.Sprintf(queryFmtRevokeOAuth2SessionByRequestID, tableOAuth2RefreshTokenSession),
		sqlDeactivateOAuth2RefreshTokenSession:            fmt.Sprintf(queryFmtDeactivateOAuth2Session, tableOAuth2RefreshTokenSession),
		sqlDeactivateOAuth2RefreshTokenSessionByRequestID: fmt.Sprintf(queryFmtDeactivateOAuth2SessionByRequestID, tableOAuth2RefreshTokenSession),

		sqlInsertMigration:       fmt.Sprintf(queryFmtInsertMigration, tableMigrations),
		sqlSelectMigrations:      fmt.Sprintf(queryFmtSelectMigrations, tableMigrations),
		sqlSelectLatestMigration: fmt.Sprintf(queryFmtSelectLatestMigration, tableMigrations),

		sqlUpsertEncryptionValue: fmt.Sprintf(queryFmtUpsertEncryptionValue, tableEncryption),
		sqlSelectEncryptionValue: fmt.Sprintf(queryFmtSelectEncryptionValue, tableEncryption),

		sqlFmtRenameTable: queryFmtRenameTable,
	}

	return provider
}

// SQLProvider is a storage provider persisting data in a SQL database.
type SQLProvider struct {
	db *sqlx.DB

	name       string
	driverName string
	schema     string
	config     *schema.Configuration
	errOpen    error

	keys SQLProviderKeys

	log *logrus.Logger

	// Table: authentication_logs.
	sqlInsertAuthenticationAttempt                         string
	sqlSelectAuthenticationLogsRegulationRecordsByUsername string
	sqlSelectAuthenticationLogsRegulationRecordsByRemoteIP string

	// Table: banned_user.
	sqlInsertBannedUser         string
	sqlSelectBannedUser         string
	sqlSelectBannedUserByID     string
	sqlSelectBannedUsers        string
	sqlSelectBannedUserLastTime string
	sqlRevokeBannedUser         string

	// Table: banned_ip.
	sqlInsertBannedIP         string
	sqlSelectBannedIP         string
	sqlSelectBannedIPByID     string
	sqlSelectBannedIPs        string
	sqlSelectBannedIPLastTime string
	sqlRevokeBannedIP         string

	// Table: cached_data.
	sqlUpsertCachedData string
	sqlSelectCachedData string
	sqlDeleteCachedData string

	// Table: user_Metadata.
	sqlInsertUserMetadata                               string
	sqlInsertNewUserMetadata                            string
	sqlInsertExistingUserAtLoginMetadata                string
	sqlInsertExistingUserAtPasswordChangeMetadata       string
	sqlSelectUserMetadata                               string
	sqlSelectMultipleUserMetadataByUsername             string
	sqlSelectUserMetadataByUsername                     string
	sqlSelectAllUserInfoAndMetadata                     string
	sqlUpdateUserRecordSignInByUsername                 string
	sqlDeleteUserMetadataByUsername                     string
	sqlUpdateUserRecordPasswordChangedAtByUsername      string
	sqlUpdateUserRecordRequirePasswordChangedByUsername string

	// Table: identity_verification.
	sqlInsertIdentityVerification  string
	sqlConsumeIdentityVerification string
	sqlRevokeIdentityVerification  string
	sqlSelectIdentityVerification  string

	// Table: one_time_code.
	sqlInsertOneTimeCode            string
	sqlConsumeOneTimeCode           string
	sqlRevokeOneTimeCode            string
	sqlSelectOneTimeCode            string
	sqlSelectOneTimeCodeBySignature string
	sqlSelectOneTimeCodeByID        string
	sqlSelectOneTimeCodeByPublicID  string

	// Table: totp_configurations.
	sqlUpsertTOTPConfig  string
	sqlDeleteTOTPConfig  string
	sqlSelectTOTPConfig  string
	sqlSelectTOTPConfigs string

	sqlUpdateTOTPConfigRecordSignIn           string
	sqlUpdateTOTPConfigRecordSignInByUsername string

	// Table: totp_history.
	sqlInsertTOTPHistory string
	sqlSelectTOTPHistory string

	// Table: webauthn_users.
	sqlInsertWebAuthnUser         string
	sqlSelectWebAuthnUser         string
	sqlSelectWebAuthnUserByUserID string

	// Table: webauthn_credentials.
	sqlInsertWebAuthnCredential                  string
	sqlSelectWebAuthnCredentials                 string
	sqlSelectWebAuthnCredentialsByUsername       string
	sqlSelectWebAuthnCredentialsByRPIDByUsername string
	sqlSelectWebAuthnCredentialByID              string

	sqlUpdateWebAuthnCredentialDescriptionByUsernameAndID string
	sqlUpdateWebAuthnCredentialRecordSignIn               string

	sqlDeleteWebAuthnCredential                         string
	sqlDeleteWebAuthnCredentialByUsername               string
	sqlDeleteWebAuthnCredentialByUsernameAndDisplayName string

	// Table: duo_devices.
	sqlUpsertDuoDevice string
	sqlDeleteDuoDevice string
	sqlSelectDuoDevice string

	// Table: user_preferences.
	sqlUpsertPreferred2FAMethod string
	sqlSelectPreferred2FAMethod string
	sqlSelectUserInfo           string
	sqlSelectAllUserInfo        string

	// Table: user_opaque_identifier.
	sqlInsertUserOpaqueIdentifier            string
	sqlSelectUserOpaqueIdentifier            string
	sqlSelectUserOpaqueIdentifiers           string
	sqlSelectUserOpaqueIdentifierBySignature string

	// Table: migrations.
	sqlInsertMigration       string
	sqlSelectMigrations      string
	sqlSelectLatestMigration string

	// Table: encryption.
	sqlUpsertEncryptionValue string
	sqlSelectEncryptionValue string

	// Table: oauth2_consent_preconfiguration.
	sqlInsertOAuth2ConsentPreConfiguration  string
	sqlSelectOAuth2ConsentPreConfigurations string

	// Table: oauth2_consent_session.
	sqlInsertOAuth2ConsentSession              string
	sqlUpdateOAuth2ConsentSessionResponse      string
	sqlUpdateOAuth2ConsentSessionGranted       string
	sqlSelectOAuth2ConsentSessionByChallengeID string

	// Table: oauth2_authorization_code_session.
	sqlInsertOAuth2AuthorizeCodeSession                string
	sqlSelectOAuth2AuthorizeCodeSession                string
	sqlRevokeOAuth2AuthorizeCodeSession                string
	sqlRevokeOAuth2AuthorizeCodeSessionByRequestID     string
	sqlDeactivateOAuth2AuthorizeCodeSession            string
	sqlDeactivateOAuth2AuthorizeCodeSessionByRequestID string

	// Table: oauth2_device_code_session.
	sqlInsertOAuth2DeviceCodeSession           string
	sqlSelectOAuth2DeviceCodeSession           string
	sqlUpdateOAuth2DeviceCodeSession           string
	sqlUpdateOAuth2DeviceCodeSessionData       string
	sqlDeactivateOAuth2DeviceCodeSession       string
	sqlSelectOAuth2DeviceCodeSessionByUserCode string

	// Table: oauth2_access_token_session.
	sqlInsertOAuth2AccessTokenSession                string
	sqlSelectOAuth2AccessTokenSession                string
	sqlRevokeOAuth2AccessTokenSession                string
	sqlRevokeOAuth2AccessTokenSessionByRequestID     string
	sqlDeactivateOAuth2AccessTokenSession            string
	sqlDeactivateOAuth2AccessTokenSessionByRequestID string

	// Table: oauth2_openid_connect_session.
	sqlInsertOAuth2OpenIDConnectSession                string
	sqlSelectOAuth2OpenIDConnectSession                string
	sqlRevokeOAuth2OpenIDConnectSession                string
	sqlRevokeOAuth2OpenIDConnectSessionByRequestID     string
	sqlDeactivateOAuth2OpenIDConnectSession            string
	sqlDeactivateOAuth2OpenIDConnectSessionByRequestID string

	// Table: oauth2_par_context.
	sqlInsertOAuth2PARContext string
	sqlUpdateOAuth2PARContext string
	sqlSelectOAuth2PARContext string
	sqlRevokeOAuth2PARContext string

	// Table: oauth2_pkce_request_session.
	sqlInsertOAuth2PKCERequestSession                string
	sqlSelectOAuth2PKCERequestSession                string
	sqlRevokeOAuth2PKCERequestSession                string
	sqlRevokeOAuth2PKCERequestSessionByRequestID     string
	sqlDeactivateOAuth2PKCERequestSession            string
	sqlDeactivateOAuth2PKCERequestSessionByRequestID string

	// Table: oauth2_refresh_token_session.
	sqlInsertOAuth2RefreshTokenSession                string
	sqlSelectOAuth2RefreshTokenSession                string
	sqlRevokeOAuth2RefreshTokenSession                string
	sqlRevokeOAuth2RefreshTokenSessionByRequestID     string
	sqlDeactivateOAuth2RefreshTokenSession            string
	sqlDeactivateOAuth2RefreshTokenSessionByRequestID string

	sqlUpsertOAuth2BlacklistedJTI string
	sqlSelectOAuth2BlacklistedJTI string

	// Utility.
	sqlSelectExistingTables string
	sqlFmtRenameTable       string
}

// SQLProviderKeys are the cryptography keys used by a SQLProvider.
type SQLProviderKeys struct {
	encryption [32]byte
	otcHMAC    []byte
	otpHMAC    []byte
}

// StartupCheck implements the provider startup check interface.
func (p *SQLProvider) StartupCheck() (err error) {
	if p.errOpen != nil {
		return fmt.Errorf("error opening database: %w", p.errOpen)
	}

	// TODO: Decide if this is needed, or if it should be configurable.
	for i := 0; i < 19; i++ {
		if err = p.db.Ping(); err == nil {
			break
		}

		time.Sleep(time.Millisecond * 500)
	}

	if err != nil {
		return fmt.Errorf("error pinging database: %w", err)
	}

	p.log.Infof("Storage schema is being checked for updates")

	ctx := context.Background()

	var result EncryptionValidationResult

	if result, err = p.SchemaEncryptionCheckKey(ctx, false); err != nil && !errors.Is(err, ErrSchemaEncryptionVersionUnsupported) {
		return err
	}

	if !result.Success() {
		return ErrSchemaEncryptionInvalidKey
	}

	switch err = p.SchemaMigrate(ctx, true, SchemaLatest); err {
	case nil:
		break
	case ErrSchemaAlreadyUpToDate:
		p.log.Infof("Storage schema is already up to date")
	default:
		return fmt.Errorf("error during schema migrate: %w", err)
	}

	if p.keys.otcHMAC, err = p.getHMACOneTimeCode(ctx); err != nil {
		return fmt.Errorf("failed to initialize the hmac one-time code signature key during startup: %w", err)
	}

	if p.keys.otpHMAC, err = p.getHMACOneTimePassword(ctx); err != nil {
		return fmt.Errorf("failed to initialize the hmac one-time password signature key during startup: %w", err)
	}

	return nil
}

// BeginTX begins a transaction with the storage provider when applicable.
func (p *SQLProvider) BeginTX(ctx context.Context) (c context.Context, err error) {
	var tx *sql.Tx

	if tx, err = p.db.Begin(); err != nil {
		return nil, err
	}

	return context.WithValue(ctx, ctxKeyTransaction, tx), nil
}

// Commit performs a storage provider commit when applicable.
func (p *SQLProvider) Commit(ctx context.Context) (err error) {
	tx, ok := ctx.Value(ctxKeyTransaction).(*sql.Tx)

	if !ok {
		return errors.New("could not retrieve tx")
	}

	return tx.Commit()
}

// Rollback performs a storage provider rollback when applicable.
func (p *SQLProvider) Rollback(ctx context.Context) (err error) {
	tx, ok := ctx.Value(ctxKeyTransaction).(*sql.Tx)

	if !ok {
		return errors.New("could not retrieve tx")
	}

	return tx.Rollback()
}

// Close the underlying storage provider.
func (p *SQLProvider) Close() (err error) {
	return p.db.Close()
}

// CreateExistingUserMetadata add an existing user's metadata record to the storage provider without setting creation date.
func (p *SQLProvider) CreateExistingUserMetadata(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertUserMetadata, username); err != nil {
		return fmt.Errorf("error creating exiting user Metadata for user '%s': %w", username, err)
	}

	return nil
}

// CreateNewUserMetadata adds a new user's metadata record to the storage provider along with the user's creation date.
func (p *SQLProvider) CreateNewUserMetadata(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertNewUserMetadata, username, time.Now()); err != nil {
		return fmt.Errorf("error creating new user Metadata for user '%s': %w", username, err)
	}

	return nil
}

// CreateExistingUserMetadataAtLogin adds a new user's attribute record to the storage provider, and sets their last login time.
func (p *SQLProvider) CreateExistingUserMetadataAtLogin(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertExistingUserAtLoginMetadata, username, time.Now()); err != nil {
		return fmt.Errorf("error creating new user Metadata for user '%s': %w", username, err)
	}

	return nil
}

func (p *SQLProvider) CreateExistingUserMetadataAtPasswordChange(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertExistingUserAtPasswordChangeMetadata, username, time.Now()); err != nil {
		return fmt.Errorf("error creating new user Metadata for user '%s': %w", username, err)
	}

	return nil
}

// LoadAllUsersMetadata load all user Metadata from the database.
func (p *SQLProvider) LoadAllUsersMetadata(ctx context.Context) (allUserMetadata []model.UserInfo, err error) {
	rows, err := p.db.QueryContext(ctx, p.sqlSelectUserMetadata)
	if err != nil {
		return nil, fmt.Errorf("error querying user Metadata: %w", err)
	}
	defer rows.Close()

	var users []model.UserInfo

	for rows.Next() {
		var user model.UserInfo

		var lastLoggedIn, lastPasswordChange, userCreatedAt sql.NullTime

		err := rows.Scan(
			&user.Username,
			&lastLoggedIn,
			&lastPasswordChange,
			&userCreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user Metadata: %w", err)
		}

		if lastLoggedIn.Valid {
			user.LastLoggedIn = &lastLoggedIn.Time
		}

		if lastPasswordChange.Valid {
			user.LastPasswordChange = &lastPasswordChange.Time
		}

		if userCreatedAt.Valid {
			user.UserCreatedAt = &userCreatedAt.Time
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user Metadata: %w", err)
	}

	return users, nil
}

// LoadAllUsersMetadata load all user Metadata from the database.
func (p *SQLProvider) LoadMultipleUsersMetadataByUsername(ctx context.Context, usernames []string) (allUserMetadata []model.UserInfo, err error) {
	rows, err := p.db.QueryContext(ctx, p.sqlSelectMultipleUserMetadataByUsername, usernames)
	if err != nil {
		return nil, fmt.Errorf("error querying user Metadata: %w", err)
	}
	defer rows.Close()

	var users []model.UserInfo

	for rows.Next() {
		var user model.UserInfo

		var lastLoggedIn, lastPasswordChange, userCreatedAt sql.NullTime

		err := rows.Scan(
			&user.Username,
			&lastLoggedIn,
			&lastPasswordChange,
			&userCreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning user Metadata: %w", err)
		}

		if lastLoggedIn.Valid {
			user.LastLoggedIn = &lastLoggedIn.Time
		}

		if lastPasswordChange.Valid {
			user.LastPasswordChange = &lastPasswordChange.Time
		}

		if userCreatedAt.Valid {
			user.UserCreatedAt = &userCreatedAt.Time
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user Metadata: %w", err)
	}

	return users, nil
}

// LoadUserMetadataByUsername load the user Metadata for a specific user.
func (p *SQLProvider) LoadUserMetadataByUsername(ctx context.Context, username string) (userMetadata model.UserInfo, err error) {
	var lastLoggedIn, lastPasswordChange, userCreatedAt sql.NullTime

	err = p.db.QueryRowContext(ctx, p.sqlSelectUserMetadataByUsername, username).Scan(
		&userMetadata.Username,
		&lastLoggedIn,
		&lastPasswordChange,
		&userCreatedAt,
	)

	switch {
	case err == nil:
		if lastLoggedIn.Valid {
			userMetadata.LastLoggedIn = &lastLoggedIn.Time
		}

		if lastPasswordChange.Valid {
			userMetadata.LastPasswordChange = &lastPasswordChange.Time
		}

		if userCreatedAt.Valid {
			userMetadata.UserCreatedAt = &userCreatedAt.Time
		}

		return userMetadata, nil
	case errors.Is(err, sql.ErrNoRows):
		return model.UserInfo{}, sql.ErrNoRows
	default:
		return model.UserInfo{}, fmt.Errorf("error retrieving user Metadata for user '%s': %w", username, err)
	}
}

// UpdateUserSignInDateByUsername save the current time as the last time a user logged in successfully.
func (p *SQLProvider) UpdateUserSignInDateByUsername(ctx context.Context, username string) (err error) {
	result, err := p.db.ExecContext(ctx, p.sqlUpdateUserRecordSignInByUsername, time.Now(), username)
	if err != nil {
		return fmt.Errorf("error updating latest sign in time for user '%s: %w", username, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected for user '%s': %w", username, err)
	}

	if rowsAffected == 0 {
		if err := p.CreateExistingUserMetadataAtLogin(ctx, username); err != nil {
			return err
		}
	}

	return nil
}

// DeleteUserByUsername deletes a specific user from the storage provider.
func (p *SQLProvider) DeleteUserByUsername(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteUserMetadataByUsername, username); err != nil {
		return fmt.Errorf("error deleting user Metadata for user '%s: %w", username, err)
	}

	return nil
}

// UpdatePasswordChangedDateByUsername save the current time as the last time a user changed their password.
func (p *SQLProvider) UpdatePasswordChangedDateByUsername(ctx context.Context, username string) (err error) {
	result, err := p.db.ExecContext(ctx, p.sqlUpdateUserRecordPasswordChangedAtByUsername, time.Now(), username)
	if err != nil {
		return fmt.Errorf("error updating 'latest password change date' for user '%s: %w", username, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("error getting rows affected for user '%s': %w", username, err)
	}

	if rowsAffected == 0 {
		p.log.Infof(("CreateExistingUserMetadataAtPasswordChange"))

		if err := p.CreateExistingUserMetadataAtPasswordChange(ctx, username); err != nil {
			return err
		}
	}

	return nil
}

// SavePreferred2FAMethod save the preferred method for 2FA for a username to the storage provider.
func (p *SQLProvider) SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpsertPreferred2FAMethod, username, method); err != nil {
		return fmt.Errorf("error inserting preferred two factor method for user '%s': %w", username, err)
	}

	return nil
}

// LoadPreferred2FAMethod load the preferred method for 2FA for a username from the storage provider.
func (p *SQLProvider) LoadPreferred2FAMethod(ctx context.Context, username string) (method string, err error) {
	err = p.db.GetContext(ctx, &method, p.sqlSelectPreferred2FAMethod, username)

	switch {
	case err == nil:
		return method, nil
	case errors.Is(err, sql.ErrNoRows):
		return "", sql.ErrNoRows
	default:
		return "", fmt.Errorf("error selecting preferred two factor method for user '%s': %w", username, err)
	}
}

// LoadUserInfo loads the model.UserInfo from the storage provider.
func (p *SQLProvider) LoadUserInfo(ctx context.Context, username string) (info model.UserInfo, err error) {
	err = p.db.GetContext(ctx, &info, p.sqlSelectUserInfo, username, username, username, username)

	switch {
	case err == nil, errors.Is(err, sql.ErrNoRows):
		return info, nil
	default:
		return model.UserInfo{}, fmt.Errorf("error selecting user info for user '%s': %w", username, err)
	}
}

// LoadAllUserInfo loads the model.UserInfo from the storage provider for all users.
func (p *SQLProvider) LoadAllUserInfo(ctx context.Context) (info []model.UserInfo, err error) {
	rows, err := p.db.QueryContext(ctx, p.sqlSelectAllUserInfo)
	if err != nil {
		return nil, fmt.Errorf("error selecting user info for all users: %w", err)
	}

	defer rows.Close()

	var users []model.UserInfo

	for rows.Next() {
		var user model.UserInfo

		err := rows.Scan(
			&user.Username,
			&user.Method,
			&user.HasTOTP,
			&user.HasWebAuthn,
			&user.HasDuo,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning all user info: %w", err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user info: %w", err)
	}

	return users, nil
}

// LoadAllUserInfoAndMetadata loads the model.UserInfo from the storage provider for all users.
func (p *SQLProvider) LoadAllUserInfoAndMetadata(ctx context.Context) (info []model.UserInfo, err error) {
	rows, err := p.db.QueryContext(ctx, p.sqlSelectAllUserInfoAndMetadata)
	if err != nil {
		return nil, fmt.Errorf("error selecting user info and Metadata for all users: %w", err)
	}

	defer rows.Close()

	var users []model.UserInfo

	for rows.Next() {
		var user model.UserInfo

		err := rows.Scan(
			&user.Username,
			&user.Method,
			&user.LastLoggedIn,
			&user.LastPasswordChange,
			&user.UserCreatedAt,
			&user.HasTOTP,
			&user.HasWebAuthn,
			&user.HasDuo,
		)
		if err != nil {
			return nil, fmt.Errorf("error scanning all user info and metadata : %w", err)
		}

		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user info and metadata: %w", err)
	}

	return users, nil
}

// SaveUserOpaqueIdentifier saves a new opaque user identifier to the storage provider.
func (p *SQLProvider) SaveUserOpaqueIdentifier(ctx context.Context, subject model.UserOpaqueIdentifier) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertUserOpaqueIdentifier, subject.Service, subject.SectorID, subject.Username, subject.Identifier); err != nil {
		return fmt.Errorf("error inserting user opaque id for user '%s' with opaque id '%s': %w", subject.Username, subject.Identifier, err)
	}

	return nil
}

// LoadUserOpaqueIdentifier selects an opaque user identifier from the storage provider.
func (p *SQLProvider) LoadUserOpaqueIdentifier(ctx context.Context, identifier uuid.UUID) (subject *model.UserOpaqueIdentifier, err error) {
	subject = &model.UserOpaqueIdentifier{}

	if err = p.db.GetContext(ctx, subject, p.sqlSelectUserOpaqueIdentifier, identifier); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, nil
		default:
			return nil, fmt.Errorf("error selecting user opaque id with value '%s': %w", identifier, err)
		}
	}

	return subject, nil
}

// LoadUserOpaqueIdentifiers selects an opaque user identifiers from the storage provider.
func (p *SQLProvider) LoadUserOpaqueIdentifiers(ctx context.Context) (identifiers []model.UserOpaqueIdentifier, err error) {
	var rows *sqlx.Rows

	if rows, err = p.db.QueryxContext(ctx, p.sqlSelectUserOpaqueIdentifiers); err != nil {
		return nil, fmt.Errorf("error selecting user opaque identifiers: %w", err)
	}

	var opaqueID *model.UserOpaqueIdentifier

	for rows.Next() {
		opaqueID = &model.UserOpaqueIdentifier{}

		if err = rows.StructScan(opaqueID); err != nil {
			return nil, fmt.Errorf("error selecting user opaque identifiers: error scanning row: %w", err)
		}

		identifiers = append(identifiers, *opaqueID)
	}

	return identifiers, nil
}

// LoadUserOpaqueIdentifierBySignature selects an opaque user identifier from the storage provider given a service name, sector id, and username.
func (p *SQLProvider) LoadUserOpaqueIdentifierBySignature(ctx context.Context, service, sectorID, username string) (subject *model.UserOpaqueIdentifier, err error) {
	subject = &model.UserOpaqueIdentifier{}

	if err = p.db.GetContext(ctx, subject, p.sqlSelectUserOpaqueIdentifierBySignature, service, sectorID, username); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, nil
		default:
			return nil, fmt.Errorf("error selecting user opaque with service '%s' and sector '%s' for username '%s': %w", service, sectorID, username, err)
		}
	}

	return subject, nil
}

// SaveTOTPConfiguration save a TOTP configuration of a given user in the storage provider.
func (p *SQLProvider) SaveTOTPConfiguration(ctx context.Context, config model.TOTPConfiguration) (err error) {
	if config.Secret, err = p.encrypt(config.Secret); err != nil {
		return fmt.Errorf("error encrypting TOTP configuration secret for user '%s': %w", config.Username, err)
	}

	if _, err = p.db.ExecContext(ctx, p.sqlUpsertTOTPConfig,
		config.CreatedAt, config.LastUsedAt,
		config.Username, config.Issuer,
		config.Algorithm, config.Digits, config.Period, config.Secret); err != nil {
		return fmt.Errorf("error upserting TOTP configuration for user '%s': %w", config.Username, err)
	}

	return nil
}

// UpdateTOTPConfigurationSignIn updates a registered TOTP configuration in the storage provider with the relevant sign in information.
func (p *SQLProvider) UpdateTOTPConfigurationSignIn(ctx context.Context, id int, lastUsedAt sql.NullTime) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateTOTPConfigRecordSignIn, lastUsedAt, id); err != nil {
		return fmt.Errorf("error updating TOTP configuration id %d: %w", id, err)
	}

	return nil
}

// DeleteTOTPConfiguration delete a TOTP configuration from the storage provider given a username.
func (p *SQLProvider) DeleteTOTPConfiguration(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteTOTPConfig, username); err != nil {
		return fmt.Errorf("error deleting TOTP configuration for user '%s': %w", username, err)
	}

	return nil
}

// LoadTOTPConfiguration load a TOTP configuration given a username from the storage provider.
func (p *SQLProvider) LoadTOTPConfiguration(ctx context.Context, username string) (config *model.TOTPConfiguration, err error) {
	config = &model.TOTPConfiguration{}

	if err = p.db.GetContext(ctx, config, p.sqlSelectTOTPConfig, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoTOTPConfiguration
		}

		return nil, fmt.Errorf("error selecting TOTP configuration for user '%s': %w", username, err)
	}

	if config.Secret, err = p.decrypt(config.Secret); err != nil {
		return nil, fmt.Errorf("error decrypting TOTP secret for user '%s': %w", username, err)
	}

	return config, nil
}

// SaveTOTPHistory saves a TOTP history item in the storage provider.
func (p *SQLProvider) SaveTOTPHistory(ctx context.Context, username string, step uint64) (err error) {
	signature := p.otpHMACSignature([]byte(strconv.FormatUint(step, 10)), []byte(username))

	if _, err = p.db.ExecContext(ctx, p.sqlInsertTOTPHistory, username, signature); err != nil {
		return fmt.Errorf("error inserting TOTP history for user '%s': %w", username, err)
	}

	return nil
}

// ExistsTOTPHistory checks if a TOTP history item exists in the storage provider.
func (p *SQLProvider) ExistsTOTPHistory(ctx context.Context, username string, step uint64) (exists bool, err error) {
	var count int

	signature := p.otpHMACSignature([]byte(strconv.FormatUint(step, 10)), []byte(username))

	if err = p.db.GetContext(ctx, &count, p.sqlSelectTOTPHistory, username, signature); err != nil {
		return false, fmt.Errorf("error checking if TOTP history exists: %w", err)
	}

	return count != 0, nil
}

// LoadTOTPConfigurations load a set of TOTP configurations from the storage provider.
func (p *SQLProvider) LoadTOTPConfigurations(ctx context.Context, limit, page int) (configs []model.TOTPConfiguration, err error) {
	configs = make([]model.TOTPConfiguration, 0, limit)

	if err = p.db.SelectContext(ctx, &configs, p.sqlSelectTOTPConfigs, limit, limit*page); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting TOTP configurations: %w", err)
	}

	for i, c := range configs {
		if configs[i].Secret, err = p.decrypt(c.Secret); err != nil {
			return nil, fmt.Errorf("error decrypting TOTP configuration for user '%s': %w", c.Username, err)
		}
	}

	return configs, nil
}

// SaveWebAuthnUser saves a registered WebAuthn user to the storage provider.
func (p *SQLProvider) SaveWebAuthnUser(ctx context.Context, user model.WebAuthnUser) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertWebAuthnUser, user.RPID, user.Username, user.UserID); err != nil {
		return fmt.Errorf("error inserting WebAuthn user '%s' with relying party id '%s': %w", user.Username, user.RPID, err)
	}

	return nil
}

// LoadWebAuthnUser loads a registered WebAuthn user from the storage provider.
func (p *SQLProvider) LoadWebAuthnUser(ctx context.Context, rpid, username string) (user *model.WebAuthnUser, err error) {
	user = &model.WebAuthnUser{}

	if err = p.db.GetContext(ctx, user, p.sqlSelectWebAuthnUser, rpid, username); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, nil
		default:
			return nil, fmt.Errorf("error selecting WebAuthn user '%s' with relying party id '%s': %w", username, rpid, err)
		}
	}

	return user, nil
}

// LoadWebAuthnUserByUserID loads a registered WebAuthn user from the storage provider.
func (p *SQLProvider) LoadWebAuthnUserByUserID(ctx context.Context, rpid, userID string) (user *model.WebAuthnUser, err error) {
	user = &model.WebAuthnUser{}

	if err = p.db.GetContext(ctx, user, p.sqlSelectWebAuthnUserByUserID, rpid, userID); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, nil
		default:
			return nil, fmt.Errorf("error selecting WebAuthn user with user id '%s' and relying party id '%s': %w", userID, rpid, err)
		}
	}

	return user, nil
}

// SaveWebAuthnCredential saves a registered WebAuthn credential to the storage provider.
func (p *SQLProvider) SaveWebAuthnCredential(ctx context.Context, credential model.WebAuthnCredential) (err error) {
	if credential.PublicKey, err = p.encrypt(credential.PublicKey); err != nil {
		return fmt.Errorf("error encrypting WebAuthn credential public key for user '%s' kid '%x': %w", credential.Username, credential.KID, err)
	}

	if len(credential.Attestation) != 0 {
		if credential.Attestation, err = p.encrypt(credential.Attestation); err != nil {
			return fmt.Errorf("error encrypting WebAuthn credential attestation for user '%s' kid '%x': %w", credential.Username, credential.KID, err)
		}
	}

	if _, err = p.db.ExecContext(ctx, p.sqlInsertWebAuthnCredential,
		credential.CreatedAt, credential.LastUsedAt, credential.RPID, credential.Username, credential.Description,
		credential.KID, credential.AAGUID, credential.AttestationType, credential.Attachment, credential.Transport,
		credential.SignCount, credential.CloneWarning, credential.Discoverable, credential.Present, credential.Verified,
		credential.BackupEligible, credential.BackupState, credential.PublicKey, credential.Attestation,
	); err != nil {
		return fmt.Errorf("error inserting WebAuthn credential for user '%s' kid '%x': %w", credential.Username, credential.KID, err)
	}

	return nil
}

// UpdateWebAuthnCredentialDescription updates a registered WebAuthn credential in the storage provider changing the
// description.
func (p *SQLProvider) UpdateWebAuthnCredentialDescription(ctx context.Context, username string, credentialID int, description string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateWebAuthnCredentialDescriptionByUsernameAndID, description, username, credentialID); err != nil {
		return fmt.Errorf("error updating WebAuthn credential description to '%s' for credential id '%d': %w", description, credentialID, err)
	}

	return nil
}

// UpdateWebAuthnCredentialSignIn updates a registered WebAuthn credential in the storage provider changing the
// information that should be changed in the event of a successful sign in.
func (p *SQLProvider) UpdateWebAuthnCredentialSignIn(ctx context.Context, credential model.WebAuthnCredential) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateWebAuthnCredentialRecordSignIn,
		credential.RPID, credential.LastUsedAt, credential.SignCount, credential.Discoverable, credential.Present, credential.Verified,
		credential.BackupEligible, credential.BackupState, credential.CloneWarning, credential.ID,
	); err != nil {
		return fmt.Errorf("error updating WebAuthn credentials authentication metadata for id '%x': %w", credential.ID, err)
	}

	return nil
}

// DeleteWebAuthnCredential deletes a registered WebAuthn credential from the storage provider.
func (p *SQLProvider) DeleteWebAuthnCredential(ctx context.Context, kid string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteWebAuthnCredential, kid); err != nil {
		return fmt.Errorf("error deleting WebAuthn credential with kid '%s': %w", kid, err)
	}

	return nil
}

// DeleteWebAuthnCredentialByUsername deletes registered WebAuthn credential from the storage provider by username or
// username and description.
func (p *SQLProvider) DeleteWebAuthnCredentialByUsername(ctx context.Context, username, displayname string) (err error) {
	if len(username) == 0 {
		return fmt.Errorf("error deleting WebAuthn credential with username '%s' and displayname '%s': username must not be empty", username, displayname)
	}

	if len(displayname) == 0 {
		if _, err = p.db.ExecContext(ctx, p.sqlDeleteWebAuthnCredentialByUsername, username); err != nil {
			return fmt.Errorf("error deleting WebAuthn credential for username '%s': %w", username, err)
		}
	} else {
		if _, err = p.db.ExecContext(ctx, p.sqlDeleteWebAuthnCredentialByUsernameAndDisplayName, username, displayname); err != nil {
			return fmt.Errorf("error deleting WebAuthn credential with username '%s' and displayname '%s': %w", username, displayname, err)
		}
	}

	return nil
}

// LoadWebAuthnCredentials loads WebAuthn credential registrations from the storage provider.
func (p *SQLProvider) LoadWebAuthnCredentials(ctx context.Context, limit, page int) (credentials []model.WebAuthnCredential, err error) {
	credentials = make([]model.WebAuthnCredential, 0, limit)

	if err = p.db.SelectContext(ctx, &credentials, p.sqlSelectWebAuthnCredentials, limit, limit*page); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting WebAuthn credentials: %w", err)
	}

	for i, credential := range credentials {
		if credentials[i].PublicKey, err = p.decrypt(credential.PublicKey); err != nil {
			return nil, fmt.Errorf("error decrypting WebAuthn credential public key of credential with id '%d' for user '%s': %w", credential.ID, credential.Username, err)
		}

		if len(credential.Attestation) != 0 {
			if credentials[i].Attestation, err = p.decrypt(credential.Attestation); err != nil {
				return nil, fmt.Errorf("error decrypting WebAuthn credential attestation of credential with id '%d' for user '%s': %w", credential.ID, credential.Username, err)
			}
		}
	}

	return credentials, nil
}

// LoadWebAuthnCredentialByID loads a WebAuthn credential registration from the storage provider for a given id.
func (p *SQLProvider) LoadWebAuthnCredentialByID(ctx context.Context, id int) (credential *model.WebAuthnCredential, err error) {
	credential = &model.WebAuthnCredential{}

	if err = p.db.GetContext(ctx, credential, p.sqlSelectWebAuthnCredentialByID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sql.ErrNoRows
		}

		return nil, fmt.Errorf("error selecting WebAuthn credential with id '%d': %w", id, err)
	}

	if credential.PublicKey, err = p.decrypt(credential.PublicKey); err != nil {
		return nil, fmt.Errorf("error decrypting WebAuthn credential public key of credential with id '%d' for user '%s': %w", credential.ID, credential.Username, err)
	}

	if len(credential.Attestation) != 0 {
		if credential.Attestation, err = p.decrypt(credential.Attestation); err != nil {
			return nil, fmt.Errorf("error decrypting WebAuthn credential attestation of credential with id '%d' for user '%s': %w", credential.ID, credential.Username, err)
		}
	}

	return credential, nil
}

// LoadWebAuthnCredentialsByUsername loads all WebAuthn credential registrations from the storage provider for a
// given username.
func (p *SQLProvider) LoadWebAuthnCredentialsByUsername(ctx context.Context, rpid, username string) (credentials []model.WebAuthnCredential, err error) {
	switch len(rpid) {
	case 0:
		err = p.db.SelectContext(ctx, &credentials, p.sqlSelectWebAuthnCredentialsByUsername, username, false)
	default:
		err = p.db.SelectContext(ctx, &credentials, p.sqlSelectWebAuthnCredentialsByRPIDByUsername, rpid, username, false)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return credentials, ErrNoWebAuthnCredential
		}

		return nil, fmt.Errorf("error selecting WebAuthn credentials for user '%s': %w", username, err)
	}

	for i, credential := range credentials {
		if credentials[i].PublicKey, err = p.decrypt(credential.PublicKey); err != nil {
			return nil, fmt.Errorf("error decrypting WebAuthn credential public key of credential with id '%d' for user '%s': %w", credential.ID, credential.Username, err)
		}

		if len(credential.Attestation) != 0 {
			if credentials[i].Attestation, err = p.decrypt(credential.Attestation); err != nil {
				return nil, fmt.Errorf("error decrypting WebAuthn credential attestation of credential with id '%d' for user '%s': %w", credential.ID, credential.Username, err)
			}
		}
	}

	return credentials, nil
}

func (p *SQLProvider) LoadWebAuthnPasskeyCredentialsByUsername(ctx context.Context, rpid, username string) (credentials []model.WebAuthnCredential, err error) {
	switch len(rpid) {
	case 0:
		err = p.db.SelectContext(ctx, &credentials, p.sqlSelectWebAuthnCredentialsByUsername, username, true)
	default:
		err = p.db.SelectContext(ctx, &credentials, p.sqlSelectWebAuthnCredentialsByRPIDByUsername, rpid, username, true)
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return credentials, ErrNoWebAuthnCredential
		}

		return nil, fmt.Errorf("error selecting passkey WebAuthn credentials for user '%s': %w", username, err)
	}

	for i, credential := range credentials {
		if credentials[i].PublicKey, err = p.decrypt(credential.PublicKey); err != nil {
			return nil, fmt.Errorf("error decrypting passkey WebAuthn credential public key of credential with id '%d' for user '%s': %w", credential.ID, credential.Username, err)
		}

		if len(credential.Attestation) != 0 {
			if credentials[i].Attestation, err = p.decrypt(credential.Attestation); err != nil {
				return nil, fmt.Errorf("error decrypting passkey WebAuthn credential attestation of credential with id '%d' for user '%s': %w", credential.ID, credential.Username, err)
			}
		}
	}

	return credentials, nil
}

// SavePreferredDuoDevice saves a Duo device to the storage provider.
func (p *SQLProvider) SavePreferredDuoDevice(ctx context.Context, device model.DuoDevice) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpsertDuoDevice, device.Username, device.Device, device.Method); err != nil {
		return fmt.Errorf("error upserting preferred duo device for user '%s': %w", device.Username, err)
	}

	return nil
}

// DeletePreferredDuoDevice deletes a Duo device from the storage provider for a given username.
func (p *SQLProvider) DeletePreferredDuoDevice(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteDuoDevice, username); err != nil {
		return fmt.Errorf("error deleting preferred duo device for user '%s': %w", username, err)
	}

	return nil
}

// LoadPreferredDuoDevice loads a Duo device from the storage provider for a given username.
func (p *SQLProvider) LoadPreferredDuoDevice(ctx context.Context, username string) (device *model.DuoDevice, err error) {
	device = &model.DuoDevice{}

	if err = p.db.QueryRowxContext(ctx, p.sqlSelectDuoDevice, username).StructScan(device); err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrNoDuoDevice
		}

		return nil, fmt.Errorf("error selecting preferred duo device for user '%s': %w", username, err)
	}

	return device, nil
}

// SaveIdentityVerification save an identity verification record to the storage provider.
func (p *SQLProvider) SaveIdentityVerification(ctx context.Context, verification model.IdentityVerification) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertIdentityVerification,
		verification.JTI, verification.IssuedAt, verification.IssuedIP, verification.ExpiresAt,
		verification.Username, verification.Action); err != nil {
		return fmt.Errorf("error inserting identity verification for user '%s' with uuid '%s': %w", verification.Username, verification.JTI, err)
	}

	return nil
}

// ConsumeIdentityVerification marks an identity verification record in the storage provider as consumed.
func (p *SQLProvider) ConsumeIdentityVerification(ctx context.Context, jti string, ip model.NullIP) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlConsumeIdentityVerification, time.Now(), ip, jti); err != nil {
		return fmt.Errorf("error updating identity verification: %w", err)
	}

	return nil
}

// RevokeIdentityVerification marks an identity verification record in the storage provider as revoked.
func (p *SQLProvider) RevokeIdentityVerification(ctx context.Context, jti string, ip model.NullIP) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlRevokeIdentityVerification, time.Now(), ip, jti); err != nil {
		return fmt.Errorf("error updating identity verification: %w", err)
	}

	return nil
}

// FindIdentityVerification checks if an identity verification record is in the storage provider and active.
func (p *SQLProvider) FindIdentityVerification(ctx context.Context, jti string) (found bool, err error) {
	verification := model.IdentityVerification{}
	if err = p.db.GetContext(ctx, &verification, p.sqlSelectIdentityVerification, jti); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("error selecting identity verification exists: %w", err)
	}

	switch {
	case verification.RevokedAt.Valid:
		return false, fmt.Errorf("the token has been revoked")
	case verification.ConsumedAt.Valid:
		return false, fmt.Errorf("the token has already been consumed")
	case verification.ExpiresAt.Before(time.Now()):
		return false, fmt.Errorf("the token expired %s ago", time.Since(verification.ExpiresAt))
	default:
		return true, nil
	}
}

// LoadIdentityVerification loads an Identity Verification but does not do any validation.
// For easy validation you should use FindIdentityVerification which ensures the JWT is still valid.
func (p *SQLProvider) LoadIdentityVerification(ctx context.Context, jti string) (verification *model.IdentityVerification, err error) {
	verification = &model.IdentityVerification{}

	if err = p.db.GetContext(ctx, verification, p.sqlSelectIdentityVerification, jti); err != nil {
		return nil, fmt.Errorf("error selecting identity verification: %w", err)
	}

	return verification, nil
}

// SaveOneTimeCode saves a One-Time Code to the storage provider after generating the signature which is returned
// along with any error.
func (p *SQLProvider) SaveOneTimeCode(ctx context.Context, code model.OneTimeCode) (signature string, err error) {
	code.Signature = p.otcHMACSignature([]byte(code.Username), []byte(code.Intent), code.Code)

	if code.Code, err = p.encrypt(code.Code); err != nil {
		return "", fmt.Errorf("error encrypting the one-time code value for user '%s' with signature '%s': %w", code.Username, code.Signature, err)
	}

	if _, err = p.db.ExecContext(ctx, p.sqlInsertOneTimeCode,
		code.PublicID, code.Signature, code.IssuedAt, code.IssuedIP, code.ExpiresAt,
		code.Username, code.Intent, code.Code); err != nil {
		return "", fmt.Errorf("error inserting one-time code for user '%s' with signature '%s': %w", code.Username, code.Signature, err)
	}

	return code.Signature, nil
}

// ConsumeOneTimeCode consumes a one-time code using the signature.
func (p *SQLProvider) ConsumeOneTimeCode(ctx context.Context, code *model.OneTimeCode) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlConsumeOneTimeCode, code.ConsumedAt, code.ConsumedIP, code.Signature); err != nil {
		return fmt.Errorf("error updating one-time code (consume): %w", err)
	}

	return nil
}

// RevokeOneTimeCode revokes a one-time code in the storage provider using the public ID.
func (p *SQLProvider) RevokeOneTimeCode(ctx context.Context, publicID uuid.UUID, ip model.IP) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlRevokeOneTimeCode, time.Now(), ip, publicID); err != nil {
		return fmt.Errorf("error updating one-time code (revoke): %w", err)
	}

	return nil
}

// LoadOneTimeCode loads a one-time code from the storage provider given a username, intent, and code.
func (p *SQLProvider) LoadOneTimeCode(ctx context.Context, username, intent, raw string) (code *model.OneTimeCode, err error) {
	code = &model.OneTimeCode{}

	signature := p.otcHMACSignature([]byte(username), []byte(intent), []byte(raw))

	if err = p.db.GetContext(ctx, code, p.sqlSelectOneTimeCode, signature, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting one-time code: %w", err)
	}

	if code.Code, err = p.decrypt(code.Code); err != nil {
		return nil, fmt.Errorf("error decrypting the one-time code value for user '%s' with signature '%s': %w", code.Username, code.Signature, err)
	}

	return code, nil
}

// LoadOneTimeCodeBySignature loads a one-time code from the storage provider given the signature.
// This method should NOT be used to validate a One-Time Code, LoadOneTimeCode should be used instead.
func (p *SQLProvider) LoadOneTimeCodeBySignature(ctx context.Context, signature string) (code *model.OneTimeCode, err error) {
	code = &model.OneTimeCode{}

	if err = p.db.GetContext(ctx, code, p.sqlSelectOneTimeCodeBySignature, signature); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting one-time code: %w", err)
	}

	if code.Code, err = p.decrypt(code.Code); err != nil {
		return nil, fmt.Errorf("error decrypting the one-time code value for user '%s' with signature '%s': %w", code.Username, code.Signature, err)
	}

	return code, nil
}

// LoadOneTimeCodeByID loads a one-time code from the storage provider given the id.
// This does not decrypt the code. This method should NOT be used to validate a One-Time Code,
// LoadOneTimeCode should be used instead.
func (p *SQLProvider) LoadOneTimeCodeByID(ctx context.Context, id int) (code *model.OneTimeCode, err error) {
	code = &model.OneTimeCode{}

	if err = p.db.GetContext(ctx, code, p.sqlSelectOneTimeCodeByID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting one-time code: %w", err)
	}

	return code, nil
}

// LoadOneTimeCodeByPublicID loads a one-time code from the storage provider given the public identifier.
// This does not decrypt the code. This method SHOULD ONLY be used to find the One-Time Code for the
// purpose of deletion.
func (p *SQLProvider) LoadOneTimeCodeByPublicID(ctx context.Context, id uuid.UUID) (code *model.OneTimeCode, err error) {
	code = &model.OneTimeCode{}

	if err = p.db.GetContext(ctx, code, p.sqlSelectOneTimeCodeByPublicID, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting one-time code: %w", err)
	}

	return code, nil
}

// SaveOAuth2ConsentPreConfiguration inserts an OAuth2.0 consent pre-configuration in the storage provider.
func (p *SQLProvider) SaveOAuth2ConsentPreConfiguration(ctx context.Context, config model.OAuth2ConsentPreConfig) (insertedID int64, err error) {
	switch p.name {
	case providerPostgres:
		err = p.db.GetContext(ctx, &insertedID, p.sqlInsertOAuth2ConsentPreConfiguration,
			config.ClientID, config.Subject, config.CreatedAt, config.ExpiresAt,
			config.Revoked, config.Scopes, config.Audience,
			config.RequestedClaims, config.SignatureClaims, config.GrantedClaims)
	default:
		var result sql.Result

		if result, err = p.db.ExecContext(ctx, p.sqlInsertOAuth2ConsentPreConfiguration,
			config.ClientID, config.Subject, config.CreatedAt, config.ExpiresAt,
			config.Revoked, config.Scopes, config.Audience,
			config.RequestedClaims, config.SignatureClaims, config.GrantedClaims); err == nil {
			insertedID, err = result.LastInsertId()
		}
	}

	if err != nil {
		return -1, fmt.Errorf("error inserting oauth2 consent pre-configuration for subject '%s' with client id '%s' and scopes '%s': %w", config.Subject.String(), config.ClientID, strings.Join(config.Scopes, " "), err)
	}

	return insertedID, nil
}

// LoadOAuth2ConsentPreConfigurations returns an OAuth2.0 consents pre-configurations from the storage provider given the consent signature.
func (p *SQLProvider) LoadOAuth2ConsentPreConfigurations(ctx context.Context, clientID string, subject uuid.UUID, now time.Time) (rows *ConsentPreConfigRows, err error) {
	var r *sqlx.Rows

	if r, err = p.db.QueryxContext(ctx, p.sqlSelectOAuth2ConsentPreConfigurations, clientID, subject, now); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &ConsentPreConfigRows{}, nil
		}

		return &ConsentPreConfigRows{}, fmt.Errorf("error selecting oauth2 consent pre-configurations by signature with client id '%s' and subject '%s': %w", clientID, subject, err)
	}

	return &ConsentPreConfigRows{rows: r}, nil
}

// SaveOAuth2ConsentSession inserts an OAuth2.0 consent session to the storage provider.
func (p *SQLProvider) SaveOAuth2ConsentSession(ctx context.Context, consent *model.OAuth2ConsentSession) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertOAuth2ConsentSession,
		consent.ChallengeID, consent.ClientID, consent.Subject, consent.Authorized, consent.Granted,
		consent.RequestedAt, consent.ExpiresAt, consent.RespondedAt, consent.Form,
		consent.RequestedScopes, consent.GrantedScopes, consent.RequestedAudience, consent.GrantedAudience, consent.GrantedClaims, consent.PreConfiguration); err != nil {
		return fmt.Errorf("error inserting oauth2 consent session with challenge id '%s' for subject '%s': %w", consent.ChallengeID.String(), consent.Subject.UUID.String(), err)
	}

	return nil
}

// SaveOAuth2ConsentSessionResponse updates an OAuth2.0 consent session in the storage provider with the response.
func (p *SQLProvider) SaveOAuth2ConsentSessionResponse(ctx context.Context, consent *model.OAuth2ConsentSession, authorized bool) (err error) {
	if !consent.RespondedAt.Valid {
		consent.SetRespondedAt(time.Now(), 0)
	}

	if _, err = p.db.ExecContext(ctx, p.sqlUpdateOAuth2ConsentSessionResponse, consent.Subject, consent.RespondedAt, authorized, consent.GrantedScopes, consent.GrantedAudience, consent.GrantedClaims, consent.PreConfiguration, consent.ID); err != nil {
		return fmt.Errorf("error updating oauth2 consent session (authorized  '%t') with id '%d' and challenge id '%s' for subject '%s': %w", authorized, consent.ID, consent.ChallengeID, consent.Subject.UUID, err)
	}

	return nil
}

// SaveOAuth2ConsentSessionGranted updates an OAuth2.0 consent session in the storage provider recording that it
// has been granted by the authorization endpoint.
func (p *SQLProvider) SaveOAuth2ConsentSessionGranted(ctx context.Context, id int) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateOAuth2ConsentSessionGranted, id); err != nil {
		return fmt.Errorf("error updating oauth2 consent session (granted) with id '%d': %w", id, err)
	}

	return nil
}

// LoadOAuth2ConsentSessionByChallengeID returns an OAuth2.0 consent session in the storage provider given the challenge ID.
func (p *SQLProvider) LoadOAuth2ConsentSessionByChallengeID(ctx context.Context, challengeID uuid.UUID) (consent *model.OAuth2ConsentSession, err error) {
	consent = &model.OAuth2ConsentSession{}

	if err = p.db.GetContext(ctx, consent, p.sqlSelectOAuth2ConsentSessionByChallengeID, challengeID); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 consent session with challenge id '%s': %w", challengeID, err)
	}

	return consent, nil
}

// SaveOAuth2Session saves an OAut2.0 session to the storage provider.
func (p *SQLProvider) SaveOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, session model.OAuth2Session) (err error) {
	var query string

	switch sessionType {
	case OAuth2SessionTypeAccessToken:
		query = p.sqlInsertOAuth2AccessTokenSession
	case OAuth2SessionTypeAuthorizeCode:
		query = p.sqlInsertOAuth2AuthorizeCodeSession
	case OAuth2SessionTypeOpenIDConnect:
		query = p.sqlInsertOAuth2OpenIDConnectSession
	case OAuth2SessionTypePKCEChallenge:
		query = p.sqlInsertOAuth2PKCERequestSession
	case OAuth2SessionTypeRefreshToken:
		query = p.sqlInsertOAuth2RefreshTokenSession
	default:
		return fmt.Errorf("error inserting oauth2 session for subject '%s' and request id '%s': unknown oauth2 session type '%s'", session.Subject.String, session.RequestID, sessionType)
	}

	if session.Session, err = p.encrypt(session.Session); err != nil {
		return fmt.Errorf("error encrypting oauth2 %s session data for subject '%s' and request id '%s' and challenge id '%s': %w", sessionType, session.Subject.String, session.RequestID, session.ChallengeID.UUID, err)
	}

	_, err = p.db.ExecContext(ctx, query,
		session.ChallengeID, session.RequestID, session.ClientID, session.Signature,
		session.Subject, session.RequestedAt, session.RequestedScopes, session.GrantedScopes,
		session.RequestedAudience, session.GrantedAudience,
		session.Active, session.Revoked, session.Form, session.Session)
	if err != nil {
		return fmt.Errorf("error inserting oauth2 %s session with signature '%s' for subject '%s' and request id '%s' and challenge id '%s': %w", sessionType, session.Signature, session.Subject.String, session.RequestID, session.ChallengeID.UUID, err)
	}

	return nil
}

// RevokeOAuth2Session marks an OAuth2.0 session as revoked in the storage provider.
func (p *SQLProvider) RevokeOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, signature string) (err error) {
	var query string

	switch sessionType {
	case OAuth2SessionTypeAccessToken:
		query = p.sqlRevokeOAuth2AccessTokenSession
	case OAuth2SessionTypeAuthorizeCode:
		query = p.sqlRevokeOAuth2AuthorizeCodeSession
	case OAuth2SessionTypeOpenIDConnect:
		query = p.sqlRevokeOAuth2OpenIDConnectSession
	case OAuth2SessionTypePKCEChallenge:
		query = p.sqlRevokeOAuth2PKCERequestSession
	case OAuth2SessionTypeRefreshToken:
		query = p.sqlRevokeOAuth2RefreshTokenSession
	default:
		return fmt.Errorf("error revoking oauth2 session with signature '%s': unknown oauth2 session type '%s'", signature, sessionType)
	}

	if _, err = p.db.ExecContext(ctx, query, signature); err != nil {
		return fmt.Errorf("error revoking oauth2 %s session with signature '%s': %w", sessionType, signature, err)
	}

	return nil
}

// RevokeOAuth2SessionByRequestID marks an OAuth2.0 session as revoked in the storage provider.
func (p *SQLProvider) RevokeOAuth2SessionByRequestID(ctx context.Context, sessionType OAuth2SessionType, requestID string) (err error) {
	var query string

	switch sessionType {
	case OAuth2SessionTypeAccessToken:
		query = p.sqlRevokeOAuth2AccessTokenSessionByRequestID
	case OAuth2SessionTypeAuthorizeCode:
		query = p.sqlRevokeOAuth2AuthorizeCodeSessionByRequestID
	case OAuth2SessionTypeOpenIDConnect:
		query = p.sqlRevokeOAuth2OpenIDConnectSessionByRequestID
	case OAuth2SessionTypePKCEChallenge:
		query = p.sqlRevokeOAuth2PKCERequestSessionByRequestID
	case OAuth2SessionTypeRefreshToken:
		query = p.sqlRevokeOAuth2RefreshTokenSessionByRequestID
	default:
		return fmt.Errorf("error revoking oauth2 session with request id '%s': unknown oauth2 session type '%s'", requestID, sessionType)
	}

	if _, err = p.db.ExecContext(ctx, query, requestID); err != nil {
		return fmt.Errorf("error revoking oauth2 %s session with request id '%s': %w", sessionType, requestID, err)
	}

	return nil
}

// DeactivateOAuth2Session marks an OAuth2.0 session as inactive in the storage provider.
func (p *SQLProvider) DeactivateOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, signature string) (err error) {
	var query string

	switch sessionType {
	case OAuth2SessionTypeAccessToken:
		query = p.sqlDeactivateOAuth2AccessTokenSession
	case OAuth2SessionTypeAuthorizeCode:
		query = p.sqlDeactivateOAuth2AuthorizeCodeSession
	case OAuth2SessionTypeOpenIDConnect:
		query = p.sqlDeactivateOAuth2OpenIDConnectSession
	case OAuth2SessionTypePKCEChallenge:
		query = p.sqlDeactivateOAuth2PKCERequestSession
	case OAuth2SessionTypeRefreshToken:
		query = p.sqlDeactivateOAuth2RefreshTokenSession
	default:
		return fmt.Errorf("error deactivating oauth2 session with signature '%s': unknown oauth2 session type '%s'", signature, sessionType)
	}

	if _, err = p.db.ExecContext(ctx, query, signature); err != nil {
		return fmt.Errorf("error deactivating oauth2 %s session with signature '%s': %w", sessionType, signature, err)
	}

	return nil
}

// DeactivateOAuth2SessionByRequestID marks an OAuth2.0 session as inactive in the storage provider.
func (p *SQLProvider) DeactivateOAuth2SessionByRequestID(ctx context.Context, sessionType OAuth2SessionType, requestID string) (err error) {
	var query string

	switch sessionType {
	case OAuth2SessionTypeAccessToken:
		query = p.sqlDeactivateOAuth2AccessTokenSessionByRequestID
	case OAuth2SessionTypeAuthorizeCode:
		query = p.sqlDeactivateOAuth2AuthorizeCodeSession
	case OAuth2SessionTypeOpenIDConnect:
		query = p.sqlDeactivateOAuth2OpenIDConnectSessionByRequestID
	case OAuth2SessionTypePKCEChallenge:
		query = p.sqlDeactivateOAuth2PKCERequestSessionByRequestID
	case OAuth2SessionTypeRefreshToken:
		query = p.sqlDeactivateOAuth2RefreshTokenSessionByRequestID
	default:
		return fmt.Errorf("error deactivating oauth2 session with request id '%s': unknown oauth2 session type '%s'", requestID, sessionType)
	}

	if _, err = p.db.ExecContext(ctx, query, requestID); err != nil {
		return fmt.Errorf("error deactivating oauth2 %s session with request id '%s': %w", sessionType, requestID, err)
	}

	return nil
}

// LoadOAuth2Session saves an OAuth2.0 session from the storage provider.
func (p *SQLProvider) LoadOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, signature string) (session *model.OAuth2Session, err error) {
	var query string

	switch sessionType {
	case OAuth2SessionTypeAccessToken:
		query = p.sqlSelectOAuth2AccessTokenSession
	case OAuth2SessionTypeAuthorizeCode:
		query = p.sqlSelectOAuth2AuthorizeCodeSession
	case OAuth2SessionTypeOpenIDConnect:
		query = p.sqlSelectOAuth2OpenIDConnectSession
	case OAuth2SessionTypePKCEChallenge:
		query = p.sqlSelectOAuth2PKCERequestSession
	case OAuth2SessionTypeRefreshToken:
		query = p.sqlSelectOAuth2RefreshTokenSession
	default:
		return nil, fmt.Errorf("error selecting oauth2 session: unknown oauth2 session type '%s'", sessionType)
	}

	session = &model.OAuth2Session{}

	if err = p.db.GetContext(ctx, session, query, signature); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 %s session with signature '%s': %w", sessionType, signature, err)
	}

	if session.Session, err = p.decrypt(session.Session); err != nil {
		return nil, fmt.Errorf("error decrypting the oauth2 %s session data with signature '%s' for subject '%s' and request id '%s': %w", sessionType, signature, session.Subject.String, session.RequestID, err)
	}

	return session, nil
}

func (p *SQLProvider) SaveOAuth2DeviceCodeSession(ctx context.Context, session *model.OAuth2DeviceCodeSession) (err error) {
	if session.Session, err = p.encrypt(session.Session); err != nil {
		return fmt.Errorf("error encrypting oauth2 device code session data for session with signature '%s' for subject '%s' and request id '%s': %w", session.Subject.String, session.Signature, session.RequestID, err)
	}

	_, err = p.db.ExecContext(ctx, p.sqlInsertOAuth2DeviceCodeSession,
		session.ChallengeID, session.RequestID, session.ClientID, session.Signature, session.UserCodeSignature,
		session.Status, session.Subject, session.RequestedAt, session.CheckedAt,
		session.RequestedScopes, session.GrantedScopes,
		session.RequestedAudience, session.GrantedAudience,
		session.Active, session.Revoked, session.Form, session.Session)
	if err != nil {
		return fmt.Errorf("error inserting oauth2 device code session with device code signature '%s' and user code signature '%s' for subject '%s' and request id '%s': %w", session.Signature, session.UserCodeSignature, session.Subject.String, session.RequestID, err)
	}

	return nil
}

func (p *SQLProvider) UpdateOAuth2DeviceCodeSession(ctx context.Context, session *model.OAuth2DeviceCodeSession) (err error) {
	if session.Session, err = p.encrypt(session.Session); err != nil {
		return fmt.Errorf("error encrypting oauth2 device code session data for session with signature '%s' for subject '%s' and request id '%s': %w", session.Subject.String, session.Signature, session.RequestID, err)
	}

	_, err = p.db.ExecContext(ctx, p.sqlUpdateOAuth2DeviceCodeSession,
		session.ChallengeID, session.RequestID, session.ClientID, session.Status, session.Subject, session.RequestedAt,
		session.CheckedAt, session.RequestedScopes, session.RequestedAudience, session.GrantedScopes, session.GrantedAudience,
		session.Active, session.Revoked, session.Form, session.Session, session.Signature)
	if err != nil {
		return fmt.Errorf("error updating oauth2 device code session with device code signature '%s': %w", session.Signature, err)
	}

	return nil
}

func (p *SQLProvider) UpdateOAuth2DeviceCodeSessionData(ctx context.Context, session *model.OAuth2DeviceCodeSession) (err error) {
	if session.Session, err = p.encrypt(session.Session); err != nil {
		return fmt.Errorf("error encrypting oauth2 device code session data for session with signature '%s' for subject '%s' and request id '%s': %w", session.Subject.String, session.Signature, session.RequestID, err)
	}

	_, err = p.db.ExecContext(ctx, p.sqlUpdateOAuth2DeviceCodeSessionData,
		session.ChallengeID, session.ClientID, session.Status, session.Subject,
		session.RequestedScopes, session.RequestedAudience, session.GrantedScopes, session.GrantedAudience,
		session.Form, session.Session, session.Signature)
	if err != nil {
		return fmt.Errorf("error updating oauth2 device code session data with device code signature '%s': %w", session.Signature, err)
	}

	return nil
}

func (p *SQLProvider) DeactivateOAuth2DeviceCodeSession(ctx context.Context, signature string) (err error) {
	_, err = p.db.ExecContext(ctx, p.sqlDeactivateOAuth2DeviceCodeSession, signature)
	if err != nil {
		return fmt.Errorf("error deactivating oauth2 device code session with device code signature '%s': %w", signature, err)
	}

	return nil
}

func (p *SQLProvider) LoadOAuth2DeviceCodeSession(ctx context.Context, signature string) (session *model.OAuth2DeviceCodeSession, err error) {
	session = &model.OAuth2DeviceCodeSession{}

	if err = p.db.GetContext(ctx, session, p.sqlSelectOAuth2DeviceCodeSession, signature); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 device code session with device code signature '%s': %w", signature, err)
	}

	if session.Session, err = p.decrypt(session.Session); err != nil {
		return nil, fmt.Errorf("error decrypting the oauth2 device code session data with device code signature '%s' for subject '%s' and request id '%s': %w", signature, session.Subject.String, session.RequestID, err)
	}

	return session, nil
}

func (p *SQLProvider) LoadOAuth2DeviceCodeSessionByUserCode(ctx context.Context, signature string) (session *model.OAuth2DeviceCodeSession, err error) {
	session = &model.OAuth2DeviceCodeSession{}

	if err = p.db.GetContext(ctx, session, p.sqlSelectOAuth2DeviceCodeSessionByUserCode, signature); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 device code session with user code signature '%s': %w", signature, err)
	}

	if session.Session, err = p.decrypt(session.Session); err != nil {
		return nil, fmt.Errorf("error decrypting the oauth2 device code session data with user code signature '%s' for subject '%s' and request id '%s': %w", signature, session.Subject.String, session.RequestID, err)
	}

	return session, nil
}

// SaveOAuth2PARContext save an OAuth2.0 PAR context to the storage provider.
func (p *SQLProvider) SaveOAuth2PARContext(ctx context.Context, par model.OAuth2PARContext) (err error) {
	if par.Session, err = p.encrypt(par.Session); err != nil {
		return fmt.Errorf("error encrypting oauth2 pushed authorization request context data for with signature '%s' and request id '%s': %w", par.Signature, par.RequestID, err)
	}

	if _, err = p.db.ExecContext(ctx, p.sqlInsertOAuth2PARContext,
		par.Signature, par.RequestID, par.ClientID, par.RequestedAt, par.Scopes, par.Audience, par.HandledResponseTypes,
		par.ResponseMode, par.DefaultResponseMode, par.Revoked, par.Form, par.Session); err != nil {
		return fmt.Errorf("error inserting oauth2 pushed authorization request context data for with signature '%s' and request id '%s': %w", par.Signature, par.RequestID, err)
	}

	return nil
}

// LoadOAuth2PARContext loads an OAuth2.0 PAR context from the storage provider.
func (p *SQLProvider) LoadOAuth2PARContext(ctx context.Context, signature string) (par *model.OAuth2PARContext, err error) {
	par = &model.OAuth2PARContext{}

	if err = p.db.GetContext(ctx, par, p.sqlSelectOAuth2PARContext, signature); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 pushed authorization request context with signature '%s': %w", signature, err)
	}

	if par.Session, err = p.decrypt(par.Session); err != nil {
		return nil, fmt.Errorf("error decrypting oauth2 oauth2 pushed authorization request context data with signature '%s' and request id '%s': %w", signature, par.RequestID, err)
	}

	return par, nil
}

// RevokeOAuth2PARContext marks an OAuth2.0 PAR context as revoked in the storage provider.
func (p *SQLProvider) RevokeOAuth2PARContext(ctx context.Context, signature string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlRevokeOAuth2PARContext, signature); err != nil {
		return fmt.Errorf("error revoking oauth2 pushed authorization request context with signature '%s': %w", signature, err)
	}

	return nil
}

// UpdateOAuth2PARContext updates an existing OAuth2.0 PAR context in the storage provider.
func (p *SQLProvider) UpdateOAuth2PARContext(ctx context.Context, par model.OAuth2PARContext) (err error) {
	if par.ID == 0 {
		return fmt.Errorf("error updating oauth2 pushed authorization request context data with signature '%s' and request id '%s': the id was a zero value", par.Signature, par.RequestID)
	}

	if par.Session, err = p.encrypt(par.Session); err != nil {
		return fmt.Errorf("error encrypting oauth2 pushed authorization request context data with id '%d' and signature '%s' and request id '%s': %w", par.ID, par.Signature, par.RequestID, err)
	}

	if _, err = p.db.ExecContext(ctx, p.sqlUpdateOAuth2PARContext,
		par.Signature, par.RequestID, par.ClientID, par.RequestedAt, par.Scopes, par.Audience, par.HandledResponseTypes,
		par.ResponseMode, par.DefaultResponseMode, par.Revoked, par.Form, par.Session, par.ID); err != nil {
		return fmt.Errorf("error updating oauth2 pushed authorization request context data with id '%d' and signature '%s' and request id '%s': %w", par.ID, par.Signature, par.RequestID, err)
	}

	return nil
}

// SaveOAuth2BlacklistedJTI saves an OAuth2.0 blacklisted JTI to the storage provider.
func (p *SQLProvider) SaveOAuth2BlacklistedJTI(ctx context.Context, blacklistedJTI model.OAuth2BlacklistedJTI) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpsertOAuth2BlacklistedJTI, blacklistedJTI.Signature, blacklistedJTI.ExpiresAt); err != nil {
		return fmt.Errorf("error inserting oauth2 blacklisted JTI with signature '%s': %w", blacklistedJTI.Signature, err)
	}

	return nil
}

// LoadOAuth2BlacklistedJTI loads an OAuth2.0 blacklisted JTI from the storage provider.
func (p *SQLProvider) LoadOAuth2BlacklistedJTI(ctx context.Context, signature string) (blacklistedJTI *model.OAuth2BlacklistedJTI, err error) {
	blacklistedJTI = &model.OAuth2BlacklistedJTI{}

	if err = p.db.GetContext(ctx, blacklistedJTI, p.sqlSelectOAuth2BlacklistedJTI, signature); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 blacklisted JTI with signature '%s': %w", signature, err)
	}

	return blacklistedJTI, nil
}

// AppendAuthenticationLog saves an authentication attempt to the storage provider.
func (p *SQLProvider) AppendAuthenticationLog(ctx context.Context, attempt model.AuthenticationAttempt) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertAuthenticationAttempt,
		attempt.Time, attempt.Successful, attempt.Banned, attempt.Username,
		attempt.Type, attempt.RemoteIP, attempt.RequestURI, attempt.RequestMethod); err != nil {
		return fmt.Errorf("error inserting authentication attempt for user '%s': %w", attempt.Username, err)
	}

	return nil
}

func (p *SQLProvider) LoadRegulationRecordsByUser(ctx context.Context, username string, since time.Time, limit int) (records []model.RegulationRecord, err error) {
	exp := banExpiresExpired{}

	if err = p.db.GetContext(ctx, &exp, p.sqlSelectBannedUserLastTime, username); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error selecting last banned user time for username '%s': %w", username, err)
		}
	}

	if expiration := exp.Expiration(); expiration.After(since) {
		since = expiration
	}

	records = make([]model.RegulationRecord, 0, limit)

	if err = p.db.SelectContext(ctx, &records, p.sqlSelectAuthenticationLogsRegulationRecordsByUsername, since, username, false, limit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting regulation records for username '%s': %w", username, err)
	}

	return records, nil
}

func (p *SQLProvider) SaveBannedUser(ctx context.Context, ban *model.BannedUser) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertBannedUser, ban.Expires, ban.Username, ban.Source, ban.Reason); err != nil {
		return fmt.Errorf("error inserting banned user with username '%s' and source '%s' and reason '%s': %w", ban.Username, ban.Source, ban.Reason.String, err)
	}

	return nil
}

func (p *SQLProvider) LoadBannedUser(ctx context.Context, username string) (bans []model.BannedUser, err error) {
	bans = []model.BannedUser{}

	if err = p.db.SelectContext(ctx, &bans, p.sqlSelectBannedUser, username, time.Now()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting banned user records for username '%s': %w", username, err)
	}

	return bans, nil
}

func (p *SQLProvider) LoadBannedUserByID(ctx context.Context, id int) (ban model.BannedUser, err error) {
	if err = p.db.GetContext(ctx, &ban, p.sqlSelectBannedUserByID, id); err != nil {
		return model.BannedUser{}, fmt.Errorf("error selecting banned user with id '%d': %w", id, err)
	}

	return ban, nil
}

func (p *SQLProvider) LoadBannedUsers(ctx context.Context, limit, page int) (bans []model.BannedUser, err error) {
	bans = []model.BannedUser{}

	if err = p.db.SelectContext(ctx, &bans, p.sqlSelectBannedUsers, false, time.Now(), limit, limit*page); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting banned user records: %w", err)
	}

	return bans, nil
}

func (p *SQLProvider) RevokeBannedUser(ctx context.Context, id int, expired time.Time) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlRevokeBannedUser, expired, id); err != nil {
		return fmt.Errorf("error revoking banned user with id '%d': %w", id, err)
	}

	return nil
}

func (p *SQLProvider) LoadRegulationRecordsByIP(ctx context.Context, ip model.IP, since time.Time, limit int) (records []model.RegulationRecord, err error) {
	exp := banExpiresExpired{}

	if err = p.db.GetContext(ctx, &exp, p.sqlSelectBannedIPLastTime, ip); err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("error selecting last banned time for ip '%s': %w", ip, err)
		}
	}

	if expiration := exp.Expiration(); expiration.After(since) {
		since = expiration
	}

	records = make([]model.RegulationRecord, 0, limit)

	if err = p.db.SelectContext(ctx, &records, p.sqlSelectAuthenticationLogsRegulationRecordsByRemoteIP, since, ip, false, limit); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting regulation records for ip '%s': %w", ip, err)
	}

	return records, nil
}

func (p *SQLProvider) SaveBannedIP(ctx context.Context, ban *model.BannedIP) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertBannedIP, ban.Expires, ban.IP, ban.Source, ban.Reason); err != nil {
		return fmt.Errorf("error inserting banned ip with ip '%s' and source '%s' and reason '%s': %w", ban.IP, ban.Source, ban.Reason.String, err)
	}

	return nil
}

func (p *SQLProvider) LoadBannedIP(ctx context.Context, ip model.IP) (bans []model.BannedIP, err error) {
	bans = []model.BannedIP{}

	if err = p.db.SelectContext(ctx, &bans, p.sqlSelectBannedIP, ip, false, time.Now()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting banned ip records for ip '%s': %w", ip, err)
	}

	return bans, nil
}

func (p *SQLProvider) LoadBannedIPByID(ctx context.Context, id int) (ban model.BannedIP, err error) {
	if err = p.db.GetContext(ctx, &ban, p.sqlSelectBannedIPByID, id); err != nil {
		return model.BannedIP{}, fmt.Errorf("error selecting banned ip with id '%d': %w", id, err)
	}

	return ban, nil
}

func (p *SQLProvider) LoadBannedIPs(ctx context.Context, limit, page int) (bans []model.BannedIP, err error) {
	bans = []model.BannedIP{}

	if err = p.db.SelectContext(ctx, &bans, p.sqlSelectBannedIPs, time.Now(), limit, limit*page); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting banned ip records: %w", err)
	}

	return bans, nil
}

func (p *SQLProvider) RevokeBannedIP(ctx context.Context, id int, expired time.Time) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlRevokeBannedIP, expired, id); err != nil {
		return fmt.Errorf("error revoking banned ip with id '%d': %w", id, err)
	}

	return nil
}

func (p *SQLProvider) SaveCachedData(ctx context.Context, data model.CachedData) (err error) {
	if data.Encrypted {
		if data.Value, err = p.encrypt(data.Value); err != nil {
			return fmt.Errorf("error encrypting cached data name '%s': %w", data.Name, err)
		}
	}

	if _, err = p.db.ExecContext(ctx, p.sqlUpsertCachedData, data.Name, time.Now(), data.Encrypted, data.Value); err != nil {
		return fmt.Errorf("error inserting cached data with name '%s': %w", data.Name, err)
	}

	return nil
}

func (p *SQLProvider) LoadCachedData(ctx context.Context, name string) (data *model.CachedData, err error) {
	data = &model.CachedData{}

	if err = p.db.GetContext(ctx, data, p.sqlSelectCachedData, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting cached data with name '%s': %w", name, err)
	}

	if data.Encrypted {
		if data.Value, err = p.decrypt(data.Value); err != nil {
			return nil, fmt.Errorf("error decrypting cached data with name '%s': %w", name, err)
		}
	}

	return data, nil
}

func (p *SQLProvider) DeleteCachedData(ctx context.Context, name string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteCachedData, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}

		return fmt.Errorf("error deleting cached data with name '%s': %w", name, err)
	}

	return nil
}

var (
	_ Provider = (*SQLProvider)(nil)
)
