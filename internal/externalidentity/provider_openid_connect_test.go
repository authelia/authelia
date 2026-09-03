// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package externalidentity

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"authelia.com/provider/oauth2/token/jose"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestProviderResolveShouldSelectTheIDTokenSigningAlg(t *testing.T) {
	testCases := []struct {
		Name       string
		Configured string
		Advertised string
		Disable    bool
		Expected   string
		Requests   int
		Error      string
	}{
		{Name: "ShouldDefaultToRS256WhenNothingIsAdvertised", Expected: "RS256", Requests: 1},
		{Name: "ShouldPreferRS256WhenAdvertised", Advertised: `["ES256","RS256"]`, Expected: "RS256", Requests: 1},
		{Name: "ShouldSelectTheFirstSupportedAdvertisedAlg", Advertised: `["HS256","ES384","PS256"]`, Expected: "ES384", Requests: 1},
		{Name: "ShouldUseTheConfiguredAlg", Configured: "PS512", Advertised: `["RS256","PS512"]`, Expected: "PS512", Requests: 1},
		{Name: "ShouldDefaultToRS256WithDiscoveryDisabled", Disable: true, Expected: "RS256"},
		{
			Name:       "ShouldRaiseErrorWhenNoAdvertisedAlgIsSupported",
			Advertised: `["HS256","EdDSA"]`,
			Requests:   1,
			Error:      "error resolving provider 'example': error validating the discovery document: the discovery document does not advertise an id token signing algorithm which is supported: the 'id_token_signing_alg_values_supported' values are 'HS256', 'EdDSA'",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				server   *httptest.Server
				requests int
			)

			server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				requests++

				metadata := ""

				if tc.Advertised != "" {
					metadata = `,"id_token_signing_alg_values_supported":` + tc.Advertised
				}

				rw.Header().Set(headerContentType, mimeApplicationJSON)

				_, _ = rw.Write([]byte(`{"issuer":"` + server.URL + `","authorization_endpoint":"` + server.URL + `/authorize","token_endpoint":"` + server.URL + `/token","jwks_uri":"` + server.URL + `/jwks.json"` + metadata + `}`))
			}))

			defer server.Close()

			config := schema.AuthenticationBackendExternalIdentityProvider{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				IDTokenSignedResponseAlg: tc.Configured,
				PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
				Discovery:                schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: tc.Disable},
			}

			if tc.Disable {
				config.Endpoints = schema.AuthenticationBackendExternalIdentityProviderEndpoints{
					Authorization: server.URL + "/authorize",
					Token:         server.URL + "/token",
					JSONWebKeys:   server.URL + "/jwks.json",
				}
			}

			provider, ok := getOpenIDConnectProvider(NewProviders(&schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{config},
			}, nil))

			require.True(t, ok)

			err := provider.Resolve(context.Background())

			assert.Equal(t, tc.Requests, requests)

			if tc.Error != "" {
				require.EqualError(t, err, tc.Error)
				assert.ErrorIs(t, err, ErrDiscoveryNoSupportedAlg)
				assert.Empty(t, provider.algIDToken, "a provider which failed to resolve must not have an algorithm selected")

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, provider.algIDToken)
		})
	}
}

func TestProviderResolveShouldValidateDiscoveryCapabilities(t *testing.T) {
	testCases := []struct {
		Name         string
		Metadata     string
		Alg          string
		AuthMethod   string
		ResponseMode string
		Error        string
	}{
		{
			Name:       "ShouldAcceptOmittedCapabilities",
			Alg:        "ES256",
			AuthMethod: "client_secret_post",
		},
		{
			Name:         "ShouldAcceptAdvertisedCapabilities",
			Metadata:     `,"id_token_signing_alg_values_supported":["RS256","ES256"],"code_challenge_methods_supported":["plain","S256"],"token_endpoint_auth_methods_supported":["client_secret_basic","client_secret_post"],"response_modes_supported":["query","form_post"]`,
			Alg:          "ES256",
			AuthMethod:   "client_secret_post",
			ResponseMode: "form_post",
		},
		{
			Name:     "ShouldRaiseErrorOnUnadvertisedAlg",
			Metadata: `,"id_token_signing_alg_values_supported":["RS256"]`,
			Alg:      "ES256",
			Error:    "error resolving provider 'example': error validating the discovery document: the discovery document does not advertise support for the configured value: the value 'ES256' is not included in the 'id_token_signing_alg_values_supported' values 'RS256'",
		},
		{
			Name:     "ShouldRaiseErrorOnUnadvertisedCodeChallengeMethod",
			Metadata: `,"code_challenge_methods_supported":["plain"]`,
			Alg:      "RS256",
			Error:    "error resolving provider 'example': error validating the discovery document: the discovery document does not advertise support for the configured value: the value 'S256' is not included in the 'code_challenge_methods_supported' values 'plain'",
		},
		{
			Name:       "ShouldRaiseErrorOnUnadvertisedTokenEndpointAuthMethod",
			Metadata:   `,"token_endpoint_auth_methods_supported":["client_secret_basic","private_key_jwt"]`,
			Alg:        "RS256",
			AuthMethod: "client_secret_post",
			Error:      "error resolving provider 'example': error validating the discovery document: the discovery document does not advertise support for the configured value: the value 'client_secret_post' is not included in the 'token_endpoint_auth_methods_supported' values 'client_secret_basic', 'private_key_jwt'",
		},
		{
			Name:         "ShouldRaiseErrorOnUnadvertisedFormPostResponseMode",
			Metadata:     `,"response_modes_supported":["query","fragment"]`,
			Alg:          "RS256",
			ResponseMode: "form_post",
			Error:        "error resolving provider 'example': error validating the discovery document: the discovery document does not advertise support for the configured value: the value 'form_post' is not included in the 'response_modes_supported' values 'query', 'fragment'",
		},
		{
			Name:         "ShouldNotRequireTheQueryResponseModeToBeAdvertised",
			Metadata:     `,"response_modes_supported":["form_post"]`,
			Alg:          "RS256",
			ResponseMode: "query",
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
						IDTokenSignedResponseAlg: tc.Alg,
						TokenEndpointAuthMethod:  tc.AuthMethod,
						ResponseMode:             tc.ResponseMode,
						PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
					},
				},
			}, nil))

			require.True(t, ok)

			err := provider.Resolve(context.Background())

			if tc.Error != "" {
				require.EqualError(t, err, tc.Error)
				assert.ErrorIs(t, err, ErrDiscoveryUnsupported)

				return
			}

			require.NoError(t, err)
		})
	}
}

func TestProviderAuthorizationResponseIssuerRequired(t *testing.T) {
	testCases := []struct {
		Name       string
		Configured bool
		Advertised bool
		Disable    bool
		Status     int
		Expected   bool
		Requests   int
		Error      string
	}{
		{Name: "ShouldNotRequireWhenNeitherConfiguredNorAdvertised", Status: http.StatusOK, Requests: 1},
		{Name: "ShouldRequireWhenAdvertised", Advertised: true, Status: http.StatusOK, Expected: true, Requests: 1},
		{Name: "ShouldRequireWhenConfigured", Configured: true, Status: http.StatusOK, Expected: true, Requests: 1},
		{Name: "ShouldRequireWhenConfiguredAndAdvertised", Configured: true, Advertised: true, Status: http.StatusOK, Expected: true, Requests: 1},
		{Name: "ShouldRequireWhenConfiguredWithDiscoveryDisabled", Configured: true, Disable: true, Status: http.StatusOK, Expected: true},
		{Name: "ShouldNotRequireWithDiscoveryDisabled", Disable: true, Status: http.StatusOK},
		{Name: "ShouldRaiseErrorWhenDiscoveryFails", Configured: true, Status: http.StatusNotFound, Requests: 1, Error: "error resolving provider 'example': error discovering the provider: the discovery endpoint returned status code 404"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			var (
				server   *httptest.Server
				requests int
			)

			server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
				requests++

				rw.Header().Set(headerContentType, mimeApplicationJSON)
				rw.WriteHeader(tc.Status)

				_, _ = rw.Write([]byte(`{"issuer":"` + server.URL + `","authorization_endpoint":"` + server.URL + `/authorize","token_endpoint":"` + server.URL + `/token","jwks_uri":"` + server.URL + `/jwks.json","authorization_response_iss_parameter_supported":` + strconv.FormatBool(tc.Advertised) + `}`))
			}))

			defer server.Close()

			config := schema.AuthenticationBackendExternalIdentityProvider{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
				Discovery:                schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: tc.Disable},

				AuthorizationResponseIssParameterSupported: tc.Configured,
			}

			if tc.Disable {
				config.Endpoints = schema.AuthenticationBackendExternalIdentityProviderEndpoints{
					Authorization: server.URL + "/authorize",
					Token:         server.URL + "/token",
					JSONWebKeys:   server.URL + "/jwks.json",
				}
			}

			provider, ok := getOpenIDConnectProvider(NewProviders(&schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{config},
			}, nil))

			require.True(t, ok)

			required, err := provider.AuthorizationResponseIssuerRequired(context.Background())

			if tc.Error != "" {
				require.EqualError(t, err, tc.Error)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tc.Expected, required)
			assert.Equal(t, tc.Requests, requests)
		})
	}
}

//nolint:gosec // Test Credentials.
func TestNewProviders(t *testing.T) {
	testCases := []struct {
		Name   string
		Have   *schema.AuthenticationBackendExternalIdentity
		Assert func(t *testing.T, providers *Providers)
	}{
		{
			Name: "ShouldBuildProviderWithDiscoveryDisabled",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "example", Name: "Example", Issuer: "https://op.example.com",
						ClientID: "client", ClientSecret: "secret",
						Scopes:                   []string{"openid", "email"},
						TokenEndpointAuthMethod:  "client_secret_basic",
						IDTokenSignedResponseAlg: "RS256",
						PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
						Discovery:                schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true},
						Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{
							Authorization: "https://op.example.com/authorize",
							Token:         "https://op.example.com/token",
							JSONWebKeys:   "https://op.example.com/jwks.json",
						},
					},
				},
			},
			Assert: func(t *testing.T, providers *Providers) {
				provider, ok := getOpenIDConnectProvider(providers)

				require.True(t, ok)
				assert.Equal(t, "Example", provider.Name())
				assert.Equal(t, "https://op.example.com/authorize", provider.AuthorizationEndpoint())
				assert.Len(t, providers.All(), 1)
			},
		},
		{
			Name: "ShouldReturnNotOkForUnknownProvider",
			Have: &schema.AuthenticationBackendExternalIdentity{
				Providers: []schema.AuthenticationBackendExternalIdentityProvider{
					{
						ID: "example", Name: "Example", Issuer: "https://op.example.com",
						ClientID: "client", ClientSecret: "secret",
						IDTokenSignedResponseAlg: "RS256",
						PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
						Discovery:                schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true},
						Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{
							Authorization: "https://op.example.com/authorize",
							Token:         "https://op.example.com/token",
							JSONWebKeys:   "https://op.example.com/jwks.json",
						},
					},
				},
			},
			Assert: func(t *testing.T, providers *Providers) {
				provider, ok := providers.Get("missing")

				assert.False(t, ok)
				assert.Nil(t, provider)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			providers := NewProviders(tc.Have, nil)

			if tc.Assert != nil {
				tc.Assert(t, providers)
			}
		})
	}
}

func TestNewProvidersShouldNotPerformDiscoveryDuringConstruction(t *testing.T) {
	var requests int

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requests++

		rw.WriteHeader(http.StatusNotFound)
	}))

	defer server.Close()

	providers := NewProviders(&schema.AuthenticationBackendExternalIdentity{
		Providers: []schema.AuthenticationBackendExternalIdentityProvider{
			{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
			},
		},
	}, nil)

	require.NotNil(t, providers)
	assert.Equal(t, 0, requests)

	provider, ok := getOpenIDConnectProvider(providers)

	require.True(t, ok)
	require.EqualError(t, provider.Resolve(context.Background()), "error resolving provider 'example': error discovering the provider: the discovery endpoint returned status code 404")

	assert.NotEqual(t, 0, requests)
}

func TestProviderResolveShouldCacheTheDiscoveryDocument(t *testing.T) {
	var requests int

	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requests++

		rw.Header().Set(headerContentType, mimeApplicationJSON)

		_, _ = rw.Write([]byte(`{"issuer":"` + server.URL + `","authorization_endpoint":"` + server.URL + `/authorize","token_endpoint":"` + server.URL + `/token","jwks_uri":"` + server.URL + `/jwks.json"}`))
	}))

	defer server.Close()

	providers := NewProviders(&schema.AuthenticationBackendExternalIdentity{
		Providers: []schema.AuthenticationBackendExternalIdentityProvider{
			{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
			},
		},
	}, nil)

	provider, ok := getOpenIDConnectProvider(providers)

	require.True(t, ok)
	require.NoError(t, provider.Resolve(context.Background()))
	require.NoError(t, provider.Resolve(context.Background()))

	assert.Equal(t, 1, requests)
	assert.Equal(t, server.URL+"/authorize", provider.AuthorizationEndpoint())
	assert.Equal(t, server.URL+"/token", provider.TokenEndpoint())
}

func TestNewProvidersShouldUseTheTrustedCertificatePool(t *testing.T) {
	var server *httptest.Server

	server = httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(headerContentType, mimeApplicationJSON)

		_, _ = rw.Write([]byte(`{"issuer":"` + server.URL + `","authorization_endpoint":"` + server.URL + `/authorize","token_endpoint":"` + server.URL + `/token","jwks_uri":"` + server.URL + `/jwks.json"}`))
	}))

	defer server.Close()

	config := &schema.AuthenticationBackendExternalIdentity{
		Providers: []schema.AuthenticationBackendExternalIdentityProvider{
			{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
			},
		},
	}

	untrusted, ok := getOpenIDConnectProvider(NewProviders(config, nil))

	require.True(t, ok)
	require.ErrorContains(t, untrusted.Resolve(context.Background()), "certificate signed by unknown authority")

	pool := x509.NewCertPool()

	pool.AddCert(server.Certificate())

	trusted, ok := getOpenIDConnectProvider(NewProviders(config, pool))

	require.True(t, ok)
	require.NoError(t, trusted.Resolve(context.Background()))

	assert.Equal(t, server.URL+"/token", trusted.TokenEndpoint())
}

func TestProviderShouldFetchTheJSONWebKeySetWithTheTrustedCertificatePool(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	jwks, err := json.Marshal(&jose.JSONWebKeySet{Keys: []jose.JSONWebKey{{KeyID: "kid1", Key: key.Public(), Algorithm: "RS256", Use: "sig"}}})
	require.NoError(t, err)

	var server *httptest.Server

	server = httptest.NewTLSServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(headerContentType, mimeApplicationJSON)

		if r.URL.Path == "/jwks.json" {
			_, _ = rw.Write(jwks)

			return
		}

		_, _ = rw.Write([]byte(`{"issuer":"` + server.URL + `","authorization_endpoint":"` + server.URL + `/authorize","token_endpoint":"` + server.URL + `/token","jwks_uri":"` + server.URL + `/jwks.json"}`))
	}))

	defer server.Close()

	now := time.Now()

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": server.URL, "sub": "abc123", "aud": []string{"client"},
		"exp": now.Add(time.Hour).Unix(), "iat": now.Unix(), "nonce": "the-nonce",
	})

	token.Header["kid"] = "kid1"

	raw, err := token.SignedString(key)
	require.NoError(t, err)

	config := &schema.AuthenticationBackendExternalIdentity{
		Providers: []schema.AuthenticationBackendExternalIdentityProvider{
			{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
			},
		},
	}

	untrusted, ok := getOpenIDConnectProvider(NewProviders(config, nil))

	require.True(t, ok)

	_, err = untrusted.ValidateIDToken(context.Background(), raw, "the-nonce", now)

	require.ErrorContains(t, err, "certificate signed by unknown authority")

	pool := x509.NewCertPool()

	pool.AddCert(server.Certificate())

	trusted, ok := getOpenIDConnectProvider(NewProviders(config, pool))

	require.True(t, ok)

	claims, err := trusted.ValidateIDToken(context.Background(), raw, "the-nonce", now)

	require.NoError(t, err)
	assert.Equal(t, "abc123", claims.Subject)
}

func TestProviderShouldExchangeAndRequestUserInfoWithTheTrustedCertificatePool(t *testing.T) {
	mux := http.NewServeMux()

	mux.HandleFunc("/token", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(headerContentType, mimeApplicationJSON)

		_, _ = rw.Write([]byte(`{"access_token":"at","token_type":"Bearer","id_token":"the-id-token"}`))
	})

	mux.HandleFunc("/userinfo", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(headerContentType, mimeApplicationJSON)

		_, _ = rw.Write([]byte(`{"sub":"abc123","name":"John Smith"}`))
	})

	server := httptest.NewTLSServer(mux)

	defer server.Close()

	config := &schema.AuthenticationBackendExternalIdentity{
		Providers: []schema.AuthenticationBackendExternalIdentityProvider{
			{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				TokenEndpointAuthMethod:  "client_secret_basic",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendExternalIdentityProviderPKCE{ChallengeMethod: "S256"},
				Discovery:                schema.AuthenticationBackendExternalIdentityProviderDiscovery{Disable: true},
				Endpoints: schema.AuthenticationBackendExternalIdentityProviderEndpoints{
					Authorization: server.URL + "/authorize",
					Token:         server.URL + "/token",
					UserInfo:      server.URL + "/userinfo",
					JSONWebKeys:   server.URL + "/jwks.json",
				},
			},
		},
	}

	untrusted, ok := getOpenIDConnectProvider(NewProviders(config, nil))

	require.True(t, ok)

	token, err := untrusted.Exchange(context.Background(), "the-code", "the-verifier", "https://auth.example.com/cb")

	assert.Nil(t, token)
	require.ErrorContains(t, err, "certificate signed by unknown authority")

	claims := &IdentityClaims{Subject: "abc123"}

	require.ErrorContains(t, untrusted.UserInfo(context.Background(), "at", claims), "certificate signed by unknown authority")

	pool := x509.NewCertPool()

	pool.AddCert(server.Certificate())

	trusted, ok := getOpenIDConnectProvider(NewProviders(config, pool))

	require.True(t, ok)

	token, err = trusted.Exchange(context.Background(), "the-code", "the-verifier", "https://auth.example.com/cb")

	require.NoError(t, err)
	assert.Equal(t, &Token{IDToken: "the-id-token", AccessToken: "at"}, token)

	require.NoError(t, trusted.UserInfo(context.Background(), token.AccessToken, claims))
	assert.Equal(t, "John Smith", claims.Name)
}
