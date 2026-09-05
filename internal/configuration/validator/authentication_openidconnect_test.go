package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidateAuthenticationBackendOpenIDConnect(t *testing.T) {
	testCases := []struct {
		Name     string
		Have     *schema.AuthenticationBackendOpenIDConnect
		Expected *schema.AuthenticationBackendOpenIDConnect
		Errors   []string
	}{
		{
			Name: "ShouldValidateMinimalProviderAndApplyDefaults",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
				},
			},
			Expected: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Scopes:                   []string{"openid", "profile", "email"},
						TokenEndpointAuthMethod:  "client_secret_basic",
						IDTokenSignedResponseAlg: "RS256",
						PKCE:                     schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "S256"},
					},
				},
			},
		},
		{
			Name: "ShouldForceOpenIDScope",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", Scopes: []string{"email"}},
				},
			},
			Expected: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Scopes:                   []string{"openid", "email"},
						TokenEndpointAuthMethod:  "client_secret_basic",
						IDTokenSignedResponseAlg: "RS256",
						PKCE:                     schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "S256"},
					},
				},
			},
		},
		{
			Name: "ShouldRaiseErrorOnInvalidID",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{ID: "Google!", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: openid_connect: providers: provider #1: option 'id' must match the pattern '^[a-z0-9][a-z0-9_-]{0,31}$' but it's configured as 'Google!'"},
		},
		{
			Name: "ShouldRaiseErrorOnDuplicateID",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
					{ID: "google", Name: "Google 2", Issuer: "https://accounts.google.com/2", ClientID: "abc", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: openid_connect: providers: provider 'google': option 'id' must be unique but it's configured multiple times"},
		},
		{
			Name: "ShouldRaiseErrorOnInsecureIssuer",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{ID: "google", Name: "Google", Issuer: "http://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: openid_connect: providers: provider 'google': option 'issuer' must have the 'https' scheme but it's configured as 'http'"},
		},
		{
			Name: "ShouldRaiseErrorOnInsecureEndpoints",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Endpoints: schema.AuthenticationBackendOpenIDConnectProviderEndpoints{ //nolint:gosec // Test URLs.
							Authorization: "http://accounts.google.com/authorize",
							Token:         "http://accounts.google.com/token",
							JSONWebKeys:   "http://accounts.google.com/jwks.json",
						},
					},
				},
			},
			Errors: []string{
				"authentication_backend: openid_connect: providers: provider 'google': endpoints: option 'authorization' must have the 'https' scheme but it's configured as 'http'",
				"authentication_backend: openid_connect: providers: provider 'google': endpoints: option 'token' must have the 'https' scheme but it's configured as 'http'",
				"authentication_backend: openid_connect: providers: provider 'google': endpoints: option 'jwks' must have the 'https' scheme but it's configured as 'http'",
			},
		},
		{
			Name: "ShouldRaiseErrorOnUnparsableEndpoint",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Endpoints: schema.AuthenticationBackendOpenIDConnectProviderEndpoints{ //nolint:gosec // Test URLs.
							Token: "https://accounts.google.com/token\x00",
						},
					},
				},
			},
			Errors: []string{
				"authentication_backend: openid_connect: providers: provider 'google': endpoints: option 'token' must be a valid URL but it could not be parsed: parse \"https://accounts.google.com/token\\x00\": net/url: invalid control character in URL",
			},
		},
		{
			Name: "ShouldRaiseErrorOnSymmetricAlg",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", IDTokenSignedResponseAlg: "HS256"},
				},
			},
			Errors: []string{"authentication_backend: openid_connect: providers: provider 'google': option 'id_token_signed_response_alg' must be one of 'ES256', 'ES384', 'ES512', 'PS256', 'PS384', 'PS512', 'RS256', 'RS384', or 'RS512' but it's configured as 'HS256'"},
		},
		{
			Name: "ShouldRaiseErrorOnPlainPKCE",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", PKCE: schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "plain"}},
				},
			},
			Errors: []string{"authentication_backend: openid_connect: providers: provider 'google': pkce: option 'challenge_method' must be 'S256' but it's configured as 'plain'"},
		},
		{
			Name: "ShouldRaiseErrorOnDiscoveryDisabledWithoutEndpoints",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", Discovery: schema.AuthenticationBackendOpenIDConnectProviderDiscovery{Disable: true}},
				},
			},
			Errors: []string{
				"authentication_backend: openid_connect: providers: provider 'google': endpoints: option 'authorization' is required when discovery is disabled but it's not configured",
				"authentication_backend: openid_connect: providers: provider 'google': endpoints: option 'token' is required when discovery is disabled but it's not configured",
				"authentication_backend: openid_connect: providers: provider 'google': endpoints: option 'jwks' is required when discovery is disabled and no 'jwks' are configured but it's not configured",
			},
		},
		{
			Name:   "ShouldRaiseErrorOnNoProviders",
			Have:   &schema.AuthenticationBackendOpenIDConnect{},
			Errors: []string{"authentication_backend: openid_connect: option 'providers' is required but it's not configured"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			val := schema.NewStructValidator()

			ValidateAuthenticationBackendOpenIDConnect(tc.Have, val)

			assert.Len(t, val.Warnings(), 0)

			errs := val.Errors()

			require.Len(t, errs, len(tc.Errors))

			for i, expected := range tc.Errors {
				assert.EqualError(t, errs[i], expected)
			}

			if tc.Expected != nil {
				assert.Equal(t, tc.Expected, tc.Have)
			}
		})
	}
}
