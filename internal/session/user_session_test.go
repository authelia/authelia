package session

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
)

func TestUserSession_SetFactors(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(session *UserSession)
		expect *UserSession
	}{
		{
			"ShouldSetOneFactorPassword",
			func(session *UserSession) {
				*session = NewUserSession("john")
				session.SetOneFactorPassword(time.Unix(10000, 0), true)
			},
			&UserSession{
				Username:                  "john",
				KeepMeLoggedIn:            true,
				LastActivity:              10000,
				FirstFactorAuthnTimestamp: 10000,
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					UsernameAndPassword:          true,
				},
			},
		},
		{
			"ShouldSetOneFactorPasskey",
			func(session *UserSession) {
				*session = NewUserSession("john")
				session.SetOneFactorPasskey(time.Unix(10000, 0), true, true, true, true)
			},
			&UserSession{
				Username:                  "john",
				KeepMeLoggedIn:            true,
				LastActivity:              10000,
				FirstFactorAuthnTimestamp: 10000,
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					WebAuthn:             true,
					WebAuthnHardware:     true,
					WebAuthnUserVerified: true,
					WebAuthnUserPresence: true,
				},
			},
		},
		{
			"ShouldSetTwoFactorPassword",
			func(session *UserSession) {
				*session = NewUserSession("john")
				session.SetOneFactorPasskey(time.Unix(10000, 0), true, true, true, true)
				session.SetTwoFactorPassword(time.Unix(20000, 0))
			},
			&UserSession{
				Username:                   "john",
				KeepMeLoggedIn:             true,
				LastActivity:               20000,
				FirstFactorAuthnTimestamp:  10000,
				SecondFactorAuthnTimestamp: 20000,
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					UsernameAndPassword:          true,
					WebAuthn:                     true,
					WebAuthnHardware:             true,
					WebAuthnUserVerified:         true,
					WebAuthnUserPresence:         true,
				},
			},
		},
		{
			"ShouldSetOneFactorPasswordAndTwoFactorDuo",
			func(session *UserSession) {
				*session = NewUserSession("john")
				session.SetOneFactorPassword(time.Unix(10000, 0), true)
				session.SetTwoFactorDuo(time.Unix(20000, 0))
			},
			&UserSession{
				Username:                   "john",
				KeepMeLoggedIn:             true,
				LastActivity:               20000,
				FirstFactorAuthnTimestamp:  10000,
				SecondFactorAuthnTimestamp: 20000,
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					UsernameAndPassword:          true,
					Duo:                          true,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			session := &UserSession{}

			tc.setup(session)

			assert.Equal(t, tc.expect, session)
		})
	}
}

func TestUserSession_AuthenticationLevel(t *testing.T) {
	testCases := []struct {
		name     string
		have     *UserSession
		passkey  bool
		expected authentication.Level
	}{
		{
			"ShouldHandleAnonymous",
			&UserSession{},
			false,
			authentication.NotAuthenticated,
		},
		{
			"ShouldHandleTwoFactorTOTP",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					TOTP:                         true,
				},
			},
			false,
			authentication.TwoFactor,
		},
		{
			"ShouldHandleTwoFactorWebAuthn",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					WebAuthn:                     true,
				},
			},
			false,
			authentication.TwoFactor,
		},
		{
			"ShouldHandleTwoFactorDuo",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					Duo:                          true,
				},
			},
			false,
			authentication.TwoFactor,
		},
		{
			"ShouldHandleTwoFactorWebAuthnPasskey",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					KnowledgeBasedAuthentication: true,
					WebAuthn:                     true,
				},
			},
			true,
			authentication.TwoFactor,
		},
		{
			"ShouldHandleTwoFactorWebAuthnPasskeyWithoutKnowledge",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					WebAuthn: true,
				},
			},
			true,
			authentication.OneFactor,
		},
		{
			"ShouldHandleTwoFactorWebAuthnPasskeyWithoutKnowledgeWithUserVerification",
			&UserSession{
				Username: "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					WebAuthn:             true,
					WebAuthnUserVerified: true,
				},
			},
			true,
			authentication.TwoFactor,
		},
		{
			"ShouldHandleNoAMR",
			&UserSession{
				Username:                 "john",
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{},
			},
			true,
			authentication.NotAuthenticated,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.have.AuthenticationLevel(tc.passkey))
		})
	}
}

func TestUserSession_SetTwoFactorWebAuthn(t *testing.T) {
	testCases := []struct {
		name                         string
		at                           time.Time
		hardware, presence, verified bool
		expected                     authorization.AuthenticationMethodsReferences
	}{
		{
			"ShouldHandleHardware",
			time.Unix(1000, 0),
			true,
			true,
			true,
			authorization.AuthenticationMethodsReferences{
				WebAuthn:             true,
				WebAuthnHardware:     true,
				WebAuthnUserPresence: true,
				WebAuthnUserVerified: true,
			},
		},
		{
			"ShouldHandleSoftware",
			time.Unix(1000, 0),
			false,
			true,
			true,
			authorization.AuthenticationMethodsReferences{
				WebAuthn:             true,
				WebAuthnSoftware:     true,
				WebAuthnUserPresence: true,
				WebAuthnUserVerified: true,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := &UserSession{}

			actual.SetTwoFactorWebAuthn(tc.at, tc.hardware, tc.presence, tc.verified)

			assert.Equal(t, tc.expected, actual.AuthenticationMethodRefs)
		})
	}
}

func TestUserSession_Misc(t *testing.T) {
	session := &UserSession{}

	assert.True(t, session.IsAnonymous())
}
