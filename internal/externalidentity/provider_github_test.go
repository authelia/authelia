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

func TestGitHubProviderAuthorizationRequest(t *testing.T) {
	provider := newGitHubProvider(&schema.AuthenticationBackendExternalIdentityProvider{
		ID: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret",
		Scopes: []string{"read:user", "user:email"}, TokenEndpointAuthMethod: "client_secret_basic",
	}, newProviderClient(nil))

	request, err := provider.AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/api/firstfactor/external-identity/github/callback"})
	require.NoError(t, err)

	assert.Len(t, request.State, 43)
	assert.Empty(t, request.Nonce, "GitHub issues no ID Token, so there is no nonce to bind it to")
	assert.Len(t, request.CodeVerifier, 43)

	uri, err := url.Parse(request.URL)
	require.NoError(t, err)

	assert.Equal(t, "https", uri.Scheme)
	assert.Equal(t, "github.com", uri.Host)
	assert.Equal(t, "/login/oauth/authorize", uri.Path)

	query := uri.Query()

	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "Iv1.abc", query.Get("client_id"))
	assert.Equal(t, "read:user user:email", query.Get("scope"))
	assert.Equal(t, "https://auth.example.com/api/firstfactor/external-identity/github/callback", query.Get("redirect_uri"))
	assert.Equal(t, request.State, query.Get("state"))
	assert.Equal(t, "S256", query.Get("code_challenge_method"))
	assert.NotEmpty(t, query.Get("code_challenge"))
	assert.False(t, query.Has("nonce"))
}

func TestGitHubProviderComplete(t *testing.T) {
	testCases := []struct {
		Name         string
		Method       string
		UserStatus   int
		User         any
		EmailsStatus int
		Emails       any
		Expected     *IdentityClaims
		Error        string
	}{
		{
			Name:         "ShouldReturnTheUserWithTheVerifiedPrimaryEmail",
			Method:       "client_secret_basic",
			UserStatus:   http.StatusOK,
			User:         map[string]any{"id": 583231, "login": "octocat", "name": "The Octocat"},
			EmailsStatus: http.StatusOK,
			Emails: []map[string]any{
				{"email": "octocat@users.noreply.github.com", "primary": false, "verified": true},
				{"email": "octocat@github.com", "primary": true, "verified": true},
			},
			Expected: &IdentityClaims{Issuer: "https://github.com", Subject: "583231", PreferredUsername: "octocat", Name: "The Octocat", Email: "octocat@github.com"},
		},
		{
			Name:         "ShouldAuthenticateWithThePostMethod",
			Method:       "client_secret_post",
			UserStatus:   http.StatusOK,
			User:         map[string]any{"id": 583231, "login": "octocat", "name": "The Octocat"},
			EmailsStatus: http.StatusOK,
			Emails:       []map[string]any{},
			Expected:     &IdentityClaims{Issuer: "https://github.com", Subject: "583231", PreferredUsername: "octocat", Name: "The Octocat"},
		},
		{
			Name:         "ShouldUseTheLoginWithoutAName",
			Method:       "client_secret_basic",
			UserStatus:   http.StatusOK,
			User:         map[string]any{"id": 583231, "login": "octocat", "name": nil},
			EmailsStatus: http.StatusOK,
			Emails:       []map[string]any{},
			Expected:     &IdentityClaims{Issuer: "https://github.com", Subject: "583231", PreferredUsername: "octocat", Name: "octocat"},
		},
		{
			Name:         "ShouldOmitAnUnverifiedPrimaryEmail",
			Method:       "client_secret_basic",
			UserStatus:   http.StatusOK,
			User:         map[string]any{"id": 583231, "login": "octocat"},
			EmailsStatus: http.StatusOK,
			Emails:       []map[string]any{{"email": "octocat@github.com", "primary": true, "verified": false}},
			Expected:     &IdentityClaims{Issuer: "https://github.com", Subject: "583231", PreferredUsername: "octocat", Name: "octocat"},
		},
		{
			Name:         "ShouldOmitTheEmailWithoutTheEmailScope",
			Method:       "client_secret_basic",
			UserStatus:   http.StatusOK,
			User:         map[string]any{"id": 583231, "login": "octocat"},
			EmailsStatus: http.StatusNotFound,
			Emails:       map[string]any{"message": "Not Found"},
			Expected:     &IdentityClaims{Issuer: "https://github.com", Subject: "583231", PreferredUsername: "octocat", Name: "octocat"},
		},
		{
			Name:       "ShouldRaiseErrorOnAbsentID",
			Method:     "client_secret_basic",
			UserStatus: http.StatusOK,
			User:       map[string]any{"login": "octocat"},
			Error:      "error requesting the github user: the github response is invalid: the 'id' is required but it is absent",
		},
		{
			Name:       "ShouldRaiseErrorOnErrorStatus",
			Method:     "client_secret_basic",
			UserStatus: http.StatusUnauthorized,
			User:       map[string]any{"message": "Bad credentials"},
			Error:      "error requesting the github user: the github response is invalid: the endpoint returned status code 401",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mux := http.NewServeMux()

			mux.HandleFunc("/login/oauth/access_token", func(rw http.ResponseWriter, r *http.Request) {
				require.NoError(t, r.ParseForm())

				assert.Equal(t, "authorization_code", r.PostForm.Get("grant_type"))
				assert.Equal(t, "the-code", r.PostForm.Get("code"))
				assert.Equal(t, "the-verifier", r.PostForm.Get("code_verifier"))
				assert.Equal(t, "https://auth.example.com/cb", r.PostForm.Get("redirect_uri"))

				switch tc.Method {
				case "client_secret_basic":
					username, password, ok := r.BasicAuth()

					assert.True(t, ok)
					assert.Equal(t, "Iv1.abc", username)
					assert.Equal(t, "secret", password)
				case "client_secret_post":
					assert.Equal(t, "Iv1.abc", r.PostForm.Get("client_id"))
					assert.Equal(t, "secret", r.PostForm.Get("client_secret"))
				}

				rw.Header().Set(headerContentType, mimeApplicationJSON)

				_ = json.NewEncoder(rw).Encode(map[string]any{"access_token": "gho_token", "token_type": "bearer", "scope": "read:user,user:email"})
			})

			mux.HandleFunc("/user", func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "Bearer gho_token", r.Header.Get(headerAuthorization))
				assert.Equal(t, "application/vnd.github+json", r.Header.Get(headerAccept))
				assert.Equal(t, "2022-11-28", r.Header.Get(headerGitHubAPIVersion))

				rw.Header().Set(headerContentType, "application/json; charset=utf-8")
				rw.WriteHeader(tc.UserStatus)

				_ = json.NewEncoder(rw).Encode(tc.User)
			})

			mux.HandleFunc("/user/emails", func(rw http.ResponseWriter, r *http.Request) {
				if tc.Emails == nil {
					t.Errorf("the emails must not be requested")
				}

				assert.Equal(t, "Bearer gho_token", r.Header.Get(headerAuthorization))

				rw.Header().Set(headerContentType, "application/json; charset=utf-8")
				rw.WriteHeader(tc.EmailsStatus)

				_ = json.NewEncoder(rw).Encode(tc.Emails)
			})

			server := httptest.NewServer(mux)

			defer server.Close()

			claims, err := newTestGitHubProvider(server.URL, tc.Method).Complete(context.Background(), CompletionRequest{Code: "the-code", CodeVerifier: "the-verifier", RedirectURI: "https://auth.example.com/cb", Now: time.Now()})

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

func TestGitHubProviderMetadata(t *testing.T) {
	provider := newGitHubProvider(&schema.AuthenticationBackendExternalIdentityProvider{
		ID: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret",
		AuthenticationMethodsReference: schema.AuthenticationBackendExternalIdentityProviderAMR{Default: []string{"pwd"}},
	}, newProviderClient(nil))

	assert.Equal(t, "github", provider.ID())
	assert.Equal(t, "GitHub", provider.Name())
	assert.Equal(t, ProviderTypeGitHub, provider.Type())
	assert.Equal(t, "https://github.com", provider.Issuer())
	assert.Equal(t, ResponseModeQuery, provider.ResponseMode())
	assert.Equal(t, []string{"pwd"}, provider.AuthenticationMethodsReference(nil))
	assert.Equal(t, []string{"pwd", "kba"}, newTestGitHubProvider("https://github.com", "client_secret_basic").AuthenticationMethodsReference(nil))
}

func newTestGitHubProvider(base, method string) *GitHubProvider {
	provider := newGitHubProvider(&schema.AuthenticationBackendExternalIdentityProvider{
		ID: "github", Name: "GitHub", ClientID: "Iv1.abc", ClientSecret: "secret",
		Scopes: []string{"read:user", "user:email"}, TokenEndpointAuthMethod: method,
	}, newProviderClient(nil))

	provider.client.RetryMax = 0
	provider.tokenEndpoint = base + "/login/oauth/access_token"
	provider.userEndpoint = base + "/user"
	provider.emailsEndpoint = base + "/user/emails"

	return provider
}
