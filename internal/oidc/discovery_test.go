package oidc_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestNewOpenIDConnectWellKnownConfiguration(t *testing.T) {
	testCases := []struct {
		desc               string
		pkcePlainChallenge bool
		enforcePAR         bool
		clients            map[string]oidc.Client
		discovery          schema.OpenIDConnectDiscovery

		expectCodeChallengeMethodsSupported, expectSubjectTypesSupported  []string
		expectedIDTokenSigAlgsSupported, expectedUserInfoSigAlgsSupported []string

		expectedRequestObjectSigAlgsSupported, expectedRevocationSigAlgsSupported, expectedTokenAuthSigAlgsSupported []string
	}{
		{
			desc:                                  "ShouldHaveChallengeMethodsS256ANDSubjectTypesSupportedPublic",
			pkcePlainChallenge:                    false,
			clients:                               map[string]oidc.Client{"a": &oidc.BaseClient{}},
			expectCodeChallengeMethodsSupported:   []string{oidc.PKCEChallengeMethodSHA256},
			expectSubjectTypesSupported:           []string{oidc.SubjectTypePublic, oidc.SubjectTypePairwise},
			expectedIDTokenSigAlgsSupported:       []string{oidc.SigningAlgRSAUsingSHA256},
			expectedUserInfoSigAlgsSupported:      []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRequestObjectSigAlgsSupported: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRevocationSigAlgsSupported:    []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
			expectedTokenAuthSigAlgsSupported:     []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
		},
		{
			desc:               "ShouldIncludeDiscoveryInfo",
			pkcePlainChallenge: false,
			clients:            map[string]oidc.Client{"a": &oidc.BaseClient{}},
			discovery: schema.OpenIDConnectDiscovery{
				ResponseObjectSigningAlgs: []string{oidc.SigningAlgECDSAUsingP521AndSHA512},
				RequestObjectSigningAlgs:  []string{oidc.SigningAlgECDSAUsingP256AndSHA256},
			},
			expectCodeChallengeMethodsSupported:   []string{oidc.PKCEChallengeMethodSHA256},
			expectSubjectTypesSupported:           []string{oidc.SubjectTypePublic, oidc.SubjectTypePairwise},
			expectedIDTokenSigAlgsSupported:       []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgECDSAUsingP521AndSHA512},
			expectedUserInfoSigAlgsSupported:      []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgECDSAUsingP521AndSHA512, oidc.SigningAlgNone},
			expectedRequestObjectSigAlgsSupported: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgECDSAUsingP256AndSHA256, oidc.SigningAlgNone},
			expectedRevocationSigAlgsSupported:    []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512, oidc.SigningAlgECDSAUsingP256AndSHA256},
			expectedTokenAuthSigAlgsSupported:     []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512, oidc.SigningAlgECDSAUsingP256AndSHA256},
		},
		{
			desc:                                  "ShouldHaveChallengeMethodsS256PlainANDSubjectTypesSupportedPublic",
			pkcePlainChallenge:                    true,
			clients:                               map[string]oidc.Client{"a": &oidc.BaseClient{}},
			expectCodeChallengeMethodsSupported:   []string{oidc.PKCEChallengeMethodSHA256, oidc.PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:           []string{oidc.SubjectTypePublic, oidc.SubjectTypePairwise},
			expectedIDTokenSigAlgsSupported:       []string{oidc.SigningAlgRSAUsingSHA256},
			expectedUserInfoSigAlgsSupported:      []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRequestObjectSigAlgsSupported: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRevocationSigAlgsSupported:    []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
			expectedTokenAuthSigAlgsSupported:     []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
		},
		{
			desc:                                  "ShouldHaveChallengeMethodsS256ANDSubjectTypesSupportedPublicPairwise",
			pkcePlainChallenge:                    false,
			clients:                               map[string]oidc.Client{"a": &oidc.BaseClient{SectorIdentifier: "yes"}},
			expectCodeChallengeMethodsSupported:   []string{oidc.PKCEChallengeMethodSHA256},
			expectSubjectTypesSupported:           []string{oidc.SubjectTypePublic, oidc.SubjectTypePairwise},
			expectedIDTokenSigAlgsSupported:       []string{oidc.SigningAlgRSAUsingSHA256},
			expectedUserInfoSigAlgsSupported:      []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRequestObjectSigAlgsSupported: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRevocationSigAlgsSupported:    []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
			expectedTokenAuthSigAlgsSupported:     []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
		},
		{
			desc:                                  "ShouldHaveChallengeMethodsS256PlainANDSubjectTypesSupportedPublicPairwise",
			pkcePlainChallenge:                    true,
			clients:                               map[string]oidc.Client{"a": &oidc.BaseClient{SectorIdentifier: "yes"}},
			expectCodeChallengeMethodsSupported:   []string{oidc.PKCEChallengeMethodSHA256, oidc.PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:           []string{oidc.SubjectTypePublic, oidc.SubjectTypePairwise},
			expectedIDTokenSigAlgsSupported:       []string{oidc.SigningAlgRSAUsingSHA256},
			expectedUserInfoSigAlgsSupported:      []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRequestObjectSigAlgsSupported: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRevocationSigAlgsSupported:    []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
			expectedTokenAuthSigAlgsSupported:     []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
		},
		{
			desc:                                  "ShouldHaveTokenAuthMethodsNone",
			pkcePlainChallenge:                    true,
			clients:                               map[string]oidc.Client{"a": &oidc.BaseClient{SectorIdentifier: "yes"}},
			expectCodeChallengeMethodsSupported:   []string{oidc.PKCEChallengeMethodSHA256, oidc.PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:           []string{oidc.SubjectTypePublic, oidc.SubjectTypePairwise},
			expectedIDTokenSigAlgsSupported:       []string{oidc.SigningAlgRSAUsingSHA256},
			expectedUserInfoSigAlgsSupported:      []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRequestObjectSigAlgsSupported: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRevocationSigAlgsSupported:    []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
			expectedTokenAuthSigAlgsSupported:     []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
		},
		{
			desc:               "ShouldHaveTokenAuthMethodsNone",
			pkcePlainChallenge: true,
			clients: map[string]oidc.Client{
				"a": &oidc.BaseClient{SectorIdentifier: "yes"},
				"b": &oidc.BaseClient{SectorIdentifier: "yes"},
			},
			expectCodeChallengeMethodsSupported:   []string{oidc.PKCEChallengeMethodSHA256, oidc.PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:           []string{oidc.SubjectTypePublic, oidc.SubjectTypePairwise},
			expectedIDTokenSigAlgsSupported:       []string{oidc.SigningAlgRSAUsingSHA256},
			expectedUserInfoSigAlgsSupported:      []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRequestObjectSigAlgsSupported: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRevocationSigAlgsSupported:    []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
			expectedTokenAuthSigAlgsSupported:     []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
		},
		{
			desc:               "ShouldHaveTokenAuthMethodsNone",
			pkcePlainChallenge: true,
			clients: map[string]oidc.Client{
				"a": &oidc.BaseClient{SectorIdentifier: "yes"},
				"b": &oidc.BaseClient{SectorIdentifier: "yes"},
			},
			expectCodeChallengeMethodsSupported:   []string{oidc.PKCEChallengeMethodSHA256, oidc.PKCEChallengeMethodPlain},
			expectSubjectTypesSupported:           []string{oidc.SubjectTypePublic, oidc.SubjectTypePairwise},
			expectedIDTokenSigAlgsSupported:       []string{oidc.SigningAlgRSAUsingSHA256},
			expectedUserInfoSigAlgsSupported:      []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRequestObjectSigAlgsSupported: []string{oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone},
			expectedRevocationSigAlgsSupported:    []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
			expectedTokenAuthSigAlgsSupported:     []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			c := schema.OpenIDConnectConfiguration{
				EnablePKCEPlainChallenge: tc.pkcePlainChallenge,
				PAR: schema.OpenIDConnectPARConfiguration{
					Enforce: tc.enforcePAR,
				},
				Discovery: tc.discovery,
			}

			actual := oidc.NewOpenIDConnectWellKnownConfiguration(&c)
			for _, codeChallengeMethod := range tc.expectCodeChallengeMethodsSupported {
				assert.Contains(t, actual.CodeChallengeMethodsSupported, codeChallengeMethod)
			}

			for _, subjectType := range tc.expectSubjectTypesSupported {
				assert.Contains(t, actual.SubjectTypesSupported, subjectType)
			}

			for _, codeChallengeMethod := range actual.CodeChallengeMethodsSupported {
				assert.Contains(t, tc.expectCodeChallengeMethodsSupported, codeChallengeMethod)
			}

			for _, subjectType := range actual.SubjectTypesSupported {
				assert.Contains(t, tc.expectSubjectTypesSupported, subjectType)
			}

			assert.Equal(t, tc.expectedUserInfoSigAlgsSupported, actual.UserinfoSigningAlgValuesSupported)
			assert.Equal(t, tc.expectedIDTokenSigAlgsSupported, actual.IDTokenSigningAlgValuesSupported)
			assert.Equal(t, tc.expectedRequestObjectSigAlgsSupported, actual.RequestObjectSigningAlgValuesSupported)
			assert.Equal(t, tc.expectedRevocationSigAlgsSupported, actual.RevocationEndpointAuthSigningAlgValuesSupported)
			assert.Equal(t, tc.expectedTokenAuthSigAlgsSupported, actual.TokenEndpointAuthSigningAlgValuesSupported)
		})
	}
}

func TestNewOpenIDConnectProviderDiscovery(t *testing.T) {
	provider := oidc.NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain:   schema.X509CertificateChain{},
		IssuerPrivateKey:         keyRSA2048,
		HMACSecret:               "asbdhaaskmdlkamdklasmdlkams",
		EnablePKCEPlainChallenge: true,
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "a-client",
				Secret: tOpenIDConnectPlainTextClientSecret,
				Policy: onefactor,
				RedirectURIs: []string{
					"https://google.com",
				},
			},
		},
	}, nil, nil)

	a := provider.GetOpenIDConnectWellKnownConfiguration("https://auth.example.com")

	data, err := json.Marshal(&a)
	assert.NoError(t, err)

	b := oidc.OpenIDConnectWellKnownConfiguration{}

	assert.NoError(t, json.Unmarshal(data, &b))

	assert.Equal(t, a, b)

	y := provider.GetOAuth2WellKnownConfiguration("https://auth.example.com")

	data, err = json.Marshal(&y)
	assert.NoError(t, err)

	z := oidc.OAuth2WellKnownConfiguration{}

	assert.NoError(t, json.Unmarshal(data, &z))

	assert.Equal(t, y, z)
}

func TestNewOpenIDConnectProvider_GetOpenIDConnectWellKnownConfiguration(t *testing.T) {
	provider := oidc.NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       keyRSA2048,
		HMACSecret:             "asbdhaaskmdlkamdklasmdlkams",
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "a-client",
				Secret: tOpenIDConnectPlainTextClientSecret,
				Policy: onefactor,
				RedirectURIs: []string{
					"https://google.com",
				},
			},
		},
	}, nil, nil)

	require.NotNil(t, provider)

	disco := provider.GetOpenIDConnectWellKnownConfiguration(examplecom)

	assert.Equal(t, examplecom, disco.Issuer)
	assert.Equal(t, "https://example.com/jwks.json", disco.JWKSURI)
	assert.Equal(t, "https://example.com/api/oidc/authorization", disco.AuthorizationEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/token", disco.TokenEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/userinfo", disco.UserinfoEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/introspection", disco.IntrospectionEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/revocation", disco.RevocationEndpoint)
	assert.Equal(t, "", disco.RegistrationEndpoint)

	assert.Len(t, disco.CodeChallengeMethodsSupported, 1)
	assert.Contains(t, disco.CodeChallengeMethodsSupported, oidc.PKCEChallengeMethodSHA256)

	assert.Len(t, disco.ScopesSupported, 5)
	assert.Contains(t, disco.ScopesSupported, oidc.ScopeOpenID)
	assert.Contains(t, disco.ScopesSupported, oidc.ScopeOfflineAccess)
	assert.Contains(t, disco.ScopesSupported, oidc.ScopeProfile)
	assert.Contains(t, disco.ScopesSupported, oidc.ScopeGroups)
	assert.Contains(t, disco.ScopesSupported, oidc.ScopeEmail)

	assert.Len(t, disco.ResponseModesSupported, 3)
	assert.Contains(t, disco.ResponseModesSupported, oidc.ResponseModeFormPost)
	assert.Contains(t, disco.ResponseModesSupported, oidc.ResponseModeQuery)
	assert.Contains(t, disco.ResponseModesSupported, oidc.ResponseModeFragment)

	assert.Len(t, disco.SubjectTypesSupported, 2)
	assert.Contains(t, disco.SubjectTypesSupported, oidc.SubjectTypePublic)
	assert.Contains(t, disco.SubjectTypesSupported, oidc.SubjectTypePairwise)

	assert.Len(t, disco.ResponseTypesSupported, 7)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeAuthorizationCodeFlow)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeImplicitFlowIDToken)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeImplicitFlowToken)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeImplicitFlowBoth)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeHybridFlowIDToken)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeHybridFlowToken)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeHybridFlowBoth)

	assert.Len(t, disco.TokenEndpointAuthMethodsSupported, 5)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, oidc.ClientAuthMethodClientSecretBasic)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, oidc.ClientAuthMethodClientSecretPost)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, oidc.ClientAuthMethodClientSecretJWT)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, oidc.ClientAuthMethodPrivateKeyJWT)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, oidc.ClientAuthMethodNone)

	assert.Len(t, disco.RevocationEndpointAuthMethodsSupported, 5)
	assert.Contains(t, disco.RevocationEndpointAuthMethodsSupported, oidc.ClientAuthMethodClientSecretBasic)
	assert.Contains(t, disco.RevocationEndpointAuthMethodsSupported, oidc.ClientAuthMethodClientSecretPost)
	assert.Contains(t, disco.RevocationEndpointAuthMethodsSupported, oidc.ClientAuthMethodClientSecretJWT)
	assert.Contains(t, disco.RevocationEndpointAuthMethodsSupported, oidc.ClientAuthMethodPrivateKeyJWT)
	assert.Contains(t, disco.RevocationEndpointAuthMethodsSupported, oidc.ClientAuthMethodNone)

	assert.Len(t, disco.IntrospectionEndpointAuthMethodsSupported, 2)
	assert.Contains(t, disco.IntrospectionEndpointAuthMethodsSupported, oidc.ClientAuthMethodClientSecretBasic)
	assert.Contains(t, disco.IntrospectionEndpointAuthMethodsSupported, oidc.ClientAuthMethodNone)

	assert.Len(t, disco.GrantTypesSupported, 3)
	assert.Contains(t, disco.GrantTypesSupported, oidc.GrantTypeAuthorizationCode)
	assert.Contains(t, disco.GrantTypesSupported, oidc.GrantTypeRefreshToken)
	assert.Contains(t, disco.GrantTypesSupported, oidc.GrantTypeImplicit)

	assert.Len(t, disco.RevocationEndpointAuthSigningAlgValuesSupported, 3)
	assert.Equal(t, disco.RevocationEndpointAuthSigningAlgValuesSupported[0], oidc.SigningAlgHMACUsingSHA256)
	assert.Equal(t, disco.RevocationEndpointAuthSigningAlgValuesSupported[1], oidc.SigningAlgHMACUsingSHA384)
	assert.Equal(t, disco.RevocationEndpointAuthSigningAlgValuesSupported[2], oidc.SigningAlgHMACUsingSHA512)

	assert.Len(t, disco.TokenEndpointAuthSigningAlgValuesSupported, 3)
	assert.Equal(t, disco.TokenEndpointAuthSigningAlgValuesSupported[0], oidc.SigningAlgHMACUsingSHA256)
	assert.Equal(t, disco.TokenEndpointAuthSigningAlgValuesSupported[1], oidc.SigningAlgHMACUsingSHA384)
	assert.Equal(t, disco.TokenEndpointAuthSigningAlgValuesSupported[2], oidc.SigningAlgHMACUsingSHA512)

	assert.Len(t, disco.IDTokenSigningAlgValuesSupported, 1)
	assert.Contains(t, disco.IDTokenSigningAlgValuesSupported, oidc.SigningAlgRSAUsingSHA256)

	assert.Len(t, disco.UserinfoSigningAlgValuesSupported, 2)
	assert.Equal(t, disco.UserinfoSigningAlgValuesSupported[0], oidc.SigningAlgRSAUsingSHA256)
	assert.Equal(t, disco.UserinfoSigningAlgValuesSupported[1], oidc.SigningAlgNone)

	require.Len(t, disco.RequestObjectSigningAlgValuesSupported, 2)
	assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, disco.RequestObjectSigningAlgValuesSupported[0])
	assert.Equal(t, oidc.SigningAlgNone, disco.RequestObjectSigningAlgValuesSupported[1])

	assert.Len(t, disco.ClaimsSupported, 18)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimAuthenticationMethodsReference)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimAudience)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimAuthorizedParty)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimClientIdentifier)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimExpirationTime)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimIssuedAt)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimIssuer)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimJWTID)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimRequestedAt)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimSubject)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimAuthenticationTime)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimNonce)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimPreferredEmail)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimEmailVerified)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimEmailAlts)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimGroups)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimPreferredUsername)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimFullName)

	assert.Len(t, disco.PromptValuesSupported, 2)
	assert.Contains(t, disco.PromptValuesSupported, oidc.PromptConsent)
	assert.Contains(t, disco.PromptValuesSupported, oidc.PromptNone)
}

func TestNewOpenIDConnectProvider_GetOAuth2WellKnownConfiguration(t *testing.T) {
	provider := oidc.NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       keyRSA2048,
		HMACSecret:             "asbdhaaskmdlkamdklasmdlkams",
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "a-client",
				Secret: tOpenIDConnectPlainTextClientSecret,
				Policy: onefactor,
				RedirectURIs: []string{
					"https://google.com",
				},
			},
		},
	}, nil, nil)

	require.NotNil(t, provider)

	disco := provider.GetOAuth2WellKnownConfiguration(examplecom)

	assert.Equal(t, examplecom, disco.Issuer)
	assert.Equal(t, "https://example.com/jwks.json", disco.JWKSURI)
	assert.Equal(t, "https://example.com/api/oidc/authorization", disco.AuthorizationEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/token", disco.TokenEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/introspection", disco.IntrospectionEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/revocation", disco.RevocationEndpoint)
	assert.Equal(t, "", disco.RegistrationEndpoint)

	require.Len(t, disco.CodeChallengeMethodsSupported, 1)
	assert.Equal(t, "S256", disco.CodeChallengeMethodsSupported[0])

	assert.Len(t, disco.ScopesSupported, 5)
	assert.Contains(t, disco.ScopesSupported, oidc.ScopeOpenID)
	assert.Contains(t, disco.ScopesSupported, oidc.ScopeOfflineAccess)
	assert.Contains(t, disco.ScopesSupported, oidc.ScopeProfile)
	assert.Contains(t, disco.ScopesSupported, oidc.ScopeGroups)
	assert.Contains(t, disco.ScopesSupported, oidc.ScopeEmail)

	assert.Len(t, disco.ResponseModesSupported, 3)
	assert.Contains(t, disco.ResponseModesSupported, oidc.ResponseModeFormPost)
	assert.Contains(t, disco.ResponseModesSupported, oidc.ResponseModeQuery)
	assert.Contains(t, disco.ResponseModesSupported, oidc.ResponseModeFragment)

	assert.Len(t, disco.SubjectTypesSupported, 2)
	assert.Contains(t, disco.SubjectTypesSupported, oidc.SubjectTypePublic)
	assert.Contains(t, disco.SubjectTypesSupported, oidc.SubjectTypePairwise)

	assert.Len(t, disco.ResponseTypesSupported, 7)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeAuthorizationCodeFlow)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeImplicitFlowIDToken)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeImplicitFlowToken)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeImplicitFlowBoth)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeHybridFlowIDToken)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeHybridFlowToken)
	assert.Contains(t, disco.ResponseTypesSupported, oidc.ResponseTypeHybridFlowBoth)

	assert.Len(t, disco.TokenEndpointAuthMethodsSupported, 5)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, oidc.ClientAuthMethodClientSecretBasic)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, oidc.ClientAuthMethodClientSecretPost)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, oidc.ClientAuthMethodClientSecretJWT)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, oidc.ClientAuthMethodPrivateKeyJWT)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, oidc.ClientAuthMethodNone)

	assert.Len(t, disco.GrantTypesSupported, 3)
	assert.Contains(t, disco.GrantTypesSupported, oidc.GrantTypeAuthorizationCode)
	assert.Contains(t, disco.GrantTypesSupported, oidc.GrantTypeRefreshToken)
	assert.Contains(t, disco.GrantTypesSupported, oidc.GrantTypeImplicit)

	assert.Len(t, disco.ClaimsSupported, 18)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimAuthenticationMethodsReference)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimAudience)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimAuthorizedParty)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimClientIdentifier)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimExpirationTime)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimIssuedAt)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimIssuer)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimJWTID)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimRequestedAt)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimSubject)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimAuthenticationTime)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimNonce)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimPreferredEmail)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimEmailVerified)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimEmailAlts)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimGroups)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimPreferredUsername)
	assert.Contains(t, disco.ClaimsSupported, oidc.ClaimFullName)
}

func TestNewOpenIDConnectProvider_GetOpenIDConnectWellKnownConfigurationWithPlainPKCE(t *testing.T) {
	provider := oidc.NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain:   schema.X509CertificateChain{},
		IssuerPrivateKey:         keyRSA2048,
		HMACSecret:               "asbdhaaskmdlkamdklasmdlkams",
		EnablePKCEPlainChallenge: true,
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "a-client",
				Secret: tOpenIDConnectPlainTextClientSecret,
				Policy: onefactor,
				RedirectURIs: []string{
					"https://google.com",
				},
			},
		},
	}, nil, nil)

	require.NotNil(t, provider)

	disco := provider.GetOpenIDConnectWellKnownConfiguration(examplecom)

	require.Len(t, disco.CodeChallengeMethodsSupported, 2)
	assert.Equal(t, oidc.PKCEChallengeMethodSHA256, disco.CodeChallengeMethodsSupported[0])
	assert.Equal(t, oidc.PKCEChallengeMethodPlain, disco.CodeChallengeMethodsSupported[1])
}

func TestNewOpenIDConnectWellKnownConfiguration_Copy(t *testing.T) {
	config := &oidc.OpenIDConnectWellKnownConfiguration{
		OAuth2WellKnownConfiguration: oidc.OAuth2WellKnownConfiguration{
			CommonDiscoveryOptions: oidc.CommonDiscoveryOptions{
				Issuer:                            "https://example.com",
				JWKSURI:                           "https://example.com/jwks.json",
				AuthorizationEndpoint:             "",
				TokenEndpoint:                     "",
				SubjectTypesSupported:             nil,
				ResponseTypesSupported:            nil,
				GrantTypesSupported:               nil,
				ResponseModesSupported:            nil,
				ScopesSupported:                   nil,
				ClaimsSupported:                   nil,
				UILocalesSupported:                nil,
				TokenEndpointAuthMethodsSupported: nil,
				TokenEndpointAuthSigningAlgValuesSupported: nil,
				ServiceDocumentation:                       "",
				OPPolicyURI:                                "",
				OPTOSURI:                                   "",
				SignedMetadata:                             "",
			},
			OAuth2DiscoveryOptions: oidc.OAuth2DiscoveryOptions{
				IntrospectionEndpoint:                              "",
				RevocationEndpoint:                                 "",
				RegistrationEndpoint:                               "",
				IntrospectionEndpointAuthMethodsSupported:          nil,
				RevocationEndpointAuthMethodsSupported:             nil,
				RevocationEndpointAuthSigningAlgValuesSupported:    nil,
				IntrospectionEndpointAuthSigningAlgValuesSupported: nil,
				CodeChallengeMethodsSupported:                      nil,
			},
			OAuth2DeviceAuthorizationGrantDiscoveryOptions: &oidc.OAuth2DeviceAuthorizationGrantDiscoveryOptions{
				DeviceAuthorizationEndpoint: "",
			},
			OAuth2MutualTLSClientAuthenticationDiscoveryOptions: &oidc.OAuth2MutualTLSClientAuthenticationDiscoveryOptions{
				TLSClientCertificateBoundAccessTokens: false,
				MutualTLSEndpointAliases: oidc.OAuth2MutualTLSClientAuthenticationAliasesDiscoveryOptions{
					AuthorizationEndpoint:              "",
					TokenEndpoint:                      "",
					IntrospectionEndpoint:              "",
					RevocationEndpoint:                 "",
					EndSessionEndpoint:                 "",
					UserinfoEndpoint:                   "",
					BackChannelAuthenticationEndpoint:  "",
					FederationRegistrationEndpoint:     "",
					PushedAuthorizationRequestEndpoint: "",
					RegistrationEndpoint:               "",
				},
			},
			OAuth2IssuerIdentificationDiscoveryOptions: &oidc.OAuth2IssuerIdentificationDiscoveryOptions{
				AuthorizationResponseIssuerParameterSupported: false,
			},
			OAuth2JWTIntrospectionResponseDiscoveryOptions: &oidc.OAuth2JWTIntrospectionResponseDiscoveryOptions{
				IntrospectionSigningAlgValuesSupported:    nil,
				IntrospectionEncryptionAlgValuesSupported: nil,
				IntrospectionEncryptionEncValuesSupported: nil,
			},
			OAuth2JWTSecuredAuthorizationRequestDiscoveryOptions: &oidc.OAuth2JWTSecuredAuthorizationRequestDiscoveryOptions{
				RequireSignedRequestObject: false,
			},
			OAuth2PushedAuthorizationDiscoveryOptions: &oidc.OAuth2PushedAuthorizationDiscoveryOptions{
				PushedAuthorizationRequestEndpoint: "",
				RequirePushedAuthorizationRequests: false,
			},
		},
		OpenIDConnectDiscoveryOptions: oidc.OpenIDConnectDiscoveryOptions{
			UserinfoEndpoint:                          "",
			IDTokenSigningAlgValuesSupported:          nil,
			UserinfoSigningAlgValuesSupported:         nil,
			RequestObjectSigningAlgValuesSupported:    nil,
			IDTokenEncryptionAlgValuesSupported:       nil,
			UserinfoEncryptionAlgValuesSupported:      nil,
			RequestObjectEncryptionAlgValuesSupported: nil,
			IDTokenEncryptionEncValuesSupported:       nil,
			UserinfoEncryptionEncValuesSupported:      nil,
			RequestObjectEncryptionEncValuesSupported: nil,
			ACRValuesSupported:                        nil,
			DisplayValuesSupported:                    nil,
			ClaimTypesSupported:                       nil,
			ClaimLocalesSupported:                     nil,
			RequestParameterSupported:                 false,
			RequestURIParameterSupported:              false,
			RequireRequestURIRegistration:             false,
			ClaimsParameterSupported:                  false,
		},
		OpenIDConnectFrontChannelLogoutDiscoveryOptions: &oidc.OpenIDConnectFrontChannelLogoutDiscoveryOptions{
			FrontChannelLogoutSupported:        false,
			FrontChannelLogoutSessionSupported: false,
		},
		OpenIDConnectBackChannelLogoutDiscoveryOptions: &oidc.OpenIDConnectBackChannelLogoutDiscoveryOptions{
			BackChannelLogoutSupported:        false,
			BackChannelLogoutSessionSupported: false,
		},
		OpenIDConnectSessionManagementDiscoveryOptions: &oidc.OpenIDConnectSessionManagementDiscoveryOptions{
			CheckSessionIFrame: "",
		},
		OpenIDConnectRPInitiatedLogoutDiscoveryOptions: &oidc.OpenIDConnectRPInitiatedLogoutDiscoveryOptions{
			EndSessionEndpoint: "",
		},
		OpenIDConnectPromptCreateDiscoveryOptions: &oidc.OpenIDConnectPromptCreateDiscoveryOptions{
			PromptValuesSupported: nil,
		},
		OpenIDConnectClientInitiatedBackChannelAuthFlowDiscoveryOptions: &oidc.OpenIDConnectClientInitiatedBackChannelAuthFlowDiscoveryOptions{
			BackChannelAuthenticationEndpoint:               "",
			BackChannelTokenDeliveryModesSupported:          nil,
			BackChannelAuthRequestSigningAlgValuesSupported: nil,
			BackChannelUserCodeParameterSupported:           false,
		},
		OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions: &oidc.OpenIDConnectJWTSecuredAuthorizationResponseModeDiscoveryOptions{
			AuthorizationSigningAlgValuesSupported:    nil,
			AuthorizationEncryptionAlgValuesSupported: nil,
			AuthorizationEncryptionEncValuesSupported: nil,
		},
		OpenIDFederationDiscoveryOptions: &oidc.OpenIDFederationDiscoveryOptions{
			FederationRegistrationEndpoint:                 "",
			ClientRegistrationTypesSupported:               nil,
			RequestAuthenticationMethodsSupported:          nil,
			RequestAuthenticationSigningAlgValuesSupproted: nil,
		},
	}

	x := config.Copy()

	assert.Equal(t, config, &x)

	y := config.OAuth2WellKnownConfiguration.Copy()

	assert.Equal(t, config.OAuth2WellKnownConfiguration, y)
}
