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
	OneFactor

	// TwoFactor if the user has passed two factors.
	TwoFactor
)

const (
	ldapSupportedExtensionAttribute = "supportedExtension"

	// LDAP Extension OID: Password Modify Extended Operation.
	//
	// RFC3062: https://datatracker.ietf.org/doc/html/rfc3062
	//
	// OID Reference: http://oidref.com/1.3.6.1.4.1.4203.1.11.1
	//
	// See the linked documents for more information.
	ldapOIDExtensionPwdModifyExOp = "1.3.6.1.4.1.4203.1.11.1"

	// LDAP Extension OID: Transport Layer Security.
	//
	// RFC2830: https://datatracker.ietf.org/doc/html/rfc2830
	//
	// OID Reference: https://oidref.com/1.3.6.1.4.1.1466.20037
	//
	// See the linked documents for more information.
	ldapOIDExtensionTLS = "1.3.6.1.4.1.1466.20037"
)

const (
	ldapSupportedControlAttribute = "supportedControl"

	// LDAP Control OID: Microsoft Password Policy Hints.
	//
	// MS ADTS: https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/4add7bce-e502-4e0f-9d69-1a3f153713e2
	//
	// OID Reference: https://oidref.com/1.2.840.113556.1.4.2239
	//
	// See the linked documents for more information.
	ldapOIDControlMsftServerPolicyHints = "1.2.840.113556.1.4.2239"

	// LDAP Control OID: Microsoft Password Policy Hints (deprecated).
	//
	// MS ADTS: https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/49751d58-8115-4277-8faf-64c83a5f658f
	//
	// OID Reference: https://oidref.com/1.2.840.113556.1.4.2066
	//
	// See the linked documents for more information.
	ldapOIDControlMsftServerPolicyHintsDeprecated = "1.2.840.113556.1.4.2066"
)

const (
	ldapAttributeUnicodePwd   = "unicodePwd"
	ldapAttributeUserPassword = "userPassword"
)

const (
	ldapPlaceholderInput             = "{input}"
	ldapPlaceholderDistinguishedName = "{dn}"
	ldapPlaceholderUsername          = "{username}"
)

const (
	none = "none"
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
