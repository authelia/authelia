// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestAuthenticationMethodsReferencePolicy(t *testing.T) {
	testCases := []struct {
		name     string
		config   schema.AuthenticationBackendExternalIdentityProviderAMR
		asserted []string
		expected []string
	}{
		{"ShouldAdoptNothingByDefault", schema.AuthenticationBackendExternalIdentityProviderAMR{}, []string{"pwd", "otp"}, nil},
		{"ShouldAdoptNothingWithoutAnAssertionByDefault", schema.AuthenticationBackendExternalIdentityProviderAMR{}, nil, nil},
		{"ShouldAdoptTrustedAssertion", schema.AuthenticationBackendExternalIdentityProviderAMR{Trust: true}, []string{"pwd", "otp"}, []string{"pwd", "otp"}},
		{"ShouldAdoptTrustedAssertionOverTheDefault", schema.AuthenticationBackendExternalIdentityProviderAMR{Trust: true, Default: []string{"pwd"}}, []string{"hwk"}, []string{"hwk"}},
		{"ShouldAdoptTheDefaultWithoutAnAssertion", schema.AuthenticationBackendExternalIdentityProviderAMR{Trust: true, Default: []string{"pwd"}}, nil, []string{"pwd"}},
		{"ShouldAdoptTheDefaultWithAnEmptyAssertion", schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd"}}, []string{}, []string{"pwd"}},
		{"ShouldNotAdoptTheDefaultOverAnUntrustedAssertion", schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd"}}, []string{"otp"}, nil},
		{"ShouldAdoptTheOverride", schema.AuthenticationBackendExternalIdentityProviderAMR{Trust: true, Default: []string{"pwd", "otp"}, Override: true}, []string{"hwk"}, []string{"pwd", "otp"}},
		{"ShouldAdoptTheOverrideWithoutTrust", schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd"}, Override: true}, []string{"hwk"}, []string{"pwd"}},
		{"ShouldAdoptTheOverrideWithoutAnAssertion", schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd"}, Override: true}, nil, []string{"pwd"}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, newAuthenticationMethodsReferencePolicy(tc.config).resolve(tc.asserted))
		})
	}
}
