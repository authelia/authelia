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
			"ShouldSetOneFactorOpenIDConnectWithOnlyTheFederatedIdentityReference",
			func(session *UserSession) {
				session.SetOneFactorOpenIDConnect(time.Unix(10000, 0), &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"abc", "123"}}, true)
			},
			&UserSession{
				Username:                  "john",
				Groups:                    []string{"abc", "123"},
				Emails:                    []string{"john@example.com"},
				KeepMeLoggedIn:            true,
				LastActivity:              10000,
				FirstFactorAuthnTimestamp: 10000,
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					FederatedIdentity: true,
				},
			},
		},
		{
			"ShouldSetOneFactorOpenIDConnectAndMergeTrustedAuthenticationMethodReferences",
			func(session *UserSession) {
				session.SetOneFactorOpenIDConnect(time.Unix(10000, 0), &authentication.UserDetails{Username: "john", Emails: []string{"john@example.com"}, Groups: []string{"abc", "123"}}, true)
				session.AuthenticationMethodRefs = session.AuthenticationMethodRefs.Merge(authorization.NewAuthenticationMethodsReferencesFromClaim([]string{"pwd", "otp"}))
			},
			&UserSession{
				Username:                  "john",
				Groups:                    []string{"abc", "123"},
				Emails:                    []string{"john@example.com"},
				KeepMeLoggedIn:            true,
				LastActivity:              10000,
				FirstFactorAuthnTimestamp: 10000,
				AuthenticationMethodRefs: authorization.AuthenticationMethodsReferences{
					FederatedIdentity:   true,
					UsernameAndPassword: true,
					TOTP:                true,
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

func TestUserSession_SetOneFactorOpenIDConnectAuthenticationLevel(t *testing.T) {
	testCases := []struct {
		Name     string
		Claim    []string
		Trust    bool
		Expected authentication.Level
	}{
		{
			Name:     "ShouldBeOneFactorWithoutTrustedAuthenticationMethodReferences",
			Claim:    []string{"pwd"},
			Trust:    false,
			Expected: authentication.OneFactor,
		},
		{
			Name:     "ShouldBeOneFactorWithTrustedKnowledgeFactor",
			Claim:    []string{"pwd"},
			Trust:    true,
			Expected: authentication.OneFactor,
		},
		{
			Name:     "ShouldBeTwoFactorWithTrustedKnowledgeAndPossessionFactors",
			Claim:    []string{"pwd", "otp"},
			Trust:    true,
			Expected: authentication.TwoFactor,
		},
		{
			Name:     "ShouldBeOneFactorWithTrustedButUnmappedAuthenticationMethodReferences",
			Claim:    []string{"face"},
			Trust:    true,
			Expected: authentication.OneFactor,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			s := &UserSession{}

			s.SetOneFactorOpenIDConnect(time.Unix(10000, 0), &authentication.UserDetails{Username: "john"}, false)

			if tc.Trust {
				s.AuthenticationMethodRefs = s.AuthenticationMethodRefs.Merge(authorization.NewAuthenticationMethodsReferencesFromClaim(tc.Claim))
			}

			assert.Equal(t, tc.Expected, s.AuthenticationLevel(false))
			assert.False(t, s.IsAnonymous())
		})
	}
}

func TestUserSession_FederatedIdentitySecondFactor(t *testing.T) {
	testCases := []struct {
		Name     string
		Setup    func(s *UserSession)
		Second   func(s *UserSession)
		Expected authentication.Level
		RFC8176  []string
	}{
		{
			Name:     "ShouldBeOneFactorWithFederatedIdentityAlone",
			Setup:    nil,
			Second:   nil,
			Expected: authentication.OneFactor,
			RFC8176:  nil,
		},
		{
			Name:     "ShouldBeTwoFactorWithFederatedIdentityAndTOTP",
			Setup:    nil,
			Second:   func(s *UserSession) { s.SetTwoFactorTOTP(time.Unix(20000, 0)) },
			Expected: authentication.TwoFactor,
			RFC8176:  []string{"otp"},
		},
		{
			Name:     "ShouldBeTwoFactorWithFederatedIdentityAndDuo",
			Setup:    nil,
			Second:   func(s *UserSession) { s.SetTwoFactorDuo(time.Unix(20000, 0)) },
			Expected: authentication.TwoFactor,
			RFC8176:  []string{"sms"},
		},
		{
			Name:     "ShouldBeTwoFactorWithFederatedIdentityAndWebAuthn",
			Setup:    nil,
			Second:   func(s *UserSession) { s.SetTwoFactorWebAuthn(time.Unix(20000, 0), true, true, false) },
			Expected: authentication.TwoFactor,
			RFC8176:  []string{"pop", "hwk", "user"},
		},
		{
			Name:     "ShouldRemainOneFactorWithFederatedIdentityAndAKnowledgeFactor",
			Setup:    nil,
			Second:   func(s *UserSession) { s.SetTwoFactorPassword(time.Unix(20000, 0)) },
			Expected: authentication.OneFactor,
			RFC8176:  []string{"pwd", "kba"},
		},
		{
			Name: "ShouldBeTwoFactorWithPasswordAndTOTP",
			Setup: func(s *UserSession) {
				s.SetOneFactorPassword(time.Unix(10000, 0), &authentication.UserDetails{Username: "john"}, false)
			},
			Second:   func(s *UserSession) { s.SetTwoFactorTOTP(time.Unix(20000, 0)) },
			Expected: authentication.TwoFactor,
			RFC8176:  []string{"pwd", "kba", "otp", "mfa"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			s := &UserSession{}

			if tc.Setup != nil {
				tc.Setup(s)
			} else {
				s.SetOneFactorOpenIDConnect(time.Unix(10000, 0), &authentication.UserDetails{Username: "john"}, false)
			}

			if tc.Second != nil {
				tc.Second(s)
			}

			assert.Equal(t, tc.Expected, s.AuthenticationLevel(false))
			assert.Equal(t, tc.RFC8176, s.AuthenticationMethodRefs.MarshalRFC8176())
		})
	}
}

func TestUserSession_FederatedIdentityIsNotAFactor(t *testing.T) {
	s := &UserSession{}

	s.SetOneFactorOpenIDConnect(time.Unix(10000, 0), &authentication.UserDetails{Username: "john"}, false)

	assert.True(t, s.AuthenticationMethodRefs.FederatedIdentity)
	assert.False(t, s.AuthenticationMethodRefs.FactorKnowledge())
	assert.False(t, s.AuthenticationMethodRefs.FactorPossession())
	assert.False(t, s.AuthenticationMethodRefs.MultiFactorAuthentication())
	assert.Empty(t, s.AuthenticationMethodRefs.MarshalRFC8176())
	assert.Equal(t, authentication.OneFactor, s.AuthenticationLevel(false))
	assert.Equal(t, authentication.OneFactor, s.AuthenticationLevel(true))
	assert.False(t, s.IsAnonymous())
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
