package session

import (
	"errors"
	"time"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/utils"
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
	s.FirstFactorAuthnTimestamp = now.Unix()
	s.LastActivity = now.Unix()
	s.AuthenticationLevel = authentication.OneFactor

	s.KeepMeLoggedIn = keepMeLoggedIn

	s.Username = details.Username
	s.DisplayName = details.DisplayName
	s.Groups = details.Groups
	s.Emails = details.Emails

	s.AuthenticationMethodReferences = append(s.AuthenticationMethodReferences, oidc.AMRPasswordBasedAuthentication)

	if utils.IsStringSliceContainsAny([]string{oidc.AMROneTimePassword, oidc.AMRHardwareSecuredKey}, s.AuthenticationMethodReferences) {
		s.AuthenticationMethodReferences = append(s.AuthenticationMethodReferences, oidc.AMRMultiFactorAuthentication)
	}
}

// SetTwoFactor sets the expected property values for two factor authentication.
func (s *UserSession) SetTwoFactor(now time.Time) {
	s.SecondFactorAuthnTimestamp = now.Unix()
	s.LastActivity = now.Unix()
	s.AuthenticationLevel = authentication.TwoFactor

	if utils.IsStringInSlice(oidc.AMRPasswordBasedAuthentication, s.AuthenticationMethodReferences) {
		s.AuthenticationMethodReferences = append(s.AuthenticationMethodReferences, oidc.AMRMultiFactorAuthentication)
	}
}

func (s *UserSession) SetTwoFactorTOTP(now time.Time) {
	s.SetTwoFactor(now)

	s.AuthenticationMethodReferences = append(s.AuthenticationMethodReferences, oidc.AMROneTimePassword)
}

func (s *UserSession) SetTwoFactorDuo(now time.Time) {
	s.SetTwoFactor(now)

	if utils.IsStringSliceContainsAny(
		[]string{oidc.AMRPasswordBasedAuthentication, oidc.AMRHardwareSecuredKey, oidc.AMROneTimePassword},
		s.AuthenticationMethodReferences) {
		s.AuthenticationMethodReferences = append(s.AuthenticationMethodReferences, oidc.AMRMultiChannelAuthentication)
	}
}

func (s *UserSession) SetTwoFactorWebauthn(now time.Time, userPresence, userVerified bool) {
	s.SetTwoFactor(now)

	s.AuthenticationMethodReferences = append(s.AuthenticationMethodReferences, oidc.AMRHardwareSecuredKey)

	if userPresence {
		s.AuthenticationMethodReferences = append(s.AuthenticationMethodReferences, oidc.AMRUserPresence)
	}

	if userVerified {
		s.AuthenticationMethodReferences = append(s.AuthenticationMethodReferences, oidc.AMRPersonalIdentificationNumber)
	}

	s.Webauthn = nil
}

// AuthenticatedTime returns the unix timestamp this session authenticated successfully at the given level.
func (s UserSession) AuthenticatedTime(level authorization.Level) (authenticatedTime time.Time, err error) {
	switch level {
	case authorization.OneFactor:
		return time.Unix(s.FirstFactorAuthnTimestamp, 0), nil
	case authorization.TwoFactor:
		return time.Unix(s.SecondFactorAuthnTimestamp, 0), nil
	default:
		return time.Unix(0, 0), errors.New("invalid authorization level")
	}
}
