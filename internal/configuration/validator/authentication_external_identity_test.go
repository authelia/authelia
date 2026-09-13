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
			Name: "ShouldAllowOpenIDConnectProvidersWithDifferentIssuers",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "google", Name: "Google", Issuer: "https://accounts.google.com", ClientID: "abc", ClientSecret: "secret"},
					{ID: "example", Name: "Example", Issuer: "https://id.example.com", ClientID: "abc", ClientSecret: "secret"},
				},
			},
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
			Name: "ShouldRaiseErrorOnUnknownType",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{ID: "twitch", Type: "twitch", Name: "Twitch", ClientID: "123", ClientSecret: "secret"},
				},
			},
			Errors: []string{"authentication_backend: external_identity: providers: provider 'twitch': option 'type' must be one of 'openid_connect' but it's configured as 'twitch'"},
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

func TestValidateAuthenticationBackendExternalIdentityAuthenticationMethodsReference(t *testing.T) {
	testCases := []struct {
		Name   string
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
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := schema.AuthenticationBackendExternalIdentityProvider{
				ID: "example", Name: "Example", Issuer: "https://id.example.com", ClientID: "abc", ClientSecret: "secret",
				AuthenticationMethodsReference: tc.Have,
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
