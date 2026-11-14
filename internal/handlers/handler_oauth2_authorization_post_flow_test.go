// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestOAuth2AuthorizationPOSTFlow(t *testing.T) {
	testCases := []struct {
		name          string
		authenticated bool
		pushed        bool
		policy        string
		setup         func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values)
		expect        func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values, rw *httptest.ResponseRecorder)
	}{
		{
			name:          "ShouldAuthorizeWithImplicitConsent",
			authenticated: true,
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values, rw *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusSeeOther, rw.Code)

				location := mustParseTestURI(t, rw.Header().Get(fasthttp.HeaderLocation))

				assert.Equal(t, "app.example.com", location.Host)
				assert.NotEmpty(t, location.Query().Get(testOIDCFormParameterCode))
				assert.Equal(t, "abcdefghijklmnopqrstuvwxyz", location.Query().Get(oidc.FormParameterState))
				assert.Empty(t, location.Query().Get("error"))
			},
		},
		{
			name:          "ShouldAuthorizePushedRequest",
			authenticated: true,
			pushed:        true,
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values, rw *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusSeeOther, rw.Code)

				location := mustParseTestURI(t, rw.Header().Get(fasthttp.HeaderLocation))

				assert.Equal(t, "app.example.com", location.Host)
				assert.NotEmpty(t, location.Query().Get(testOIDCFormParameterCode))
				assert.Empty(t, location.Query().Get("error"))
			},
		},
		{
			name: "ShouldContinueStoredConsentSessionForAnonymousUser",
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values, rw *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusSeeOther, rw.Code)

				location := mustParseTestURI(t, rw.Header().Get(fasthttp.HeaderLocation))

				assert.Equal(t, "login.example.com:8080", location.Host)
				assert.Equal(t, flowNameOpenIDConnect, location.Query().Get(queryArgFlow))
				assert.Equal(t, reference.Get(queryArgConsentID), location.Query().Get(queryArgFlowID))
			},
		},
		{
			name:          "ShouldContinueStoredConsentSessionForInsufficientlyAuthenticatedUser",
			authenticated: true,
			policy:        "two_factor",
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values, rw *httptest.ResponseRecorder) {
				require.Equal(t, http.StatusSeeOther, rw.Code)

				location := mustParseTestURI(t, rw.Header().Get(fasthttp.HeaderLocation))

				assert.Equal(t, "login.example.com:8080", location.Host)
				assert.Equal(t, flowNameOpenIDConnect, location.Query().Get(queryArgFlow))
				assert.Equal(t, reference.Get(queryArgConsentID), location.Query().Get(queryArgFlowID))
			},
		},
		{
			name:          "ShouldRejectReferenceFromAnotherUserAgent",
			authenticated: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values) {
				mock.Ctx.Request.Header.SetCookie(session.AuthorizationBindingCookieName("authelia_session"), newTestSessionCookie("another-user-agent"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values, rw *httptest.ResponseRecorder) {
				assertTestOAuth2AuthorizationConsentReferenceRejected(t, mock, rw, "the consent session is not bound to the user agent")
			},
		},
		{
			name:          "ShouldRejectReferenceWithoutBinding",
			authenticated: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values) {
				mock.Ctx.Request.Header.DelCookie(session.AuthorizationBindingCookieName("authelia_session"))
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values, rw *httptest.ResponseRecorder) {
				assertTestOAuth2AuthorizationConsentReferenceRejected(t, mock, rw, "the consent session is not bound to the user agent")
			},
		},
		{
			name:          "ShouldRejectReferenceForAnotherClient",
			authenticated: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values) {
				reference.Set(oidc.FormParameterClientID, "another-client")
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values, rw *httptest.ResponseRecorder) {
				assertTestOAuth2AuthorizationConsentReferenceRejected(t, mock, rw, nil)
			},
		},
		{
			name:          "ShouldRejectMalformedReference",
			authenticated: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values) {
				reference.Set(queryArgConsentID, "not-a-uuid")
			},
			expect: func(t *testing.T, mock *mocks.MockAutheliaCtx, reference url.Values, rw *httptest.ResponseRecorder) {
				assertTestOAuth2AuthorizationConsentReferenceRejected(t, mock, rw, "invalid UUID length: 10")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var mock *mocks.MockAutheliaCtx

			if tc.authenticated {
				mock = mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			} else {
				mock = mocks.NewMockAutheliaCtx(t)
			}

			defer mock.Close()

			client := newTestOIDCAuthorizationCodeClient(t)
			client.ConsentMode = "implicit"

			if tc.policy != "" {
				client.AuthorizationPolicy = tc.policy
			}

			config := newTestOIDCConfig(t)
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

			setupTestOIDCProvider(t, mock, config)
			setupTestOIDCSessionStore(t, mock)
			setupTestOIDCConsentStore(t, mock)
			setupTestOIDCSubjectStore(t, mock)
			setupTestOIDCUserDetails(t, mock)
			setupTestOIDCPARStore(t, mock)

			values := newTestOIDCAuthorizationValues()

			if tc.pushed {
				values = url.Values{
					oidc.FormParameterClientID:   []string{testOIDCAuthorizationCodeID},
					oidc.FormParameterRequestURI: []string{mustPushTestOIDCAuthorizationRequest(t, mock)},
				}
			}

			rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCAuthorizationEndpoint, values)

			OAuth2AuthorizationPOST(mock.Ctx, rw, r)

			require.Equal(t, http.StatusSeeOther, rw.Code)

			reference := mustParseTestURI(t, rw.Header().Get(fasthttp.HeaderLocation)).Query()

			require.Len(t, reference, 2)

			mock.Ctx.Response.Reset()

			if tc.setup != nil {
				tc.setup(t, mock, reference)
			}

			mock.Ctx.Request.SetRequestURI(oidc.EndpointPathAuthorization + "?" + reference.Encode())

			rw, r = newTestOAuth2Request(t, fasthttp.MethodGet, testOIDCAuthorizationEndpoint, reference)

			OAuth2AuthorizationGET(mock.Ctx, rw, r)

			if tc.expect != nil {
				tc.expect(t, mock, reference, rw)
			}
		})
	}
}

func assertTestOAuth2AuthorizationConsentReferenceRejected(t *testing.T, mock *mocks.MockAutheliaCtx, rw *httptest.ResponseRecorder, err any) {
	t.Helper()

	location := mustParseTestURI(t, rw.Header().Get(fasthttp.HeaderLocation))

	assert.Equal(t, oidc.FrontendEndpointPathConsentCompletion, location.Path)
	assert.NotEmpty(t, location.Query().Get("error"))
	assert.Empty(t, location.Query().Get(testOIDCFormParameterCode))

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpAuthorizationFailed, err)
}
