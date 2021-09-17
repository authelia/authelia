package authentication

import (
	"github.com/sirupsen/logrus"
)

// UserProvider is the interface for checking user password and
// gathering user details.
type UserProvider interface {
	CheckUserPassword(username string, password string) (valid bool, err error)
	GetDetails(username string) (details *UserDetails, err error)
	UpdatePassword(username string, newPassword string) (err error)
	StartupCheck(logger *logrus.Logger) (err error)
}
