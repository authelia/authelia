package oidc

import (
	"fmt"
	"testing"

	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/square/go-jose.v2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestNewClient(t *testing.T) {
	config := schema.OpenIDConnectClientConfiguration{}
	client := NewClient(config)
	assert.Equal(t, "", client.GetID())
	assert.Equal(t, "", client.GetDescription())
	assert.Len(t, client.GetResponseModes(), 0)
	assert.Len(t, client.GetResponseTypes(), 1)
	assert.Equal(t, "", client.GetSectorIdentifier())

	bclient, ok := client.(*BaseClient)
	require.True(t, ok)
	assert.Equal(t, "", bclient.UserinfoSigningAlg)
	assert.Equal(t, SigningAlgNone, client.GetUserinfoSigningAlg())

	_, ok = client.(*FullClient)
	assert.False(t, ok)

	config = schema.OpenIDConnectClientConfiguration{
		ID:            myclient,
		Description:   myclientdesc,
		Policy:        twofactor,
		Secret:        MustDecodeSecret(badsecret),
		RedirectURIs:  []string{examplecom},
		Scopes:        schema.DefaultOpenIDConnectClientConfiguration.Scopes,
		ResponseTypes: schema.DefaultOpenIDConnectClientConfiguration.ResponseTypes,
		GrantTypes:    schema.DefaultOpenIDConnectClientConfiguration.GrantTypes,
		ResponseModes: schema.DefaultOpenIDConnectClientConfiguration.ResponseModes,
	}

	client = NewClient(config)
	assert.Equal(t, myclient, client.GetID())
	require.Len(t, client.GetResponseModes(), 1)
	assert.Equal(t, fosite.ResponseModeFormPost, client.GetResponseModes()[0])
	assert.Equal(t, authorization.TwoFactor, client.GetAuthorizationPolicy())

	config = schema.OpenIDConnectClientConfiguration{
		TokenEndpointAuthMethod: ClientAuthMethodClientSecretBasic,
	}

	client = NewClient(config)

	fclient, ok := client.(*FullClient)

	var niljwks *jose.JSONWebKeySet

	require.True(t, ok)
	assert.Equal(t, "", fclient.UserinfoSigningAlg)
	assert.Equal(t, ClientAuthMethodClientSecretBasic, fclient.TokenEndpointAuthMethod)
	assert.Equal(t, ClientAuthMethodClientSecretBasic, fclient.GetTokenEndpointAuthMethod())
	assert.Equal(t, SigningAlgNone, client.GetUserinfoSigningAlg())
	assert.Equal(t, "", fclient.TokenEndpointAuthSigningAlgorithm)
	assert.Equal(t, SigningAlgRSAUsingSHA256, fclient.GetTokenEndpointAuthSigningAlgorithm())
	assert.Equal(t, "", fclient.RequestObjectSigningAlgorithm)
	assert.Equal(t, "", fclient.GetRequestObjectSigningAlgorithm())
	assert.Equal(t, "", fclient.JSONWebKeysURI)
	assert.Equal(t, "", fclient.GetJSONWebKeysURI())
	assert.Equal(t, niljwks, fclient.JSONWebKeys)
	assert.Equal(t, niljwks, fclient.GetJSONWebKeys())
	assert.Equal(t, []string(nil), fclient.RequestURIs)
	assert.Equal(t, []string(nil), fclient.GetRequestURIs())
}

func TestBaseClient_ValidatePARPolicy(t *testing.T) {
	testCases := []struct {
		name     string
		client   *BaseClient
		have     *fosite.Request
		expected string
	}{
		{
			"ShouldNotEnforcePAR",
			&BaseClient{
				EnforcePAR: false,
			},
			&fosite.Request{},
			"",
		},
		{
			"ShouldEnforcePARAndErrorWithoutCorrectRequestURI",
			&BaseClient{
				EnforcePAR: true,
			},
			&fosite.Request{
				Form: map[string][]string{
					FormParameterRequestURI: {"https://google.com"},
				},
			},
			"invalid_request",
		},
		{
			"ShouldEnforcePARAndErrorWithEmptyRequestURI",
			&BaseClient{
				EnforcePAR: true,
			},
			&fosite.Request{
				Form: map[string][]string{
					FormParameterRequestURI: {""},
				},
			},
			"invalid_request",
		},
		{
			"ShouldEnforcePARAndNotErrorWithCorrectRequestURI",
			&BaseClient{
				EnforcePAR: true,
			},
			&fosite.Request{
				Form: map[string][]string{
					FormParameterRequestURI: {urnPARPrefix + "abc"},
				},
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.client.ValidatePARPolicy(tc.have, urnPARPrefix)

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
	c := &FullClient{BaseClient: &BaseClient{}}

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
	c := &FullClient{BaseClient: &BaseClient{}}

	consentRequestBody := c.GetConsentResponseBody(nil)
	assert.Equal(t, "", consentRequestBody.ClientID)
	assert.Equal(t, "", consentRequestBody.ClientDescription)
	assert.Equal(t, []string(nil), consentRequestBody.Scopes)
	assert.Equal(t, []string(nil), consentRequestBody.Audience)

	c.ID = myclient
	c.Description = myclientdesc

	consent := &model.OAuth2ConsentSession{
		RequestedAudience: []string{examplecom},
		RequestedScopes:   []string{ScopeOpenID, ScopeGroups},
	}

	expectedScopes := []string{ScopeOpenID, ScopeGroups}
	expectedAudiences := []string{examplecom}

	consentRequestBody = c.GetConsentResponseBody(consent)
	assert.Equal(t, myclient, consentRequestBody.ClientID)
	assert.Equal(t, myclientdesc, consentRequestBody.ClientDescription)
	assert.Equal(t, expectedScopes, consentRequestBody.Scopes)
	assert.Equal(t, expectedAudiences, consentRequestBody.Audience)
}

func TestClient_GetAudience(t *testing.T) {
	c := &FullClient{BaseClient: &BaseClient{}}

	audience := c.GetAudience()
	assert.Len(t, audience, 0)

	c.Audience = []string{examplecom}

	audience = c.GetAudience()
	require.Len(t, audience, 1)
	assert.Equal(t, examplecom, audience[0])
}

func TestClient_GetScopes(t *testing.T) {
	c := &FullClient{BaseClient: &BaseClient{}}

	scopes := c.GetScopes()
	assert.Len(t, scopes, 0)

	c.Scopes = []string{ScopeOpenID}

	scopes = c.GetScopes()
	require.Len(t, scopes, 1)
	assert.Equal(t, ScopeOpenID, scopes[0])
}

func TestClient_GetGrantTypes(t *testing.T) {
	c := &FullClient{BaseClient: &BaseClient{}}

	grantTypes := c.GetGrantTypes()
	require.Len(t, grantTypes, 1)
	assert.Equal(t, GrantTypeAuthorizationCode, grantTypes[0])

	c.GrantTypes = []string{"device_code"}

	grantTypes = c.GetGrantTypes()
	require.Len(t, grantTypes, 1)
	assert.Equal(t, "device_code", grantTypes[0])
}

func TestClient_Hashing(t *testing.T) {
	c := &FullClient{BaseClient: &BaseClient{}}

	hashedSecret := c.GetHashedSecret()
	assert.Equal(t, []byte(nil), hashedSecret)

	c.Secret = MustDecodeSecret(badsecret)

	assert.True(t, c.Secret.MatchBytes([]byte("a_bad_secret")))
}

func TestClient_GetHashedSecret(t *testing.T) {
	c := &FullClient{BaseClient: &BaseClient{}}

	hashedSecret := c.GetHashedSecret()
	assert.Equal(t, []byte(nil), hashedSecret)

	c.Secret = MustDecodeSecret(badsecret)

	hashedSecret = c.GetHashedSecret()
	assert.Equal(t, []byte(badsecret), hashedSecret)
}

func TestClient_GetID(t *testing.T) {
	c := &FullClient{BaseClient: &BaseClient{}}

	id := c.GetID()
	assert.Equal(t, "", id)

	c.ID = myclient

	id = c.GetID()
	assert.Equal(t, myclient, id)
}

func TestClient_GetRedirectURIs(t *testing.T) {
	c := &FullClient{BaseClient: &BaseClient{}}

	redirectURIs := c.GetRedirectURIs()
	require.Len(t, redirectURIs, 0)

	c.RedirectURIs = []string{examplecom}

	redirectURIs = c.GetRedirectURIs()
	require.Len(t, redirectURIs, 1)
	assert.Equal(t, examplecom, redirectURIs[0])
}

func TestClient_GetResponseModes(t *testing.T) {
	c := &FullClient{BaseClient: &BaseClient{}}

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
	c := &FullClient{BaseClient: &BaseClient{}}

	responseTypes := c.GetResponseTypes()
	require.Len(t, responseTypes, 1)
	assert.Equal(t, ResponseTypeAuthorizationCodeFlow, responseTypes[0])

	c.ResponseTypes = []string{ResponseTypeAuthorizationCodeFlow, ResponseTypeImplicitFlowIDToken}

	responseTypes = c.GetResponseTypes()
	require.Len(t, responseTypes, 2)
	assert.Equal(t, ResponseTypeAuthorizationCodeFlow, responseTypes[0])
	assert.Equal(t, ResponseTypeImplicitFlowIDToken, responseTypes[1])
}

func TestNewClientPKCE(t *testing.T) {
	testCases := []struct {
		name                               string
		have                               schema.OpenIDConnectClientConfiguration
		expectedEnforcePKCE                bool
		expectedEnforcePKCEChallengeMethod bool
		expected                           string
		r                                  *fosite.Request
		err                                string
		desc                               string
	}{
		{
			"ShouldNotEnforcePKCEAndNotErrorOnNonPKCERequest",
			schema.OpenIDConnectClientConfiguration{},
			false,
			false,
			"",
			&fosite.Request{},
			"",
			"",
		},
		{
			"ShouldEnforcePKCEAndErrorOnNonPKCERequest",
			schema.OpenIDConnectClientConfiguration{EnforcePKCE: true},
			true,
			false,
			"",
			&fosite.Request{},
			"invalid_request",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for this client.",
		},
		{
			"ShouldEnforcePKCEAndNotErrorOnPKCERequest",
			schema.OpenIDConnectClientConfiguration{EnforcePKCE: true},
			true,
			false,
			"",
			&fosite.Request{Form: map[string][]string{"code_challenge": {"abc"}}},
			"",
			"",
		},
		{"ShouldEnforcePKCEFromChallengeMethodAndErrorOnNonPKCERequest",
			schema.OpenIDConnectClientConfiguration{PKCEChallengeMethod: "S256"},
			true,
			true,
			"S256",
			&fosite.Request{},
			"invalid_request",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Clients must include a code_challenge when performing the authorize code flow, but it is missing. The server is configured in a way that enforces PKCE for this client.",
		},
		{"ShouldEnforcePKCEFromChallengeMethodAndErrorOnInvalidChallengeMethod",
			schema.OpenIDConnectClientConfiguration{PKCEChallengeMethod: "S256"},
			true,
			true,
			"S256",
			&fosite.Request{Form: map[string][]string{"code_challenge": {"abc"}}},
			"invalid_request",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Client must use code_challenge_method=S256,  is not allowed. The server is configured in a way that enforces PKCE S256 as challenge method for this client.",
		},
		{"ShouldEnforcePKCEFromChallengeMethodAndNotErrorOnValidRequest",
			schema.OpenIDConnectClientConfiguration{PKCEChallengeMethod: "S256"},
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
			client := NewClient(tc.have)

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
		have     schema.OpenIDConnectClientConfiguration
		expected bool
		r        *fosite.Request
		err      string
		desc     string
	}{
		{
			"ShouldNotEnforcEPARAndNotErrorOnNonPARRequest",
			schema.OpenIDConnectClientConfiguration{},
			false,
			&fosite.Request{},
			"",
			"",
		},
		{
			"ShouldEnforcePARAndErrorOnNonPARRequest",
			schema.OpenIDConnectClientConfiguration{EnforcePAR: true},
			true,
			&fosite.Request{},
			"invalid_request",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Pushed Authorization Requests are enforced for this client but no such request was sent. The request_uri parameter was empty.",
		},
		{
			"ShouldEnforcePARAndErrorOnNonPARRequest",
			schema.OpenIDConnectClientConfiguration{EnforcePAR: true},
			true,
			&fosite.Request{Form: map[string][]string{FormParameterRequestURI: {"https://example.com"}}},
			"invalid_request",
			"The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Pushed Authorization Requests are enforced for this client but no such request was sent. The request_uri parameter 'https://example.com' is malformed."},
		{
			"ShouldEnforcePARAndNotErrorOnPARRequest",
			schema.OpenIDConnectClientConfiguration{EnforcePAR: true},
			true,
			&fosite.Request{Form: map[string][]string{FormParameterRequestURI: {fmt.Sprintf("%sabc", urnPARPrefix)}}},
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient(tc.have)

			assert.Equal(t, tc.expected, client.GetPAREnforcement())

			if tc.r != nil {
				err := client.ValidatePARPolicy(tc.r, urnPARPrefix)

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

func TestNewClientResponseModes(t *testing.T) {
	testCases := []struct {
		name     string
		have     schema.OpenIDConnectClientConfiguration
		expected []fosite.ResponseModeType
		r        *fosite.AuthorizeRequest
		err      string
		desc     string
	}{
		{
			"ShouldEnforceResponseModePolicyAndAllowDefaultModeQuery",
			schema.OpenIDConnectClientConfiguration{ResponseModes: []string{ResponseModeQuery}},
			[]fosite.ResponseModeType{fosite.ResponseModeQuery},
			&fosite.AuthorizeRequest{DefaultResponseMode: fosite.ResponseModeQuery, ResponseMode: fosite.ResponseModeDefault, Request: fosite.Request{Form: map[string][]string{FormParameterResponseMode: nil}}},
			"",
			"",
		},
		{
			"ShouldEnforceResponseModePolicyAndFailOnDefaultMode",
			schema.OpenIDConnectClientConfiguration{ResponseModes: []string{ResponseModeFormPost}},
			[]fosite.ResponseModeType{fosite.ResponseModeFormPost},
			&fosite.AuthorizeRequest{DefaultResponseMode: fosite.ResponseModeQuery, ResponseMode: fosite.ResponseModeDefault, Request: fosite.Request{Form: map[string][]string{FormParameterResponseMode: nil}}},
			"unsupported_response_mode",
			"The authorization server does not support obtaining a response using this response mode. The request omitted the response_mode making the default response_mode 'query' based on the other authorization request parameters but registered OAuth 2.0 client doesn't support this response_mode",
		},
		{
			"ShouldNotEnforceConfiguredResponseMode",
			schema.OpenIDConnectClientConfiguration{ResponseModes: []string{ResponseModeFormPost}},
			[]fosite.ResponseModeType{fosite.ResponseModeFormPost},
			&fosite.AuthorizeRequest{DefaultResponseMode: fosite.ResponseModeQuery, ResponseMode: fosite.ResponseModeQuery, Request: fosite.Request{Form: map[string][]string{FormParameterResponseMode: {ResponseModeQuery}}}},
			"",
			"",
		},
		{
			"ShouldNotEnforceUnconfiguredResponseMode",
			schema.OpenIDConnectClientConfiguration{ResponseModes: []string{}},
			[]fosite.ResponseModeType{},
			&fosite.AuthorizeRequest{DefaultResponseMode: fosite.ResponseModeQuery, ResponseMode: fosite.ResponseModeDefault, Request: fosite.Request{Form: map[string][]string{FormParameterResponseMode: {ResponseModeQuery}}}},
			"",
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := NewClient(tc.have)

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
	c := &FullClient{BaseClient: &BaseClient{}}

	assert.False(t, c.IsPublic())

	c.Public = true
	assert.True(t, c.IsPublic())
}
