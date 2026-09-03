// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestValidateAuthenticationBackendExternalIdentity(t *testing.T) {
	testCases := []struct {
		Name     string
		Have     *schema.AuthenticationBackendExternalIdentity
		Expected *schema.AuthenticationBackendExternalIdentity
		Errors   []string
	}{
		{
			Name: "ShouldValidateMinimalProviderAndApplyDefaults",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
				},
			},
			Expected: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Scopes:                  []string{"openid", "profile", "email"},
						Type:                    "openid_connect",
						ResponseMode:            "query",
						TokenEndpointAuthMethod: "client_secret_basic",
						PKCE:                    schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
					},
				},
			},
		},
		{
			Name: "ShouldDefaultIDTokenSigningAlgWhenDiscoveryIsDisabled",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Discovery: schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true},
						Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{ //nolint:gosec // Test URLs.
							Authorization: "https://accounts.google.com/authorize",
							Token:         "https://accounts.google.com/token",
							JSONWebKeys:   "https://accounts.google.com/jwks.json",
						},
					},
				},
			},
			Expected: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Scopes:                   []string{"openid", "profile", "email"},
						Type:                     "openid_connect",
						ResponseMode:             "query",
						TokenEndpointAuthMethod:  "client_secret_basic",
						IDTokenSignedResponseAlg: "RS256",
						PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
						Discovery:                schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true},
						Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{ //nolint:gosec // Test URLs.
							Authorization: "https://accounts.google.com/authorize",
							Token:         "https://accounts.google.com/token",
							JSONWebKeys:   "https://accounts.google.com/jwks.json",
						},
					},
				},
			},
		},
		{
			Name: "ShouldForceOpenIDScope",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", Scopes: []string{"email"}},
				},
			},
			Expected: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Scopes:                  []string{"openid", "email"},
						Type:                    "openid_connect",
						ResponseMode:            "query",
						TokenEndpointAuthMethod: "client_secret_basic",
						PKCE:                    schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
					},
				},
			},
		},
		{
			Name: "ShouldRaiseErrorOnInvalidID",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "Google!", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider #1: option 'id' must match the pattern '^[a-z0-9][a-z0-9_-]{0,31}$' but it's configured as 'Google!'"},
		},
		{
			Name: "ShouldRaiseErrorOnDuplicateID",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
					{ID: "google", Name: "Google 2", Issuer: "https://accounts.google.com/2", ClientID: "abc", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'google': option 'id' must be unique but it's configured multiple times"},
		},
		{
			Name: "ShouldAllowOneDiscordAndPlexProviderAndOpenIDConnectProvidersWithDifferentIssuers",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
					{ID: "example", Name: "Example", Issuer: "https://id.example.com", ClientID: "abc", ClientSecret: "secret"},
					{ID: "discord", Type: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret"},
					{ID: "plex", Type: "plex", Name: "Plex", ClientID: "authelia-client"},
				},
			},
		},
		{
			Name: "ShouldRaiseErrorOnMultipleDiscordProviders",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "discord", Type: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret"},
					{ID: "discord2", Type: "discord", Name: "Discord 2", ClientID: "456", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: option 'type' must only be 'discord' for one provider but it's configured for the providers 'discord' and 'discord2'"},
		},
		{
			Name: "ShouldRaiseErrorOnMultiplePlexProviders",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "plex", Type: "plex", Name: "Plex", ClientID: "a"},
					{ID: "plex2", Type: "plex", Name: "Plex 2", ClientID: "b"},
					{ID: "plex3", Type: "plex", Name: "Plex 3", ClientID: "c"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: option 'type' must only be 'plex' for one provider but it's configured for the providers 'plex', 'plex2', and 'plex3'"},
		},
		{
			Name: "ShouldRaiseErrorOnMultipleGitHubProviders",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "github", Type: "github", Name: "GitHub", ClientID: "a", ClientSecret: "secret"},
					{ID: "github2", Type: "github", Name: "GitHub 2", ClientID: "b", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: option 'type' must only be 'github' for one provider but it's configured for the providers 'github' and 'github2'"},
		},
		{
			Name: "ShouldRaiseErrorOnDuplicateOpenIDConnectIssuer",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
					{ID: "example", Name: "Example", Issuer: "https://id.example.com", ClientID: "abc", ClientSecret: "secret"},
					{ID: "google2", Type: "openid_connect", Name: "Google 2", Issuer: "https://accounts.google.com", ClientID: "def", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: option 'issuer' must be unique for 'openid_connect' providers but 'https://accounts.google.com' is configured for the providers 'google' and 'google2'"},
		},
		{
			Name: "ShouldNotReportDuplicateTypesForProvidersWithADuplicateID",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "discord", Type: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret"},
					{ID: "discord", Type: "discord", Name: "Discord 2", ClientID: "456", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'discord': option 'id' must be unique but it's configured multiple times"},
		},
		{
			Name: "ShouldRaiseErrorOnInsecureIssuer",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "http://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'google': option 'issuer' must have the 'https' scheme but it's configured as 'http'"},
		},
		{
			Name: "ShouldRaiseErrorOnInsecureEndpoints",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{ //nolint:gosec // Test URLs.
							Authorization: "http://accounts.google.com/authorize",
							Token:         "http://accounts.google.com/token",
							UserInfo:      "http://accounts.google.com/userinfo",
							JSONWebKeys:   "http://accounts.google.com/jwks.json",
						},
					},
				},
			},
			Errors: []string{
				"authentication_backend: external_identity: providers: provider 'google': endpoints: option 'authorization' must have the 'https' scheme but it's configured as 'http'",
				"authentication_backend: external_identity: providers: provider 'google': endpoints: option 'token' must have the 'https' scheme but it's configured as 'http'",
				"authentication_backend: external_identity: providers: provider 'google': endpoints: option 'userinfo' must have the 'https' scheme but it's configured as 'http'",
				"authentication_backend: external_identity: providers: provider 'google': endpoints: option 'jwks' must have the 'https' scheme but it's configured as 'http'",
			},
		},
		{
			Name: "ShouldRaiseErrorOnUnparsableEndpoint",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{ //nolint:gosec // Test URLs.
							Token: "https://accounts.google.com/token\x00",
						},
					},
				},
			},
			Errors: []string{
				"authentication_backend: external_identity: providers: provider 'google': endpoints: option 'token' must be a valid URL but it could not be parsed: parse \"https://accounts.google.com/token\\x00\": net/url: invalid control character in URL",
			},
		},
		{
			Name: "ShouldRaiseErrorOnInvalidResponseMode",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", ResponseMode: "fragment"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'google': option 'response_mode' must be one of 'query' or 'form_post' but it's configured as 'fragment'"},
		},
		{
			Name: "ShouldAcceptFormPostResponseMode",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", ResponseMode: "form_post"},
				},
			},
		},
		{
			Name: "ShouldRaiseErrorOnSymmetricAlg",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", IDTokenSignedResponseAlg: "HS256"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'google': option 'id_token_signed_response_alg' must be one of 'ES256', 'ES384', 'ES512', 'PS256', 'PS384', 'PS512', 'RS256', 'RS384', or 'RS512' but it's configured as 'HS256'"},
		},
		{
			Name: "ShouldRaiseErrorOnUnsupportedUserInfoSigningAlg",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", UserInfoSignedResponseAlg: "none"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'google': option 'userinfo_signed_response_alg' must be one of 'ES256', 'ES384', 'ES512', 'PS256', 'PS384', 'PS512', 'RS256', 'RS384', or 'RS512' but it's configured as 'none'"},
		},
		{
			Name: "ShouldAllowUserInfoSigningAlg",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", UserInfoSignedResponseAlg: "ES256"},
				},
			},
		},
		{
			Name: "ShouldRaiseErrorOnPlainPKCE",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", PKCE: schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "plain"}},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'google': pkce: option 'challenge_method' must be 'S256' but it's configured as 'plain'"},
		},
		{
			Name: "ShouldRaiseErrorOnDiscoveryDisabledWithoutEndpoints",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret", Discovery: schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true}},
				},
			},
			Errors: []string{
				"authentication_backend: external_identity: providers: provider 'google': endpoints: option 'authorization' is required when discovery is disabled but it's not configured",
				"authentication_backend: external_identity: providers: provider 'google': endpoints: option 'token' is required when discovery is disabled but it's not configured",
				"authentication_backend: external_identity: providers: provider 'google': endpoints: option 'jwks' is required when discovery is disabled and no 'jwks' are configured but it's not configured",
			},
		},
		{
			Name: "ShouldRaiseErrorOnRequiredPushedAuthorizationRequestsWithoutEndpointWhenDiscoveryIsDisabled",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						RequirePushedAuthorizationRequests: true,
						Discovery:                          schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true},
						Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{ //nolint:gosec // Test URLs.
							Authorization: "https://accounts.google.com/authorize",
							Token:         "https://accounts.google.com/token",
							JSONWebKeys:   "https://accounts.google.com/jwks.json",
						},
					},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'google': endpoints: option 'pushed_authorization_request' is required when 'require_pushed_authorization_requests' is enabled and discovery is disabled but it's not configured"},
		},
		{
			Name: "ShouldRaiseErrorOnInsecurePushedAuthorizationRequestEndpoint",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret",
						Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{PushedAuthorizationRequest: "http://accounts.google.com/par"},
					},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'google': endpoints: option 'pushed_authorization_request' must have the 'https' scheme but it's configured as 'http'"},
		},
		{
			Name:   "ShouldRaiseErrorOnNoProviders",
			Have:   &schema.AuthenticationBackendExternalIdentity{},
			Errors: []string{"authentication_backend: external_identity: option 'providers' is required but it's not configured"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			val := schema.NewStructValidator()

			ValidateAuthenticationBackendExternalIdentity(tc.Have, val)

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

func TestValidateAuthenticationBackendExternalIdentityAllowDuplicateIssuer(t *testing.T) {
	have := func() *schema.AuthenticationBackendExternalIdentity {
		return &schema.AuthenticationBackendExternalIdentity{
			Providers: []schema.AuthenticationBackendExternalIdentityProvider{
				{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
				{ID: "google2", Name: "Google 2", Issuer: "https://accounts.google.com", ClientID: "def", ClientSecret: "secret"},
			},
		}
	}

	errDuplicate := "authentication_backend: external_identity: providers: option 'issuer' must be unique for 'openid_connect' providers but 'https://accounts.google.com' is configured for the providers 'google' and 'google2'"

	testCases := []struct {
		name  string
		value string
		skip  bool
	}{
		{"ShouldSkipWhenTrue", "true", utils.Dev},
		{"ShouldNotSkipWhenTrueUppercase", "TRUE", false},
		{"ShouldNotSkipWhenOne", "1", false},
		{"ShouldNotSkipWhenEmpty", "", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(envExternalIdentityOpenIDConnectAllowDuplicateIssuer, tc.value)

			val := schema.NewStructValidator()

			ValidateAuthenticationBackendExternalIdentity(have(), val)

			if tc.skip {
				assert.Len(t, val.Errors(), 0)
			} else {
				require.Len(t, val.Errors(), 1)
				assert.EqualError(t, val.Errors()[0], errDuplicate)
			}
		})
	}
}

func TestValidateAuthenticationBackendExternalIdentityDiscord(t *testing.T) {
	testCases := []struct {
		Name     string
		Have     schema.AuthenticationBackendExternalIdentityProvider
		Expected *schema.AuthenticationBackendExternalIdentityProvider
		Errors   []string
	}{
		{
			Name: "ShouldApplyDefaults",
			Have: schema.AuthenticationBackendExternalIdentityProvider{ID: "discord", Type: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret"},
			Expected: &schema.AuthenticationBackendExternalIdentityProvider{
				ID: "discord", Type: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret",
				Scopes:                  []string{"identify", "email"},
				ResponseMode:            "query",
				TokenEndpointAuthMethod: "client_secret_basic",
				PKCE:                    schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
			},
		},
		{
			Name: "ShouldForceIdentifyScope",
			Have: schema.AuthenticationBackendExternalIdentityProvider{ID: "discord", Type: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret", Scopes: []string{"email"}, TokenEndpointAuthMethod: "client_secret_post"},
			Expected: &schema.AuthenticationBackendExternalIdentityProvider{
				ID: "discord", Type: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret",
				Scopes:                  []string{"identify", "email"},
				ResponseMode:            "query",
				TokenEndpointAuthMethod: "client_secret_post",
				PKCE:                    schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
			},
		},
		{
			Name: "ShouldRaiseErrorOnOpenIDConnectOptions",
			Have: schema.AuthenticationBackendExternalIdentityProvider{
				ID: "discord", Type: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret",
				Issuer:                         "https://discord.com",
				IDTokenSignedResponseAlg:       "RS256",
				AuthenticationMethodsReference: schema.AuthenticationBackendExternalIdentityProviderAMR{Trust: true},
				Discovery:                      schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true},
				Endpoints:                      schema.AuthenticationBackendExternalIdentityProviderEndpoints{Authorization: "https://discord.com/oauth2/authorize", Token: "https://discord.com/api/oauth2/token", UserInfo: "https://discord.com/api/v10/users/@me", JSONWebKeys: "https://discord.com/api/oauth2/keys"}, //nolint:gosec // Test URLs.
				JSONWebKeys:                    []schema.JWK{{KeyID: "kid1"}},
			},
			Errors: []string{
				"authentication_backend: external_identity: providers: provider 'discord': option 'issuer' is not supported by the 'discord' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'discord': option 'id_token_signed_response_alg' is not supported by the 'discord' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'discord': option 'authentication_methods_reference.trust' is not supported by the 'discord' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'discord': option 'discovery.disable' is not supported by the 'discord' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'discord': option 'endpoints' is not supported by the 'discord' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'discord': option 'jwks' is not supported by the 'discord' provider type but it's configured",
			},
		},
		{
			Name:   "ShouldRaiseErrorOnFormPostResponseMode",
			Have:   schema.AuthenticationBackendExternalIdentityProvider{ID: "discord", Type: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret", ResponseMode: "form_post"},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'discord': option 'response_mode' must be one of 'query' but it's configured as 'form_post'"},
		},
		{
			Name:   "ShouldRaiseErrorOnNoneAuthMethod",
			Have:   schema.AuthenticationBackendExternalIdentityProvider{ID: "discord", Type: "discord", Name: "Discord", ClientID: "123", TokenEndpointAuthMethod: "none"},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'discord': option 'token_endpoint_auth_method' must be one of 'client_secret_basic' or 'client_secret_post' but it's configured as 'none'"},
		},
		{
			Name:   "ShouldRaiseErrorOnMissingSecret",
			Have:   schema.AuthenticationBackendExternalIdentityProvider{ID: "discord", Type: "discord", Name: "Discord", ClientID: "123"},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'discord': option 'client_secret' is required when the 'token_endpoint_auth_method' is 'client_secret_basic' but it's not configured"},
		},
		{
			Name:   "ShouldRaiseErrorOnUnknownType",
			Have:   schema.AuthenticationBackendExternalIdentityProvider{ID: "twitch", Type: "twitch", Name: "Twitch", ClientID: "123", ClientSecret: "secret"},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'twitch': option 'type' must be one of 'openid_connect', 'discord', 'plex', or 'github' but it's configured as 'twitch'"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			config := &schema.AuthenticationBackendExternalIdentity{Providers: []schema.AuthenticationBackendExternalIdentityProvider{tc.Have}}
			val := schema.NewStructValidator()

			ValidateAuthenticationBackendExternalIdentity(config, val)

			errs := make([]string, 0, len(val.Errors()))

			for _, err := range val.Errors() {
				errs = append(errs, err.Error())
			}

			if len(tc.Errors) == 0 {
				assert.Empty(t, errs)
			} else {
				assert.Equal(t, tc.Errors, errs)
			}

			if tc.Expected != nil {
				assert.Equal(t, *tc.Expected, config.Providers[0])
			}
		})
	}
}

func TestValidateAuthenticationBackendExternalIdentityGitHub(t *testing.T) {
	testCases := []struct {
		Name     string
		Have     schema.AuthenticationBackendExternalIdentityProvider
		Expected *schema.AuthenticationBackendExternalIdentityProvider
		Errors   []string
	}{
		{
			Name: "ShouldApplyDefaults",
			Have: schema.AuthenticationBackendExternalIdentityProvider{ID: "github", Type: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret"},
			Expected: &schema.AuthenticationBackendExternalIdentityProvider{
				ID: "github", Type: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret",
				Scopes:                  []string{"read:user", "user:email"},
				ResponseMode:            "query",
				TokenEndpointAuthMethod: "client_secret_basic",
				PKCE:                    schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
			},
		},
		{
			Name: "ShouldKeepTheConfiguredScopes",
			Have: schema.AuthenticationBackendExternalIdentityProvider{ID: "github", Type: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret", Scopes: []string{"read:user"}, TokenEndpointAuthMethod: "client_secret_post"},
			Expected: &schema.AuthenticationBackendExternalIdentityProvider{
				ID: "github", Type: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret",
				Scopes:                  []string{"read:user"},
				ResponseMode:            "query",
				TokenEndpointAuthMethod: "client_secret_post",
				PKCE:                    schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
			},
		},
		{
			Name: "ShouldRaiseErrorOnOpenIDConnectOptions",
			Have: schema.AuthenticationBackendExternalIdentityProvider{
				ID: "github", Type: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret",
				Issuer:                   "https://github.com",
				IDTokenSignedResponseAlg: "RS256",
			},
			Errors: []string{
				"authentication_backend: external_identity: providers: provider 'github': option 'issuer' is not supported by the 'github' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'github': option 'id_token_signed_response_alg' is not supported by the 'github' provider type but it's configured",
			},
		},
		{
			Name:   "ShouldRaiseErrorOnFormPostResponseMode",
			Have:   schema.AuthenticationBackendExternalIdentityProvider{ID: "github", Type: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret", ResponseMode: "form_post"},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'github': option 'response_mode' must be one of 'query' but it's configured as 'form_post'"},
		},
		{
			Name:   "ShouldRaiseErrorOnMissingSecret",
			Have:   schema.AuthenticationBackendExternalIdentityProvider{ID: "github", Type: "github", Name: "GitHub", ClientID: "Iv1.abc"},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'github': option 'client_secret' is required when the 'token_endpoint_auth_method' is 'client_secret_basic' but it's not configured"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			config := &schema.AuthenticationBackendExternalIdentity{Providers: []schema.AuthenticationBackendExternalIdentityProvider{tc.Have}}
			val := schema.NewStructValidator()

			ValidateAuthenticationBackendExternalIdentity(config, val)

			errs := make([]string, 0, len(val.Errors()))

			for _, err := range val.Errors() {
				errs = append(errs, err.Error())
			}

			if len(tc.Errors) == 0 {
				assert.Empty(t, errs)
			} else {
				assert.Equal(t, tc.Errors, errs)
			}

			if tc.Expected != nil {
				assert.Equal(t, *tc.Expected, config.Providers[0])
			}
		})
	}
}

func TestValidateAuthenticationBackendExternalIdentityPlex(t *testing.T) {
	testCases := []struct {
		Name     string
		Have     schema.AuthenticationBackendExternalIdentityProvider
		Expected *schema.AuthenticationBackendExternalIdentityProvider
		Errors   []string
	}{
		{
			Name: "ShouldApplyDefaults",
			Have: schema.AuthenticationBackendExternalIdentityProvider{ID: "plex", Type: "plex", Name: "Plex", ClientID: "authelia-client"},
			Expected: &schema.AuthenticationBackendExternalIdentityProvider{
				ID: "plex", Type: "plex", Name: "Plex", ClientID: "authelia-client",
				ResponseMode: "query",
			},
		},
		{
			Name: "ShouldAllowDefaultAuthenticationMethodsReference",
			Have: schema.AuthenticationBackendExternalIdentityProvider{ID: "plex", Type: "plex", Name: "Plex", ClientID: "authelia-client", AuthenticationMethodsReference: schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd"}}},
		},
		{
			Name: "ShouldRaiseErrorOnUnsupportedOptions",
			Have: schema.AuthenticationBackendExternalIdentityProvider{
				ID: "plex", Type: "plex", Name: "Plex", ClientID: "authelia-client",
				ClientSecret:                   "secret",
				Scopes:                         []string{"openid"},
				TokenEndpointAuthMethod:        "client_secret_basic",
				PKCE:                           schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
				Issuer:                         "https://plex.tv",
				IDTokenSignedResponseAlg:       "RS256",
				AuthenticationMethodsReference: schema.AuthenticationBackendExternalIdentityProviderAMR{Trust: true, Default: []string{"pwd"}, Override: true},
				Discovery:                      schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true},
				Endpoints:                      schema.AuthenticationBackendExternalIdentityProviderEndpoints{Token: "https://plex.tv/api/v2/pins"}, //nolint:gosec // Test URL.
				JSONWebKeys:                    []schema.JWK{{KeyID: "kid1"}},
			},
			Errors: []string{
				"authentication_backend: external_identity: providers: provider 'plex': option 'client_secret' is not supported by the 'plex' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'plex': option 'scopes' is not supported by the 'plex' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'plex': option 'token_endpoint_auth_method' is not supported by the 'plex' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'plex': option 'pkce.challenge_method' is not supported by the 'plex' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'plex': option 'issuer' is not supported by the 'plex' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'plex': option 'id_token_signed_response_alg' is not supported by the 'plex' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'plex': option 'authentication_methods_reference.trust' is not supported by the 'plex' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'plex': option 'authentication_methods_reference.override' is not supported by the 'plex' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'plex': option 'discovery.disable' is not supported by the 'plex' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'plex': option 'endpoints' is not supported by the 'plex' provider type but it's configured",
				"authentication_backend: external_identity: providers: provider 'plex': option 'jwks' is not supported by the 'plex' provider type but it's configured",
			},
		},
		{
			Name:   "ShouldRaiseErrorOnFormPostResponseMode",
			Have:   schema.AuthenticationBackendExternalIdentityProvider{ID: "plex", Type: "plex", Name: "Plex", ClientID: "authelia-client", ResponseMode: "form_post"},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'plex': option 'response_mode' must be one of 'query' but it's configured as 'form_post'"},
		},
		{
			Name:   "ShouldRaiseErrorOnMissingClientID",
			Have:   schema.AuthenticationBackendExternalIdentityProvider{ID: "plex", Type: "plex", Name: "Plex"},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'plex': option 'client_id' is required but it's not configured"},
		},
		{
			Name:   "ShouldRaiseErrorOnEmptyDefaultAuthenticationMethodsReference",
			Have:   schema.AuthenticationBackendExternalIdentityProvider{ID: "plex", Type: "plex", Name: "Plex", ClientID: "authelia-client", AuthenticationMethodsReference: schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{""}}},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'plex': authentication_methods_reference: option 'default' must not contain empty values"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			config := &schema.AuthenticationBackendExternalIdentity{Providers: []schema.AuthenticationBackendExternalIdentityProvider{tc.Have}}
			val := schema.NewStructValidator()

			ValidateAuthenticationBackendExternalIdentity(config, val)

			errs := make([]string, 0, len(val.Errors()))

			for _, err := range val.Errors() {
				errs = append(errs, err.Error())
			}

			if len(tc.Errors) == 0 {
				assert.Empty(t, errs)
			} else {
				assert.Equal(t, tc.Errors, errs)
			}

			if tc.Expected != nil {
				assert.Equal(t, *tc.Expected, config.Providers[0])
			}
		})
	}
}

func TestValidateAuthenticationBackendExternalIdentityAuthenticationMethodsReference(t *testing.T) {
	testCases := []struct {
		Name   string
		Type   string
		Have   schema.AuthenticationBackendExternalIdentityProviderAMR
		Errors []string
	}{
		{
			Name: "ShouldAllowDefaultWithoutTrust",
			Have: schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd", "otp"}},
		},
		{
			Name: "ShouldAllowOverride",
			Have: schema.AuthenticationBackendExternalIdentityProviderAMR{Trust: true, Default: []string{"pwd"}, Override: true},
		},
		{
			Name:   "ShouldRaiseErrorOnOverrideWithoutDefault",
			Have:   schema.AuthenticationBackendExternalIdentityProviderAMR{Trust: true, Override: true},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'example': authentication_methods_reference: option 'default' is required when 'override' is enabled but it's not configured"},
		},
		{
			Name:   "ShouldRaiseErrorOnEmptyDefault",
			Have:   schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd", ""}},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'example': authentication_methods_reference: option 'default' must not contain empty values"},
		},
		{
			Name: "ShouldAllowDefaultForDiscord",
			Type: "discord",
			Have: schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd"}},
		},
		{
			Name:   "ShouldRaiseErrorOnEmptyDefaultForDiscord",
			Type:   "discord",
			Have:   schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{""}},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'example': authentication_methods_reference: option 'default' must not contain empty values"},
		},
		{
			Name: "ShouldRaiseErrorOnOverrideForDiscord",
			Type: "discord",
			Have: schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd"}, Override: true},
			Errors: []string{
				"authentication_backend: external_identity: providers: provider 'example': option 'authentication_methods_reference.override' is not supported by the 'discord' provider type but it's configured",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := schema.AuthenticationBackendExternalIdentityProvider{
				ID: "example", Type: tc.Type, Name: "Example", ClientID: "abc", ClientSecret: "secret",
				AuthenticationMethodsReference: tc.Have,
			}

			if tc.Type != "discord" {
				provider.Issuer = "https://id.example.com"
			}

			config := &schema.AuthenticationBackendExternalIdentity{Providers: []schema.AuthenticationBackendExternalIdentityProvider{provider}}
			val := schema.NewStructValidator()

			ValidateAuthenticationBackendExternalIdentity(config, val)

			errs := make([]string, 0, len(val.Errors()))

			for _, err := range val.Errors() {
				errs = append(errs, err.Error())
			}

			if len(tc.Errors) == 0 {
				assert.Empty(t, errs)
			} else {
				assert.Equal(t, tc.Errors, errs)
			}

			assert.Equal(t, tc.Have, config.Providers[0].AuthenticationMethodsReference)
		})
	}
}
