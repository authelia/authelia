package schema

import (
	"time"
)

const argon2id = "argon2id"

// ProfileRefreshDisabled represents a value for refresh_interval that disables the check entirely.
const ProfileRefreshDisabled = "disable"

// ProfileRefreshAlways represents a value for refresh_interval that's the same as 0ms.
const ProfileRefreshAlways = "always"

// RefreshIntervalDefault represents the default value of refresh_interval.
const RefreshIntervalDefault = "5m"

// RefreshIntervalAlways represents the duration value refresh interval should have if set to always.
const RefreshIntervalAlways = 0 * time.Millisecond

// LDAPImplementationCustom is the string for the custom LDAP implementation.
const LDAPImplementationCustom = "custom"

// LDAPImplementationActiveDirectory is the string for the Active Directory LDAP implementation.
const LDAPImplementationActiveDirectory = "activedirectory"

// TOTP Algorithm.
const (
	TOTPAlgorithmSHA1   = "SHA1"
	TOTPAlgorithmSHA256 = "SHA256"
	TOTPAlgorithmSHA512 = "SHA512"
)

var (
	// TOTPPossibleAlgorithms is a list of valid TOTP Algorithms.
	TOTPPossibleAlgorithms = []string{TOTPAlgorithmSHA1, TOTPAlgorithmSHA256, TOTPAlgorithmSHA512}
)
