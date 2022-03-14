package oidc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticationMethodsReferences(t *testing.T) {
	var amr AuthenticationMethodsReferences

	amr = AuthenticationMethodsReferences{
		UsernameAndPassword: true,
	}

	assert.True(t, amr.FactorKnowledge())
	assert.False(t, amr.FactorPossession())
	assert.False(t, amr.MultiFactorAuthentication())
	assert.True(t, amr.ChannelBrowser())
	assert.False(t, amr.ChannelService())
	assert.False(t, amr.MultiChannelAuthentication())
	assert.Equal(t, []string{"pwd"}, amr.MarshalRFC8176())

	amr = AuthenticationMethodsReferences{
		TOTP: true,
	}

	assert.False(t, amr.FactorKnowledge())
	assert.True(t, amr.FactorPossession())
	assert.False(t, amr.MultiFactorAuthentication())
	assert.True(t, amr.ChannelBrowser())
	assert.False(t, amr.ChannelService())
	assert.False(t, amr.MultiChannelAuthentication())
	assert.Equal(t, []string{"otp"}, amr.MarshalRFC8176())

	amr = AuthenticationMethodsReferences{
		Webauthn: true,
	}

	assert.False(t, amr.FactorKnowledge())
	assert.True(t, amr.FactorPossession())
	assert.False(t, amr.MultiFactorAuthentication())
	assert.True(t, amr.ChannelBrowser())
	assert.False(t, amr.ChannelService())
	assert.False(t, amr.MultiChannelAuthentication())
	assert.Equal(t, []string{"hwk"}, amr.MarshalRFC8176())

	amr = AuthenticationMethodsReferences{
		WebauthnUserPresence: true,
	}

	assert.False(t, amr.FactorKnowledge())
	assert.False(t, amr.FactorPossession())
	assert.False(t, amr.MultiFactorAuthentication())
	assert.False(t, amr.ChannelBrowser())
	assert.False(t, amr.ChannelService())
	assert.False(t, amr.MultiChannelAuthentication())
	assert.Equal(t, []string{"user"}, amr.MarshalRFC8176())

	amr = AuthenticationMethodsReferences{
		WebauthnUserVerified: true,
	}

	assert.False(t, amr.FactorKnowledge())
	assert.False(t, amr.FactorPossession())
	assert.False(t, amr.MultiFactorAuthentication())
	assert.False(t, amr.ChannelBrowser())
	assert.False(t, amr.ChannelService())
	assert.False(t, amr.MultiChannelAuthentication())
	assert.Equal(t, []string{"pin"}, amr.MarshalRFC8176())

	amr = AuthenticationMethodsReferences{
		Webauthn:             true,
		WebauthnUserPresence: true,
		WebauthnUserVerified: true,
	}

	assert.False(t, amr.FactorKnowledge())
	assert.True(t, amr.FactorPossession())
	assert.False(t, amr.MultiFactorAuthentication())
	assert.True(t, amr.ChannelBrowser())
	assert.False(t, amr.ChannelService())
	assert.False(t, amr.MultiChannelAuthentication())
	assert.Equal(t, []string{"hwk", "user", "pin"}, amr.MarshalRFC8176())

	amr = AuthenticationMethodsReferences{
		Duo: true,
	}

	assert.False(t, amr.FactorKnowledge())
	assert.True(t, amr.FactorPossession())
	assert.False(t, amr.MultiFactorAuthentication())
	assert.False(t, amr.ChannelBrowser())
	assert.True(t, amr.ChannelService())
	assert.False(t, amr.MultiChannelAuthentication())
	assert.Equal(t, []string{"sms"}, amr.MarshalRFC8176())

	amr = AuthenticationMethodsReferences{
		Duo:      true,
		Webauthn: true,
		TOTP:     true,
	}

	assert.False(t, amr.FactorKnowledge())
	assert.True(t, amr.FactorPossession())
	assert.False(t, amr.MultiFactorAuthentication())
	assert.True(t, amr.ChannelBrowser())
	assert.True(t, amr.ChannelService())
	assert.True(t, amr.MultiChannelAuthentication())
	assert.Equal(t, []string{"otp", "sms", "hwk", "mca"}, amr.MarshalRFC8176())

	amr = AuthenticationMethodsReferences{
		Duo:  true,
		TOTP: true,
	}

	assert.False(t, amr.FactorKnowledge())
	assert.True(t, amr.FactorPossession())
	assert.False(t, amr.MultiFactorAuthentication())
	assert.True(t, amr.ChannelBrowser())
	assert.True(t, amr.ChannelService())
	assert.True(t, amr.MultiChannelAuthentication())
	assert.Equal(t, []string{"otp", "sms", "mca"}, amr.MarshalRFC8176())

	amr = AuthenticationMethodsReferences{
		Duo:                 true,
		UsernameAndPassword: true,
	}

	assert.True(t, amr.FactorKnowledge())
	assert.True(t, amr.FactorPossession())
	assert.True(t, amr.MultiFactorAuthentication())
	assert.True(t, amr.ChannelBrowser())
	assert.True(t, amr.ChannelService())
	assert.True(t, amr.MultiChannelAuthentication())
	assert.Equal(t, []string{"pwd", "sms", "mfa", "mca"}, amr.MarshalRFC8176())
}
