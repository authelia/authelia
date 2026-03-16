package authentication

type UserManagementProvider interface {
	AddUser(userData *UserDetailsExtended) (err error)
	UpdateUserWithMask(username string, userData *UserDetailsExtended, updateMask []string) (err error)
	DeleteUser(username string) (err error)

	AddGroup(newGroup string) error
	DeleteGroup(group string) (err error)
	ListGroups() ([]string, error)

	GetRequiredAttributes() []string
	GetSupportedAttributes() map[string]UserManagementAttributeMetadata
	ValidateUserData(userData *UserDetailsExtended) error
	ValidatePartialUpdate(userData *UserDetailsExtended, updateMask []string) error
}
