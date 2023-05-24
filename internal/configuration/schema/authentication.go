package schema

import (
	"crypto/tls"
	"net/url"
	"time"
)

// AuthenticationBackend represents the configuration related to the authentication backend.
type AuthenticationBackend struct {
	PasswordReset PasswordResetAuthenticationBackend `koanf:"password_reset"`

	RefreshInterval string `koanf:"refresh_interval"`

	File *FileAuthenticationBackend `koanf:"file"`
	LDAP *LDAPAuthenticationBackend `koanf:"ldap"`
}

// PasswordResetAuthenticationBackend represents the configuration related to password reset functionality.
type PasswordResetAuthenticationBackend struct {
	Disable   bool    `koanf:"disable"`
	CustomURL url.URL `koanf:"custom_url"`
}

// FileAuthenticationBackend represents the configuration related to file-based backend.
type FileAuthenticationBackend struct {
	Path     string   `koanf:"path"`
	Watch    bool     `koanf:"watch"`
	Password Password `koanf:"password"`

	Search FileSearchAuthenticationBackend `koanf:"search"`
}

// FileSearchAuthenticationBackend represents the configuration related to file-based backend searching.
type FileSearchAuthenticationBackend struct {
	Email           bool `koanf:"email"`
	CaseInsensitive bool `koanf:"case_insensitive"`
}

// Password represents the configuration related to password hashing.
type Password struct {
	Algorithm string `koanf:"algorithm"`

	Argon2    Argon2Password    `koanf:"argon2"`
	SHA2Crypt SHA2CryptPassword `koanf:"sha2crypt"`
	PBKDF2    PBKDF2Password    `koanf:"pbkdf2"`
	BCrypt    BCryptPassword    `koanf:"bcrypt"`
	SCrypt    SCryptPassword    `koanf:"scrypt"`

	Iterations  int `koanf:"iterations"`
	Memory      int `koanf:"memory"`
	Parallelism int `koanf:"parallelism"`
	KeyLength   int `koanf:"key_length"`
	SaltLength  int `koanf:"salt_length"`
}

// Argon2Password represents the argon2 hashing settings.
type Argon2Password struct {
	Variant     string `koanf:"variant"`
	Iterations  int    `koanf:"iterations"`
	Memory      int    `koanf:"memory"`
	Parallelism int    `koanf:"parallelism"`
	KeyLength   int    `koanf:"key_length"`
	SaltLength  int    `koanf:"salt_length"`
}

// SHA2CryptPassword represents the sha2crypt hashing settings.
type SHA2CryptPassword struct {
	Variant    string `koanf:"variant"`
	Iterations int    `koanf:"iterations"`
	SaltLength int    `koanf:"salt_length"`
}

// PBKDF2Password represents the PBKDF2 hashing settings.
type PBKDF2Password struct {
	Variant    string `koanf:"variant"`
	Iterations int    `koanf:"iterations"`
	SaltLength int    `koanf:"salt_length"`
}

// BCryptPassword represents the bcrypt hashing settings.
type BCryptPassword struct {
	Variant string `koanf:"variant"`
	Cost    int    `koanf:"cost"`
}

// SCryptPassword represents the scrypt hashing settings.
type SCryptPassword struct {
	Iterations  int `koanf:"iterations"`
	BlockSize   int `koanf:"block_size"`
	Parallelism int `koanf:"parallelism"`
	KeyLength   int `koanf:"key_length"`
	SaltLength  int `koanf:"salt_length"`
}

// LDAPAuthenticationBackend represents the configuration related to LDAP server.
type LDAPAuthenticationBackend struct {
	Address        *AddressLDAP  `koanf:"address"`
	Implementation string        `koanf:"implementation"`
	Timeout        time.Duration `koanf:"timeout"`
	StartTLS       bool          `koanf:"start_tls"`
	TLS            *TLSConfig    `koanf:"tls"`

	BaseDN string `koanf:"base_dn"`

	AdditionalUsersDN string `koanf:"additional_users_dn"`
	UsersFilter       string `koanf:"users_filter"`
	UsersResetFilter  string `koanf:"users_reset_filter"`

	AdditionalGroupsDN string `koanf:"additional_groups_dn"`
	GroupsFilter       string `koanf:"groups_filter"`
	GroupSearchMode    string `koanf:"group_search_mode"`

	Attributes LDAPAuthenticationAttributes `koanf:"attributes"`

	PermitReferrals               bool `koanf:"permit_referrals"`
	PermitUnauthenticatedBind     bool `koanf:"permit_unauthenticated_bind"`
	PermitFeatureDetectionFailure bool `koanf:"permit_feature_detection_failure"`

	User     string `koanf:"user"`
	Password string `koanf:"password"`
}

// LDAPAuthenticationAttributes represents the configuration related to LDAP server attributes.
type LDAPAuthenticationAttributes struct {
	DistinguishedName string `koanf:"distinguished_name"`
	Username          string `koanf:"username"`
	DisplayName       string `koanf:"display_name"`
	Mail              string `koanf:"mail"`
	MemberOf          string `koanf:"member_of"`
	GroupName         string `koanf:"group_name"`
}

// DefaultPasswordConfig represents the default configuration related to Argon2id hashing.
var DefaultPasswordConfig = Password{
	Algorithm: argon2,
	Argon2: Argon2Password{
		Variant:     argon2id,
		Iterations:  3,
		Memory:      64 * 1024,
		Parallelism: 4,
		KeyLength:   32,
		SaltLength:  16,
	},
	SHA2Crypt: SHA2CryptPassword{
		Variant:    sha512,
		Iterations: 50000,
		SaltLength: 16,
	},
	PBKDF2: PBKDF2Password{
		Variant:    sha512,
		Iterations: 310000,
		SaltLength: 16,
	},
	BCrypt: BCryptPassword{
		Variant: "standard",
		Cost:    12,
	},
	SCrypt: SCryptPassword{
		Iterations:  16,
		BlockSize:   8,
		Parallelism: 1,
		KeyLength:   32,
		SaltLength:  16,
	},
}

// DefaultCIPasswordConfig represents the default configuration related to Argon2id hashing for CI.
var DefaultCIPasswordConfig = Password{
	Algorithm: argon2,
	Argon2: Argon2Password{
		Iterations:  3,
		Memory:      64,
		Parallelism: 4,
		KeyLength:   32,
		SaltLength:  16,
	},
	SHA2Crypt: SHA2CryptPassword{
		Variant:    sha512,
		Iterations: 50000,
		SaltLength: 16,
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationCustom represents the default LDAP config.
var DefaultLDAPAuthenticationBackendConfigurationImplementationCustom = LDAPAuthenticationBackend{
	GroupSearchMode: ldapGroupSearchModeFilter,
	Attributes: LDAPAuthenticationAttributes{
		Username:    ldapAttrUserID,
		DisplayName: ldapAttrDisplayName,
		Mail:        ldapAttrMail,
		GroupName:   ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLSConfig{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory represents the default LDAP config for the LDAPImplementationActiveDirectory Implementation.
var DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory = LDAPAuthenticationBackend{
	UsersFilter:     "(&(|({username_attribute}={input})({mail_attribute}={input}))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))(|(!(accountExpires=*))(accountExpires=0)(accountExpires>={date-time:microsoft-nt})))",
	GroupsFilter:    "(&(member={dn})(|(sAMAccountType=268435456)(sAMAccountType=536870912)))",
	GroupSearchMode: ldapGroupSearchModeFilter,
	Attributes: LDAPAuthenticationAttributes{
		DistinguishedName: ldapAttrDistinguishedName,
		Username:          ldapAttrSAMAccountName,
		DisplayName:       ldapAttrDisplayName,
		Mail:              ldapAttrMail,
		MemberOf:          ldapAttrMemberOf,
		GroupName:         ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLSConfig{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis represents the default LDAP config for the LDAPImplementationRFC2307bis Implementation.
var DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis = LDAPAuthenticationBackend{
	UsersFilter:     "(&(|({username_attribute}={input})({mail_attribute}={input}))(|(objectClass=inetOrgPerson)(objectClass=organizationalPerson)))",
	GroupsFilter:    "(&(|(member={dn})(uniqueMember={dn}))(|(objectClass=groupOfNames)(objectClass=groupOfUniqueNames)(objectClass=groupOfMembers)))",
	GroupSearchMode: ldapGroupSearchModeFilter,
	Attributes: LDAPAuthenticationAttributes{
		Username:    ldapAttrUserID,
		DisplayName: ldapAttrDisplayName,
		Mail:        ldapAttrMail,
		MemberOf:    ldapAttrMemberOf,
		GroupName:   ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLSConfig{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA represents the default LDAP config for the LDAPImplementationFreeIPA Implementation.
var DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA = LDAPAuthenticationBackend{
	UsersFilter:     "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person)(!(nsAccountLock=TRUE))(krbPasswordExpiration>={date-time:generalized})(|(!(krbPrincipalExpiration=*))(krbPrincipalExpiration>={date-time:generalized})))",
	GroupsFilter:    "(&(member={dn})(objectClass=groupOfNames))",
	GroupSearchMode: ldapGroupSearchModeFilter,
	Attributes: LDAPAuthenticationAttributes{
		Username:    ldapAttrUserID,
		DisplayName: ldapAttrDisplayName,
		Mail:        ldapAttrMail,
		MemberOf:    ldapAttrMemberOf,
		GroupName:   ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLSConfig{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP represents the default LDAP config for the LDAPImplementationLLDAP Implementation.
var DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP = LDAPAuthenticationBackend{
	AdditionalUsersDN:  "OU=people",
	AdditionalGroupsDN: "OU=groups",
	UsersFilter:        "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
	GroupsFilter:       "(&(member={dn})(objectClass=groupOfUniqueNames))",
	GroupSearchMode:    ldapGroupSearchModeFilter,
	Attributes: LDAPAuthenticationAttributes{
		Username:    ldapAttrUserID,
		DisplayName: ldapAttrCommonName,
		Mail:        ldapAttrMail,
		GroupName:   ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLSConfig{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth represents the default LDAP config for the LDAPImplementationGLAuth Implementation.
var DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth = LDAPAuthenticationBackend{
	UsersFilter:     "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=posixAccount)(!(accountStatus=inactive)))",
	GroupsFilter:    "(&(uniqueMember={dn})(objectClass=posixGroup))",
	GroupSearchMode: ldapGroupSearchModeFilter,
	Attributes: LDAPAuthenticationAttributes{
		Username:    ldapAttrCommonName,
		DisplayName: ldapAttrDescription,
		Mail:        ldapAttrMail,
		GroupName:   ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLSConfig{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}
