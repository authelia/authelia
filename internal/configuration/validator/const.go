package validator

import (
	"regexp"

	"github.com/go-webauthn/webauthn/protocol"

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
	errSuffixMustBeOneOf = "is configured as '%s' but must be one of the following values: '%s'"
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
		"'url' must have either the 'ldap' or 'ldaps' scheme but it is configured as '%s'"
	errFmtLDAPAuthBackendFilterEnclosingParenthesis = "authentication_backend: ldap: option " +
		"'%s' must contain enclosing parenthesis: '%s' should probably be '(%s)'"
	errFmtLDAPAuthBackendFilterMissingPlaceholder = "authentication_backend: ldap: option " +
		"'%s' must contain the placeholder '{%s}' but it is required"
)

// TOTP Error constants.
const (
	errFmtTOTPInvalidAlgorithm  = "totp: option 'algorithm' must be one of '%s' but it is configured as '%s'"
	errFmtTOTPInvalidPeriod     = "totp: option 'period' option must be 15 or more but it is configured as '%d'"
	errFmtTOTPInvalidDigits     = "totp: option 'digits' must be 6 or 8 but it is configured as '%d'"
	errFmtTOTPInvalidSecretSize = "totp: option 'secret_size' must be %d or higher but it is configured as '%d'" //nolint:gosec
)

// Storage Error constants.
const (
	errStrStorage                                 = "storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided"
	errStrStorageEncryptionKeyMustBeProvided      = "storage: option 'encryption_key' is required"
	errStrStorageEncryptionKeyTooShort            = "storage: option 'encryption_key' must be 20 characters or longer"
	errFmtStorageUserPassMustBeProvided           = "storage: %s: option 'username' and 'password' are required" //nolint:gosec
	errFmtStorageOptionMustBeProvided             = "storage: %s: option '%s' is required"
	errFmtStorageTLSConfigInvalid                 = "storage: %s: tls: %w"
	errFmtStoragePostgreSQLInvalidSSLMode         = "storage: postgres: ssl: option 'mode' must be one of '%s' but it is configured as '%s'"
	errFmtStoragePostgreSQLInvalidSSLAndTLSConfig = "storage: postgres: can't define both 'tls' and 'ssl' configuration options"
	warnFmtStoragePostgreSQLInvalidSSLDeprecated  = "storage: postgres: ssl: the ssl configuration options are deprecated and we recommend the tls options instead"
)

// Telemetry Error constants.
const (
	errFmtTelemetryMetricsScheme = "telemetry: metrics: option 'address' must have a scheme 'tcp://' but it is configured as '%s'"
)

// OpenID Error constants.
const (
	errFmtOIDCNoClientsConfigured = "identity_providers: oidc: option 'clients' must have one or " +
		"more clients configured"
	errFmtOIDCNoPrivateKey             = "identity_providers: oidc: option 'issuer_private_key' is required"
	errFmtOIDCInvalidPrivateKeyBitSize = "identity_providers: oidc: option 'issuer_private_key' must be an RSA private key with %d bits or more but it only has %d bits"
	errFmtOIDCCertificateMismatch      = "identity_providers: oidc: option 'issuer_private_key' does not appear to be the private key the certificate provided by option 'issuer_certificate_chain'"
	errFmtOIDCCertificateChain         = "identity_providers: oidc: option 'issuer_certificate_chain' produced an error during validation of the chain: %w"
	errFmtOIDCEnforcePKCEInvalidValue  = "identity_providers: oidc: option 'enforce_pkce' must be 'never', " +
		"'public_clients_only' or 'always', but it is configured as '%s'"

	errFmtOIDCCORSInvalidOrigin                    = "identity_providers: oidc: cors: option 'allowed_origins' contains an invalid value '%s' as it has a %s: origins must only be scheme, hostname, and an optional port"
	errFmtOIDCCORSInvalidOriginWildcard            = "identity_providers: oidc: cors: option 'allowed_origins' contains the wildcard origin '*' with more than one origin but the wildcard origin must be defined by itself"
	errFmtOIDCCORSInvalidOriginWildcardWithClients = "identity_providers: oidc: cors: option 'allowed_origins' contains the wildcard origin '*' cannot be specified with option 'allowed_origins_from_client_redirect_uris' enabled"
	errFmtOIDCCORSInvalidEndpoint                  = "identity_providers: oidc: cors: option 'endpoints' contains an invalid value '%s': must be one of '%s'"

	errFmtOIDCClientsDuplicateID = "identity_providers: oidc: one or more clients have the same id but all client" +
		"id's must be unique"
	errFmtOIDCClientsWithEmptyID = "identity_providers: oidc: one or more clients have been configured with " +
		"an empty id"

	errFmtOIDCClientInvalidSecret       = "identity_providers: oidc: client '%s': option 'secret' is required"
	errFmtOIDCClientPublicInvalidSecret = "identity_providers: oidc: client '%s': option 'secret' is " +
		"required to be empty when option 'public' is true"
	errFmtOIDCClientRedirectURICantBeParsed = "identity_providers: oidc: client '%s': option 'redirect_uris' has an " +
		"invalid value: redirect uri '%s' could not be parsed: %v"
	errFmtOIDCClientRedirectURIPublic = "identity_providers: oidc: client '%s': option 'redirect_uris' has the " +
		"redirect uri '%s' when option 'public' is false but this is invalid as this uri is not valid " +
		"for the openid connect confidential client type"
	errFmtOIDCClientRedirectURIAbsolute = "identity_providers: oidc: client '%s': option 'redirect_uris' has an " +
		"invalid value: redirect uri '%s' must have the scheme but it is absent"
	errFmtOIDCClientInvalidPolicy = "identity_providers: oidc: client '%s': option 'policy' must be 'one_factor' " +
		"or 'two_factor' but it is configured as '%s'"
	errFmtOIDCClientInvalidConsentMode = "identity_providers: oidc: client '%s': consent: option 'mode' must be one of " +
		"'%s' but it is configured as '%s'"
	errFmtOIDCClientInvalidEntry = "identity_providers: oidc: client '%s': option '%s' must only have the values " +
		"'%s' but one option is configured as '%s'"
	errFmtOIDCClientInvalidUserinfoAlgorithm = "identity_providers: oidc: client '%s': option " +
		"'userinfo_signing_algorithm' must be one of '%s' but it is configured as '%s'"
	errFmtOIDCClientInvalidSectorIdentifier = "identity_providers: oidc: client '%s': option " +
		"'sector_identifier' with value '%s': must be a URL with only the host component for example '%s' but it has a %s with the value '%s'"
	errFmtOIDCClientInvalidSectorIdentifierWithoutValue = "identity_providers: oidc: client '%s': option " +
		"'sector_identifier' with value '%s': must be a URL with only the host component for example '%s' but it has a %s"
	errFmtOIDCClientInvalidSectorIdentifierHost = "identity_providers: oidc: client '%s': option " +
		"'sector_identifier' with value '%s': must be a URL with only the host component but appears to be invalid"
	errFmtOIDCServerInsecureParameterEntropy = "openid connect provider: SECURITY ISSUE - minimum parameter entropy is " +
		"configured to an unsafe value, it should be above 8 but it's configured to %d"
)

// Webauthn Error constants.
const (
	errFmtWebauthnConveyancePreference = "webauthn: option 'attestation_conveyance_preference' must be one of '%s' but it is configured as '%s'"
	errFmtWebauthnUserVerification     = "webauthn: option 'user_verification' must be one of 'discouraged', 'preferred', 'required' but it is configured as '%s'"
)

// Access Control error constants.
const (
	errFmtAccessControlDefaultPolicyValue = "access control: option 'default_policy' must be one of '%s' but it is " +
		"configured as '%s'"
	errFmtAccessControlDefaultPolicyWithoutRules = "access control: 'default_policy' option '%s' is invalid: when " +
		"no rules are specified it must be 'two_factor' or 'one_factor'"
	errFmtAccessControlNetworkGroupIPCIDRInvalid = "access control: networks: network group '%s' is invalid: the " +
		"network '%s' is not a valid IP or CIDR notation"
	errFmtAccessControlWarnNoRulesDefaultPolicy = "access control: no rules have been specified so the " +
		"'default_policy' of '%s' is going to be applied to all requests"
	errFmtAccessControlRuleNoDomains = "access control: rule %s: rule is invalid: must have the option " +
		"'domain' or 'domain_regex' configured"
	errFmtAccessControlRuleInvalidPolicy = "access control: rule %s: rule 'policy' option '%s' " +
		"is invalid: must be one of 'deny', 'two_factor', 'one_factor' or 'bypass'"
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
	errFmtAccessControlRuleMethodInvalid = "access control: rule %s: 'methods' option '%s' is " +
		"invalid: must be one of '%s'"
	errFmtAccessControlRuleQueryInvalid = "access control: rule %s: 'query' option 'operator' with value '%s' is " +
		"invalid: must be one of '%s'"
	errFmtAccessControlRuleQueryInvalidNoValue = "access control: rule %s: 'query' option '%s' is " +
		"invalid: must have a value"
	errFmtAccessControlRuleQueryInvalidNoValueOperator = "access control: rule %s: 'query' option '%s' is " +
		"invalid: must have a value when the operator is '%s'"
	errFmtAccessControlRuleQueryInvalidValue = "access control: rule %s: 'query' option '%s' is " +
		"invalid: must not have a value when the operator is '%s'"
	errFmtAccessControlRuleQueryInvalidValueParse = "access control: rule %s: 'query' option '%s' is " +
		"invalid: %w"
	errFmtAccessControlRuleQueryInvalidValueType = "access control: rule %s: 'query' option 'value' is " +
		"invalid: expected type was string but got %T"
)

// Theme Error constants.
const (
	errFmtThemeName = "option 'theme' must be one of '%s' but it is configured as '%s'"
)

// NTP Error constants.
const (
	errFmtNTPVersion = "ntp: option 'version' must be either 3 or 4 but it is configured as '%d'"
)

// Session error constants.
const (
	errFmtSessionOptionRequired           = "session: option '%s' is required"
	errFmtSessionDomainMustBeRoot         = "session: option 'domain' must be the domain you wish to protect not a wildcard domain but it is configured as '%s'"
	errFmtSessionSameSite                 = "session: option 'same_site' must be one of '%s' but is configured as '%s'"
	errFmtSessionSecretRequired           = "session: option 'secret' is required when using the '%s' provider"
	errFmtSessionRedisPortRange           = "session: redis: option 'port' must be between 1 and 65535 but is configured as '%d'"
	errFmtSessionRedisHostRequired        = "session: redis: option 'host' is required"
	errFmtSessionRedisHostOrNodesRequired = "session: redis: option 'host' or the 'high_availability' option 'nodes' is required"
	errFmtSessionRedisTLSConfigInvalid    = "session: redis: tls: %w"

	errFmtSessionRedisSentinelMissingName     = "session: redis: high_availability: option 'sentinel_name' is required"
	errFmtSessionRedisSentinelNodeHostMissing = "session: redis: high_availability: option 'nodes': option 'host' is required for each node but one or more nodes are missing this"
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

	errFmtServerEndpointsAuthzImplementation    = "server: endpoints: authz: %s: option 'implementation' must be one of '%s' but is configured as '%s'"
	errFmtServerEndpointsAuthzStrategy          = "server: endpoints: authz: %s: authn_strategies: option 'name' must be one of '%s' but is configured as '%s'"
	errFmtServerEndpointsAuthzStrategyDuplicate = "server: endpoints: authz: %s: authn_strategies: duplicate strategy name detected with name '%s'"
)

const (
	errPasswordPolicyMultipleDefined                        = "password_policy: only a single password policy mechanism can be specified"
	errFmtPasswordPolicyStandardMinLengthNotGreaterThanZero = "password_policy: standard: option 'min_length' must be greater than 0 but is configured as %d"
	errFmtPasswordPolicyZXCVBNMinScoreInvalid               = "password_policy: zxcvbn: option 'min_score' is invalid: must be between 1 and 4 but it's configured as %d"
)

const (
	errFmtDuoMissingOption = "duo_api: option '%s' is required when duo is enabled but it is missing"
)

// Error constants.
const (
	/*
		errFmtDeprecatedConfigurationKey = "the %s configuration option is deprecated and will be " +
			"removed in %s, please use %s instead"

		Uncomment for use when deprecating keys.

		TODO: Create a method from within Koanf to automatically remap deprecated keys and produce warnings.
		TODO (cont): The main consideration is making sure we do not overwrite the destination key name if it already exists.
	*/

	errFmtInvalidDefault2FAMethod = "option 'default_2fa_method' is configured as '%s' but must be one of " +
		"the following values: '%s'"
	errFmtInvalidDefault2FAMethodDisabled = "option 'default_2fa_method' is configured as '%s' " +
		"but must be one of the following enabled method values: '%s'"

	errFmtReplacedConfigurationKey = "invalid configuration key '%s' was replaced by '%s'"

	errFmtLoggingLevelInvalid = "log: option 'level' must be one of '%s' but it is configured as '%s'"

	errFileHashing  = "config key incorrect: authentication_backend.file.hashing should be authentication_backend.file.password"
	errFilePHashing = "config key incorrect: authentication_backend.file.password_hashing should be authentication_backend.file.password"
	errFilePOptions = "config key incorrect: authentication_backend.file.password_options should be authentication_backend.file.password"
)

var validArgon2Variants = []string{"argon2id", "id", "argon2i", "i", "argon2d", "d"}

var validSHA2CryptVariants = []string{digestSHA256, digestSHA512}

var validPBKDF2Variants = []string{digestSHA1, digestSHA224, digestSHA256, digestSHA384, digestSHA512}

var validBCryptVariants = []string{"standard", digestSHA256}

var validHashAlgorithms = []string{hashSHA2Crypt, hashPBKDF2, hashSCrypt, hashBCrypt, hashArgon2}

var validStoragePostgreSQLSSLModes = []string{"disable", "require", "verify-ca", "verify-full"}

var validThemeNames = []string{"light", "dark", "grey", "auto"}

var validSessionSameSiteValues = []string{"none", "lax", "strict"}

var validLogLevels = []string{"trace", "debug", "info", "warn", "error"}

var validAuthzImplementations = []string{"AuthRequest", "ForwardAuth", "ExtAuthz", "Legacy"}

var validAuthzAuthnStrategies = []string{"CookieSession", "HeaderAuthorization", "HeaderProxyAuthorization", "HeaderAuthRequestProxyAuthorization", "HeaderLegacy"}

var validWebauthnConveyancePreferences = []string{string(protocol.PreferNoAttestation), string(protocol.PreferIndirectAttestation), string(protocol.PreferDirectAttestation)}

var validWebauthnUserVerificationRequirement = []string{string(protocol.VerificationDiscouraged), string(protocol.VerificationPreferred), string(protocol.VerificationRequired)}

var validRFC7231HTTPMethodVerbs = []string{"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "TRACE", "CONNECT", "OPTIONS"}

var validRFC4918HTTPMethodVerbs = []string{"COPY", "LOCK", "MKCOL", "MOVE", "PROPFIND", "PROPPATCH", "UNLOCK"}

const (
	operatorPresent    = "present"
	operatorAbsent     = "absent"
	operatorEqual      = "equal"
	operatorNotEqual   = "not equal"
	operatorPattern    = "pattern"
	operatorNotPattern = "not pattern"
)

var (
	validACLHTTPMethodVerbs = append(validRFC7231HTTPMethodVerbs, validRFC4918HTTPMethodVerbs...)
	validACLRulePolicies    = []string{policyBypass, policyOneFactor, policyTwoFactor, policyDeny}
	validACLRuleOperators   = []string{operatorPresent, operatorAbsent, operatorEqual, operatorNotEqual, operatorPattern, operatorNotPattern}
)

var validDefault2FAMethods = []string{"totp", "webauthn", "mobile_push"}

var (
	validOIDCScopes             = []string{oidc.ScopeOpenID, oidc.ScopeEmail, oidc.ScopeProfile, oidc.ScopeGroups, oidc.ScopeOfflineAccess}
	validOIDCGrantTypes         = []string{oidc.GrantTypeImplicit, oidc.GrantTypeRefreshToken, oidc.GrantTypeAuthorizationCode, oidc.GrantTypePassword, oidc.GrantTypeClientCredentials}
	validOIDCResponseModes      = []string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery, oidc.ResponseModeFragment}
	validOIDCUserinfoAlgorithms = []string{oidc.SigningAlgorithmNone, oidc.SigningAlgorithmRSAWithSHA256}
	validOIDCCORSEndpoints      = []string{oidc.EndpointAuthorization, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo}
	validOIDCClientConsentModes = []string{"auto", oidc.ClientConsentModeImplicit.String(), oidc.ClientConsentModeExplicit.String(), oidc.ClientConsentModePreConfigured.String()}
)

var reKeyReplacer = regexp.MustCompile(`\[\d+]`)

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
