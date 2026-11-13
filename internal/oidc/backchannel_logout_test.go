// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"authelia.com/provider/jose"
	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
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

	parsed, err := jose.ParseSignedCompact(received[0], []jose.SignatureAlgorithm{jose.RS256})

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

	parsed, err := jose.ParseSignedCompact(received[0], []jose.SignatureAlgorithm{jose.RS256})

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

func TestBackChannelLogoutDeliverer(t *testing.T) {
	var (
		mu       sync.Mutex
		received []string
	)

	respond := func(status int) *httptest.Server {
		server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			_ = r.ParseForm()

			mu.Lock()

			received = append(received, r.PostFormValue("logout_token"))
			mu.Unlock()

			rw.WriteHeader(status)
		}))

		t.Cleanup(server.Close)

		return server
	}

	ok, unavailable, limited, rejected := respond(http.StatusOK), respond(http.StatusServiceUnavailable), respond(http.StatusTooManyRequests), respond(http.StatusBadRequest)

	provider := newTestBackChannelLogoutProvider(t,
		schema.IdentityProvidersOpenIDConnectClient{ID: "ok", BackChannelLogoutURI: ok.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "unavailable", BackChannelLogoutURI: unavailable.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "limited", BackChannelLogoutURI: limited.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "rejected", BackChannelLogoutURI: rejected.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "no-uri"},
	)

	client := oidc.NewHTTPClient(nil)
	client.RetryMax = 0

	provider.HTTPClient = client

	deliverer := oidc.NewBackChannelLogoutDeliverer(provider, oidc.DetachedContextProviders{Clock: clock.New()})

	testCases := []struct {
		name     string
		clientID string
		issuer   string
		done     bool
		err      string
		received int
	}{
		{"ShouldDeliver", "ok", examplecom, true, "", 1},
		{"ShouldRetryServerError", "unavailable", examplecom, false, "the client responded with status code '503' and the request will be retried: ", 1},
		{"ShouldRetryRateLimiting", "limited", examplecom, false, "the client responded with status code '429' and the request will be retried: ", 1},
		{"ShouldNotRetryClientError", "rejected", examplecom, true, "the client responded with status code '400': ", 1},
		{"ShouldSkipClientWithoutURI", "no-uri", examplecom, true, "", 0},
		{"ShouldSkipUnregisteredClient", "removed", examplecom, true, "", 0},
		{"ShouldNotDeliverWithInvalidIssuer", "ok", "http://[::1", true, "the issuer could not be parsed: ", 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mu.Lock()
			received = nil
			mu.Unlock()

			done, err := deliverer.DeliverBackChannelLogout(context.Background(), model.OAuth2SessionLogoutTask{
				Kind:      model.OAuth2SessionLogoutTaskKindBackChannel,
				SessionID: "sid",
				ClientID:  tc.clientID,
				Subject:   "subject",
				IssuerURL: tc.issuer,
			})

			assert.Equal(t, tc.done, done)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.True(t, strings.HasPrefix(err.Error(), tc.err), err.Error())
			}

			mu.Lock()
			assert.Len(t, received, tc.received)
			mu.Unlock()
		})
	}

	t.Run("ShouldGenerateANewLogoutTokenForEachAttempt", func(t *testing.T) {
		mu.Lock()
		received = nil
		mu.Unlock()

		task := model.OAuth2SessionLogoutTask{Kind: model.OAuth2SessionLogoutTaskKindBackChannel, SessionID: "sid", ClientID: "ok", IssuerURL: examplecom}

		for range 2 {
			_, err := deliverer.DeliverBackChannelLogout(context.Background(), task)
			require.NoError(t, err)
		}

		mu.Lock()
		defer mu.Unlock()

		require.Len(t, received, 2)
		assert.NotEqual(t, received[0], received[1])
	})
}

func TestBackChannelLogoutDelivererShouldNotBeBoundByTheDeadline(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 200)

		rw.WriteHeader(http.StatusOK)
	}))

	t.Cleanup(server.Close)

	provider := newTestBackChannelLogoutProvider(t, schema.IdentityProvidersOpenIDConnectClient{ID: "slow", BackChannelLogoutURI: server.URL})

	client := oidc.NewHTTPClient(nil)
	client.RetryMax = 0

	provider.HTTPClient = client

	deliverer := oidc.NewBackChannelLogoutDeliverer(provider, oidc.DetachedContextProviders{Clock: clock.New()})

	task := model.OAuth2SessionLogoutTask{Kind: model.OAuth2SessionLogoutTaskKindBackChannel, SessionID: "sid", ClientID: "slow", IssuerURL: examplecom}

	testCases := []struct {
		name string
		ctx  func() (context.Context, context.CancelFunc)
		done bool
		err  string
	}{
		{
			name: "ShouldDeliverAfterTheDeadline",
			ctx: func() (context.Context, context.CancelFunc) {
				return context.WithTimeout(context.Background(), time.Millisecond*50)
			},
			done: true,
		},
		{
			name: "ShouldStopWhenCancelled",
			ctx: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())

				time.AfterFunc(time.Millisecond*50, cancel)

				return ctx, cancel
			},
			err: "the client responded with status code '0' and the request will be retried: ",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := tc.ctx()
			defer cancel()

			done, err := deliverer.DeliverBackChannelLogout(ctx, task)

			assert.Equal(t, tc.done, done)

			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.True(t, strings.HasPrefix(err.Error(), tc.err), err.Error())
			}
		})
	}
}

func TestIsBackChannelLogoutRetryable(t *testing.T) {
	testCases := []struct {
		name     string
		status   int
		expected bool
	}{
		{"ShouldRetryUnreachable", 0, true},
		{"ShouldRetryRateLimiting", http.StatusTooManyRequests, true},
		{"ShouldRetryServerError", http.StatusBadGateway, true},
		{"ShouldNotRetryClientError", http.StatusBadRequest, false},
		{"ShouldNotRetrySuccess", http.StatusOK, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.IsBackChannelLogoutRetryable(tc.status))
		})
	}
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
			Lifespan:    time.Minute * 3,
			Concurrency: 5,
		},
	}, oidc.NewIssuer(nil), nil)

	ctx := context.Background()

	assert.NotNil(t, config.GetBackChannelLogoutTokenStrategy(ctx))
	assert.Equal(t, time.Minute*3, config.GetBackChannelLogoutLifespan(ctx))
	assert.Equal(t, 5, config.GetBackChannelLogoutConcurrency(ctx))

	config = oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{}, oidc.NewIssuer(nil), nil)

	assert.Equal(t, time.Minute*2, config.GetBackChannelLogoutLifespan(ctx))
	assert.Equal(t, 10, config.GetBackChannelLogoutConcurrency(ctx))
}

func TestConfig_AcquireBackChannelLogoutSlot(t *testing.T) {
	config := oidc.NewConfig(&schema.IdentityProvidersOpenIDConnect{
		BackChannelLogout: schema.IdentityProvidersOpenIDConnectBackChannelLogout{
			Concurrency: 1,
		},
	}, oidc.NewIssuer(nil), nil)

	release, err := config.AcquireBackChannelLogoutSlot(context.Background())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	blocked, err := config.AcquireBackChannelLogoutSlot(ctx)

	assert.Nil(t, blocked)
	assert.EqualError(t, err, "context deadline exceeded")

	release()

	release, err = config.AcquireBackChannelLogoutSlot(context.Background())
	require.NoError(t, err)

	release()
}

func TestOpenIDConnectProviderMetadata_Logout(t *testing.T) {
	provider := newDiscoveryTestProvider(t, nil)

	metadata, err := provider.NewOpenIDConnectProviderMetadata(context.Background())
	require.NoError(t, err)

	require.NotNil(t, metadata.OpenIDConnectRPInitiatedLogoutMetadata)
	assert.Equal(t, "https://example.com/api/oidc/end-session", metadata.EndSessionEndpoint)

	require.NotNil(t, metadata.OpenIDConnectBackChannelLogoutMetadata)
	assert.True(t, metadata.BackChannelLogoutSupported)
	assert.True(t, metadata.BackChannelLogoutSessionSupported)
}

func newTestBackChannelLogoutProvider(t *testing.T, clients ...schema.IdentityProvidersOpenIDConnectClient) (provider *oidc.OpenIDConnectProvider) {
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
				Clients: clients,
				SpecialUseAddressGuard: schema.IdentityProvidersOpenIDConnectSpecialUseAddressGuard{
					AllowedNetworks: []*net.IPNet{
						{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
						{IP: net.IPv6loopback, Mask: net.CIDRMask(128, 128)},
					},
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
