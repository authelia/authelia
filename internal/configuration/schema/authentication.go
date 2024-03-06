package schema

import (
	"crypto/tls"
	"net/url"
	"time"
)

// AuthenticationBackend represents the configuration related to the authentication backend.
type AuthenticationBackend struct {
	PasswordReset AuthenticationBackendPasswordReset `koanf:"password_reset" json:"password_reset" jsonschema:"title=Password Reset" jsonschema_description:"Allows configuration of the password reset behaviour."`

	RefreshInterval RefreshIntervalDuration `koanf:"refresh_interval" json:"refresh_interval" jsonschema:"default=5 minutes,title=Refresh Interval" jsonschema_description:"How frequently the user details are refreshed from the backend."`

	// The file authentication backend configuration.
	File *AuthenticationBackendFile `koanf:"file" json:"file" jsonschema:"title=File Backend" jsonschema_description:"The file authentication backend configuration."`
	LDAP *AuthenticationBackendLDAP `koanf:"ldap" json:"ldap" jsonschema:"title=LDAP Backend" jsonschema_description:"The LDAP authentication backend configuration."`
}

// AuthenticationBackendPasswordReset represents the configuration related to password reset functionality.
type AuthenticationBackendPasswordReset struct {
	Disable   bool    `koanf:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables the Password Reset option."`
	CustomURL url.URL `koanf:"custom_url" json:"custom_url" jsonschema:"title=Custom URL" jsonschema_description:"Disables the internal Password Reset option and instead redirects users to this specified URL."`
}

// AuthenticationBackendFile represents the configuration related to file-based backend.
type AuthenticationBackendFile struct {
	Path  string `koanf:"path" json:"path" jsonschema:"title=Path" jsonschema_description:"The file path to the user database."`
	Watch bool   `koanf:"watch" json:"watch" jsonschema:"default=false,title=Watch" jsonschema_description:"Enables watching the file for external changes and dynamically reloading the database."`

	Password AuthenticationBackendFilePassword `koanf:"password" json:"password" jsonschema:"title=Password Options" jsonschema_description:"Allows configuration of the password hashing options when the user passwords are changed directly by Authelia."`

	Search AuthenticationBackendFileSearch `koanf:"search" json:"search" jsonschema:"title=Search" jsonschema_description:"Configures the user searching behaviour."`
}

// AuthenticationBackendFileSearch represents the configuration related to file-based backend searching.
type AuthenticationBackendFileSearch struct {
	Email           bool `koanf:"email" json:"email" jsonschema:"default=false,title=Email Searching" jsonschema_description:"Allows users to either use their username or their configured email as a username."`
	CaseInsensitive bool `koanf:"case_insensitive" json:"case_insensitive" jsonschema:"default=false,title=Case Insensitive Searching" jsonschema_description:"Allows usernames to be any case during the search."`
}

// AuthenticationBackendFilePassword represents the configuration related to password hashing.
type AuthenticationBackendFilePassword struct {
	Algorithm string `koanf:"algorithm" json:"algorithm" jsonschema:"default=argon2,enum=argon2,enum=sha2crypt,enum=pbkdf2,enum=bcrypt,enum=scrypt,title=Algorithm" jsonschema_description:"The password hashing algorithm to use."`

	Argon2    AuthenticationBackendFilePasswordArgon2    `koanf:"argon2" json:"argon2" jsonschema:"title=Argon2" jsonschema_description:"Configure the Argon2 password hashing parameters."`
	SHA2Crypt AuthenticationBackendFilePasswordSHA2Crypt `koanf:"sha2crypt" json:"sha2crypt" jsonschema:"title=SHA2Crypt" jsonschema_description:"Configure the SHA2Crypt password hashing parameters."`
	PBKDF2    AuthenticationBackendFilePasswordPBKDF2    `koanf:"pbkdf2" json:"pbkdf2" jsonschema:"title=PBKDF2" jsonschema_description:"Configure the PBKDF2 password hashing parameters."`
	BCrypt    AuthenticationBackendFilePasswordBCrypt    `koanf:"bcrypt" json:"bcrypt" jsonschema:"title=BCrypt" jsonschema_description:"Configure the BCrypt password hashing parameters."`
	SCrypt    AuthenticationBackendFilePasswordSCrypt    `koanf:"scrypt" json:"scrypt" jsonschema:"title=SCrypt" jsonschema_description:"Configure the SCrypt password hashing parameters."`

	// Deprecated: Use individual password options instead.
	Iterations int `koanf:"iterations" json:"iterations" jsonschema:"deprecated,title=Iterations"`

	// Deprecated: Use individual password options instead.
	Memory int `koanf:"memory" json:"memory" jsonschema:"deprecated,title=Memory"`

	// Deprecated: Use individual password options instead.
	Parallelism int `koanf:"parallelism" json:"parallelism" jsonschema:"deprecated,title=Parallelism"`

	// Deprecated: Use individual password options instead.
	KeyLength int `koanf:"key_length" json:"key_length" jsonschema:"deprecated,title=Key Length"`

	// Deprecated: Use individual password options instead.
	SaltLength int `koanf:"salt_length" json:"salt_length" jsonschema:"deprecated,title=Salt Length"`
}

// AuthenticationBackendFilePasswordArgon2 represents the argon2 hashing settings.
type AuthenticationBackendFilePasswordArgon2 struct {
	Variant     string `koanf:"variant" json:"variant" jsonschema:"default=argon2id,enum=argon2id,enum=argon2i,enum=argon2d,title=Variant" jsonschema_description:"The Argon2 variant to be used."`
	Iterations  int    `koanf:"iterations" json:"iterations" jsonschema:"default=3,title=Iterations" jsonschema_description:"The number of Argon2 iterations (parameter t) to be used."`
	Memory      int    `koanf:"memory" json:"memory" jsonschema:"default=65536,minimum=8,maximum=4294967295,title=Memory" jsonschema_description:"The Argon2 amount of memory in kibibytes (parameter m) to be used."`
	Parallelism int    `koanf:"parallelism" json:"parallelism" jsonschema:"default=4,minimum=1,maximum=16777215,title=Parallelism" jsonschema_description:"The Argon2 degree of parallelism (parameter p) to be used."`
	KeyLength   int    `koanf:"key_length" json:"key_length" jsonschema:"default=32,minimum=4,maximum=2147483647,title=Key Length" jsonschema_description:"The Argon2 key output length."`
	SaltLength  int    `koanf:"salt_length" json:"salt_length" jsonschema:"default=16,minimum=1,maximum=2147483647,title=Salt Length" jsonschema_description:"The Argon2 salt length."`
}

// AuthenticationBackendFilePasswordSHA2Crypt represents the sha2crypt hashing settings.
type AuthenticationBackendFilePasswordSHA2Crypt struct {
	Variant    string `koanf:"variant" json:"variant" jsonschema:"default=sha512,enum=sha256,enum=sha512,title=Variant" jsonschema_description:"The SHA2Crypt variant to be used."`
	Iterations int    `koanf:"iterations" json:"iterations" jsonschema:"default=50000,minimum=1000,maximum=999999999,title=Iterations" jsonschema_description:"The SHA2Crypt iterations (parameter rounds) to be used."`
	SaltLength int    `koanf:"salt_length" json:"salt_length" jsonschema:"default=16,minimum=1,maximum=16,title=Salt Length" jsonschema_description:"The SHA2Crypt salt length to be used."`
}

// AuthenticationBackendFilePasswordPBKDF2 represents the PBKDF2 hashing settings.
type AuthenticationBackendFilePasswordPBKDF2 struct {
	Variant    string `koanf:"variant" json:"variant" jsonschema:"default=sha512,enum=sha1,enum=sha224,enum=sha256,enum=sha384,enum=sha512,title=Variant" jsonschema_description:"The PBKDF2 variant to be used."`
	Iterations int    `koanf:"iterations" json:"iterations" jsonschema:"default=310000,minimum=100000,maximum=2147483647,title=Iterations" jsonschema_description:"The PBKDF2 iterations to be used."`
	SaltLength int    `koanf:"salt_length" json:"salt_length" jsonschema:"default=16,minimum=8,maximum=2147483647,title=Salt Length" jsonschema_description:"The PBKDF2 salt length to be used."`
}

// AuthenticationBackendFilePasswordBCrypt represents the bcrypt hashing settings.
type AuthenticationBackendFilePasswordBCrypt struct {
	Variant string `koanf:"variant" json:"variant" jsonschema:"default=standard,enum=standard,enum=sha256,title=Variant" jsonschema_description:"The BCrypt variant to be used."`
	Cost    int    `koanf:"cost" json:"cost" jsonschema:"default=12,minimum=10,maximum=31,title=Cost" jsonschema_description:"The BCrypt cost to be used."`
}

// AuthenticationBackendFilePasswordSCrypt represents the scrypt hashing settings.
type AuthenticationBackendFilePasswordSCrypt struct {
	Iterations  int `koanf:"iterations" json:"iterations" jsonschema:"default=16,minimum=1,maximum=58,title=Iterations" jsonschema_description:"The SCrypt iterations to be used."`
	BlockSize   int `koanf:"block_size" json:"block_size" jsonschema:"default=8,minimum=1,maximum=36028797018963967,title=Key Length" jsonschema_description:"The SCrypt block size to be used."`
	Parallelism int `koanf:"parallelism" json:"parallelism" jsonschema:"default=1,minimum=1,maximum=1073741823,title=Key Length" jsonschema_description:"The SCrypt parallelism factor to be used."`
	KeyLength   int `koanf:"key_length" json:"key_length" jsonschema:"default=32,minimum=1,maximum=137438953440,title=Key Length" jsonschema_description:"The SCrypt key length to be used."`
	SaltLength  int `koanf:"salt_length" json:"salt_length" jsonschema:"default=16,minimum=8,maximum=1024,title=Salt Length" jsonschema_description:"The SCrypt salt length to be used."`
}

// AuthenticationBackendLDAP represents the configuration related to LDAP server.
type AuthenticationBackendLDAP struct {
	Address        *AddressLDAP  `koanf:"address" json:"address" jsonschema:"title=Address" jsonschema_description:"The address of the LDAP directory server."`
	Implementation string        `koanf:"implementation" json:"implementation" jsonschema:"default=custom,enum=custom,enum=activedirectory,enum=rfc2307bis,enum=freeipa,enum=lldap,enum=glauth,title=Implementation" jsonschema_description:"The implementation which mostly decides the default values."`
	Timeout        time.Duration `koanf:"timeout" json:"timeout" jsonschema:"default=5 seconds,title=Timeout" jsonschema_description:"The LDAP directory server connection timeout."`
	StartTLS       bool          `koanf:"start_tls" json:"start_tls" jsonschema:"default=false,title=StartTLS" jsonschema_description:"Enables the use of StartTLS."`
	TLS            *TLS          `koanf:"tls" json:"tls" jsonschema:"title=TLS" jsonschema_description:"The LDAP directory server TLS connection properties."`

	BaseDN string `koanf:"base_dn" json:"base_dn" jsonschema:"title=Base DN" jsonschema_description:"The base for all directory server operations."`

	AdditionalUsersDN string `koanf:"additional_users_dn" json:"additional_users_dn" jsonschema:"title=Additional User Base" jsonschema_description:"The base in addition to the Base DN for all directory server operations for users."`
	UsersFilter       string `koanf:"users_filter" json:"users_filter" jsonschema:"title=Users Filter" jsonschema_description:"The LDAP filter used to search for user objects."`

	AdditionalGroupsDN string `koanf:"additional_groups_dn" json:"additional_groups_dn" jsonschema:"title=Additional Group Base" jsonschema_description:"The base in addition to the Base DN for all directory server operations for groups."`
	GroupsFilter       string `koanf:"groups_filter" json:"groups_filter" jsonschema:"title=Groups Filter" jsonschema_description:"The LDAP filter used to search for group objects."`
	GroupSearchMode    string `koanf:"group_search_mode" json:"group_search_mode" jsonschema:"default=filter,enum=filter,enum=memberof,title=Groups Search Mode" jsonschema_description:"The LDAP group search mode used to search for group objects."`

	Attributes AuthenticationBackendLDAPAttributes `koanf:"attributes" json:"attributes"`

	PermitReferrals               bool `koanf:"permit_referrals" json:"permit_referrals" jsonschema:"default=false,title=Permit Referrals" jsonschema_description:"Enables chasing LDAP referrals."`
	PermitUnauthenticatedBind     bool `koanf:"permit_unauthenticated_bind" json:"permit_unauthenticated_bind" jsonschema:"default=false,title=Permit Unauthenticated Bind" jsonschema_description:"Enables omission of the password to perform an unauthenticated bind."`
	PermitFeatureDetectionFailure bool `koanf:"permit_feature_detection_failure" json:"permit_feature_detection_failure" jsonschema:"default=false,title=Permit Feature Detection Failure" jsonschema_description:"Enables failures when detecting directory server features using the Root DSE lookup."`

	User     string `koanf:"user" json:"user" jsonschema:"title=User" jsonschema_description:"The user distinguished name for LDAP binding."`
	Password string `koanf:"password" json:"password" jsonschema:"title=Password" jsonschema_description:"The password for LDAP authenticated binding."`
}

// AuthenticationBackendLDAPAttributes represents the configuration related to LDAP server attributes.
type AuthenticationBackendLDAPAttributes struct {
	DistinguishedName string `koanf:"distinguished_name" json:"distinguished_name" jsonschema:"title=Attribute: Distinguished Name" jsonschema_description:"The directory server attribute which contains the distinguished name for all objects."`
	Username          string `koanf:"username" json:"username" jsonschema:"title=Attribute: User Username" jsonschema_description:"The directory server attribute which contains the username for all users."`
	DisplayName       string `koanf:"display_name" json:"display_name" jsonschema:"title=Attribute: User Display Name" jsonschema_description:"The directory server attribute which contains the display name for all users."`
	Mail              string `koanf:"mail" json:"mail" jsonschema:"title=Attribute: User Mail" jsonschema_description:"The directory server attribute which contains the mail address for all users and groups."`
	MemberOf          string `koanf:"member_of" jsonschema:"title=Attribute: Member Of" jsonschema_description:"The directory server attribute which contains the objects that an object is a member of."`
	GroupName         string `koanf:"group_name" json:"group_name" jsonschema:"title=Attribute: Group Name" jsonschema_description:"The directory server attribute which contains the group name for all groups."`
}

var DefaultAuthenticationBackendConfig = AuthenticationBackend{
	RefreshInterval: NewRefreshIntervalDuration(time.Minute * 5),
}

// DefaultPasswordConfig represents the default configuration related to Argon2id hashing.
var DefaultPasswordConfig = AuthenticationBackendFilePassword{
	Algorithm: argon2,
	Argon2: AuthenticationBackendFilePasswordArgon2{
		Variant:     argon2id,
		Iterations:  3,
		Memory:      64 * 1024,
		Parallelism: 4,
		KeyLength:   32,
		SaltLength:  16,
	},
	SHA2Crypt: AuthenticationBackendFilePasswordSHA2Crypt{
		Variant:    sha512,
		Iterations: 50000,
		SaltLength: 16,
	},
	PBKDF2: AuthenticationBackendFilePasswordPBKDF2{
		Variant:    sha512,
		Iterations: 310000,
		SaltLength: 16,
	},
	BCrypt: AuthenticationBackendFilePasswordBCrypt{
		Variant: "standard",
		Cost:    12,
	},
	SCrypt: AuthenticationBackendFilePasswordSCrypt{
		Iterations:  16,
		BlockSize:   8,
		Parallelism: 1,
		KeyLength:   32,
		SaltLength:  16,
	},
}

// DefaultCIPasswordConfig represents the default configuration related to Argon2id hashing for CI.
var DefaultCIPasswordConfig = AuthenticationBackendFilePassword{
	Algorithm: argon2,
	Argon2: AuthenticationBackendFilePasswordArgon2{
		Iterations:  3,
		Memory:      64,
		Parallelism: 4,
		KeyLength:   32,
		SaltLength:  16,
	},
	SHA2Crypt: AuthenticationBackendFilePasswordSHA2Crypt{
		Variant:    sha512,
		Iterations: 50000,
		SaltLength: 16,
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationCustom represents the default LDAP config.
var DefaultLDAPAuthenticationBackendConfigurationImplementationCustom = AuthenticationBackendLDAP{
	GroupSearchMode: ldapGroupSearchModeFilter,
	Attributes: AuthenticationBackendLDAPAttributes{
		Username:    ldapAttrUserID,
		DisplayName: ldapAttrDisplayName,
		Mail:        ldapAttrMail,
		GroupName:   ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLS{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory represents the default LDAP config for the LDAPImplementationActiveDirectory Implementation.
var DefaultLDAPAuthenticationBackendConfigurationImplementationActiveDirectory = AuthenticationBackendLDAP{
	UsersFilter:     "(&(|({username_attribute}={input})({mail_attribute}={input}))(sAMAccountType=805306368)(!(userAccountControl:1.2.840.113556.1.4.803:=2))(!(pwdLastSet=0))(|(!(accountExpires=*))(accountExpires=0)(accountExpires>={date-time:microsoft-nt})))",
	GroupsFilter:    "(&(member={dn})(|(sAMAccountType=268435456)(sAMAccountType=536870912)))",
	GroupSearchMode: ldapGroupSearchModeFilter,
	Attributes: AuthenticationBackendLDAPAttributes{
		DistinguishedName: ldapAttrDistinguishedName,
		Username:          ldapAttrSAMAccountName,
		DisplayName:       ldapAttrDisplayName,
		Mail:              ldapAttrMail,
		MemberOf:          ldapAttrMemberOf,
		GroupName:         ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLS{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis represents the default LDAP config for the LDAPImplementationRFC2307bis Implementation.
var DefaultLDAPAuthenticationBackendConfigurationImplementationRFC2307bis = AuthenticationBackendLDAP{
	UsersFilter:     "(&(|({username_attribute}={input})({mail_attribute}={input}))(|(objectClass=inetOrgPerson)(objectClass=organizationalPerson)))",
	GroupsFilter:    "(&(|(member={dn})(uniqueMember={dn}))(|(objectClass=groupOfNames)(objectClass=groupOfUniqueNames)(objectClass=groupOfMembers))(!(pwdReset=TRUE)))",
	GroupSearchMode: ldapGroupSearchModeFilter,
	Attributes: AuthenticationBackendLDAPAttributes{
		Username:    ldapAttrUserID,
		DisplayName: ldapAttrDisplayName,
		Mail:        ldapAttrMail,
		MemberOf:    ldapAttrMemberOf,
		GroupName:   ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLS{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA represents the default LDAP config for the LDAPImplementationFreeIPA Implementation.
var DefaultLDAPAuthenticationBackendConfigurationImplementationFreeIPA = AuthenticationBackendLDAP{
	UsersFilter:     "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person)(!(nsAccountLock=TRUE))(krbPasswordExpiration>={date-time:generalized})(|(!(krbPrincipalExpiration=*))(krbPrincipalExpiration>={date-time:generalized})))",
	GroupsFilter:    "(&(member={dn})(objectClass=groupOfNames))",
	GroupSearchMode: ldapGroupSearchModeFilter,
	Attributes: AuthenticationBackendLDAPAttributes{
		Username:    ldapAttrUserID,
		DisplayName: ldapAttrDisplayName,
		Mail:        ldapAttrMail,
		MemberOf:    ldapAttrMemberOf,
		GroupName:   ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLS{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP represents the default LDAP config for the LDAPImplementationLLDAP Implementation.
var DefaultLDAPAuthenticationBackendConfigurationImplementationLLDAP = AuthenticationBackendLDAP{
	AdditionalUsersDN:  "OU=people",
	AdditionalGroupsDN: "OU=groups",
	UsersFilter:        "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=person))",
	GroupsFilter:       "(&(member={dn})(objectClass=groupOfUniqueNames))",
	GroupSearchMode:    ldapGroupSearchModeFilter,
	Attributes: AuthenticationBackendLDAPAttributes{
		Username:    ldapAttrUserID,
		DisplayName: ldapAttrCommonName,
		Mail:        ldapAttrMail,
		MemberOf:    ldapAttrMemberOf,
		GroupName:   ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLS{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}

// DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth represents the default LDAP config for the LDAPImplementationGLAuth Implementation.
var DefaultLDAPAuthenticationBackendConfigurationImplementationGLAuth = AuthenticationBackendLDAP{
	UsersFilter:     "(&(|({username_attribute}={input})({mail_attribute}={input}))(objectClass=posixAccount)(!(accountStatus=inactive)))",
	GroupsFilter:    "(&(uniqueMember={dn})(objectClass=posixGroup))",
	GroupSearchMode: ldapGroupSearchModeFilter,
	Attributes: AuthenticationBackendLDAPAttributes{
		Username:    ldapAttrCommonName,
		DisplayName: ldapAttrDescription,
		Mail:        ldapAttrMail,
		MemberOf:    ldapAttrMemberOf,
		GroupName:   ldapAttrCommonName,
	},
	Timeout: time.Second * 5,
	TLS: &TLS{
		MinimumVersion: TLSVersion{tls.VersionTLS12},
	},
}
