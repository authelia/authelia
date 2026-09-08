package handlers

import (
	"database/sql"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestHandleOAuth2ConsentFlowIDGETExtra(t *testing.T) {
	t.Run("ShouldHandleSessionProviderError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCConsent(t, mock, nil)

		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "unknown.example.org")

		handleOAuth2ConsentFlowIDGET(mock.Ctx, []byte(uuid.Must(uuid.NewRandom()).String()))

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching user session", regexpAnyError)
	})

	t.Run("ShouldHandleMalformedFormOnConsentSession", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))
		consent.Form = "%zz"

		setupTestOIDCConsent(t, mock, consent)

		handleOAuth2ConsentFlowIDGET(mock.Ctx, []byte(consent.ChallengeID.String()))

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred getting form from consent session", regexpAnyError)
	})
}

func TestHandleOAuth2ConsentFlowIDPOSTExtra(t *testing.T) {
	newBody := func(consent *model.OAuth2ConsentSession, grant, preConfigure bool) string {
		return fmt.Sprintf(`{"flow_id":"%s","client_id":"%s","consent":%t,"pre_configure":%t}`, consent.ChallengeID, testOIDCAuthorizationCodeID, grant, preConfigure)
	}

	t.Run("ShouldHandleMalformedForm", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))
		consent.Form = "%zz"

		setupTestOIDCConsent(t, mock, consent)

		mock.Ctx.Request.SetBodyString(newBody(consent, true, false))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to obtain the request form from the consent session", regexpAnyError)
	})

	t.Run("ShouldHandleSubjectLookupError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		consent := newTestOIDCConsentSession(t, mock, uuid.Nil)

		setupTestOIDCConsent(t, mock, consent)

		mock.StorageMock.EXPECT().
			LoadUserOpaqueIdentifierBySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("error in db"))

		mock.Ctx.Request.SetBodyString(newBody(consent, true, false))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to determine the subject for the consent session", regexpAnyError)
	})

	t.Run("ShouldHandlePushedAuthorizationRequestLookupError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		form := url.Values{
			oidc.FormParameterClientID:   []string{testOIDCAuthorizationCodeID},
			oidc.FormParameterRequestURI: []string{"urn:ietf:params:oauth:request_uri:not-a-real-request"},
		}

		consent.Form = form.Encode()

		setupTestOIDCConsent(t, mock, consent)

		mock.StorageMock.EXPECT().
			LoadOAuth2PushedAuthorizationSession(gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("error in db"))

		mock.Ctx.Request.SetBodyString(newBody(consent, true, false))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to obtain the actual authorization parameters from the request form", regexpAnyError)
	})

	t.Run("ShouldHandleFormRequiringLogin", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		form := url.Values{
			oidc.FormParameterClientID: []string{testOIDCAuthorizationCodeID},
			oidc.FormParameterPrompt:   []string{oidc.PromptLogin},
		}

		consent.Form = form.Encode()

		consent.RequestedAt = mock.Ctx.GetClock().Now().Add(time.Hour)

		setupTestOIDCConsent(t, mock, consent)

		mock.Ctx.Request.SetBodyString(newBody(consent, true, false))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "The authorization request requires the user performs a login even prior to providing consent", nil)
	})

	t.Run("ShouldIgnorePreConfigureWhenFormRequiresExplicitConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCPreConfiguredClient(t)

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		form := url.Values{
			oidc.FormParameterClientID: []string{testOIDCAuthorizationCodeID},
			oidc.FormParameterPrompt:   []string{oidc.PromptConsent},
		}

		consent.Form = form.Encode()

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), true).
			Return(nil)

		mock.Ctx.Request.SetBodyString(newBody(consent, true, true))

		OAuth2ConsentPOST(mock.Ctx)

		body := oidc.ConsentPostResponseBody{}

		mock.GetResponseData(t, &body)

		assert.NotEmpty(t, body.RedirectURI)

		AssertLogEntryMessageAndError(t, MustGetLogLastSeq(t, mock.Hook, 0), "Ignored saving pre-configuration as it is not permitted due to constraints within the authorization request form", nil)
	})

	t.Run("ShouldHandleSaveConsentSessionResponseError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		setupTestOIDCConsent(t, mock, consent)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), true).
			Return(fmt.Errorf("error in db"))

		mock.Ctx.Request.SetBodyString(newBody(consent, true, false))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred saving the consent session response to the database", regexpAnyError)
	})

	t.Run("ShouldHandleIssuerError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		setupTestOIDCConsent(t, mock, consent)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), true).
			DoAndReturn(func(_, _ any, _ bool) error {
				clearForwardedHeaders(mock)

				return nil
			})

		mock.Ctx.Request.SetBodyString(newBody(consent, true, false))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to determine the issuer URL", "missing required X-Forwarded-Host header")
	})
}

func TestHandleSavePreConfiguredConsentExtra(t *testing.T) {
	t.Run("ShouldLogClaimsParseError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCPreConfiguredClient(t)

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		registered, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, testOIDCAuthorizationCodeID)

		require.NoError(t, err)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentPreConfiguration(gomock.Any(), gomock.Any()).
			Return(int64(5), nil)

		handleSavePreConfiguredConsent(mock.Ctx, newTestOIDCUserSession(1), consent, registered, url.Values{oidc.FormParameterClaims: []string{"not json"}}, nil)

		assert.Equal(t, sql.NullInt64{Int64: 5, Valid: true}, consent.PreConfiguration)

		AssertLogEntryMessageAndError(t, MustGetLogLastSeq(t, mock.Hook, 0), "Error occurred parsing request form claims parameter", regexpAnyError)
	})

	t.Run("ShouldSerializeClaimsRequests", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCPreConfiguredClient(t)

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		registered, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, testOIDCAuthorizationCodeID)

		require.NoError(t, err)

		var saved model.OAuth2ConsentPreConfig

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentPreConfiguration(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, config model.OAuth2ConsentPreConfig) (int64, error) {
				saved = config

				return int64(7), nil
			})

		handleSavePreConfiguredConsent(mock.Ctx, newTestOIDCUserSession(1), consent, registered, url.Values{oidc.FormParameterClaims: []string{`{"id_token":{"email":null}}`}}, []string{"email"})

		assert.Equal(t, sql.NullInt64{Int64: 7, Valid: true}, consent.PreConfiguration)

		assert.True(t, saved.RequestedClaims.Valid)
		assert.True(t, saved.SignatureClaims.Valid)
		assert.Equal(t, model.StringSlicePipeDelimited{"email"}, saved.GrantedClaims)
		assert.Equal(t, mock.Ctx.GetClock().Now().Add(testOIDCPreConfiguredDuration).Unix(), saved.ExpiresAt.Time.Unix())
	})
}

func TestHandleOAuth2ConsentDeviceAuthorizationPOSTExtra(t *testing.T) {
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

	body := func(userCode string) string {
		return fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":true,"user_code":"%s"}`, testOIDCDeviceCodeID, userCode)
	}

	t.Run("ShouldHandleRevokedDeviceCodeSession", func(t *testing.T) {
		mock, userCode := setup(t, func(device *model.OAuth2DeviceCodeSession) {
			device.Revoked = true
		})
		defer mock.Close()

		mock.Ctx.Request.SetBodyString(body(userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to determine if the device code session is active during the Consent Flow stage of the Device Authorization Flow", nil)
	})

	t.Run("ShouldHandleDeviceCodeSessionWithExistingChallenge", func(t *testing.T) {
		mock, userCode := setup(t, func(device *model.OAuth2DeviceCodeSession) {
			device.ChallengeID = uuid.NullUUID{UUID: uuid.Must(uuid.NewRandom()), Valid: true}
		})
		defer mock.Close()

		mock.Ctx.Request.SetBodyString(body(userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to advance the Consent Flow stage of the Device Authorization Flow as the device code session already has a challenge id", nil)
	})

	t.Run("ShouldHandleInsufficientAuthenticationLevel", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		client := newTestOIDCDeviceCodeClient(t)

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		client.AuthorizationPolicy = "two_factor"

		configTwoFactor := newTestOIDCConfig(t)
		configTwoFactor.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, configTwoFactor)

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetBodyString(body(userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "User is not sufficiently authenticated to provide consent given the client authorization policy during the Consent Flow stage of the Device Authorization Flow", nil)
	})

	t.Run("ShouldHandleSubjectLookupError", func(t *testing.T) {
		mock, userCode := setup(t, nil)
		defer mock.Close()

		mock.StorageMock.EXPECT().
			LoadUserOpaqueIdentifierBySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("error in db"))

		mock.Ctx.Request.SetBodyString(body(userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to determine the subject during the Consent Flow stage of the Device Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSubjectMismatch", func(t *testing.T) {
		mock, userCode := setup(t, func(device *model.OAuth2DeviceCodeSession) {
			device.Subject = sql.NullString{String: uuid.Must(uuid.NewRandom()).String(), Valid: true}
		})
		defer mock.Close()

		setupTestOIDCSubjectStore(t, mock)

		mock.Ctx.Request.SetBodyString(body(userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to determine the subject during the Consent Flow stage of the Device Authorization Flow as the subject of the device code session does not match the subject of the user session", nil)
	})

	t.Run("ShouldHandleSaveConsentSessionError", func(t *testing.T) {
		mock, userCode := setup(t, nil)
		defer mock.Close()

		setupTestOIDCSubjectStore(t, mock)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSession(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("error in db"))

		mock.Ctx.Request.SetBodyString(body(userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred saving the consent session to the database during the Consent Flow stage of the Device Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSaveConsentSessionResponseError", func(t *testing.T) {
		mock, userCode := setup(t, nil)
		defer mock.Close()

		setupTestOIDCSubjectStore(t, mock)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSession(gomock.Any(), gomock.Any()).
			Return(nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("error in db"))

		mock.Ctx.Request.SetBodyString(body(userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred saving the consent session response to the database during the Consent Flow stage of the Device Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleUpdateDeviceCodeSessionError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSubjectStore(t, mock)

		var device *model.OAuth2DeviceCodeSession

		mock.StorageMock.EXPECT().
			SaveOAuth2DeviceCodeSession(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, session *model.OAuth2DeviceCodeSession) error {
				value := *session
				device = &value

				return nil
			})

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByUserCode(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ any, _ string) (*model.OAuth2DeviceCodeSession, error) {
				value := *device

				return &value, nil
			})

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSession(gomock.Any(), gomock.Any()).
			Return(nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		mock.StorageMock.EXPECT().
			UpdateOAuth2DeviceCodeSession(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("error in db"))

		userCode := mustGetTestOIDCUserCode(t, mock)

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetBodyString(body(userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred saving the device code session challenge id to the database during the Consent Flow stage of the Device Authorization Flow", regexpAnyError)
	})
}

func TestHandleOAuth2ConsentSessionProviderErrors(t *testing.T) {
	newMockWithoutSessionProvider := func(t *testing.T) *mocks.MockAutheliaCtx {
		t.Helper()

		mock := mocks.NewMockAutheliaCtx(t)

		setupTestOIDCProvider(t, mock, nil)

		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "unknown.example.org")

		return mock
	}

	t.Run("ShouldHandleGetSessionsAndClientSessionError", func(t *testing.T) {
		mock := newMockWithoutSessionProvider(t)
		defer mock.Close()

		_, consent, client, handled := handleOAuth2ConsentGetSessionsAndClient(mock.Ctx, uuid.Must(uuid.NewRandom()))

		assert.True(t, handled)
		assert.Nil(t, consent)
		assert.Nil(t, client)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching user session during the Consent Flow stage of the Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleDeviceGetSessionsAndClientSessionError", func(t *testing.T) {
		mock := newMockWithoutSessionProvider(t)
		defer mock.Close()

		_, device, client, handled := handleOAuth2ConsentDeviceAuthorizationGetSessionsAndClient(mock.Ctx, "ABCDEFGH")

		assert.True(t, handled)
		assert.Nil(t, device)
		assert.Nil(t, client)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching user session during the Consent Flow stage of the Device Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleDeviceAuthorizationPOSTSessionError", func(t *testing.T) {
		mock := newMockWithoutSessionProvider(t)
		defer mock.Close()

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":true,"user_code":"ABCDEFGH"}`, testOIDCDeviceCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching user session during the Consent Flow stage of the Device Authorization Flow", regexpAnyError)
	})
}

func TestHandleOAuth2ConsentDeviceUnregisteredClient(t *testing.T) {
	setup := func(t *testing.T) (mock *mocks.MockAutheliaCtx, userCode string) {
		t.Helper()

		mock = mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode = mustGetTestOIDCUserCode(t, mock)

		configWithoutClient := newTestOIDCConfig(t)
		configWithoutClient.Clients = nil

		setupTestOIDCProvider(t, mock, configWithoutClient)

		mock.Ctx.Response.Reset()

		return mock, userCode
	}

	t.Run("ShouldHandleUnregisteredClientOnGET", func(t *testing.T) {
		mock, userCode := setup(t)
		defer mock.Close()

		mock.Ctx.Request.SetRequestURI("/api/oidc/consent?user_code=" + userCode)

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred loading registered client using client id during the Consent Flow stage of the Device Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleUnregisteredClientOnPOST", func(t *testing.T) {
		mock, userCode := setup(t)
		defer mock.Close()

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":true,"user_code":"%s"}`, testOIDCDeviceCodeID, userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching client configuration during the Consent Flow stage of the Device Authorization Flow", regexpAnyError)
	})
}

func TestHandleOAuth2ConsentUseCodeGETExtra(t *testing.T) {
	t.Run("ShouldHandleMalformedFormOnDeviceCodeSession", func(t *testing.T) {
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

		device.Form = "%zz"

		require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetRequestURI("/api/oidc/consent?user_code=" + userCode)

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred getting form from device code session", regexpAnyError)
	})
}

func TestOAuth2ConsentPOSTSubflowDispatch(t *testing.T) {
	t.Run("ShouldFallBackToFlowIDForUnknownSubflow", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		setupTestOIDCConsent(t, mock, consent)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), true).
			Return(nil)

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"not-a-subflow","flow_id":"%s","client_id":"%s","consent":true}`, consent.ChallengeID, testOIDCAuthorizationCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		body := oidc.ConsentPostResponseBody{}

		mock.GetResponseData(t, &body)

		assert.NotEmpty(t, body.RedirectURI)
	})

	t.Run("ShouldStopWhenConsentSessionCannotBeResolved", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCConsent(t, mock, nil)

		flowID := uuid.Must(uuid.NewRandom())

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), flowID).
			Return(nil, fmt.Errorf("error in db"))

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flow_id":"%s","client_id":"%s","consent":true}`, flowID, testOIDCAuthorizationCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching consent session during the Consent Flow stage of the Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldResolveSubjectWhenConsentSessionHasNone", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		consent := newTestOIDCConsentSession(t, mock, uuid.Nil)

		setupTestOIDCConsent(t, mock, consent)
		setupTestOIDCSubjectStore(t, mock)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), true).
			Return(nil)

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flow_id":"%s","client_id":"%s","consent":true}`, consent.ChallengeID, testOIDCAuthorizationCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		body := oidc.ConsentPostResponseBody{}

		mock.GetResponseData(t, &body)

		assert.NotEmpty(t, body.RedirectURI)

		assert.True(t, consent.Subject.Valid)
		assert.NotEqual(t, uuid.Nil, consent.Subject.UUID)
	})
}

func TestHandleSavePreConfiguredConsentWithoutClaims(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	config := newTestOIDCConfig(t)
	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCPreConfiguredClient(t)}

	setupTestOIDCProvider(t, mock, config)

	consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

	registered, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, testOIDCAuthorizationCodeID)

	require.NoError(t, err)

	var saved model.OAuth2ConsentPreConfig

	mock.StorageMock.EXPECT().
		SaveOAuth2ConsentPreConfiguration(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ any, config model.OAuth2ConsentPreConfig) (int64, error) {
			saved = config

			return int64(3), nil
		})

	handleSavePreConfiguredConsent(mock.Ctx, newTestOIDCUserSession(1), consent, registered, url.Values{}, nil)

	assert.Equal(t, sql.NullInt64{Int64: 3, Valid: true}, consent.PreConfiguration)

	assert.False(t, saved.RequestedClaims.Valid)
	assert.False(t, saved.SignatureClaims.Valid)
	assert.Empty(t, saved.GrantedClaims)
}

func TestHandleGetConsentForm(t *testing.T) {
	t.Run("ShouldReturnOriginalFormWhenNotPushed", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		original := url.Values{oidc.FormParameterClientID: []string{testOIDCAuthorizationCodeID}}

		form, err := handleGetConsentForm(mock.Ctx, original)

		require.NoError(t, err)

		assert.Equal(t, original, form)
	})

	t.Run("ShouldReturnPushedRequestForm", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCPARStore(t, mock)

		requestURI := mustPushTestOIDCAuthorizationRequest(t, mock)

		form, err := handleGetConsentForm(mock.Ctx, url.Values{
			oidc.FormParameterClientID:   []string{testOIDCAuthorizationCodeID},
			oidc.FormParameterRequestURI: []string{requestURI},
		})

		require.NoError(t, err)

		assert.Equal(t, testOIDCAuthorizationCodeID, form.Get(oidc.FormParameterClientID))
		assert.Equal(t, testOIDCRedirectURI, form.Get(oidc.FormParameterRedirectURI))
		assert.Equal(t, oidc.ScopeOpenID, form.Get(oidc.FormParameterScope))
		assert.Empty(t, form.Get(oidc.FormParameterRequestURI))
	})

	t.Run("ShouldPropagatePushedRequestLookupError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		mock.StorageMock.EXPECT().
			LoadOAuth2PushedAuthorizationSession(gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("error in db"))

		form, err := handleGetConsentForm(mock.Ctx, url.Values{
			oidc.FormParameterRequestURI: []string{"urn:ietf:params:oauth:request_uri:not-a-real-request"},
		})

		assert.Nil(t, form)
		assert.Error(t, err)
	})
}

func TestHandleOAuth2ConsentDeviceAuthorizationPOSTCorruptSession(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	config := newTestOIDCConfig(t)
	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCDeviceCodeStore(t, mock)
	setupTestOIDCSubjectStore(t, mock)

	userCode := mustGetTestOIDCUserCode(t, mock)

	signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(mock.Ctx, userCode)

	require.NoError(t, err)

	device, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(mock.Ctx, signature)

	require.NoError(t, err)

	device.Session = []byte("not json")

	require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))

	mock.Ctx.Response.Reset()
	mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":true,"user_code":"%s"}`, testOIDCDeviceCodeID, userCode))

	OAuth2ConsentPOST(mock.Ctx)

	mock.Assert200KO(t, messageOperationFailed)

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to restore the requester during the Consent Flow stage of the Device Authorization Flow", regexpAnyError)
}

func setupTestOIDCConsent(t *testing.T, mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) {
	t.Helper()

	config := newTestOIDCConfig(t)
	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

	setupTestOIDCProvider(t, mock, config)

	if consent != nil {
		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)
	}
}
