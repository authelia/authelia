// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jose"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOpenIDConnectProvider_SendBackChannelLogout(t *testing.T) {
	var received []string

	rpOK := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()

		received = append(received, r.PostFormValue("logout_token"))

		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))

		rw.WriteHeader(http.StatusOK)
	}))

	defer rpOK.Close()

	rpBad := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusBadRequest)
	}))

	defer rpBad.Close()

	provider := newTestBackChannelLogoutProvider(t)
	ctx := newTestBackChannelLogoutContext()

	clients := []oauthelia2.Client{
		&oidc.RegisteredClient{ID: "acknowledges", BackChannelLogoutURI: rpOK.URL},
		&oidc.RegisteredClient{ID: "rejects", BackChannelLogoutURI: rpBad.URL},
		&oidc.RegisteredClient{ID: "unregistered"},
		&oidc.RegisteredClient{ID: "requires-sid", BackChannelLogoutURI: rpOK.URL, BackChannelLogoutSessionRequired: true},
	}

	results, err := provider.SendBackChannelLogout(ctx, oauthelia2.NewBackChannelLogoutRequest("john", "", clients))

	require.NoError(t, err)
	require.Len(t, results, 4)

	assert.Equal(t, "acknowledges", results[0].ClientID)
	assert.True(t, results[0].Success())
	assert.Equal(t, http.StatusOK, results[0].Status)
	assert.Equal(t, "rejects", results[1].ClientID)
	assert.False(t, results[1].Success())
	assert.Equal(t, http.StatusBadRequest, results[1].Status)
	assert.Error(t, results[1].Err)
	assert.Equal(t, "unregistered", results[2].ClientID)
	assert.True(t, results[2].Skipped)
	assert.Equal(t, "The OAuth 2.0 Client does not have a registered 'backchannel_logout_uri'.", results[2].Reason)
	assert.Equal(t, "requires-sid", results[3].ClientID)
	assert.True(t, results[3].Skipped)
	assert.Equal(t, "The OAuth 2.0 Client requires the 'sid' claim but no session identifier was supplied.", results[3].Reason)
	require.Len(t, received, 1)

	parsed, err := jose.ParseSigned(received[0], []jose.SignatureAlgorithm{jose.RS256})

	require.NoError(t, err)
	assert.Equal(t, "logout+jwt", parsed.Signatures[0].Header.ExtraHeaders[jose.HeaderType])

	payload, err := parsed.Verify(x509PrivateKeyRSA2048.Public())

	require.NoError(t, err)

	claims := map[string]any{}

	require.NoError(t, json.Unmarshal(payload, &claims))

	assert.Equal(t, "john", claims[oidc.ClaimSubject])
	assert.Equal(t, examplecom, claims[oidc.ClaimIssuer])
	assert.Equal(t, []any{"acknowledges"}, claims[oidc.ClaimAudience])
	assert.NotEmpty(t, claims[oidc.ClaimJWTID])
	assert.NotNil(t, claims[oidc.ClaimIssuedAt])
	assert.NotNil(t, claims[oidc.ClaimExpirationTime])
	assert.NotContains(t, claims, oidc.ClaimSessionID)

	events, ok := claims["events"].(map[string]any)

	require.True(t, ok)
	assert.Contains(t, events, "http://schemas.openid.net/event/backchannel-logout")
	assert.NotContains(t, claims, oidc.ClaimNonce)
}

func TestOpenIDConnectProvider_SendBackChannelLogout_ShouldIncludeSessionID(t *testing.T) {
	var received []string

	rp := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()

		received = append(received, r.PostFormValue("logout_token"))

		rw.WriteHeader(http.StatusOK)
	}))

	defer rp.Close()

	provider := newTestBackChannelLogoutProvider(t)
	ctx := newTestBackChannelLogoutContext()

	clients := []oauthelia2.Client{
		&oidc.RegisteredClient{ID: "requires-sid", BackChannelLogoutURI: rp.URL, BackChannelLogoutSessionRequired: true},
	}

	results, err := provider.SendBackChannelLogout(ctx, oauthelia2.NewBackChannelLogoutRequest("john", "session-abc", clients))

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.True(t, results[0].Success())
	require.Len(t, received, 1)

	parsed, err := jose.ParseSigned(received[0], []jose.SignatureAlgorithm{jose.RS256})

	require.NoError(t, err)

	payload, err := parsed.Verify(x509PrivateKeyRSA2048.Public())

	require.NoError(t, err)

	claims := map[string]any{}

	require.NoError(t, json.Unmarshal(payload, &claims))
	assert.Equal(t, "session-abc", claims[oidc.ClaimSessionID])
	assert.Equal(t, "john", claims[oidc.ClaimSubject])
}

func TestOpenIDConnectProvider_SendBackChannelLogout_ShouldErrorWithoutSubjectOrSessionID(t *testing.T) {
	provider := newTestBackChannelLogoutProvider(t)
	ctx := newTestBackChannelLogoutContext()

	clients := []oauthelia2.Client{
		&oidc.RegisteredClient{ID: "example", BackChannelLogoutURI: examplecom},
	}

	results, err := provider.SendBackChannelLogout(ctx, oauthelia2.NewBackChannelLogoutRequest("", "", clients))

	assert.Nil(t, results)
	assert.EqualError(t, oauthelia2.ErrorToRFC6749Error(err), "invalid_request")
}

func TestOpenIDConnectProvider_SendBackChannelLogout_ShouldNotErrorWithoutClients(t *testing.T) {
	provider := newTestBackChannelLogoutProvider(t)
	ctx := newTestBackChannelLogoutContext()

	results, err := provider.SendBackChannelLogout(ctx, oauthelia2.NewBackChannelLogoutRequest("john", "", nil))

	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestRegisteredClient_BackChannelLogout(t *testing.T) {
	client := &oidc.RegisteredClient{}

	assert.Empty(t, client.GetBackChannelLogoutURI())
	assert.False(t, client.GetBackChannelLogoutSessionRequired())

	client = &oidc.RegisteredClient{
		BackChannelLogoutURI:             "https://app.example.com/logout/backchannel",
		BackChannelLogoutSessionRequired: true,
	}

	assert.Equal(t, "https://app.example.com/logout/backchannel", client.GetBackChannelLogoutURI())
	assert.True(t, client.GetBackChannelLogoutSessionRequired())
}

func TestNewClient_BackChannelLogout(t *testing.T) {
	client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{
		ID:                               "example",
		BackChannelLogoutURI:             "https://app.example.com/logout/backchannel",
		BackChannelLogoutSessionRequired: true,
	}, &schema.IdentityProvidersOpenIDConnect{}, nil)

	require.NotNil(t, client)

	assert.Equal(t, "https://app.example.com/logout/backchannel", client.GetBackChannelLogoutURI())
	assert.True(t, client.GetBackChannelLogoutSessionRequired())
}

func TestConfig_GetBackChannelLogout(t *testing.T) {
	config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{
		BackChannelLogout: schema.IdentityProvidersOpenIDConnectBackChannelLogout{
			Lifespan:    time.Minute * 2,
			Concurrency: 5,
		},
	}, oidc.NewIssuer(nil), nil)

	ctx := context.Background()

	assert.NotNil(t, config.GetBackChannelLogoutTokenStrategy(ctx))
	assert.Equal(t, time.Minute*2, config.GetBackChannelLogoutLifespan(ctx))
	assert.Equal(t, 5, config.GetBackChannelLogoutConcurrency(ctx))

	config = oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{}, oidc.NewIssuer(nil), nil)

	assert.Equal(t, time.Minute*5, config.GetBackChannelLogoutLifespan(ctx))
	assert.Equal(t, 10, config.GetBackChannelLogoutConcurrency(ctx))
}

func TestNewOpenIDConnectWellKnownConfiguration_BackChannelLogout(t *testing.T) {
	disco := oidc.NewOpenIDConnectWellKnownConfiguration(&schema.IdentityProvidersOpenIDConnect{})

	require.NotNil(t, disco.OpenIDConnectBackChannelLogoutDiscoveryOptions)
	assert.True(t, disco.BackChannelLogoutSupported)
	assert.False(t, disco.BackChannelLogoutSessionSupported)
}

func newTestBackChannelLogoutProvider(t *testing.T) (provider *oidc.OpenIDConnectProvider) {
	t.Helper()

	provider = oidc.NewOpenIDConnectProvider(&schema.Configuration{
		IdentityProviders: schema.IdentityProviders{
			OIDC: &schema.IdentityProvidersOpenIDConnect{
				HMACSecret: badhmac,
				JSONWebKeys: []schema.JWK{
					{KeyID: "example", Use: oidc.KeyUseSignature, Algorithm: oidc.SigningAlgRSAUsingSHA256, Key: x509PrivateKeyRSA2048},
				},
				BackChannelLogout: schema.IdentityProvidersOpenIDConnectBackChannelLogout{
					Lifespan:    time.Minute * 5,
					Concurrency: 10,
				},
			},
		},
	}, nil, nil)

	require.NotNil(t, provider)

	return provider
}

func newTestBackChannelLogoutContext() (ctx *TestContext) {
	return &TestContext{
		Context:       context.Background(),
		MockIssuerURL: MustParseRequestURI(examplecom),
		Clock:         clock.New(),
	}
}
