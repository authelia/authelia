package authentication

import (
	"errors"
)

// Level is the type representing a level of authentication.
type Level int

const (
	// NotAuthenticated if the user is not authenticated yet.
	NotAuthenticated Level = iota
	// OneFactor if the user has passed first factor only.
	OneFactor Level = iota
	// TwoFactor if the user has passed two factors.
	TwoFactor Level = iota
)

const (
	ldapSupportedExtensionAttribute = "supportedExtension"
	ldapOIDPasswdModifyExtension    = "1.3.6.1.4.1.4203.1.11.1" // http://oidref.com/1.3.6.1.4.1.4203.1.11.1
)

const (
	ldapPlaceholderInput             = "{input}"
	ldapPlaceholderDistinguishedName = "{dn}"
	ldapPlaceholderUsername          = "{username}"
)

// CryptAlgo the crypt representation of an algorithm used in the prefix of the hash.
type CryptAlgo string

const (
	// HashingAlgorithmArgon2id Argon2id hash identifier.
	HashingAlgorithmArgon2id CryptAlgo = argon2id
	// HashingAlgorithmSHA512 SHA512 hash identifier.
	HashingAlgorithmSHA512 CryptAlgo = "6"
)

// These are the default values from the upstream crypt module we use them to for GetInt
// and they need to be checked when updating github.com/simia-tech/crypt.
const (
	HashingDefaultArgon2idTime        = 1
	HashingDefaultArgon2idMemory      = 32 * 1024
	HashingDefaultArgon2idParallelism = 4
	HashingDefaultArgon2idKeyLength   = 32
	HashingDefaultSHA512Iterations    = 5000
)

// HashingPossibleSaltCharacters represents valid hashing runes.
var HashingPossibleSaltCharacters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789+/"

// ErrUserNotFound indicates the user wasn't found in the authentication backend.
var ErrUserNotFound = errors.New("user not found")

const argon2id = "argon2id"
const sha512 = "sha512"

const testPassword = "my;secure*password"

const fileAuthenticationMode = 0600

// OWASP recommends to escape some special characters.
// https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/LDAP_Injection_Prevention_Cheat_Sheet.md
const specialLDAPRunes = ",#+<>;\"="
