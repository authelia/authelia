package storage

import (
	"regexp"
)

const (
	tableUserPreferences      = "user_preferences"
	tableIdentityVerification = "identity_verification"
	tableTOTPConfigurations   = "totp_configurations"
	tableWebauthnDevices      = "webauthn_devices"
	tableDuoDevices           = "duo_devices"
	tableAuthenticationLogs   = "authentication_logs"
	tableMigrations           = "migrations"
	tableEncryption           = "encryption"

	tableOAuth2AuthorizeCodeSessions = "oauth2_authorize_code_sessions"
	tableOAuth2AccessTokenSessions   = "oauth2_access_token_sessions"  //nolint:gosec // This is not a hardcoded credential.
	tableOAuth2RefreshTokenSessions  = "oauth2_refresh_token_sessions" //nolint:gosec // This is not a hardcoded credential.
	tableOAuth2PKCERequestSessions   = "oauth2_pkce_request_sessions"
	tableOAuth2OpenIDConnectSessions = "oauth2_openid_connect_sessions"
	tableOAuth2Subjects              = "oauth2_subjects"
	tableOAuth2BlacklistedJTI        = "oauth2_blacklisted_jti"

	tablePrefixBackup = "_bkp_"
)

// OAuth2SessionType represents the potential OAuth 2.0 session types.
type OAuth2SessionType string

// Representation of specific OAuth 2.0 session types.
const (
	OAuth2SessionTypeAuthorizeCode OAuth2SessionType = "authorize code"
	OAuth2SessionTypeAccessToken   OAuth2SessionType = "access token"
	OAuth2SessionTypeRefreshToken  OAuth2SessionType = "refresh token"
	OAuth2SessionTypePKCEChallenge OAuth2SessionType = "pkce challenge"
	OAuth2SessionTypeOpenIDConnect OAuth2SessionType = "openid connect"
)

const (
	encryptionNameCheck = "check"
)

// WARNING: Do not change/remove these consts. They are used for Pre1 migrations.
const (
	tablePre1TOTPSecrets                = "totp_secrets"
	tablePre1IdentityVerificationTokens = "identity_verification_tokens"
	tablePre1U2FDevices                 = "u2f_devices"

	tablePre1Config = "config"

	tableAlphaAuthenticationLogs         = "AuthenticationLogs"
	tableAlphaIdentityVerificationTokens = "IdentityVerificationTokens"
	tableAlphaPreferences                = "Preferences"
	tableAlphaPreferencesTableName       = "PreferencesTableName"
	tableAlphaSecondFactorPreferences    = "SecondFactorPreferences"
	tableAlphaTOTPSecrets                = "TOTPSecrets"
	tableAlphaU2FDeviceHandles           = "U2FDeviceHandles"
)

var tablesPre1 = []string{
	tablePre1TOTPSecrets,
	tablePre1IdentityVerificationTokens,
	tablePre1U2FDevices,

	tableUserPreferences,
	tableAuthenticationLogs,
}

const (
	providerAll      = "all"
	providerMySQL    = "mysql"
	providerPostgres = "postgres"
	providerSQLite   = "sqlite"
)

const (
	// This is the latest schema version for the purpose of tests.
	testLatestVersion = 4
)

const (
	// SchemaLatest represents the value expected for a "migrate to latest" migration. It's the maximum 32bit signed integer.
	SchemaLatest = 2147483647
)

type ctxKey int

const (
	ctxKeyTransaction ctxKey = iota
)

var (
	reMigration = regexp.MustCompile(`^V(\d{4})\.([^.]+)\.(all|sqlite|postgres|mysql)\.(up|down)\.sql$`)
)
