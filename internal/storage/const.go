package storage

import (
	"regexp"
)

const (
	tableAuthenticationLogs   = "authentication_logs"
	tableBannedUser           = "banned_user"
	tableBannedIP             = "banned_ip"
	tableCachedData           = "cached_data"
	tableUserMetadata         = "user_metadata"
	tableDuoDevices           = "duo_devices"
	tableIdentityVerification = "identity_verification"
	tableOneTimeCode          = "one_time_code"
	tableTOTPConfigurations   = "totp_configurations"
	tableTOTPHistory          = "totp_history"
	tableUserOpaqueIdentifier = "user_opaque_identifier"
	tableUserPreferences      = "user_preferences"
	tableWebAuthnCredentials  = "webauthn_credentials" //nolint:gosec // This is a table name, not a credential.
	tableWebAuthnUsers        = "webauthn_users"

	tableOAuth2BlacklistedJTI          = "oauth2_blacklisted_jti"
	tableOAuth2ConsentSession          = "oauth2_consent_session"
	tableOAuth2ConsentPreConfiguration = "oauth2_consent_preconfiguration"

	tableOAuth2AccessTokenSession   = "oauth2_access_token_session" //nolint:gosec // This is not a hardcoded credential.
	tableOAuth2AuthorizeCodeSession = "oauth2_authorization_code_session"
	tableOAuth2DeviceCodeSession    = "oauth2_device_code_session"
	tableOAuth2OpenIDConnectSession = "oauth2_openid_connect_session"
	tableOAuth2PARContext           = "oauth2_par_context"
	tableOAuth2PKCERequestSession   = "oauth2_pkce_request_session"
	tableOAuth2RefreshTokenSession  = "oauth2_refresh_token_session" //nolint:gosec // This is not a hardcoded credential.

	tableMigrations = "migrations"
	tableEncryption = "encryption"
)

const (
	encryptionNameCheck = "check"
)

// WARNING: Do not change/remove these consts. They are used for Pre1 migrations.
const (
	tablePre1TOTPSecrets                = "totp_secrets"
	tablePre1IdentityVerificationTokens = "identity_verification_tokens"
	tablePre1U2FDevices                 = "u2f_devices"
)

var tablesPre1 = []string{
	tablePre1TOTPSecrets,
	tablePre1IdentityVerificationTokens,
	tablePre1U2FDevices,

	tableUserPreferences,
	tableAuthenticationLogs,
}

const (
	pathMigrations   = "migrations"
	providerMySQL    = "mysql"
	providerPostgres = "postgres"
	providerSQLite   = "sqlite"
)

const (
	driverParameterFmtAppName = "authelia %s"
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
	reMigration                  = regexp.MustCompile(`^V(?P<Version>\d{4})\.(?P<Name>[^.]+)\.(?P<Direction>(up|down))\.sql$`)
	rePostgreSQLUnixDomainSocket = regexp.MustCompile(`^\.s\.PGSQL\.(\d+)$`)
)

const (
	na      = "N/A"
	invalid = "invalid"
)
