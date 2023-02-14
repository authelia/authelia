package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"
	"github.com/ory/fosite/storage"

	"github.com/authelia/authelia/v4/internal/model"
)

// Provider is an interface providing storage capabilities for persisting any kind of data related to Authelia.
type Provider interface {
	model.StartupCheck

	RegulatorProvider

	storage.Transactional

	SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error)
	LoadPreferred2FAMethod(ctx context.Context, username string) (method string, err error)
	LoadUserInfo(ctx context.Context, username string) (info model.UserInfo, err error)

	SaveUserOpaqueIdentifier(ctx context.Context, subject model.UserOpaqueIdentifier) (err error)
	LoadUserOpaqueIdentifier(ctx context.Context, opaqueUUID uuid.UUID) (subject *model.UserOpaqueIdentifier, err error)
	LoadUserOpaqueIdentifiers(ctx context.Context) (opaqueIDs []model.UserOpaqueIdentifier, err error)
	LoadUserOpaqueIdentifierBySignature(ctx context.Context, service, sectorID, username string) (subject *model.UserOpaqueIdentifier, err error)

	SaveIdentityVerification(ctx context.Context, verification model.IdentityVerification) (err error)
	ConsumeIdentityVerification(ctx context.Context, jti string, ip model.NullIP) (err error)
	FindIdentityVerification(ctx context.Context, jti string) (found bool, err error)

	SaveTOTPConfiguration(ctx context.Context, config model.TOTPConfiguration) (err error)
	UpdateTOTPConfigurationSignIn(ctx context.Context, id int, lastUsedAt sql.NullTime) (err error)
	DeleteTOTPConfiguration(ctx context.Context, username string) (err error)
	LoadTOTPConfiguration(ctx context.Context, username string) (config *model.TOTPConfiguration, err error)
	LoadTOTPConfigurations(ctx context.Context, limit, page int) (configs []model.TOTPConfiguration, err error)

	SaveWebauthnDevice(ctx context.Context, device model.WebauthnDevice) (err error)
	UpdateWebauthnDeviceDescription(ctx context.Context, username string, deviceID int, description string) (err error)
	UpdateWebauthnDeviceSignIn(ctx context.Context, device model.WebauthnDevice) (err error)
	DeleteWebauthnDevice(ctx context.Context, kid string) (err error)
	DeleteWebauthnDeviceByUsername(ctx context.Context, username, description string) (err error)
	LoadWebauthnDevices(ctx context.Context, limit, page int) (devices []model.WebauthnDevice, err error)
	LoadWebauthnDevicesByUsername(ctx context.Context, rpid, username string) (devices []model.WebauthnDevice, err error)
	LoadWebauthnDeviceByID(ctx context.Context, id int) (device *model.WebauthnDevice, err error)

	SavePreferredDuoDevice(ctx context.Context, device model.DuoDevice) (err error)
	DeletePreferredDuoDevice(ctx context.Context, username string) (err error)
	LoadPreferredDuoDevice(ctx context.Context, username string) (device *model.DuoDevice, err error)

	SaveOAuth2ConsentPreConfiguration(ctx context.Context, config model.OAuth2ConsentPreConfig) (insertedID int64, err error)
	LoadOAuth2ConsentPreConfigurations(ctx context.Context, clientID string, subject uuid.UUID) (rows *ConsentPreConfigRows, err error)

	SaveOAuth2ConsentSession(ctx context.Context, consent model.OAuth2ConsentSession) (err error)
	SaveOAuth2ConsentSessionSubject(ctx context.Context, consent model.OAuth2ConsentSession) (err error)
	SaveOAuth2ConsentSessionResponse(ctx context.Context, consent model.OAuth2ConsentSession, rejection bool) (err error)
	SaveOAuth2ConsentSessionGranted(ctx context.Context, id int) (err error)
	LoadOAuth2ConsentSessionByChallengeID(ctx context.Context, challengeID uuid.UUID) (consent *model.OAuth2ConsentSession, err error)

	SaveOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, session model.OAuth2Session) (err error)
	RevokeOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, signature string) (err error)
	RevokeOAuth2SessionByRequestID(ctx context.Context, sessionType OAuth2SessionType, requestID string) (err error)
	DeactivateOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, signature string) (err error)
	DeactivateOAuth2SessionByRequestID(ctx context.Context, sessionType OAuth2SessionType, requestID string) (err error)
	LoadOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, signature string) (session *model.OAuth2Session, err error)

	SaveOAuth2BlacklistedJTI(ctx context.Context, blacklistedJTI model.OAuth2BlacklistedJTI) (err error)
	LoadOAuth2BlacklistedJTI(ctx context.Context, signature string) (blacklistedJTI *model.OAuth2BlacklistedJTI, err error)

	SchemaTables(ctx context.Context) (tables []string, err error)
	SchemaVersion(ctx context.Context) (version int, err error)
	SchemaLatestVersion() (version int, err error)

	SchemaMigrate(ctx context.Context, up bool, version int) (err error)
	SchemaMigrationHistory(ctx context.Context) (migrations []model.Migration, err error)
	SchemaMigrationsUp(ctx context.Context, version int) (migrations []model.SchemaMigration, err error)
	SchemaMigrationsDown(ctx context.Context, version int) (migrations []model.SchemaMigration, err error)

	SchemaEncryptionChangeKey(ctx context.Context, key string) (err error)
	SchemaEncryptionCheckKey(ctx context.Context, verbose bool) (result EncryptionValidationResult, err error)

	Close() (err error)
}

// RegulatorProvider is an interface providing storage capabilities for persisting any kind of data related to the regulator.
type RegulatorProvider interface {
	AppendAuthenticationLog(ctx context.Context, attempt model.AuthenticationAttempt) (err error)
	LoadAuthenticationLogs(ctx context.Context, username string, fromDate time.Time, limit, page int) (attempts []model.AuthenticationAttempt, err error)
}
