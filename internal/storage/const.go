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
	tableWebauthnUsers        = "webauthn_users"

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
type OAuth2SessionType int

// Representation of specific OAuth 2.0 session types.
const (
	OAuth2SessionTypeAuthorizeCode OAuth2SessionType = iota
	OAuth2SessionTypeAccessToken
	OAuth2SessionTypeRefreshToken
	OAuth2SessionTypePKCEChallenge
	OAuth2SessionTypeOpenIDConnect
)

// String returns a string representation of this OAuth2SessionType.
func (s OAuth2SessionType) String() string {
	switch s {
	case OAuth2SessionTypeAuthorizeCode:
		return "authorization code"
	case OAuth2SessionTypeAccessToken:
		return "access token"
	case OAuth2SessionTypeRefreshToken:
		return "refresh token"
	case OAuth2SessionTypePKCEChallenge:
		return "pkce challenge"
	case OAuth2SessionTypeOpenIDConnect:
		return "openid connect"
	default:
		return "invalid"
	}
}

// Table returns the table name for this session type.
func (s OAuth2SessionType) Table() string {
	switch s {
	case OAuth2SessionTypeAuthorizeCode:
		return tableOAuth2AuthorizeCodeSession
	case OAuth2SessionTypeAccessToken:
		return tableOAuth2AccessTokenSession
	case OAuth2SessionTypeRefreshToken:
		return tableOAuth2RefreshTokenSession
	case OAuth2SessionTypePKCEChallenge:
		return tableOAuth2PKCERequestSession
	case OAuth2SessionTypeOpenIDConnect:
		return tableOAuth2OpenIDConnectSession
	default:
		return ""
	}
}

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

const (
	na      = "N/A"
	invalid = "invalid"
)
