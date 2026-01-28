package validator

import (
	"regexp"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/oidc"
)

const (
	loopback = "127.0.0.1"
)

// Policy constants.
const (
	policyBypass    = "bypass"
	policyOneFactor = "one_factor"
	policyTwoFactor = "two_factor"
	policyDeny      = "deny"
)

const (
	i18nAuthelia = "{{authelia}}"
)

const (
	durationZero = time.Duration(0)
)

// Hashing constants.
const (
	hashLegacyArgon2id = "argon2id"
	hashLegacySHA512   = schema.SHA512Lower

	hashArgon2    = "argon2"
	hashSHA2Crypt = "sha2crypt"
	hashPBKDF2    = "pbkdf2"
	hashScrypt    = "scrypt"
	hashBcrypt    = "bcrypt"
)

// Scheme constants.
const (
	schemeHTTP  = "http"
	schemeHTTPS = "https"
	schemeSep   = "://"
)

// General fmt consts.
const (
	errFmtMustBeOneOf = "'%s' must be one of %s but it's configured as '%s'"
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
	errFmtNotifierSMTPAddress                     = "notifier: smtp: option 'address' with value '%s' is invalid: %w"
	errFmtNotifierSMTPAddressLegacyAndModern      = "notifier: smtp: option 'host' and 'port' can't be configured at the same time as 'address'"

	errFmtNotifierStartTlsDisabled = "notifier: smtp: option 'disable_starttls' is enabled: " +
		"opportunistic STARTTLS is explicitly disabled which means all emails will be sent insecurely over plaintext " +
		"and this setting is only necessary for non-compliant SMTP servers which advertise they support STARTTLS " +
		"when they actually don't support STARTTLS"
)

const (
	errSuffixMustBeOneOf = "must be one of %s but it's configured as '%s'"
)

const (
	errFmtDefinitionsUserAttributesReservedOrDefined = "definitions: user_attributes: %s: attribute name '%s' is either reserved or already defined in the authentication backend"
)

// Authentication Backend Error constants.
const (
	errFmtAuthBackendNotConfigured = "authentication_backend: you must ensure either the 'file' or 'ldap' " +
		"authentication backend is configured"
	errFmtAuthBackendMultipleConfigured = "authentication_backend: please ensure only one of the 'file' or 'ldap' " +
		"backend is configured"
	errFmtAuthBackendRefreshInterval = "authentication_backend: option 'refresh_interval' is configured to '%s' but " +
		"it must be either in duration common syntax or one of 'disable', or 'always': %w"
	errFmtAuthBackendPasswordResetCustomURLScheme = "authentication_backend: password_reset: option 'custom_url' is" +
		" configured to '%s' which has the scheme '%s' but the scheme must be either 'http' or 'https'"

	errFmtFileAuthBackendPathNotConfigured              = "authentication_backend: file: option 'path' is required"
	errFmtFileAuthBackendExtraAttributeValueTypeMissing = "authentication_backend: file: extra_attributes: %s: option 'value_type' is required"
	errFmtFileAuthBackendExtraAttributeValueType        = "authentication_backend: file: extra_attributes: %s: option 'value_type' must be one of 'string', 'integer', or 'boolean' but it's configured as '%s'"
	errFmtFileAuthBackendExtraAttributeReserved         = "authentication_backend: file: extra_attributes: %s: attribute name '%s' is reserved"
	errFmtFileAuthBackendPasswordUnknownAlg             = "authentication_backend: file: password: option 'algorithm' " +
		errSuffixMustBeOneOf
	errFmtFileAuthBackendPassword               = "authentication_backend: file: password: %s: "
	errFmtFileAuthBackendPasswordInvalidVariant = errFmtFileAuthBackendPassword +
		"option 'variant' " + errSuffixMustBeOneOf
	errFmtFileAuthBackendPasswordOptionInvalid = errFmtFileAuthBackendPassword +
		"option '%s' is configured as '%d' but must be '%d' when '%s' is set to '%v'"
	errFmtFileAuthBackendPasswordOptionTooLarge = errFmtFileAuthBackendPassword +
		"option '%s' is configured as '%d' but must be less than or equal to '%d'"
	errFmtFileAuthBackendPasswordOptionTooSmall = errFmtFileAuthBackendPassword +
		"option '%s' is configured as '%d' but must be greater than or equal to '%d'"
	errFmtFileAuthBackendPasswordArgon2MemoryTooLow = "authentication_backend: file: password: argon2: " +
		"option 'memory' is configured as '%d' but must be greater than or equal to '%d' or '%d' (the value of 'parallelism) multiplied by '%d'"

	errFmtLDAPAuthBackendUnauthenticatedBindWithPassword     = "authentication_backend: ldap: option 'permit_unauthenticated_bind' can't be enabled when a password is specified"
	errFmtLDAPAuthBackendUnauthenticatedBindWithResetEnabled = "authentication_backend: ldap: option 'permit_unauthenticated_bind' can't be enabled when password reset is enabled"

	errFmtLDAPAuthBackendMissingOption                  = "authentication_backend: ldap: option '%s' is required"
	errFmtLDAPAuthBackendExtraAttributeValueTypeMissing = "authentication_backend: ldap: attributes: extra: %s: option 'value_type' is required"
	errFmtLDAPAuthBackendExtraAttributeValueType        = "authentication_backend: ldap: attributes: extra: %s: option 'value_type' must be one of 'string', 'integer', or 'boolean' but it's configured as '%s'"
	errFmtLDAPAuthBackendExtraAttributeReserved         = "authentication_backend: ldap: attributes: extra: %s: attribute name '%s' is reserved"
	errFmtLDAPAuthBackendTLSConfigInvalid               = "authentication_backend: ldap: tls: %w"
	errFmtLDAPAuthBackendOption                         = "authentication_backend: ldap: option '%s' "
	errFmtLDAPAuthBackendOptionMustBeOneOf              = errFmtLDAPAuthBackendOption +
		errSuffixMustBeOneOf
	errFmtLDAPAuthBackendFilterReplacedPlaceholders = errFmtLDAPAuthBackendOption +
		"has an invalid placeholder: '%s' has been removed, please use '%s' instead"
	errFmtLDAPAuthBackendAddress                    = "authentication_backend: ldap: option 'address' with value '%s' is invalid: %w"
	errFmtLDAPAuthBackendFilterEnclosingParenthesis = errFmtLDAPAuthBackendOption +
		"must contain enclosing parenthesis: '%s' should probably be '(%s)'"
	errFmtLDAPAuthBackendFilterMissingPlaceholder = errFmtLDAPAuthBackendOption +
		"must contain the placeholder '{%s}' but it's absent"
	errFmtLDAPAuthBackendFilterMissingPlaceholderGroupSearchMode = errFmtLDAPAuthBackendOption +
		"must contain one of the %s placeholders when using a group_search_mode of '%s' but they're absent"
	errFmtLDAPAuthBackendFilterMissingAttribute = "authentication_backend: ldap: attributes: option '%s' " +
		"must be provided when using the %s placeholder but it's absent"
)

// TOTP Error constants.
const (
	errFmtTOTPInvalidAlgorithm        = "totp: option 'algorithm' must be one of %s but it's configured as '%s'"
	errFmtTOTPInvalidAllowedAlgorithm = "totp: option 'allowed_algorithm' must be one of %s but one of the values is '%s'"
	errFmtTOTPInvalidPeriod           = "totp: option 'period' option must be 15 or more but it's configured as '%d'"
	errFmtTOTPInvalidAllowedPeriod    = "totp: option 'allowed_periods' option must be 15 or more but one of the values is '%d'"
	errFmtTOTPInvalidDigits           = "totp: option 'digits' must be 6 or 8 but it's configured as '%d'"
	errFmtTOTPInvalidAllowedDigit     = "totp: option 'allowed_digits' must only have the values 6 or 8 but one of the values is '%d'"
	errFmtTOTPInvalidSecretSize       = "totp: option 'secret_size' must be %d or higher but it's configured as '%d'" //nolint:gosec
)

// Storage Error constants.
const (
	errStrStorage                                  = "storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided"
	errStrStorageMultiple                          = "storage: option 'local', 'mysql' and 'postgres' are mutually exclusive but %s have been configured"
	errStrStorageEncryptionKeyMustBeProvided       = "storage: option 'encryption_key' is required"
	errStrStorageEncryptionKeyTooShort             = "storage: option 'encryption_key' must be 20 characters or longer"
	errFmtStorageAddressValidate                   = "storage: %s: option 'address' with value '%s' is invalid: %w"
	errFmtStorageUserMustBeProvided                = "storage: %s: option 'username' is required"
	errFmtStorageOptionMustBeProvided              = "storage: %s: option '%s' is required"
	errFmtStorageOptionAddressConflictWithHostPort = "storage: %s: option 'host' and 'port' can't be configured at the same time as 'address'"
	errFmtStorageFailedToConvertHostPortToAddress  = "storage: %s: option 'address' failed to parse options 'host' and 'port' as address: %w"

	errFmtStorageTLSConfigInvalid                 = "storage: %s: tls: %w"
	errFmtStoragePostgreSQLInvalidSSLMode         = "storage: postgres: ssl: option 'mode' must be one of %s but it's configured as '%s'"
	errFmtStoragePostgreSQLInvalidSSLAndTLSConfig = "storage: postgres: can't define both 'tls' and 'ssl' configuration options"
	warnFmtStoragePostgreSQLInvalidSSLDeprecated  = "storage: postgres: ssl: the ssl configuration options are deprecated and we recommend the tls options instead"
)

// Telemetry Error constants.
const (
	errFmtTelemetryMetricsAddress = "telemetry: metrics: option 'address' with value '%s' is invalid: %w"
)

// OpenID Error constants.
const (
	errFmtOIDCProviderNoClientsConfigured = "identity_providers: oidc: option 'clients' must have one or " +
		"more clients configured"
	errFmtOIDCProviderNoPrivateKey            = "identity_providers: oidc: option `jwks` is required"
	errFmtOIDCProviderEnforcePKCEInvalidValue = "identity_providers: oidc: option 'enforce_pkce' must be 'never', " +
		"'public_clients_only' or 'always', but it's configured as '%s'"
	errFmtOIDCProviderInsecureParameterEntropy       = "identity_providers: oidc: option 'minimum_parameter_entropy' is "
	errFmtOIDCProviderInsecureParameterEntropyUnsafe = errFmtOIDCProviderInsecureParameterEntropy +
		"configured to an unsafe and insecure value, it should at least be %d but it's configured to %d"
	errFmtOIDCProviderInsecureDisabledParameterEntropy = errFmtOIDCProviderInsecureParameterEntropy +
		"disabled which is considered unsafe and insecure"
	errFmtOIDCProviderPrivateKeysInvalid                 = "identity_providers: oidc: jwks: key #%d: option 'key' must be a valid private key but the provided data is malformed as it's missing the public key bits"
	errFmtOIDCProviderPrivateKeysMissing                 = "identity_providers: oidc: jwks: key #%d: option 'key' must be provided"
	errFmtOIDCProviderPrivateKeysWithKeyID               = "identity_providers: oidc: jwks: key #%d with key id '%s': option 'key' must be provided"
	errFmtOIDCProviderPrivateKeysCalcThumbprint          = "identity_providers: oidc: jwks: key #%d: option 'key' failed to calculate thumbprint to configure key id value: %w"
	errFmtOIDCProviderPrivateKeysKeyIDLength             = "identity_providers: oidc: jwks: key #%d with key id '%s': option `key_id` must be 100 characters or less"
	errFmtOIDCProviderPrivateKeysAttributeNotUnique      = "identity_providers: oidc: jwks: key #%d with key id '%s': option '%s' must be unique"
	errFmtOIDCProviderPrivateKeysKeyIDNotValid           = "identity_providers: oidc: jwks: key #%d with key id '%s': option 'key_id' must only contain RFC3986 unreserved characters and must only start and end with alphanumeric characters"
	errFmtOIDCProviderPrivateKeysProperties              = "identity_providers: oidc: jwks: key #%d with key id '%s': option 'key' failed to get key properties: %w"
	errFmtOIDCProviderPrivateKeysInvalidOptionOneOf      = "identity_providers: oidc: jwks: key #%d with key id '%s': option '%s' must be one of %s but it's configured as '%s'"
	errFmtOIDCProviderPrivateKeysRSAKeyLessThan2048Bits  = "identity_providers: oidc: jwks: key #%d with key id '%s': option 'key' is an RSA %d bit private key but it must at minimum be a RSA 2048 bit private key"
	errFmtOIDCProviderPrivateKeysKeyNotRSAOrECDSA        = "identity_providers: oidc: jwks: key #%d with key id '%s': option 'key' must be a RSA private key or ECDSA private key but it's type is %T"
	errFmtOIDCProviderPrivateKeysKeyCertificateMismatch  = "identity_providers: oidc: jwks: key #%d with key id '%s': option 'certificate_chain' does not appear to contain the public key for the private key provided by option 'key'"
	errFmtOIDCProviderPrivateKeysCertificateChainInvalid = "identity_providers: oidc: jwks: key #%d with key id '%s': option 'certificate_chain' produced an error during validation of the chain: %w"
	errFmtOIDCProviderPrivateKeysNoRS256                 = "identity_providers: oidc: jwks: keys: must at least have one key supporting the '%s' algorithm but only has %s"
	errFmtOIDCProviderInvalidValue                       = "identity_providers: oidc: option " +
		errFmtMustBeOneOf

	errFmtOIDCCORSInvalidOrigin                    = "identity_providers: oidc: cors: option 'allowed_origins' contains an invalid value '%s' as it has a %s: origins must only be scheme, hostname, and an optional port"
	errFmtOIDCCORSInvalidOriginWildcard            = "identity_providers: oidc: cors: option 'allowed_origins' contains the wildcard origin '*' with more than one origin but the wildcard origin must be defined by itself"
	errFmtOIDCCORSInvalidOriginWildcardWithClients = "identity_providers: oidc: cors: option 'allowed_origins' contains the wildcard origin '*' cannot be specified with option 'allowed_origins_from_client_redirect_uris' enabled"
	errFmtOIDCCORSInvalidEndpoint                  = "identity_providers: oidc: cors: option 'endpoints' contains an invalid value '%s': must be one of %s"

	errFmtOIDCPolicyInvalidName          = "identity_providers: oidc: authorization_policies: authorization policies must have a name but one with a blank name exists"
	errFmtOIDCPolicyInvalidNameStandard  = "identity_providers: oidc: authorization_policies: policy '%s': option '%s' must not be one of %s but it's configured as '%s'"
	errFmtOIDCPolicyMissingOption        = "identity_providers: oidc: authorization_policies: policy '%s': option '%s' is required"
	errFmtOIDCPolicyRuleMissingOption    = "identity_providers: oidc: authorization_policies: policy '%s': rules: rule #%d: option 'subject' or 'networks' is required"
	errFmtOIDCPolicyRuleInvalidSubject   = "identity_providers: oidc: authorization_policies: policy '%s': rules: rule #%d: option 'subject' with value '%s' is invalid: must start with 'user:' or 'group:'"
	errFmtOIDCPolicyInvalidDefaultPolicy = "identity_providers: oidc: authorization_policies: policy '%s': option 'default_policy' must be one of %s but it's configured as '%s'"
	errFmtOIDCPolicyRuleInvalidPolicy    = "identity_providers: oidc: authorization_policies: policy '%s': rules: rule #%d: option 'policy' must be one of %s but it's configured as '%s'"

	errFmtOIDCClientsDuplicateID = "identity_providers: oidc: clients: option 'id' must be unique for every client but one or more clients share the following 'id' values %s"
	errFmtOIDCClientsWithEmptyID = "identity_providers: oidc: clients: option 'id' is required but was absent on the clients in positions %s"
	errFmtOIDCClientsDeprecated  = "identity_providers: oidc: clients: warnings for clients above indicate deprecated functionality and it's strongly suggested these issues are checked and fixed if they're legitimate issues or reported if they are not as in a future version these warnings will become errors"

	errFmtMustOnlyHaveValues                  = "'%s' must only have the values %s "
	errFmtMustBeConfiguredAs                  = "'%s' must be configured as %s "
	errFmtOIDCClientOption                    = "identity_providers: oidc: clients: client '%s': option "
	errFmtOIDCWhenScope                       = "when configured with scope '%s'"
	errFmtOIDCClientInvalidSecretIs           = errFmtOIDCClientOption + "'client_secret' is "
	errFmtOIDCClientInvalidSecret             = errFmtOIDCClientInvalidSecretIs + "required"
	errFmtOIDCClientInvalidSecretPlainText    = errFmtOIDCClientInvalidSecretIs + "plaintext but for clients not using any endpoint authentication method 'client_secret_jwt' it should be a hashed value as plaintext values are deprecated with the exception of 'client_secret_jwt' and will be removed in the near future"
	errFmtOIDCClientInvalidSecretNotPlainText = errFmtOIDCClientOption + "'client_secret' must be plaintext with option '%s' with a value of '%s'"

	errFmtOIDCClientPublicInvalidSecret = errFmtOIDCClientInvalidSecretIs +
		"required to be empty when option 'public' is true"
	errFmtOIDCClientPublicInvalidSecretClientAuthMethod = errFmtOIDCClientInvalidSecretIs +
		"required to be empty when option '%s' is configured as '%s'"
	errFmtOIDCClientIDTooLong           = errFmtOIDCClientOption + "'id' must not be more than 100 characters but it has %d characters"
	errFmtOIDCClientIDInvalidCharacters = errFmtOIDCClientOption + "'id' must only contain RFC3986 unreserved characters"

	errFmtOIDCClientRedirectURIHas          = errFmtOIDCClientOption + "'redirect_uris' has "
	errFmtOIDCClientRedirectURICantBeParsed = errFmtOIDCClientRedirectURIHas +
		"an invalid value: redirect uri '%s' could not be parsed: %v"
	errFmtOIDCClientRedirectURIPublic = errFmtOIDCClientRedirectURIHas +
		"the redirect uri '%s' when option 'public' is false but this is invalid as this uri is not valid " +
		"for the openid connect confidential client type"
	errFmtOIDCClientRedirectURIAbsolute = errFmtOIDCClientRedirectURIHas +
		"an invalid value: redirect uri '%s' must have a scheme but it's absent"

	errFmtOIDCClientRequestURIHas          = errFmtOIDCClientOption + "'request_uris' has "
	errFmtOIDCClientRequestURICantBeParsed = errFmtOIDCClientRequestURIHas +
		"an invalid value: request uri '%s' could not be parsed: %v"
	errFmtOIDCClientRequestURINotAbsolute = errFmtOIDCClientRequestURIHas +
		"an invalid value: request uri '%s' must have a scheme but it's absent"
	errFmtOIDCClientRequestURIInvalidScheme = errFmtOIDCClientRequestURIHas +
		"an invalid scheme: scheme must be 'https' but request uri '%s' has a '%s' scheme"

	errFmtOIDCClientInvalidConsentMode = "identity_providers: oidc: clients: client '%s': consent: option 'mode' must be one of " +
		"%s but it's configured as '%s'"
	errFmtOIDCClientInvalidEntries = errFmtOIDCClientOption + errFmtMustOnlyHaveValues +
		"but the values %s are present"
	errFmtOIDCClientUnknownScopeEntries = errFmtOIDCClientOption + "'%s' only expects the values " +
		"%s but the unknown values %s are present and should generally only be used if a particular client requires a scope outside of our standard scopes"
	errFmtOIDCClientInvalidEntriesScope = errFmtOIDCClientOption + errFmtMustOnlyHaveValues +
		errFmtOIDCWhenScope + " but the values %s are present"
	errFmtOIDCClientEmptyEntriesScope = errFmtOIDCClientOption + errFmtMustOnlyHaveValues +
		errFmtOIDCWhenScope + " but it's not configured"
	errFmtOIDCClientOptionRequiredScope             = errFmtOIDCClientOption + "'%s' must be configured " + errFmtOIDCWhenScope + " but it's absent"
	errFmtOIDCClientOptionMustScope                 = errFmtOIDCClientOption + errFmtMustBeConfiguredAs + errFmtOIDCWhenScope + " but it's configured as '%s'"
	errFmtOIDCClientOptionMustScopeClientType       = errFmtOIDCClientOption + errFmtMustBeConfiguredAs + errFmtOIDCWhenScope + " and the '%s' client type but it's configured as '%s'"
	errFmtOIDCClientInvalidEntriesClientCredentials = errFmtOIDCClientOption + "'scopes' has the values " +
		"%s however when utilizing the 'client_credentials' value for the 'grant_types' the values %s are not allowed"
	errFmtOIDCClientInvalidEntryDuplicates = errFmtOIDCClientOption + "'%s' must have unique values but the values %s are duplicated"
	errFmtOIDCClientInvalidValue           = errFmtOIDCClientOption +
		errFmtMustBeOneOf
	errFmtOIDCClientInvalidLifespan = errFmtOIDCClientOption +
		"'lifespan' must not be configured when no custom lifespans are configured but it's configured as '%s'"
	errFmtOIDCClientInvalidEndpointAuthMethod = errFmtOIDCClientOption +
		"'%s' must be one of %s when configured as the confidential client type unless it only includes implicit flow response types such as %s but it's configured as '%s'"
	errFmtOIDCClientInvalidEndpointAuthMethodPublic = errFmtOIDCClientOption +
		"'%s' must be 'none' when configured as the public client type but it's configured as '%s'"
	errFmtOIDCClientInvalidEndpointAuthSigAlg = errFmtOIDCClientOption +
		"'%s' must be one of %s when option '%s' is configured to '%s'"
	errFmtOIDCClientInvalidTokenEndpointAuthSigAlgReg = errFmtOIDCClientOption +
		"'token_endpoint_auth_signing_alg' must be one of the registered public key algorithm values %s when option 'token_endpoint_auth_method' is configured to '%s'"
	errFmtOIDCClientInvalidTokenEndpointAuthSigAlgMissingPrivateKeyJWT = errFmtOIDCClientOption +
		"'token_endpoint_auth_signing_alg' is required when option 'token_endpoint_auth_method' is configured to 'private_key_jwt'"
	errFmtOIDCClientInvalidPublicKeysPrivateKeyJWT = errFmtOIDCClientOption +
		"'jwks_uri' or 'jwks' is required with 'token_endpoint_auth_method' set to 'private_key_jwt'"
	errFmtOIDCClientInvalidSectorIdentifierAbsolute = errFmtOIDCClientOption +
		"'sector_identifier_uri' with value '%s': should be an absolute URI"
	errFmtOIDCClientInvalidSectorIdentifierScheme = errFmtOIDCClientOption +
		"'sector_identifier_uri' with value '%s': must have the 'https' scheme but has the '%s' scheme"
	errFmtOIDCClientInvalidSectorIdentifier = errFmtOIDCClientOption +
		"'sector_identifier_uri' with value '%s': must not have a %s but it has a %s with the value '%s'"
	errFmtOIDCClientInvalidSectorIdentifierRedirect = errFmtOIDCClientOption +
		"'sector_identifier_uri' with value '%s': must be a json document that contains all of the 'redirect_uris' for the client but had an error validating it: %w"
	errFmtOIDCClientInvalidGrantTypeMatch = errFmtOIDCClientOption +
		"'grant_types' should only have grant type values which are valid with the configured 'response_types' for the client but '%s' expects a response type %s such as %s but the response types are %s"
	errFmtOIDCClientInvalidGrantTypeRefresh = errFmtOIDCClientOption +
		"'grant_types' should only have the 'refresh_token' value if the client is also configured with the 'offline_access' scope"
	errFmtOIDCClientInvalidGrantTypePublic = errFmtOIDCClientOption + "'grant_types' " +
		"should only have the '%s' value if it is of the confidential client type but it's of the public client type"

	errFmtOIDCClientInvalidRefreshTokenOptionWithoutCodeResponseType = errFmtOIDCClientOption +
		"'%s' should only have the values %s if the client is also configured with a 'response_type' such as %s which respond with authorization codes"

	errFmtOIDCClientPublicKeysBothURIAndValuesConfigured      = "identity_providers: oidc: clients: client '%s': option 'jwks_uri' must not be defined at the same time as option 'jwks'"
	errFmtOIDCClientPublicKeysURIInvalidScheme                = "identity_providers: oidc: clients: client '%s': option 'jwks_uri' must have the 'https' scheme but the scheme is '%s'"
	errFmtOIDCClientPublicKeysProperties                      = "identity_providers: oidc: clients: client '%s': jwks: key #%d with key id '%s': option 'key' failed to get key properties: %w"
	errFmtOIDCClientPublicKeysInvalidOptionOneOf              = "identity_providers: oidc: clients: client '%s': jwks: key #%d with key id '%s': option '%s' must be one of %s but it's configured as '%s'"
	errFmtOIDCClientPublicKeysInvalidOptionMissingOneOf       = "identity_providers: oidc: clients: client '%s': jwks: key #%d: option '%s' must be provided"
	errFmtOIDCClientPublicKeysWithIDInvalidOptionMissingOneOf = "identity_providers: oidc: clients: client '%s': jwks: key #%d with key id '%s': option '%s' must be provided"
	errFmtOIDCClientPublicKeysKeyMalformed                    = "identity_providers: oidc: clients: client '%s': jwks: key #%d: option 'key' option 'key' must be a valid private key but the provided data is malformed as it's missing the public key bits"
	errFmtOIDCClientPublicKeysRSAKeyLessThan2048Bits          = "identity_providers: oidc: clients: client '%s': jwks: key #%d with key id '%s': option 'key' is an RSA %d bit private key but it must at minimum be a RSA 2048 bit private key"
	errFmtOIDCClientPublicKeysKeyNotRSAOrECDSA                = "identity_providers: oidc: clients: client '%s': jwks: key #%d with key id '%s': option 'key' must be a RSA public key or ECDSA public key but it's type is %T"
	errFmtOIDCClientPublicKeysCertificateChainKeyMismatch     = "identity_providers: oidc: clients: client '%s': jwks: key #%d with key id '%s': option 'certificate_chain' does not appear to contain the public key for the public key provided by option 'key'"
	errFmtOIDCClientPublicKeysCertificateChainInvalid         = "identity_providers: oidc: clients: client '%s': jwks: key #%d with key id '%s': option 'certificate_chain' produced an error during validation of the chain: %w"
	errFmtOIDCClientPublicKeysROSAMissingAlgorithm            = errFmtOIDCClientOption + "'request_object_signing_alg' must be one of %s configured in the client option 'jwks'"
)

// WebAuthn Error constants.
const (
	errFmtWebAuthnConveyancePreference   = "webauthn: option 'attestation_conveyance_preference' must be one of %s but it's configured as '%s'"
	errFmtWebAuthnSelectionCriteria      = "webauthn: selection_criteria: option '%s' must be one of %s but it's configured as '%s'"
	errFmtWebAuthnPasskeyDiscoverability = "webauthn: selection_criteria: option 'discoverability' should generally be configured as '%s' or '%s' when passkey logins are enabled" //nolint:gosec
	errFmtWebAuthnFiltering              = "webauthn: filtering: option 'permitted_aaguids' and 'prohibited_aaguids' are mutually exclusive however both have values"
	errFmtWebAuthnBoolean                = "webauthn: option '%s' is %t but it must be %t when '%s' is %t"
	errFmtWebAuthnMetadataString         = "webauthn: metadata: option '%s' is '%s' but it must be %s"
)

// Access Control error constants.
const (
	errFmtAccessControlDefaultPolicyValue = "access_control: option 'default_policy' must be one of %s but it's " +
		"configured as '%s'"
	errFmtAccessControlDefaultPolicyWithoutRules = "access_control: 'default_policy' option '%s' is invalid: when " +
		"no rules are specified it must be 'two_factor' or 'one_factor'"
	errFmtAccessControlNetworkGroupIPCIDRInvalid = "access_control: networks: network group '%s' is invalid: the " +
		"network '%s' is not a valid IP or CIDR notation"
	errFmtAccessControlWarnNoRulesDefaultPolicy = "access_control: no rules have been specified so the " +
		"'default_policy' of '%s' is going to be applied to all requests"
	errFmtAccessControlRuleNoDomains                    = "access_control: rule %s: option 'domain' or 'domain_regex' must be present but are both absent"
	errFmtAccessControlRuleNoPolicy                     = "access_control: rule %s: option 'policy' must be present but it's absent"
	errFmtAccessControlRuleInvalidPolicy                = "access_control: rule %s: option 'policy' must be one of %s but it's configured as '%s'"
	errAccessControlRuleBypassPolicyOptionBypassIs      = "access_control: rule %s: 'policy' option 'bypass' is "
	errAccessControlRuleBypassPolicyInvalidWithSubjects = errAccessControlRuleBypassPolicyOptionBypassIs +
		"not supported when 'subject' option is configured: see " +
		"https://www.authelia.com/c/acl#bypass"
	errAccessControlRuleBypassPolicyInvalidWithSubjectsWithGroupDomainRegex = errAccessControlRuleBypassPolicyOptionBypassIs +
		"not supported when 'domain_regex' option contains the user or group named matches. For more information see: " +
		"https://www.authelia.com/c/acl-match-concept-2"
	errFmtAccessControlRuleNetworksInvalid = "access_control: rule %s: the network '%s' is not a " +
		"valid Group Name, IP, or CIDR notation"
	errFmtAccessControlRuleSubjectInvalid = "access_control: rule %s: 'subject' option '%s' is " +
		"invalid: must start with 'user:', 'group:', or 'oauth2:client:'"
	errFmtAccessControlRuleOAuth2ClientSubjectInvalid = "access_control: rule %s: option 'subject' with value '%s' is " +
		"invalid: the client id '%s' does not belong to a registered client"
	errFmtAccessControlRuleInvalidEntries              = "access_control: rule %s: option '%s' must only have the values %s but the values %s are present"
	errFmtAccessControlRuleInvalidDuplicates           = "access_control: rule %s: option '%s' must have unique values but the values %s are duplicated"
	errFmtAccessControlRuleQueryInvalid                = "access_control: rule %s: query: option 'operator' must be one of %s but it's configured as '%s'"
	errFmtAccessControlRuleQueryInvalidNoValue         = "access_control: rule %s: query: option '%s' is required but it's absent"
	errFmtAccessControlRuleQueryInvalidNoValueOperator = "access_control: rule %s: query: option '%s' must be present when the option 'operator' is '%s' but it's absent"
	errFmtAccessControlRuleQueryInvalidValue           = "access_control: rule %s: query: option '%s' must not be present when the option 'operator' is '%s' but it's present"
	errFmtAccessControlRuleQueryInvalidValueParse      = "access_control: rule %s: query: option '%s' is " +
		"invalid: %w"
	errFmtAccessControlRuleQueryInvalidValueType = "access_control: rule %s: query: option 'value' is " +
		"invalid: expected type was string but got %T"
)

// Theme Error constants.
const (
	errFmtThemeName = "option 'theme' must be one of %s but it's configured as '%s'"
)

// NTP Error constants.
const (
	errFmtNTPVersion       = "ntp: option 'version' must be either 3 or 4 but it's configured as '%d'"
	errFmtNTPAddressScheme = "ntp: option 'address' with value '%s' is invalid: %w"
)

// Session error constants.
const (
	errFmtSessionDomainLegacy             = "session: option 'domain' is deprecated in v4.38.0 and has been replaced by a multi-domain configuration: this has automatically been mapped for you but you will need to adjust your configuration to remove this message and receive the latest messages"
	errFmtSessionLegacyRedirectionURL     = "session: option 'cookies' must be configured with the per cookie option 'default_redirection_url' but the global one is configured which is not supported"
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

	errFmtSessionFilePathRequired       = "session: file: option 'path' is required"
	errFmtSessionFilePathNotAbsolute    = "session: file: option 'path' must be an absolute path but it's configured as '%s'"
	errFmtSessionFileAndRedisConfigured = "session: only one of 'redis' or 'file' can be configured"

	errFmtSessionDomainMustBeRoot                        = "session: domain config %s: option 'domain' must be the domain you wish to protect not a wildcard domain but it's configured as '%s'"
	errFmtSessionDomainSameSite                          = "session: domain config %s: option 'same_site' must be one of %s but it's configured as '%s'"
	errFmtSessionDomainOptionRequired                    = "session: domain config %s: option '%s' is required"
	errFmtSessionDomainHasPeriodPrefix                   = "session: domain config %s: option 'domain' has a prefix of '.' which is not supported or intended behaviour: you can use this at your own risk but we recommend removing it"
	errFmtSessionDomainDuplicate                         = "session: domain config %s: option 'domain' is a duplicate value for another configured session domain"
	errFmtSessionDomainDuplicateCookieScope              = "session: domain config %s: option 'domain' shares the same cookie domain scope as another configured session domain"
	errFmtSessionDomainURLNotAbsolute                    = "session: domain config %s: option '%s' is not absolute with a value of '%s'"
	errFmtSessionDomainURLInsecure                       = "session: domain config %s: option '%s' does not have a secure scheme with a value of '%s'"
	errFmtSessionDomainURLNotInCookieScope               = "session: domain config %s: option '%s' does not share a cookie scope with domain '%s' with a value of '%s'"
	errFmtSessionDomainAutheliaURLAndRedirectionURLEqual = "session: domain config %s: option 'default_redirection_url' with value '%s' is effectively equal to option 'authelia_url' with value '%s' which is not permitted"
	errFmtSessionDomainInvalidDomain                     = "session: domain config %s: option 'domain' does not appear to be a valid cookie domain or an ip address"
	errFmtSessionDomainInvalidDomainNoDots               = "session: domain config %s: option 'domain' is not a valid cookie domain: must have at least a single period or be an ip address"
	errFmtSessionDomainInvalidDomainPublic               = "session: domain config %s: option 'domain' is not a valid cookie domain: the domain is part of the special public suffix list"
)

// Regulation Error Consts.
const (
	errFmtRegulationFindTimeGreaterThanBanTime = "regulation: option 'find_time' must be less than or equal to option 'ban_time'"
	errFmtRegulationInvalidMode                = "regulation: option 'modes' must only contain the values 'user' and 'ip' but contains the value '%s'"
)

// Server Error constants.
const (
	errFmtServerTLSCert             = "server: tls: option 'key' must also be accompanied by option 'certificate'"
	errFmtServerTLSKey              = "server: tls: option 'certificate' must also be accompanied by option 'key'"
	errFmtServerTLSClientAuthNoAuth = "server: tls: client authentication cannot be configured if no server certificate and key are provided"

	errFmtServerAddress = "server: option 'address' with value '%s' is invalid: %w"

	errFmtServerPathNotEndForwardSlash = "server: option 'address' must not have a path with a forward slash but it's configured as '%s'"
	errFmtServerPathAlphaNumeric       = "server: option 'address' must have a path with only alphanumeric characters but it's configured as '%s'"

	errFmtServerEndpointsAuthzImplementation            = "server: endpoints: authz: %s: option 'implementation' must be one of %s but it's configured as '%s'"
	errFmtServerEndpointsAuthzStrategy                  = "server: endpoints: authz: %s: authn_strategies: option 'name' must be one of %s but it's configured as '%s'"
	errFmtServerEndpointsAuthzSchemes                   = "server: endpoints: authz: %s: authn_strategies: strategy #%d (%s): option 'schemes' must only include the values %s but has '%s'"
	errFmtServerEndpointsAuthzSchemesInvalidForStrategy = "server: endpoints: authz: %s: authn_strategies: strategy #%d (%s): option 'schemes' is not valid for the strategy"
	errFmtServerEndpointsAuthzStrategyNoName            = "server: endpoints: authz: %s: authn_strategies: strategy #%d: option 'name' must be configured"
	errFmtServerEndpointsAuthzStrategySchemeOnlyOption  = "server: endpoints: authz: %s: authn_strategies: strategy #%d: option '%s' can't be configured unless the '%s' scheme is configured but only the %s schemes are configured"
	errFmtServerEndpointsAuthzStrategyDuplicate         = "server: endpoints: authz: %s: authn_strategies: duplicate strategy name detected with name '%s'"
	errFmtServerEndpointsAuthzPrefixDuplicate           = "server: endpoints: authz: %s: endpoint starts with the same prefix as the '%s' endpoint with the '%s' implementation which accepts prefixes as part of its implementation"
	errFmtServerEndpointsRateLimitsBucketPeriodZero     = "server: endpoints: rate_limits: %s: bucket %d: option 'period' must have a value"
	errFmtServerEndpointsRateLimitsBucketPeriodTooLow   = "server: endpoints: rate_limits: %s: bucket %d: option 'period' has a value of '%s' but it must be greater than 10 seconds"
	errFmtServerEndpointsRateLimitsBucketRequestsZero   = "server: endpoints: rate_limits: %s: bucket %d: option 'requests' has a value of '%d' but it must be greater than 1"
	errFmtServerEndpointsAuthzInvalidName               = "server: endpoints: authz: %s: contains invalid characters"

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

	errFmtLoggingInvalid = "log: option '%s' must be one of %s but it's configured as '%s'"

	errFmtCookieDomainInPSL = "%s is a suffix"

	errFileHashing  = "config key incorrect: authentication_backend.file.hashing should be authentication_backend.file.password"
	errFilePHashing = "config key incorrect: authentication_backend.file.password_hashing should be authentication_backend.file.password"
	errFilePOptions = "config key incorrect: authentication_backend.file.password_options should be authentication_backend.file.password"
)

const (
	errFmtIdentityValidationResetPasswordJWTAlgorithm      = "identity_validation: reset_password: option 'jwt_algorithm' must be one of %s but it's configured as '%s'"
	errFmtIdentityValidationResetPasswordJWTSecret         = "identity_validation: reset_password: option 'jwt_secret' is required when the reset password functionality isn't disabled"
	errFmtIdentityValidationElevatedSessionCharacterLength = "identity_validation: elevated_session: option 'characters' must be 20 or less but it's configured as %d"
)

const (
	operatorPresent    = "present"
	operatorAbsent     = "absent"
	operatorEqual      = "equal"
	operatorNotEqual   = "not equal"
	operatorPattern    = "pattern"
	operatorNotPattern = "not pattern"
)

const (
	legacy = "legacy"
)

const (
	auto = "auto"
)

var (
	validAuthzImplementations       = []string{schema.AuthzImplementationAuthRequest, schema.AuthzImplementationForwardAuth, schema.AuthzImplementationExtAuthz, schema.AuthzImplementationLegacy}
	validAuthzAuthnStrategies       = []string{schema.AuthzStrategyHeaderCookieSession, schema.AuthzStrategyHeaderAuthorization, schema.AuthzStrategyHeaderProxyAuthorization, schema.AuthzStrategyHeaderAuthRequestProxyAuthorization, schema.AuthzStrategyHeaderLegacy}
	validAuthzAuthnHeaderStrategies = []string{schema.AuthzStrategyHeaderAuthorization, schema.AuthzStrategyHeaderProxyAuthorization, schema.AuthzStrategyHeaderAuthRequestProxyAuthorization}
	validAuthzAuthnStrategySchemes  = []string{schema.SchemeBasic, schema.SchemeBearer}
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

	validLDAPGroupSearchModes = []string{
		schema.LDAPGroupSearchModeFilter,
		schema.LDAPGroupSearchModeMemberOf,
	}
)

var (
	validArgon2Variants    = []string{"argon2id", "id", "argon2i", "i", "argon2d", "d"}
	validSHA2CryptVariants = []string{schema.SHA256Lower, schema.SHA512Lower}
	validPBKDF2Variants    = []string{schema.SHA1Lower, schema.SHA224Lower, schema.SHA256Lower, schema.SHA384Lower, schema.SHA512Lower}
	validBcryptVariants    = []string{"standard", schema.SHA256Lower}
	validScryptVariants    = []string{"scrypt", "yescrypt"}
	validHashAlgorithms    = []string{hashSHA2Crypt, hashPBKDF2, hashScrypt, hashBcrypt, hashArgon2}
)

var (
	validStoragePostgreSQLSSLModes           = []string{"disable", "require", "verify-ca", "verify-full"}
	validThemeNames                          = []string{"light", "dark", "grey", "oled", auto}
	validSessionSameSiteValues               = []string{"none", "lax", "strict"}
	validLogLevels                           = []string{logging.LevelTrace, logging.LevelDebug, logging.LevelInfo, logging.LevelWarn, logging.LevelError}
	validLogFormats                          = []string{logging.FormatText, logging.FormatJSON}
	validWebAuthnConveyancePreferences       = []string{string(protocol.PreferNoAttestation), string(protocol.PreferIndirectAttestation), string(protocol.PreferDirectAttestation)}
	validWebAuthnUserVerificationRequirement = []string{string(protocol.VerificationDiscouraged), string(protocol.VerificationPreferred), string(protocol.VerificationRequired)}
	validWebAuthnAttachment                  = []string{string(protocol.Platform), string(protocol.CrossPlatform)}
	validWebAuthnDiscoverability             = []string{string(protocol.ResidentKeyRequirementDiscouraged), string(protocol.ResidentKeyRequirementPreferred), string(protocol.ResidentKeyRequirementRequired)}
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
	attrOIDCKey                         = "key"
	attrOIDCKeyID                       = "key_id"
	attrOIDCKeyUse                      = "use"
	attrOIDCAlgorithm                   = "algorithm"
	attrOIDCScopes                      = "scopes"
	attrOIDCResponseTypes               = "response_types"
	attrOIDCResponseModes               = "response_modes"
	attrOIDCGrantTypes                  = "grant_types"
	attrOIDCRedirectURIs                = "redirect_uris"
	attrOIDCRequestURIs                 = "request_uris"
	attrOIDCTokenAuthMethod             = "token_endpoint_auth_method"
	attrOIDCTokenAuthSigningAlg         = "token_endpoint_auth_signing_alg"
	attrOIDCRevocationAuthMethod        = "revocation_endpoint_auth_method"
	attrOIDCRevocationAuthSigningAlg    = "revocation_endpoint_auth_signing_alg"
	attrOIDCIntrospectionAuthMethod     = "introspection_endpoint_auth_method"
	attrOIDCIntrospectionAuthSigningAlg = "introspection_endpoint_auth_signing_alg"
	attrOIDCPARAuthMethod               = "pushed_authorization_request_endpoint_auth_method"
	attrOIDCPARAuthSigningAlg           = "pushed_authorization_request_endpoint_auth_signing_alg"
	attrOIDCDiscoSigAlg                 = "discovery_signed_response_alg"
	attrOIDCDiscoSigKID                 = "discovery_signed_response_key_id"
	attrOIDCAuthorizationPrefix         = "authorization"
	attrOIDCIDTokenPrefix               = "id_token"
	attrOIDCAccessTokenPrefix           = "access_token"
	attrOIDCUserinfoPrefix              = "userinfo"
	attrOIDCIntrospectionPrefix         = "introspection"
	attrOIDCPKCEChallengeMethod         = "pkce_challenge_method"
	attrOIDCRequestedAudienceMode       = "requested_audience_mode"
	attrSessionAutheliaURL              = "authelia_url"
	attrSessionDomain                   = "domain"
	attrDefaultRedirectionURL           = "default_redirection_url"
)

var (
	validIdentityValidationJWTAlgorithms = []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512}
)

var (
	validOIDCCORSEndpoints = []string{oidc.EndpointAuthorization, oidc.EndpointDeviceAuthorization, oidc.EndpointPushedAuthorizationRequest, oidc.EndpointToken, oidc.EndpointIntrospection, oidc.EndpointRevocation, oidc.EndpointUserinfo}

	validOIDCReservedClaims                  = []string{oidc.ClaimJWTID, oidc.ClaimAuthorizedParty, oidc.ClaimClientIdentifier, oidc.ClaimScope, oidc.ClaimScopeNonStandard, oidc.ClaimIssuer, oidc.ClaimSubject, oidc.ClaimAudience, oidc.ClaimSessionID, oidc.ClaimStateHash, oidc.ClaimCodeHash, oidc.ClaimIssuedAt, oidc.ClaimUpdatedAt, oidc.ClaimNotBefore, oidc.ClaimExpirationTime, oidc.ClaimAuthenticationTime, oidc.ClaimAuthenticationMethodsReference, oidc.ClaimAuthenticationContextClassReference, oidc.ClaimNonce}
	validOIDCReservedIDTokenClaims           = []string{oidc.ClaimJWTID, oidc.ClaimAuthorizedParty, oidc.ClaimScope, oidc.ClaimIssuer, oidc.ClaimSubject, oidc.ClaimAudience, oidc.ClaimSessionID, oidc.ClaimStateHash, oidc.ClaimCodeHash, oidc.ClaimIssuedAt, oidc.ClaimNotBefore, oidc.ClaimExpirationTime, oidc.ClaimAuthenticationTime, oidc.ClaimAuthenticationMethodsReference, oidc.ClaimAuthenticationContextClassReference, oidc.ClaimNonce}
	validOIDCClientClaims                    = []string{oidc.ClaimFullName, oidc.ClaimGivenName, oidc.ClaimFamilyName, oidc.ClaimMiddleName, oidc.ClaimNickname, oidc.ClaimPreferredUsername, oidc.ClaimProfile, oidc.ClaimPicture, oidc.ClaimWebsite, oidc.ClaimEmail, oidc.ClaimEmailVerified, oidc.ClaimGender, oidc.ClaimBirthdate, oidc.ClaimZoneinfo, oidc.ClaimLocale, oidc.ClaimPhoneNumber, oidc.ClaimPhoneNumberVerified, oidc.ClaimAddress, oidc.ClaimGroups, oidc.ClaimEmailAlts, oidc.ClaimRequestedAt, oidc.ClaimUpdatedAt}
	validOIDCClientScopes                    = []string{oidc.ScopeOpenID, oidc.ScopeEmail, oidc.ScopeProfile, oidc.ScopeAddress, oidc.ScopePhone, oidc.ScopeGroups, oidc.ScopeOfflineAccess, oidc.ScopeOffline, oidc.ScopeAutheliaBearerAuthz}
	validOIDCClientConsentModes              = []string{auto, oidc.ClientConsentModeImplicit.String(), oidc.ClientConsentModeExplicit.String(), oidc.ClientConsentModePreConfigured.String()}
	validOIDCClientResponseModes             = []string{oidc.ResponseModeFormPost, oidc.ResponseModeQuery, oidc.ResponseModeFragment, oidc.ResponseModeJWT, oidc.ResponseModeFormPostJWT, oidc.ResponseModeQueryJWT, oidc.ResponseModeFragmentJWT}
	validOIDCClientResponseTypes             = []string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowIDToken, oidc.ResponseTypeImplicitFlowToken, oidc.ResponseTypeImplicitFlowBoth, oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeHybridFlowBoth}
	validOIDCClientResponseTypesImplicitFlow = []string{oidc.ResponseTypeImplicitFlowIDToken, oidc.ResponseTypeImplicitFlowToken, oidc.ResponseTypeImplicitFlowBoth}
	validOIDCClientResponseTypesHybridFlow   = []string{oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeHybridFlowBoth}
	validOIDCClientResponseTypesRefreshToken = []string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeHybridFlowBoth}
	validOIDCClientGrantTypes                = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeImplicit, oidc.GrantTypeClientCredentials, oidc.GrantTypeRefreshToken, oidc.GrantTypeDeviceCode}

	validOIDCClientTokenEndpointAuthMethods                = []string{oidc.ClientAuthMethodNone, oidc.ClientAuthMethodClientSecretPost, oidc.ClientAuthMethodClientSecretBasic, oidc.ClientAuthMethodPrivateKeyJWT, oidc.ClientAuthMethodClientSecretJWT}
	validOIDCClientTokenEndpointAuthMethodsConfidential    = []string{oidc.ClientAuthMethodClientSecretPost, oidc.ClientAuthMethodClientSecretBasic, oidc.ClientAuthMethodPrivateKeyJWT}
	validOIDCClientTokenEndpointAuthSigAlgsClientSecretJWT = []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512}
	validOIDCIssuerJWKSigningAlgs                          = []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgRSAPSSUsingSHA256, oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgRSAUsingSHA384, oidc.SigningAlgRSAPSSUsingSHA384, oidc.SigningAlgECDSAUsingP384AndSHA384, oidc.SigningAlgRSAUsingSHA512, oidc.SigningAlgRSAPSSUsingSHA512, oidc.SigningAlgECDSAUsingP521AndSHA512}
	validOIDCClientJWKEncryptionKeyAlgs                    = []string{oidc.EncryptionAlgNone, oidc.EncryptionAlgRSA15, oidc.EncryptionAlgRSAOAEP, oidc.EncryptionAlgRSAOAEP256, oidc.EncryptionAlgECDHES, oidc.EncryptionAlgECDHESA128KW, oidc.EncryptionAlgECDHESA192KW, oidc.EncryptionAlgECDHESA256KW, oidc.EncryptionAlgA128KW, oidc.EncryptionAlgA192KW, oidc.EncryptionAlgA256KW, oidc.EncryptionAlgA128GCMKW, oidc.EncryptionAlgA192GCMKW, oidc.EncryptionAlgA256GCMKW, oidc.EncryptionAlgPBES2HS256A128KW, oidc.EncryptionAlgPBES2HS284A192KW, oidc.EncryptionAlgPBES2HS512A256KW}
	validOIDCClientJWKContentEncryptionAlgs                = []string{oidc.EncryptionEncA128GCM, oidc.EncryptionEncA192GCM, oidc.EncryptionEncA256GCM, oidc.EncryptionEncA128CBCHS256, oidc.EncryptionEncA192CBCHS384, oidc.EncryptionEncA256CBCHS512}

	validOIDCJWKEncryptionAlgs = []string{oidc.EncryptionAlgRSA15, oidc.EncryptionAlgRSAOAEP, oidc.EncryptionAlgRSAOAEP256, oidc.EncryptionAlgA128KW, oidc.EncryptionAlgA192KW, oidc.EncryptionAlgA256KW, oidc.EncryptionAlgDirect, oidc.EncryptionAlgECDHES, oidc.EncryptionAlgECDHESA128KW, oidc.EncryptionAlgECDHESA192KW, oidc.EncryptionAlgECDHESA256KW, oidc.EncryptionAlgA128GCMKW, oidc.EncryptionAlgA192GCMKW, oidc.EncryptionAlgA256GCMKW, oidc.EncryptionAlgPBES2HS256A128KW, oidc.EncryptionAlgPBES2HS284A192KW, oidc.EncryptionAlgPBES2HS512A256KW}

	validOIDCClientScopesBearerAuthz        = []string{oidc.ScopeOfflineAccess, oidc.ScopeOffline, oidc.ScopeAutheliaBearerAuthz}
	validOIDCClientResponseModesBearerAuthz = []string{oidc.ResponseModeFormPost, oidc.ResponseModeFormPostJWT}
	validOIDCClientResponseTypesBearerAuthz = []string{oidc.ResponseTypeAuthorizationCodeFlow}
	validOIDCClientGrantTypesBearerAuthz    = []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeRefreshToken, oidc.GrantTypeClientCredentials}
)

var (
	reKeyReplacer       = regexp.MustCompile(`\[\d+]`)
	reDomainCharacters  = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?(?:\.[a-z0-9](?:[a-z0-9-]*[a-z0-9])?)+$`)
	reAuthzEndpointName = regexp.MustCompile(`^[a-zA-Z](([a-zA-Z0-9/._-]*)([a-zA-Z]))?$`)
	reOpenIDConnectKID  = regexp.MustCompile(`^([a-zA-Z0-9](([a-zA-Z0-9._~-]*)([a-zA-Z0-9]))?)?$`)
	reRFC3986Unreserved = regexp.MustCompile(`^[a-zA-Z0-9._~-]+$`)
)

const (
	attributeUserUsername       = "username"
	attributeUserGroups         = "groups"
	attributeUserDisplayName    = "display_name"
	attributeUserEmail          = "email"
	attributeUserEmails         = "emails"
	attributeUserGivenName      = "given_name"
	attributeUserMiddleName     = "middle_name"
	attributeUserFamilyName     = "family_name"
	attributeUserNickname       = "nickname"
	attributeUserProfile        = "profile"
	attributeUserPicture        = "picture"
	attributeUserWebsite        = "website"
	attributeUserGender         = "gender"
	attributeUserBirthdate      = "birthdate"
	attributeUserZoneInfo       = "zoneinfo"
	attributeUserLocale         = "locale"
	attributeUserPhoneNumber    = "phone_number"
	attributeUserPhoneExtension = "phone_extension"
	attributeUserStreetAddress  = "street_address"
	attributeUserLocality       = "locality"
	attributeUserRegion         = "region"
	attributeUserPostalCode     = "postal_code"
	attributeUserCountry        = "country"
)

var validUserAttributes = []string{
	attributeUserUsername,
	attributeUserGroups,
	attributeUserDisplayName,
	attributeUserEmail,
	attributeUserEmails,
	attributeUserGivenName,
	attributeUserMiddleName,
	attributeUserFamilyName,
	attributeUserNickname,
	attributeUserProfile,
	attributeUserPicture,
	attributeUserWebsite,
	attributeUserGender,
	attributeUserBirthdate,
	attributeUserZoneInfo,
	attributeUserLocale,
	attributeUserPhoneNumber,
	attributeUserPhoneExtension,
	attributeUserStreetAddress,
	attributeUserLocality,
	attributeUserRegion,
	attributeUserPostalCode,
	attributeUserCountry,
}

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
