package storage

import (
	"context"
	"database/sql"
	"time"

	"authelia.com/provider/oauth2/storage"
	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/model"
)

// Provider is an interface providing storage capabilities for persisting any kind of data related to Authelia.
type Provider interface {
	model.StartupCheck

	storage.Transactional

	// Close the underlying storage provider.
	Close() (err error)

	/*
		Implementation for Basic User Information.
	*/

	// SavePreferred2FAMethod save the preferred method for 2FA for a username to the storage provider.
	SavePreferred2FAMethod(ctx context.Context, username string, method string) (err error)

	// LoadPreferred2FAMethod load the preferred method for 2FA for a username from the storage provider.
	LoadPreferred2FAMethod(ctx context.Context, username string) (method string, err error)

	// LoadUserInfo loads the model.UserInfo from the storage provider.
	LoadUserInfo(ctx context.Context, username string) (info model.UserInfo, err error)

	/*
		Implementation for User Opaque Identifiers.
	*/

	// SaveUserOpaqueIdentifier saves a new opaque user identifier to the storage provider.
	SaveUserOpaqueIdentifier(ctx context.Context, subject model.UserOpaqueIdentifier) (err error)

	// LoadUserOpaqueIdentifier selects an opaque user identifier from the storage provider.
	LoadUserOpaqueIdentifier(ctx context.Context, identifier uuid.UUID) (subject *model.UserOpaqueIdentifier, err error)

	// LoadUserOpaqueIdentifiers selects an opaque user identifiers from the storage provider.
	LoadUserOpaqueIdentifiers(ctx context.Context) (identifiers []model.UserOpaqueIdentifier, err error)

	// LoadUserOpaqueIdentifierBySignature selects an opaque user identifier from the storage provider given a service
	// name, sector id, and username.
	LoadUserOpaqueIdentifierBySignature(ctx context.Context, service, sectorID, username string) (subject *model.UserOpaqueIdentifier, err error)

	/*
		Implementation for User TOTP Configurations.
	*/

	// SaveTOTPConfiguration save a TOTP configuration of a given user in the storage provider.
	SaveTOTPConfiguration(ctx context.Context, config model.TOTPConfiguration) (err error)

	// UpdateTOTPConfigurationSignIn updates a registered TOTP configuration in the storage provider with the relevant
	// sign in information.
	UpdateTOTPConfigurationSignIn(ctx context.Context, id int, lastUsedAt sql.NullTime) (err error)

	// DeleteTOTPConfiguration delete a TOTP configuration from the storage provider given a username.
	DeleteTOTPConfiguration(ctx context.Context, username string) (err error)

	// LoadTOTPConfiguration load a TOTP configuration given a username from the storage provider.
	LoadTOTPConfiguration(ctx context.Context, username string) (config *model.TOTPConfiguration, err error)

	// LoadTOTPConfigurations load a set of TOTP configurations from the storage provider.
	LoadTOTPConfigurations(ctx context.Context, limit, page int) (configs []model.TOTPConfiguration, err error)

	/*
		Implementation for User TOTP History.
	*/

	// SaveTOTPHistory saves a TOTP history item in the storage provider.
	SaveTOTPHistory(ctx context.Context, username string, step uint64) (err error)

	// ExistsTOTPHistory checks if a TOTP history item exists in the storage provider.
	ExistsTOTPHistory(ctx context.Context, username string, step uint64) (exists bool, err error)

	/*
		Implementation for User WebAuthn Information.
	*/

	// SaveWebAuthnUser saves a registered WebAuthn user to the storage provider.
	SaveWebAuthnUser(ctx context.Context, user model.WebAuthnUser) (err error)

	// LoadWebAuthnUser loads a registered WebAuthn user from the storage provider.
	LoadWebAuthnUser(ctx context.Context, rpid, username string) (user *model.WebAuthnUser, err error)

	// LoadWebAuthnUserByUserID loads a registered WebAuthn user from the storage provider.
	LoadWebAuthnUserByUserID(ctx context.Context, rpid, userID string) (user *model.WebAuthnUser, err error)

	/*
		Implementation for User WebAuthn Device Registrations.
	*/

	// SaveWebAuthnCredential saves a registered WebAuthn credential to the storage provider.
	SaveWebAuthnCredential(ctx context.Context, credential model.WebAuthnCredential) (err error)

	// UpdateWebAuthnCredentialDescription updates a registered WebAuthn credential in the storage provider changing the
	// description.
	UpdateWebAuthnCredentialDescription(ctx context.Context, username string, credentialID int, description string) (err error)

	// UpdateWebAuthnCredentialSignIn updates a registered WebAuthn credential in the storage provider changing the
	// information that should be changed in the event of a successful sign in.
	UpdateWebAuthnCredentialSignIn(ctx context.Context, credential model.WebAuthnCredential) (err error)

	// DeleteWebAuthnCredential deletes a registered WebAuthn credential from the storage provider.
	DeleteWebAuthnCredential(ctx context.Context, kid string) (err error)

	// DeleteWebAuthnCredentialByUsername deletes registered WebAuthn credential from the storage provider by username
	// or username and description.
	DeleteWebAuthnCredentialByUsername(ctx context.Context, username, description string) (err error)

	// LoadWebAuthnCredentials loads WebAuthn credential registrations from the storage provider.
	LoadWebAuthnCredentials(ctx context.Context, limit, page int) (credentials []model.WebAuthnCredential, err error)

	// LoadWebAuthnCredentialsByUsername loads all WebAuthn credential registrations from the storage provider for a
	// given username.
	LoadWebAuthnCredentialsByUsername(ctx context.Context, rpid, username string) (credential []model.WebAuthnCredential, err error)

	// LoadWebAuthnPasskeyCredentialsByUsername loads passkey WebAuthn credential registrations from the storage provider
	// for a given username.
	LoadWebAuthnPasskeyCredentialsByUsername(ctx context.Context, rpid, username string) (credentials []model.WebAuthnCredential, err error)

	// LoadWebAuthnCredentialByID loads a WebAuthn credential registration from the storage provider for a given id.
	LoadWebAuthnCredentialByID(ctx context.Context, id int) (credential *model.WebAuthnCredential, err error)

	// SavePreferredDuoDevice saves a Duo device to the storage provider.
	SavePreferredDuoDevice(ctx context.Context, device model.DuoDevice) (err error)

	// DeletePreferredDuoDevice deletes a Duo device from the storage provider for a given username.
	DeletePreferredDuoDevice(ctx context.Context, username string) (err error)

	// LoadPreferredDuoDevice loads a Duo device from the storage provider for a given username.
	LoadPreferredDuoDevice(ctx context.Context, username string) (device *model.DuoDevice, err error)

	/*
		Implementation for Identity Verification (JWT).
	*/

	// SaveIdentityVerification save an identity verification record to the storage provider.
	SaveIdentityVerification(ctx context.Context, verification model.IdentityVerification) (err error)

	// ConsumeIdentityVerification marks an identity verification record in the storage provider as consumed.
	ConsumeIdentityVerification(ctx context.Context, jti string, ip model.NullIP) (err error)

	// RevokeIdentityVerification marks an identity verification record in the storage provider as revoked.
	RevokeIdentityVerification(ctx context.Context, jti string, ip model.NullIP) (err error)

	// FindIdentityVerification checks if an identity verification record is in the storage provider and active.
	FindIdentityVerification(ctx context.Context, jti string) (found bool, err error)

	// LoadIdentityVerification loads an Identity Verification but does not do any validation.
	// For easy validation you should use FindIdentityVerification which ensures the JWT is still valid.
	LoadIdentityVerification(ctx context.Context, jti string) (verification *model.IdentityVerification, err error)

	/*
		Implementation for Identity Verification (OTP).
	*/

	// SaveOneTimeCode saves a one-time code to the storage provider after generating the signature which is returned
	// along with any error.
	SaveOneTimeCode(ctx context.Context, code model.OneTimeCode) (signature string, err error)

	// ConsumeOneTimeCode consumes a one-time code using the signature.
	ConsumeOneTimeCode(ctx context.Context, code *model.OneTimeCode) (err error)

	// RevokeOneTimeCode revokes a one-time code in the storage provider using the public ID.
	RevokeOneTimeCode(ctx context.Context, id uuid.UUID, ip model.IP) (err error)

	// LoadOneTimeCode loads a one-time code from the storage provider given a username, intent, and code.
	LoadOneTimeCode(ctx context.Context, username, intent, raw string) (code *model.OneTimeCode, err error)

	// LoadOneTimeCodeBySignature loads a one-time code from the storage provider given the signature.
	// This method should NOT be used to validate a One-Time Code, LoadOneTimeCode should be used instead.
	LoadOneTimeCodeBySignature(ctx context.Context, signature string) (code *model.OneTimeCode, err error)

	// LoadOneTimeCodeByID loads a one-time code from the storage provider given the id.
	// This does not decrypt the code. This method should NOT be used to validate a One-Time Code,
	// LoadOneTimeCode should be used instead.
	LoadOneTimeCodeByID(ctx context.Context, id int) (code *model.OneTimeCode, err error)

	// LoadOneTimeCodeByPublicID loads a one-time code from the storage provider given the public identifier.
	// This does not decrypt the code. This method SHOULD ONLY be used to find the One-Time Code for the
	// purpose of deletion.
	LoadOneTimeCodeByPublicID(ctx context.Context, id uuid.UUID) (code *model.OneTimeCode, err error)

	/*
		Implementation for OAuth2.0 Consent Pre-Configurations.
	*/

	// SaveOAuth2ConsentPreConfiguration inserts an OAuth2.0 consent pre-configuration in the storage provider.
	SaveOAuth2ConsentPreConfiguration(ctx context.Context, config model.OAuth2ConsentPreConfig) (insertedID int64, err error)

	// LoadOAuth2ConsentPreConfigurations returns an OAuth2.0 consents pre-configurations from the storage provider given the consent signature.
	LoadOAuth2ConsentPreConfigurations(ctx context.Context, clientID string, subject uuid.UUID, now time.Time) (rows *ConsentPreConfigRows, err error)

	/*
		Implementation for OAuth2.0 Consent Sessions.
	*/

	// SaveOAuth2ConsentSession inserts an OAuth2.0 consent session to the storage provider.
	SaveOAuth2ConsentSession(ctx context.Context, consent *model.OAuth2ConsentSession) (err error)

	// SaveOAuth2ConsentSessionResponse updates an OAuth2.0 consent session in the storage provider with the response.
	SaveOAuth2ConsentSessionResponse(ctx context.Context, consent *model.OAuth2ConsentSession, rejection bool) (err error)

	// SaveOAuth2ConsentSessionGranted updates an OAuth2.0 consent session in the storage provider recording that it
	// has been granted by the authorization endpoint.
	SaveOAuth2ConsentSessionGranted(ctx context.Context, id int) (err error)

	// LoadOAuth2ConsentSessionByChallengeID returns an OAuth2.0 consent session in the storage provider given the
	// challenge ID.
	LoadOAuth2ConsentSessionByChallengeID(ctx context.Context, challengeID uuid.UUID) (consent *model.OAuth2ConsentSession, err error)

	/*
		Implementation for OAuth2.0 General Sessions.
	*/

	// SaveOAuth2Session saves an OAut2.0 session to the storage provider.
	SaveOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, session model.OAuth2Session) (err error)

	// RevokeOAuth2Session marks an OAuth2.0 session as revoked in the storage provider.
	RevokeOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, signature string) (err error)

	// RevokeOAuth2SessionByRequestID marks an OAuth2.0 session as revoked in the storage provider.
	RevokeOAuth2SessionByRequestID(ctx context.Context, sessionType OAuth2SessionType, requestID string) (err error)

	// DeactivateOAuth2Session marks an OAuth2.0 session as inactive in the storage provider.
	DeactivateOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, signature string) (err error)

	// DeactivateOAuth2SessionByRequestID marks an OAuth2.0 session as inactive in the storage provider.
	DeactivateOAuth2SessionByRequestID(ctx context.Context, sessionType OAuth2SessionType, requestID string) (err error)

	// LoadOAuth2Session loads an OAuth2.0 session from the storage provider.
	LoadOAuth2Session(ctx context.Context, sessionType OAuth2SessionType, signature string) (session *model.OAuth2Session, err error)

	/*
		Implementation for OAuth2.0 Device Code Sessions.
	*/

	// SaveOAuth2DeviceCodeSession saves an OAuth2.0 device code session to the storage provider.
	SaveOAuth2DeviceCodeSession(ctx context.Context, session *model.OAuth2DeviceCodeSession) (err error)

	// UpdateOAuth2DeviceCodeSession updates an OAuth2.0 device code session in the storage provider.
	UpdateOAuth2DeviceCodeSession(ctx context.Context, session *model.OAuth2DeviceCodeSession) (err error)

	// UpdateOAuth2DeviceCodeSessionData updates an OAuth2.0 device code session data in the storage provider.
	UpdateOAuth2DeviceCodeSessionData(ctx context.Context, session *model.OAuth2DeviceCodeSession) (err error)

	// DeactivateOAuth2DeviceCodeSession marks an OAuth2.0 device code session as inactive in the storage provider.
	DeactivateOAuth2DeviceCodeSession(ctx context.Context, signature string) (err error)

	// LoadOAuth2DeviceCodeSession loads an OAuth2.0 device code session from the storage provider given the signature
	// of the device code.
	LoadOAuth2DeviceCodeSession(ctx context.Context, signature string) (session *model.OAuth2DeviceCodeSession, err error)

	// LoadOAuth2DeviceCodeSessionByUserCode loads an OAuth2.0 device code session from the storage provider given the
	// signature of a user code.
	LoadOAuth2DeviceCodeSessionByUserCode(ctx context.Context, signature string) (session *model.OAuth2DeviceCodeSession, err error)

	/*
		Implementation for OAuth2.0 PAR Contexts.
	*/

	// SaveOAuth2PARContext save an OAuth2.0 PAR context to the storage provider.
	SaveOAuth2PARContext(ctx context.Context, par model.OAuth2PARContext) (err error)

	// LoadOAuth2PARContext loads an OAuth2.0 PAR context from the storage provider.
	LoadOAuth2PARContext(ctx context.Context, signature string) (par *model.OAuth2PARContext, err error)

	// RevokeOAuth2PARContext marks an OAuth2.0 PAR context as revoked in the storage provider.
	RevokeOAuth2PARContext(ctx context.Context, signature string) (err error)

	// UpdateOAuth2PARContext updates an existing OAuth2.0 PAR context in the storage provider.
	UpdateOAuth2PARContext(ctx context.Context, par model.OAuth2PARContext) (err error)

	/*
		Implementation for OAuth2.0 Blacklisted JTI's.
	*/

	// SaveOAuth2BlacklistedJTI saves an OAuth2.0 blacklisted JTI to the storage provider.
	SaveOAuth2BlacklistedJTI(ctx context.Context, blacklistedJTI model.OAuth2BlacklistedJTI) (err error)

	// LoadOAuth2BlacklistedJTI loads an OAuth2.0 blacklisted JTI from the storage provider.
	LoadOAuth2BlacklistedJTI(ctx context.Context, signature string) (blacklistedJTI *model.OAuth2BlacklistedJTI, err error)

	/*
		Implementation for Schema controls.
	*/

	// SchemaTables returns a list of tables from the storage provider.
	SchemaTables(ctx context.Context) (tables []string, err error)

	// SchemaVersion returns the version of the schema from the storage provider.
	SchemaVersion(ctx context.Context) (version int, err error)

	// SchemaLatestVersion returns the latest version available for migration for the storage provider.
	SchemaLatestVersion() (version int, err error)

	// SchemaMigrationHistory returns the storage provider migration history rows.
	SchemaMigrationHistory(ctx context.Context) (migrations []model.Migration, err error)

	// SchemaMigrationsUp returns a list of storage provider up migrations available between the current version
	// and the provided version.
	SchemaMigrationsUp(ctx context.Context, version int) (migrations []model.SchemaMigration, err error)

	// SchemaMigrationsDown returns a list of storage provider down migrations available between the current version
	// and the provided version.
	SchemaMigrationsDown(ctx context.Context, version int) (migrations []model.SchemaMigration, err error)

	// SchemaMigrate migrates from the storage provider's current schema version to the provided schema version.
	SchemaMigrate(ctx context.Context, up bool, version int) (err error)

	// SchemaEncryptionChangeKey uses the currently configured key to decrypt values in the storage provider and the key
	// provided by this command to encrypt the values again and update them using a transaction.
	SchemaEncryptionChangeKey(ctx context.Context, key string) (err error)

	// SchemaEncryptionCheckKey checks the encryption key configured is valid for the storage provider.
	SchemaEncryptionCheckKey(ctx context.Context, verbose bool) (result EncryptionValidationResult, err error)

	RegulatorProvider
	CachedDataProvider
	SessionProvider
}

// SessionProvider is an interface providing storage capabilities for HTTP session data.
type SessionProvider interface {
	// SaveSession saves or updates a session in the database.
	SaveSession(ctx context.Context, sessionID string, data []byte, lastActiveAt, expiresAt time.Time) (err error)

	// LoadSession loads session data from the database.
	LoadSession(ctx context.Context, sessionID string) (data []byte, err error)

	// DeleteSession deletes a session from the database.
	DeleteSession(ctx context.Context, sessionID string) (err error)

	// DeleteExpiredSessions removes all expired sessions from the database.
	DeleteExpiredSessions(ctx context.Context) (err error)

	// CountSessions returns the count of non-expired sessions.
	CountSessions(ctx context.Context) (count int, err error)
}

type CachedDataProvider interface {
	// LoadCachedData loads cached data from the database.
	LoadCachedData(ctx context.Context, name string) (data *model.CachedData, err error)

	// SaveCachedData saves cached data to the database.
	SaveCachedData(ctx context.Context, data model.CachedData) (err error)

	// DeleteCachedData deletes cached data from the database.
	DeleteCachedData(ctx context.Context, name string) (err error)
}

// RegulatorProvider is an interface providing storage capabilities for persisting any kind of data related to the regulator.
type RegulatorProvider interface {
	// AppendAuthenticationLog saves an authentication attempt to the storage provider.
	AppendAuthenticationLog(ctx context.Context, attempt model.AuthenticationAttempt) (err error)

	// LoadRegulationRecordsByUser loads authentication logs for a given username for the purpose of regulation. As such
	// compared to standard authentication logs the amount of available data is very low.
	LoadRegulationRecordsByUser(ctx context.Context, username string, since time.Time, limit int) (records []model.RegulationRecord, err error)

	// SaveBannedUser saves a banned user to the database.
	SaveBannedUser(ctx context.Context, ban *model.BannedUser) (err error)

	// LoadBannedUser loads banned users from the database given a username.
	LoadBannedUser(ctx context.Context, username string) (bans []model.BannedUser, err error)

	// LoadBannedUserByID loads a banned user record given an id.
	LoadBannedUserByID(ctx context.Context, id int) (ban model.BannedUser, err error)

	// LoadBannedUsers loads pages of banned users from the database.
	LoadBannedUsers(ctx context.Context, limit, page int) (bans []model.BannedUser, err error)

	// RevokeBannedUser revokes a user ban in the database.
	RevokeBannedUser(ctx context.Context, id int, expired time.Time) (err error)

	// LoadRegulationRecordsByIP loads authentication logs for a given ip for the purpose of regulation. As such
	// compared to standard authentication logs the amount of available data is very low.
	LoadRegulationRecordsByIP(ctx context.Context, ip model.IP, since time.Time, limit int) (records []model.RegulationRecord, err error)

	// SaveBannedIP saves a banned ip to the database.
	SaveBannedIP(ctx context.Context, ban *model.BannedIP) (err error)

	// LoadBannedIP loads banned ip's from the database given an ip.
	LoadBannedIP(ctx context.Context, remoteIP model.IP) (bans []model.BannedIP, err error)

	// LoadBannedIPByID loads a banned ip record given an id.
	LoadBannedIPByID(ctx context.Context, id int) (ban model.BannedIP, err error)

	// LoadBannedIPs loads pages of banned ip's from the database.
	LoadBannedIPs(ctx context.Context, limit, page int) (bans []model.BannedIP, err error)

	// RevokeBannedIP revokes an ip ban in the database.
	RevokeBannedIP(ctx context.Context, id int, expired time.Time) (err error)
}
