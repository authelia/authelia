// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/identity"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestFirstFactorExternalIdentityCallbackResponseModes(t *testing.T) {
	key, server, _ := newTestOpenIDConnectUpstream(t)

	defer server.Close()

	testCases := []struct {
		Name     string
		Mode     string
		Method   string
		Query    url.Values
		Form     url.Values
		Setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Location string
	}{
		{
			Name:   "ShouldHandleFormPostResponse",
			Mode:   identity.ResponseModeFormPost,
			Method: fasthttp.MethodPost,
			Form:   url.Values{"code": {"the-code"}, "state": {"the-state"}, "external_identity_relay": {"true"}},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadExternalIdentityLinkBySubject(mock.Ctx, "openid_connect", "https://op.example.com", "abc123").
					Return(nil, storage.ErrNoExternalIdentityLink)
			},
			Location: "https://login.example.com:8080/external-identity/link?link_provider=example",
		},
		{
			Name:     "ShouldIgnoreQueryParametersOfFormPostResponse",
			Mode:     identity.ResponseModeFormPost,
			Method:   fasthttp.MethodPost,
			Query:    url.Values{"code": {"the-code"}, "state": {"the-state"}},
			Form:     url.Values{"external_identity_relay": {"true"}},
			Location: "https://login.example.com:8080/?external_identity_error=true",
		},
		{
			Name:     "ShouldRejectQueryResponseForFormPostProvider",
			Mode:     identity.ResponseModeFormPost,
			Method:   fasthttp.MethodGet,
			Query:    url.Values{"code": {"the-code"}, "state": {"the-state"}},
			Location: "https://login.example.com:8080/?external_identity_error=true",
		},
		{
			Name:     "ShouldRejectFormPostResponseForQueryProvider",
			Mode:     identity.ResponseModeQuery,
			Method:   fasthttp.MethodPost,
			Form:     url.Values{"code": {"the-code"}, "state": {"the-state"}},
			Location: "https://login.example.com:8080/?external_identity_error=true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := newTestOpenIDConnectCallbackMock(t)

			defer mock.Close()

			mock.Ctx.Providers.ExternalIdentity = newTestExternalIdentityProvidersWithResponseMode(server.URL+"/token", key, tc.Mode)

			mock.Ctx.SetUserValue("provider", "example")
			mock.Ctx.Request.Header.SetMethod(tc.Method)
			mock.Ctx.Request.SetRequestURI("/api/identity/example/callback?" + tc.Query.Encode())
			mock.Ctx.Request.SetHost("login.example.com:8080")

			if tc.Form != nil {
				mock.Ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
				mock.Ctx.Request.SetBodyString(tc.Form.Encode())
			}

			setTestExternalIdentityFlow(t, mock)

			if tc.Setup != nil {
				tc.Setup(t, mock)
			}

			if tc.Method == fasthttp.MethodPost {
				FirstFactorExternalIdentityCallbackPOST(mock.Ctx)
			} else {
				FirstFactorExternalIdentityCallbackGET(mock.Ctx)
			}

			assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.Location, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			assert.Nil(t, userSession.ExternalIdentity, "the flow must be consumed whatever the response mode")
			assert.True(t, userSession.IsAnonymous())
		})
	}
}

func TestFirstFactorExternalIdentityCallbackFormPostRelay(t *testing.T) {
	key, server, requests := newTestOpenIDConnectUpstream(t)

	defer server.Close()

	t.Run("ShouldRelayFormPostResponseFromThePortalOrigin", func(t *testing.T) {
		mock := newTestOpenIDConnectCallbackMock(t)

		defer mock.Close()

		mock.Ctx.Providers.ExternalIdentity = newTestExternalIdentityProvidersWithResponseMode(server.URL+"/token", key, identity.ResponseModeFormPost)

		mock.Ctx.SetUserValue("provider", "example")
		mock.Ctx.Request.Header.SetMethod(fasthttp.MethodPost)
		mock.Ctx.Request.SetRequestURI("/api/identity/example/callback")
		mock.Ctx.Request.SetHost("login.example.com:8080")
		mock.Ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
		mock.Ctx.Request.SetBodyString(url.Values{
			"code":       {"the-code"},
			"state":      {`the-state"><script>alert(1)</script>`},
			"iss":        {"https://op.example.com"},
			"unexpected": {"planted"},
		}.Encode())

		setTestExternalIdentityFlow(t, mock)

		observed := requests.Load()

		FirstFactorExternalIdentityCallbackPOST(mock.Ctx)

		assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
		assert.Equal(t, "text/html; charset=utf-8", string(mock.Ctx.Response.Header.ContentType()))
		assert.Equal(t, true, mock.Ctx.UserValue(middlewares.UserValueKeyOpenIDConnectResponseModeFormPost))

		body := string(mock.Ctx.Response.Body())

		assert.Contains(t, body, `action="https://login.example.com:8080/api/identity/example/callback"`)
		assert.Contains(t, body, `name="code" value="the-code"`)
		assert.Contains(t, body, `name="iss" value="https://op.example.com"`)
		assert.Contains(t, body, `name="external_identity_relay" value="true"`)
		assert.NotContains(t, body, "planted", "only the parameters of an authorization response are relayed")
		assert.NotContains(t, body, `"><script>alert(1)</script>`, "relayed values must be escaped")

		assert.Equal(t, observed, requests.Load(), "the provider must not be contacted until the response is relayed")

		userSession, err := mock.Ctx.GetSession()
		require.NoError(t, err)

		assert.NotNil(t, userSession.ExternalIdentity, "the flow must not be consumed until the response is relayed")
	})

	t.Run("ShouldNotRelayForAnUnknownProvider", func(t *testing.T) {
		mock := newTestOpenIDConnectCallbackMock(t)

		defer mock.Close()

		mock.Ctx.Providers.ExternalIdentity = newTestExternalIdentityProvidersWithResponseMode(server.URL+"/token", key, identity.ResponseModeFormPost)

		mock.Ctx.SetUserValue("provider", "missing")
		mock.Ctx.Request.Header.SetMethod(fasthttp.MethodPost)
		mock.Ctx.Request.SetRequestURI("/api/identity/missing/callback")
		mock.Ctx.Request.SetHost("login.example.com:8080")
		mock.Ctx.Request.Header.SetContentType("application/x-www-form-urlencoded")
		mock.Ctx.Request.SetBodyString(url.Values{"code": {"the-code"}, "state": {"the-state"}}.Encode())

		FirstFactorExternalIdentityCallbackPOST(mock.Ctx)

		assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
		assert.Equal(t, "https://login.example.com:8080/?external_identity_error=true", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
	})
}

func TestFirstFactorExternalIdentityCallbackUserInfo(t *testing.T) {
	key, server, requests := newTestOpenIDConnectUpstream(t)

	defer server.Close()

	testCases := []struct {
		Name     string
		Path     string
		Setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Assert   func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Location string
	}{
		{
			Name: "ShouldAuthenticateLinkedUserWhenTheUserInfoSubjectMatches",
			Path: "/userinfo",
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				setupTestOpenIDConnectCallbackLinked(mock)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Equal(t, "john", userSession.Username)
			},
			Location: "https://app.example.com/",
		},
		{
			Name: "ShouldRejectUserInfoSubjectMismatch",
			Path: "/userinfo-invalid-sub",
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, userSession.IsAnonymous())
				assert.Nil(t, userSession.ExternalIdentityPending)
			},
			Location: "https://login.example.com:8080/?external_identity_error=true",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := newTestOpenIDConnectCallbackMock(t)

			defer mock.Close()

			mock.Ctx.Providers.ExternalIdentity = newTestRelyingPartyProvidersWithUpstreamUserInfo(server.URL+"/token", server.URL+tc.Path, key, false)
			mock.Ctx.SetUserValue("provider", "example")
			mock.Ctx.Request.SetRequestURI("/api/identity/example/callback?code=the-code&state=the-state")
			mock.Ctx.Request.SetHost("login.example.com:8080")

			setTestExternalIdentityFlow(t, mock)

			if tc.Setup != nil {
				tc.Setup(t, mock)
			}

			before := requests.Load()

			FirstFactorExternalIdentityCallbackGET(mock.Ctx)

			assert.Equal(t, int64(2), requests.Load()-before, "the token and userinfo endpoints must each be requested once")
			assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.Location, string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))

			tc.Assert(t, mock)
		})
	}
}

func setTestExternalIdentityFlow(t *testing.T, mock *mocks.MockAutheliaCtx) {
	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	userSession.ExternalIdentity = &session.ExternalIdentityFlow{
		Provider:      "example",
		State:         "the-state",
		Nonce:         "the-nonce",
		CodeVerifier:  "the-verifier",
		TargetURL:     "https://app.example.com",
		RequestMethod: fasthttp.MethodGet,
		Expires:       mock.Ctx.GetClock().Now().Add(time.Minute * 3),
	}

	require.NoError(t, mock.Ctx.SaveSession(userSession))
}
