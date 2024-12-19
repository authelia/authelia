package authentication

import (
	"github.com/authelia/authelia/v4/internal/model"
)

// UserProvider is the interface for checking user password and
// gathering user details.
type UserProvider interface {
	model.StartupCheck

	// AddUser adds a user given the new user's information.
	AddUser(username, displayname, password string, opts ...func(options *NewUserDetailsOpts)) (err error)

	// DeleteUser deletes user given the username.
	DeleteUser(username string) (err error)

	// CheckUserPassword checks if provided password matches for the given user.
	CheckUserPassword(username, password string) (valid bool, err error)

	// GetDetails retrieve the details for a user.
	GetDetails(username string) (details *UserDetails, err error)

	// UpdatePassword updates the password of the given user.
	UpdatePassword(username, newPassword string) (err error)

	// ChangePassword validates the old password then changes the password of the given user.
	ChangePassword(username, oldPassword, newPassword string) (err error)

	// ChangeDisplayName changes the display name for a specific user.
	ChangeDisplayName(username, newDisplayName string) (err error)

	// ChangeEmail changes the email for a specific user.
	ChangeEmail(username, newEmail string) (err error)

	// ChangeGroups changes the groups for a specific user.
	ChangeGroups(username string, newGroups []string) (err error)

	// ListUsers returns a list of all users and their attributes.
	ListUsers() (userList []UserDetails, err error)
}
