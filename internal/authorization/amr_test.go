package authorization_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/authorization"
)

func TestNewAuthenticationMethodsReferencesFromClaim(t *testing.T) {
	testCases := []struct {
		name     string
		have     []string
		expected authorization.AuthenticationMethodsReferences
	}{
		{
			"ShouldHandleWebAuthnSoftware",
			[]string{"pop", "swk", "mca", "mfa", "pwd", "otp"},
			authorization.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnSoftware: true, UsernameAndPassword: true, TOTP: true},
		},
		{
			"ShouldHandleWebAuthnHardware",
			[]string{"pop", "hwk", "mca", "mfa", "pwd", "sms", "user"},
			authorization.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnHardware: true, UsernameAndPassword: true, Duo: true, WebAuthnUserVerified: true},
		},
		{
			"ShouldHandleWebAuthnHardware",
			[]string{"kba", "pop", "hwk", "mca", "mfa", "pwd", "sms", "user"},
			authorization.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnHardware: true, UsernameAndPassword: true, Duo: true, WebAuthnUserVerified: true, KnowledgeBasedAuthentication: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := authorization.NewAuthenticationMethodsReferencesFromClaim(tc.have)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestAuthenticationMethodsReferences(t *testing.T) {
	testCases := []struct {
		name     string
		have     authorization.AuthenticationMethodsReferences
		expected testAMRWant
	}{

		{
			name: "UsernameAndPassword",

			have: authorization.AuthenticationMethodsReferences{UsernameAndPassword: true, KnowledgeBasedAuthentication: true},
			expected: testAMRWant{
				FactorKnowledge:            true,
				FactorPossession:           false,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             true,
				ChannelService:             false,
				MultiChannelAuthentication: false,
				RFC8176:                    []string{"pwd", "kba"},
			},
		},
		{
			name: "TOTP",

			have: authorization.AuthenticationMethodsReferences{TOTP: true},
			expected: testAMRWant{
				FactorKnowledge:            false,
				FactorPossession:           true,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             true,
				ChannelService:             false,
				MultiChannelAuthentication: false,
				RFC8176:                    []string{"otp"},
			},
		},
		{
			name: "WebAuthn",

			have: authorization.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnHardware: true},
			expected: testAMRWant{
				FactorKnowledge:            false,
				FactorPossession:           true,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             true,
				ChannelService:             false,
				MultiChannelAuthentication: false,
				RFC8176:                    []string{"hwk", "pop"},
			},
		},
		{
			name: "WebAuthnSoftware",

			have: authorization.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnSoftware: true},
			expected: testAMRWant{
				FactorKnowledge:            false,
				FactorPossession:           true,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             true,
				ChannelService:             false,
				MultiChannelAuthentication: false,
				RFC8176:                    []string{"swk", "pop"},
			},
		},
		{
			name: "WebAuthnUserPresence",

			have: authorization.AuthenticationMethodsReferences{WebAuthnUserPresence: true},
			expected: testAMRWant{
				FactorKnowledge:            false,
				FactorPossession:           false,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             false,
				ChannelService:             false,
				MultiChannelAuthentication: false,
				RFC8176:                    []string{"user"},
			},
		},
		{
			name: "WebAuthnUserVerified",

			have: authorization.AuthenticationMethodsReferences{WebAuthnUserVerified: true},
			expected: testAMRWant{
				FactorKnowledge:            false,
				FactorPossession:           false,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             false,
				ChannelService:             false,
				MultiChannelAuthentication: false,
				RFC8176:                    []string{"pin"},
			},
		},
		{
			name: "WebAuthnWithUserPresenceAndVerified",

			have: authorization.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnHardware: true, WebAuthnUserVerified: true, WebAuthnUserPresence: true},
			expected: testAMRWant{
				FactorKnowledge:            false,
				FactorPossession:           true,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             true,
				ChannelService:             false,
				MultiChannelAuthentication: false,
				RFC8176:                    []string{"hwk", "pop", "user", "pin"},
			},
		},
		{
			name: "Duo",

			have: authorization.AuthenticationMethodsReferences{Duo: true},
			expected: testAMRWant{
				FactorKnowledge:            false,
				FactorPossession:           true,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             false,
				ChannelService:             true,
				MultiChannelAuthentication: false,
				RFC8176:                    []string{"sms"},
			},
		},
		{
			name: "DuoWebAuthnTOTP",

			have: authorization.AuthenticationMethodsReferences{Duo: true, WebAuthn: true, WebAuthnHardware: true, TOTP: true},
			expected: testAMRWant{
				FactorKnowledge:            false,
				FactorPossession:           true,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             true,
				ChannelService:             true,
				MultiChannelAuthentication: true,
				RFC8176:                    []string{"sms", "hwk", "pop", "otp", "mca"},
			},
		},
		{
			name: "DuoTOTP",

			have: authorization.AuthenticationMethodsReferences{Duo: true, TOTP: true},
			expected: testAMRWant{
				FactorKnowledge:            false,
				FactorPossession:           true,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             true,
				ChannelService:             true,
				MultiChannelAuthentication: true,
				RFC8176:                    []string{"sms", "otp", "mca"},
			},
		},
		{
			name: "UsernameAndPasswordWithDuo",

			have: authorization.AuthenticationMethodsReferences{Duo: true, UsernameAndPassword: true},
			expected: testAMRWant{
				FactorKnowledge:            true,
				FactorPossession:           true,
				MultiFactorAuthentication:  true,
				ChannelBrowser:             true,
				ChannelService:             true,
				MultiChannelAuthentication: true,
				RFC8176:                    []string{"pwd", "sms", "mfa", "mca"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected.FactorKnowledge, tc.have.FactorKnowledge())
			assert.Equal(t, tc.expected.FactorPossession, tc.have.FactorPossession())
			assert.Equal(t, tc.expected.MultiFactorAuthentication, tc.have.MultiFactorAuthentication())
			assert.Equal(t, tc.expected.ChannelBrowser, tc.have.ChannelBrowser())
			assert.Equal(t, tc.expected.ChannelService, tc.have.ChannelService())
			assert.Equal(t, tc.expected.MultiChannelAuthentication, tc.have.MultiChannelAuthentication())

			isRFC8176 := tc.have.MarshalRFC8176()

			for _, amr := range tc.expected.RFC8176 {
				t.Run(fmt.Sprintf("has all wanted/%s", amr), func(t *testing.T) {
					assert.Contains(t, isRFC8176, amr)
				})
			}

			for _, amr := range isRFC8176 {
				t.Run(fmt.Sprintf("only has wanted/%s", amr), func(t *testing.T) {
					assert.Contains(t, tc.expected.RFC8176, amr)
				})
			}
		})
	}
}

type testAMRWant struct {
	FactorKnowledge, FactorPossession, MultiFactorAuthentication bool
	ChannelBrowser, ChannelService, MultiChannelAuthentication   bool

	RFC8176 []string
}
