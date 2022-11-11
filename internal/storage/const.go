package storage

import (
	"regexp"
)

const (
	tableAuthenticationLogs   = "authentication_logs"
	tableDuoDevices           = "duo_devices"
	tableIdentityVerification = "identity_verification"
	tableTOTPConfigurations   = "totp_configurations"
	tableUserOpaqueIdentifier = "user_opaque_identifier"
	tableUserPreferences      = "user_preferences"
	tableWebauthnDevices      = "webauthn_devices"

	tableOAuth2ConsentSession          = "oauth2_consent_session"
	tableOAuth2ConsentPreConfiguration = "oauth2_consent_preconfiguration"

	tableOAuth2AuthorizeCodeSession = "oauth2_authorization_code_session"
	tableOAuth2AccessTokenSession   = "oauth2_access_token_session"  //nolint:gosec // This is not a hardcoded credential.
	tableOAuth2RefreshTokenSession  = "oauth2_refresh_token_session" //nolint:gosec // This is not a hardcoded credential.
	tableOAuth2PKCERequestSession   = "oauth2_pkce_request_session"
	tableOAuth2OpenIDConnectSession = "oauth2_openid_connect_session"
	tableOAuth2BlacklistedJTI       = "oauth2_blacklisted_jti"

	tableMigrations = "migrations"
	tableEncryption = "encryption"
)

// OAuth2SessionType represents the potential OAuth 2.0 session types.
type OAuth2SessionType string

// Representation of specific OAuth 2.0 session types.
const (
	OAuth2SessionTypeAuthorizeCode OAuth2SessionType = "authorization code"
	OAuth2SessionTypeAccessToken   OAuth2SessionType = "access token"
	OAuth2SessionTypeRefreshToken  OAuth2SessionType = "refresh token"
	OAuth2SessionTypePKCEChallenge OAuth2SessionType = "pkce challenge"
	OAuth2SessionTypeOpenIDConnect OAuth2SessionType = "openid connect"
)

const (
	sqlNetworkTypeTCP        = "tcp"
	sqlNetworkTypeUnixSocket = "unix"
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
	providerAll      = "all"
	providerMySQL    = "mysql"
	providerPostgres = "postgres"
	providerSQLite   = "sqlite"
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
