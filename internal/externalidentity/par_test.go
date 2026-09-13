// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/random"
)

func TestProviderAuthorizationRequestPushed(t *testing.T) {
	testCases := []struct {
		Name     string
		Method   string
		Status   int
		Body     string
		Endpoint bool
		Error    string
	}{
		{Name: "ShouldPushWithBasicAuth", Method: "client_secret_basic", Status: http.StatusCreated, Body: `{"request_uri":"urn:ietf:params:oauth:request_uri:abc","expires_in":60}`, Endpoint: true},
		{Name: "ShouldPushWithPostAuth", Method: "client_secret_post", Status: http.StatusCreated, Body: `{"request_uri":"urn:ietf:params:oauth:request_uri:abc","expires_in":60}`, Endpoint: true},
		{Name: "ShouldPushWithNoneAuth", Method: "none", Status: http.StatusCreated, Body: `{"request_uri":"urn:ietf:params:oauth:request_uri:abc","expires_in":60}`, Endpoint: true},
		{Name: "ShouldAcceptOKStatus", Method: "client_secret_basic", Status: http.StatusOK, Body: `{"request_uri":"urn:ietf:params:oauth:request_uri:abc","expires_in":60}`, Endpoint: true},
		{
			Name:     "ShouldRaiseErrorOnErrorResponse",
			Method:   "client_secret_basic",
			Status:   http.StatusBadRequest,
			Body:     `{"error":"invalid_request","error_description":"The request is\nmissing a parameter.","error_hint":"The 'redirect_uri' is not registered."}`,
			Endpoint: true,
			Error:    "error pushing the authorization request: the pushed authorization request endpoint returned an error response with status code 400: error 'invalid_request', description 'The request ismissing a parameter.', hint 'The 'redirect_uri' is not registered.'",
		},
		{
			Name:     "ShouldRaiseErrorOnErrorResponseWithoutJSON",
			Method:   "client_secret_basic",
			Status:   http.StatusForbidden,
			Body:     `<html>forbidden</html>`,
			Endpoint: true,
			Error:    "error pushing the authorization request: the pushed authorization request endpoint returned an error response with status code 403",
		},
		{
			Name:     "ShouldRaiseErrorOnAbsentRequestURI",
			Method:   "client_secret_basic",
			Status:   http.StatusCreated,
			Body:     `{"expires_in":60}`,
			Endpoint: true,
			Error:    "error pushing the authorization request: the pushed authorization response is invalid: the 'request_uri' is absent",
		},
		{
			Name:   "ShouldRaiseErrorWhenTheEndpointIsNotKnown",
			Method: "client_secret_basic",
			Error:  "error pushing the authorization request: the pushed authorization request endpoint is not known",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				form     url.Values
				requests int
			)

			server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				requests++

				assert.Equal(t, http.MethodPost, r.Method)
				assert.Equal(t, "/par", r.URL.Path)
				assert.Equal(t, mimeApplicationXWWWFormURLEncoded, r.Header.Get(headerContentType))

				require.NoError(t, r.ParseForm())

				form = r.PostForm

				username, password, basic := r.BasicAuth()

				switch tc.Method {
				case "client_secret_basic":
					assert.True(t, basic)
					assert.Equal(t, "client", username)
					assert.Equal(t, "secret", password)
					assert.False(t, r.PostForm.Has("client_secret"))
				case "client_secret_post":
					assert.False(t, basic)
					assert.Equal(t, "secret", r.PostForm.Get("client_secret"))
				case "none":
					assert.False(t, basic)
					assert.False(t, r.PostForm.Has("client_secret"))
				}

				rw.Header().Set(headerContentType, mimeApplicationJSON)
				rw.WriteHeader(tc.Status)

				_, _ = rw.Write([]byte(tc.Body))
			}))

			defer server.Close()

			client := retryablehttp.NewClient()
			client.Logger = nil
			client.RetryMax = 0

			provider := &OpenIDConnectProvider{
				id: "example", name: "Example", issuer: "https://op.example.com", clientID: "client",
				clientSecret: "secret", tokenEndpointAuthMethod: tc.Method, algIDToken: "RS256",
				scopes: []string{"openid", "email"}, endpointAuthorization: "https://op.example.com/authorize",
				parRequired: true, client: client,
			}

			if tc.Endpoint {
				provider.endpointPAR = server.URL + "/par"
			}

			request, err := provider.AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/cb", Language: "fr"})

			if tc.Error != "" {
				assert.Nil(t, request)
				require.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, request)
			require.Equal(t, 1, requests)

			assert.Equal(t, "https://op.example.com/authorize?client_id=client&request_uri=urn%3Aietf%3Aparams%3Aoauth%3Arequest_uri%3Aabc", request.URL,
				"the browser must only be sent the client identifier and the request URI")

			assert.Equal(t, "code", form.Get("response_type"))
			assert.Equal(t, "client", form.Get("client_id"))
			assert.Equal(t, "https://auth.example.com/cb", form.Get("redirect_uri"))
			assert.Equal(t, "openid email", form.Get("scope"))
			assert.Equal(t, request.State, form.Get("state"))
			assert.Equal(t, request.Nonce, form.Get("nonce"))
			assert.Equal(t, "S256", form.Get("code_challenge_method"))
			assert.NotEmpty(t, form.Get("code_challenge"))
			assert.Equal(t, "fr", form.Get("ui_locales"))
		})
	}
}

func TestProviderAuthorizationRequestShouldNotPushUnlessRequired(t *testing.T) {
	var requests int

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requests++
	}))

	defer server.Close()

	provider := &OpenIDConnectProvider{
		id: "example", name: "Example", issuer: "https://op.example.com", clientID: "client",
		clientSecret: "secret", tokenEndpointAuthMethod: "client_secret_basic", algIDToken: "RS256",
		scopes: []string{"openid"}, endpointAuthorization: "https://op.example.com/authorize",
		endpointPAR: server.URL + "/par",
	}

	request, err := provider.AuthorizationRequest(context.Background(), &random.Cryptographical{}, AuthorizationRequestOptions{RedirectURI: "https://auth.example.com/cb"})
	require.NoError(t, err)

	uri, err := url.Parse(request.URL)
	require.NoError(t, err)

	assert.Equal(t, 0, requests, "an advertised endpoint alone must not cause the authorization request to be pushed")
	assert.Equal(t, request.State, uri.Query().Get("state"))
	assert.False(t, uri.Query().Has("request_uri"))
}

func TestProviderResolveShouldResolvePushedAuthorizationRequests(t *testing.T) {
	testCases := []struct {
		Name             string
		Configured       bool
		ConfiguredURI    string
		Metadata         string
		ExpectedRequired bool
		ExpectedEndpoint string
		Error            string
	}{
		{
			Name: "ShouldNotRequireWhenNothingIsAdvertised",
		},
		{
			Name:             "ShouldStoreTheAdvertisedEndpointWithoutRequiringIt",
			Metadata:         `,"pushed_authorization_request_endpoint":"https://op.example.com/par"`,
			ExpectedEndpoint: "https://op.example.com/par",
		},
		{
			Name:             "ShouldRequireWhenTheProviderRequiresIt",
			Metadata:         `,"pushed_authorization_request_endpoint":"https://op.example.com/par","require_pushed_authorization_requests":true`,
			ExpectedRequired: true,
			ExpectedEndpoint: "https://op.example.com/par",
		},
		{
			Name:             "ShouldRequireWhenConfigured",
			Configured:       true,
			Metadata:         `,"pushed_authorization_request_endpoint":"https://op.example.com/par"`,
			ExpectedRequired: true,
			ExpectedEndpoint: "https://op.example.com/par",
		},
		{
			Name:             "ShouldPreferTheConfiguredEndpoint",
			Configured:       true,
			ConfiguredURI:    "https://op.example.com/configured/par",
			Metadata:         `,"pushed_authorization_request_endpoint":"https://op.example.com/par"`,
			ExpectedRequired: true,
			ExpectedEndpoint: "https://op.example.com/configured/par",
		},
		{
			Name:       "ShouldRaiseErrorWhenRequiredWithoutAnEndpoint",
			Configured: true,
			Error:      "error resolving provider 'example': the discovery document is missing a required endpoint: pushed authorization requests are required but the discovery document does not include the 'pushed_authorization_request_endpoint'",
		},
		{
			Name:     "ShouldRaiseErrorWhenTheProviderRequiresItWithoutAnEndpoint",
			Metadata: `,"require_pushed_authorization_requests":true`,
			Error:    "error resolving provider 'example': the discovery document is missing a required endpoint: pushed authorization requests are required but the discovery document does not include the 'pushed_authorization_request_endpoint'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var server *httptest.Server

			server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				rw.Header().Set(headerContentType, mimeApplicationJSON)

				_, _ = rw.Write([]byte(`{"issuer":"` + server.URL + `","authorization_endpoint":"` + server.URL + `/authorize","token_endpoint":"` + server.URL + `/token","jwks_uri":"` + server.URL + `/jwks.json"` + tc.Metadata + `}`))
			}))

			defer server.Close()

			provider, ok := getOpenIDConnectProvider(NewProviders(&schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "example", Name: "Example", Issuer: server.URL,
						ClientID: "client", ClientSecret: "secret",
						PKCE:                               schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
						RequirePushedAuthorizationRequests: tc.Configured,
						Endpoints:                          schema.AuthenticationBackendExternalIdentityProviderEndpoints{PushedAuthorizationRequest: tc.ConfiguredURI},
					},
				},
			}, nil))

			require.True(t, ok)

			err := provider.Resolve(context.Background())

			if tc.Error != "" {
				require.EqualError(t, err, tc.Error)
				assert.ErrorIs(t, err, ErrDiscoveryEndpointMissing)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.ExpectedRequired, provider.parRequired)
			assert.Equal(t, tc.ExpectedEndpoint, provider.endpointPAR)
		})
	}
}
