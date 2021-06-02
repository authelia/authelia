package authentication

import "github.com/authelia/authelia/v4"

// UserProvider is the interface for checking user password and
// gathering user details.
type UserProvider interface {
	CheckUserPassword(username string, password string) (bool, error)
	GetDetails(username string) (*authelia.UserDetails, error)
	UpdatePassword(username string, newPassword string) error
}
