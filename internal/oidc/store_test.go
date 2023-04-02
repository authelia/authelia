package oidc

import (
	"context"
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestOpenIDConnectStore_GetClientPolicy(t *testing.T) {
	s := NewStore(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       mustParseRSAPrivateKey(exampleIssuerPrivateKey),
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:          "myclient",
				Description: "myclient desc",
				Policy:      "one_factor",
				Scopes:      []string{ScopeOpenID, ScopeProfile},
				Secret:      MustDecodeSecret("$plaintext$mysecret"),
			},
			{
				ID:          "myotherclient",
				Description: "myclient desc",
				Policy:      "two_factor",
				Scopes:      []string{ScopeOpenID, ScopeProfile},
				Secret:      MustDecodeSecret("$plaintext$mysecret"),
			},
		},
	}, nil)

	policyOne := s.GetClientPolicy("myclient")
	assert.Equal(t, authorization.OneFactor, policyOne)

	policyTwo := s.GetClientPolicy("myotherclient")
	assert.Equal(t, authorization.TwoFactor, policyTwo)

	policyInvalid := s.GetClientPolicy("invalidclient")
	assert.Equal(t, authorization.TwoFactor, policyInvalid)
}

func TestOpenIDConnectStore_GetInternalClient(t *testing.T) {
	s := NewStore(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       mustParseRSAPrivateKey(exampleIssuerPrivateKey),
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:          "myclient",
				Description: "myclient desc",
				Policy:      "one_factor",
				Scopes:      []string{ScopeOpenID, ScopeProfile},
				Secret:      MustDecodeSecret("$plaintext$mysecret"),
			},
		},
	}, nil)

	client, err := s.GetClient(context.Background(), "myinvalidclient")
	assert.EqualError(t, err, "not_found")
	assert.Nil(t, client)

	client, err = s.GetClient(context.Background(), "myclient")
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, "myclient", client.GetID())
}

func TestOpenIDConnectStore_GetInternalClient_ValidClient(t *testing.T) {
	id := "myclient"

	c1 := schema.OpenIDConnectClientConfiguration{
		ID:          id,
		Description: "myclient desc",
		Policy:      "one_factor",
		Scopes:      []string{ScopeOpenID, ScopeProfile},
		Secret:      MustDecodeSecret("$plaintext$mysecret"),
	}

	s := NewStore(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       mustParseRSAPrivateKey(exampleIssuerPrivateKey),
		Clients:                []schema.OpenIDConnectClientConfiguration{c1},
	}, nil)

	client, err := s.GetFullClient(id)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, id, client.GetID())
	assert.Equal(t, "myclient desc", client.GetDescription())
	assert.Equal(t, fosite.Arguments(c1.Scopes), client.GetScopes())
	assert.Equal(t, fosite.Arguments([]string{GrantTypeAuthorizationCode}), client.GetGrantTypes())
	assert.Equal(t, fosite.Arguments([]string{ResponseTypeAuthorizationCodeFlow}), client.GetResponseTypes())
	assert.Equal(t, []string(nil), client.GetRedirectURIs())
	assert.Equal(t, authorization.OneFactor, client.GetAuthorizationPolicy())
	assert.Equal(t, "$plaintext$mysecret", client.GetSecret().Encode())
}

func TestOpenIDConnectStore_GetInternalClient_InvalidClient(t *testing.T) {
	c1 := schema.OpenIDConnectClientConfiguration{
		ID:          "myclient",
		Description: "myclient desc",
		Policy:      "one_factor",
		Scopes:      []string{ScopeOpenID, ScopeProfile},
		Secret:      MustDecodeSecret("$plaintext$mysecret"),
	}

	s := NewStore(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       mustParseRSAPrivateKey(exampleIssuerPrivateKey),
		Clients:                []schema.OpenIDConnectClientConfiguration{c1},
	}, nil)

	client, err := s.GetFullClient("another-client")
	assert.Nil(t, client)
	assert.EqualError(t, err, "not_found")
}

func TestOpenIDConnectStore_IsValidClientID(t *testing.T) {
	s := NewStore(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       mustParseRSAPrivateKey(exampleIssuerPrivateKey),
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:          "myclient",
				Description: "myclient desc",
				Policy:      "one_factor",
				Scopes:      []string{ScopeOpenID, ScopeProfile},
				Secret:      MustDecodeSecret("$plaintext$mysecret"),
			},
		},
	}, nil)

	validClient := s.IsValidClientID("myclient")
	invalidClient := s.IsValidClientID("myinvalidclient")

	assert.True(t, validClient)
	assert.False(t, invalidClient)
}
