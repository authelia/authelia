package schema

const argon2id = "argon2id"

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
