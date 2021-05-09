package oidc

import (
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/internal/authentication"
	"github.com/authelia/authelia/internal/authorization"
	"github.com/authelia/authelia/internal/session"
)

func TestIsAuthenticationLevelSufficient(t *testing.T) {
	c := InternalClient{}

	c.Policy = authorization.Bypass
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.NotAuthenticated))
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.OneFactor))
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.TwoFactor))

	c.Policy = authorization.OneFactor
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.NotAuthenticated))
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.OneFactor))
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.TwoFactor))

	c.Policy = authorization.TwoFactor
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.NotAuthenticated))
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.OneFactor))
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.TwoFactor))

	c.Policy = authorization.Denied
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.NotAuthenticated))
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.OneFactor))
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.TwoFactor))
}

func TestInternalClient_GetConsentRequestBody(t *testing.T) {
	c := InternalClient{}

	consentRequestBody := c.GetConsentRequestBody(nil)
	assert.Equal(t, "", consentRequestBody.ClientID)
	assert.Equal(t, "", consentRequestBody.ClientDescription)
	assert.Equal(t, []Scope(nil), consentRequestBody.Scopes)
	assert.Equal(t, []Audience(nil), consentRequestBody.Audience)

	c.ID = "myclient"
	c.Description = "My Client"

	workflow := &session.OIDCWorkflowSession{
		RequestedAudience: []string{"https://example.com"},
		RequestedScopes:   []string{"openid", "groups"},
	}
	expectedScopes := []Scope{
		{"openid", "Use OpenID to verify your identity"},
		{"groups", "Access your group membership"},
	}
	expectedAudiences := []Audience{
		{"https://example.com", "https://example.com"},
	}

	consentRequestBody = c.GetConsentRequestBody(workflow)
	assert.Equal(t, "myclient", consentRequestBody.ClientID)
	assert.Equal(t, "My Client", consentRequestBody.ClientDescription)
	assert.Equal(t, expectedScopes, consentRequestBody.Scopes)
	assert.Equal(t, expectedAudiences, consentRequestBody.Audience)
}

func TestInternalClient_GetAudience(t *testing.T) {
	c := InternalClient{}

	audience := c.GetAudience()
	assert.Len(t, audience, 0)

	c.Audience = []string{"https://example.com"}

	audience = c.GetAudience()
	require.Len(t, audience, 1)
	assert.Equal(t, "https://example.com", audience[0])
}

func TestInternalClient_GetScopes(t *testing.T) {
	c := InternalClient{}

	scopes := c.GetScopes()
	assert.Len(t, scopes, 0)

	c.Scopes = []string{"openid"}

	scopes = c.GetScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, "openid", scopes[0])
}

func TestInternalClient_GetGrantTypes(t *testing.T) {
	c := InternalClient{}

	grantTypes := c.GetGrantTypes()
	require.Len(t, grantTypes, 1)
	assert.Equal(t, "authorization_code", grantTypes[0])

	c.GrantTypes = []string{"device_code"}

	grantTypes = c.GetGrantTypes()
	require.Len(t, grantTypes, 1)
	assert.Equal(t, "device_code", grantTypes[0])
}

func TestInternalClient_GetHashedSecret(t *testing.T) {
	c := InternalClient{}

	hashedSecret := c.GetHashedSecret()
	assert.Equal(t, []byte(nil), hashedSecret)

	c.Secret = []byte("a_bad_secret")

	hashedSecret = c.GetHashedSecret()
	assert.Equal(t, []byte("a_bad_secret"), hashedSecret)
}

func TestInternalClient_GetID(t *testing.T) {
	c := InternalClient{}

	id := c.GetID()
	assert.Equal(t, "", id)

	c.ID = "myid"

	id = c.GetID()
	assert.Equal(t, "myid", id)
}

func TestInternalClient_GetRedirectURIs(t *testing.T) {
	c := InternalClient{}

	redirectURIs := c.GetRedirectURIs()
	require.Len(t, redirectURIs, 0)

	c.RedirectURIs = []string{"https://example.com/oauth2/callback"}

	redirectURIs = c.GetRedirectURIs()
	require.Len(t, redirectURIs, 1)
	assert.Equal(t, "https://example.com/oauth2/callback", redirectURIs[0])
}

func TestInternalClient_GetRequestURIs(t *testing.T) {
	c := InternalClient{}

	requestURIs := c.GetRequestURIs()
	require.Len(t, requestURIs, 0)

	c.RequestURIs = []string{"https://example.com/oauth2/callback"}

	requestURIs = c.GetRequestURIs()
	require.Len(t, requestURIs, 1)
	assert.Equal(t, "https://example.com/oauth2/callback", requestURIs[0])
}

func TestInternalClient_GetResponseModes(t *testing.T) {
	c := InternalClient{}

	responseModes := c.GetResponseModes()
	require.Len(t, responseModes, 0)

	c.ResponseModes = []fosite.ResponseModeType{
		fosite.ResponseModeDefault, fosite.ResponseModeFormPost,
		fosite.ResponseModeQuery, fosite.ResponseModeFragment,
	}

	responseModes = c.GetResponseModes()
	require.Len(t, responseModes, 4)
	assert.Equal(t, fosite.ResponseModeDefault, responseModes[0])
	assert.Equal(t, fosite.ResponseModeFormPost, responseModes[1])
	assert.Equal(t, fosite.ResponseModeQuery, responseModes[2])
	assert.Equal(t, fosite.ResponseModeFragment, responseModes[3])
}

func TestInternalClient_GetResponseTypes(t *testing.T) {
	c := InternalClient{}

	responseTypes := c.GetResponseTypes()
	require.Len(t, responseTypes, 1)
	assert.Equal(t, "code", responseTypes[0])

	c.ResponseTypes = []string{"code", "id_token"}

	responseTypes = c.GetResponseTypes()
	require.Len(t, responseTypes, 2)
	assert.Equal(t, "code", responseTypes[0])
	assert.Equal(t, "id_token", responseTypes[1])
}

func TestInternalClient_IsPublic(t *testing.T) {
	c := InternalClient{}

	assert.False(t, c.IsPublic())

	c.Public = true
	assert.True(t, c.IsPublic())
}

func TestInternalClient_GetTokenEndpointAuthSigningAlgorithm(t *testing.T) {
	c := InternalClient{}
	assert.Equal(t, "RS256", c.GetTokenEndpointAuthSigningAlgorithm())

	c.TokenEndpointAuthSigningAlgorithm = "HS256"
	assert.Equal(t, "HS256", c.GetTokenEndpointAuthSigningAlgorithm())
}

func TestInternalClient_GetTokenEndpointAuthMethod(t *testing.T) {
	c := InternalClient{}
	assert.Equal(t, "client_secret_basic", c.GetTokenEndpointAuthMethod())

	c.TokenEndpointAuthMethod = "token-endpoint-auth-method"
	assert.Equal(t, "token-endpoint-auth-method", c.GetTokenEndpointAuthMethod())
}
