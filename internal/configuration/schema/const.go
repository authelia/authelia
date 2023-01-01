package schema

import (
	"errors"
	"regexp"
	"time"
)

const (
	argon2   = "argon2"
	argon2id = "argon2id"
	sha512   = "sha512"
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

// ProfileRefreshDisabled represents a Value for refresh_interval that disables the check entirely.
const ProfileRefreshDisabled = "disable"

const (
	// ProfileRefreshAlways represents a value for refresh_interval that's the same as 0ms.
	ProfileRefreshAlways = "always"

	// RefreshIntervalDefault represents the default value of refresh_interval.
	RefreshIntervalDefault = "5m"

	// RefreshIntervalAlways represents the duration value refresh interval should have if set to always.
	RefreshIntervalAlways = 0 * time.Millisecond
)

const (
	// LDAPImplementationCustom is the string for the custom LDAP implementation.
	LDAPImplementationCustom = "custom"

	// LDAPImplementationActiveDirectory is the string for the Active Directory LDAP implementation.
	LDAPImplementationActiveDirectory = "activedirectory"

	// LDAPImplementationFreeIPA is the string for the FreeIPA LDAP implementation.
	LDAPImplementationFreeIPA = "freeipa"

	// LDAPImplementationLLDAP is the string for the lldap LDAP implementation.
	LDAPImplementationLLDAP = "lldap"

	// LDAPImplementationGLAuth is the string for the GLAuth LDAP implementation.
	LDAPImplementationGLAuth = "glauth"
)

const (
	LDAPUserAuthenticationMethodBind   = "bind"
	LDAPUserAuthenticationMethodNTHash = "nthash"
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

// regexpHasScheme checks if a string has a scheme. Valid characters for schemes include alphanumeric, hyphen,
// period, and plus characters.
var regexpHasScheme = regexp.MustCompile(`^[-+.a-zA-Z\d]+://`)

const (
	blockCERTIFICATE   = "CERTIFICATE"
	blockRSAPRIVATEKEY = "RSA PRIVATE KEY"
)

const (
	ldapAttrMail        = "mail"
	ldapAttrUserID      = "uid"
	ldapAttrDisplayName = "displayName"
	ldapAttrDescription = "description"
	ldapAttrCommonName  = "cn"
)
