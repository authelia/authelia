// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestFlowContinuePOST(t *testing.T) {
	t.Run("ShouldRedirectToAuthorizationWhenAuthenticationSufficient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		provider, err := mock.Ctx.GetSessionProvider()
		require.NoError(t, err)
		require.NoError(t, provider.SaveSession(mock.Ctx.RequestCtx, newTestOIDCUserSession(1)))

		mock.Ctx.Request.SetBodyString(`{"flow":"openid_connect","flowID":"` + consent.ChallengeID.String() + `"}`)

		FlowContinuePOST(mock.Ctx)

		body := redirectResponse{}

		mock.GetResponseData(t, &body)

		target, err := url.Parse(body.Redirect)

		require.NoError(t, err)

		assert.Equal(t, oidc.EndpointPathAuthorization, target.Path)
		assert.Equal(t, consent.ChallengeID.String(), target.Query().Get(queryArgConsentID))
	})

	t.Run("ShouldHandleAnonymousUser", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.Ctx.Request.SetBodyString(`{"flow":"openid_connect","flowID":"` + uuid.Must(uuid.NewRandom()).String() + `"}`)

		FlowContinuePOST(mock.Ctx)

		mock.Assert200KO(t, messageAuthenticationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to continue the flow as the user is anonymous", nil)
	})

	t.Run("ShouldHandleMalformedBody", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Request.SetBodyString(`not json`)

		FlowContinuePOST(mock.Ctx)

		mock.Assert200KO(t, messageAuthenticationFailed)
	})

	t.Run("ShouldHandleMissingFlow", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		provider, err := mock.Ctx.GetSessionProvider()
		require.NoError(t, err)
		require.NoError(t, provider.SaveSession(mock.Ctx.RequestCtx, newTestOIDCUserSession(1)))

		mock.Ctx.Request.SetBodyString(`{"flowID":"` + uuid.Must(uuid.NewRandom()).String() + `"}`)

		FlowContinuePOST(mock.Ctx)

		mock.Assert200KO(t, messageAuthenticationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to continue the flow as no flow was provided", nil)
	})
}
