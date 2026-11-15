// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/identity"
)

func TestFirstFactorExternalIdentitySharedCallbackGET(t *testing.T) {
	key, server, requests := newTestOpenIDConnectUpstream(t)

	defer server.Close()

	const failure = "https://login.example.com:8080/?external_identity_error=true"

	testCases := []struct {
		Name     string
		Shared   bool
		Delivery bool
		Query    url.Values
		Linked   bool
		Location string
	}{
		{
			Name:     "ShouldSignInWhenTheIssuerIsDeliveredToTheSharedRedirectURI",
			Shared:   true,
			Delivery: true,
			Query:    url.Values{"code": {"the-code"}, "state": {"the-state"}, "iss": {"https://op.example.com"}},
			Linked:   true,
			Location: "https://app.example.com/",
		},
		{
			Name:     "ShouldRejectAResponseWithoutTheIssuer",
			Shared:   true,
			Delivery: true,
			Query:    url.Values{"code": {"the-code"}, "state": {"the-state"}},
			Location: failure,
		},
		{
			Name:     "ShouldRejectAnErrorResponseWithoutTheIssuer",
			Shared:   true,
			Delivery: true,
			Query:    url.Values{"error": {"access_denied"}, "state": {"the-state"}},
			Location: failure,
		},
		{
			Name:     "ShouldRejectAResponseWithAnotherIssuer",
			Shared:   true,
			Delivery: true,
			Query:    url.Values{"code": {"the-code"}, "state": {"the-state"}, "iss": {"https://other.example.com"}},
			Location: failure,
		},
		{
			Name:     "ShouldRejectAResponseForAProviderWhichDoesNotUseTheSharedRedirectURI",
			Delivery: true,
			Query:    url.Values{"code": {"the-code"}, "state": {"the-state"}, "iss": {"https://op.example.com"}},
			Location: failure,
		},
		{
			Name:     "ShouldRejectAResponseDeliveredToTheRedirectURIOfAProviderWhichUsesTheSharedRedirectURI",
			Shared:   true,
			Query:    url.Values{"code": {"the-code"}, "state": {"the-state"}, "iss": {"https://op.example.com"}},
			Location: failure,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := newTestOpenIDConnectCallbackMock(t)

			defer mock.Close()

			if tc.Shared {
				mock.Ctx.Providers.ExternalIdentity = newTestRelyingPartySharedProviders(server.URL+"/token", key, identity.ResponseModeQuery)
			} else {
				mock.Ctx.Providers.ExternalIdentity = newTestRelyingPartyProvidersWithUpstream(server.URL+"/token", key, false)
			}

			if tc.Delivery {
				mock.Ctx.Request.SetRequestURI("/api/identity/callback?" + tc.Query.Encode())
			} else {
				mock.Ctx.SetUserValue("provider", "example")
				mock.Ctx.Request.SetRequestURI("/api/identity/example/callback?" + tc.Query.Encode())
			}

			mock.Ctx.Request.SetHost("login.example.com:8080")

			setTestExternalIdentityFlow(t, mock)

			if tc.Linked {
				setupTestOpenIDConnectCallbackLinked(mock)
			}

			observed := requests.Load()

			if tc.Delivery {
				FirstFactorExternalIdentitySharedCallbackGET(mock.Ctx)
			} else {
				FirstFactorExternalIdentityCallbackGET(mock.Ctx)
			}

			assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.Location, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

			if !tc.Linked {
				assert.Equal(t, observed, requests.Load(), "the provider must not be contacted for a response which is rejected")
			}

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			assert.Nil(t, userSession.ExternalIdentity, "the flow must be consumed whatever the outcome")
		})
	}
}

func TestFirstFactorExternalIdentitySharedCallbackFormPostRelay(t *testing.T) {
	key, server, _ := newTestOpenIDConnectUpstream(t)

	defer server.Close()

	testCases := []struct {
		Name    string
		Issuer  string
		Relayed bool
	}{
		{"ShouldRelayTheResponseOfTheProviderIdentifiedByTheIssuer", "https://op.example.com", true},
		{"ShouldNotRelayWithoutTheIssuer", "", false},
		{"ShouldNotRelayForAnUnknownIssuer", "https://other.example.com", false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := newTestOpenIDConnectCallbackMock(t)

			defer mock.Close()

			mock.Ctx.Providers.ExternalIdentity = newTestRelyingPartySharedProviders(server.URL+"/token", key, identity.ResponseModeFormPost)

			values := url.Values{"code": {"the-code"}, "state": {"the-state"}}

			if tc.Issuer != "" {
				values.Set("iss", tc.Issuer)
			}

			mock.Ctx.Request.Header.SetMethod(fasthttp.MethodPost)
			mock.Ctx.Request.SetRequestURI("/api/identity/callback")
			mock.Ctx.Request.SetHost("login.example.com:8080")
			mock.Ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
			mock.Ctx.Request.SetBodyString(values.Encode())

			FirstFactorExternalIdentitySharedCallbackPOST(mock.Ctx)

			if !tc.Relayed {
				assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
				assert.Equal(t, "https://login.example.com:8080/?external_identity_error=true", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

				return
			}

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

			body := string(mock.Ctx.Response.Body())

			assert.Contains(t, body, `action="https://login.example.com:8080/api/identity/callback"`)
			assert.Contains(t, body, `name="iss" value="https://op.example.com"`)
			assert.Contains(t, body, `name="external_identity_relay" value="true"`)
		})
	}
}
