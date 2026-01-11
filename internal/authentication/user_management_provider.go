package authentication

type UserManagementProvider interface {
	AddUser(userData *UserDetailsExtended) (err error)
	UpdateUser(username string, userData *UserDetailsExtended) (err error)
	UpdateUserWithMask(username string, userData *UserDetailsExtended, updateMask []string) (err error)
	DeleteUser(username string) (err error)

	GetRequiredFields() []string
	GetSupportedFields() []string
	GetFieldMetadata() map[string]FieldMetadata
	ValidateUserData(userData *UserDetailsExtended) error
	ValidatePartialUpdate(userData *UserDetailsExtended, updateMask []string) error
}
