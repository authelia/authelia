package schema

import (
	"regexp"
	"time"
)

const (
	argon2   = "argon2"
	argon2id = "argon2id"
	sha512   = "sha512"
	sha256   = "sha256"
)

// ProfileRefreshDisabled represents a value for refresh_interval that disables the check entirely.
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
