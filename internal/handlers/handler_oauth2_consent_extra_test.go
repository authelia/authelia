// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

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

		handleOAuth2ConsentFlowIDGET(mock.Ctx, []byte(uuid.Must(uuid.NewRandom()).String()), false)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching user session during the Consent Flow stage of the Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleMalformedFormOnConsentSession", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))
		consent.Form = "%zz"

		setupTestOIDCConsent(t, mock, consent)
		setupTestOIDCSubjectStore(t, mock)

		consent.Subject.UUID = mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID)

		handleOAuth2ConsentFlowIDGET(mock.Ctx, []byte(consent.ChallengeID.String()), false)

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

		setupTestOIDCUserDetails(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))
		consent.Form = "%zz"

		setupTestOIDCConsent(t, mock, consent)
		setupTestOIDCSubjectStore(t, mock)

		consent.Subject.UUID = mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(nil, sql.ErrNoRows)

		mock.Ctx.Request.SetBodyString(newBody(consent, true, false))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to obtain the request form from the consent session", regexpAnyError)
	})

	t.Run("ShouldHandleSubjectLookupError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Nil)

		setupTestOIDCConsent(t, mock, consent)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(nil, sql.ErrNoRows)

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

		setupTestOIDCUserDetails(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		form := url.Values{
			oidc.FormParameterClientID:   []string{testOIDCAuthorizationCodeID},
			oidc.FormParameterRequestURI: []string{"urn:ietf:params:oauth:request_uri:not-a-real-request"},
		}

		consent.Form = form.Encode()

		setupTestOIDCConsent(t, mock, consent)
		setupTestOIDCSubjectStore(t, mock)

		consent.Subject.UUID = mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(nil, sql.ErrNoRows)

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

		setupTestOIDCUserDetails(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		form := url.Values{
			oidc.FormParameterClientID: []string{testOIDCAuthorizationCodeID},
			oidc.FormParameterPrompt:   []string{oidc.PromptLogin},
		}

		consent.Form = form.Encode()

		consent.RequestedAt = mock.Ctx.GetClock().Now().Add(time.Hour)

		setupTestOIDCConsent(t, mock, consent)
		setupTestOIDCSubjectStore(t, mock)

		consent.Subject.UUID = mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(nil, sql.ErrNoRows)

		mock.Ctx.Request.SetBodyString(newBody(consent, true, false))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "The authorization request requires the user performs a login even prior to providing consent", nil)
	})

	t.Run("ShouldIgnorePreConfigureWhenFormRequiresExplicitConsent", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		client := newTestOIDCPreConfiguredClient(t)

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCSubjectStore(t, mock)

		consent := newTestOIDCConsentSession(t, mock, mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID))

		form := url.Values{
			oidc.FormParameterClientID: []string{testOIDCAuthorizationCodeID},
			oidc.FormParameterPrompt:   []string{oidc.PromptConsent},
		}

		consent.Form = form.Encode()

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(nil, sql.ErrNoRows)

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

		setupTestOIDCUserDetails(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		setupTestOIDCConsent(t, mock, consent)
		setupTestOIDCSubjectStore(t, mock)

		consent.Subject.UUID = mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(nil, sql.ErrNoRows)

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

		setupTestOIDCUserDetails(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		setupTestOIDCConsent(t, mock, consent)
		setupTestOIDCSubjectStore(t, mock)

		consent.Subject.UUID = mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(nil, sql.ErrNoRows)

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

		setupTestOIDCUserDetails(t, mock)

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

		setupTestOIDCUserDetails(t, mock)

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

		_, _, consent, client, handled := handleOAuth2ConsentGetSessionsAndClient(mock.Ctx, uuid.Must(uuid.NewRandom()))

		assert.True(t, handled)
		assert.Nil(t, consent)
		assert.Nil(t, client)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching user session during the Consent Flow stage of the Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleDeviceAuthorizationPOSTSessionError", func(t *testing.T) {
		mock := newMockWithoutSessionProvider(t)
		defer mock.Close()

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","flow_id":"%s","client_id":"%s","consent":true}`, uuid.Must(uuid.NewRandom()), testOIDCDeviceCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching user session during the Consent Flow stage of the Authorization Flow", regexpAnyError)
	})
}

func TestHandleOAuth2ConsentGetSessionsAndClientSubject(t *testing.T) {
	t.Run("ShouldReturnConsentSessionWithSubjectOfUser", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		expected := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		setupTestOIDCConsent(t, mock, expected)
		setupTestOIDCSubjectStore(t, mock)

		expected.Subject.UUID = mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID)

		_, _, consent, client, handled := handleOAuth2ConsentGetSessionsAndClient(mock.Ctx, expected.ChallengeID)

		assert.False(t, handled)
		assert.Equal(t, expected, consent)
		assert.NotNil(t, client)
	})

	t.Run("ShouldReturnConsentSessionWithoutSubject", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		expected := newTestOIDCConsentSession(t, mock, uuid.Nil)

		setupTestOIDCConsent(t, mock, expected)

		_, _, consent, client, handled := handleOAuth2ConsentGetSessionsAndClient(mock.Ctx, expected.ChallengeID)

		assert.False(t, handled)
		assert.Equal(t, expected, consent)
		assert.NotNil(t, client)
	})

	t.Run("ShouldRejectConsentSessionWithSubjectOfAnotherUser", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		expected := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		setupTestOIDCConsent(t, mock, expected)
		setupTestOIDCSubjectStore(t, mock)

		_, _, consent, client, handled := handleOAuth2ConsentGetSessionsAndClient(mock.Ctx, expected.ChallengeID)

		assert.True(t, handled)
		assert.Nil(t, consent)
		assert.Nil(t, client)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred performing consent during the Consent Flow stage of the Authorization Flow as the subject of the consent session does not match the subject of the user", nil)
	})

	t.Run("ShouldHandleSubjectLookupError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		expected := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		setupTestOIDCConsent(t, mock, expected)

		mock.StorageMock.EXPECT().
			LoadUserOpaqueIdentifierBySignature(gomock.Any(), gomock.Any(), gomock.Any(), testUsername).
			Return(nil, fmt.Errorf("error in db"))

		_, _, consent, client, handled := handleOAuth2ConsentGetSessionsAndClient(mock.Ctx, expected.ChallengeID)

		assert.True(t, handled)
		assert.Nil(t, consent)
		assert.Nil(t, client)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred performing consent during the Consent Flow stage of the Authorization Flow as an error occurred determining the subject of the user", regexpAnyError)
	})
}

func TestOAuth2ConsentPOSTSubflowDispatch(t *testing.T) {
	t.Run("ShouldFallBackToFlowIDForUnknownSubflow", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		setupTestOIDCConsent(t, mock, consent)
		setupTestOIDCSubjectStore(t, mock)

		consent.Subject.UUID = mustGetTestOIDCSubject(t, mock, testOIDCAuthorizationCodeID)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(nil, sql.ErrNoRows)

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

		setupTestOIDCUserDetails(t, mock)

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

		setupTestOIDCUserDetails(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Nil)

		setupTestOIDCConsent(t, mock, consent)
		setupTestOIDCSubjectStore(t, mock)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(nil, sql.ErrNoRows)

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

	setupTestOIDCUserDetails(t, mock)

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
