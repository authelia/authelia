package oidc

import (
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestNewClient(t *testing.T) {
	blankConfig := schema.OpenIDConnectClientConfiguration{}
	blankClient := NewClient(blankConfig)
	assert.Equal(t, "", blankClient.ID)
	assert.Equal(t, "", blankClient.Description)
	assert.Equal(t, "", blankClient.Description)
	require.Len(t, blankClient.ResponseModes, 1)
	assert.Equal(t, fosite.ResponseModeDefault, blankClient.ResponseModes[0])

	exampleConfig := schema.OpenIDConnectClientConfiguration{
		ID:            "myapp",
		Description:   "My App",
		Policy:        "two_factor",
		Secret:        MustDecodeSecret("$plaintext$abcdef"),
		RedirectURIs:  []string{"https://google.com/callback"},
		Scopes:        schema.DefaultOpenIDConnectClientConfiguration.Scopes,
		ResponseTypes: schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes,
		GrantTypes:    schema.DefaultOpenIDConnectClientConfiguration.GrantTypes,
		ResponseModes: schema.DefaultOpenIDConnectClientConfiguration.ResponseModes,
	}

	exampleClient := NewClient(exampleConfig)
	assert.Equal(t, "myapp", exampleClient.ID)
	require.Len(t, exampleClient.ResponseModes, 4)
	assert.Equal(t, fosite.ResponseModeDefault, exampleClient.ResponseModes[0])
	assert.Equal(t, fosite.ResponseModeFormPost, exampleClient.ResponseModes[1])
	assert.Equal(t, fosite.ResponseModeQuery, exampleClient.ResponseModes[2])
	assert.Equal(t, fosite.ResponseModeFragment, exampleClient.ResponseModes[3])
	assert.Equal(t, authorization.TwoFactor, exampleClient.Policy)
}

func TestIsAuthenticationLevelSufficient(t *testing.T) {
	c := Client{}

	c.Policy = authorization.Bypass
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.NotAuthenticated))
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

func TestClient_GetConsentResponseBody(t *testing.T) {
	c := Client{}

	consentRequestBody := c.GetConsentResponseBody(nil)
	assert.Equal(t, "", consentRequestBody.ClientID)
	assert.Equal(t, "", consentRequestBody.ClientDescription)
	assert.Equal(t, []string(nil), consentRequestBody.Scopes)
	assert.Equal(t, []string(nil), consentRequestBody.Audience)

	c.ID = "myclient"
	c.Description = "My Client"

	consent := &model.OAuth2ConsentSession{
		RequestedAudience: []string{"https://example.com"},
		RequestedScopes:   []string{"openid", "groups"},
	}

	expectedScopes := []string{"openid", "groups"}
	expectedAudiences := []string{"https://example.com"}

	consentRequestBody = c.GetConsentResponseBody(consent)
	assert.Equal(t, "myclient", consentRequestBody.ClientID)
	assert.Equal(t, "My Client", consentRequestBody.ClientDescription)
	assert.Equal(t, expectedScopes, consentRequestBody.Scopes)
	assert.Equal(t, expectedAudiences, consentRequestBody.Audience)
}

func TestClient_GetAudience(t *testing.T) {
	c := Client{}

	audience := c.GetAudience()
	assert.Len(t, audience, 0)

	c.Audience = []string{"https://example.com"}

	audience = c.GetAudience()
	require.Len(t, audience, 1)
	assert.Equal(t, "https://example.com", audience[0])
}

func TestClient_GetScopes(t *testing.T) {
	c := Client{}

	scopes := c.GetScopes()
	assert.Len(t, scopes, 0)

	c.Scopes = []string{"openid"}

	scopes = c.GetScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, "openid", scopes[0])
}

func TestClient_GetGrantTypes(t *testing.T) {
	c := Client{}

	grantTypes := c.GetGrantTypes()
	require.Len(t, grantTypes, 1)
	assert.Equal(t, "authorization_code", grantTypes[0])

	c.GrantTypes = []string{"device_code"}

	grantTypes = c.GetGrantTypes()
	require.Len(t, grantTypes, 1)
	assert.Equal(t, "device_code", grantTypes[0])
}

func TestClient_Hashing(t *testing.T) {
	c := Client{}

	hashedSecret := c.GetHashedSecret()
	assert.Equal(t, []byte(nil), hashedSecret)

	c.Secret = MustDecodeSecret("$plaintext$a_bad_secret")

	assert.True(t, c.Secret.MatchBytes([]byte("a_bad_secret")))
}

func TestClient_GetHashedSecret(t *testing.T) {
	c := Client{}

	hashedSecret := c.GetHashedSecret()
	assert.Equal(t, []byte(nil), hashedSecret)

	c.Secret = MustDecodeSecret("$plaintext$a_bad_secret")

	hashedSecret = c.GetHashedSecret()
	assert.Equal(t, []byte("$plaintext$a_bad_secret"), hashedSecret)
}

func TestClient_GetID(t *testing.T) {
	c := Client{}

	id := c.GetID()
	assert.Equal(t, "", id)

	c.ID = "myid"

	id = c.GetID()
	assert.Equal(t, "myid", id)
}

func TestClient_GetRedirectURIs(t *testing.T) {
	c := Client{}

	redirectURIs := c.GetRedirectURIs()
	require.Len(t, redirectURIs, 0)

	c.RedirectURIs = []string{"https://example.com/oauth2/callback"}

	redirectURIs = c.GetRedirectURIs()
	require.Len(t, redirectURIs, 1)
	assert.Equal(t, "https://example.com/oauth2/callback", redirectURIs[0])
}

func TestClient_GetResponseModes(t *testing.T) {
	c := Client{}

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

func TestClient_GetResponseTypes(t *testing.T) {
	c := Client{}

	responseTypes := c.GetResponseTypes()
	require.Len(t, responseTypes, 1)
	assert.Equal(t, "code", responseTypes[0])

	c.ResponseTypes = []string{"code", "id_token"}

	responseTypes = c.GetResponseTypes()
	require.Len(t, responseTypes, 2)
	assert.Equal(t, "code", responseTypes[0])
	assert.Equal(t, "id_token", responseTypes[1])
}

func TestClient_IsPublic(t *testing.T) {
	c := Client{}

	assert.False(t, c.IsPublic())

	c.Public = true
	assert.True(t, c.IsPublic())
}

func MustDecodeSecret(value string) *schema.PasswordDigest {
	if secret, err := schema.DecodePasswordDigest(value); err != nil {
		panic(err)
	} else {
		return secret
	}
}
