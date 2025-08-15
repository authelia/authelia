package schema

import (
	"errors"
	"regexp"
	"time"
)

const (
	argon2   = "argon2"
	argon2id = "argon2id"
)

const (
	SHA1Lower   = "sha1"
	SHA224Lower = "sha224"
	SHA256Lower = "sha256"
	SHA384Lower = "sha384"
	SHA512Lower = "sha512"
)

const (
	// TLSVersion13 is the textual representation of TLS 1.3.
	TLSVersion13 = "TLS1.3"

	// TLSVersion12 is the textual representation of TLS 1.2.
	TLSVersion12 = "TLS1.2"

	// TLSVersion11 is the textual representation of TLS 1.1.
	TLSVersion11 = "TLS1.1"

	// TLSVersion10 is the textual representation of TLS 1.0.
	TLSVersion10 = "TLS1.0"

	// SSLVersion30 is the textual representation of SSL 3.0.
	SSLVersion30 = "SSL3.0"

	// Version13 is the textual representation of version 1.3.
	Version13 = "1.3"

	// Version12 is the textual representation of version 1.2.
	Version12 = "1.2"

	// Version11 is the textual representation of version 1.1.
	Version11 = "1.1"

	// Version10 is the textual representation of version 1.0.
	Version10 = "1.0"
)

// ErrTLSVersionNotSupported returned when an unknown TLS version supplied.
var ErrTLSVersionNotSupported = errors.New("supplied tls version isn't supported")

const (
	// ProfileRefreshAlways represents a value for refresh_interval that's the same as 0ms.
	ProfileRefreshAlways = "always"

	// ProfileRefreshDisabled represents a Value for refresh_interval that disables the check entirely.
	ProfileRefreshDisabled = "disable"

	// RefreshIntervalDefault represents the default value of refresh_interval.
	RefreshIntervalDefault = time.Minute * 5
)

const (
	// LDAPImplementationCustom is the string for the custom LDAP implementation.
	LDAPImplementationCustom = "custom"

	// LDAPImplementationActiveDirectory is the string for the Active Directory LDAP implementation.
	LDAPImplementationActiveDirectory = "activedirectory"

	// LDAPImplementationRFC2307bis is the string for the RFC2307bis LDAP implementation.
	LDAPImplementationRFC2307bis = "rfc2307bis"

	// LDAPImplementationFreeIPA is the string for the FreeIPA LDAP implementation.
	LDAPImplementationFreeIPA = "freeipa"

	// LDAPImplementationLLDAP is the string for the lldap LDAP implementation.
	LDAPImplementationLLDAP = "lldap"

	// LDAPImplementationGLAuth is the string for the GLAuth LDAP implementation.
	LDAPImplementationGLAuth = "glauth"
)

const (
	// LDAPGroupSearchModeFilter is the string for the filter group search mode.
	LDAPGroupSearchModeFilter = "filter"

	// LDAPGroupSearchModeMemberOf is the string for the memberOf group search mode.
	LDAPGroupSearchModeMemberOf = "memberof"
)

// TOTP Algorithm.
const (
	TOTPAlgorithmSHA1   = "SHA1"
	TOTPAlgorithmSHA256 = "SHA256"
	TOTPAlgorithmSHA512 = "SHA512"
)

const (
	// RememberMeDisabled represents the duration for a disabled remember me session configuration.
	RememberMeDisabled = time.Second * -1
)

var (
	// TOTPPossibleAlgorithms is a list of valid TOTP Algorithms.
	TOTPPossibleAlgorithms = []string{TOTPAlgorithmSHA1, TOTPAlgorithmSHA256, TOTPAlgorithmSHA512}
)

const (
	// TOTPSecretSizeDefault is the default secret size.
	TOTPSecretSizeDefault = 32

	// TOTPSecretSizeMinimum is the minimum secret size.
	TOTPSecretSizeMinimum = 20
)

var (
	// regexpHasScheme checks if a string has a scheme. Valid characters for schemes include alphanumeric, hyphen,
	// period, and plus characters.
	regexpHasScheme = regexp.MustCompile(`^[-+.a-zA-Z\d]*(://|:$)`)

	regexpIsUmask = regexp.MustCompile(`^[0-7]{3,4}$`)
)

const (
	policyTwoFactor = "two_factor"
)

const (
	addressQueryParamUmask = "umask"
	addressQueryParamPath  = "path"
)

const (
	blockCERTIFICATE = "CERTIFICATE"
)

// Authorization Schemes.
const (
	SchemeBasic  = "basic"
	SchemeBearer = "bearer"
)

// Authz values.
const (
	AuthzEndpointNameLegacy      = "legacy"
	AuthzEndpointNameAuthRequest = "auth-request"
	AuthzEndpointNameExtAuthz    = "ext-authz"
	AuthzEndpointNameForwardAuth = "forward-auth"

	AuthzImplementationLegacy      = "Legacy"
	AuthzImplementationAuthRequest = "AuthRequest"
	AuthzImplementationExtAuthz    = "ExtAuthz"
	AuthzImplementationForwardAuth = "ForwardAuth"

	AuthzStrategyHeaderCookieSession                 = "CookieSession"
	AuthzStrategyHeaderAuthorization                 = "HeaderAuthorization"
	AuthzStrategyHeaderProxyAuthorization            = "HeaderProxyAuthorization"
	AuthzStrategyHeaderAuthRequestProxyAuthorization = "HeaderAuthRequestProxyAuthorization"
	AuthzStrategyHeaderLegacy                        = "HeaderLegacy"
)

const (
	ldapGroupSearchModeFilter = "filter"
)

const (
	ldapAttrDistinguishedName = "distinguishedName"
	ldapAttrMail              = "mail"
	ldapAttrUserID            = "uid"
	ldapAttrSAMAccountName    = "sAMAccountName"
	ldapAttrDisplayName       = "displayName"
	ldapAttrSurname           = "sn"
	ldapAttrGivenName         = "givenName"
	ldapAttrMiddleName        = "middleName"
	ldapAttrDescription       = "description"
	ldapAttrCommonName        = "cn"
	ldapAttrMemberOf          = "memberOf"
	ldapAttrGroupMember       = "member"
)

// Address Schemes.
const (
	AddressSchemeTCP            = "tcp"
	AddressSchemeTCP4           = "tcp4"
	AddressSchemeTCP6           = "tcp6"
	AddressSchemeUDP            = "udp"
	AddressSchemeUDP4           = "udp4"
	AddressSchemeUDP6           = "udp6"
	AddressSchemeUnix           = "unix"
	AddressSchemeLDAP           = "ldap"
	AddressSchemeLDAPS          = "ldaps"
	AddressSchemeLDAPI          = "ldapi"
	AddressSchemeSMTP           = "smtp"
	AddressSchemeSUBMISSION     = "submission"
	AddressSchemeSUBMISSIONS    = "submissions"
	AddressSchemeFileDescriptor = "fd"
)
