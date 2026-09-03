package handlers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2ConsentDeviceAuthorizationGET(t *testing.T) {
	t.Run("ShouldHandleUnknownUserCode", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		mock.Ctx.Request.SetRequestURI("/api/oidc/consent?user_code=ABCDEFGH")

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred loading device code session using user code signature during the Consent Flow stage of the Device Authorization Flow", "sql: no rows in result set")
	})

	t.Run("ShouldHandleInactiveDeviceCodeSession", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(mock.Ctx, userCode)

		require.NoError(t, err)

		device, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(mock.Ctx, signature)

		require.NoError(t, err)

		device.Revoked = true

		require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetRequestURI("/api/oidc/consent?user_code=" + userCode)

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Flow failed to retrieve Consent Flow data as device code session is not active or has been revoked", nil)
	})

	t.Run("ShouldHandleInsufficientAuthenticationLevel", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCDeviceCodeClient(t)
		client.AuthorizationPolicy = "two_factor"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetRequestURI("/api/oidc/consent?user_code=" + userCode)

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Flow failed to retrieve Consent Flow data as the user is not sufficiently authenticated", nil)
	})

	t.Run("ShouldReturnConsentInformation", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetRequestURI("/api/oidc/consent?user_code=" + userCode)

		OAuth2ConsentGET(mock.Ctx)

		body := map[string]any{}

		mock.GetResponseData(t, &body)

		assert.Equal(t, testOIDCDeviceCodeID, body["client_id"])
		assert.Contains(t, body, "scopes")
	})
}

func TestOAuth2ConsentDeviceAuthorizationPOST(t *testing.T) {
	t.Run("ShouldHandleMissingUserCode", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		mock.Ctx.Request.SetBodyString(`{"subflow":"device_authorization","client_id":"device-code","consent":true}`)

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Request is missing the required field 'user_code' from the JSON body", nil)
	})

	t.Run("ShouldHandleAnonymousUser", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		mock.Ctx.Request.SetBodyString(`{"subflow":"device_authorization","client_id":"device-code","consent":true,"user_code":"ABCDEFGH"}`)

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching user session during the Consent Flow stage of the Device Authorization Flow as the user is anonymous", nil)
	})

	t.Run("ShouldHandleUnknownUserCode", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		mock.Ctx.Request.SetBodyString(`{"subflow":"device_authorization","client_id":"device-code","consent":true,"user_code":"ABCDEFGH"}`)

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred using the signature of the user code session to retrieve the device code session during the Consent Flow stage of the Device Authorization Flow", "sql: no rows in result set")
	})

	t.Run("ShouldHandleClientIDMismatch", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","client_id":"a-different-client","consent":true,"user_code":"%s"}`, userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred matching the user code to the device code session during the Consent Flow stage of the Device Authorization Flow as the client id of the form and the client id of the consent session do not match", nil)
	})

	t.Run("ShouldGrantConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":true,"user_code":"%s"}`, testOIDCDeviceCodeID, userCode))

		OAuth2ConsentPOST(mock.Ctx)

		body := oidc.ConsentPostResponseBody{}

		mock.GetResponseData(t, &body)

		assert.NotEmpty(t, body.FlowID)
	})

	t.Run("ShouldRejectConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		var saved bool

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSession(gomock.Any(), gomock.Any()).
			Return(nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), false).
			DoAndReturn(func(_ any, _ any, _ bool) error {
				saved = true

				return nil
			})

		userCode := mustGetTestOIDCUserCode(t, mock)

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":false,"user_code":"%s"}`, testOIDCDeviceCodeID, userCode))

		OAuth2ConsentPOST(mock.Ctx)

		body := oidc.ConsentPostResponseBody{}

		mock.GetResponseData(t, &body)

		require.True(t, saved)
		assert.NotEmpty(t, body.FlowID)
	})
}
