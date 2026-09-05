package handlers

import (
	"database/sql"
	"net/url"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
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

	t.Run("ShouldHandleUnknownUserCode", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 1)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		handleFlowResponse(mock.Ctx, userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, "ABCDEFGH")

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred using the signature of the user code session to retrieve the device code session preventing a successful flow response", "sql: no rows in result set")
	})

	t.Run("ShouldRedirectToConsentDecisionWithUserCode", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 1)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		mock.Ctx.Response.Reset()

		handleFlowResponse(mock.Ctx, userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, userCode)

		body := redirectResponse{}

		mock.GetResponseData(t, &body)

		target, err := url.Parse(body.Redirect)

		require.NoError(t, err)

		assert.Equal(t, oidc.FrontendEndpointPathConsentDecision, target.Path)
		assert.Equal(t, userCode, target.Query().Get(queryArgUserCode))
		assert.Equal(t, flowOpenIDConnectSubFlowNameDeviceAuthorization, target.Query().Get(queryArgSubflow))
	})

	t.Run("ShouldReply200WhenClientRequiresSecondFactor", func(t *testing.T) {
		mock, userSession := newMockWithSession(t, 1)
		defer mock.Close()

		client := newTestOIDCDeviceCodeClient(t)
		client.AuthorizationPolicy = "two_factor"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		mock.Ctx.Response.Reset()

		handleFlowResponse(mock.Ctx, userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, userCode)

		mock.Assert200OK(t, nil)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "OpenID Connect 1.0 client requires 2FA", nil)
	})
}

func TestHandleFlowResponseOpenIDConnectDeviceAuthSubflowDeviceState(t *testing.T) {
	setup := func(t *testing.T, mutate func(device *model.OAuth2DeviceCodeSession)) (mock *mocks.MockAutheliaCtx, userCode string) {
		t.Helper()

		mock = mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode = mustGetTestOIDCUserCode(t, mock)

		if mutate != nil {
			signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(mock.Ctx, userCode)

			require.NoError(t, err)

			device, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(mock.Ctx, signature)

			require.NoError(t, err)

			mutate(device)

			require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))
		}

		mock.Ctx.Response.Reset()

		return mock, userCode
	}

	testCases := []struct {
		name   string
		mutate func(device *model.OAuth2DeviceCodeSession)
	}{
		{
			name: "ShouldRejectSessionWithSubject",
			mutate: func(device *model.OAuth2DeviceCodeSession) {
				device.Subject = sql.NullString{String: uuid.Must(uuid.NewRandom()).String(), Valid: true}
			},
		},
		{
			name: "ShouldRejectSessionWithChallengeID",
			mutate: func(device *model.OAuth2DeviceCodeSession) {
				device.ChallengeID = uuid.NullUUID{UUID: uuid.Must(uuid.NewRandom()), Valid: true}
			},
		},
		{
			name: "ShouldRejectSessionWithNonNewStatus",
			mutate: func(device *model.OAuth2DeviceCodeSession) {
				device.Status = int(oauthelia2.DeviceAuthorizeStatusApproved)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock, userCode := setup(t, tc.mutate)
			defer mock.Close()

			userSession := newTestOIDCUserSession(1)

			handleFlowResponse(mock.Ctx, &userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, userCode)

			mock.Assert200KO(t, messageOperationFailed)

			AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to handle flow response as the device code session is in an invalid state", nil)
		})
	}

	t.Run("ShouldHandleUnregisteredClient", func(t *testing.T) {
		mock, userCode := setup(t, nil)
		defer mock.Close()

		configWithoutClient := newTestOIDCConfig(t)
		configWithoutClient.Clients = nil

		setupTestOIDCProvider(t, mock, configWithoutClient)

		userSession := newTestOIDCUserSession(1)

		handleFlowResponse(mock.Ctx, &userSession, "", flowNameOpenIDConnect, flowOpenIDConnectSubFlowNameDeviceAuthorization, userCode)

		mock.Assert200KO(t, messageAuthenticationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred loading the client for the device code session", regexpAnyError)
	})
}
