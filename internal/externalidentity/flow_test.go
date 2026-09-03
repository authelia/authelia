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

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/random"
)

func TestProviderAuthorizationRequest(t *testing.T) {
	provider := &OpenIDConnectProvider{
		id: "example", name: "Example", issuer: "https://op.example.com", clientID: "client",
		scopes: []string{"openid", "email"}, algIDToken: "RS256",
		endpointAuthorization: "https://op.example.com/authorize",
	}

	request, err := provider.AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/api/firstfactor/external-identity/example/callback"})

	require.NoError(t, err)
	require.NotNil(t, request)

	assert.Len(t, request.State, 43)
	assert.Len(t, request.Nonce, 43)
	assert.Len(t, request.CodeVerifier, 43)

	uri, err := url.Parse(request.URL)
	require.NoError(t, err)

	assert.Equal(t, "op.example.com", uri.Host)
	assert.Equal(t, "/authorize", uri.Path)

	query := uri.Query()

	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "client", query.Get("client_id"))
	assert.Equal(t, "openid email", query.Get("scope"))
	assert.Equal(t, "https://auth.example.com/api/firstfactor/external-identity/example/callback", query.Get("redirect_uri"))
	assert.Equal(t, request.State, query.Get("state"))
	assert.Equal(t, request.Nonce, query.Get("nonce"))
	assert.Equal(t, "S256", query.Get("code_challenge_method"))
	assert.NotEmpty(t, query.Get("code_challenge"))
	assert.NotEqual(t, request.CodeVerifier, query.Get("code_challenge"))
}

func TestProviderAuthorizationRequestUILocales(t *testing.T) {
	testCases := []struct {
		Name     string
		Language string
		Expected string
		Present  bool
	}{
		{Name: "ShouldRequestTheChosenLanguage", Language: "fr-CA", Expected: "fr-CA", Present: true},
		{Name: "ShouldNotRequestALanguageWhenNoneWasChosen"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := &OpenIDConnectProvider{
				id: "example", name: "Example", issuer: "https://op.example.com", clientID: "client",
				scopes: []string{"openid"}, algIDToken: "RS256",
				endpointAuthorization: "https://op.example.com/authorize",
			}

			request, err := provider.AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/cb", Language: tc.Language})
			require.NoError(t, err)

			uri, err := url.Parse(request.URL)
			require.NoError(t, err)

			assert.Equal(t, tc.Present, uri.Query().Has("ui_locales"))
			assert.Equal(t, tc.Expected, uri.Query().Get("ui_locales"))
		})
	}
}

func TestProviderExchange(t *testing.T) {
	testCases := []struct {
		Name     string
		Status   int
		Body     any
		Method   string
		Expected string
		Error    string
	}{
		{
			Name:     "ShouldReturnIDTokenWithBasicAuth",
			Status:   http.StatusOK,
			Body:     map[string]any{"access_token": "at", "token_type": "Bearer", "id_token": "the-id-token"},
			Method:   "client_secret_basic",
			Expected: "the-id-token",
		},
		{
			Name:     "ShouldReturnIDTokenWithPostAuth",
			Status:   http.StatusOK,
			Body:     map[string]any{"access_token": "at", "token_type": "Bearer", "id_token": "the-id-token"},
			Method:   "client_secret_post",
			Expected: "the-id-token",
		},
		{
			Name:     "ShouldReturnIDTokenWithNoneAuth",
			Status:   http.StatusOK,
			Body:     map[string]any{"access_token": "at", "token_type": "Bearer", "id_token": "the-id-token"},
			Method:   "none",
			Expected: "the-id-token",
		},
		{
			Name:   "ShouldRaiseErrorOnMissingIDToken",
			Status: http.StatusOK,
			Body:   map[string]any{"access_token": "at", "token_type": "Bearer"},
			Method: "client_secret_basic",
			Error:  "error exchanging the authorization code: the token response did not contain an id token",
		},
		{
			Name:   "ShouldRaiseErrorOnErrorResponse",
			Status: http.StatusBadRequest,
			Body:   map[string]any{"error": "invalid_grant"},
			Method: "client_secret_basic",
			Error:  "error exchanging the authorization code: the token endpoint returned an error response with status code 400: error 'invalid_grant'",
		},
		{
			Name:   "ShouldIncludeTheDescriptionAndHintOfErrorResponse",
			Status: http.StatusBadRequest,
			Body:   map[string]any{"error": "invalid_grant", "error_description": "The provided authorization grant is invalid.", "error_hint": "The PKCE code verifier does not match the challenge."},
			Method: "client_secret_basic",
			Error:  "error exchanging the authorization code: the token endpoint returned an error response with status code 400: error 'invalid_grant', description 'The provided authorization grant is invalid.', hint 'The PKCE code verifier does not match the challenge.'",
		},
		{
			Name:   "ShouldSanitizeErrorResponse",
			Status: http.StatusUnauthorized,
			Body:   map[string]any{"error": "invalid_client", "error_description": "bad\nclient\r\tsecret"},
			Method: "client_secret_basic",
			Error:  "error exchanging the authorization code: the token endpoint returned an error response with status code 401: error 'invalid_client', description 'badclientsecret'",
		},
		{
			Name:   "ShouldNotIncludeTheBodyOfNonOAuthErrorResponse",
			Status: http.StatusForbidden,
			Body:   "<html>access denied by the web application firewall</html>",
			Method: "client_secret_basic",
			Error:  "error exchanging the authorization code: the token endpoint returned an error response with status code 403",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				require.NoError(t, r.ParseForm())

				assert.Equal(t, "authorization_code", r.PostForm.Get("grant_type"))
				assert.Equal(t, "the-code", r.PostForm.Get("code"))
				assert.Equal(t, "the-verifier", r.PostForm.Get("code_verifier"))
				assert.Equal(t, "https://auth.example.com/cb", r.PostForm.Get("redirect_uri"))

				switch tc.Method {
				case "client_secret_basic":
					username, password, ok := r.BasicAuth()

					assert.True(t, ok)
					assert.Equal(t, "client", username)
					assert.Equal(t, "secret", password)
					assert.Empty(t, r.PostForm.Get("client_secret"))
				case "client_secret_post":
					assert.Equal(t, "client", r.PostForm.Get("client_id"))
					assert.Equal(t, "secret", r.PostForm.Get("client_secret"))
				case "none":
					assert.Equal(t, "client", r.PostForm.Get("client_id"))
					assert.Empty(t, r.PostForm.Get("client_secret"))

					_, _, ok := r.BasicAuth()

					assert.False(t, ok)
				}

				rw.Header().Set(headerContentType, mimeApplicationJSON)
				rw.WriteHeader(tc.Status)

				_ = json.NewEncoder(rw).Encode(tc.Body)
			}))

			defer server.Close()

			provider := newTestProvider(server.URL, tc.Method)

			token, err := provider.Exchange(context.Background(), "the-code", "the-verifier", "https://auth.example.com/cb")

			if tc.Error != "" {
				assert.Nil(t, token)
				require.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, token)
			assert.Equal(t, tc.Expected, token.IDToken)
			assert.Equal(t, "at", token.AccessToken)
		})
	}
}

func newTestProvider(tokenEndpoint, method string) *OpenIDConnectProvider {
	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryMax = 0

	return &OpenIDConnectProvider{
		id: "example", name: "Example", issuer: "https://op.example.com", clientID: "client",
		clientSecret: "secret", tokenEndpointAuthMethod: method, algIDToken: "RS256",
		endpointToken: tokenEndpoint, client: client,
	}
}

func TestProviderAuthorizationRequestResponseMode(t *testing.T) {
	testCases := []struct {
		Name     string
		Mode     string
		Expected string
	}{
		{Name: "ShouldOmitTheQueryResponseMode", Mode: ResponseModeQuery, Expected: ""},
		{Name: "ShouldRequestTheFormPostResponseMode", Mode: ResponseModeFormPost, Expected: ResponseModeFormPost},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			provider := &OpenIDConnectProvider{
				id: "example", name: "Example", issuer: "https://op.example.com", clientID: "client",
				scopes: []string{"openid"}, responseMode: tc.Mode, algIDToken: "RS256",
				endpointAuthorization: "https://op.example.com/authorize",
			}

			request, err := provider.AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/cb"})
			require.NoError(t, err)

			uri, err := url.Parse(request.URL)
			require.NoError(t, err)

			assert.Equal(t, tc.Expected, uri.Query().Get("response_mode"))
		})
	}
}
