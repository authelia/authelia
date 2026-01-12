package schema

type Administration struct {
	Enabled                bool   `koanf:"enabled" json:"enabled" jsonschema:"default=false,title=Enabled" jsonschema_description:"Enables administration interface features."`
	AdminGroup             string `koanf:"admin_group" json:"admin_group" jsonschema:"default=authelia-admin,title=Admin Group" jsonschema_description:"The Group that will allow a user to access authelia administration features."`
	EnableUserManagement   bool   `koanf:"enable_user_management" json:"enable_user_management" jsonschema:"default=false,title=Enable User Management"  jsonschema_description:"Enables the user management interface for users with the admin_group group."`
	AllowAdminsToAddAdmins bool   `koanf:"allow_admins_to_add_admins" json:"allow_admins_to_add_admins" jsonschema:"default=false,title=Allow admins to create other admins"  jsonschema_description:"Allows admins with the admin group to give the admin group to other users."`
}

var DefaultAdministrationConfiguration = Administration{
	Enabled:                false,
	AdminGroup:             "authelia-admin",
	EnableUserManagement:   false,
	AllowAdminsToAddAdmins: false,
}
