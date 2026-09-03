// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewProvidersShouldBuildEachType(t *testing.T) {
	providers := NewProviders(&schema.AuthenticationBackendExternalIdentity{
		Providers: []schema.AuthenticationBackendExternalIdentityProvider{
			{ID: "example", Type: ProviderTypeOpenIDConnect, Name: "Example", Issuer: "https://op.example.com", ClientID: "client", ResponseMode: ResponseModeFormPost, AuthenticationMethodsReference: schema.AuthenticationBackendExternalIdentityProviderAMR{Trust: true}},
			{ID: "discord", Type: ProviderTypeDiscord, Name: "Discord", ClientID: "123", ClientSecret: "secret"},
			{ID: "untyped", Name: "Untyped", Issuer: "https://untyped.example.com", ClientID: "client"},
		},
	}, nil)

	require.NotNil(t, providers)
	require.Len(t, providers.All(), 3)

	assert.Equal(t, []string{"example", "discord", "untyped"}, []string{providers.All()[0].ID(), providers.All()[1].ID(), providers.All()[2].ID()})

	example, ok := providers.Get("example")
	require.True(t, ok)
	require.IsType(t, &OpenIDConnectProvider{}, example)

	assert.Equal(t, "Example", example.Name())
	assert.Equal(t, ProviderTypeOpenIDConnect, example.Type())
	assert.Equal(t, "https://op.example.com", example.Issuer())
	assert.Equal(t, ResponseModeFormPost, example.ResponseMode())
	assert.Equal(t, []string{"pwd", "otp"}, example.AuthenticationMethodsReference([]string{"pwd", "otp"}))

	discord, ok := providers.Get("discord")
	require.True(t, ok)
	require.IsType(t, &DiscordProvider{}, discord)

	assert.Equal(t, "Discord", discord.Name())
	assert.Equal(t, ProviderTypeDiscord, discord.Type())
	assert.Equal(t, "https://discord.com", discord.Issuer())
	assert.Equal(t, ResponseModeQuery, discord.ResponseMode())
	assert.Equal(t, []string{"pwd", "kba"}, discord.AuthenticationMethodsReference(nil), "without default values a provider which asserts nothing adopts the values of a password sign in")
	assert.Equal(t, []string{"pwd", "kba"}, discord.AuthenticationMethodsReference([]string{"otp"}), "a provider which asserts nothing never adopts asserted values")

	untyped, ok := providers.Get("untyped")
	require.True(t, ok)
	assert.IsType(t, &OpenIDConnectProvider{}, untyped, "a provider without a type is an OpenID Connect 1.0 Provider")
	assert.Equal(t, ResponseModeQuery, untyped.ResponseMode())

	_, ok = providers.Get("missing")
	assert.False(t, ok)
}

func TestProvidersShouldHandleNil(t *testing.T) {
	var providers *Providers

	assert.Nil(t, NewProviders(nil, nil))
	assert.Nil(t, NewProviders(&schema.AuthenticationBackendExternalIdentity{}, nil))
	assert.Nil(t, providers.All())

	provider, ok := providers.Get("example")

	assert.Nil(t, provider)
	assert.False(t, ok)
}

func getOpenIDConnectProvider(providers *Providers) (*OpenIDConnectProvider, bool) {
	provider, ok := providers.Get("example")
	if !ok {
		return nil, false
	}

	oidc, ok := provider.(*OpenIDConnectProvider)

	return oidc, ok
}
