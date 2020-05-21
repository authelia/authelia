package session

import (
	"github.com/authelia/authelia/internal/authentication"
)

// NewDefaultUserSession create a default user session.
func NewDefaultUserSession() UserSession {
	return UserSession{
		KeepMeLoggedIn:      false,
		AuthenticationLevel: authentication.NotAuthenticated,
		LastActivity:        0,
	}
}
