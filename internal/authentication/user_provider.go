package authentication

import (
	"github.com/authelia/authelia/v4/internal/model"
)

// UserProvider is the interface for interacting with the authentication backends.
type UserProvider interface {
	model.StartupCheck

	// CheckUserPassword is used to check if a password matches for a specific user.
	CheckUserPassword(username string, password string) (valid bool, err error)

	// GetDetails is used to get a user's information.
	GetDetails(username string) (details *UserDetails, err error)

	// UpdatePassword is used to change a user's password without verifying their old password.
	UpdatePassword(username string, newPassword string) (err error)

	// ChangePassword is used to change a user's password but requires their old password to be successfully verified.
	ChangePassword(username string, oldPassword string, newPassword string) (err error)

	GetUser(username string) (details *UserDetailsExtended, err error)

	GetDetailsExtended(username string) (details *UserDetailsExtended, err error)

	ListUsers() (userList []UserDetailsExtended, err error)

	AddUser(userData *UserDetailsExtended) (err error)
	UpdateUser(username string, userData *UserDetailsExtended) (err error)
	UpdateUserWithMask(username string, userData *UserDetailsExtended, updateMask []string) error
	DeleteUser(username string) (err error)

	AddGroup(newGroup string) (err error)
	DeleteGroup(group string) (err error)
	ListGroups() ([]string, error)

	GetRequiredFields() []string
	GetSupportedFields() []string
	GetFieldMetadata() map[string]FieldMetadata
	ValidateUserData(userData *UserDetailsExtended) error
	ValidatePartialUpdate(userData *UserDetailsExtended, updateMask []string) error

	Close() (err error)
}
