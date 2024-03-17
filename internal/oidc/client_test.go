package oidc_test

import (
	"context"
	"testing"
	"time"

	oauthelia2 "authelia.com/provider/oauth2"
	"github.com/go-jose/go-jose/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewClient(t *testing.T) {
	config := schema.IdentityProvidersOpenIDConnectClient{}
	client := oidc.NewClient(config, &schema.IdentityProvidersOpenIDConnect{})
	assert.Equal(t, "", client.GetID())
	assert.Equal(t, "", client.GetName())
	assert.Len(t, client.GetResponseModes(), 0)
	assert.Len(t, client.GetResponseTypes(), 1)
	assert.Equal(t, "", client.GetSectorIdentifierURI())

	bclient, ok := client.(*oidc.RegisteredClient)
	require.True(t, ok)
	assert.Equal(t, "", bclient.UserinfoSignedResponseAlg)
	assert.Equal(t, oidc.SigningAlgNone, client.GetUserinfoSignedResponseAlg())
	assert.Equal(t, "", client.GetUserinfoSignedResponseKeyID())
	assert.Equal(t, oidc.SigningAlgNone, client.GetIntrospectionSignedResponseAlg())
	assert.Equal(t, "", client.GetIntrospectionSignedResponseKeyID())

	_, ok = client.(*oidc.RegisteredClient)
	assert.True(t, ok)

	config = schema.IdentityProvidersOpenIDConnectClient{
		ID:                  myclient,
		Name:                myclientdesc,
		AuthorizationPolicy: twofactor,
		Secret:              tOpenIDConnectPlainTextClientSecret,
		RedirectURIs:        []string{examplecom},
		Scopes:              schema.DefaultOpenIDConnectClientConfiguration.Scopes,
		ResponseTypes:       schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes,
		GrantTypes:          schema.DefaultOpenIDConnectClientConfiguration.GrantTypes,
		ResponseModes:       schema.DefaultOpenIDConnectClientConfiguration.ResponseModes,
	}

	client = oidc.NewClient(config, &schema.IdentityProvidersOpenIDConnect{})
	assert.Equal(t, myclient, client.GetID())
	require.Len(t, client.GetResponseModes(), 1)
	assert.Equal(t, oauthelia2.ResponseModeFormPost, client.GetResponseModes()[0])
	assert.Equal(t, authorization.TwoFactor, client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{}))
	assert.Equal(t, oauthelia2.Arguments(nil), client.GetAudience())

	config = schema.IdentityProvidersOpenIDConnectClient{
		TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
	}

	client = oidc.NewClient(config, &schema.IdentityProvidersOpenIDConnect{})

	fclient, ok := client.(*oidc.RegisteredClient)

	require.True(t, ok)

	assert.Equal(t, "", fclient.UserinfoSignedResponseAlg)
	assert.Equal(t, oidc.SigningAlgNone, client.GetUserinfoSignedResponseAlg())
	assert.Equal(t, oidc.SigningAlgNone, fclient.GetUserinfoSignedResponseAlg())
	assert.Equal(t, oidc.SigningAlgNone, fclient.UserinfoSignedResponseAlg)

	assert.Equal(t, "", fclient.UserinfoSignedResponseKeyID)
	assert.Equal(t, "", client.GetUserinfoSignedResponseKeyID())
	assert.Equal(t, "", fclient.GetUserinfoSignedResponseKeyID())

	fclient.UserinfoSignedResponseKeyID = "aukeyid"

	assert.Equal(t, "aukeyid", client.GetUserinfoSignedResponseKeyID())
	assert.Equal(t, "aukeyid", fclient.GetUserinfoSignedResponseKeyID())

	assert.Equal(t, "", fclient.IntrospectionSignedResponseAlg)
	assert.Equal(t, oidc.SigningAlgNone, client.GetIntrospectionSignedResponseAlg())
	assert.Equal(t, oidc.SigningAlgNone, fclient.GetIntrospectionSignedResponseAlg())
	assert.Equal(t, oidc.SigningAlgNone, fclient.IntrospectionSignedResponseAlg)

	assert.Equal(t, "", fclient.IntrospectionSignedResponseKeyID)
	assert.Equal(t, "", client.GetIntrospectionSignedResponseKeyID())
	assert.Equal(t, "", fclient.GetIntrospectionSignedResponseKeyID())

	fclient.IntrospectionSignedResponseKeyID = "aikeyid"

	assert.Equal(t, "aikeyid", client.GetIntrospectionSignedResponseKeyID())
	assert.Equal(t, "aikeyid", fclient.GetIntrospectionSignedResponseKeyID())

	fclient.IntrospectionSignedResponseAlg = oidc.SigningAlgRSAUsingSHA512

	assert.Equal(t, oidc.SigningAlgRSAUsingSHA512, client.GetIntrospectionSignedResponseAlg())
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA512, fclient.GetIntrospectionSignedResponseAlg())

	assert.Equal(t, "", fclient.IDTokenSignedResponseAlg)
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, client.GetIDTokenSignedResponseAlg())
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, fclient.GetIDTokenSignedResponseAlg())
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, fclient.IDTokenSignedResponseAlg)

	assert.Equal(t, "", fclient.IDTokenSignedResponseKeyID)
	assert.Equal(t, "", client.GetIDTokenSignedResponseKeyID())
	assert.Equal(t, "", fclient.GetIDTokenSignedResponseKeyID())

	fclient.IDTokenSignedResponseKeyID = "akeyid"

	assert.Equal(t, "akeyid", client.GetIDTokenSignedResponseKeyID())
	assert.Equal(t, "akeyid", fclient.GetIDTokenSignedResponseKeyID())

	assert.Equal(t, "", fclient.AccessTokenSignedResponseAlg)
	assert.Equal(t, oidc.SigningAlgNone, client.GetAccessTokenSignedResponseAlg())
	assert.Equal(t, oidc.SigningAlgNone, fclient.GetAccessTokenSignedResponseAlg())
	assert.Equal(t, oidc.SigningAlgNone, fclient.AccessTokenSignedResponseAlg)

	assert.Equal(t, "", fclient.AccessTokenSignedResponseKeyID)
	assert.Equal(t, "", client.GetAccessTokenSignedResponseKeyID())
	assert.Equal(t, "", fclient.GetAccessTokenSignedResponseKeyID())

	fclient.AccessTokenSignedResponseKeyID = "atkeyid"

	assert.Equal(t, "atkeyid", client.GetAccessTokenSignedResponseKeyID())
	assert.Equal(t, "atkeyid", fclient.GetAccessTokenSignedResponseKeyID())

	assert.Equal(t, oidc.ClientAuthMethodClientSecretPost, fclient.TokenEndpointAuthMethod)
	assert.Equal(t, oidc.ClientAuthMethodClientSecretPost, fclient.GetTokenEndpointAuthMethod())

	assert.Equal(t, "", fclient.TokenEndpointAuthSigningAlg)
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, fclient.GetTokenEndpointAuthSigningAlg())
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, fclient.TokenEndpointAuthSigningAlg)

	assert.Equal(t, "", fclient.RequestObjectSigningAlg)
	assert.Equal(t, "", fclient.GetRequestObjectSigningAlg())

	fclient.RequestObjectSigningAlg = oidc.SigningAlgRSAUsingSHA256

	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, fclient.GetRequestObjectSigningAlg())

	assert.Nil(t, fclient.JSONWebKeysURI)
	assert.Equal(t, "", fclient.GetJSONWebKeysURI())

	fclient.JSONWebKeysURI = MustParseRequestURI("https://example.com")
	assert.Equal(t, "https://example.com", fclient.GetJSONWebKeysURI())

	var niljwks *jose.JSONWebKeySet

	assert.Equal(t, niljwks, fclient.JSONWebKeys)
	assert.Equal(t, niljwks, fclient.GetJSONWebKeys())

	assert.Equal(t, oidc.ClientConsentMode(0), fclient.ConsentPolicy.Mode)
	assert.Equal(t, time.Second*0, fclient.ConsentPolicy.Duration)
	assert.Equal(t, oidc.ClientConsentPolicy{Mode: oidc.ClientConsentModeExplicit}, fclient.GetConsentPolicy())

	fclient.TokenEndpointAuthMethod = ""
	fclient.Public = false
	assert.Equal(t, oidc.ClientAuthMethodClientSecretBasic, fclient.GetTokenEndpointAuthMethod())
	assert.Equal(t, oidc.ClientAuthMethodClientSecretBasic, fclient.TokenEndpointAuthMethod)

	fclient.TokenEndpointAuthMethod = ""
	fclient.Public = true
	assert.Equal(t, oidc.ClientAuthMethodNone, fclient.GetTokenEndpointAuthMethod())
	assert.Equal(t, oidc.ClientAuthMethodNone, fclient.TokenEndpointAuthMethod)

	assert.Equal(t, []string(nil), fclient.RequestURIs)
	assert.Equal(t, []string(nil), fclient.GetRequestURIs())
}

func TestBaseClient_Misc(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(client *oidc.RegisteredClient)
		expected func(t *testing.T, client *oidc.RegisteredClient)
	}{
		{
			"ShouldReturnGetRefreshFlowIgnoreOriginalGrantedScopes",
			func(client *oidc.RegisteredClient) {
				client.RefreshFlowIgnoreOriginalGrantedScopes = true
			},
			func(t *testing.T, client *oidc.RegisteredClient) {
				assert.True(t, client.GetRefreshFlowIgnoreOriginalGrantedScopes(context.TODO()))
			},
		},
		{
			"ShouldReturnGetRefreshFlowIgnoreOriginalGrantedScopesFalse",
			func(client *oidc.RegisteredClient) {
				client.RefreshFlowIgnoreOriginalGrantedScopes = false
			},
			func(t *testing.T, client *oidc.RegisteredClient) {
				assert.False(t, client.GetRefreshFlowIgnoreOriginalGrantedScopes(context.TODO()))
			},
		},
		{
			"ShouldReturnClientAuthorizationPolicy",
			func(client *oidc.RegisteredClient) {
				client.AuthorizationPolicy = oidc.ClientAuthorizationPolicy{
					DefaultPolicy: authorization.OneFactor,
				}
			},
			func(t *testing.T, client *oidc.RegisteredClient) {
				assert.Equal(t, authorization.OneFactor, client.GetAuthorizationPolicy().DefaultPolicy)
			},
		},
		{
			"ShouldReturnClientAuthorizationPolicyEmpty",
			func(client *oidc.RegisteredClient) {
				client.AuthorizationPolicy = oidc.ClientAuthorizationPolicy{}
			},
			func(t *testing.T, client *oidc.RegisteredClient) {
				assert.Equal(t, authorization.Bypass, client.GetAuthorizationPolicy().DefaultPolicy)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := &oidc.RegisteredClient{}

			tc.setup(client)

			tc.expected(t, client)
		})
	}
}

func TestIsAuthenticationLevelSufficient(t *testing.T) {
	c := &oidc.RegisteredClient{}

	c.AuthorizationPolicy = oidc.ClientAuthorizationPolicy{DefaultPolicy: authorization.Bypass}
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.NotAuthenticated, authorization.Subject{}))
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.OneFactor, authorization.Subject{}))
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.TwoFactor, authorization.Subject{}))

	c.AuthorizationPolicy = oidc.ClientAuthorizationPolicy{DefaultPolicy: authorization.OneFactor}
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.NotAuthenticated, authorization.Subject{}))
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.OneFactor, authorization.Subject{}))
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.TwoFactor, authorization.Subject{}))

	c.AuthorizationPolicy = oidc.ClientAuthorizationPolicy{DefaultPolicy: authorization.TwoFactor}
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.NotAuthenticated, authorization.Subject{}))
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.OneFactor, authorization.Subject{}))
	assert.True(t, c.IsAuthenticationLevelSufficient(authentication.TwoFactor, authorization.Subject{}))

	c.AuthorizationPolicy = oidc.ClientAuthorizationPolicy{DefaultPolicy: authorization.Denied}
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.NotAuthenticated, authorization.Subject{}))
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.OneFactor, authorization.Subject{}))
	assert.False(t, c.IsAuthenticationLevelSufficient(authentication.TwoFactor, authorization.Subject{}))
}

func TestClient_GetConsentResponseBody(t *testing.T) {
	c := &oidc.RegisteredClient{}

	consentRequestBody := c.GetConsentResponseBody(nil)
	assert.Equal(t, "", consentRequestBody.ClientID)
	assert.Equal(t, "", consentRequestBody.ClientDescription)
	assert.Equal(t, []string(nil), consentRequestBody.Scopes)
	assert.Equal(t, []string(nil), consentRequestBody.Audience)

	c.ID = myclient
	c.Name = myclientdesc

	consent := &model.OAuth2ConsentSession{
		RequestedAudience: []string{examplecom},
		RequestedScopes:   []string{oidc.ScopeOpenID, oidc.ScopeGroups},
	}

	expectedScopes := []string{oidc.ScopeOpenID, oidc.ScopeGroups}
	expectedAudiences := []string{examplecom}

	consentRequestBody = c.GetConsentResponseBody(consent)
	assert.Equal(t, myclient, consentRequestBody.ClientID)
	assert.Equal(t, myclientdesc, consentRequestBody.ClientDescription)
	assert.Equal(t, expectedScopes, consentRequestBody.Scopes)
	assert.Equal(t, expectedAudiences, consentRequestBody.Audience)
}

func TestClient_GetAudience(t *testing.T) {
	c := &oidc.RegisteredClient{}

	audience := c.GetAudience()
	assert.Len(t, audience, 0)

	c.Audience = []string{examplecom}

	audience = c.GetAudience()
	require.Len(t, audience, 1)
	assert.Equal(t, examplecom, audience[0])
}

func TestClient_GetScopes(t *testing.T) {
	c := &oidc.RegisteredClient{}

	scopes := c.GetScopes()
	assert.Len(t, scopes, 0)

	c.Scopes = []string{oidc.ScopeOpenID}

	scopes = c.GetScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, oidc.ScopeOpenID, scopes[0])
}

func TestClient_GetGrantTypes(t *testing.T) {
	c := &oidc.RegisteredClient{}

	grantTypes := c.GetGrantTypes()
	require.Len(t, grantTypes, 1)
	assert.Equal(t, oidc.GrantTypeAuthorizationCode, grantTypes[0])

	c.GrantTypes = []string{"device_code"}

	grantTypes = c.GetGrantTypes()
	require.Len(t, grantTypes, 1)
	assert.Equal(t, "device_code", grantTypes[0])
}

func TestClient_Hashing(t *testing.T) {
	c := &oidc.RegisteredClient{}

	c.ClientSecret = &oidc.ClientSecretDigest{PasswordDigest: tOpenIDConnectPlainTextClientSecret}

	assert.True(t, c.ClientSecret.MatchBytes([]byte("client-secret")))
}

func TestClient_GetID(t *testing.T) {
	c := &oidc.RegisteredClient{}

	id := c.GetID()
	assert.Equal(t, "", id)

	c.ID = myclient

	id = c.GetID()
	assert.Equal(t, myclient, id)
}

func TestClient_GetRedirectURIs(t *testing.T) {
	c := &oidc.RegisteredClient{}

	redirectURIs := c.GetRedirectURIs()
	require.Len(t, redirectURIs, 0)

	c.RedirectURIs = []string{examplecom}

	redirectURIs = c.GetRedirectURIs()
	require.Len(t, redirectURIs, 1)
	assert.Equal(t, examplecom, redirectURIs[0])
}

func TestClient_GetResponseModes(t *testing.T) {
	c := &oidc.RegisteredClient{}

	responseModes := c.GetResponseModes()
	require.Len(t, responseModes, 0)

	c.ResponseModes = []oauthelia2.ResponseModeType{
		oauthelia2.ResponseModeDefault, oauthelia2.ResponseModeFormPost,
		oauthelia2.ResponseModeQuery, oauthelia2.ResponseModeFragment,
	}

	responseModes = c.GetResponseModes()
	require.Len(t, responseModes, 4)
	assert.Equal(t, oauthelia2.ResponseModeDefault, responseModes[0])
	assert.Equal(t, oauthelia2.ResponseModeFormPost, responseModes[1])
	assert.Equal(t, oauthelia2.ResponseModeQuery, responseModes[2])
	assert.Equal(t, oauthelia2.ResponseModeFragment, responseModes[3])
}

func TestClient_GetResponseTypes(t *testing.T) {
	c := &oidc.RegisteredClient{}

	responseTypes := c.GetResponseTypes()
	require.Len(t, responseTypes, 1)
	assert.Equal(t, oidc.ResponseTypeAuthorizationCodeFlow, responseTypes[0])

	c.ResponseTypes = []string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeImplicitFlowIDToken}

	responseTypes = c.GetResponseTypes()
	require.Len(t, responseTypes, 2)
	assert.Equal(t, oidc.ResponseTypeAuthorizationCodeFlow, responseTypes[0])
	assert.Equal(t, oidc.ResponseTypeImplicitFlowIDToken, responseTypes[1])
}

func TestNewClientPKCE(t *testing.T) {
	testCases := []struct {
		name                               string
		have                               schema.IdentityProvidersOpenIDConnectClient
		expectedEnforcePKCE                bool
		expectedEnforcePKCEChallengeMethod bool
		expected                           string
	}{
		{
			"ShouldNotEnforcePKCEAndNotErrorOnNonPKCERequest",
			schema.IdentityProvidersOpenIDConnectClient{},
			false,
			false,
			"",
		},
		{
			"ShouldEnforcePKCEAndErrorOnNonPKCERequest",
			schema.IdentityProvidersOpenIDConnectClient{RequirePKCE: true},
			true,
			false,
			"",
		},
		{
			"ShouldEnforcePKCEAndNotErrorOnPKCERequest",
			schema.IdentityProvidersOpenIDConnectClient{RequirePKCE: true},
			true,
			false,
			"",
		},
		{"ShouldEnforcePKCEFromChallengeMethodAndErrorOnNonPKCERequest",
			schema.IdentityProvidersOpenIDConnectClient{PKCEChallengeMethod: "S256"},
			true,
			true,
			"S256",
		},
		{"ShouldEnforcePKCEFromChallengeMethodAndErrorOnInvalidChallengeMethod",
			schema.IdentityProvidersOpenIDConnectClient{PKCEChallengeMethod: "S256"},
			true,
			true,
			"S256",
		},
		{"ShouldEnforcePKCEFromChallengeMethodAndNotErrorOnValidRequest",
			schema.IdentityProvidersOpenIDConnectClient{PKCEChallengeMethod: "S256"},
			true,
			true,
			"S256",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := oidc.NewClient(tc.have, &schema.IdentityProvidersOpenIDConnect{})

			assert.Equal(t, tc.expectedEnforcePKCE, client.GetEnforcePKCE())
			assert.Equal(t, tc.expectedEnforcePKCEChallengeMethod, client.GetEnforcePKCEChallengeMethod())
			assert.Equal(t, tc.expected, client.GetPKCEChallengeMethod())
		})
	}
}

func TestNewClientPAR(t *testing.T) {
	testCases := []struct {
		name     string
		have     schema.IdentityProvidersOpenIDConnectClient
		expected bool
	}{
		{
			"ShouldNotEnforcEPARAndNotErrorOnNonPARRequest",
			schema.IdentityProvidersOpenIDConnectClient{},
			false,
		},
		{
			"ShouldEnforcePARAndErrorOnNonPARRequest",
			schema.IdentityProvidersOpenIDConnectClient{RequirePushedAuthorizationRequests: true},
			true,
		},
		{
			"ShouldEnforcePARAndErrorOnNonPARRequest",
			schema.IdentityProvidersOpenIDConnectClient{RequirePushedAuthorizationRequests: true},
			true,
		},
		{
			"ShouldEnforcePARAndNotErrorOnPARRequest",
			schema.IdentityProvidersOpenIDConnectClient{RequirePushedAuthorizationRequests: true},
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := oidc.NewClient(tc.have, &schema.IdentityProvidersOpenIDConnect{})

			assert.Equal(t, tc.expected, client.GetRequirePushedAuthorizationRequests())
		})
	}
}

func TestClient_GetEffectiveLifespan(t *testing.T) {
	type subcase struct {
		name     string
		gt       oauthelia2.GrantType
		tt       oauthelia2.TokenType
		fallback time.Duration
		expected time.Duration
	}

	testCases := []struct {
		name     string
		have     schema.IdentityProvidersOpenIDConnectLifespan
		subcases []subcase
	}{
		{
			"ShouldHandleEdgeCases",
			schema.IdentityProvidersOpenIDConnectLifespan{
				IdentityProvidersOpenIDConnectLifespanToken: schema.IdentityProvidersOpenIDConnectLifespanToken{
					AccessToken:   time.Hour * 1,
					RefreshToken:  time.Hour * 2,
					IDToken:       time.Hour * 3,
					AuthorizeCode: time.Minute * 5,
				},
			},
			[]subcase{
				{
					"ShouldHandleInvalidTokenTypeFallbackToProvidedFallback",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.TokenType(abc),
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleInvalidGrantTypeFallbackToTokenType",
					oauthelia2.GrantType(abc),
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
			},
		},
		{
			"ShouldHandleUnconfiguredClient",
			schema.IdentityProvidersOpenIDConnectLifespan{},
			[]subcase{
				{
					"ShouldHandleAuthorizationCodeFlowAuthorizationCode",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleAuthorizationCodeFlowAccessToken",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.AccessToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleAuthorizationCodeFlowRefreshToken",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleAuthorizationCodeFlowIDToken",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.IDToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleImplicitFlowAuthorizationCode",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleImplicitFlowAccessToken",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.AccessToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleImplicitFlowRefreshToken",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleImplicitFlowIDToken",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.IDToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleClientCredentialsFlowAuthorizationCode",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleClientCredentialsFlowAccessToken",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.AccessToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleClientCredentialsFlowRefreshToken",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleClientCredentialsFlowIDToken",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.IDToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleRefreshTokenFlowAuthorizationCode",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleRefreshTokenFlowAccessToken",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.AccessToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleRefreshTokenFlowRefreshToken",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleRefreshTokenFlowIDToken",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.IDToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleJWTBearerFlowAuthorizationCode",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleJWTBearerFlowAccessToken",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.AccessToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleJWTBearerFlowRefreshToken",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleJWTBearerFlowIDToken",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.IDToken,
					time.Minute,
					time.Minute,
				},
			},
		},
		{
			"ShouldHandleConfiguredClientByTokenType",
			schema.IdentityProvidersOpenIDConnectLifespan{
				IdentityProvidersOpenIDConnectLifespanToken: schema.IdentityProvidersOpenIDConnectLifespanToken{
					AccessToken:   time.Hour * 1,
					RefreshToken:  time.Hour * 2,
					IDToken:       time.Hour * 3,
					AuthorizeCode: time.Minute * 5,
				},
			},
			[]subcase{
				{
					"ShouldHandleAuthorizationCodeFlowAuthorizationCode",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
				{
					"ShouldHandleAuthorizationCodeFlowAccessToken",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.AccessToken,
					time.Minute,
					time.Hour * 1,
				},
				{
					"ShouldHandleAuthorizationCodeFlowRefreshToken",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Hour * 2,
				},
				{
					"ShouldHandleAuthorizationCodeFlowIDToken",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.IDToken,
					time.Minute,
					time.Hour * 3,
				},
				{
					"ShouldHandleImplicitFlowAuthorizationCode",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
				{
					"ShouldHandleImplicitFlowAccessToken",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.AccessToken,
					time.Minute,
					time.Hour * 1,
				},
				{
					"ShouldHandleImplicitFlowRefreshToken",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Hour * 2,
				},
				{
					"ShouldHandleImplicitFlowIDToken",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.IDToken,
					time.Minute,
					time.Hour * 3,
				},
				{
					"ShouldHandleClientCredentialsFlowAuthorizationCode",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
				{
					"ShouldHandleClientCredentialsFlowAccessToken",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.AccessToken,
					time.Minute,
					time.Hour * 1,
				},
				{
					"ShouldHandleClientCredentialsFlowRefreshToken",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Hour * 2,
				},
				{
					"ShouldHandleClientCredentialsFlowIDToken",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.IDToken,
					time.Minute,
					time.Hour * 3,
				},
				{
					"ShouldHandleRefreshTokenFlowAuthorizationCode",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
				{
					"ShouldHandleRefreshTokenFlowAccessToken",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.AccessToken,
					time.Minute,
					time.Hour * 1,
				},
				{
					"ShouldHandleRefreshTokenFlowRefreshToken",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Hour * 2,
				},
				{
					"ShouldHandleRefreshTokenFlowIDToken",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.IDToken,
					time.Minute,
					time.Hour * 3,
				},
				{
					"ShouldHandleJWTBearerFlowAuthorizationCode",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
				{
					"ShouldHandleJWTBearerFlowAccessToken",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.AccessToken,
					time.Minute,
					time.Hour * 1,
				},
				{
					"ShouldHandleJWTBearerFlowRefreshToken",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Hour * 2,
				},
				{
					"ShouldHandleJWTBearerFlowIDToken",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.IDToken,
					time.Minute,
					time.Hour * 3,
				},
			},
		},
		{
			"ShouldHandleConfiguredClientByTokenTypeByGrantType",
			schema.IdentityProvidersOpenIDConnectLifespan{
				IdentityProvidersOpenIDConnectLifespanToken: schema.IdentityProvidersOpenIDConnectLifespanToken{
					AccessToken:   time.Hour * 1,
					RefreshToken:  time.Hour * 2,
					IDToken:       time.Hour * 3,
					AuthorizeCode: time.Minute * 5,
				},
				Grants: schema.IdentityProvidersOpenIDConnectLifespanGrants{
					AuthorizeCode: schema.IdentityProvidersOpenIDConnectLifespanToken{
						AccessToken:   time.Hour * 11,
						RefreshToken:  time.Hour * 12,
						IDToken:       time.Hour * 13,
						AuthorizeCode: time.Minute * 15,
					},
					Implicit: schema.IdentityProvidersOpenIDConnectLifespanToken{
						AccessToken:   time.Hour * 21,
						RefreshToken:  time.Hour * 22,
						IDToken:       time.Hour * 23,
						AuthorizeCode: time.Minute * 25,
					},
					ClientCredentials: schema.IdentityProvidersOpenIDConnectLifespanToken{
						AccessToken:   time.Hour * 31,
						RefreshToken:  time.Hour * 32,
						IDToken:       time.Hour * 33,
						AuthorizeCode: time.Minute * 35,
					},
					RefreshToken: schema.IdentityProvidersOpenIDConnectLifespanToken{
						AccessToken:   time.Hour * 41,
						RefreshToken:  time.Hour * 42,
						IDToken:       time.Hour * 43,
						AuthorizeCode: time.Minute * 45,
					},
					JWTBearer: schema.IdentityProvidersOpenIDConnectLifespanToken{
						AccessToken:   time.Hour * 51,
						RefreshToken:  time.Hour * 52,
						IDToken:       time.Hour * 53,
						AuthorizeCode: time.Minute * 55,
					},
				},
			},
			[]subcase{
				{
					"ShouldHandleAuthorizationCodeFlowAuthorizationCode",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 15,
				},
				{
					"ShouldHandleAuthorizationCodeFlowAccessToken",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.AccessToken,
					time.Minute,
					time.Hour * 11,
				},
				{
					"ShouldHandleAuthorizationCodeFlowRefreshToken",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Hour * 12,
				},
				{
					"ShouldHandleAuthorizationCodeFlowIDToken",
					oauthelia2.GrantTypeAuthorizationCode,
					oauthelia2.IDToken,
					time.Minute,
					time.Hour * 13,
				},
				{
					"ShouldHandleImplicitFlowAuthorizationCode",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 25,
				},
				{
					"ShouldHandleImplicitFlowAccessToken",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.AccessToken,
					time.Minute,
					time.Hour * 21,
				},
				{
					"ShouldHandleImplicitFlowRefreshToken",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Hour * 22,
				},
				{
					"ShouldHandleImplicitFlowIDToken",
					oauthelia2.GrantTypeImplicit,
					oauthelia2.IDToken,
					time.Minute,
					time.Hour * 23,
				},
				{
					"ShouldHandleClientCredentialsFlowAuthorizationCode",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 35,
				},
				{
					"ShouldHandleClientCredentialsFlowAccessToken",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.AccessToken,
					time.Minute,
					time.Hour * 31,
				},
				{
					"ShouldHandleClientCredentialsFlowRefreshToken",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Hour * 32,
				},
				{
					"ShouldHandleClientCredentialsFlowIDToken",
					oauthelia2.GrantTypeClientCredentials,
					oauthelia2.IDToken,
					time.Minute,
					time.Hour * 33,
				},
				{
					"ShouldHandleRefreshTokenFlowAuthorizationCode",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 45,
				},
				{
					"ShouldHandleRefreshTokenFlowAccessToken",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.AccessToken,
					time.Minute,
					time.Hour * 41,
				},
				{
					"ShouldHandleRefreshTokenFlowRefreshToken",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Hour * 42,
				},
				{
					"ShouldHandleRefreshTokenFlowIDToken",
					oauthelia2.GrantTypeRefreshToken,
					oauthelia2.IDToken,
					time.Minute,
					time.Hour * 43,
				},
				{
					"ShouldHandleJWTBearerFlowAuthorizationCode",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.AuthorizeCode,
					time.Minute,
					time.Minute * 55,
				},
				{
					"ShouldHandleJWTBearerFlowAccessToken",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.AccessToken,
					time.Minute,
					time.Hour * 51,
				},
				{
					"ShouldHandleJWTBearerFlowRefreshToken",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.RefreshToken,
					time.Minute,
					time.Hour * 52,
				},
				{
					"ShouldHandleJWTBearerFlowIDToken",
					oauthelia2.GrantTypeJWTBearer,
					oauthelia2.IDToken,
					time.Minute,
					time.Hour * 53,
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{
				ID:       "test",
				Lifespan: "test",
			}, &schema.IdentityProvidersOpenIDConnect{
				Lifespans: schema.IdentityProvidersOpenIDConnectLifespans{
					Custom: map[string]schema.IdentityProvidersOpenIDConnectLifespan{
						"test": tc.have,
					},
				},
			})

			for _, stc := range tc.subcases {
				t.Run(stc.name, func(t *testing.T) {
					assert.Equal(t, stc.expected, client.GetEffectiveLifespan(stc.gt, stc.tt, stc.fallback))
				})
			}
		})
	}
}

func TestNewClientResponseModes(t *testing.T) {
	testCases := []struct {
		name     string
		have     schema.IdentityProvidersOpenIDConnectClient
		expected []oauthelia2.ResponseModeType
		r        *oauthelia2.AuthorizeRequest
		err      string
		desc     string
	}{
		{
			"ShouldEnforceResponseModePolicyAndAllowDefaultModeQuery",
			schema.IdentityProvidersOpenIDConnectClient{ResponseModes: []string{oidc.ResponseModeQuery}},
			[]oauthelia2.ResponseModeType{oauthelia2.ResponseModeQuery},
			&oauthelia2.AuthorizeRequest{DefaultResponseMode: oauthelia2.ResponseModeQuery, ResponseMode: oauthelia2.ResponseModeDefault, Request: oauthelia2.Request{Form: map[string][]string{oidc.FormParameterResponseMode: nil}}},
			"",
			"",
		},
		{
			"ShouldEnforceResponseModePolicyAndFailOnDefaultMode",
			schema.IdentityProvidersOpenIDConnectClient{ResponseModes: []string{oidc.ResponseModeFormPost}},
			[]oauthelia2.ResponseModeType{oauthelia2.ResponseModeFormPost},
			&oauthelia2.AuthorizeRequest{DefaultResponseMode: oauthelia2.ResponseModeQuery, ResponseMode: oauthelia2.ResponseModeDefault, Request: oauthelia2.Request{Form: map[string][]string{oidc.FormParameterResponseMode: nil}}},
			"unsupported_response_mode",
			"The authorization server does not support obtaining a response using this response mode. The request omitted the response_mode making the default response_mode 'query' based on the other authorization request parameters but registered OAuth 2.0 client doesn't support this response_mode",
		},
		{
			"ShouldNotEnforceConfiguredResponseMode",
			schema.IdentityProvidersOpenIDConnectClient{ResponseModes: []string{oidc.ResponseModeFormPost}},
			[]oauthelia2.ResponseModeType{oauthelia2.ResponseModeFormPost},
			&oauthelia2.AuthorizeRequest{DefaultResponseMode: oauthelia2.ResponseModeQuery, ResponseMode: oauthelia2.ResponseModeQuery, Request: oauthelia2.Request{Form: map[string][]string{oidc.FormParameterResponseMode: {oidc.ResponseModeQuery}}}},
			"",
			"",
		},
		{
			"ShouldNotEnforceUnconfiguredResponseMode",
			schema.IdentityProvidersOpenIDConnectClient{ResponseModes: []string{}},
			[]oauthelia2.ResponseModeType{},
			&oauthelia2.AuthorizeRequest{DefaultResponseMode: oauthelia2.ResponseModeQuery, ResponseMode: oauthelia2.ResponseModeDefault, Request: oauthelia2.Request{Form: map[string][]string{oidc.FormParameterResponseMode: {oidc.ResponseModeQuery}}}},
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := oidc.NewClient(tc.have, &schema.IdentityProvidersOpenIDConnect{})

			assert.Equal(t, tc.expected, client.GetResponseModes())

			if tc.r != nil {
				err := client.ValidateResponseModePolicy(tc.r)

				if tc.err != "" {
					require.NotNil(t, err)
					assert.EqualError(t, err, tc.err)
					assert.Equal(t, tc.desc, oauthelia2.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestClient_IsPublic(t *testing.T) {
	c := &oidc.RegisteredClient{}

	assert.False(t, c.IsPublic())

	c.Public = true
	assert.True(t, c.IsPublic())
}

func TestNewClient_JSONWebKeySetURI(t *testing.T) {
	var (
		client     oidc.Client
		registered *oidc.RegisteredClient
		ok         bool
	)

	client = oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{
		TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
		JSONWebKeysURI:          MustParseRequestURI("https://google.com"),
	}, &schema.IdentityProvidersOpenIDConnect{})

	require.NotNil(t, client)

	registered, ok = client.(*oidc.RegisteredClient)

	require.True(t, ok)

	assert.Equal(t, "https://google.com", registered.GetJSONWebKeysURI())

	client = oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{
		TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
		JSONWebKeysURI:          nil,
	}, &schema.IdentityProvidersOpenIDConnect{})

	require.NotNil(t, client)

	registered, ok = client.(*oidc.RegisteredClient)

	require.True(t, ok)

	assert.Equal(t, "", registered.GetJSONWebKeysURI())
}
