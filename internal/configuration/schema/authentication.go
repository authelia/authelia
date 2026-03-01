package schema

import (
	"crypto/tls"
	"net/url"
	"time"
)

// AuthenticationBackend represents the configuration related to the authentication backend.
type AuthenticationBackend struct {
	PasswordReset  AuthenticationBackendPasswordReset  `koanf:"password_reset" yaml:"password_reset,omitempty" toml:"password_reset,omitempty" json:"password_reset,omitempty" jsonschema:"title=Password Reset" jsonschema_description:"Allows configuration of the password reset behaviour."`
	PasswordChange AuthenticationBackendPasswordChange `koanf:"password_change" yaml:"password_change,omitempty" toml:"password_change,omitempty" json:"password_change,omitempty" jsonschema:"title=Password Reset" jsonschema_description:"Allows configuration of the password reset behaviour."`

	RefreshInterval RefreshIntervalDuration `koanf:"refresh_interval" yaml:"refresh_interval,omitempty" toml:"refresh_interval,omitempty" json:"refresh_interval,omitempty" jsonschema:"default=5 minutes,title=Refresh Interval" jsonschema_description:"How frequently the user details are refreshed from the backend."`

	// The file authentication backend configuration.
	File *AuthenticationBackendFile `koanf:"file" yaml:"file,omitempty" toml:"file,omitempty" json:"file,omitempty" jsonschema:"title=File Backend" jsonschema_description:"The file authentication backend configuration."`
	LDAP *AuthenticationBackendLDAP `koanf:"ldap" yaml:"ldap,omitempty" toml:"ldap,omitempty" json:"ldap,omitempty" jsonschema:"title=LDAP Backend" jsonschema_description:"The LDAP authentication backend configuration."`
}

// AuthenticationBackendPasswordChange represents the configuration related to password reset functionality.
type AuthenticationBackendPasswordChange struct {
	Disable bool `koanf:"disable" yaml:"disable" toml:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables the Password Change option."`
}

// AuthenticationBackendPasswordReset represents the configuration related to password reset functionality.
type AuthenticationBackendPasswordReset struct {
	Disable   bool    `koanf:"disable" yaml:"disable" toml:"disable" json:"disable" jsonschema:"default=false,title=Disable" jsonschema_description:"Disables the Password Reset option."`
	CustomURL url.URL `koanf:"custom_url" yaml:"custom_url,omitempty" toml:"custom_url,omitempty" json:"custom_url,omitempty" jsonschema:"title=Custom URL" jsonschema_description:"Disables the internal Password Reset option and instead redirects users to this specified URL."`
}

// AuthenticationBackendFile represents the configuration related to file-based backend.
type AuthenticationBackendFile struct {
	Path  string `koanf:"path" yaml:"path,omitempty" toml:"path,omitempty" json:"path,omitempty" jsonschema:"title=Path" jsonschema_description:"The file path to the user database."`
	Watch bool   `koanf:"watch" yaml:"watch" toml:"watch" json:"watch" jsonschema:"default=false,title=Watch" jsonschema_description:"Enables watching the file for external changes and dynamically reloading the database."`

	Password AuthenticationBackendFilePassword `koanf:"password" yaml:"password,omitempty" toml:"password,omitempty" json:"password,omitempty" jsonschema:"title=Password Options" jsonschema_description:"Allows configuration of the password hashing options when the user passwords are changed directly by Authelia."`

	Search AuthenticationBackendFileSearch `koanf:"search" yaml:"search,omitempty" toml:"search,omitempty" json:"search,omitempty" jsonschema:"title=Search" jsonschema_description:"Configures the user searching behaviour."`

	ExtraAttributes map[string]AuthenticationBackendExtraAttribute `koanf:"extra_attributes" yaml:"extra_attributes,omitempty" toml:"extra_attributes,omitempty" json:"extra_attributes,omitempty" jsonschema:"title=Extra Attributes" jsonschema_description:"Configures the extra attributes available in expressions and other areas of Authelia."`
}

type AuthenticationBackendExtraAttribute struct {
	MultiValued bool   `koanf:"multi_valued" yaml:"multi_valued" toml:"multi_valued" json:"multi_valued" jsonschema:"title=Multi-Valued" jsonschema_description:"Defines the attribute as multi-valued."`
	ValueType   string `koanf:"value_type" yaml:"value_type,omitempty" toml:"value_type,omitempty" json:"value_type,omitempty" jsonschema:"enum=boolean,enum=integer,enum=string,title=Value Type" jsonschema_description:"Defines the value type for the attribute."`
}

func (a AuthenticationBackendExtraAttribute) IsMultiValued() (multi bool) {
	return a.MultiValued
}

func (a AuthenticationBackendExtraAttribute) GetValueType() (vtype string) {
	return a.ValueType
}

// AuthenticationBackendFileSearch represents the configuration related to file-based backend searching.
type AuthenticationBackendFileSearch struct {
	Email           bool `koanf:"email" yaml:"email" toml:"email" json:"email" jsonschema:"default=false,title=Email Searching" jsonschema_description:"Allows users to either use their username or their configured email as a username."`
	CaseInsensitive bool `koanf:"case_insensitive" yaml:"case_insensitive" toml:"case_insensitive" json:"case_insensitive" jsonschema:"default=false,title=Case Insensitive Searching" jsonschema_description:"Allows usernames to be any case during the search."`
}

// AuthenticationBackendFilePassword represents the configuration related to password hashing.
type AuthenticationBackendFilePassword struct {
	Algorithm string `koanf:"algorithm" yaml:"algorithm,omitempty" toml:"algorithm,omitempty" json:"algorithm,omitempty" jsonschema:"default=argon2,enum=argon2,enum=sha2crypt,enum=pbkdf2,enum=bcrypt,enum=scrypt,title=Algorithm" jsonschema_description:"The password hashing algorithm to use."`

	Argon2    AuthenticationBackendFilePasswordArgon2    `koanf:"argon2" yaml:"argon2,omitempty" toml:"argon2,omitempty" json:"argon2,omitempty" jsonschema:"title=Argon2" jsonschema_description:"Configure the Argon2 password hashing parameters."`
	SHA2Crypt AuthenticationBackendFilePasswordSHA2Crypt `koanf:"sha2crypt" yaml:"sha2crypt,omitempty" toml:"sha2crypt,omitempty" json:"sha2crypt,omitempty" jsonschema:"title=SHA2Crypt" jsonschema_description:"Configure the SHA2Crypt password hashing parameters."`
	PBKDF2    AuthenticationBackendFilePasswordPBKDF2    `koanf:"pbkdf2" yaml:"pbkdf2,omitempty" toml:"pbkdf2,omitempty" json:"pbkdf2,omitempty" jsonschema:"title=PBKDF2" jsonschema_description:"Configure the PBKDF2 password hashing parameters."`
	Bcrypt    AuthenticationBackendFilePasswordBcrypt    `koanf:"bcrypt" yaml:"bcrypt,omitempty" toml:"bcrypt,omitempty" json:"bcrypt,omitempty" jsonschema:"title=Bcrypt" jsonschema_description:"Configure the Bcrypt password hashing parameters."`
	Scrypt    AuthenticationBackendFilePasswordScrypt    `koanf:"scrypt" yaml:"scrypt,omitempty" toml:"scrypt,omitempty" json:"scrypt,omitempty" jsonschema:"title=Scrypt" jsonschema_description:"Configure the Scrypt password hashing parameters."`

	// Deprecated: Use individual password options instead.
	Iterations int `koanf:"iterations" yaml:"iterations" toml:"iterations" json:"iterations" jsonschema:"deprecated,title=Iterations"`

	// Deprecated: Use individual password options instead.
	Memory int `koanf:"memory" yaml:"memory" toml:"memory" json:"memory" jsonschema:"deprecated,title=Memory"`

	// Deprecated: Use individual password options instead.
	Parallelism int `koanf:"parallelism" yaml:"parallelism" toml:"parallelism" json:"parallelism" jsonschema:"deprecated,title=Parallelism"`

	// Deprecated: Use individual password options instead.
	KeyLength int `koanf:"key_length" yaml:"key_length" toml:"key_length" json:"key_length" jsonschema:"deprecated,title=Key Length"`

	// Deprecated: Use individual password options instead.
	SaltLength int `koanf:"salt_length" yaml:"salt_length" toml:"salt_length" json:"salt_length" jsonschema:"deprecated,title=Salt Length"`
}

// AuthenticationBackendFilePasswordArgon2 represents the argon2 hashing settings.
type AuthenticationBackendFilePasswordArgon2 struct {
	Variant     string `koanf:"variant" yaml:"variant,omitempty" toml:"variant,omitempty" json:"variant,omitempty" jsonschema:"default=argon2id,enum=argon2id,enum=argon2i,enum=argon2d,title=Variant" jsonschema_description:"The Argon2 variant to be used."`
	Iterations  int    `koanf:"iterations" yaml:"iterations" toml:"iterations" json:"iterations" jsonschema:"default=3,title=Iterations" jsonschema_description:"The number of Argon2 iterations (parameter t) to be used."`
	Memory      int    `koanf:"memory" yaml:"memory" toml:"memory" json:"memory" jsonschema:"default=65536,minimum=8,maximum=4294967295,title=Memory" jsonschema_description:"The Argon2 amount of memory in kibibytes (parameter m) to be used."`
	Parallelism int    `koanf:"parallelism" yaml:"parallelism" toml:"parallelism" json:"parallelism" jsonschema:"default=4,minimum=1,maximum=16777215,title=Parallelism" jsonschema_description:"The Argon2 degree of parallelism (parameter p) to be used."`
	KeyLength   int    `koanf:"key_length" yaml:"key_length" toml:"key_length" json:"key_length" jsonschema:"default=32,minimum=4,maximum=2147483647,title=Key Length" jsonschema_description:"The Argon2 key output length."`
	SaltLength  int    `koanf:"salt_length" yaml:"salt_length" toml:"salt_length" json:"salt_length" jsonschema:"default=16,minimum=1,maximum=2147483647,title=Salt Length" jsonschema_description:"The Argon2 salt length."`
}

// AuthenticationBackendFilePasswordSHA2Crypt represents the sha2crypt hashing settings.
type AuthenticationBackendFilePasswordSHA2Crypt struct {
	Variant    string `koanf:"variant" yaml:"variant,omitempty" toml:"variant,omitempty" json:"variant,omitempty" jsonschema:"default=sha512,enum=sha256,enum=sha512,title=Variant" jsonschema_description:"The SHA2Crypt variant to be used."`
	Iterations int    `koanf:"iterations" yaml:"iterations" toml:"iterations" json:"iterations" jsonschema:"default=50000,minimum=1000,maximum=999999999,title=Iterations" jsonschema_description:"The SHA2Crypt iterations (parameter rounds) to be used."`
	SaltLength int    `koanf:"salt_length" yaml:"salt_length" toml:"salt_length" json:"salt_length" jsonschema:"default=16,minimum=1,maximum=16,title=Salt Length" jsonschema_description:"The SHA2Crypt salt length to be used."`
}

// AuthenticationBackendFilePasswordPBKDF2 represents the PBKDF2 hashing settings.
type AuthenticationBackendFilePasswordPBKDF2 struct {
	Variant    string `koanf:"variant" yaml:"variant,omitempty" toml:"variant,omitempty" json:"variant,omitempty" jsonschema:"default=sha512,enum=sha1,enum=sha224,enum=sha256,enum=sha384,enum=sha512,title=Variant" jsonschema_description:"The PBKDF2 variant to be used."`
	Iterations int    `koanf:"iterations" yaml:"iterations" toml:"iterations" json:"iterations" jsonschema:"default=310000,minimum=100000,maximum=2147483647,title=Iterations" jsonschema_description:"The PBKDF2 iterations to be used."`
	SaltLength int    `koanf:"salt_length" yaml:"salt_length" toml:"salt_length" json:"salt_length" jsonschema:"default=16,minimum=8,maximum=2147483647,title=Salt Length" jsonschema_description:"The PBKDF2 salt length to be used."`
}

// AuthenticationBackendFilePasswordBcrypt represents the bcrypt hashing settings.
type AuthenticationBackendFilePasswordBcrypt struct {
	Variant string `koanf:"variant" yaml:"variant,omitempty" toml:"variant,omitempty" json:"variant,omitempty" jsonschema:"default=standard,enum=standard,enum=sha256,title=Variant" jsonschema_description:"The Bcrypt variant to be used."`
	Cost    int    `koanf:"cost" yaml:"cost" toml:"cost" json:"cost" jsonschema:"default=12,minimum=10,maximum=31,title=Cost" jsonschema_description:"The Bcrypt cost to be used."`
}

// AuthenticationBackendFilePasswordScrypt represents the scrypt hashing settings.
type AuthenticationBackendFilePasswordScrypt struct {
	Variant     string `koanf:"variant" yaml:"variant,omitempty" toml:"variant,omitempty" json:"variant,omitempty" jsonschema:"default=scrypt,enum=scrypt,enum=yescrypt,titleVariant" jsonschema_description:"The Scrypt variant to be used."`
	Iterations  int    `koanf:"iterations" yaml:"iterations" toml:"iterations" json:"iterations" jsonschema:"default=16,minimum=1,maximum=58,title=Iterations" jsonschema_description:"The Scrypt iterations to be used."`
	BlockSize   int    `koanf:"block_size" yaml:"block_size" toml:"block_size" json:"block_size" jsonschema:"default=8,minimum=1,maximum=36028797018963967,title=Key Length" jsonschema_description:"The Scrypt block size to be used."`
	Parallelism int    `koanf:"parallelism" yaml:"parallelism" toml:"parallelism" json:"parallelism" jsonschema:"default=1,minimum=1,maximum=1073741823,title=Key Length" jsonschema_description:"The Scrypt parallelism factor to be used."`
	KeyLength   int    `koanf:"key_length" yaml:"key_length" toml:"key_length" json:"key_length" jsonschema:"default=32,minimum=1,maximum=137438953440,title=Key Length" jsonschema_description:"The Scrypt key length to be used."`
	SaltLength  int    `koanf:"salt_length" yaml:"salt_length" toml:"salt_length" json:"salt_length" jsonschema:"default=16,minimum=8,maximum=1024,title=Salt Length" jsonschema_description:"The Scrypt salt length to be used."`
}

// AuthenticationBackendLDAP represents the configuration related to LDAP server.
type AuthenticationBackendLDAP struct {
	Address        *AddressLDAP  `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty" jsonschema:"title=Address" jsonschema_description:"The address of the LDAP directory server."`
	Implementation string        `koanf:"implementation" yaml:"implementation,omitempty" toml:"implementation,omitempty" json:"implementation,omitempty" jsonschema:"default=custom,enum=custom,enum=activedirectory,enum=rfc2307bis,enum=freeipa,enum=lldap,enum=glauth,title=Implementation" jsonschema_description:"The implementation which mostly decides the default values."`
	Timeout        time.Duration `koanf:"timeout" yaml:"timeout,omitempty" toml:"timeout,omitempty" json:"timeout,omitempty" jsonschema:"default=20 seconds,title=Timeout" jsonschema_description:"The LDAP directory server connection timeout."`
	StartTLS       bool          `koanf:"start_tls" yaml:"start_tls" toml:"start_tls" json:"start_tls" jsonschema:"default=false,title=StartTLS" jsonschema_description:"Forces the use of StartTLS."`
	TLS            *TLS          `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty" jsonschema:"title=TLS" jsonschema_description:"The LDAP directory server TLS connection properties."`

	Pooling AuthenticationBackendLDAPPooling `koanf:"pooling" yaml:"pooling,omitempty" toml:"pooling,omitempty" json:"pooling,omitempty" jsonschema:"title=Pooling" jsonschema_description:"The LDAP Connection Pooling properties."`

	BaseDN string `koanf:"base_dn" yaml:"base_dn,omitempty" toml:"base_dn,omitempty" json:"base_dn,omitempty" jsonschema:"title=Base DN" jsonschema_description:"The base for all directory server operations."`

	AdditionalUsersDN string `koanf:"additional_users_dn" yaml:"additional_users_dn,omitempty" toml:"additional_users_dn,omitempty" json:"additional_users_dn,omitempty" jsonschema:"title=Additional User Base" jsonschema_description:"The base in addition to the Base DN for all directory server operations for users."`
	UsersFilter       string `koanf:"users_filter" yaml:"users_filter,omitempty" toml:"users_filter,omitempty" json:"users_filter,omitempty" jsonschema:"title=Users Filter" jsonschema_description:"The LDAP filter used to search for user objects."`

	AdditionalGroupsDN string `koanf:"additional_groups_dn" yaml:"additional_groups_dn,omitempty" toml:"additional_groups_dn,omitempty" json:"additional_groups_dn,omitempty" jsonschema:"title=Additional Group Base" jsonschema_description:"The base in addition to the Base DN for all directory server operations for groups."`
	GroupsFilter       string `koanf:"groups_filter" yaml:"groups_filter,omitempty" toml:"groups_filter,omitempty" json:"groups_filter,omitempty" jsonschema:"title=Groups Filter" jsonschema_description:"The LDAP filter used to search for group objects."`
	GroupSearchMode    string `koanf:"group_search_mode" yaml:"group_search_mode,omitempty" toml:"group_search_mode,omitempty" json:"group_search_mode,omitempty" jsonschema:"default=filter,enum=filter,enum=memberof,title=Groups Search Modes" jsonschema_description:"The LDAP group search mode used to search for group objects."`

	Attributes AuthenticationBackendLDAPAttributes `koanf:"attributes" yaml:"attributes,omitempty" toml:"attributes,omitempty" json:"attributes,omitempty"`

	PermitReferrals               bool `koanf:"permit_referrals" yaml:"permit_referrals" toml:"permit_referrals" json:"permit_referrals" jsonschema:"default=false,title=Permit Referrals" jsonschema_description:"Enables chasing LDAP referrals."`
	PermitUnauthenticatedBind     bool `koanf:"permit_unauthenticated_bind" yaml:"permit_unauthenticated_bind" toml:"permit_unauthenticated_bind" json:"permit_unauthenticated_bind" jsonschema:"default=false,title=Permit Unauthenticated Bind" jsonschema_description:"Enables omission of the password to perform an unauthenticated bind."`
	PermitFeatureDetectionFailure bool `koanf:"permit_feature_detection_failure" yaml:"permit_feature_detection_failure" toml:"permit_feature_detection_failure" json:"permit_feature_detection_failure" jsonschema:"default=false,title=Permit Feature Detection Failure" jsonschema_description:"Enables failures when detecting directory server features using the Root DSE lookup."`

	User     string `koanf:"user" yaml:"user,omitempty" toml:"user,omitempty" json:"user,omitempty" jsonschema:"title=User" jsonschema_description:"The user distinguished name for LDAP binding."`
	Password string `koanf:"password" yaml:"password,omitempty" toml:"password,omitempty" json:"password,omitempty" jsonschema:"title=Password" jsonschema_description:"The password for LDAP authenticated binding."` //nolint:gosec
}

type AuthenticationBackendLDAPPooling struct {
	Enable  bool          `koanf:"enable" yaml:"enable" toml:"enable" json:"enable" jsonschema:"title=Enable,default=false" jsonschema_description:"Enable LDAP connection pooling."`
	Count   int           `koanf:"count" yaml:"count" toml:"count" json:"count" jsonschema:"title=Count,default=5" jsonschema_description:"The number of connections to keep open for LDAP connection pooling."`
	Retries int           `koanf:"retries" yaml:"retries" toml:"retries" json:"retries" jsonschema:"title=Retries,default=2" jsonschema_description:"The number of attempts to retrieve a connection from the pool during the timeout."`
	Timeout time.Duration `koanf:"timeout" yaml:"timeout,omitempty" toml:"timeout,omitempty" json:"timeout,omitempty" jsonschema:"title=Timeout,default=10 seconds" jsonschema_description:"The duration of time to wait for a connection to become available in the connection pool."`
}

// AuthenticationBackendLDAPAttributes represents the configuration related to LDAP server attributes.
type AuthenticationBackendLDAPAttributes struct {
	DistinguishedName string `koanf:"distinguished_name" yaml:"distinguished_name,omitempty" toml:"distinguished_name,omitempty" json:"distinguished_name,omitempty" jsonschema:"title=Attribute: Distinguished Name" jsonschema_description:"The directory server attribute which contains the distinguished name for all objects."`
	Username          string `koanf:"username" yaml:"username,omitempty" toml:"username,omitempty" json:"username,omitempty" jsonschema:"title=Attribute: User Username" jsonschema_description:"The directory server attribute which contains the username for all users."`
	DisplayName       string `koanf:"display_name" yaml:"display_name,omitempty" toml:"display_name,omitempty" json:"display_name,omitempty" jsonschema:"title=Attribute: User Display Name" jsonschema_description:"The directory server attribute which contains the display name for all users."`
	FamilyName        string `koanf:"family_name" yaml:"family_name,omitempty" toml:"family_name,omitempty" json:"family_name,omitempty" jsonschema:"title=Attribute: Family Name" jsonschema_description:"The directory server attribute which contains the family name for all users."`
	GivenName         string `koanf:"given_name" yaml:"given_name,omitempty" toml:"given_name,omitempty" json:"given_name,omitempty" jsonschema:"title=Attribute: Given Name" jsonschema_description:"The directory server attribute which contains the given name for all users."`
	MiddleName        string `koanf:"middle_name" yaml:"middle_name,omitempty" toml:"middle_name,omitempty" json:"middle_name,omitempty" jsonschema:"title=Attribute: Middle Name" jsonschema_description:"The directory server attribute which contains the middle name for all users."`
	Nickname          string `koanf:"nickname" yaml:"nickname,omitempty" toml:"nickname,omitempty" json:"nickname,omitempty" jsonschema:"title=Attribute: Nickname" jsonschema_description:"The directory server attribute which contains the nickname for all users."`
	Gender            string `koanf:"gender" yaml:"gender,omitempty" toml:"gender,omitempty" json:"gender,omitempty" jsonschema:"title=Attribute: Gender" jsonschema_description:"The directory server attribute which contains the gender for all users."`
	Birthdate         string `koanf:"birthdate" yaml:"birthdate,omitempty" toml:"birthdate,omitempty" json:"birthdate,omitempty" jsonschema:"title=Attribute: Birthdate" jsonschema_description:"The directory server attribute which contains the birthdate for all users."`
	Website           string `koanf:"website" yaml:"website,omitempty" toml:"website,omitempty" json:"website,omitempty" jsonschema:"title=Attribute: Website" jsonschema_description:"The directory server attribute which contains the website URL for all users."`
	Profile           string `koanf:"profile" yaml:"profile,omitempty" toml:"profile,omitempty" json:"profile,omitempty" jsonschema:"title=Attribute: Profile" jsonschema_description:"The directory server attribute which contains the profile URL for all users."`
	Picture           string `koanf:"picture" yaml:"picture,omitempty" toml:"picture,omitempty" json:"picture,omitempty" jsonschema:"title=Attribute: Picture" jsonschema_description:"The directory server attribute which contains the picture URL for all users."`
	ZoneInfo          string `koanf:"zoneinfo" yaml:"zoneinfo,omitempty" toml:"zoneinfo,omitempty" json:"zoneinfo,omitempty" jsonschema:"title=Attribute: Zone Information" jsonschema_description:"The directory server attribute which contains the time zone information for all users."`
	Locale            string `koanf:"locale" yaml:"locale,omitempty" toml:"locale,omitempty" json:"locale,omitempty" jsonschema:"title=Attribute: Locale" jsonschema_description:"The directory server attribute which contains the locale information for all users."`
	PhoneNumber       string `koanf:"phone_number" yaml:"phone_number,omitempty" toml:"phone_number,omitempty" json:"phone_number,omitempty" jsonschema:"title=Attribute: Phone Number" jsonschema_description:"The directory server attribute which contains the phone number for all users."`
	PhoneExtension    string `koanf:"phone_extension" yaml:"phone_extension,omitempty" toml:"phone_extension,omitempty" json:"phone_extension,omitempty" jsonschema:"title=Attribute: Phone Extension" jsonschema_description:"The directory server attribute which contains the phone extension for all users."`
	StreetAddress     string `koanf:"street_address" yaml:"street_address,omitempty" toml:"street_address,omitempty" json:"street_address,omitempty" jsonschema:"title=Attribute: Street Address" jsonschema_description:"The directory server attribute which contains the street address for all users."`
	Locality          string `koanf:"locality" yaml:"locality,omitempty" toml:"locality,omitempty" json:"locality,omitempty" jsonschema:"title=Attribute: Locality" jsonschema_description:"The directory server attribute which contains the locality for all users."`
	Region            string `koanf:"region" yaml:"region,omitempty" toml:"region,omitempty" json:"region,omitempty" jsonschema:"title=Attribute: Region" jsonschema_description:"The directory server attribute which contains the region for all users."`
	PostalCode        string `koanf:"postal_code" yaml:"postal_code,omitempty" toml:"postal_code,omitempty" json:"postal_code,omitempty" jsonschema:"title=Attribute: Postal Code" jsonschema_description:"The directory server attribute which contains the postal code for all users."`
	Country           string `koanf:"country" yaml:"country,omitempty" toml:"country,omitempty" json:"country,omitempty" jsonschema:"title=Attribute: Country" jsonschema_description:"The directory server attribute which contains the country for all users."`
	Mail              string `koanf:"mail" yaml:"mail,omitempty" toml:"mail,omitempty" json:"mail,omitempty" jsonschema:"title=Attribute: User Mail" jsonschema_description:"The directory server attribute which contains the mail address for all users and groups."`
	MemberOf          string `koanf:"member_of" yaml:"member_of,omitempty" toml:"member_of,omitempty" json:"member_of,omitempty" jsonschema:"title=Attribute: Member Of" jsonschema_description:"The directory server attribute which contains the objects that an object is a member of."`
	GroupName         string `koanf:"group_name" yaml:"group_name,omitempty" toml:"group_name,omitempty" json:"group_name,omitempty" jsonschema:"title=Attribute: Group Name" jsonschema_description:"The directory server attribute which contains the group name for all groups."`

	Extra map[string]AuthenticationBackendLDAPAttributesAttribute `koanf:"extra" yaml:"extra,omitempty" toml:"extra,omitempty" json:"extra,omitempty" jsonschema:"title=Extra Attributes" jsonschema_description:"Configures the extra attributes available in expressions and other areas of Authelia."`
}

type AuthenticationBackendLDAPAttributesAttribute struct {
	Name string `koanf:"name" yaml:"name,omitempty" toml:"name,omitempty" json:"name,omitempty" jsonschema:"title=Name" jsonschema_description:"The name of the attribute within Authelia. This does not adjust the attribute queried from the LDAP server."`

	AuthenticationBackendExtraAttribute `koanf:",squash"`
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
		Variant:    SHA512Lower,
		Iterations: 50000,
		SaltLength: 16,
	},
	PBKDF2: AuthenticationBackendFilePasswordPBKDF2{
		Variant:    SHA512Lower,
		Iterations: defaultIterationsPBKDF2SHA512,
		SaltLength: 16,
	},
	Bcrypt: AuthenticationBackendFilePasswordBcrypt{
		Variant: "standard",
		Cost:    12,
	},
	Scrypt: AuthenticationBackendFilePasswordScrypt{
		Variant:     "scrypt",
		Iterations:  16,
		BlockSize:   8,
		Parallelism: 1,
		KeyLength:   32,
		SaltLength:  16,
	},
}

const (
	defaultIterationsPBKDF2SHA512 = 310000
	defaultIterationsPBKDF2SHA384 = 280000
	defaultIterationsPBKDF2SHA256 = 700000
	defaultIterationsPBKDF2SHA224 = 900000
	defaultIterationsPBKDF2SHA1   = 1600000
)

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
		Variant:    SHA512Lower,
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
	Timeout: time.Second * 20,
	Pooling: AuthenticationBackendLDAPPooling{
		Count:   5,
		Retries: 2,
		Timeout: time.Second * 10,
	},
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
		FamilyName:        ldapAttrSurname,
		GivenName:         ldapAttrGivenName,
		MiddleName:        ldapAttrMiddleName,
		Website:           "wWWHomePage",
		Mail:              ldapAttrMail,
		PhoneNumber:       "telephoneNumber",
		StreetAddress:     "streetAddress",
		Locality:          "l",
		Region:            "st",
		PostalCode:        "postalCode",
		Country:           "c",
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
