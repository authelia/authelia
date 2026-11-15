// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOpenIDConnectProviderMetadata(t *testing.T) {
	provider := newDiscoveryTestProvider(t, nil)

	metadata, err := provider.NewOpenIDConnectProviderMetadata(context.Background())
	require.NoError(t, err)

	assert.Equal(t, examplecom, metadata.Issuer)
	assert.Equal(t, "https://example.com/jwks.json", metadata.JWKSURI)
	assert.Equal(t, "https://example.com/api/oidc/authorization", metadata.AuthorizationEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/token", metadata.TokenEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/userinfo", metadata.UserinfoEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/introspection", metadata.IntrospectionEndpoint)
	assert.Equal(t, "https://example.com/api/oidc/revocation", metadata.RevocationEndpoint)
	assert.Equal(t, "", metadata.RegistrationEndpoint)

	assert.Equal(t, []string{oidc.PKCEChallengeMethodSHA256}, metadata.CodeChallengeMethodsSupported)
	assert.Equal(t, []string{oidc.ScopeAddress, oidc.ScopeEmail, oidc.ScopeGroups, oidc.ScopeOffline, oidc.ScopeOfflineAccess, oidc.ScopeOpenID, oidc.ScopePhone, oidc.ScopeProfile}, metadata.ScopesSupported)
	assert.Equal(t, []string{oidc.ResponseModeFormPost, oidc.ResponseModeFormPostJWT, oidc.ResponseModeFragment, oidc.ResponseModeFragmentJWT, oidc.ResponseModeJWT, oidc.ResponseModeQuery, oidc.ResponseModeQueryJWT}, metadata.ResponseModesSupported)
	assert.Equal(t, []string{oidc.ResponseTypeAuthorizationCodeFlow, oidc.ResponseTypeHybridFlowIDToken, oidc.ResponseTypeHybridFlowBoth, oidc.ResponseTypeHybridFlowToken, oidc.ResponseTypeImplicitFlowIDToken, oidc.ResponseTypeImplicitFlowBoth, oidc.ResponseTypeImplicitFlowToken}, metadata.ResponseTypesSupported)
	assert.Equal(t, []string{oidc.GrantTypeAuthorizationCode, oidc.GrantTypeClientCredentials, oidc.GrantTypeImplicit, oidc.GrantTypeRefreshToken, oidc.GrantTypeDeviceCode}, metadata.GrantTypesSupported)
	assert.Equal(t, []string{oidc.SubjectTypePublic, oidc.SubjectTypePairwise}, metadata.SubjectTypesSupported)
	assert.Equal(t, []string{oidc.ClaimTypeNormal}, metadata.ClaimTypesSupported)
	assert.Equal(t, []string{oidc.PromptConsent, oidc.PromptLogin, oidc.PromptNone, oidc.PromptSelectAccount}, metadata.PromptValuesSupported)

	assert.Equal(t, []string{oidc.ClientAuthMethodClientSecretBasic, oidc.ClientAuthMethodClientSecretPost, oidc.ClientAuthMethodClientSecretJWT, oidc.ClientAuthMethodPrivateKeyJWT, oidc.ClientAuthMethodNone}, metadata.TokenEndpointAuthMethodsSupported)
	assert.Equal(t, []string{oidc.ClientAuthMethodClientSecretBasic, oidc.ClientAuthMethodClientSecretPost, oidc.ClientAuthMethodClientSecretJWT, oidc.ClientAuthMethodPrivateKeyJWT, oidc.ClientAuthMethodNone}, metadata.RevocationEndpointAuthMethodsSupported)
	assert.Equal(t, []string{oidc.ClientAuthMethodClientSecretBasic, oidc.ClientAuthMethodClientSecretPost, oidc.ClientAuthMethodClientSecretJWT, oidc.ClientAuthMethodPrivateKeyJWT}, metadata.IntrospectionEndpointAuthMethodsSupported)

	assert.Equal(t, oidc.SigningAlgValuesSupported(), metadata.TokenEndpointAuthSigningAlgValuesSupported)
	assert.Equal(t, oidc.SigningAlgValuesSupported(), metadata.RevocationEndpointAuthSigningAlgValuesSupported)
	assert.Equal(t, oidc.SigningAlgValuesSupported(), metadata.IntrospectionEndpointAuthSigningAlgValuesSupported)
	assert.Equal(t, oidc.SigningAlgValuesSupported(oidc.SigningAlgNone), metadata.RequestObjectSigningAlgValuesSupported)

	issued := []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512, oidc.SigningAlgRSAUsingSHA256}

	assert.Equal(t, issued, metadata.IDTokenSigningAlgValuesSupported)
	assert.Equal(t, []string{oidc.SigningAlgHMACUsingSHA256, oidc.SigningAlgHMACUsingSHA384, oidc.SigningAlgHMACUsingSHA512, oidc.SigningAlgRSAUsingSHA256, oidc.SigningAlgNone}, metadata.UserinfoSigningAlgValuesSupported)

	require.NotNil(t, metadata.RFC9701JWTIntrospectionResponseMetadata)
	assert.Equal(t, issued, metadata.IntrospectionSigningAlgValuesSupported)

	require.NotNil(t, metadata.OpenIDConnectJARMMetadata)
	assert.Equal(t, issued, metadata.AuthorizationSigningAlgValuesSupported)

	assert.Equal(t, oidc.EncryptionAlgValuesSupported(), metadata.IDTokenEncryptionAlgValuesSupported)
	assert.Equal(t, oidc.EncryptionEncValuesSupported(), metadata.IDTokenEncryptionEncValuesSupported)

	assert.NotContains(t, metadata.ClaimsSupported, oidc.ClaimAuthenticationContextClassReference)
	assert.NotContains(t, metadata.ClaimsSupported, oidc.ClaimSessionID)
	assert.Contains(t, metadata.ClaimsSupported, oidc.ClaimEmailAlts)
	assert.Contains(t, metadata.ClaimsSupported, oidc.ClaimAccessTokenHash)
	assert.Contains(t, metadata.ClaimsSupported, oidc.ClaimCodeHash)

	assert.True(t, metadata.ClaimsParameterSupported)
	assert.Equal(t, "", metadata.SignedMetadata)
}

func TestOAuth2AuthorizationServerMetadata(t *testing.T) {
	provider := newDiscoveryTestProvider(t, nil)

	metadata, err := provider.NewOAuth2AuthorizationServerMetadata(context.Background())
	require.NoError(t, err)

	data, err := json.Marshal(metadata)
	require.NoError(t, err)

	raw := map[string]any{}

	require.NoError(t, json.Unmarshal(data, &raw))

	assert.Equal(t, examplecom, raw["issuer"])
	assert.Equal(t, "https://example.com/api/oidc/token", raw["token_endpoint"])
	assert.Equal(t, "https://example.com/api/oidc/device-authorization", raw["device_authorization_endpoint"])
	assert.Equal(t, "https://example.com/api/oidc/pushed-authorization-request", raw["pushed_authorization_request_endpoint"])
	assert.Equal(t, []any{oidc.ClientAuthMethodClientSecretBasic, oidc.ClientAuthMethodClientSecretPost, oidc.ClientAuthMethodClientSecretJWT, oidc.ClientAuthMethodPrivateKeyJWT}, raw["introspection_endpoint_auth_methods_supported"])
	assert.Equal(t, true, raw["authorization_response_iss_parameter_supported"])

	assert.NotContains(t, raw, "userinfo_endpoint")
	assert.NotContains(t, raw, "claims_supported")
	assert.NotContains(t, raw, "signed_metadata")
}

func TestOpenIDConnectProviderMetadataOptions(t *testing.T) {
	testCases := []struct {
		Name   string
		Setup  func(config *schema.IdentityProvidersOpenIDConnect)
		Assert func(t *testing.T, raw map[string]any)
	}{
		{
			Name: "ShouldOmitFalseRequirements",
			Assert: func(t *testing.T, raw map[string]any) {
				assert.NotContains(t, raw, "require_signed_request_object")
				assert.NotContains(t, raw, "require_pushed_authorization_requests")
				assert.Contains(t, raw["request_object_signing_alg_values_supported"], oidc.SigningAlgNone)
			},
		},
		{
			Name: "ShouldRequireSignedRequestObject",
			Setup: func(config *schema.IdentityProvidersOpenIDConnect) {
				config.RequireSignedRequestObject = true
			},
			Assert: func(t *testing.T, raw map[string]any) {
				assert.Equal(t, true, raw["require_signed_request_object"])
				assert.NotContains(t, raw["request_object_signing_alg_values_supported"], oidc.SigningAlgNone)
			},
		},
		{
			Name: "ShouldRequirePushedAuthorizationRequests",
			Setup: func(config *schema.IdentityProvidersOpenIDConnect) {
				config.RequirePushedAuthorizationRequests = true
			},
			Assert: func(t *testing.T, raw map[string]any) {
				assert.Equal(t, true, raw["require_pushed_authorization_requests"])
			},
		},
		{
			Name: "ShouldIncludePlainChallengeMethod",
			Setup: func(config *schema.IdentityProvidersOpenIDConnect) {
				config.EnablePKCEPlainChallenge = true
			},
			Assert: func(t *testing.T, raw map[string]any) {
				assert.Equal(t, []any{oidc.PKCEChallengeMethodSHA256, oidc.PKCEChallengeMethodPlain}, raw["code_challenge_methods_supported"])
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := newDiscoveryTestProvider(t, tc.Setup)

			metadata, err := provider.NewOpenIDConnectProviderMetadata(context.Background())
			require.NoError(t, err)

			data, err := json.Marshal(metadata)
			require.NoError(t, err)

			raw := map[string]any{}

			require.NoError(t, json.Unmarshal(data, &raw))

			tc.Assert(t, raw)
		})
	}
}

func TestOpenIDConnectProviderMetadataSigned(t *testing.T) {
	testCases := []struct {
		Name     string
		KeyID    string
		Alg      string
		Signed   bool
		Expected string
	}{
		{"ShouldNotSignByDefault", "", "", false, ""},
		{"ShouldNotSignAlgNone", "", oidc.SigningAlgNone, false, ""},
		{"ShouldSignAlg", "", oidc.SigningAlgRSAUsingSHA256, true, "rs256"},
		{"ShouldSignKeyID", "rs256", oidc.SigningAlgRSAUsingSHA256, true, "rs256"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := newDiscoveryTestProvider(t, func(config *schema.IdentityProvidersOpenIDConnect) {
				config.DiscoverySignedResponseKeyID = tc.KeyID
				config.DiscoverySignedResponseAlg = tc.Alg
			})

			metadata, err := provider.NewOpenIDConnectProviderMetadata(context.Background())
			require.NoError(t, err)

			if !tc.Signed {
				assert.Equal(t, "", metadata.SignedMetadata)

				return
			}

			parts := strings.Split(metadata.SignedMetadata, ".")
			require.Len(t, parts, 3)

			header, claims := map[string]any{}, map[string]any{}

			raw, err := base64.RawURLEncoding.DecodeString(parts[0])
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(raw, &header))

			raw, err = base64.RawURLEncoding.DecodeString(parts[1])
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(raw, &claims))

			assert.Equal(t, oidc.SigningAlgRSAUsingSHA256, header["alg"])
			assert.Equal(t, tc.Expected, header["kid"])
			assert.Equal(t, examplecom, claims[oidc.ClaimIssuer])
			assert.Equal(t, "https://example.com/api/oidc/token", claims["token_endpoint"])
			assert.NotEmpty(t, claims[oidc.ClaimJWTID])
			assert.NotEmpty(t, claims[oidc.ClaimIssuedAt])
			assert.NotEmpty(t, claims[oidc.ClaimExpirationTime])
			assert.NotContains(t, claims, "signed_metadata")
		})
	}
}

func TestSigningAlgValuesSupported(t *testing.T) {
	algs := oidc.SigningAlgValuesSupported(oidc.SigningAlgNone)

	assert.Equal(t, oidc.SigningAlgHMACUsingSHA256, algs[0])
	assert.Equal(t, oidc.SigningAlgNone, algs[len(algs)-1])
	assert.Len(t, algs, len(oidc.SigningAlgValuesSupported())+1)
}

func newDiscoveryTestProvider(t *testing.T, setup func(config *schema.IdentityProvidersOpenIDConnect)) *oidc.OpenIDConnectProvider {
	t.Helper()

	config := &schema.IdentityProvidersOpenIDConnect{
		HMACSecret:                   badhmac,
		DiscoverySignedResponseAlg:   oidc.SigningAlgNone,
		DiscoverySignedResponseKeyID: "",
		JSONWebKeys: []schema.JWK{
			{KeyID: "rs256", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: x509PrivateKeyRSA2048},
		},
		Clients: []schema.IdentityProvidersOpenIDConnectClient{
			{
				ID:                  myclient,
				Secret:              tOpenIDConnectPlainTextClientSecret,
				AuthorizationPolicy: onefactor,
				RedirectURIs:        []string{examplecom},
			},
		},
	}

	if setup != nil {
		setup(config)
	}

	provider := oidc.NewOpenIDConnectProvider(&schema.Configuration{IdentityProviders: schema.IdentityProviders{OIDC: config}}, nil, nil)
	require.NotNil(t, provider)

	provider.Config.Issuer = examplecom

	return provider
}
