package oidc_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewAuthenticationMethodsReferencesFromClaim(t *testing.T) {
	testCases := []struct {
		name     string
		have     []string
		expected oidc.AuthenticationMethodsReferences
	}{
		{
			"ShouldHandleWebAuthnSoftware",
			[]string{"pop", "swk", "mca", "mfa", "pwd", "otp"},
			oidc.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnSoftware: true, UsernameAndPassword: true, TOTP: true},
		},
		{
			"ShouldHandleWebAuthnHardware",
			[]string{"pop", "hwk", "mca", "mfa", "pwd", "sms", "user"},
			oidc.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnHardware: true, UsernameAndPassword: true, Duo: true, WebAuthnUserVerified: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := oidc.NewAuthenticationMethodsReferencesFromClaim(tc.have)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestAuthenticationMethodsReferences(t *testing.T) {
	testCases := []struct {
		desc string
		is   oidc.AuthenticationMethodsReferences
		want testAMRWant
	}{

		{
			desc: "Username and Password",

			is: oidc.AuthenticationMethodsReferences{UsernameAndPassword: true},
			want: testAMRWant{
				FactorKnowledge:            true,
				FactorPossession:           false,
				MultiFactorAuthentication:  false,
				ChannelBrowser:             true,
				ChannelService:             false,
				MultiChannelAuthentication: false,
				RFC8176:                    []string{"pwd"},
			},
		},
		{
			desc: "TOTP",

			is: oidc.AuthenticationMethodsReferences{TOTP: true},
			want: testAMRWant{
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
			desc: "WebAuthn",

			is: oidc.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnHardware: true},
			want: testAMRWant{
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
			desc: "WebAuthnSoftware",

			is: oidc.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnSoftware: true},
			want: testAMRWant{
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
			desc: "WebAuthn User Presence",

			is: oidc.AuthenticationMethodsReferences{WebAuthnUserPresence: true},
			want: testAMRWant{
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
			desc: "WebAuthn User Verified",

			is: oidc.AuthenticationMethodsReferences{WebAuthnUserVerified: true},
			want: testAMRWant{
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
			desc: "WebAuthn with User Presence and Verified",

			is: oidc.AuthenticationMethodsReferences{WebAuthn: true, WebAuthnHardware: true, WebAuthnUserVerified: true, WebAuthnUserPresence: true},
			want: testAMRWant{
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
			desc: "Duo",

			is: oidc.AuthenticationMethodsReferences{Duo: true},
			want: testAMRWant{
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
			desc: "Duo WebAuthn TOTP",

			is: oidc.AuthenticationMethodsReferences{Duo: true, WebAuthn: true, WebAuthnHardware: true, TOTP: true},
			want: testAMRWant{
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
			desc: "Duo TOTP",

			is: oidc.AuthenticationMethodsReferences{Duo: true, TOTP: true},
			want: testAMRWant{
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
			desc: "Username and Password with Duo",

			is: oidc.AuthenticationMethodsReferences{Duo: true, UsernameAndPassword: true},
			want: testAMRWant{
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
		t.Run(tc.desc, func(t *testing.T) {
			assert.Equal(t, tc.want.FactorKnowledge, tc.is.FactorKnowledge())
			assert.Equal(t, tc.want.FactorPossession, tc.is.FactorPossession())
			assert.Equal(t, tc.want.MultiFactorAuthentication, tc.is.MultiFactorAuthentication())
			assert.Equal(t, tc.want.ChannelBrowser, tc.is.ChannelBrowser())
			assert.Equal(t, tc.want.ChannelService, tc.is.ChannelService())
			assert.Equal(t, tc.want.MultiChannelAuthentication, tc.is.MultiChannelAuthentication())

			isRFC8176 := tc.is.MarshalRFC8176()

			for _, amr := range tc.want.RFC8176 {
				t.Run(fmt.Sprintf("has all wanted/%s", amr), func(t *testing.T) {
					assert.Contains(t, isRFC8176, amr)
				})
			}

			for _, amr := range isRFC8176 {
				t.Run(fmt.Sprintf("only has wanted/%s", amr), func(t *testing.T) {
					assert.Contains(t, tc.want.RFC8176, amr)
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
