package oidc

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestOpenIDConnectProvider_NewOpenIDConnectProvider_NotConfigured(t *testing.T) {
	provider := NewOpenIDConnectProvider(nil, nil, nil)

	assert.Nil(t, provider)
}

func TestNewOpenIDConnectProvider_ShouldEnableOptionalDiscoveryValues(t *testing.T) {
	provider := NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain:   schema.X509CertificateChain{},
		IssuerPrivateKey:         keyRSA2048,
		EnablePKCEPlainChallenge: true,
		HMACSecret:               badhmac,
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:               myclient,
				Secret:           MustDecodeSecret(badsecret),
				SectorIdentifier: url.URL{Host: examplecomsid},
				Policy:           onefactor,
				RedirectURIs: []string{
					examplecom,
				},
			},
		},
	}, nil, nil)

	require.NotNil(t, provider)

	disco := provider.GetOpenIDConnectWellKnownConfiguration(examplecom)

	assert.Len(t, disco.SubjectTypesSupported, 2)
	assert.Contains(t, disco.SubjectTypesSupported, SubjectTypePublic)
	assert.Contains(t, disco.SubjectTypesSupported, SubjectTypePairwise)

	assert.Len(t, disco.CodeChallengeMethodsSupported, 2)
	assert.Contains(t, disco.CodeChallengeMethodsSupported, PKCEChallengeMethodSHA256)
	assert.Contains(t, disco.CodeChallengeMethodsSupported, PKCEChallengeMethodSHA256)
}

func TestOpenIDConnectProvider_NewOpenIDConnectProvider_GoodConfiguration(t *testing.T) {
	provider := NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       keyRSA2048,
		HMACSecret:             badhmac,
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "a-client",
				Secret: MustDecodeSecret("$plaintext$a-client-secret"),
				Policy: onefactor,
				RedirectURIs: []string{
					"https://google.com",
				},
			},
			{
				ID:          "b-client",
				Description: "Normal Description",
				Secret:      MustDecodeSecret("$plaintext$b-client-secret"),
				Policy:      twofactor,
				RedirectURIs: []string{
					"https://google.com",
				},
				Scopes: []string{
					ScopeGroups,
				},
				GrantTypes: []string{
					GrantTypeRefreshToken,
				},
				ResponseTypes: []string{
					"token",
					"code",
				},
			},
		},
	}, nil, nil)

	assert.NotNil(t, provider)
}

func TestOpenIDConnectProvider_NewOpenIDConnectProvider_GetOpenIDConnectWellKnownConfiguration(t *testing.T) {
	provider := NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       keyRSA2048,
		HMACSecret:             "asbdhaaskmdlkamdklasmdlkams",
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "a-client",
				Secret: MustDecodeSecret("$plaintext$a-client-secret"),
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
	assert.Contains(t, disco.CodeChallengeMethodsSupported, PKCEChallengeMethodSHA256)

	assert.Len(t, disco.ScopesSupported, 5)
	assert.Contains(t, disco.ScopesSupported, ScopeOpenID)
	assert.Contains(t, disco.ScopesSupported, ScopeOfflineAccess)
	assert.Contains(t, disco.ScopesSupported, ScopeProfile)
	assert.Contains(t, disco.ScopesSupported, ScopeGroups)
	assert.Contains(t, disco.ScopesSupported, ScopeEmail)

	assert.Len(t, disco.ResponseModesSupported, 3)
	assert.Contains(t, disco.ResponseModesSupported, ResponseModeFormPost)
	assert.Contains(t, disco.ResponseModesSupported, ResponseModeQuery)
	assert.Contains(t, disco.ResponseModesSupported, ResponseModeFragment)

	assert.Len(t, disco.SubjectTypesSupported, 2)
	assert.Contains(t, disco.SubjectTypesSupported, SubjectTypePublic)
	assert.Contains(t, disco.SubjectTypesSupported, SubjectTypePairwise)

	assert.Len(t, disco.ResponseTypesSupported, 7)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeAuthorizationCodeFlow)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeImplicitFlowIDToken)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeImplicitFlowToken)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeImplicitFlowBoth)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeHybridFlowIDToken)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeHybridFlowToken)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeHybridFlowBoth)

	assert.Len(t, disco.TokenEndpointAuthMethodsSupported, 5)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, ClientAuthMethodClientSecretBasic)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, ClientAuthMethodClientSecretPost)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, ClientAuthMethodClientSecretJWT)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, ClientAuthMethodPrivateKeyJWT)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, ClientAuthMethodNone)

	assert.Len(t, disco.RevocationEndpointAuthMethodsSupported, 5)
	assert.Contains(t, disco.RevocationEndpointAuthMethodsSupported, ClientAuthMethodClientSecretBasic)
	assert.Contains(t, disco.RevocationEndpointAuthMethodsSupported, ClientAuthMethodClientSecretPost)
	assert.Contains(t, disco.RevocationEndpointAuthMethodsSupported, ClientAuthMethodClientSecretJWT)
	assert.Contains(t, disco.RevocationEndpointAuthMethodsSupported, ClientAuthMethodPrivateKeyJWT)
	assert.Contains(t, disco.RevocationEndpointAuthMethodsSupported, ClientAuthMethodNone)

	assert.Len(t, disco.IntrospectionEndpointAuthMethodsSupported, 2)
	assert.Contains(t, disco.IntrospectionEndpointAuthMethodsSupported, ClientAuthMethodClientSecretBasic)
	assert.Contains(t, disco.IntrospectionEndpointAuthMethodsSupported, ClientAuthMethodNone)

	assert.Len(t, disco.GrantTypesSupported, 3)
	assert.Contains(t, disco.GrantTypesSupported, GrantTypeAuthorizationCode)
	assert.Contains(t, disco.GrantTypesSupported, GrantTypeRefreshToken)
	assert.Contains(t, disco.GrantTypesSupported, GrantTypeImplicit)

	assert.Len(t, disco.RevocationEndpointAuthSigningAlgValuesSupported, 3)
	assert.Equal(t, disco.RevocationEndpointAuthSigningAlgValuesSupported[0], SigningAlgHMACUsingSHA256)
	assert.Equal(t, disco.RevocationEndpointAuthSigningAlgValuesSupported[1], SigningAlgHMACUsingSHA384)
	assert.Equal(t, disco.RevocationEndpointAuthSigningAlgValuesSupported[2], SigningAlgHMACUsingSHA512)

	assert.Len(t, disco.TokenEndpointAuthSigningAlgValuesSupported, 3)
	assert.Equal(t, disco.TokenEndpointAuthSigningAlgValuesSupported[0], SigningAlgHMACUsingSHA256)
	assert.Equal(t, disco.TokenEndpointAuthSigningAlgValuesSupported[1], SigningAlgHMACUsingSHA384)
	assert.Equal(t, disco.TokenEndpointAuthSigningAlgValuesSupported[2], SigningAlgHMACUsingSHA512)

	assert.Len(t, disco.IDTokenSigningAlgValuesSupported, 1)
	assert.Contains(t, disco.IDTokenSigningAlgValuesSupported, SigningAlgRSAUsingSHA256)

	assert.Len(t, disco.UserinfoSigningAlgValuesSupported, 2)
	assert.Equal(t, disco.UserinfoSigningAlgValuesSupported[0], SigningAlgRSAUsingSHA256)
	assert.Equal(t, disco.UserinfoSigningAlgValuesSupported[1], SigningAlgNone)

	require.Len(t, disco.RequestObjectSigningAlgValuesSupported, 2)
	assert.Equal(t, SigningAlgRSAUsingSHA256, disco.RequestObjectSigningAlgValuesSupported[0])
	assert.Equal(t, SigningAlgNone, disco.RequestObjectSigningAlgValuesSupported[1])

	assert.Len(t, disco.ClaimsSupported, 18)
	assert.Contains(t, disco.ClaimsSupported, ClaimAuthenticationMethodsReference)
	assert.Contains(t, disco.ClaimsSupported, ClaimAudience)
	assert.Contains(t, disco.ClaimsSupported, ClaimAuthorizedParty)
	assert.Contains(t, disco.ClaimsSupported, ClaimClientIdentifier)
	assert.Contains(t, disco.ClaimsSupported, ClaimExpirationTime)
	assert.Contains(t, disco.ClaimsSupported, ClaimIssuedAt)
	assert.Contains(t, disco.ClaimsSupported, ClaimIssuer)
	assert.Contains(t, disco.ClaimsSupported, ClaimJWTID)
	assert.Contains(t, disco.ClaimsSupported, ClaimRequestedAt)
	assert.Contains(t, disco.ClaimsSupported, ClaimSubject)
	assert.Contains(t, disco.ClaimsSupported, ClaimAuthenticationTime)
	assert.Contains(t, disco.ClaimsSupported, ClaimNonce)
	assert.Contains(t, disco.ClaimsSupported, ClaimPreferredEmail)
	assert.Contains(t, disco.ClaimsSupported, ClaimEmailVerified)
	assert.Contains(t, disco.ClaimsSupported, ClaimEmailAlts)
	assert.Contains(t, disco.ClaimsSupported, ClaimGroups)
	assert.Contains(t, disco.ClaimsSupported, ClaimPreferredUsername)
	assert.Contains(t, disco.ClaimsSupported, ClaimFullName)

	assert.Len(t, disco.PromptValuesSupported, 2)
	assert.Contains(t, disco.PromptValuesSupported, PromptConsent)
	assert.Contains(t, disco.PromptValuesSupported, PromptNone)
}

func TestOpenIDConnectProvider_NewOpenIDConnectProvider_GetOAuth2WellKnownConfiguration(t *testing.T) {
	provider := NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       keyRSA2048,
		HMACSecret:             "asbdhaaskmdlkamdklasmdlkams",
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "a-client",
				Secret: MustDecodeSecret("$plaintext$a-client-secret"),
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
	assert.Contains(t, disco.ScopesSupported, ScopeOpenID)
	assert.Contains(t, disco.ScopesSupported, ScopeOfflineAccess)
	assert.Contains(t, disco.ScopesSupported, ScopeProfile)
	assert.Contains(t, disco.ScopesSupported, ScopeGroups)
	assert.Contains(t, disco.ScopesSupported, ScopeEmail)

	assert.Len(t, disco.ResponseModesSupported, 3)
	assert.Contains(t, disco.ResponseModesSupported, ResponseModeFormPost)
	assert.Contains(t, disco.ResponseModesSupported, ResponseModeQuery)
	assert.Contains(t, disco.ResponseModesSupported, ResponseModeFragment)

	assert.Len(t, disco.SubjectTypesSupported, 2)
	assert.Contains(t, disco.SubjectTypesSupported, SubjectTypePublic)
	assert.Contains(t, disco.SubjectTypesSupported, SubjectTypePairwise)

	assert.Len(t, disco.ResponseTypesSupported, 7)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeAuthorizationCodeFlow)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeImplicitFlowIDToken)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeImplicitFlowToken)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeImplicitFlowBoth)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeHybridFlowIDToken)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeHybridFlowToken)
	assert.Contains(t, disco.ResponseTypesSupported, ResponseTypeHybridFlowBoth)

	assert.Len(t, disco.TokenEndpointAuthMethodsSupported, 5)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, ClientAuthMethodClientSecretBasic)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, ClientAuthMethodClientSecretPost)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, ClientAuthMethodClientSecretJWT)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, ClientAuthMethodPrivateKeyJWT)
	assert.Contains(t, disco.TokenEndpointAuthMethodsSupported, ClientAuthMethodNone)

	assert.Len(t, disco.GrantTypesSupported, 3)
	assert.Contains(t, disco.GrantTypesSupported, GrantTypeAuthorizationCode)
	assert.Contains(t, disco.GrantTypesSupported, GrantTypeRefreshToken)
	assert.Contains(t, disco.GrantTypesSupported, GrantTypeImplicit)

	assert.Len(t, disco.ClaimsSupported, 18)
	assert.Contains(t, disco.ClaimsSupported, ClaimAuthenticationMethodsReference)
	assert.Contains(t, disco.ClaimsSupported, ClaimAudience)
	assert.Contains(t, disco.ClaimsSupported, ClaimAuthorizedParty)
	assert.Contains(t, disco.ClaimsSupported, ClaimClientIdentifier)
	assert.Contains(t, disco.ClaimsSupported, ClaimExpirationTime)
	assert.Contains(t, disco.ClaimsSupported, ClaimIssuedAt)
	assert.Contains(t, disco.ClaimsSupported, ClaimIssuer)
	assert.Contains(t, disco.ClaimsSupported, ClaimJWTID)
	assert.Contains(t, disco.ClaimsSupported, ClaimRequestedAt)
	assert.Contains(t, disco.ClaimsSupported, ClaimSubject)
	assert.Contains(t, disco.ClaimsSupported, ClaimAuthenticationTime)
	assert.Contains(t, disco.ClaimsSupported, ClaimNonce)
	assert.Contains(t, disco.ClaimsSupported, ClaimPreferredEmail)
	assert.Contains(t, disco.ClaimsSupported, ClaimEmailVerified)
	assert.Contains(t, disco.ClaimsSupported, ClaimEmailAlts)
	assert.Contains(t, disco.ClaimsSupported, ClaimGroups)
	assert.Contains(t, disco.ClaimsSupported, ClaimPreferredUsername)
	assert.Contains(t, disco.ClaimsSupported, ClaimFullName)
}

func TestOpenIDConnectProvider_NewOpenIDConnectProvider_GetOpenIDConnectWellKnownConfigurationWithPlainPKCE(t *testing.T) {
	provider := NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain:   schema.X509CertificateChain{},
		IssuerPrivateKey:         keyRSA2048,
		HMACSecret:               "asbdhaaskmdlkamdklasmdlkams",
		EnablePKCEPlainChallenge: true,
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "a-client",
				Secret: MustDecodeSecret("$plaintext$a-client-secret"),
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
	assert.Equal(t, PKCEChallengeMethodSHA256, disco.CodeChallengeMethodsSupported[0])
	assert.Equal(t, PKCEChallengeMethodPlain, disco.CodeChallengeMethodsSupported[1])
}

func TestNewOpenIDConnectProviderDiscovery(t *testing.T) {
	provider := NewOpenIDConnectProvider(&schema.OpenIDConnectConfiguration{
		IssuerCertificateChain:   schema.X509CertificateChain{},
		IssuerPrivateKey:         keyRSA2048,
		HMACSecret:               "asbdhaaskmdlkamdklasmdlkams",
		EnablePKCEPlainChallenge: true,
		Clients: []schema.OpenIDConnectClientConfiguration{
			{
				ID:     "a-client",
				Secret: MustDecodeSecret("$plaintext$a-client-secret"),
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

	b := OpenIDConnectWellKnownConfiguration{}

	assert.NoError(t, json.Unmarshal(data, &b))

	assert.Equal(t, a, b)

	y := provider.GetOAuth2WellKnownConfiguration("https://auth.example.com")

	data, err = json.Marshal(&y)
	assert.NoError(t, err)

	z := OAuth2WellKnownConfiguration{}

	assert.NoError(t, json.Unmarshal(data, &z))

	assert.Equal(t, y, z)
}
