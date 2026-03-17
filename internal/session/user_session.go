package session

import (
	"errors"
	"time"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
)

// NewDefaultUserSession create a default user session.
func NewDefaultUserSession() UserSession {
	return UserSession{
		KeepMeLoggedIn: false,
		LastActivity:   0,
	}
}

// IsAnonymous returns true if the username is empty or the AuthenticationLevel is authentication.NotAuthenticated.
func (s *UserSession) IsAnonymous() bool {
	return s.AuthenticationLevel(false) == authentication.NotAuthenticated
}

// AuthenticationLevel returns the authentication.Level for this session.
func (s *UserSession) AuthenticationLevel(passkey2FA bool) authentication.Level {
	switch {
	case s.Username == "":
		return authentication.NotAuthenticated
	case s.AuthenticationMethodRefs.FactorPossession() && s.AuthenticationMethodRefs.FactorKnowledge():
		return authentication.TwoFactor
	case passkey2FA && s.AuthenticationMethodRefs.WebAuthn && s.AuthenticationMethodRefs.WebAuthnUserVerified:
		return authentication.TwoFactor
	case s.AuthenticationMethodRefs.FactorPossession() || s.AuthenticationMethodRefs.FactorKnowledge():
		return authentication.OneFactor
	default:
		return authentication.NotAuthenticated
	}
}

// SetOneFactorPassword sets the 1FA AMR's and expected property values for one factor password authentication.
func (s *UserSession) SetOneFactorPassword(now time.Time, username string, keepMeLoggedIn bool) {
	s.setOneFactor(now, username, keepMeLoggedIn)

	s.AuthenticationMethodRefs.KnowledgeBasedAuthentication = true
	s.AuthenticationMethodRefs.UsernameAndPassword = true
}

// SetOneFactorPasskey sets the 1FA AMR's and expected property values for one factor passkey authentication.
func (s *UserSession) SetOneFactorPasskey(now time.Time, username string, keepMeLoggedIn, hardware, userPresence, userVerified bool) {
	s.setOneFactor(now, username, keepMeLoggedIn)

	s.setWebAuthn(hardware, userPresence, userVerified)
}

func (s *UserSession) setOneFactor(now time.Time, username string, keepMeLoggedIn bool) {
	s.FirstFactorAuthnTimestamp = now.Unix()
	s.LastActivity = now.Unix()

	s.KeepMeLoggedIn = keepMeLoggedIn
	s.Username = username

	s.SetOneFactorReauthenticate(now)
}

// SetOneFactorReauthenticate sets the relevant session values when a user reauthenticates with the first factor.
func (s *UserSession) SetOneFactorReauthenticate(now time.Time) {
	s.FirstFactorAuthnTimestamp = now.Unix()
	s.LastActivity = now.Unix()
}

// SetTwoFactorTOTP sets the relevant TOTP AMR's and sets the factor to 2FA.
func (s *UserSession) SetTwoFactorTOTP(now time.Time) {
	s.setTwoFactor(now)
	s.AuthenticationMethodRefs.TOTP = true
}

// SetTwoFactorDuo sets the relevant Duo AMR's and sets the factor to 2FA.
func (s *UserSession) SetTwoFactorDuo(now time.Time) {
	s.setTwoFactor(now)
	s.AuthenticationMethodRefs.Duo = true
}

// SetTwoFactorWebAuthn sets the relevant WebAuthn AMR's and sets the factor to 2FA.
func (s *UserSession) SetTwoFactorWebAuthn(now time.Time, hardware, userPresence, userVerified bool) {
	s.setTwoFactor(now)

	s.setWebAuthn(hardware, userPresence, userVerified)
}

// SetTwoFactorPassword sets the relevant session values when a user authenticates with a password as the second factor.
func (s *UserSession) SetTwoFactorPassword(now time.Time) {
	s.setTwoFactor(now)

	s.AuthenticationMethodRefs.KnowledgeBasedAuthentication = true
	s.AuthenticationMethodRefs.UsernameAndPassword = true
}

func (s *UserSession) setTwoFactor(now time.Time) {
	s.SecondFactorAuthnTimestamp = now.Unix()
	s.LastActivity = now.Unix()
}

func (s *UserSession) setWebAuthn(hardware, userPresence, userVerified bool) {
	s.AuthenticationMethodRefs.WebAuthn = true
	s.AuthenticationMethodRefs.WebAuthnUserPresence, s.AuthenticationMethodRefs.WebAuthnUserVerified = userPresence, userVerified

	if hardware {
		s.AuthenticationMethodRefs.WebAuthnHardware = true
	} else {
		s.AuthenticationMethodRefs.WebAuthnSoftware = true
	}

	s.WebAuthn = nil
}

// GetFirstFactorAuthn returns the time the first factor authentication occurred.
func (s *UserSession) GetFirstFactorAuthn() time.Time {
	return time.Unix(s.FirstFactorAuthnTimestamp, 0).UTC()
}

// GetSecondFactorAuthn returns the time the second factor authentication occurred.
func (s *UserSession) GetSecondFactorAuthn() time.Time {
	return time.Unix(s.SecondFactorAuthnTimestamp, 0).UTC()
}

// AuthenticatedTime returns the unix timestamp this session authenticated successfully at the given level.
func (s *UserSession) AuthenticatedTime(level authorization.Level) (authenticatedTime time.Time, err error) {
	switch level {
	case authorization.OneFactor:
		return s.GetFirstFactorAuthn(), nil
	case authorization.TwoFactor:
		return s.GetSecondFactorAuthn(), nil
	default:
		return time.Unix(0, 0).UTC(), errors.New("invalid authorization level")
	}
}

// LastAuthenticatedTime returns the most recent time either factor was authenticated.
func (s *UserSession) LastAuthenticatedTime() (authenticated time.Time) {
	if s.FirstFactorAuthnTimestamp > s.SecondFactorAuthnTimestamp {
		return s.GetFirstFactorAuthn()
	}

	return s.GetSecondFactorAuthn()
}
