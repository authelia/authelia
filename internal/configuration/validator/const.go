package validator

import (
	"regexp"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

const (
	loopback           = "127.0.0.1"
	oauth2InstalledApp = "urn:ietf:wg:oauth:2.0:oob"
)

// Policy constants.
const (
	policyBypass    = "bypass"
	policyOneFactor = "one_factor"
	policyTwoFactor = "two_factor"
	policyDeny      = "deny"
)

const (
	digestSHA1   = "sha1"
	digestSHA224 = "sha224"
	digestSHA256 = "sha256"
	digestSHA384 = "sha384"
	digestSHA512 = "sha512"
)

// Hashing constants.
const (
	hashLegacyArgon2id = "argon2id"
	hashLegacySHA512   = digestSHA512

	hashArgon2    = "argon2"
	hashSHA2Crypt = "sha2crypt"
	hashPBKDF2    = "pbkdf2"
	hashSCrypt    = "scrypt"
	hashBCrypt    = "bcrypt"
)

// Scheme constants.
const (
	schemeLDAP  = "ldap"
	schemeLDAPS = "ldaps"
	schemeHTTP  = "http"
	schemeHTTPS = "https"
)

// Notifier Error constants.
const (
	errFmtNotifierMultipleConfigured = "notifier: please ensure only one of the 'smtp' or 'filesystem' notifier is configured"
	errFmtNotifierNotConfigured      = "notifier: you must ensure either the 'smtp' or 'filesystem' notifier " +
		"is configured"
	errFmtNotifierTemplatePathNotExist            = "notifier: option 'template_path' refers to location '%s' which does not exist"
	errFmtNotifierTemplatePathUnknownError        = "notifier: option 'template_path' refers to location '%s' which couldn't be opened: %w"
	errFmtNotifierFileSystemFileNameNotConfigured = "notifier: filesystem: option 'filename' is required"
	errFmtNotifierSMTPNotConfigured               = "notifier: smtp: option '%s' is required"
	errFmtNotifierSMTPTLSConfigInvalid            = "notifier: smtp: tls: %w"
	errFmtNotifierStartTlsDisabled                = "notifier: smtp: option 'disable_starttls' is enabled: " +
		"opportunistic STARTTLS is explicitly disabled which means all emails will be sent insecurely over plaintext " +
		"and this setting is only necessary for non-compliant SMTP servers which advertise they support STARTTLS " +
		"when they actually don't support STARTTLS"
)

const (
	errSuffixMustBeOneOf = "must be one of %s but it's configured as '%s'"
)

// Authentication Backend Error constants.
const (
	errFmtAuthBackendNotConfigured = "authentication_backend: you must ensure either the 'file' or 'ldap' " +
		"authentication backend is configured"
	errFmtAuthBackendMultipleConfigured = "authentication_backend: please ensure only one of the 'file' or 'ldap' " +
		"backend is configured"
	errFmtAuthBackendRefreshInterval = "authentication_backend: option 'refresh_interval' is configured to '%s' but " +
		"it must be either a duration notation or one of 'disable', or 'always': %w"
	errFmtAuthBackendPasswordResetCustomURLScheme = "authentication_backend: password_reset: option 'custom_url' is" +
		" configured to '%s' which has the scheme '%s' but the scheme must be either 'http' or 'https'"

	errFmtFileAuthBackendPathNotConfigured  = "authentication_backend: file: option 'path' is required"
	errFmtFileAuthBackendPasswordUnknownAlg = "authentication_backend: file: password: option 'algorithm' " +
		errSuffixMustBeOneOf
	errFmtFileAuthBackendPasswordInvalidVariant = "authentication_backend: file: password: %s: " +
		"option 'variant' " + errSuffixMustBeOneOf
	errFmtFileAuthBackendPasswordOptionTooLarge = "authentication_backend: file: password: %s: " +
		"option '%s' is configured as '%d' but must be less than or equal to '%d'"
	errFmtFileAuthBackendPasswordOptionTooSmall = "authentication_backend: file: password: %s: " +
		"option '%s' is configured as '%d' but must be greater than or equal to '%d'"
	errFmtFileAuthBackendPasswordArgon2MemoryTooLow = "authentication_backend: file: password: argon2: " +
		"option 'memory' is configured as '%d' but must be greater than or equal to '%d' or '%d' (the value of 'parallelism) multiplied by '%d'"

	errFmtLDAPAuthBackendUnauthenticatedBindWithPassword     = "authentication_backend: ldap: option 'permit_unauthenticated_bind' can't be enabled when a password is specified"
	errFmtLDAPAuthBackendUnauthenticatedBindWithResetEnabled = "authentication_backend: ldap: option 'permit_unauthenticated_bind' can't be enabled when password reset is enabled"

	errFmtLDAPAuthBackendMissingOption    = "authentication_backend: ldap: option '%s' is required"
	errFmtLDAPAuthBackendTLSConfigInvalid = "authentication_backend: ldap: tls: %w"
	errFmtLDAPAuthBackendImplementation   = "authentication_backend: ldap: option 'implementation' " +
		errSuffixMustBeOneOf
	errFmtLDAPAuthBackendFilterReplacedPlaceholders = "authentication_backend: ldap: option " +
		"'%s' has an invalid placeholder: '%s' has been removed, please use '%s' instead"
	errFmtLDAPAuthBackendURLNotParsable = "authentication_backend: ldap: option " +
		"'url' could not be parsed: %w"
	errFmtLDAPAuthBackendURLInvalidScheme = "authentication_backend: ldap: option " +
		"'url' must have either the 'ldap' or 'ldaps' scheme but it's configured as '%s'"
	errFmtLDAPAuthBackendFilterEnclosingParenthesis = "authentication_backend: ldap: option " +
		"'%s' must contain enclosing parenthesis: '%s' should probably be '(%s)'"
	errFmtLDAPAuthBackendFilterMissingPlaceholder = "authentication_backend: ldap: option " +
		"'%s' must contain the placeholder '{%s}' but it's absent"
)

// TOTP Error constants.
const (
	errFmtTOTPInvalidAlgorithm  = "totp: option 'algorithm' must be one of %s but it's configured as '%s'"
	errFmtTOTPInvalidPeriod     = "totp: option 'period' option must be 15 or more but it's configured as '%d'"
	errFmtTOTPInvalidDigits     = "totp: option 'digits' must be 6 or 8 but it's configured as '%d'"
	errFmtTOTPInvalidSecretSize = "totp: option 'secret_size' must be %d or higher but it's configured as '%d'" //nolint:gosec
)

// Storage Error constants.
const (
	errStrStorage                                 = "storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided"
	errStrStorageEncryptionKeyMustBeProvided      = "storage: option 'encryption_key' is required"
	errStrStorageEncryptionKeyTooShort            = "storage: option 'encryption_key' must be 20 characters or longer"
	errFmtStorageUserPassMustBeProvided           = "storage: %s: option 'username' and 'password' are required" //nolint:gosec
	errFmtStorageOptionMustBeProvided             = "storage: %s: option '%s' is required"
	errFmtStorageTLSConfigInvalid                 = "storage: %s: tls: %w"
	errFmtStoragePostgreSQLInvalidSSLMode         = "storage: postgres: ssl: option 'mode' must be one of %s but it's configured as '%s'"
	errFmtStoragePostgreSQLInvalidSSLAndTLSConfig = "storage: postgres: can't define both 'tls' and 'ssl' configuration options"
	warnFmtStoragePostgreSQLInvalidSSLDeprecated  = "storage: postgres: ssl: the ssl configuration options are deprecated and we recommend the tls options instead"
)

// Telemetry Error constants.
const (
	errFmtTelemetryMetricsScheme = "telemetry: metrics: option 'address' must have a scheme 'tcp://' but it's configured as '%s'"
)

// OpenID Error constants.
const (
	errFmtOIDCNoClientsConfigured = "identity_providers: oidc: option 'clients' must have one or " +
		"more clients configured"
	errFmtOIDCNoPrivateKey                               = "identity_providers: oidc: option 'issuer_private_key' or `issuer_jwks` is required"
	errFmtOIDCInvalidPrivateKeyBitSize                   = "identity_providers: oidc: option 'issuer_private_key' must be an RSA private key with %d bits or more but it only has %d bits"
	errFmtOIDCInvalidPrivateKeyMalformedMissingPublicKey = "identity_providers: oidc: option 'issuer_private_key' must be a valid RSA private key but the provided data is missing the public key bits"
	errFmtOIDCCertificateMismatch                        = "identity_providers: oidc: option 'issuer_private_key' does not appear to be the private key the certificate provided by option 'issuer_certificate_chain'"
	errFmtOIDCCertificateChain                           = "identity_providers: oidc: option 'issuer_certificate_chain' produced an error during validation of the chain: %w"
	errFmtOIDCEnforcePKCEInvalidValue                    = "identity_providers: oidc: option 'enforce_pkce' must be 'never', " +
		"'public_clients_only' or 'always', but it's configured as '%s'"

	errFmtOIDCCORSInvalidOrigin                    = "identity_providers: oidc: cors: option 'allowed_origins' contains an invalid value '%s' as it has a %s: origins must only be scheme, hostname, and an optional port"
	errFmtOIDCCORSInvalidOriginWildcard            = "identity_providers: oidc: cors: option 'allowed_origins' contains the wildcard origin '*' with more than one origin but the wildcard origin must be defined by itself"
	errFmtOIDCCORSInvalidOriginWildcardWithClients = "identity_providers: oidc: cors: option 'allowed_origins' contains the wildcard origin '*' cannot be specified with option 'allowed_origins_from_client_redirect_uris' enabled"
	errFmtOIDCCORSInvalidEndpoint                  = "identity_providers: oidc: cors: option 'endpoints' contains an invalid value '%s': must be one of %s"

	errFmtOIDCClientsDuplicateID = "identity_providers: oidc: clients: option 'id' must be unique for every client but one or more clients share the following 'id' values %s"
	errFmtOIDCClientsWithEmptyID = "identity_providers: oidc: clients: option 'id' is required but was absent on the clients in positions %s"
	errFmtOIDCClientsDeprecated  = "identity_providers: oidc: clients: warnings for clients above indicate deprecated functionality and it's strongly suggested these issues are checked and fixed if they're legitimate issues or reported if they are not as in a future version these warnings will become errors"

	errFmtOIDCClientInvalidSecret             = "identity_providers: oidc: client '%s': option 'secret' is required"
	errFmtOIDCClientInvalidSecretPlainText    = "identity_providers: oidc: client '%s': option 'secret' is plaintext but for clients not using the 'token_endpoint_auth_method' of 'client_secret_jwt' it should be a hashed value as plaintext values are deprecated with the exception of 'client_secret_jwt' and will be removed when oidc becomes stable"
	errFmtOIDCClientInvalidSecretNotPlainText = "identity_providers: oidc: client '%s': option 'secret' must be plaintext with option 'token_endpoint_auth_method' with a value of 'client_secret_jwt'"
	errFmtOIDCClientPublicInvalidSecret       = "identity_providers: oidc: client '%s': option 'secret' is " +
		"required to be empty when option 'public' is true"
	errFmtOIDCClientRedirectURICantBeParsed = "identity_providers: oidc: client '%s': option 'redirect_uris' has an " +
		"invalid value: redirect uri '%s' could not be parsed: %v"
	errFmtOIDCClientRedirectURIPublic = "identity_providers: oidc: client '%s': option 'redirect_uris' has the " +
		"redirect uri '%s' when option 'public' is false but this is invalid as this uri is not valid " +
		"for the openid connect confidential client type"
	errFmtOIDCClientRedirectURIAbsolute = "identity_providers: oidc: client '%s': option 'redirect_uris' has an " +
		"invalid value: redirect uri '%s' must have a scheme but it's absent"
	errFmtOIDCClientInvalidConsentMode = "identity_providers: oidc: client '%s': consent: option 'mode' must be one of " +
		"%s but it's configured as '%s'"
	errFmtOIDCClientInvalidEntries = "identity_providers: oidc: client '%s': option '%s' must only have the values " +
		"%s but the values %s are present"
	errFmtOIDCClientInvalidEntryDuplicates = "identity_providers: oidc: client '%s': option '%s' must have unique values but the values %s are duplicated"
	errFmtOIDCClientInvalidValue           = "identity_providers: oidc: client '%s': option " +
		"'%s' must be one of %s but it's configured as '%s'"
	errFmtOIDCClientInvalidTokenEndpointAuthMethod = "identity_providers: oidc: client '%s': option " +
		"'token_endpoint_auth_method' must be one of %s when configured as the confidential client type unless it only includes implicit flow response types such as %s but it's configured as '%s'"
	errFmtOIDCClientInvalidTokenEndpointAuthMethodPublic = "identity_providers: oidc: client '%s': option " +
		"'token_endpoint_auth_method' must be 'none' when configured as the public client type but it's configured as '%s'"
	errFmtOIDCClientInvalidTokenEndpointAuthSigAlg = "identity_providers: oidc: client '%s': option " +
		"'token_endpoint_auth_signing_alg' must be %s when option 'token_endpoint_auth_method' is %s"
	errFmtOIDCClientInvalidSectorIdentifier = "identity_providers: oidc: client '%s': option " +
		"'sector_identifier' with value '%s': must be a URL with only the host component for example '%s' but it has a %s with the value '%s'"
	errFmtOIDCClientInvalidSectorIdentifierWithoutValue = "identity_providers: oidc: client '%s': option " +
		"'sector_identifier' with value '%s': must be a URL with only the host component for example '%s' but it has a %s"
	errFmtOIDCClientInvalidSectorIdentifierHost = "identity_providers: oidc: client '%s': option " +
		"'sector_identifier' with value '%s': must be a URL with only the host component but appears to be invalid"
	errFmtOIDCClientInvalidGrantTypeMatch = "identity_providers: oidc: client '%s': option " +
		"'grant_types' should only have grant type values which are valid with the configured 'response_types' for the client but '%s' expects a response type %s such as %s but the response types are %s"
	errFmtOIDCClientInvalidGrantTypeRefresh = "identity_providers: oidc: client '%s': option " +
		"'grant_types' should only have the 'refresh_token' value if the client is also configured with the 'offline_access' scope"
	errFmtOIDCClientInvalidRefreshTokenOptionWithoutCodeResponseType = "identity_providers: oidc: client '%s': option " +
		"'%s' should only have the values %s if the client is also configured with a 'response_type' such as %s which respond with authorization codes"
	errFmtOIDCServerInsecureParameterEntropy = "openid connect provider: SECURITY ISSUE - minimum parameter entropy is " +
		"configured to an unsafe value, it should be above 8 but it's configured to %d"
)

// WebAuthn Error constants.
const (
	errFmtWebAuthnConveyancePreference = "webauthn: option 'attestation_conveyance_preference' must be one of %s but it's configured as '%s'"
	errFmtWebAuthnUserVerification     = "webauthn: option 'user_verification' must be one of %s but it's configured as '%s'"
)

// Access Control error constants.
const (
	errFmtAccessControlDefaultPolicyValue = "access control: option 'default_policy' must be one of %s but it's " +
		"configured as '%s'"
	errFmtAccessControlDefaultPolicyWithoutRules = "access control: 'default_policy' option '%s' is invalid: when " +
		"no rules are specified it must be 'two_factor' or 'one_factor'"
	errFmtAccessControlNetworkGroupIPCIDRInvalid = "access control: networks: network group '%s' is invalid: the " +
		"network '%s' is not a valid IP or CIDR notation"
	errFmtAccessControlWarnNoRulesDefaultPolicy = "access control: no rules have been specified so the " +
		"'default_policy' of '%s' is going to be applied to all requests"
	errFmtAccessControlRuleNoDomains                    = "access control: rule %s: option 'domain' or 'domain_regex' must be present but are both absent"
	errFmtAccessControlRuleNoPolicy                     = "access control: rule %s: option 'policy' must be present but it's absent"
	errFmtAccessControlRuleInvalidPolicy                = "access control: rule %s: option 'policy' must be one of %s but it's configured as '%s'"
	errAccessControlRuleBypassPolicyInvalidWithSubjects = "access control: rule %s: 'policy' option 'bypass' is " +
		"not supported when 'subject' option is configured: see " +
		"https://www.authelia.com/c/acl#bypass"
	errAccessControlRuleBypassPolicyInvalidWithSubjectsWithGroupDomainRegex = "access control: rule %s: 'policy' option 'bypass' is " +
		"not supported when 'domain_regex' option contains the user or group named matches. For more information see: " +
		"https://www.authelia.com/c/acl-match-concept-2"
	errFmtAccessControlRuleNetworksInvalid = "access control: rule %s: the network '%s' is not a " +
		"valid Group Name, IP, or CIDR notation"
	errFmtAccessControlRuleSubjectInvalid = "access control: rule %s: 'subject' option '%s' is " +
		"invalid: must start with 'user:' or 'group:'"
	errFmtAccessControlRuleInvalidEntries              = "access control: rule %s: option '%s' must only have the values %s but the values %s are present"
	errFmtAccessControlRuleInvalidDuplicates           = "access control: rule %s: option '%s' must have unique values but the values %s are duplicated"
	errFmtAccessControlRuleQueryInvalid                = "access control: rule %s: query: option 'operator' must be one of %s but it's configured as '%s'"
	errFmtAccessControlRuleQueryInvalidNoValue         = "access control: rule %s: query: option '%s' is required but it's absent"
	errFmtAccessControlRuleQueryInvalidNoValueOperator = "access control: rule %s: query: option '%s' must be present when the option 'operator' is '%s' but it's absent"
	errFmtAccessControlRuleQueryInvalidValue           = "access control: rule %s: query: option '%s' must not be present when the option 'operator' is '%s' but it's present"
	errFmtAccessControlRuleQueryInvalidValueParse      = "access control: rule %s: query: option '%s' is " +
		"invalid: %w"
	errFmtAccessControlRuleQueryInvalidValueType = "access control: rule %s: query: option 'value' is " +
		"invalid: expected type was string but got %T"
)

// Theme Error constants.
const (
	errFmtThemeName = "option 'theme' must be one of %s but it's configured as '%s'"
)

// NTP Error constants.
const (
	errFmtNTPVersion = "ntp: option 'version' must be either 3 or 4 but it's configured as '%d'"
)

// Session error constants.
const (
	errFmtSessionOptionRequired           = "session: option '%s' is required"
	errFmtSessionLegacyAndWarning         = "session: option 'domain' and option 'cookies' can't be specified at the same time"
	errFmtSessionSameSite                 = "session: option 'same_site' must be one of %s but it's configured as '%s'"
	errFmtSessionSecretRequired           = "session: option 'secret' is required when using the '%s' provider"
	errFmtSessionRedisPortRange           = "session: redis: option 'port' must be between 1 and 65535 but it's configured as '%d'"
	errFmtSessionRedisHostRequired        = "session: redis: option 'host' is required"
	errFmtSessionRedisHostOrNodesRequired = "session: redis: option 'host' or the 'high_availability' option 'nodes' is required"
	errFmtSessionRedisTLSConfigInvalid    = "session: redis: tls: %w"

	errFmtSessionRedisSentinelMissingName     = "session: redis: high_availability: option 'sentinel_name' is required"
	errFmtSessionRedisSentinelNodeHostMissing = "session: redis: high_availability: option 'nodes': option 'host' is required for each node but one or more nodes are missing this"

	errFmtSessionDomainMustBeRoot                = "session: domain config %s: option 'domain' must be the domain you wish to protect not a wildcard domain but it's configured as '%s'"
	errFmtSessionDomainSameSite                  = "session: domain config %s: option 'same_site' must be one of %s but it's configured as '%s'"
	errFmtSessionDomainRequired                  = "session: domain config %s: option 'domain' is required"
	errFmtSessionDomainHasPeriodPrefix           = "session: domain config %s: option 'domain' has a prefix of '.' which is not supported or intended behaviour: you can use this at your own risk but we recommend removing it"
	errFmtSessionDomainDuplicate                 = "session: domain config %s: option 'domain' is a duplicate value for another configured session domain"
	errFmtSessionDomainDuplicateCookieScope      = "session: domain config %s: option 'domain' shares the same cookie domain scope as another configured session domain"
	errFmtSessionDomainPortalURLInsecure         = "session: domain config %s: option 'authelia_url' does not have a secure scheme with a value of '%s'"
	errFmtSessionDomainPortalURLNotInCookieScope = "session: domain config %s: option 'authelia_url' does not share a cookie scope with domain '%s' with a value of '%s'"
	errFmtSessionDomainInvalidDomain             = "session: domain config %s: option 'domain' is not a valid cookie domain"
	errFmtSessionDomainInvalidDomainNoDots       = "session: domain config %s: option 'domain' is not a valid cookie domain: must have at least a single period"
	errFmtSessionDomainInvalidDomainPublic       = "session: domain config %s: option 'domain' is not a valid cookie domain: the domain is part of the special public suffix list"
)

// Regulation Error Consts.
const (
	errFmtRegulationFindTimeGreaterThanBanTime = "regulation: option 'find_time' must be less than or equal to option 'ban_time'"
)

// Server Error constants.
const (
	errFmtServerTLSCert                           = "server: tls: option 'key' must also be accompanied by option 'certificate'"
	errFmtServerTLSKey                            = "server: tls: option 'certificate' must also be accompanied by option 'key'"
	errFmtServerTLSCertFileDoesNotExist           = "server: tls: file path %s provided in 'certificate' does not exist"
	errFmtServerTLSKeyFileDoesNotExist            = "server: tls: file path %s provided in 'key' does not exist"
	errFmtServerTLSClientAuthCertFileDoesNotExist = "server: tls: client_certificates: certificates: file path %s does not exist"
	errFmtServerTLSClientAuthNoAuth               = "server: tls: client authentication cannot be configured if no server certificate and key are provided"

	errFmtServerPathNoForwardSlashes = "server: option 'path' must not contain any forward slashes"
	errFmtServerPathAlphaNum         = "server: option 'path' must only contain alpha numeric characters"

	errFmtServerEndpointsAuthzImplementation    = "server: endpoints: authz: %s: option 'implementation' must be one of %s but it's configured as '%s'"
	errFmtServerEndpointsAuthzStrategy          = "server: endpoints: authz: %s: authn_strategies: option 'name' must be one of %s but it's configured as '%s'"
	errFmtServerEndpointsAuthzStrategyDuplicate = "server: endpoints: authz: %s: authn_strategies: duplicate strategy name detected with name '%s'"
	errFmtServerEndpointsAuthzPrefixDuplicate   = "server: endpoints: authz: %s: endpoint starts with the same prefix as the '%s' endpoint with the '%s' implementation which accepts prefixes as part of its implementation"
	errFmtServerEndpointsAuthzInvalidName       = "server: endpoints: authz: %s: contains invalid characters"

	errFmtServerEndpointsAuthzLegacyInvalidImplementation = "server: endpoints: authz: %s: option 'implementation' is invalid: the endpoint with the name 'legacy' must use the 'Legacy' implementation"
)

const (
	errPasswordPolicyMultipleDefined                        = "password_policy: only a single password policy mechanism can be specified"
	errFmtPasswordPolicyStandardMinLengthNotGreaterThanZero = "password_policy: standard: option 'min_length' must be greater than 0 but it's configured as %d"
	errFmtPasswordPolicyZXCVBNMinScoreInvalid               = "password_policy: zxcvbn: option 'min_score' is invalid: must be between 1 and 4 but it's configured as %d"
)

const (
	errPrivacyPolicyEnabledWithoutURL = "privacy_policy: option 'policy_url' must be provided when the option 'enabled' is true"
	errFmtPrivacyPolicyURLNotHTTPS    = "privacy_policy: option 'policy_url' must have the 'https' scheme but it's configured as '%s'"
)

const (
	errFmtDuoMissingOption = "duo_api: option '%s' is required when duo is enabled but it's absent"
)

// Error constants.
const (
	errFmtInvalidDefault2FAMethod         = "option 'default_2fa_method' must be one of %s but it's configured as '%s'"
	errFmtInvalidDefault2FAMethodDisabled = "option 'default_2fa_method' must be one of the enabled options %s but it's configured as '%s'"

	errFmtReplacedConfigurationKey = "invalid configuration key '%s' was replaced by '%s'"

	errFmtLoggingLevelInvalid = "log: option 'level' must be one of %s but it's configured as '%s'"

	errFileHashing  = "config key incorrect: authentication_backend.file.hashing should be authentication_backend.file.password"
	errFilePHashing = "config key incorrect: authentication_backend.file.password_hashing should be authentication_backend.file.password"
	errFilePOptions = "config key incorrect: authentication_backend.file.password_options should be authentication_backend.file.password"
)

const (
	operatorPresent    = "present"
	operatorAbsent     = "absent"
	operatorEqual      = "equal"
	operatorNotEqual   = "not equal"
	operatorPattern    = "pattern"
	operatorNotPattern = "not pattern"
)

var (
	validLDAPImplementations = []string{
		schema.LDAPImplementationCustom,
		schema.LDAPImplementationActiveDirectory,
		schema.LDAPImplementationRFC2307bis,
		schema.LDAPImplementationFreeIPA,
		schema.LDAPImplementationLLDAP,
		schema.LDAPImplementationGLAuth,
	}
)

const (
	legacy                      = "legacy"
	authzImplementationLegacy   = "Legacy"
	authzImplementationExtAuthz = "ExtAuthz"
)

const (
	auto = "auto"
)

var (
	validAuthzImplementations = []string{"AuthRequest", "ForwardAuth", authzImplementationExtAuthz, authzImplementationLegacy}
	validAuthzAuthnStrategies = []string{"CookieSession", "HeaderAuthorization", "HeaderProxyAuthorization", "HeaderAuthRequestProxyAuthorization", "HeaderLegacy"}
)

var (
	validArgon2Variants    = []string{"argon2id", "id", "argon2i", "i", "argon2d", "d"}
	validSHA2CryptVariants = []string{digestSHA256, digestSHA512}
	validPBKDF2Variants    = []string{digestSHA1, digestSHA224, digestSHA256, digestSHA384, digestSHA512}
	validBCryptVariants    = []string{"standard", digestSHA256}
	validHashAlgorithms    = []string{hashSHA2Crypt, hashPBKDF2, hashSCrypt, hashBCrypt, hashArgon2}
)

var (
	validStoragePostgreSQLSSLModes           = []string{"disable", "require", "verify-ca", "verify-full"}
	validThemeNames                          = []string{"light", "dark", "grey", auto}
	validSessionSameSiteValues               = []string{"none", "lax", "strict"}
	validLogLevels                           = []string{"trace", "debug", "info", "warn", "error"}
	validWebAuthnConveyancePreferences       = []string{string(protocol.PreferNoAttestation), string(protocol.PreferIndirectAttestation), string(protocol.PreferDirectAttestation)}
	validWebAuthnUserVerificationRequirement = []string{string(protocol.VerificationDiscouraged), string(protocol.VerificationPreferred), string(protocol.VerificationRequired)}
	validRFC7231HTTPMethodVerbs              = []string{fasthttp.MethodGet, fasthttp.MethodHead, fasthttp.MethodPost, fasthttp.MethodPut, fasthttp.MethodPatch, fasthttp.MethodDelete, fasthttp.MethodTrace, fasthttp.MethodConnect, fasthttp.MethodOptions}
	validRFC4918HTTPMethodVerbs              = []string{"COPY", "LOCK", "MKCOL", "MOVE", "PROPFIND", "PROPPATCH", "UNLOCK"}
)

var (
	validACLHTTPMethodVerbs = append(validRFC7231HTTPMethodVerbs, validRFC4918HTTPMethodVerbs...)
	validACLRulePolicies    = []string{policyBypass, policyOneFactor, policyTwoFactor, policyDeny}
	validACLRuleOperators   = []string{operatorPresent, operatorAbsent, operatorEqual, operatorNotEqual, operatorPattern, operatorNotPattern}
)

var validDefault2FAMethods = []string{"totp", "webauthn", "mobile_push"}

const (
	attrOIDCScopes              = "scopes"
	attrOIDCResponseTypes       = "response_types"
	attrOIDCResponseModes       = "response_modes"
	attrOIDCGrantTypes          = "grant_types"
	attrOIDCRedirectURIs        = "redirect_uris"
	attrOIDCTokenAuthMethod     = "token_endpoint_auth_method"
	attrOIDCUsrSigAlg           = "userinfo_signing_algorithm"
	attrOIDCIDTokenSigAlg       = "id_token_signing_alg"
	attrOIDCPKCEChallengeMethod = "pkce_challenge_method"
)

var (
	validOIDCCORSEndpoints = []string{oidc.EndpointAuthorization, oidc.EndpointPushedAuthorizationRequest, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo}

	validOIDCClientScopes                    = []string{oidc.ScopeOpenID, oidc.ScopeEmail, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeOfflineAccess}
	validOIDCClientConsentModes              = []string{auto, oidc.ClientConsentModeImplicit.String(), oidc.ClientConsentModeExplicit.String(), oidc.ClientConsentModePreConfigured.String()}
	validOIDCClientResponseModes             = []string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery, oidc.ResponseModeFragment}
	validOIDCClientResponseTypes             = []string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowIDToken, oidc.ResponseTypeImplicitFlowToken, oidc.ResponseTypeImplicitFlowBoth, oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeHybridFlowBoth}
	validOIDCClientResponseTypesImplicitFlow = []string{oidc.ResponseTypeImplicitFlowIDToken, oidc.ResponseTypeImplicitFlowToken, oidc.ResponseTypeImplicitFlowBoth}
	validOIDCClientResponseTypesHybridFlow   = []string{oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeHybridFlowBoth}
	validOIDCClientResponseTypesRefreshToken = []string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeHybridFlowBoth}
	validOIDCClientGrantTypes                = []string{oidc.GrantTypeImplicit, oidc.GrantTypeRefreshToken, oidc.GrantTypeAuthorizationCode}

	validOIDCClientTokenEndpointAuthMethods             = []string{oidc.ClientAuthMethodNone, oidc.ClientAuthMethodClientSecretPost, oidc.ClientAuthMethodClientSecretBasic, oidc.ClientAuthMethodClientSecretJWT}
	validOIDCClientTokenEndpointAuthMethodsConfidential = []string{oidc.ClientAuthMethodClientSecretPost, oidc.ClientAuthMethodClientSecretBasic}
	validOIDCClientTokenEndpointAuthSigAlgs             = []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512}
	validOIDCIssuerJWKSigningAlgs                       = []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgRSAPSSUsingSHA256, oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAPSSUsingSHA384, oidc.SigningAlgECDSAUsingP384AndSHA384, oidc.SigningAlgRSAUsingSHA512, oidc.SigningAlgRSAPSSUsingSHA512, oidc.SigningAlgECDSAUsingP521AndSHA512}
)

var (
	reKeyReplacer       = regexp.MustCompile(`\[\d+]`)
	reDomainCharacters  = regexp.MustCompile(`^[a-z0-9-]+(\.[a-z0-9-]+)+[a-z0-9]$`)
	reAuthzEndpointName = regexp.MustCompile(`^[a-zA-Z](([a-zA-Z0-9/._-]*)([a-zA-Z]))?$`)
)

var replacedKeys = map[string]string{
	"authentication_backend.ldap.skip_verify":         "authentication_backend.ldap.tls.skip_verify",
	"authentication_backend.ldap.minimum_tls_version": "authentication_backend.ldap.tls.minimum_version",
	"notifier.smtp.disable_verify_cert":               "notifier.smtp.tls.skip_verify",
	"logs_level":                                      "log.level",
	"logs_file_path":                                  "log.file_path",
	"log_level":                                       "log.level",
	"log_file_path":                                   "log.file_path",
	"log_format":                                      "log.format",
	"host":                                            "server.host",
	"port":                                            "server.port",
	"tls_key":                                         "server.tls.key",
	"tls_cert":                                        "server.tls.certificate",
}

var specificErrorKeys = map[string]string{
	"google_analytics": "config key removed: google_analytics - this functionality has been deprecated",
	"notifier.smtp.trusted_cert": "invalid configuration key 'notifier.smtp.trusted_cert' it has been removed, " +
		"option has been replaced by the global option 'certificates_directory'",

	"authentication_backend.file.password_options.algorithm":   errFilePOptions,
	"authentication_backend.file.password_options.iterations":  errFilePOptions,
	"authentication_backend.file.password_options.key_length":  errFilePOptions,
	"authentication_backend.file.password_options.salt_length": errFilePOptions,
	"authentication_backend.file.password_options.memory":      errFilePOptions,
	"authentication_backend.file.password_options.parallelism": errFilePOptions,
	"authentication_backend.file.password_hashing.algorithm":   errFilePHashing,
	"authentication_backend.file.password_hashing.iterations":  errFilePHashing,
	"authentication_backend.file.password_hashing.key_length":  errFilePHashing,
	"authentication_backend.file.password_hashing.salt_length": errFilePHashing,
	"authentication_backend.file.password_hashing.memory":      errFilePHashing,
	"authentication_backend.file.password_hashing.parallelism": errFilePHashing,
	"authentication_backend.file.hashing.algorithm":            errFileHashing,
	"authentication_backend.file.hashing.iterations":           errFileHashing,
	"authentication_backend.file.hashing.key_length":           errFileHashing,
	"authentication_backend.file.hashing.salt_length":          errFileHashing,
	"authentication_backend.file.hashing.memory":               errFileHashing,
	"authentication_backend.file.hashing.parallelism":          errFileHashing,
}
