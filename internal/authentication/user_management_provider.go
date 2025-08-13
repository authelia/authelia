package authentication

type UserManagementProvider interface {
	AddUser(userData *UserDetailsExtended) (err error)
	UpdateUser(username string, userData *UserDetailsExtended) (err error)
	DeleteUser(username string) (err error)

	GetRequiredFields() []string
	GetSupportedFields() []string
	ValidateUserData(userData *UserDetailsExtended) error
}
