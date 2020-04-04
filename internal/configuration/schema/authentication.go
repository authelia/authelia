package schema

// LDAPAuthenticationBackendConfiguration represents the configuration related to LDAP server.
type LDAPAuthenticationBackendConfiguration struct {
	URL                string `mapstructure:"url"`
	SkipVerify         bool   `mapstructure:"skip_verify"`
	BaseDN             string `mapstructure:"base_dn"`
	AdditionalUsersDN  string `mapstructure:"additional_users_dn"`
	UsersFilter        string `mapstructure:"users_filter"`
	AdditionalGroupsDN string `mapstructure:"additional_groups_dn"`
	GroupsFilter       string `mapstructure:"groups_filter"`
	GroupNameAttribute string `mapstructure:"group_name_attribute"`
	UsernameAttribute  string `mapstructure:"username_attribute"`
	MailAttribute      string `mapstructure:"mail_attribute"`
	User               string `mapstructure:"user"`
	Password           string `mapstructure:"password"`
}

// FileAuthenticationBackendConfiguration represents the configuration related to file-based backend
type FileAuthenticationBackendConfiguration struct {
	Path            string                        `mapstructure:"path"`
	PasswordHashing *PasswordHashingConfiguration `mapstructure:"password"`
}

type PasswordHashingConfiguration struct {
	Iterations  int    `mapstructure:"iterations"`
	KeyLength   int    `mapstructure:"key_length"`
	SaltLength  int    `mapstructure:"salt_length"`
	Algorithm   string `mapstrucutre:"algorithm"`
	Memory      int    `mapstructure:"memory"`
	Parallelism int    `mapstructure:"parallelism"`
}

// Default Argon2id Configuration
var DefaultPasswordOptionsConfiguration = PasswordHashingConfiguration{
	Iterations:  1,
	KeyLength:   32,
	SaltLength:  16,
	Algorithm:   "argon2id",
	Memory:      1024,
	Parallelism: 8,
}

// Default Argon2id Configuration for CI testing when calling HashPassword()
var DefaultCIPasswordOptionsConfiguration = PasswordHashingConfiguration{
	Iterations:  1,
	KeyLength:   32,
	SaltLength:  16,
	Algorithm:   "argon2id",
	Memory:      128,
	Parallelism: 8,
}

// Default SHA512 Cofniguration
var DefaultPasswordOptionsSHA512Configuration = PasswordHashingConfiguration{
	Iterations: 50000,
	SaltLength: 16,
	Algorithm:  "sha512",
}

// AuthenticationBackendConfiguration represents the configuration related to the authentication backend.
type AuthenticationBackendConfiguration struct {
	DisableResetPassword bool                                    `mapstructure:"disable_reset_password"`
	Ldap                 *LDAPAuthenticationBackendConfiguration `mapstructure:"ldap"`
	File                 *FileAuthenticationBackendConfiguration `mapstructure:"file"`
}
