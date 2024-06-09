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
	return s.AuthenticationLevel() == authentication.NotAuthenticated
}

func (s *UserSession) AuthenticationLevel() authentication.Level {
	switch {
	case s.Username == "":
		return authentication.NotAuthenticated
	case s.AuthenticationMethodRefs.FactorPossession() && s.AuthenticationMethodRefs.FactorKnowledge():
		return authentication.TwoFactor
	case s.AuthenticationMethodRefs.FactorPossession() || s.AuthenticationMethodRefs.FactorKnowledge():
		return authentication.OneFactor
	default:
		return authentication.NotAuthenticated
	}
}

// SetOneFactorPassword sets the 1FA AMR's and expected property values for one factor password authentication.
func (s *UserSession) SetOneFactorPassword(now time.Time, details *authentication.UserDetails, keepMeLoggedIn bool) {
	s.setOneFactor(now, details, keepMeLoggedIn)

	s.AuthenticationMethodRefs.KnowledgeBasedAuthentication = true
	s.AuthenticationMethodRefs.UsernameAndPassword = true
}

// SetOneFactorPasskey sets the 1FA AMR's and expected property values for one factor passkey authentication.
func (s *UserSession) SetOneFactorPasskey(now time.Time, details *authentication.UserDetails, keepMeLoggedIn, hardware, userPresence, userVerified bool) {
	s.setOneFactor(now, details, keepMeLoggedIn)

	s.setWebAuthn(hardware, userPresence, userVerified)
}

func (s *UserSession) setOneFactor(now time.Time, details *authentication.UserDetails, keepMeLoggedIn bool) {
	s.FirstFactorAuthnTimestamp = now.Unix()
	s.LastActivity = now.Unix()

	s.KeepMeLoggedIn = keepMeLoggedIn

	s.Username = details.Username
	s.DisplayName = details.DisplayName
	s.Groups = details.Groups
	s.Emails = details.Emails
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

// AuthenticatedTime returns the unix timestamp this session authenticated successfully at the given level.
func (s *UserSession) AuthenticatedTime(level authorization.Level) (authenticatedTime time.Time, err error) {
	switch level {
	case authorization.OneFactor:
		return time.Unix(s.FirstFactorAuthnTimestamp, 0).UTC(), nil
	case authorization.TwoFactor:
		return time.Unix(s.SecondFactorAuthnTimestamp, 0).UTC(), nil
	default:
		return time.Unix(0, 0).UTC(), errors.New("invalid authorization level")
	}
}

// Identity value of the user session.
func (s *UserSession) Identity() Identity {
	identity := Identity{
		Username:    s.Username,
		DisplayName: s.DisplayName,
	}

	if len(s.Emails) != 0 {
		identity.Email = s.Emails[0]
	}

	return identity
}

func (s *UserSession) GetUsername() (username string) {
	return s.Username
}

func (s *UserSession) GetGroups() (groups []string) {
	return s.Groups
}

func (s *UserSession) GetDisplayName() (name string) {
	return s.DisplayName
}

func (s *UserSession) GetEmails() (emails []string) {
	return s.Emails
}
