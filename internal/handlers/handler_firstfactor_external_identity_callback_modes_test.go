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

	"github.com/authelia/authelia/v4/internal/externalidentity"
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
			Mode:   externalidentity.ResponseModeFormPost,
			Method: fasthttp.MethodPost,
			Form:   url.Values{"code": {"the-code"}, "state": {"the-state"}},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadExternalIdentityLinkBySubject(mock.Ctx, "openid_connect", "https://op.example.com", "abc123").
					Return(nil, storage.ErrNoExternalIdentityLink)
			},
			Location: "https://login.example.com:8080/external-identity/link?link_provider=example",
		},
		{
			Name:     "ShouldIgnoreQueryParametersOfFormPostResponse",
			Mode:     externalidentity.ResponseModeFormPost,
			Method:   fasthttp.MethodPost,
			Query:    url.Values{"code": {"the-code"}, "state": {"the-state"}},
			Location: "https://login.example.com:8080/?external_identity_error=true",
		},
		{
			Name:     "ShouldRejectQueryResponseForFormPostProvider",
			Mode:     externalidentity.ResponseModeFormPost,
			Method:   fasthttp.MethodGet,
			Query:    url.Values{"code": {"the-code"}, "state": {"the-state"}},
			Location: "https://login.example.com:8080/?external_identity_error=true",
		},
		{
			Name:     "ShouldRejectFormPostResponseForQueryProvider",
			Mode:     externalidentity.ResponseModeQuery,
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
			mock.Ctx.Request.SetRequestURI("/api/firstfactor/external-identity/example/callback?" + tc.Query.Encode())
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
			mock.Ctx.Request.SetRequestURI("/api/firstfactor/external-identity/example/callback?code=the-code&state=the-state")
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
