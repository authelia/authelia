package authentication

import (
	"errors"
	"time"

	"golang.org/x/text/encoding/unicode"
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
	ldapBaseObjectFilter = "(objectClass=*)"
)

const (
	ldapPlaceholderInput                             = "{input}"
	ldapPlaceholderDistinguishedName                 = "{dn}"
	ldapPlaceholderMemberOfDistinguishedName         = "{memberof:dn}"
	ldapPlaceholderMemberOfRelativeDistinguishedName = "{memberof:rdn}"
	ldapPlaceholderUsername                          = "{username}"
	ldapPlaceholderDateTimeGeneralized               = "{date-time:generalized}"
	ldapPlaceholderDateTimeMicrosoftNTTimeEpoch      = "{date-time:microsoft-nt}"
	ldapPlaceholderDateTimeUnixEpoch                 = "{date-time:unix}"
	ldapPlaceholderDistinguishedNameAttribute        = "{distinguished_name_attribute}"
	ldapPlaceholderUsernameAttribute                 = "{username_attribute}"
	ldapPlaceholderDisplayNameAttribute              = "{display_name_attribute}"
	ldapPlaceholderMailAttribute                     = "{mail_attribute}"
	ldapPlaceholderMemberOfAttribute                 = "{member_of_attribute}"
)

const (
	ldapGeneralizedTimeDateTimeFormat = "20060102150405.0Z"
)

const (
	none = "none"
)

const (
	hashArgon2    = "argon2"
	hashSHA2Crypt = "sha2crypt"
	hashPBKDF2    = "pbkdf2"
	hashSCrypt    = "scrypt"
	hashBCrypt    = "bcrypt"
)

var (
	// ErrUserNotFound indicates the user wasn't found in the authentication backend.
	ErrUserNotFound = errors.New("user not found")
	ErrUserDisabled = errors.New("the user is disabled")

	// ErrNoContent is returned when the file is empty.
	ErrNoContent = errors.New("no file content")

	ErrInvalidUsername      = errors.New("invalid username")
	ErrInvalidEmail         = errors.New("invalid email address")
	ErrInvalidPassword      = errors.New("invalid password")
	ErrCreatingUser         = errors.New("error creating user")
	ErrUserExists           = errors.New("user already exists")
	ErrUpdatingUserPassword = errors.New("error updating password")
	ErrUpdatingUser         = errors.New("error updating user")
	ErrHashError            = errors.New("error hashing password")

	ErrListingUser = errors.New("error while getting the user list")

	ErrDeletingUser = errors.New("error deleting user")
)

const fileAuthenticationMode = 0600

// OWASP recommends to escape some special characters.
// https://github.com/OWASP/CheatSheetSeries/blob/master/cheatsheets/LDAP_Injection_Prevention_Cheat_Sheet.md
const specialLDAPRunes = ",#+<>;\"="

var (
	encodingUTF16LittleEndian = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
)

const (
	contextTimeout time.Duration = 20 * time.Second
)
