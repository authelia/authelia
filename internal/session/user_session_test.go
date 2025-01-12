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
				session.SetOneFactorPassword(time.Unix(10000, 0), &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"abc", "123"}}, true)
			},
			&UserSession{
				Username:                  "john",
				Groups:                    []string{"abc", "123"},
				Emails:                    []string{"john@example.com"},
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
				session.SetOneFactorPasskey(time.Unix(10000, 0), &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"abc", "123"}}, true, true, true, true)
			},
			&UserSession{
				Username:                  "john",
				Groups:                    []string{"abc", "123"},
				Emails:                    []string{"john@example.com"},
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
				session.SetOneFactorPasskey(time.Unix(10000, 0), &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"abc", "123"}}, true, true, true, true)
				session.SetTwoFactorPassword(time.Unix(20000, 0))
			},
			&UserSession{
				Username:                   "john",
				Groups:                     []string{"abc", "123"},
				Emails:                     []string{"john@example.com"},
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
				session.SetOneFactorPassword(time.Unix(10000, 0), &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"abc", "123"}}, true)
				session.SetTwoFactorDuo(time.Unix(20000, 0))
			},
			&UserSession{
				Username:                   "john",
				Groups:                     []string{"abc", "123"},
				Emails:                     []string{"john@example.com"},
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

	assert.Equal(t, Identity{}, session.Identity())
	assert.Equal(t, "", session.GetUsername())
	assert.Equal(t, "", session.GetDisplayName())
	assert.Nil(t, session.GetGroups())
	assert.Nil(t, session.GetEmails())
	assert.True(t, session.IsAnonymous())

	session.Username = "abc"

	assert.Equal(t, Identity{Username: "abc"}, session.Identity())
	assert.Equal(t, "abc", session.GetUsername())

	session.DisplayName = "A B C"

	assert.Equal(t, Identity{Username: "abc", DisplayName: "A B C"}, session.Identity())
	assert.Equal(t, "abc", session.GetUsername())
	assert.Equal(t, "A B C", session.GetDisplayName())

	session.Emails = []string{"abc@example.com", "xyz@example.com"}

	assert.Equal(t, Identity{Username: "abc", DisplayName: "A B C", Email: "abc@example.com"}, session.Identity())
	assert.Equal(t, "abc", session.GetUsername())
	assert.Equal(t, "A B C", session.GetDisplayName())
	assert.Equal(t, []string{"abc@example.com", "xyz@example.com"}, session.GetEmails())

	session.Groups = []string{"agroup", "bgroup"}
	assert.Equal(t, "abc", session.GetUsername())
	assert.Equal(t, "A B C", session.GetDisplayName())
	assert.Equal(t, []string{"abc@example.com", "xyz@example.com"}, session.GetEmails())
	assert.Equal(t, []string{"agroup", "bgroup"}, session.GetGroups())
}
