package authentication

import (
	"errors"

	"golang.org/x/text/encoding/unicode"
)

const (
	ldapSupportedLDAPVersionAttribute    = "supportedLDAPVersion"
	ldapSupportedFeaturesAttribute       = "supportedFeatures"
	ldapSupportedSASLMechanismsAttribute = "supportedSASLMechanisms"
)

const (
	ldapVendorNameAttribute          = "vendorName"
	ldapVendorVersionAttribute       = "vendorVersion"
	ldapDomainFunctionalityAttribute = "domainFunctionality"
	ldapForestFunctionalityAttribute = "forestFunctionality"
	ldapObjectClassAttribute         = "objectClass"

	ldapVendorNameMicrosoftCorporation = "Microsoft Corporation"
	ldapVendorNameOpenLDAP             = "OpenLDAP"
	ldapVendorOpenLDAPObjectClass      = "OpenLDAProotDSE"
)

const (
	ldapSupportedExtensionAttribute = "supportedExtension"

	// LDAP Extension OID: Password Modify Extended Operation.
	//
	// See the linked documents for more information.
	//
	// RFC3062: https://datatracker.ietf.org/doc/html/rfc3062
	//
	// OID Reference: http://oidref.com/1.3.6.1.4.1.4203.1.11.1
	ldapOIDExtensionPwdModify = "1.3.6.1.4.1.4203.1.11.1"

	// LDAP Extension OID: Transport Layer Security.
	//
	// See the linked documents for more information.
	//
	// RFC2830: https://datatracker.ietf.org/doc/html/rfc2830
	//
	// OID Reference: https://oidref.com/1.3.6.1.4.1.1466.20037
	ldapOIDExtensionTLS = "1.3.6.1.4.1.1466.20037"

	// LDAP Extension OID: Who Am I?
	//
	// See the linked documents for more information.
	//
	// RFC4532: https://datatracker.ietf.org/doc/html/rfc4532
	//
	// OID Reference: https://oidref.com/1.3.6.1.4.1.4203.1.11.3
	ldapOIDExtensionWhoAmI = "1.3.6.1.4.1.4203.1.11.3"
)

const (
	ldapSupportedControlAttribute = "supportedControl"

	// LDAP Control OID: Microsoft Password Policy Hints.
	//
	// See the linked documents for more information.
	//
	// MS ADTS: https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/4add7bce-e502-4e0f-9d69-1a3f153713e2
	//
	// OID Reference: https://oidref.com/1.2.840.113556.1.4.2239
	ldapOIDControlMsftServerPolicyHints = "1.2.840.113556.1.4.2239"

	// LDAP Control OID: Microsoft Password Policy Hints (deprecated).
	//
	// See the linked documents for more information.
	//
	// MS ADTS: https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-adts/49751d58-8115-4277-8faf-64c83a5f658f
	//
	// OID Reference: https://oidref.com/1.2.840.113556.1.4.2066
	ldapOIDControlMsftServerPolicyHintsDeprecated = "1.2.840.113556.1.4.2066"
)

const (
	ldapAttributeUnicodePwd   = "unicodePwd"
	ldapAttributeUserPassword = "userPassword"

	ldapAttrMail        = "mail"
	ldapAttrCommonName  = "cn"
	ldapAttrMemberOf    = "memberOf"
	ldapAttrObjectClass = "objectClass"
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
	hashScrypt    = "scrypt"
	hashBcrypt    = "bcrypt"
)

var (
	// ErrUserNotFound indicates the user wasn't found in the authentication backend.
	ErrUserNotFound = errors.New("user not found")

	ErrGroupNotFound = errors.New("group not found")

	ErrGroupExists = errors.New("group already exists")

	// ErrWatcherNoContent is returned when the file is empty.
	ErrWatcherNoContent = errors.New("no file content")

	// ErrWatcherCooldown is returned when the file watcher is on cooldown.
	ErrWatcherCooldown = errors.New("watcher on cooldown")

	ErrOperationFailed = errors.New("operation failed")

	// ErrIncorrectPassword is returned when the password provided is incorrect.
	ErrIncorrectPassword = errors.New("incorrect password")

	ErrPasswordWeak = errors.New("your supplied password does not meet the password policy requirements")

	// ErrPasswordReuse is returned when the new password is the same as the existing password.
	ErrPasswordReuse = errors.New("you cannot reuse your old password")

	// ErrEmptyInput is returned when an empty string or nil value is used to set a value.
	ErrEmptyInput = errors.New("empty input is not valid")

	ErrPasswordEmpty = errors.New("your password cannot be blank")

	ErrAuthenticationFailed = errors.New("authentication failed")

	ErrLDAPHealthCheckFailedEntryCount = errors.New("incorrect number entries found when performing RootDSE search")
)
var (
	ErrUsernameIsRequired   = errors.New("username is required")
	ErrFamilyNameIsRequired = errors.New("family name is required")
)

const fileAuthenticationMode = 0600

var (
	encodingUTF16LittleEndian = unicode.UTF16(unicode.LittleEndian, unicode.IgnoreBOM)
)

const (
	ValueTypeString  = "string"
	ValueTypeInteger = "integer"
	ValueTypeBoolean = "boolean"
)

type errReload struct {
	err      error
	critical bool
}

func (e *errReload) Error() string {
	return e.err.Error()
}

func (e *errReload) Unwrap() error {
	return e.err
}

func (e *errReload) WatcherReloadErrorCritical() bool {
	return e.critical
}

// LDAP Boolean Syntax Values.
//
// See the linked documents for more information.
//
// RFC4517 Section 3.3.3: https://datatracker.ietf.org/doc/html/rfc4517#section-3.3.3
//
// Syntax OID Reference: https://oidref.com/1.3.6.1.4.1.1466.115.121.1.7
const (
	BooleanValueTrue  = "TRUE"
	BooleanValueFalse = "FALSE"
)

// User management attribute names for update masks.
const (
	AttributeUsername       = "username"
	AttributePassword       = "password"
	AttributeDisplayName    = "display_name"
	AttributeGivenName      = "given_name"
	AttributeFamilyName     = "family_name"
	AttributeMiddleName     = "middle_name"
	AttributeNickname       = "nickname"
	AttributeGender         = "gender"
	AttributeBirthdate      = "birthdate"
	AttributeWebsite        = "website"
	AttributeProfile        = "profile"
	AttributePicture        = "picture"
	AttributeZoneInfo       = "zoneinfo"
	AttributeLocale         = "locale"
	AttributePhoneNumber    = "phone_number"
	AttributePhoneExtension = "phone_extension"
	AttributeMail           = "mail"
	AttributeGroups         = "groups"
	AttributeAddress        = "address"
	AttributeExtra          = "extra"
)

// Address subfield attribute names.
const (
	AttributeAddressStreetAddress = "street_address"
	AttributeAddressLocality      = "locality"
	AttributeAddressRegion        = "region"
	AttributeAddressPostalCode    = "postal_code"
	AttributeAddressCountry       = "country"
)

// Attribute prefixes for composite attributes.
const (
	PrefixAttributeExtra   = "extra."
	PrefixAttributeAddress = "address."
)

var attributeMetadataMap = map[string]UserManagementAttributeMetadata{
	AttributeUsername:             {Type: Text, Multiple: false},
	AttributeGroups:               {Type: Groups, Multiple: true},
	AttributePassword:             {Type: Password, Multiple: false},
	AttributeDisplayName:          {Type: Text, Multiple: false},
	AttributeFamilyName:           {Type: Text, Multiple: false},
	AttributeGivenName:            {Type: Text, Multiple: false},
	AttributeMiddleName:           {Type: Text, Multiple: false},
	AttributeNickname:             {Type: Text, Multiple: false},
	AttributeGender:               {Type: Text, Multiple: false},
	AttributeBirthdate:            {Type: Date, Multiple: false},
	AttributeWebsite:              {Type: Url, Multiple: false},
	AttributeProfile:              {Type: Url, Multiple: false},
	AttributePicture:              {Type: Url, Multiple: false},
	AttributeZoneInfo:             {Type: Text, Multiple: false},
	AttributeLocale:               {Type: Text, Multiple: false},
	AttributePhoneNumber:          {Type: Telephone, Multiple: false},
	AttributePhoneExtension:       {Type: Text, Multiple: false},
	AttributeAddressStreetAddress: {Type: Text, Multiple: false},
	AttributeAddressLocality:      {Type: Text, Multiple: false},
	AttributeAddressRegion:        {Type: Text, Multiple: false},
	AttributeAddressPostalCode:    {Type: Text, Multiple: false},
	AttributeAddressCountry:       {Type: Text, Multiple: false},
	AttributeMail:                 {Type: Email, Multiple: false},
}
