// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package events

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryShouldContainEveryEventType(t *testing.T) {
	expected := []string{
		TypeSecurityAuthenticationFailed,
		TypeSecurityAuthenticationSucceeded,
		TypeSecurityBanApplied,
		TypeSecurityBanExpired,
		TypeSystemStartupCheck,
		TypeUserCredentialTOTPAdded,
		TypeUserCredentialTOTPRemoved,
		TypeUserCredentialWebAuthnAdded,
		TypeUserCredentialWebAuthnRemoved,
		TypeUserIdentityVerificationStarted,
		TypeUserPasswordChanged,
		TypeUserPasswordReset,
		TypeUserSessionElevationRequested,
	}

	assert.Equal(t, expected, Registered())
}

func TestMatchShouldResolveExactNames(t *testing.T) {
	assert.Equal(t, []string{TypeUserPasswordChanged}, Match(TypeUserPasswordChanged))
}

func TestMatchShouldResolveGlobs(t *testing.T) {
	assert.Equal(t, []string{
		TypeUserCredentialTOTPAdded,
		TypeUserCredentialTOTPRemoved,
		TypeUserCredentialWebAuthnAdded,
		TypeUserCredentialWebAuthnRemoved,
	}, Match("user.credential.*"))
}

func TestMatchShouldResolveGlobAll(t *testing.T) {
	assert.Equal(t, Registered(), Match("*"))
}

func TestMatchShouldReturnNothingForUnknownSelectors(t *testing.T) {
	assert.Empty(t, Match("users.*"))
	assert.Empty(t, Match("user.password.rotated"))
}

func TestDescriptorShouldProvideDataSchema(t *testing.T) {
	descriptor, ok := Lookup(TypeUserPasswordChanged)

	require.True(t, ok)
	assert.Equal(t, "https://www.authelia.com/schemas/webhooks/v1/user.password.changed.json", descriptor.DataSchema)
}

func TestEveryRegisteredTypeShouldProduceMatchingData(t *testing.T) {
	for _, name := range Registered() {
		descriptor, ok := Lookup(name)

		require.True(t, ok, name)
		require.NotNil(t, descriptor.New, name)
		assert.Equal(t, name, descriptor.New().EventType(), name)
	}
}
