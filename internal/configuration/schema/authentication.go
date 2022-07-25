package schema

import (
	"net/url"
	"time"
)

// LDAPAuthenticationBackendConfiguration represents the configuration related to LDAP server.
type LDAPAuthenticationBackendConfiguration struct {
	Implementation string        `koanf:"implementation"`
	URL            string        `koanf:"url"`
	Timeout        time.Duration `koanf:"timeout"`
	StartTLS       bool          `koanf:"start_tls"`
	TLS            *TLSConfig    `koanf:"tls"`

	BaseDN string `koanf:"base_dn"`

	AdditionalUsersDN string `koanf:"additional_users_dn"`
	UsersFilter       string `koanf:"users_filter"`

	AdditionalGroupsDN string `koanf:"additional_groups_dn"`
	GroupsFilter       string `koanf:"groups_filter"`

	GroupNameAttribute   string `koanf:"group_name_attribute"`
	UsernameAttribute    string `koanf:"username_attribute"`
	MailAttribute        string `koanf:"mail_attribute"`
	DisplayNameAttribute string `koanf:"display_name_attribute"`

	PermitReferrals           bool `koanf:"permit_referrals"`
	PermitUnauthenticatedBind bool `koanf:"permit_unauthenticated_bind"`

	User     string `koanf:"user"`
	Password string `koanf:"password"`
}

// FileAuthenticationBackendConfig represents the configuration related to file-based backend.
type FileAuthenticationBackendConfig struct {
	Path     string          `koanf:"path"`
	Password *PasswordConfig `koanf:"password"`
}

// PasswordConfig represents the configuration related to password hashing.
type PasswordConfig struct {
	Algorithm string `koanf:"algorithm"`

	Argon2    Argon2PasswordConfig    `koanf:"argon2"`
	SHA2Crypt SHA2CryptPasswordConfig `koanf:"sha2crypt"`
	PBKDF2    PBKDF2PasswordConfig    `koanf:"pbkdf2"`
	BCrypt    BCryptPasswordConfig    `koanf:"bcrypt"`
	SCrypt    SCryptPasswordConfig    `koanf:"scrypt"`

	Iterations  int `koanf:"iterations"`
	KeyLength   int `koanf:"key_length"`
	SaltLength  int `koanf:"salt_length"`
	Memory      int `koanf:"memory"`
	Parallelism int `koanf:"parallelism"`
}

type Argon2PasswordConfig struct {
	Variant     string `koanf:"variant"`
	Iterations  uint32 `koanf:"iterations"`
	Memory      uint32 `koanf:"memory"`
	Parallelism uint32 `koanf:"parallelism"`
	KeyLength   uint32 `koanf:"key_length"`
	SaltLength  uint32 `koanf:"salt_length"`
}

type SHA2CryptPasswordConfig struct {
	Variant    string `koanf:"variant"`
	Rounds     uint32 `koanf:"rounds"`
	SaltLength uint32 `koanf:"salt_length"`
}

type PBKDF2PasswordConfig struct {
	Variant    string `koanf:"variant"`
	Iterations int    `koanf:"iterations"`
	KeyLength  uint32 `koanf:"key_length"`
	SaltLength uint32 `koanf:"salt_length"`
}

type BCryptPasswordConfig struct {
	Variant string `koanf:"variant"`
	Rounds  int    `koanf:"rounds"`
}

type SCryptPasswordConfig struct {
	Rounds      uint32 `koanf:"rounds"`
	BlockSize   uint32 `koanf:"block_size"`
	Parallelism uint8  `koanf:"parallelism"`
	KeyLength   uint32 `koanf:"key_length"`
	SaltLength  uint32 `koanf:"salt_length"`
}

// AuthenticationBackendConfiguration represents the configuration related to the authentication backend.
type AuthenticationBackendConfiguration struct {
	LDAP *LDAPAuthenticationBackendConfiguration `koanf:"ldap"`
	File *FileAuthenticationBackendConfig        `koanf:"file"`

	PasswordReset PasswordResetAuthenticationBackendConfiguration `koanf:"password_reset"`

	RefreshInterval string `koanf:"refresh_interval"`
}

// PasswordResetAuthenticationBackendConfiguration represents the configuration related to password reset functionality.
type PasswordResetAuthenticationBackendConfiguration struct {
	Disable   bool    `koanf:"disable"`
	CustomURL url.URL `koanf:"custom_url"`
}

// DefaultPasswordConfiguration represents the default configuration related to Argon2id hashing.
var DefaultPasswordConfiguration = PasswordConfig{
	Iterations:  3,
	KeyLength:   32,
	SaltLength:  16,
	Algorithm:   argon2id,
	Memory:      64,
	Parallelism: 4,
}

// DefaultCIPasswordConfiguration represents the default configuration related to Argon2id hashing for CI.
var DefaultCIPasswordConfiguration = PasswordConfig{
	Algorithm: "argon2",
	Argon2: Argon2PasswordConfig{
		Iterations:  3,
		Memory:      64,
		Parallelism: 4,
		KeyLength:   32,
		SaltLength:  16,
	},
}

// DefaultPasswordSHA512Configuration represents the default configuration related to SHA512 hashing.
var DefaultPasswordSHA512Configuration = PasswordConfig{
	Iterations: 50000,
	SaltLength: 16,
	Algorithm:  "sha512",
}

// DefaultLDAPAuthenticationBackendConfiguration represents the default LDAP config.
var DefaultLDAPAuthenticationBackendConfiguration = LDAPAuthenticationBackendConfiguration{
	Implementation:       LDAPImplementationCustom,
	UsernameAttribute:    "uid",
	MailAttribute:        "mail",
	DisplayNameAttribute: "displayName",
	GroupNameAttribute:   "cn",
	Timeout:              time.Second * 5,
	TLS: &TLSConfig{
		MinimumVersion: "TLS1.2",
	},
}

// DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration represents the default LDAP config for the MSAD Implementation.
var DefaultLDAPAuthenticationBackendImplementationActiveDirectoryConfiguration = LDAPAuthenticationBackendConfiguration{
	UsersFilter:          "(&(|({username_attribute}={input})({mail_attribute}={input}))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0)))",
	UsernameAttribute:    "sAMAccountName",
	MailAttribute:        "mail",
	DisplayNameAttribute: "displayName",
	GroupsFilter:         "(&(member={dn})(objectClass=group))",
	GroupNameAttribute:   "cn",
}
