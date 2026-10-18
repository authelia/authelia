// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestHandleFlowResponse(t *testing.T) {
	t.Run("ShouldHandleUnknownFlow", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		userSession := newTestOIDCUserSession(1)

		handleFlowResponse(mock.Ctx, &userSession, "", "not-a-flow", "", "")

		mock.Assert200KO(t, messageAuthenticationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to find flow handler for the given flow parameters", nil)
	})

	t.Run("ShouldHandleUnknownSubflow", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		userSession := newTestOIDCUserSession(1)

		handleFlowResponse(mock.Ctx, &userSession, "", flowNameOpenIDConnect, "not-a-subflow", "")

		mock.Assert200KO(t, messageAuthenticationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to find flow handler for the given flow parameters", nil)
	})
}

func TestHandleFlowResponseOpenIDConnectDeviceAuthSubflow(t *testing.T) {
	newMockWithSession := func(t *testing.T, level int) (*mocks.MockAutheliaCtx, *session.UserSession) {
		t.Helper()

		userSession := newTestOIDCUserSession(level)

		mock := mocks.NewMockAutheliaCtxWithUserSession(t, userSession)

		return mock, &userSession
	}

	t.Run("ShouldHandleAnonymousUser", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		userSession := session.UserSession{}

		handleFlowResponse(mock.Ctx, &userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, "")

		mock.Assert200KO(t, messageAuthenticationFailed)

		require.NotNil(t, mock.Hook.LastEntry())
		assert.Equal(t, "Failed to handle flow response as the user is anonymous", mock.Hook.LastEntry().Message)
	})

	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 1)
		defer mock.Close()

		clearForwardedHeaders(mock)

		handleFlowResponse(mock.Ctx, userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, "")

		mock.Assert200KO(t, messageAuthenticationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred determining the issuer preventing a successful flow response", "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldRedirectToDeviceAuthorizationWhenNoUserCodeAndSecondFactorDisabled", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 1)
		defer mock.Close()

		mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
			AccessControl: schema.AccessControl{DefaultPolicy: "one_factor"},
		})

		handleFlowResponse(mock.Ctx, userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, "")

		body := redirectResponse{}

		mock.GetResponseData(t, &body)

		target, err := url.Parse(body.Redirect)

		require.NoError(t, err)

		assert.Equal(t, oidc.FrontendEndpointPathConsentDeviceAuthorization, target.Path)
		assert.Equal(t, flowNameOpenIDConnect, target.Query().Get(queryArgFlow))
		assert.Equal(t, flowOpenIDConnectSubFlowNameDeviceAuthorization, target.Query().Get(queryArgSubflow))
		assert.Empty(t, target.Query().Get(queryArgFlowID))
	})

	t.Run("ShouldIncludeFlowIDWhenProvided", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 2)
		defer mock.Close()

		handleFlowResponse(mock.Ctx, userSession, "abc", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, "")

		body := redirectResponse{}

		mock.GetResponseData(t, &body)

		target, err := url.Parse(body.Redirect)

		require.NoError(t, err)

		assert.Equal(t, "abc", target.Query().Get(queryArgFlowID))
	})

	t.Run("ShouldReply200WhenNoUserCodeAndSecondFactorRequired", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 1)
		defer mock.Close()

		mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
			AccessControl: schema.AccessControl{DefaultPolicy: "two_factor"},
		})

		handleFlowResponse(mock.Ctx, userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, "")

		mock.Assert200OK(t, nil)
	})

	t.Run("ShouldHandleUserCodeTooLong", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 1)
		defer mock.Close()

		handleFlowResponse(mock.Ctx, userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, strings.Repeat("A", 33))

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to handle flow response as the user code is too long", nil)
	})

	t.Run("ShouldRedirectToDeviceAuthorizationWithUserCode", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 2)
		defer mock.Close()

		handleFlowResponse(mock.Ctx, userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, "ABCDEFGH")

		body := redirectResponse{}

		mock.GetResponseData(t, &body)

		target, err := url.Parse(body.Redirect)

		require.NoError(t, err)

		assert.Equal(t, oidc.FrontendEndpointPathConsentDeviceAuthorization, target.Path)
		assert.Equal(t, "ABCDEFGH", target.Query().Get(queryArgUserCode))
		assert.Equal(t, flowOpenIDConnectSubFlowNameDeviceAuthorization, target.Query().Get(queryArgSubflow))
	})

	t.Run("ShouldReply200WithUserCodeWhenSecondFactorRequired", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 1)
		defer mock.Close()

		mock.Ctx.Providers.Authorizer = authorization.NewAuthorizer(&schema.Configuration{
			AccessControl: schema.AccessControl{DefaultPolicy: "two_factor"},
		})

		handleFlowResponse(mock.Ctx, userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, "ABCDEFGH")

		mock.Assert200OK(t, nil)
	})

	t.Run("ShouldNotRevealWhetherTheUserCodeIsPending", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 2)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		redirect := func(userCode string) *url.URL {
			mock.Ctx.Response.Reset()

			handleFlowResponse(mock.Ctx, userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, userCode)

			body := redirectResponse{}

			mock.GetResponseData(t, &body)

			target, err := url.Parse(body.Redirect)

			require.NoError(t, err)

			return target
		}

		pending, unknown := redirect(mustGetTestOIDCUserCode(t, mock)), redirect("BCDFGHJK")

		assert.Equal(t, pending.Path, unknown.Path)
		assert.Equal(t, oidc.FrontendEndpointPathConsentDeviceAuthorization, unknown.Path)
	})
}
