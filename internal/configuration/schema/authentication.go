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
	MailAttribute      string `mapstructure:"mail_attribute"`
	User               string `mapstructure:"user"`
	Password           string `mapstructure:"password"`
}

// FileAuthenticationBackendConfiguration represents the configuration related to file-based backend
type FileAuthenticationBackendConfiguration struct {
	Path string `mapstructure:"path"`
}

// AuthenticationBackendConfiguration represents the configuration related to the authentication backend.
type AuthenticationBackendConfiguration struct {
	Ldap *LDAPAuthenticationBackendConfiguration `mapstructure:"ldap"`
	File *FileAuthenticationBackendConfiguration `mapstructure:"file"`
}
