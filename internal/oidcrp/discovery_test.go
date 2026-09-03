package oidcrp

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
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var server *httptest.Server

			server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, pathWellKnownOpenIDConfiguration, r.URL.Path)

				rw.Header().Set(headerContentType, mimeApplicationJSON)
				rw.WriteHeader(tc.Status)

				_, _ = rw.Write([]byte(formatDiscoveryBody(tc.Body, server.URL)))
			}))

			defer server.Close()

			client := retryablehttp.NewClient()
			client.Logger = nil
			client.RetryMax = 0

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
		})
	}
}

func formatDiscoveryBody(body, url string) string {
	return strings.NewReplacer("%s", url, "%[2]s", url).Replace(body)
}
