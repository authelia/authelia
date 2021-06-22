package session

import (
	"time"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
)

// NewDefaultUserSession create a default user session.
func NewDefaultUserSession() UserSession {
	return UserSession{
		KeepMeLoggedIn:      false,
		AuthenticationLevel: authentication.NotAuthenticated,
		LastActivity:        0,
	}
}

// SetOneFactor sets the expected property values for one factor authentication.
func (s *UserSession) SetOneFactor(now time.Time, details *authentication.UserDetails, keepMeLoggedIn bool) {
	if s.FirstFactorAuthn == 0 {
		s.FirstFactorAuthn = now.Unix()
	}

	s.LastActivity = now.Unix()
	s.AuthenticationLevel = authentication.OneFactor

	s.KeepMeLoggedIn = keepMeLoggedIn

	s.Username = details.Username
	s.DisplayName = details.DisplayName
	s.Groups = details.Groups
	s.Emails = details.Emails
}

// SetTwoFactor sets the expected property values for two factor authentication.
func (s *UserSession) SetTwoFactor(now time.Time) {
	if s.SecondFactorAuthn == 0 {
		s.SecondFactorAuthn = now.Unix()
	}

	s.LastActivity = now.Unix()
	s.AuthenticationLevel = authentication.TwoFactor
}

// AuthenticatedAt returns the unix timestamp this session authenticated successfully at the given level.
func (s UserSession) AuthenticatedAt(level authorization.Level) (authenticatedAt time.Time) {
	switch level {
	case authorization.OneFactor:
		return time.Unix(s.FirstFactorAuthn, 0)
	case authorization.TwoFactor:
		return time.Unix(s.SecondFactorAuthn, 0)
	}

	if s.SecondFactorAuthn != 0 {
		return time.Unix(s.SecondFactorAuthn, 0)
	}

	if s.FirstFactorAuthn != 0 {
		return time.Unix(s.FirstFactorAuthn, 0)
	}

	return time.Now()
}
