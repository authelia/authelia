package schema

// LDAPAuthenticationBackendConfiguration represents the configuration related to LDAP server.
type LDAPAuthenticationBackendConfiguration struct {
	URL                string `yaml:"url"`
	BaseDN             string `yaml:"base_dn"`
	AdditionalUsersDN  string `yaml:"additional_users_dn"`
	UsersFilter        string `yaml:"users_filter"`
	AdditionalGroupsDN string `yaml:"additional_groups_dn"`
	GroupsFilter       string `yaml:"groups_filter"`
	GroupNameAttribute string `yaml:"group_name_attribute"`
	MailAttribute      string `yaml:"mail_attribute"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
}

// FileAuthenticationBackendConfiguration represents the configuration related to file-based backend
type FileAuthenticationBackendConfiguration struct {
	Path string `yaml:"path"`
}

// AuthenticationBackendConfiguration represents the configuration related to the authentication backend.
type AuthenticationBackendConfiguration struct {
	Ldap *LDAPAuthenticationBackendConfiguration `yaml:"ldap"`
	File *FileAuthenticationBackendConfiguration `yaml:"file"`
}
