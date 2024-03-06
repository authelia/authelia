package oidc_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOpenIDConnectProvider_NewOpenIDConnectProvider_NotConfigured(t *testing.T) {
	provider := oidc.NewOpenIDConnectProvider(nil, nil, nil)

	assert.Nil(t, provider)
}

func TestNewOpenIDConnectProvider_ShouldEnableOptionalDiscoveryValues(t *testing.T) {
	provider := oidc.NewOpenIDConnectProvider(&schema.IdentityProvidersOpenIDConnect{
		IssuerCertificateChain:   schema.X509CertificateChain{},
		IssuerPrivateKey:         x509PrivateKeyRSA2048,
		EnablePKCEPlainChallenge: true,
		HMACSecret:               badhmac,
		Clients: []schema.IdentityProvidersOpenIDConnectClient{
			{
				ID:                  myclient,
				Secret:              tOpenIDConnectPlainTextClientSecret,
				SectorIdentifierURI: &url.URL{Host: examplecomsid},
				AuthorizationPolicy: onefactor,
				RedirectURIs: []string{
					examplecom,
				},
			},
		},
	}, nil, nil)

	require.NotNil(t, provider)

	disco := provider.GetOpenIDConnectWellKnownConfiguration(examplecom)

	assert.Len(t, disco.SubjectTypesSupported, 2)
	assert.Contains(t, disco.SubjectTypesSupported, oidc.SubjectTypePublic)
	assert.Contains(t, disco.SubjectTypesSupported, oidc.SubjectTypePairwise)

	assert.Len(t, disco.CodeChallengeMethodsSupported, 2)
	assert.Contains(t, disco.CodeChallengeMethodsSupported, oidc.PKCEChallengeMethodSHA256)
	assert.Contains(t, disco.CodeChallengeMethodsSupported, oidc.PKCEChallengeMethodSHA256)
}

func TestOpenIDConnectProvider_NewOpenIDConnectProvider_GoodConfiguration(t *testing.T) {
	provider := oidc.NewOpenIDConnectProvider(&schema.IdentityProvidersOpenIDConnect{
		IssuerCertificateChain: schema.X509CertificateChain{},
		IssuerPrivateKey:       x509PrivateKeyRSA2048,
		HMACSecret:             badhmac,
		Clients: []schema.IdentityProvidersOpenIDConnectClient{
			{
				ID:                  "a-client",
				Secret:              tOpenIDConnectPlainTextClientSecret,
				AuthorizationPolicy: onefactor,
				RedirectURIs: []string{
					"https://google.com",
				},
			},
			{
				ID:                  "b-client",
				Name:                "Normal Description",
				Secret:              tOpenIDConnectPlainTextClientSecret,
				AuthorizationPolicy: twofactor,
				RedirectURIs: []string{
					"https://google.com",
				},
				Scopes: []string{
					oidc.ScopeGroups,
				},
				GrantTypes: []string{
					oidc.GrantTypeRefreshToken,
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
