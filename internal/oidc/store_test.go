package oidc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestOpenIDConnectStore_GetClientPolicy(t *testing.T) {
	s := NewOpenIDConnectStore(&schema.OpenIDConnectConfiguration{
		IssuerPrivateKey: exampleIssuerPrivateKey,
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:          "myclient",
				Description: "myclient desc",
				Policy:      "one_factor",
				Scopes:      []string{"openid", "profile"},
				//Secret:      "mysecret",
			},
			{
				ID:          "myotherclient",
				Description: "myclient desc",
				Policy:      "two_factor",
				Scopes:      []string{"openid", "profile"},
				//Secret:      "mysecret",
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
	s := NewOpenIDConnectStore(&schema.OpenIDConnectConfiguration{
		IssuerPrivateKey: exampleIssuerPrivateKey,
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:          "myclient",
				Description: "myclient desc",
				Policy:      "one_factor",
				Scopes:      []string{"openid", "profile"},
				//Secret:      "mysecret",
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
	c1 := schema.OpenIDConnectClientConfiguration{
		ID:          "myclient",
		Description: "myclient desc",
		Policy:      "one_factor",
		Scopes:      []string{"openid", "profile"},
		//Secret:      "mysecret",
	}

	s := NewOpenIDConnectStore(&schema.OpenIDConnectConfiguration{
		IssuerPrivateKey: exampleIssuerPrivateKey,
		Clients:          []schema.OpenIDConnectClientConfiguration{c1},
	}, nil)

	client, err := s.GetFullClient(c1.ID)
	require.NoError(t, err)
	require.NotNil(t, client)
	assert.Equal(t, client.ID, c1.ID)
	assert.Equal(t, client.Description, c1.Description)
	assert.Equal(t, client.Scopes, c1.Scopes)
	assert.Equal(t, client.GrantTypes, c1.GrantTypes)
	assert.Equal(t, client.ResponseTypes, c1.ResponseTypes)
	assert.Equal(t, client.RedirectURIs, c1.RedirectURIs)
	assert.Equal(t, client.Policy, authorization.OneFactor)
	//assert.Equal(t, client.Secret, []byte(c1.Secret))
}

func TestOpenIDConnectStore_GetInternalClient_InvalidClient(t *testing.T) {
	c1 := schema.OpenIDConnectClientConfiguration{
		ID:          "myclient",
		Description: "myclient desc",
		Policy:      "one_factor",
		Scopes:      []string{"openid", "profile"},
		//Secret:      "mysecret",
	}

	s := NewOpenIDConnectStore(&schema.OpenIDConnectConfiguration{
		IssuerPrivateKey: exampleIssuerPrivateKey,
		Clients:          []schema.OpenIDConnectClientConfiguration{c1},
	}, nil)

	client, err := s.GetFullClient("another-client")
	assert.Nil(t, client)
	assert.EqualError(t, err, "not_found")
}

func TestOpenIDConnectStore_IsValidClientID(t *testing.T) {
	s := NewOpenIDConnectStore(&schema.OpenIDConnectConfiguration{
		IssuerPrivateKey: exampleIssuerPrivateKey,
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:          "myclient",
				Description: "myclient desc",
				Policy:      "one_factor",
				Scopes:      []string{"openid", "profile"},
				//Secret:      "mysecret",
			},
		},
	}, nil)

	validClient := s.IsValidClientID("myclient")
	invalidClient := s.IsValidClientID("myinvalidclient")

	assert.True(t, validClient)
	assert.False(t, invalidClient)
}
