package oidc_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	jose "gopkg.in/square/go-jose.v2"

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
	assert.Equal(t, "", client.GetDescription())
	assert.Len(t, client.GetResponseModes(), 0)
	assert.Len(t, client.GetResponseTypes(), 1)
	assert.Equal(t, "", client.GetSectorIdentifier())

	bclient, ok := client.(*oidc.BaseClient)
	require.True(t, ok)
	assert.Equal(t, "", bclient.UserinfoSigningAlg)
	assert.Equal(t, oidc.SigningAlgNone, client.GetUserinfoSigningAlg())
	assert.Equal(t, "", client.GetUserinfoSigningKeyID())

	_, ok = client.(*oidc.FullClient)
	assert.False(t, ok)

	config = schema.IdentityProvidersOpenIDConnectClient{
		ID:                  myclient,
		Description:         myclientdesc,
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
	assert.Equal(t, fosite.ResponseModeFormPost, client.GetResponseModes()[0])
	assert.Equal(t, authorization.TwoFactor, client.GetAuthorizationPolicyRequiredLevel(authorization.Subject{}))

	config = schema.IdentityProvidersOpenIDConnectClient{
		TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
	}

	client = oidc.NewClient(config, &schema.IdentityProvidersOpenIDConnect{})

	fclient, ok := client.(*oidc.FullClient)

	require.True(t, ok)

	assert.Equal(t, "", fclient.UserinfoSigningAlg)
	assert.Equal(t, oidc.SigningAlgNone, client.GetUserinfoSigningAlg())
	assert.Equal(t, oidc.SigningAlgNone, fclient.GetUserinfoSigningAlg())
	assert.Equal(t, oidc.SigningAlgNone, fclient.UserinfoSigningAlg)

	assert.Equal(t, "", fclient.UserinfoSigningKeyID)
	assert.Equal(t, "", client.GetUserinfoSigningKeyID())
	assert.Equal(t, "", fclient.GetUserinfoSigningKeyID())

	fclient.UserinfoSigningKeyID = "aukeyid"

	assert.Equal(t, "aukeyid", client.GetUserinfoSigningKeyID())
	assert.Equal(t, "aukeyid", fclient.GetUserinfoSigningKeyID())

	assert.Equal(t, "", fclient.IDTokenSigningAlg)
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, client.GetIDTokenSigningAlg())
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, fclient.GetIDTokenSigningAlg())
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, fclient.IDTokenSigningAlg)

	assert.Equal(t, "", fclient.IDTokenSigningKeyID)
	assert.Equal(t, "", client.GetIDTokenSigningKeyID())
	assert.Equal(t, "", fclient.GetIDTokenSigningKeyID())

	fclient.IDTokenSigningKeyID = "akeyid"

	assert.Equal(t, "akeyid", client.GetIDTokenSigningKeyID())
	assert.Equal(t, "akeyid", fclient.GetIDTokenSigningKeyID())

	assert.Equal(t, oidc.ClientAuthMethodClientSecretPost, fclient.TokenEndpointAuthMethod)
	assert.Equal(t, oidc.ClientAuthMethodClientSecretPost, fclient.GetTokenEndpointAuthMethod())

	assert.Equal(t, "", fclient.TokenEndpointAuthSigningAlgorithm)
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, fclient.GetTokenEndpointAuthSigningAlgorithm())
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, fclient.TokenEndpointAuthSigningAlgorithm)

	assert.Equal(t, "", fclient.RequestObjectSigningAlgorithm)
	assert.Equal(t, "", fclient.GetRequestObjectSigningAlgorithm())

	fclient.RequestObjectSigningAlgorithm = oidc.SigningAlgRSAUsingSHA256

	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, fclient.GetRequestObjectSigningAlgorithm())

	assert.Equal(t, "", fclient.JSONWebKeysURI)
	assert.Equal(t, "", fclient.GetJSONWebKeysURI())

	fclient.JSONWebKeysURI = "https://example.com"
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

func TestBaseClient_ValidatePARPolicy(t *testing.T) {
	testCases := []struct {
		name     string
		client   *oidc.BaseClient
		have     *fosite.Request
		expected string
	}{
		{
			"ShouldNotEnforcePAR",
			&oidc.BaseClient{
				EnforcePAR: false,
			},
			&fosite.Request{},
			"",
		},
		{
			"ShouldEnforcePARAndErrorWithoutCorrectRequestURI",
			&oidc.BaseClient{
				EnforcePAR: true,
			},
			&fosite.Request{
				Form: map[string][]string{
					oidc.FormParameterRequestURI: {"https://google.com"},
				},
			},
			"invalid_request",
		},
		{
			"ShouldEnforcePARAndErrorWithEmptyRequestURI",
			&oidc.BaseClient{
				EnforcePAR: true,
			},
			&fosite.Request{
				Form: map[string][]string{
					oidc.FormParameterRequestURI: {""},
				},
			},
			"invalid_request",
		},
		{
			"ShouldEnforcePARAndNotErrorWithCorrectRequestURI",
			&oidc.BaseClient{
				EnforcePAR: true,
			},
			&fosite.Request{
				Form: map[string][]string{
					oidc.FormParameterRequestURI: {oidc.RedirectURIPrefixPushedAuthorizationRequestURN + "abc"},
				},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.client.ValidatePARPolicy(tc.have, oidc.RedirectURIPrefixPushedAuthorizationRequestURN)

			switch tc.expected {
			case "":
				assert.NoError(t, err)
			default:
				assert.EqualError(t, err, tc.expected)
			}
		})
	}
}

func TestIsAuthenticationLevelSufficient(t *testing.T) {
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

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
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

	consentRequestBody := c.GetConsentResponseBody(nil)
	assert.Equal(t, "", consentRequestBody.ClientID)
	assert.Equal(t, "", consentRequestBody.ClientDescription)
	assert.Equal(t, []string(nil), consentRequestBody.Scopes)
	assert.Equal(t, []string(nil), consentRequestBody.Audience)

	c.ID = myclient
	c.Description = myclientdesc

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
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

	audience := c.GetAudience()
	assert.Len(t, audience, 0)

	c.Audience = []string{examplecom}

	audience = c.GetAudience()
	require.Len(t, audience, 1)
	assert.Equal(t, examplecom, audience[0])
}

func TestClient_GetScopes(t *testing.T) {
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

	scopes := c.GetScopes()
	assert.Len(t, scopes, 0)

	c.Scopes = []string{oidc.ScopeOpenID}

	scopes = c.GetScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, oidc.ScopeOpenID, scopes[0])
}

func TestClient_GetGrantTypes(t *testing.T) {
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

	grantTypes := c.GetGrantTypes()
	require.Len(t, grantTypes, 1)
	assert.Equal(t, oidc.GrantTypeAuthorizationCode, grantTypes[0])

	c.GrantTypes = []string{"device_code"}

	grantTypes = c.GetGrantTypes()
	require.Len(t, grantTypes, 1)
	assert.Equal(t, "device_code", grantTypes[0])
}

func TestClient_Hashing(t *testing.T) {
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

	hashedSecret := c.GetHashedSecret()
	assert.Equal(t, []byte(nil), hashedSecret)

	c.Secret = tOpenIDConnectPlainTextClientSecret

	assert.True(t, c.Secret.MatchBytes([]byte("client-secret")))
}

func TestClient_GetHashedSecret(t *testing.T) {
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

	hashedSecret := c.GetHashedSecret()
	assert.Equal(t, []byte(nil), hashedSecret)

	c.Secret = tOpenIDConnectPlainTextClientSecret

	hashedSecret = c.GetHashedSecret()
	assert.Equal(t, []byte("$plaintext$client-secret"), hashedSecret)
}

func TestClient_GetID(t *testing.T) {
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

	id := c.GetID()
	assert.Equal(t, "", id)

	c.ID = myclient

	id = c.GetID()
	assert.Equal(t, myclient, id)
}

func TestClient_GetRedirectURIs(t *testing.T) {
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

	redirectURIs := c.GetRedirectURIs()
	require.Len(t, redirectURIs, 0)

	c.RedirectURIs = []string{examplecom}

	redirectURIs = c.GetRedirectURIs()
	require.Len(t, redirectURIs, 1)
	assert.Equal(t, examplecom, redirectURIs[0])
}

func TestClient_GetResponseModes(t *testing.T) {
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

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
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

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
		r                                  *fosite.Request
		err                                string
		desc                               string
	}{
		{
			"ShouldNotEnforcePKCEAndNotErrorOnNonPKCERequest",
			schema.IdentityProvidersOpenIDConnectClient{},
			false,
			false,
			"",
			&fosite.Request{},
			"",
			"",
		},
		{
			"ShouldEnforcePKCEAndErrorOnNonPKCERequest",
			schema.IdentityProvidersOpenIDConnectClient{EnforcePKCE: true},
			true,
			false,
			"",
			&fosite.Request{},
			"invalid_request",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for this client.",
		},
		{
			"ShouldEnforcePKCEAndNotErrorOnPKCERequest",
			schema.IdentityProvidersOpenIDConnectClient{EnforcePKCE: true},
			true,
			false,
			"",
			&fosite.Request{Form: map[string][]string{"code_challenge": {"abc"}}},
			"",
			"",
		},
		{"ShouldEnforcePKCEFromChallengeMethodAndErrorOnNonPKCERequest",
			schema.IdentityProvidersOpenIDConnectClient{PKCEChallengeMethod: "S256"},
			true,
			true,
			"S256",
			&fosite.Request{},
			"invalid_request",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for this client.",
		},
		{"ShouldEnforcePKCEFromChallengeMethodAndErrorOnInvalidChallengeMethod",
			schema.IdentityProvidersOpenIDConnectClient{PKCEChallengeMethod: "S256"},
			true,
			true,
			"S256",
			&fosite.Request{Form: map[string][]string{"code_challenge": {"abc"}}},
			"invalid_request",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Client must use code_challenge_method=S256,  is not allowed. The server is configured in a way that enforces PKCE S256 as challenge method for this client.",
		},
		{"ShouldEnforcePKCEFromChallengeMethodAndNotErrorOnValidRequest",
			schema.IdentityProvidersOpenIDConnectClient{PKCEChallengeMethod: "S256"},
			true,
			true,
			"S256",
			&fosite.Request{Form: map[string][]string{"code_challenge": {"abc"}, "code_challenge_method": {"S256"}}},
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := oidc.NewClient(tc.have, &schema.IdentityProvidersOpenIDConnect{})

			assert.Equal(t, tc.expectedEnforcePKCE, client.GetPKCEEnforcement())
			assert.Equal(t, tc.expectedEnforcePKCEChallengeMethod, client.GetPKCEChallengeMethodEnforcement())
			assert.Equal(t, tc.expected, client.GetPKCEChallengeMethod())

			if tc.r != nil {
				err := client.ValidatePKCEPolicy(tc.r)

				if tc.err != "" {
					require.NotNil(t, err)
					assert.EqualError(t, err, tc.err)
					assert.Equal(t, tc.desc, fosite.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestNewClientPAR(t *testing.T) {
	testCases := []struct {
		name     string
		have     schema.IdentityProvidersOpenIDConnectClient
		expected bool
		r        *fosite.Request
		err      string
		desc     string
	}{
		{
			"ShouldNotEnforcEPARAndNotErrorOnNonPARRequest",
			schema.IdentityProvidersOpenIDConnectClient{},
			false,
			&fosite.Request{},
			"",
			"",
		},
		{
			"ShouldEnforcePARAndErrorOnNonPARRequest",
			schema.IdentityProvidersOpenIDConnectClient{EnforcePAR: true},
			true,
			&fosite.Request{},
			"invalid_request",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Pushed Authorization Requests are enforced for this client but no such request was sent. The request_uri parameter was empty.",
		},
		{
			"ShouldEnforcePARAndErrorOnNonPARRequest",
			schema.IdentityProvidersOpenIDConnectClient{EnforcePAR: true},
			true,
			&fosite.Request{Form: map[string][]string{oidc.FormParameterRequestURI: {"https://example.com"}}},
			"invalid_request",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Pushed Authorization Requests are enforced for this client but no such request was sent. The request_uri parameter 'https://example.com' is malformed."},
		{
			"ShouldEnforcePARAndNotErrorOnPARRequest",
			schema.IdentityProvidersOpenIDConnectClient{EnforcePAR: true},
			true,
			&fosite.Request{Form: map[string][]string{oidc.FormParameterRequestURI: {fmt.Sprintf("%sabc", oidc.RedirectURIPrefixPushedAuthorizationRequestURN)}}},
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := oidc.NewClient(tc.have, &schema.IdentityProvidersOpenIDConnect{})

			assert.Equal(t, tc.expected, client.GetPAREnforcement())

			if tc.r != nil {
				err := client.ValidatePARPolicy(tc.r, oidc.RedirectURIPrefixPushedAuthorizationRequestURN)

				if tc.err != "" {
					require.NotNil(t, err)
					assert.EqualError(t, err, tc.err)
					assert.Equal(t, tc.desc, fosite.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestClient_GetEffectiveLifespan(t *testing.T) {
	type subcase struct {
		name     string
		gt       fosite.GrantType
		tt       fosite.TokenType
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
					fosite.GrantTypeAuthorizationCode,
					fosite.TokenType("abc"),
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleInvalidGrantTypeFallbackToTokenType",
					fosite.GrantType("abc"),
					fosite.AuthorizeCode,
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
					fosite.GrantTypeAuthorizationCode,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleAuthorizationCodeFlowAccessToken",
					fosite.GrantTypeAuthorizationCode,
					fosite.AccessToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleAuthorizationCodeFlowRefreshToken",
					fosite.GrantTypeAuthorizationCode,
					fosite.RefreshToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleAuthorizationCodeFlowIDToken",
					fosite.GrantTypeAuthorizationCode,
					fosite.IDToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleImplicitFlowAuthorizationCode",
					fosite.GrantTypeImplicit,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleImplicitFlowAccessToken",
					fosite.GrantTypeImplicit,
					fosite.AccessToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleImplicitFlowRefreshToken",
					fosite.GrantTypeImplicit,
					fosite.RefreshToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleImplicitFlowIDToken",
					fosite.GrantTypeImplicit,
					fosite.IDToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleClientCredentialsFlowAuthorizationCode",
					fosite.GrantTypeClientCredentials,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleClientCredentialsFlowAccessToken",
					fosite.GrantTypeClientCredentials,
					fosite.AccessToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleClientCredentialsFlowRefreshToken",
					fosite.GrantTypeClientCredentials,
					fosite.RefreshToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleClientCredentialsFlowIDToken",
					fosite.GrantTypeClientCredentials,
					fosite.IDToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleRefreshTokenFlowAuthorizationCode",
					fosite.GrantTypeRefreshToken,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleRefreshTokenFlowAccessToken",
					fosite.GrantTypeRefreshToken,
					fosite.AccessToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleRefreshTokenFlowRefreshToken",
					fosite.GrantTypeRefreshToken,
					fosite.RefreshToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleRefreshTokenFlowIDToken",
					fosite.GrantTypeRefreshToken,
					fosite.IDToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleJWTBearerFlowAuthorizationCode",
					fosite.GrantTypeJWTBearer,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleJWTBearerFlowAccessToken",
					fosite.GrantTypeJWTBearer,
					fosite.AccessToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleJWTBearerFlowRefreshToken",
					fosite.GrantTypeJWTBearer,
					fosite.RefreshToken,
					time.Minute,
					time.Minute,
				},
				{
					"ShouldHandleJWTBearerFlowIDToken",
					fosite.GrantTypeJWTBearer,
					fosite.IDToken,
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
					fosite.GrantTypeAuthorizationCode,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
				{
					"ShouldHandleAuthorizationCodeFlowAccessToken",
					fosite.GrantTypeAuthorizationCode,
					fosite.AccessToken,
					time.Minute,
					time.Hour * 1,
				},
				{
					"ShouldHandleAuthorizationCodeFlowRefreshToken",
					fosite.GrantTypeAuthorizationCode,
					fosite.RefreshToken,
					time.Minute,
					time.Hour * 2,
				},
				{
					"ShouldHandleAuthorizationCodeFlowIDToken",
					fosite.GrantTypeAuthorizationCode,
					fosite.IDToken,
					time.Minute,
					time.Hour * 3,
				},
				{
					"ShouldHandleImplicitFlowAuthorizationCode",
					fosite.GrantTypeImplicit,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
				{
					"ShouldHandleImplicitFlowAccessToken",
					fosite.GrantTypeImplicit,
					fosite.AccessToken,
					time.Minute,
					time.Hour * 1,
				},
				{
					"ShouldHandleImplicitFlowRefreshToken",
					fosite.GrantTypeImplicit,
					fosite.RefreshToken,
					time.Minute,
					time.Hour * 2,
				},
				{
					"ShouldHandleImplicitFlowIDToken",
					fosite.GrantTypeImplicit,
					fosite.IDToken,
					time.Minute,
					time.Hour * 3,
				},
				{
					"ShouldHandleClientCredentialsFlowAuthorizationCode",
					fosite.GrantTypeClientCredentials,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
				{
					"ShouldHandleClientCredentialsFlowAccessToken",
					fosite.GrantTypeClientCredentials,
					fosite.AccessToken,
					time.Minute,
					time.Hour * 1,
				},
				{
					"ShouldHandleClientCredentialsFlowRefreshToken",
					fosite.GrantTypeClientCredentials,
					fosite.RefreshToken,
					time.Minute,
					time.Hour * 2,
				},
				{
					"ShouldHandleClientCredentialsFlowIDToken",
					fosite.GrantTypeClientCredentials,
					fosite.IDToken,
					time.Minute,
					time.Hour * 3,
				},
				{
					"ShouldHandleRefreshTokenFlowAuthorizationCode",
					fosite.GrantTypeRefreshToken,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
				{
					"ShouldHandleRefreshTokenFlowAccessToken",
					fosite.GrantTypeRefreshToken,
					fosite.AccessToken,
					time.Minute,
					time.Hour * 1,
				},
				{
					"ShouldHandleRefreshTokenFlowRefreshToken",
					fosite.GrantTypeRefreshToken,
					fosite.RefreshToken,
					time.Minute,
					time.Hour * 2,
				},
				{
					"ShouldHandleRefreshTokenFlowIDToken",
					fosite.GrantTypeRefreshToken,
					fosite.IDToken,
					time.Minute,
					time.Hour * 3,
				},
				{
					"ShouldHandleJWTBearerFlowAuthorizationCode",
					fosite.GrantTypeJWTBearer,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute * 5,
				},
				{
					"ShouldHandleJWTBearerFlowAccessToken",
					fosite.GrantTypeJWTBearer,
					fosite.AccessToken,
					time.Minute,
					time.Hour * 1,
				},
				{
					"ShouldHandleJWTBearerFlowRefreshToken",
					fosite.GrantTypeJWTBearer,
					fosite.RefreshToken,
					time.Minute,
					time.Hour * 2,
				},
				{
					"ShouldHandleJWTBearerFlowIDToken",
					fosite.GrantTypeJWTBearer,
					fosite.IDToken,
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
					fosite.GrantTypeAuthorizationCode,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute * 15,
				},
				{
					"ShouldHandleAuthorizationCodeFlowAccessToken",
					fosite.GrantTypeAuthorizationCode,
					fosite.AccessToken,
					time.Minute,
					time.Hour * 11,
				},
				{
					"ShouldHandleAuthorizationCodeFlowRefreshToken",
					fosite.GrantTypeAuthorizationCode,
					fosite.RefreshToken,
					time.Minute,
					time.Hour * 12,
				},
				{
					"ShouldHandleAuthorizationCodeFlowIDToken",
					fosite.GrantTypeAuthorizationCode,
					fosite.IDToken,
					time.Minute,
					time.Hour * 13,
				},
				{
					"ShouldHandleImplicitFlowAuthorizationCode",
					fosite.GrantTypeImplicit,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute * 25,
				},
				{
					"ShouldHandleImplicitFlowAccessToken",
					fosite.GrantTypeImplicit,
					fosite.AccessToken,
					time.Minute,
					time.Hour * 21,
				},
				{
					"ShouldHandleImplicitFlowRefreshToken",
					fosite.GrantTypeImplicit,
					fosite.RefreshToken,
					time.Minute,
					time.Hour * 22,
				},
				{
					"ShouldHandleImplicitFlowIDToken",
					fosite.GrantTypeImplicit,
					fosite.IDToken,
					time.Minute,
					time.Hour * 23,
				},
				{
					"ShouldHandleClientCredentialsFlowAuthorizationCode",
					fosite.GrantTypeClientCredentials,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute * 35,
				},
				{
					"ShouldHandleClientCredentialsFlowAccessToken",
					fosite.GrantTypeClientCredentials,
					fosite.AccessToken,
					time.Minute,
					time.Hour * 31,
				},
				{
					"ShouldHandleClientCredentialsFlowRefreshToken",
					fosite.GrantTypeClientCredentials,
					fosite.RefreshToken,
					time.Minute,
					time.Hour * 32,
				},
				{
					"ShouldHandleClientCredentialsFlowIDToken",
					fosite.GrantTypeClientCredentials,
					fosite.IDToken,
					time.Minute,
					time.Hour * 33,
				},
				{
					"ShouldHandleRefreshTokenFlowAuthorizationCode",
					fosite.GrantTypeRefreshToken,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute * 45,
				},
				{
					"ShouldHandleRefreshTokenFlowAccessToken",
					fosite.GrantTypeRefreshToken,
					fosite.AccessToken,
					time.Minute,
					time.Hour * 41,
				},
				{
					"ShouldHandleRefreshTokenFlowRefreshToken",
					fosite.GrantTypeRefreshToken,
					fosite.RefreshToken,
					time.Minute,
					time.Hour * 42,
				},
				{
					"ShouldHandleRefreshTokenFlowIDToken",
					fosite.GrantTypeRefreshToken,
					fosite.IDToken,
					time.Minute,
					time.Hour * 43,
				},
				{
					"ShouldHandleJWTBearerFlowAuthorizationCode",
					fosite.GrantTypeJWTBearer,
					fosite.AuthorizeCode,
					time.Minute,
					time.Minute * 55,
				},
				{
					"ShouldHandleJWTBearerFlowAccessToken",
					fosite.GrantTypeJWTBearer,
					fosite.AccessToken,
					time.Minute,
					time.Hour * 51,
				},
				{
					"ShouldHandleJWTBearerFlowRefreshToken",
					fosite.GrantTypeJWTBearer,
					fosite.RefreshToken,
					time.Minute,
					time.Hour * 52,
				},
				{
					"ShouldHandleJWTBearerFlowIDToken",
					fosite.GrantTypeJWTBearer,
					fosite.IDToken,
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
		expected []fosite.ResponseModeType
		r        *fosite.AuthorizeRequest
		err      string
		desc     string
	}{
		{
			"ShouldEnforceResponseModePolicyAndAllowDefaultModeQuery",
			schema.IdentityProvidersOpenIDConnectClient{ResponseModes: []string{oidc.ResponseModeQuery}},
			[]fosite.ResponseModeType{fosite.ResponseModeQuery},
			&fosite.AuthorizeRequest{DefaultResponseMode: fosite.ResponseModeQuery, ResponseMode: fosite.ResponseModeDefault, Request: fosite.Request{Form: map[string][]string{oidc.FormParameterResponseMode: nil}}},
			"",
			"",
		},
		{
			"ShouldEnforceResponseModePolicyAndFailOnDefaultMode",
			schema.IdentityProvidersOpenIDConnectClient{ResponseModes: []string{oidc.ResponseModeFormPost}},
			[]fosite.ResponseModeType{fosite.ResponseModeFormPost},
			&fosite.AuthorizeRequest{DefaultResponseMode: fosite.ResponseModeQuery, ResponseMode: fosite.ResponseModeDefault, Request: fosite.Request{Form: map[string][]string{oidc.FormParameterResponseMode: nil}}},
			"unsupported_response_mode",
			"The authorization server does not support obtaining a response using this response mode. The request omitted the response_mode making the default response_mode 'query' based on the other authorization request parameters but registered OAuth 2.0 client doesn't support this response_mode",
		},
		{
			"ShouldNotEnforceConfiguredResponseMode",
			schema.IdentityProvidersOpenIDConnectClient{ResponseModes: []string{oidc.ResponseModeFormPost}},
			[]fosite.ResponseModeType{fosite.ResponseModeFormPost},
			&fosite.AuthorizeRequest{DefaultResponseMode: fosite.ResponseModeQuery, ResponseMode: fosite.ResponseModeQuery, Request: fosite.Request{Form: map[string][]string{oidc.FormParameterResponseMode: {oidc.ResponseModeQuery}}}},
			"",
			"",
		},
		{
			"ShouldNotEnforceUnconfiguredResponseMode",
			schema.IdentityProvidersOpenIDConnectClient{ResponseModes: []string{}},
			[]fosite.ResponseModeType{},
			&fosite.AuthorizeRequest{DefaultResponseMode: fosite.ResponseModeQuery, ResponseMode: fosite.ResponseModeDefault, Request: fosite.Request{Form: map[string][]string{oidc.FormParameterResponseMode: {oidc.ResponseModeQuery}}}},
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
					assert.Equal(t, tc.desc, fosite.ErrorToRFC6749Error(err).WithExposeDebug(true).GetDescription())
				} else {
					assert.NoError(t, err)
				}
			}
		})
	}
}

func TestClient_IsPublic(t *testing.T) {
	c := &oidc.FullClient{BaseClient: &oidc.BaseClient{}}

	assert.False(t, c.IsPublic())

	c.Public = true
	assert.True(t, c.IsPublic())
}

func TestNewClient_JSONWebKeySetURI(t *testing.T) {
	var (
		client  oidc.Client
		clientf *oidc.FullClient
		ok      bool
	)

	client = oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{
		TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
		PublicKeys: schema.IdentityProvidersOpenIDConnectClientPublicKeys{
			URI: MustParseRequestURI("https://google.com"),
		},
	}, &schema.IdentityProvidersOpenIDConnect{})

	require.NotNil(t, client)

	clientf, ok = client.(*oidc.FullClient)

	require.True(t, ok)

	assert.Equal(t, "https://google.com", clientf.GetJSONWebKeysURI())

	client = oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{
		TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
		PublicKeys: schema.IdentityProvidersOpenIDConnectClientPublicKeys{
			URI: nil,
		},
	}, &schema.IdentityProvidersOpenIDConnect{})

	require.NotNil(t, client)

	clientf, ok = client.(*oidc.FullClient)

	require.True(t, ok)

	assert.Equal(t, "", clientf.GetJSONWebKeysURI())
}
