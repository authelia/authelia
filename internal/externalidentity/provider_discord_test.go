// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

func TestDiscordProviderAuthorizationRequest(t *testing.T) {
	provider := newDiscordProvider(&schema.AuthenticationBackendExternalIdentityProvider{
		ID: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret",
		Scopes: []string{"identify", "email"}, TokenEndpointAuthMethod: "client_secret_basic",
	}, newProviderClient(nil))

	request, err := provider.AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/api/firstfactor/external-identity/discord/callback"})
	require.NoError(t, err)

	assert.Len(t, request.State, 43)
	assert.Empty(t, request.Nonce, "Discord issues no ID Token, so there is no nonce to bind it to")
	assert.Len(t, request.CodeVerifier, 43)

	uri, err := url.Parse(request.URL)
	require.NoError(t, err)

	assert.Equal(t, "https", uri.Scheme)
	assert.Equal(t, "discord.com", uri.Host)
	assert.Equal(t, "/oauth2/authorize", uri.Path)

	query := uri.Query()

	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "123", query.Get("client_id"))
	assert.Equal(t, "identify email", query.Get("scope"))
	assert.Equal(t, "https://auth.example.com/api/firstfactor/external-identity/discord/callback", query.Get("redirect_uri"))
	assert.Equal(t, request.State, query.Get("state"))
	assert.Equal(t, "S256", query.Get("code_challenge_method"))
	assert.NotEmpty(t, query.Get("code_challenge"))
	assert.False(t, query.Has("nonce"))
	assert.False(t, query.Has("response_mode"))
}

func TestDiscordProviderComplete(t *testing.T) {
	testCases := []struct {
		Name       string
		Method     string
		UserStatus int
		User       any
		Expected   *IdentityClaims
		Error      string
	}{
		{
			Name:       "ShouldReturnTheUserWithVerifiedEmail",
			Method:     "client_secret_basic",
			UserStatus: http.StatusOK,
			User:       map[string]any{"id": "80351110224678912", "username": "nelly", "global_name": "Nelly", "email": "nelly@discord.com", "verified": true},
			Expected:   &IdentityClaims{Issuer: "https://discord.com", Subject: "80351110224678912", PreferredUsername: "nelly", Name: "Nelly", Email: "nelly@discord.com"},
		},
		{
			Name:       "ShouldAuthenticateWithThePostMethod",
			Method:     "client_secret_post",
			UserStatus: http.StatusOK,
			User:       map[string]any{"id": "80351110224678912", "username": "nelly", "global_name": "Nelly"},
			Expected:   &IdentityClaims{Issuer: "https://discord.com", Subject: "80351110224678912", PreferredUsername: "nelly", Name: "Nelly"},
		},
		{
			Name:       "ShouldOmitAnUnverifiedEmail",
			Method:     "client_secret_basic",
			UserStatus: http.StatusOK,
			User:       map[string]any{"id": "80351110224678912", "username": "nelly", "email": "nelly@discord.com", "verified": false},
			Expected:   &IdentityClaims{Issuer: "https://discord.com", Subject: "80351110224678912", PreferredUsername: "nelly", Name: "nelly"},
		},
		{
			Name:       "ShouldUseTheUsernameWithoutAGlobalName",
			Method:     "client_secret_basic",
			UserStatus: http.StatusOK,
			User:       map[string]any{"id": "80351110224678912", "username": "nelly", "global_name": nil},
			Expected:   &IdentityClaims{Issuer: "https://discord.com", Subject: "80351110224678912", PreferredUsername: "nelly", Name: "nelly"},
		},
		{
			Name:       "ShouldRaiseErrorOnAbsentID",
			Method:     "client_secret_basic",
			UserStatus: http.StatusOK,
			User:       map[string]any{"username": "nelly"},
			Error:      "error requesting the discord user: the discord user response is invalid: the 'id' is required but it is absent",
		},
		{
			Name:       "ShouldRaiseErrorOnErrorStatus",
			Method:     "client_secret_basic",
			UserStatus: http.StatusUnauthorized,
			User:       map[string]any{"message": "401: Unauthorized", "code": 0},
			Error:      "error requesting the discord user: the discord user response is invalid: the user endpoint returned status code 401",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mux := http.NewServeMux()

			mux.HandleFunc("/api/oauth2/token", func(rw http.ResponseWriter, r *http.Request) {
				require.NoError(t, r.ParseForm())

				assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get(headerContentType))
				assert.Equal(t, "authorization_code", r.PostForm.Get("grant_type"))
				assert.Equal(t, "the-code", r.PostForm.Get("code"))
				assert.Equal(t, "the-verifier", r.PostForm.Get("code_verifier"))
				assert.Equal(t, "https://auth.example.com/cb", r.PostForm.Get("redirect_uri"))

				switch tc.Method {
				case "client_secret_basic":
					username, password, ok := r.BasicAuth()

					assert.True(t, ok)
					assert.Equal(t, "123", username)
					assert.Equal(t, "secret", password)
				case "client_secret_post":
					assert.Equal(t, "123", r.PostForm.Get("client_id"))
					assert.Equal(t, "secret", r.PostForm.Get("client_secret"))
				}

				rw.Header().Set(headerContentType, mimeApplicationJSON)

				_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "at", "token_type": "Bearer", "expires_in": 604800, "scope": "identify email"})
			})

			mux.HandleFunc("/api/v10/users/@me", func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer at", r.Header.Get(headerAuthorization))

				rw.Header().Set(headerContentType, mimeApplicationJSON)
				rw.WriteHeader(tc.UserStatus)

				_ = json.NewEncoder(rw).Encode(tc.User)
			})

			server := httptest.NewServer(mux)

			defer server.Close()

			provider := newTestDiscordProvider(server.URL, tc.Method)

			claims, err := provider.Complete(context.Background(), CompletionRequest{Code: "the-code", CodeVerifier: "the-verifier", RedirectURI: "https://auth.example.com/cb", Now: time.Now()})

			if tc.Error != "" {
				assert.Nil(t, claims)
				require.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, claims)
		})
	}
}

func TestDiscordProviderCompleteShouldRaiseErrorOnTokenError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/oauth2/token" {
			t.Errorf("the user endpoint must not be requested when the token exchange fails, got %s", r.URL.Path)
		}

		rw.Header().Set(headerContentType, mimeApplicationJSON)
		rw.WriteHeader(http.StatusBadRequest)

		_ = json.NewEncoder(rw).Encode(map[string]any{"error": "invalid_grant"})
	}))

	defer server.Close()

	provider := newTestDiscordProvider(server.URL, "client_secret_basic")

	claims, err := provider.Complete(context.Background(), CompletionRequest{Code: "the-code", CodeVerifier: "the-verifier", RedirectURI: "https://auth.example.com/cb"})

	assert.Nil(t, claims)
	require.EqualError(t, err, "error exchanging the authorization code: the token endpoint returned an error response with status code 400: error 'invalid_grant'")
}

func newTestDiscordProvider(base, method string) *DiscordProvider {
	provider := newDiscordProvider(&schema.AuthenticationBackendExternalIdentityProvider{
		ID: "discord", Name: "Discord", ClientID: "123", ClientSecret: "secret",
		Scopes: []string{"identify", "email"}, TokenEndpointAuthMethod: method,
	}, newProviderClient(nil))

	provider.client.RetryMax = 0
	provider.tokenEndpoint = base + "/api/oauth2/token"
	provider.userEndpoint = base + "/api/v10/users/@me"

	return provider
}
