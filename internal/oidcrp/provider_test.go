package oidcrp

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"authelia.com/provider/oauth2/token/jose"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

//nolint:gosec // Test Credentials.
func TestNewProviders(t *testing.T) {
	testCases := []struct {
		Name   string
		Have   *schema.AuthenticationBackendOpenIDConnect
		Assert func(t *testing.T, providers *Providers)
	}{
		{
			Name: "ShouldBuildProviderWithDiscoveryDisabled",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{
						ID: "example", Name: "Example", Issuer: "https://op.example.com",
						ClientID: "client", ClientSecret: "secret",
						Scopes:                   []string{"openid", "email"},
						TokenEndpointAuthMethod:  "client_secret_basic",
						IDTokenSignedResponseAlg: "RS256",
						PKCE:                     schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "S256"},
						Discovery:                schema.AuthenticationBackendOpenIDConnectProviderDiscovery{Disable: true},
						Endpoints: schema.AuthenticationBackendOpenIDConnectProviderEndpoints{
							Authorization: "https://op.example.com/authorize",
							Token:         "https://op.example.com/token",
							JSONWebKeys:   "https://op.example.com/jwks.json",
						},
					},
				},
			},
			Assert: func(t *testing.T, providers *Providers) {
				provider, ok := providers.Get("example")

				require.True(t, ok)
				assert.Equal(t, "Example", provider.Name)
				assert.Equal(t, "https://op.example.com/authorize", provider.AuthorizationEndpoint())
				assert.Len(t, providers.All(), 1)
			},
		},
		{
			Name: "ShouldReturnNotOkForUnknownProvider",
			Have: &schema.AuthenticationBackendOpenIDConnect{
				Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
					{
						ID: "example", Name: "Example", Issuer: "https://op.example.com",
						ClientID: "client", ClientSecret: "secret",
						IDTokenSignedResponseAlg: "RS256",
						PKCE:                     schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "S256"},
						Discovery:                schema.AuthenticationBackendOpenIDConnectProviderDiscovery{Disable: true},
						Endpoints: schema.AuthenticationBackendOpenIDConnectProviderEndpoints{
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

	providers := NewProviders(&schema.AuthenticationBackendOpenIDConnect{
		Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
			{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "S256"},
			},
		},
	}, nil)

	require.NotNil(t, providers)
	assert.Equal(t, 0, requests)

	provider, ok := providers.Get("example")

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

	providers := NewProviders(&schema.AuthenticationBackendOpenIDConnect{
		Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
			{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "S256"},
			},
		},
	}, nil)

	provider, ok := providers.Get("example")

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

	config := &schema.AuthenticationBackendOpenIDConnect{
		Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
			{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "S256"},
			},
		},
	}

	untrusted, ok := NewProviders(config, nil).Get("example")

	require.True(t, ok)
	require.ErrorContains(t, untrusted.Resolve(context.Background()), "certificate signed by unknown authority")

	pool := x509.NewCertPool()

	pool.AddCert(server.Certificate())

	trusted, ok := NewProviders(config, pool).Get("example")

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

	config := &schema.AuthenticationBackendOpenIDConnect{
		Providers: []schema.AuthenticationBackendOpenIDConnectProvider{
			{
				ID: "example", Name: "Example", Issuer: server.URL,
				ClientID: "client", ClientSecret: "secret",
				IDTokenSignedResponseAlg: "RS256",
				PKCE:                     schema.AuthenticationBackendOpenIDConnectProviderPKCE{ChallengeMethod: "S256"},
			},
		},
	}

	untrusted, ok := NewProviders(config, nil).Get("example")

	require.True(t, ok)

	_, err = untrusted.ValidateIDToken(context.Background(), raw, "the-nonce", now)

	require.ErrorContains(t, err, "certificate signed by unknown authority")

	pool := x509.NewCertPool()

	pool.AddCert(server.Certificate())

	trusted, ok := NewProviders(config, pool).Get("example")

	require.True(t, ok)

	claims, err := trusted.ValidateIDToken(context.Background(), raw, "the-nonce", now)

	require.NoError(t, err)
	assert.Equal(t, "abc123", claims.Subject)
}
