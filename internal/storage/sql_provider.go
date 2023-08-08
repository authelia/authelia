package storage

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/model"
)

// NewSQLProvider generates a generic SQLProvider to be used with other SQL provider NewUp's.
func NewSQLProvider(config *schema.Configuration, name, driverName, dataSourceName string) (provider SQLProvider) {
	db, err := sqlx.Open(driverName, dataSourceName)

	provider = SQLProvider{
		db:         db,
		key:        sha256.Sum256([]byte(config.Storage.EncryptionKey)),
		name:       name,
		driverName: driverName,
		config:     config,
		errOpen:    err,
		log:        logging.Logger(),

		sqlInsertAuthenticationAttempt:            fmt.Sprintf(queryFmtInsertAuthenticationLogEntry, tableAuthenticationLogs),
		sqlSelectAuthenticationAttemptsByUsername: fmt.Sprintf(queryFmtSelect1FAAuthenticationLogEntryByUsername, tableAuthenticationLogs),

		sqlInsertIdentityVerification:  fmt.Sprintf(queryFmtInsertIdentityVerification, tableIdentityVerification),
		sqlConsumeIdentityVerification: fmt.Sprintf(queryFmtConsumeIdentityVerification, tableIdentityVerification),
		sqlSelectIdentityVerification:  fmt.Sprintf(queryFmtSelectIdentityVerification, tableIdentityVerification),

		sqlUpsertTOTPConfig:  fmt.Sprintf(queryFmtUpsertTOTPConfiguration, tableTOTPConfigurations),
		sqlDeleteTOTPConfig:  fmt.Sprintf(queryFmtDeleteTOTPConfiguration, tableTOTPConfigurations),
		sqlSelectTOTPConfig:  fmt.Sprintf(queryFmtSelectTOTPConfiguration, tableTOTPConfigurations),
		sqlSelectTOTPConfigs: fmt.Sprintf(queryFmtSelectTOTPConfigurations, tableTOTPConfigurations),

		sqlUpdateTOTPConfigRecordSignIn:           fmt.Sprintf(queryFmtUpdateTOTPConfigRecordSignIn, tableTOTPConfigurations),
		sqlUpdateTOTPConfigRecordSignInByUsername: fmt.Sprintf(queryFmtUpdateTOTPConfigRecordSignInByUsername, tableTOTPConfigurations),

		sqlUpsertWebAuthnDevice:            fmt.Sprintf(queryFmtUpsertWebAuthnDevice, tableWebAuthnDevices),
		sqlSelectWebAuthnDevices:           fmt.Sprintf(queryFmtSelectWebAuthnDevices, tableWebAuthnDevices),
		sqlSelectWebAuthnDevicesByUsername: fmt.Sprintf(queryFmtSelectWebAuthnDevicesByUsername, tableWebAuthnDevices),

		sqlUpdateWebAuthnDeviceRecordSignIn:           fmt.Sprintf(queryFmtUpdateWebAuthnDeviceRecordSignIn, tableWebAuthnDevices),
		sqlUpdateWebAuthnDeviceRecordSignInByUsername: fmt.Sprintf(queryFmtUpdateWebAuthnDeviceRecordSignInByUsername, tableWebAuthnDevices),

		sqlDeleteWebAuthnDevice:                         fmt.Sprintf(queryFmtDeleteWebAuthnDevice, tableWebAuthnDevices),
		sqlDeleteWebAuthnDeviceByUsername:               fmt.Sprintf(queryFmtDeleteWebAuthnDeviceByUsername, tableWebAuthnDevices),
		sqlDeleteWebAuthnDeviceByUsernameAndDescription: fmt.Sprintf(queryFmtDeleteWebAuthnDeviceByUsernameAndDescription, tableWebAuthnDevices),

		sqlUpsertDuoDevice: fmt.Sprintf(queryFmtUpsertDuoDevice, tableDuoDevices),
		sqlDeleteDuoDevice: fmt.Sprintf(queryFmtDeleteDuoDevice, tableDuoDevices),
		sqlSelectDuoDevice: fmt.Sprintf(queryFmtSelectDuoDevice, tableDuoDevices),

		sqlUpsertPreferred2FAMethod: fmt.Sprintf(queryFmtUpsertPreferred2FAMethod, tableUserPreferences),
		sqlSelectPreferred2FAMethod: fmt.Sprintf(queryFmtSelectPreferred2FAMethod, tableUserPreferences),
		sqlSelectUserInfo:           fmt.Sprintf(queryFmtSelectUserInfo, tableTOTPConfigurations, tableWebAuthnDevices, tableDuoDevices, tableUserPreferences),

		sqlInsertUserOpaqueIdentifier:            fmt.Sprintf(queryFmtInsertUserOpaqueIdentifier, tableUserOpaqueIdentifier),
		sqlSelectUserOpaqueIdentifier:            fmt.Sprintf(queryFmtSelectUserOpaqueIdentifier, tableUserOpaqueIdentifier),
		sqlSelectUserOpaqueIdentifiers:           fmt.Sprintf(queryFmtSelectUserOpaqueIdentifiers, tableUserOpaqueIdentifier),
		sqlSelectUserOpaqueIdentifierBySignature: fmt.Sprintf(queryFmtSelectUserOpaqueIdentifierBySignature, tableUserOpaqueIdentifier),

		sqlUpsertOAuth2BlacklistedJTI: fmt.Sprintf(queryFmtUpsertOAuth2BlacklistedJTI, tableOAuth2BlacklistedJTI),
		sqlSelectOAuth2BlacklistedJTI: fmt.Sprintf(queryFmtSelectOAuth2BlacklistedJTI, tableOAuth2BlacklistedJTI),

		sqlInsertOAuth2PARContext: fmt.Sprintf(queryFmtInsertOAuth2PARContext, tableOAuth2PARContext),
		sqlSelectOAuth2PARContext: fmt.Sprintf(queryFmtSelectOAuth2PARContext, tableOAuth2PARContext),
		sqlRevokeOAuth2PARContext: fmt.Sprintf(queryFmtRevokeOAuth2Session, tableOAuth2PARContext),

		sqlInsertOAuth2ConsentPreConfiguration:  fmt.Sprintf(queryFmtInsertOAuth2ConsentPreConfiguration, tableOAuth2ConsentPreConfiguration),
		sqlSelectOAuth2ConsentPreConfigurations: fmt.Sprintf(queryFmtSelectOAuth2ConsentPreConfigurations, tableOAuth2ConsentPreConfiguration),

		sqlInsertOAuth2ConsentSession:              fmt.Sprintf(queryFmtInsertOAuth2ConsentSession, tableOAuth2ConsentSession),
		sqlUpdateOAuth2ConsentSessionSubject:       fmt.Sprintf(queryFmtUpdateOAuth2ConsentSessionSubject, tableOAuth2ConsentSession),
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
	db         *sqlx.DB
	key        [32]byte
	name       string
	driverName string
	schema     string
	config     *schema.Configuration
	errOpen    error

	log *logrus.Logger

	// Table: authentication_logs.
	sqlInsertAuthenticationAttempt            string
	sqlSelectAuthenticationAttemptsByUsername string

	// Table: identity_verification.
	sqlInsertIdentityVerification  string
	sqlConsumeIdentityVerification string
	sqlSelectIdentityVerification  string

	// Table: totp_configurations.
	sqlUpsertTOTPConfig  string
	sqlDeleteTOTPConfig  string
	sqlSelectTOTPConfig  string
	sqlSelectTOTPConfigs string

	sqlUpdateTOTPConfigRecordSignIn           string
	sqlUpdateTOTPConfigRecordSignInByUsername string

	// Table: webauthn_devices.
	sqlUpsertWebAuthnDevice            string
	sqlSelectWebAuthnDevices           string
	sqlSelectWebAuthnDevicesByUsername string

	sqlUpdateWebAuthnDeviceRecordSignIn           string
	sqlUpdateWebAuthnDeviceRecordSignInByUsername string

	sqlDeleteWebAuthnDevice                         string
	sqlDeleteWebAuthnDeviceByUsername               string
	sqlDeleteWebAuthnDeviceByUsernameAndDescription string

	// Table: duo_devices.
	sqlUpsertDuoDevice string
	sqlDeleteDuoDevice string
	sqlSelectDuoDevice string

	// Table: user_preferences.
	sqlUpsertPreferred2FAMethod string
	sqlSelectPreferred2FAMethod string
	sqlSelectUserInfo           string

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
	sqlUpdateOAuth2ConsentSessionSubject       string
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

// Close the underlying database connection.
func (p *SQLProvider) Close() (err error) {
	return p.db.Close()
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
	case ErrSchemaAlreadyUpToDate:
		p.log.Infof("Storage schema is already up to date")
		return nil
	case nil:
		return nil
	default:
		return fmt.Errorf("error during schema migrate: %w", err)
	}
}

// BeginTX begins a transaction.
func (p *SQLProvider) BeginTX(ctx context.Context) (c context.Context, err error) {
	var tx *sql.Tx

	if tx, err = p.db.Begin(); err != nil {
		return nil, err
	}

	return context.WithValue(ctx, ctxKeyTransaction, tx), nil
}

// Commit performs a database commit.
func (p *SQLProvider) Commit(ctx context.Context) (err error) {
	tx, ok := ctx.Value(ctxKeyTransaction).(*sql.Tx)

	if !ok {
		return errors.New("could not retrieve tx")
	}

	return tx.Commit()
}

// Rollback performs a database rollback.
func (p *SQLProvider) Rollback(ctx context.Context) (err error) {
	tx, ok := ctx.Value(ctxKeyTransaction).(*sql.Tx)

	if !ok {
		return errors.New("could not retrieve tx")
	}

	return tx.Rollback()
}

// SaveUserOpaqueIdentifier saves a new opaque user identifier to the database.
func (p *SQLProvider) SaveUserOpaqueIdentifier(ctx context.Context, subject model.UserOpaqueIdentifier) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertUserOpaqueIdentifier, subject.Service, subject.SectorID, subject.Username, subject.Identifier); err != nil {
		return fmt.Errorf("error inserting user opaque id for user '%s' with opaque id '%s': %w", subject.Username, subject.Identifier.String(), err)
	}

	return nil
}

// LoadUserOpaqueIdentifier selects an opaque user identifier from the database.
func (p *SQLProvider) LoadUserOpaqueIdentifier(ctx context.Context, identifier uuid.UUID) (subject *model.UserOpaqueIdentifier, err error) {
	subject = &model.UserOpaqueIdentifier{}

	if err = p.db.GetContext(ctx, subject, p.sqlSelectUserOpaqueIdentifier, identifier); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, nil
		default:
			return nil, err
		}
	}

	return subject, nil
}

// LoadUserOpaqueIdentifiers selects an opaque user identifiers from the database.
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

// LoadUserOpaqueIdentifierBySignature selects an opaque user identifier from the database given a service name, sector id, and username.
func (p *SQLProvider) LoadUserOpaqueIdentifierBySignature(ctx context.Context, service, sectorID, username string) (subject *model.UserOpaqueIdentifier, err error) {
	subject = &model.UserOpaqueIdentifier{}

	if err = p.db.GetContext(ctx, subject, p.sqlSelectUserOpaqueIdentifierBySignature, service, sectorID, username); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, nil
		default:
			return nil, err
		}
	}

	return subject, nil
}

// SaveOAuth2ConsentSession inserts an OAuth2.0 consent session.
func (p *SQLProvider) SaveOAuth2ConsentSession(ctx context.Context, consent model.OAuth2ConsentSession) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertOAuth2ConsentSession,
		consent.ChallengeID, consent.ClientID, consent.Subject, consent.Authorized, consent.Granted,
		consent.RequestedAt, consent.RespondedAt, consent.Form,
		consent.RequestedScopes, consent.GrantedScopes, consent.RequestedAudience, consent.GrantedAudience, consent.PreConfiguration); err != nil {
		return fmt.Errorf("error inserting oauth2 consent session with challenge id '%s' for subject '%s': %w", consent.ChallengeID.String(), consent.Subject.UUID.String(), err)
	}

	return nil
}

// SaveOAuth2ConsentSessionSubject updates an OAuth2.0 consent session with the subject.
func (p *SQLProvider) SaveOAuth2ConsentSessionSubject(ctx context.Context, consent model.OAuth2ConsentSession) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateOAuth2ConsentSessionSubject, consent.Subject, consent.ID); err != nil {
		return fmt.Errorf("error updating oauth2 consent session subject with id '%d' and challenge id '%s' for subject '%s': %w", consent.ID, consent.ChallengeID, consent.Subject.UUID, err)
	}

	return nil
}

// SaveOAuth2ConsentSessionResponse updates an OAuth2.0 consent session with the response.
func (p *SQLProvider) SaveOAuth2ConsentSessionResponse(ctx context.Context, consent model.OAuth2ConsentSession, authorized bool) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateOAuth2ConsentSessionResponse, authorized, consent.GrantedScopes, consent.GrantedAudience, consent.PreConfiguration, consent.ID); err != nil {
		return fmt.Errorf("error updating oauth2 consent session (authorized  '%t') with id '%d' and challenge id '%s' for subject '%s': %w", authorized, consent.ID, consent.ChallengeID, consent.Subject.UUID, err)
	}

	return nil
}

// SaveOAuth2ConsentSessionGranted updates an OAuth2.0 consent recording that it has been granted by the authorization endpoint.
func (p *SQLProvider) SaveOAuth2ConsentSessionGranted(ctx context.Context, id int) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateOAuth2ConsentSessionGranted, id); err != nil {
		return fmt.Errorf("error updating oauth2 consent session (granted) with id '%d': %w", id, err)
	}

	return nil
}

// LoadOAuth2ConsentSessionByChallengeID returns an OAuth2.0 consent given the challenge ID.
func (p *SQLProvider) LoadOAuth2ConsentSessionByChallengeID(ctx context.Context, challengeID uuid.UUID) (consent *model.OAuth2ConsentSession, err error) {
	consent = &model.OAuth2ConsentSession{}

	if err = p.db.GetContext(ctx, consent, p.sqlSelectOAuth2ConsentSessionByChallengeID, challengeID); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 consent session with challenge id '%s': %w", challengeID.String(), err)
	}

	return consent, nil
}

// SaveOAuth2ConsentPreConfiguration inserts an OAuth2.0 consent pre-configuration.
func (p *SQLProvider) SaveOAuth2ConsentPreConfiguration(ctx context.Context, config model.OAuth2ConsentPreConfig) (insertedID int64, err error) {
	switch p.name {
	case providerPostgres:
		if err = p.db.GetContext(ctx, &insertedID, p.sqlInsertOAuth2ConsentPreConfiguration,
			config.ClientID, config.Subject, config.CreatedAt, config.ExpiresAt,
			config.Revoked, config.Scopes, config.Audience); err != nil {
			return -1, fmt.Errorf("error inserting oauth2 consent pre-configuration for subject '%s' with client id '%s' and scopes '%s': %w", config.Subject.String(), config.ClientID, strings.Join(config.Scopes, " "), err)
		}

		return insertedID, nil
	default:
		var result sql.Result

		if result, err = p.db.ExecContext(ctx, p.sqlInsertOAuth2ConsentPreConfiguration,
			config.ClientID, config.Subject, config.CreatedAt, config.ExpiresAt,
			config.Revoked, config.Scopes, config.Audience); err != nil {
			return -1, fmt.Errorf("error inserting oauth2 consent pre-configuration for subject '%s' with client id '%s' and scopes '%s': %w", config.Subject.String(), config.ClientID, strings.Join(config.Scopes, " "), err)
		}

		return result.LastInsertId()
	}
}

// LoadOAuth2ConsentPreConfigurations returns an OAuth2.0 consents pre-configurations given the consent signature.
func (p *SQLProvider) LoadOAuth2ConsentPreConfigurations(ctx context.Context, clientID string, subject uuid.UUID) (rows *ConsentPreConfigRows, err error) {
	var r *sqlx.Rows

	if r, err = p.db.QueryxContext(ctx, p.sqlSelectOAuth2ConsentPreConfigurations, clientID, subject); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &ConsentPreConfigRows{}, nil
		}

		return &ConsentPreConfigRows{}, fmt.Errorf("error selecting oauth2 consent pre-configurations by signature with client id '%s' and subject '%s': %w", clientID, subject.String(), err)
	}

	return &ConsentPreConfigRows{rows: r}, nil
}

// SaveOAuth2Session saves a OAuth2Session to the database.
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
		return fmt.Errorf("error inserting oauth2 %s session data for subject '%s' and request id '%s' and challenge id '%s': %w", sessionType, session.Subject.String, session.RequestID, session.ChallengeID.UUID, err)
	}

	return nil
}

// RevokeOAuth2Session marks a OAuth2Session as revoked in the database.
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
		return fmt.Errorf("error revoking oauth2 session with signature '%s': unknown oauth2 session type '%s'", signature, sessionType.String())
	}

	if _, err = p.db.ExecContext(ctx, query, signature); err != nil {
		return fmt.Errorf("error revoking oauth2 %s session with signature '%s': %w", sessionType.String(), signature, err)
	}

	return nil
}

// RevokeOAuth2SessionByRequestID marks a OAuth2Session as revoked in the database.
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
		return fmt.Errorf("error revoking oauth2 session with request id '%s': unknown oauth2 session type '%s'", requestID, sessionType.String())
	}

	if _, err = p.db.ExecContext(ctx, query, requestID); err != nil {
		return fmt.Errorf("error revoking oauth2 %s session with request id '%s': %w", sessionType.String(), requestID, err)
	}

	return nil
}

// DeactivateOAuth2Session marks a OAuth2Session as inactive in the database.
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
		return fmt.Errorf("error deactivating oauth2 session with signature '%s': unknown oauth2 session type '%s'", signature, sessionType.String())
	}

	if _, err = p.db.ExecContext(ctx, query, signature); err != nil {
		return fmt.Errorf("error deactivating oauth2 %s session with signature '%s': %w", sessionType.String(), signature, err)
	}

	return nil
}

// DeactivateOAuth2SessionByRequestID marks a OAuth2Session as inactive in the database.
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
		return fmt.Errorf("error deactivating oauth2 session with request id '%s': unknown oauth2 session type '%s'", requestID, sessionType.String())
	}

	if _, err = p.db.ExecContext(ctx, query, requestID); err != nil {
		return fmt.Errorf("error deactivating oauth2 %s session with request id '%s': %w", sessionType, requestID, err)
	}

	return nil
}

// LoadOAuth2Session saves a OAuth2Session from the database.
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
		return nil, fmt.Errorf("error selecting oauth2 session: unknown oauth2 session type '%s'", sessionType.String())
	}

	session = &model.OAuth2Session{}

	if err = p.db.GetContext(ctx, session, query, signature); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 %s session with signature '%s': %w", sessionType.String(), signature, err)
	}

	if session.Session, err = p.decrypt(session.Session); err != nil {
		return nil, fmt.Errorf("error decrypting the oauth2 %s session data with signature '%s' for subject '%s' and request id '%s': %w", sessionType.String(), signature, session.Subject.String, session.RequestID, err)
	}

	return session, nil
}

// SaveOAuth2PARContext save a OAuth2PARContext to the database.
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

// LoadOAuth2PARContext loads a OAuth2PARContext from the database.
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

// RevokeOAuth2PARContext marks a OAuth2PARContext as revoked in the database.
func (p *SQLProvider) RevokeOAuth2PARContext(ctx context.Context, signature string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlRevokeOAuth2PARContext, signature); err != nil {
		return fmt.Errorf("error revoking oauth2 pushed authorization request context with signature '%s': %w", signature, err)
	}

	return nil
}

// SaveOAuth2BlacklistedJTI saves a OAuth2BlacklistedJTI to the database.
func (p *SQLProvider) SaveOAuth2BlacklistedJTI(ctx context.Context, blacklistedJTI model.OAuth2BlacklistedJTI) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpsertOAuth2BlacklistedJTI, blacklistedJTI.Signature, blacklistedJTI.ExpiresAt); err != nil {
		return fmt.Errorf("error inserting oauth2 blacklisted JTI with signature '%s': %w", blacklistedJTI.Signature, err)
	}

	return nil
}

// LoadOAuth2BlacklistedJTI loads a OAuth2BlacklistedJTI from the database.
func (p *SQLProvider) LoadOAuth2BlacklistedJTI(ctx context.Context, signature string) (blacklistedJTI *model.OAuth2BlacklistedJTI, err error) {
	blacklistedJTI = &model.OAuth2BlacklistedJTI{}

	if err = p.db.GetContext(ctx, blacklistedJTI, p.sqlSelectOAuth2BlacklistedJTI, signature); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 blacklisted JTI with signature '%s': %w", blacklistedJTI.Signature, err)
	}

	return blacklistedJTI, nil
}

// SavePreferred2FAMethod save the preferred method for 2FA to the database.
func (p *SQLProvider) SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpsertPreferred2FAMethod, username, method); err != nil {
		return fmt.Errorf("error upserting preferred two factor method for user '%s': %w", username, err)
	}

	return nil
}

// LoadPreferred2FAMethod load the preferred method for 2FA from the database.
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

// LoadUserInfo loads the model.UserInfo from the database.
func (p *SQLProvider) LoadUserInfo(ctx context.Context, username string) (info model.UserInfo, err error) {
	err = p.db.GetContext(ctx, &info, p.sqlSelectUserInfo, username, username, username, username)

	switch {
	case err == nil, errors.Is(err, sql.ErrNoRows):
		return info, nil
	default:
		return model.UserInfo{}, fmt.Errorf("error selecting user info for user '%s': %w", username, err)
	}
}

// SaveIdentityVerification save an identity verification record to the database.
func (p *SQLProvider) SaveIdentityVerification(ctx context.Context, verification model.IdentityVerification) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertIdentityVerification,
		verification.JTI, verification.IssuedAt, verification.IssuedIP, verification.ExpiresAt,
		verification.Username, verification.Action); err != nil {
		return fmt.Errorf("error inserting identity verification for user '%s' with uuid '%s': %w", verification.Username, verification.JTI, err)
	}

	return nil
}

// ConsumeIdentityVerification marks an identity verification record in the database as consumed.
func (p *SQLProvider) ConsumeIdentityVerification(ctx context.Context, jti string, ip model.NullIP) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlConsumeIdentityVerification, ip, jti); err != nil {
		return fmt.Errorf("error updating identity verification: %w", err)
	}

	return nil
}

// FindIdentityVerification checks if an identity verification record is in the database and active.
func (p *SQLProvider) FindIdentityVerification(ctx context.Context, jti string) (found bool, err error) {
	verification := model.IdentityVerification{}
	if err = p.db.GetContext(ctx, &verification, p.sqlSelectIdentityVerification, jti); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("error selecting identity verification exists: %w", err)
	}

	switch {
	case verification.Consumed.Valid:
		return false, fmt.Errorf("the token has already been consumed")
	case verification.ExpiresAt.Before(time.Now()):
		return false, fmt.Errorf("the token expired %s ago", time.Since(verification.ExpiresAt))
	default:
		return true, nil
	}
}

// SaveTOTPConfiguration save a TOTP configuration of a given user in the database.
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

// UpdateTOTPConfigurationSignIn updates a registered WebAuthn devices sign in information.
func (p *SQLProvider) UpdateTOTPConfigurationSignIn(ctx context.Context, id int, lastUsedAt sql.NullTime) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateTOTPConfigRecordSignIn, lastUsedAt, id); err != nil {
		return fmt.Errorf("error updating TOTP configuration id %d: %w", id, err)
	}

	return nil
}

// DeleteTOTPConfiguration delete a TOTP configuration from the database given a username.
func (p *SQLProvider) DeleteTOTPConfiguration(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteTOTPConfig, username); err != nil {
		return fmt.Errorf("error deleting TOTP configuration for user '%s': %w", username, err)
	}

	return nil
}

// LoadTOTPConfiguration load a TOTP configuration given a username from the database.
func (p *SQLProvider) LoadTOTPConfiguration(ctx context.Context, username string) (config *model.TOTPConfiguration, err error) {
	config = &model.TOTPConfiguration{}

	if err = p.db.QueryRowxContext(ctx, p.sqlSelectTOTPConfig, username).StructScan(config); err != nil {
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

// LoadTOTPConfigurations load a set of TOTP configurations.
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

// SaveWebAuthnDevice saves a registered WebAuthn device.
func (p *SQLProvider) SaveWebAuthnDevice(ctx context.Context, device model.WebAuthnDevice) (err error) {
	if device.PublicKey, err = p.encrypt(device.PublicKey); err != nil {
		return fmt.Errorf("error encrypting WebAuthn device public key for user '%s' kid '%x': %w", device.Username, device.KID, err)
	}

	if _, err = p.db.ExecContext(ctx, p.sqlUpsertWebAuthnDevice,
		device.CreatedAt, device.LastUsedAt,
		device.RPID, device.Username, device.Description,
		device.KID, device.PublicKey,
		device.AttestationType, device.Transport, device.AAGUID, device.SignCount, device.CloneWarning,
	); err != nil {
		return fmt.Errorf("error upserting WebAuthn device for user '%s' kid '%x': %w", device.Username, device.KID, err)
	}

	return nil
}

// UpdateWebAuthnDeviceSignIn updates a registered WebAuthn devices sign in information.
func (p *SQLProvider) UpdateWebAuthnDeviceSignIn(ctx context.Context, id int, rpid string, lastUsedAt sql.NullTime, signCount uint32, cloneWarning bool) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateWebAuthnDeviceRecordSignIn, rpid, lastUsedAt, signCount, cloneWarning, id); err != nil {
		return fmt.Errorf("error updating WebAuthn signin metadata for id '%x': %w", id, err)
	}

	return nil
}

// DeleteWebAuthnDevice deletes a registered WebAuthn device.
func (p *SQLProvider) DeleteWebAuthnDevice(ctx context.Context, kid string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteWebAuthnDevice, kid); err != nil {
		return fmt.Errorf("error deleting WebAuthn device with kid '%s': %w", kid, err)
	}

	return nil
}

// DeleteWebAuthnDeviceByUsername deletes registered WebAuthn devices by username or username and description.
func (p *SQLProvider) DeleteWebAuthnDeviceByUsername(ctx context.Context, username, description string) (err error) {
	if len(username) == 0 {
		return fmt.Errorf("error deleting WebAuthn device with username '%s' and description '%s': username must not be empty", username, description)
	}

	if len(description) == 0 {
		if _, err = p.db.ExecContext(ctx, p.sqlDeleteWebAuthnDeviceByUsername, username); err != nil {
			return fmt.Errorf("error deleting WebAuthn devices for username '%s': %w", username, err)
		}
	} else {
		if _, err = p.db.ExecContext(ctx, p.sqlDeleteWebAuthnDeviceByUsernameAndDescription, username, description); err != nil {
			return fmt.Errorf("error deleting WebAuthn device with username '%s' and description '%s': %w", username, description, err)
		}
	}

	return nil
}

// LoadWebAuthnDevices loads WebAuthn device registrations.
func (p *SQLProvider) LoadWebAuthnDevices(ctx context.Context, limit, page int) (devices []model.WebAuthnDevice, err error) {
	devices = make([]model.WebAuthnDevice, 0, limit)

	if err = p.db.SelectContext(ctx, &devices, p.sqlSelectWebAuthnDevices, limit, limit*page); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting WebAuthn devices: %w", err)
	}

	for i, device := range devices {
		if devices[i].PublicKey, err = p.decrypt(device.PublicKey); err != nil {
			return nil, fmt.Errorf("error decrypting WebAuthn public key for user '%s': %w", device.Username, err)
		}
	}

	return devices, nil
}

// LoadWebAuthnDevicesByUsername loads all WebAuthn devices registration for a given username.
func (p *SQLProvider) LoadWebAuthnDevicesByUsername(ctx context.Context, username string) (devices []model.WebAuthnDevice, err error) {
	if err = p.db.SelectContext(ctx, &devices, p.sqlSelectWebAuthnDevicesByUsername, username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoWebAuthnDevice
		}

		return nil, fmt.Errorf("error selecting WebAuthn devices for user '%s': %w", username, err)
	}

	for i, device := range devices {
		if devices[i].PublicKey, err = p.decrypt(device.PublicKey); err != nil {
			return nil, fmt.Errorf("error decrypting WebAuthn public key for user '%s': %w", username, err)
		}
	}

	return devices, nil
}

// SavePreferredDuoDevice saves a Duo device.
func (p *SQLProvider) SavePreferredDuoDevice(ctx context.Context, device model.DuoDevice) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpsertDuoDevice, device.Username, device.Device, device.Method); err != nil {
		return fmt.Errorf("error upserting preferred duo device for user '%s': %w", device.Username, err)
	}

	return nil
}

// DeletePreferredDuoDevice deletes a Duo device of a given user.
func (p *SQLProvider) DeletePreferredDuoDevice(ctx context.Context, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteDuoDevice, username); err != nil {
		return fmt.Errorf("error deleting preferred duo device for user '%s': %w", username, err)
	}

	return nil
}

// LoadPreferredDuoDevice loads a Duo device of a given user.
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

// AppendAuthenticationLog append a mark to the authentication log.
func (p *SQLProvider) AppendAuthenticationLog(ctx context.Context, attempt model.AuthenticationAttempt) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlInsertAuthenticationAttempt,
		attempt.Time, attempt.Successful, attempt.Banned, attempt.Username,
		attempt.Type, attempt.RemoteIP, attempt.RequestURI, attempt.RequestMethod); err != nil {
		return fmt.Errorf("error inserting authentication attempt for user '%s': %w", attempt.Username, err)
	}

	return nil
}

// LoadAuthenticationLogs retrieve the latest failed authentications from the authentication log.
func (p *SQLProvider) LoadAuthenticationLogs(ctx context.Context, username string, fromDate time.Time, limit, page int) (attempts []model.AuthenticationAttempt, err error) {
	attempts = make([]model.AuthenticationAttempt, 0, limit)

	if err = p.db.SelectContext(ctx, &attempts, p.sqlSelectAuthenticationAttemptsByUsername, fromDate, username, limit, limit*page); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoAuthenticationLogs
		}

		return nil, fmt.Errorf("error selecting authentication logs for user '%s': %w", username, err)
	}

	return attempts, nil
}
