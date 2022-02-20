package authentication

import (
	"github.com/authelia/authelia/v4/internal/model"
)

// UserProvider is the interface for checking user password and
// gathering user details.
type UserProvider interface {
	model.StartupCheck

	CheckUserPassword(username string, password string) (valid bool, err error)
	GetDetails(username string) (details *model.UserDetails, err error)

	// GetCurrentDetails is a special method that bypasses any caching. This should only be used where necessary.
	// Use GetDetails with the same signature instead.
	GetCurrentDetails(username string) (details *model.UserDetails, err error)
	UpdatePassword(username string, newPassword string) (err error)
}
