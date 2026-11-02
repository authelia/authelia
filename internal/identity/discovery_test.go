// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package identity

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiscover(t *testing.T) {
	testCases := []struct {
		Name     string
		Body     string
		Status   int
		Expected *Discovery
		Error    string
	}{
		{
			Name:   "ShouldDiscoverValidDocument",
			Status: http.StatusOK,
			Body:   `{"issuer":"%s","authorization_endpoint":"%s/authorize","token_endpoint":"%s/token","jwks_uri":"%s/jwks.json"}`,
			Expected: &Discovery{
				AuthorizationEndpoint: "/authorize",
				TokenEndpoint:         "/token",
				JWKSURI:               "/jwks.json",
			},
		},
		{
			Name:   "ShouldDiscoverAuthorizationResponseIssParameterSupport",
			Status: http.StatusOK,
			Body:   `{"issuer":"%s","authorization_endpoint":"%s/authorize","token_endpoint":"%s/token","jwks_uri":"%s/jwks.json","authorization_response_iss_parameter_supported":true}`,
			Expected: &Discovery{
				AuthorizationEndpoint: "/authorize",
				TokenEndpoint:         "/token",
				JWKSURI:               "/jwks.json",

				AuthorizationResponseIssParameterSupported: true,
			},
		},
		{
			Name:   "ShouldRaiseErrorOnIssuerMismatch",
			Status: http.StatusOK,
			Body:   `{"issuer":"https://elsewhere.example.com","authorization_endpoint":"%[2]s/authorize","token_endpoint":"%[2]s/token","jwks_uri":"%[2]s/jwks.json"}`,
			Error:  "error discovering the provider: the discovery document issuer does not match the configured issuer",
		},
		{
			Name:   "ShouldRaiseErrorOnMissingTokenEndpoint",
			Status: http.StatusOK,
			Body:   `{"issuer":"%s","authorization_endpoint":"%s/authorize","jwks_uri":"%[2]s/jwks.json"}`,
			Error:  "error discovering the provider: the discovery document is missing a required endpoint",
		},
		{
			Name:   "ShouldRaiseErrorOnNonOKStatus",
			Status: http.StatusNotFound,
			Body:   `{}`,
			Error:  "error discovering the provider: the discovery endpoint returned status code 404",
		},
		{
			Name:   "ShouldRaiseErrorOnInsecureAuthorizationEndpoint",
			Status: http.StatusOK,
			Body:   `{"issuer":"%s","authorization_endpoint":"http://op.example.com/authorize","token_endpoint":"%[2]s/token","jwks_uri":"%[2]s/jwks.json"}`,
			Error:  "error discovering the provider: the discovery document includes a url which does not use the https scheme: the 'authorization_endpoint' is 'http://op.example.com/authorize'",
		},
		{
			Name:   "ShouldRaiseErrorOnInsecureTokenEndpoint",
			Status: http.StatusOK,
			Body:   `{"issuer":"%s","authorization_endpoint":"%[2]s/authorize","token_endpoint":"http://op.example.com/token","jwks_uri":"%[2]s/jwks.json"}`,
			Error:  "error discovering the provider: the discovery document includes a url which does not use the https scheme: the 'token_endpoint' is 'http://op.example.com/token'",
		},
		{
			Name:   "ShouldRaiseErrorOnInsecureUserInfoEndpoint",
			Status: http.StatusOK,
			Body:   `{"issuer":"%s","authorization_endpoint":"%[2]s/authorize","token_endpoint":"%[2]s/token","userinfo_endpoint":"http://op.example.com/userinfo","jwks_uri":"%[2]s/jwks.json"}`,
			Error:  "error discovering the provider: the discovery document includes a url which does not use the https scheme: the 'userinfo_endpoint' is 'http://op.example.com/userinfo'",
		},
		{
			Name:   "ShouldRaiseErrorOnInsecureJSONWebKeySetURI",
			Status: http.StatusOK,
			Body:   `{"issuer":"%s","authorization_endpoint":"%[2]s/authorize","token_endpoint":"%[2]s/token","jwks_uri":"http://op.example.com/jwks.json"}`,
			Error:  "error discovering the provider: the discovery document includes a url which does not use the https scheme: the 'jwks_uri' is 'http://op.example.com/jwks.json'",
		},
		{
			Name:   "ShouldRaiseErrorOnInsecurePushedAuthorizationRequestEndpoint",
			Status: http.StatusOK,
			Body:   `{"issuer":"%s","authorization_endpoint":"%[2]s/authorize","token_endpoint":"%[2]s/token","jwks_uri":"%[2]s/jwks.json","pushed_authorization_request_endpoint":"http://op.example.com/par"}`,
			Error:  "error discovering the provider: the discovery document includes a url which does not use the https scheme: the 'pushed_authorization_request_endpoint' is 'http://op.example.com/par'",
		},
		{
			Name:   "ShouldRaiseErrorOnRelativeEndpoint",
			Status: http.StatusOK,
			Body:   `{"issuer":"%s","authorization_endpoint":"/authorize","token_endpoint":"%[2]s/token","jwks_uri":"%[2]s/jwks.json"}`,
			Error:  "error discovering the provider: the discovery document includes a url which does not use the https scheme: the 'authorization_endpoint' is '/authorize'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var server *httptest.Server

			server = httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, pathWellKnownOpenIDConfiguration, r.URL.Path)

				rw.Header().Set(headerContentType, mimeApplicationJSON)
				rw.WriteHeader(tc.Status)

				_, _ = rw.Write([]byte(formatDiscoveryBody(tc.Body, server.URL)))
			}))

			defer server.Close()

			client := retryablehttp.NewClient()
			client.Logger = nil
			client.RetryMax = 0
			client.HTTPClient = server.Client()

			discovery, err := Discover(context.Background(), client, server.URL)

			if tc.Error != "" {
				assert.Nil(t, discovery)
				require.EqualError(t, err, tc.Error)

				return
			}

			require.NoError(t, err)
			require.NotNil(t, discovery)
			assert.Equal(t, server.URL, discovery.Issuer)
			assert.Equal(t, server.URL+tc.Expected.AuthorizationEndpoint, discovery.AuthorizationEndpoint)
			assert.Equal(t, server.URL+tc.Expected.TokenEndpoint, discovery.TokenEndpoint)
			assert.Equal(t, server.URL+tc.Expected.JWKSURI, discovery.JWKSURI)
			assert.Equal(t, tc.Expected.AuthorizationResponseIssParameterSupported, discovery.AuthorizationResponseIssParameterSupported)
		})
	}
}

func TestDiscoverShouldRaiseErrorOnInsecureIssuer(t *testing.T) {
	var (
		server   *httptest.Server
		requests int
	)

	server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requests++

		rw.Header().Set(headerContentType, mimeApplicationJSON)

		_, _ = rw.Write([]byte(formatDiscoveryBody(`{"issuer":"%s","authorization_endpoint":"https://op.example.com/authorize","token_endpoint":"https://op.example.com/token","jwks_uri":"https://op.example.com/jwks.json"}`, server.URL)))
	}))

	defer server.Close()

	client := retryablehttp.NewClient()
	client.Logger = nil
	client.RetryMax = 0

	discovery, err := Discover(context.Background(), client, server.URL)

	assert.Nil(t, discovery)
	require.EqualError(t, err, "error discovering the provider: the discovery document includes a url which does not use the https scheme: the 'issuer' is '"+server.URL+"'")
	assert.ErrorIs(t, err, ErrDiscoveryURLInsecure)
	assert.Equal(t, 0, requests)
}

func formatDiscoveryBody(body, url string) string {
	return strings.NewReplacer("%s", url, "%[2]s", url).Replace(body)
}
