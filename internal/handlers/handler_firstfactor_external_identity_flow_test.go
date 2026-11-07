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

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestHandleExternalIdentityFlowRedirect(t *testing.T) {
	t.Run("ShouldRedirectToThePortalWithTheFlowWhenASecondFactorIsRequired", func(t *testing.T) {
		userSession := newTestOIDCUserSession(1)

		mock := mocks.NewMockAutheliaCtxWithUserSession(t, userSession)
		defer mock.Close()

		handleExternalIdentityFlowRedirect(mock.Ctx, &userSession, "abc", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, "")

		assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())

		location, err := url.Parse(string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
		require.NoError(t, err)

		root := mock.Ctx.TemplateRootURL()

		assert.Equal(t, root.Host, location.Host)
		assert.Equal(t, root.Path, location.Path)
		assert.Equal(t, url.Values{
			queryArgFlow:    []string{flowNameOpenIDConnect},
			queryArgFlowID:  []string{"abc"},
			queryArgSubflow: []string{flowOpenIDConnectSubFlowNameDeviceAuthorization},
		}, location.Query())
	})

	t.Run("ShouldRedirectToTheResumedFlow", func(t *testing.T) {
		userSession := newTestOIDCUserSession(2)

		mock := mocks.NewMockAutheliaCtxWithUserSession(t, userSession)
		defer mock.Close()

		handleExternalIdentityFlowRedirect(mock.Ctx, &userSession, "abc", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, "")

		assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())

		location, err := url.Parse(string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
		require.NoError(t, err)

		assert.Equal(t, oidc.FrontendEndpointPathConsentDeviceAuthorization, location.Path)
		assert.Equal(t, flowNameOpenIDConnect, location.Query().Get(queryArgFlow))
		assert.Equal(t, flowOpenIDConnectSubFlowNameDeviceAuthorization, location.Query().Get(queryArgSubflow))
		assert.Equal(t, "abc", location.Query().Get(queryArgFlowID))
	})

	t.Run("ShouldRedirectToTheErrorPageWhenTheFlowCannotBeResumed", func(t *testing.T) {
		userSession := newTestOIDCUserSession(1)

		mock := mocks.NewMockAutheliaCtxWithUserSession(t, userSession)
		defer mock.Close()

		handleExternalIdentityFlowRedirect(mock.Ctx, &userSession, "", "not-a-flow", "", "")

		assert.Equal(t, fasthttp.StatusFound, mock.Ctx.Response.StatusCode())

		location, err := url.Parse(string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderLocation)))
		require.NoError(t, err)

		assert.Equal(t, "true", location.Query().Get(queryArgExternalIdentityError))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to find flow handler for the given flow parameters", nil)
	})
}
