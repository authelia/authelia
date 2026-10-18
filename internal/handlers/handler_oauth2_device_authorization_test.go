// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestOAuth2DeviceAuthorizationPOST(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID: []string{testOIDCDeviceCodeID},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_request", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpDeviceAuthorizationIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleUnknownClient", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:        []string{"not-a-client"},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusUnauthorized, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error during the Device Authorization Flow", regexpErrorClientAuthenticationFailed)
	})

	t.Run("ShouldHandleClientWithoutDeviceCodeGrant", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:        []string{testOIDCAuthorizationCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error during the Device Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldIssueDeviceCode", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByUserCode(gomock.Any(), gomock.Any()).
			Return(nil, sql.ErrNoRows)

		mock.StorageMock.EXPECT().
			SaveOAuth2DeviceCodeSession(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:        []string{testOIDCDeviceCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.NotEmpty(t, response["device_code"])
		assert.NotEmpty(t, response["user_code"])
		assert.NotEmpty(t, response["verification_uri"])
	})
}

func TestOAuth2DeviceAuthorizationPUT(t *testing.T) {
	t.Run("ShouldHandleBadIssuer", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		setupTestOIDCProvider(t, mock, nil)

		clearForwardedHeaders(mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, nil)

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpDeviceAuthorizationUserFlowIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleUnknownUserCode", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByUserCode(gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(nil, sql.ErrNoRows)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{"ABCDEFGH"},
			oidc.FormParameterFlowID:   []string{uuid.Must(uuid.NewRandom()).String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleMalformedFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{"not-a-uuid"},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error to parse the flow ID during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleUnknownConsentSession", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCConsentStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{uuid.Must(uuid.NewRandom()).String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error to load the consent session during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSubjectMismatch", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCUserRefresh(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to match the session to the subject during the User Authorization Flow", nil)
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
		setupTestOIDCUserRefresh(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed as the user did not satisfy the client authorization policy during the User Authorization Flow", nil)
	})

	t.Run("ShouldApproveUserAuthorization", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		subject := mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID)

		consent := newTestOIDCDeviceConsentSession(t, mock, userCode, subject)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Times(1).
			Return(nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)
	})
}

func TestOAuth2DeviceAuthorizationPUTExtra(t *testing.T) {
	setupPUT := func(t *testing.T, mutate func(client *schema.IdentityProvidersOpenIDConnectClient)) (mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
		t.Helper()

		mock = mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))

		client := newTestOIDCDeviceCodeClient(t)

		if mutate != nil {
			mutate(&client)
		}

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode = mustGetTestOIDCUserCode(t, mock)

		subject := mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID)

		consent = newTestOIDCDeviceConsentSession(t, mock, userCode, subject)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		return mock, userCode, consent
	}

	newRequest := func(t *testing.T, userCode string, consent *model.OAuth2ConsentSession) (*httptest.ResponseRecorder, *http.Request) {
		t.Helper()

		return newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})
	}

	t.Run("ShouldHandleUserDetailsError", func(t *testing.T) {
		mock, userCode, consent := setupPUT(t, nil)
		defer mock.Close()

		setupTestOIDCUserRefresh(t, mock)

		mock.UserProviderMock.EXPECT().
			GetDetailsExtended(gomock.Eq(testUsername)).
			Return(nil, fmt.Errorf("error in backend"))

		rw, r := newRequest(t, userCode, consent)

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to obtain the user details during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSaveConsentSessionGrantedError", func(t *testing.T) {
		mock, userCode, consent := setupPUT(t, nil)
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Return(fmt.Errorf("error in db"))

		rw, r := newRequest(t, userCode, consent)

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request had an error while saving the session during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldMergeAudienceWhenClaimsStrategyRequires", func(t *testing.T) {
		mock, userCode, consent := setupPUT(t, func(client *schema.IdentityProvidersOpenIDConnectClient) {
			client.ClaimsPolicy = testOIDCClaimsPolicyMerged
		})

		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Return(nil)

		rw, r := newRequest(t, userCode, consent)

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusOK, rw.Code)
	})
}

func TestOAuth2DeviceAuthorizationErrorPaths(t *testing.T) {
	t.Run("ShouldHandleDeviceAuthorizeResponseError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		mock.StorageMock.EXPECT().
			LoadOAuth2DeviceCodeSessionByUserCode(gomock.Any(), gomock.Any()).
			Return(nil, sql.ErrNoRows)

		mock.StorageMock.EXPECT().
			SaveOAuth2DeviceCodeSession(gomock.Any(), gomock.Any()).
			Return(fmt.Errorf("error in db"))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:        []string{testOIDCDeviceCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request had an error while trying to create a response during the Device Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSessionProviderError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.Ctx.Request.Header.Set("X-Original-URL", "https://unknown.example.org/api/oidc/device-authorization")

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to obtain the user session during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSubjectLookupError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCUserRefresh(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			LoadUserOpaqueIdentifierBySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, fmt.Errorf("error in db"))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to obtain the user subject value for the user during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleMalformedClaimsParameter", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		rwd, rd := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterClientID:        []string{testOIDCDeviceCodeID},
			testOIDCFormParameterClientSecret: []string{testOIDCClientSecretValue},
			oidc.FormParameterScope:           []string{oidc.ScopeOpenID},
			oidc.FormParameterClaims:          []string{"not json"},
		})

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rwd, rd)

		require.Equal(t, http.StatusOK, rwd.Code)

		userCode, ok := getTestOAuth2ErrorResponse(t, rwd)[oidc.FormParameterUserCode].(string)

		require.True(t, ok)

		subject := mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID)

		consent := newTestOIDCDeviceConsentSession(t, mock, userCode, subject)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpDeviceAuthorizationClaimsParseError, regexpAnyError)
	})

	t.Run("ShouldHandleUserAuthorizeResponseError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		subject := mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID)

		consent := newTestOIDCDeviceConsentSession(t, mock, userCode, subject)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2Session(gomock.Any(), gomock.Any(), gomock.Any()).
			AnyTimes().
			Return(fmt.Errorf("error in db"))

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.NotEqual(t, http.StatusOK, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request had an error while attempting to generate the responder during the User Authorization Flow", regexpAnyError)
	})
}

func TestOAuth2DeviceAuthorizationPUTConsent(t *testing.T) {
	testCases := []struct {
		name     string
		mutate   func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession)
		expected string
		err      any
		code     int
	}{
		{
			"ShouldApprove",
			nil,
			"",
			nil,
			http.StatusOK,
		},
		{
			"ShouldNotApproveWithNullSubject",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				consent.Subject.Valid = false
			},
			"Device Authorization Request failed to match the session to the subject during the User Authorization Flow",
			nil,
			http.StatusInternalServerError,
		},
		{
			"ShouldNotApproveWithAnotherClient",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				consent.ClientID = testOIDCAuthorizationCodeID
			},
			"Device Authorization Request failed to match the consent session to the request during the User Authorization Flow",
			regexpAnyError,
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWithAnotherSubjectOnTheDeviceCodeSession",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)
				device.Subject = sql.NullString{String: uuid.Must(uuid.NewRandom()).String(), Valid: true}

				require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))
			},
			"Device Authorization Request failed to match the device code session to the consent session which claimed it during the User Authorization Flow",
			regexpAnyError,
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWhenTheDeviceCodeSessionHasNoSubject",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)
				device.Subject = sql.NullString{}

				require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))
			},
			"Device Authorization Request failed to match the device code session to the consent session which claimed it during the User Authorization Flow",
			regexpAnyError,
			http.StatusBadRequest,
		},
		{
			"ShouldApproveWhenTheDevicePolledAfterTheConsentFlow",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				// The token endpoint holds the device code session as it was before the consent flow bound it.
				device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)
				device.ChallengeID, device.Subject = uuid.NullUUID{}, sql.NullString{}

				request, err := device.ToRequest(mock.Ctx, oidc.NewSession(), mock.Ctx.Providers.OpenIDConnect.Store)
				require.NoError(t, err)

				request.SetLastChecked(time.Now())

				require.NoError(t, mock.Ctx.Providers.OpenIDConnect.UpdateDeviceCodeSession(mock.Ctx, device.Signature, request))

				device = mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

				assert.Equal(t, consent.ChallengeID, device.ChallengeID.UUID)
				assert.True(t, device.Subject.Valid)
			},
			"",
			nil,
			http.StatusOK,
		},
		{
			"ShouldNotApproveWhenTheDeviceCodeSessionWasNotClaimed",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)
				device.Status = int(oauthelia2.DeviceAuthorizeStatusNew)

				require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))
			},
			"Device Authorization Request failed to match the device code session to the consent session which claimed it during the User Authorization Flow",
			"the device code session does not have the expected status: expected 'Claimed' but it's 'New'",
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWhenTheDeviceCodeSessionHasNoFlow",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)
				device.ChallengeID = uuid.NullUUID{}

				require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))
			},
			"Device Authorization Request failed to match the device code session to the consent session which claimed it during the User Authorization Flow",
			regexpAnyError,
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWhenTheDeviceCodeSessionHasAnotherFlow",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)
				device.ChallengeID = uuid.NullUUID{UUID: uuid.Must(uuid.NewRandom()), Valid: true}

				require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))
			},
			"Device Authorization Request failed to match the device code session to the consent session which claimed it during the User Authorization Flow",
			regexpAnyError,
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWhenTheConsentSessionIsNotForTheRequestedFlow",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				consent.ChallengeID = uuid.Must(uuid.NewRandom())
			},
			"Device Authorization Request failed to match the device code session to the consent session which claimed it during the User Authorization Flow",
			regexpAnyError,
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWithDifferentScopes",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				consent.RequestedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeProfile}
			},
			"Device Authorization Request failed to match the consent session to the request during the User Authorization Flow",
			regexpAnyError,
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWithDifferentAudience",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				consent.RequestedAudience = model.StringSlicePipeDelimited{"https://app.example.com"}
			},
			"Device Authorization Request failed to match the consent session to the request during the User Authorization Flow",
			regexpAnyError,
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWithDifferentResource",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				consent.RequestedResource = model.StringSlicePipeDelimited{"https://app.example.com"}
			},
			"Device Authorization Request failed to match the consent session to the request during the User Authorization Flow",
			regexpAnyError,
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWhenExpired",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				consent.ExpiresAt = mock.Ctx.GetClock().Now().Add(-time.Second)
			},
			"Device Authorization Request failed as the consent session has either expired or already been granted during the User Authorization Flow",
			nil,
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWhenAlreadyGranted",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				consent.Granted = true
			},
			"Device Authorization Request failed as the consent session has either expired or already been granted during the User Authorization Flow",
			nil,
			http.StatusBadRequest,
		},
		{
			"ShouldNotApproveWhenDenied",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				consent.Authorized = false
			},
			"Device Authorization Request failed as the consent session was not authorized by the user during the User Authorization Flow",
			nil,
			http.StatusForbidden,
		},
		{
			"ShouldNotApproveWhenNotResponded",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
				consent.RespondedAt = sql.NullTime{}
			},
			"Device Authorization Request failed as the consent session was not authorized by the user during the User Authorization Flow",
			nil,
			http.StatusForbidden,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			defer mock.Close()

			config := newTestOIDCConfig(t)
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

			setupTestOIDCProvider(t, mock, config)
			setupTestOIDCDeviceCodeStore(t, mock)
			setupTestOIDCSessionStore(t, mock)
			setupTestOIDCSubjectStore(t, mock)
			setupTestOIDCUserDetails(t, mock)

			userCode := mustGetTestOIDCUserCode(t, mock)

			consent := newTestOIDCDeviceConsentSession(t, mock, userCode, mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID))

			flowID := consent.ChallengeID

			if tc.mutate != nil {
				tc.mutate(t, mock, userCode, consent)
			}

			status := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).Status

			mock.StorageMock.EXPECT().
				LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), flowID).
				AnyTimes().
				Return(consent, nil)

			if tc.code == http.StatusOK {
				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
					Times(1).
					Return(nil)
			}

			rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
				oidc.FormParameterUserCode: []string{userCode},
				oidc.FormParameterFlowID:   []string{flowID.String()},
			})

			OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

			assert.Equal(t, tc.code, rw.Code)

			if tc.code == http.StatusOK {
				assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusApproved), mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).Status)
			} else {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), tc.expected, tc.err)

				assert.Equal(t, status, mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).Status)
			}
		})
	}
}

func TestOAuth2DeviceAuthorizationPUTUserSession(t *testing.T) {
	const policyAdmins = "admins"

	setup := func(t *testing.T, userSession *session.UserSession, policy string) (mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
		t.Helper()

		if userSession == nil {
			mock = mocks.NewMockAutheliaCtx(t)
		} else {
			mock = mocks.NewMockAutheliaCtxWithUserSession(t, *userSession)
		}

		client := newTestOIDCDeviceCodeClient(t)

		config := newTestOIDCConfig(t)
		config.AuthorizationPolicies = map[string]schema.IdentityProvidersOpenIDConnectPolicy{
			policyAdmins: {
				DefaultPolicy: "deny",
				Rules: []schema.IdentityProvidersOpenIDConnectPolicyRule{
					{Policy: "one_factor", Subjects: schema.AccessControlRuleSubjects{{"group:admins"}}},
				},
			},
		}

		if policy != "" {
			client.AuthorizationPolicy = policy
		}

		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode = mustGetTestOIDCUserCode(t, mock)

		consent = newTestOIDCDeviceConsentSession(t, mock, userCode, mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		return mock, userCode, consent
	}

	put := func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) int {
		t.Helper()

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		return rw.Code
	}

	details := func(groups ...string) *authentication.UserDetails {
		return &authentication.UserDetails{Username: testUsername, DisplayName: testDisplayName, Emails: []string{testEmail}, Groups: groups}
	}

	t.Run("ShouldRejectAnonymousUser", func(t *testing.T) {
		mock, userCode, consent := setup(t, nil, "")
		defer mock.Close()

		assert.Equal(t, http.StatusForbidden, put(t, mock, userCode, consent))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed as the user is anonymous during the User Authorization Flow", nil)

		assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusClaimed), mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).Status)
	})

	t.Run("ShouldRejectAndDestroySessionWhenUserNoLongerExists", func(t *testing.T) {
		userSession := newTestOIDCUserSession(1)

		mock, userCode, consent := setup(t, &userSession, "")
		defer mock.Close()

		mock.UserProviderMock.EXPECT().
			GetDetails(testUsername).
			Return(nil, authentication.ErrUserNotFound)

		assert.Equal(t, http.StatusForbidden, put(t, mock, userCode, consent))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed as the user is no longer authorized to login during the User Authorization Flow", nil)

		us, err := mock.Ctx.GetSession()
		require.NoError(t, err)

		assert.True(t, us.IsAnonymous())
		assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusClaimed), mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).Status)
	})

	t.Run("ShouldRejectWhenRefreshedGroupsNoLongerSatisfyThePolicy", func(t *testing.T) {
		userSession := newTestOIDCUserSession(1)
		userSession.Groups = []string{"admins"}

		mock, userCode, consent := setup(t, &userSession, policyAdmins)
		defer mock.Close()

		mock.UserProviderMock.EXPECT().
			GetDetails(testUsername).
			Return(details("dev"), nil)

		assert.Equal(t, http.StatusForbidden, put(t, mock, userCode, consent))

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed as the user is not authorized to use the client during the User Authorization Flow", nil)

		us, err := mock.Ctx.GetSession()
		require.NoError(t, err)

		assert.Equal(t, []string{"dev"}, us.Groups)
		assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusClaimed), mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).Status)
	})

	t.Run("ShouldApproveWhenRefreshedGroupsSatisfyThePolicy", func(t *testing.T) {
		userSession := newTestOIDCUserSession(1)

		mock, userCode, consent := setup(t, &userSession, policyAdmins)
		defer mock.Close()

		mock.UserProviderMock.EXPECT().
			GetDetails(testUsername).
			Return(details("admins"), nil)

		mock.UserProviderMock.EXPECT().
			GetDetailsExtended(testUsername).
			Return(&authentication.UserDetailsExtended{UserDetails: details("admins")}, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Return(nil)

		assert.Equal(t, http.StatusOK, put(t, mock, userCode, consent))

		us, err := mock.Ctx.GetSession()
		require.NoError(t, err)

		assert.Equal(t, []string{"admins"}, us.Groups)
		assert.True(t, us.RefreshTTL.After(mock.Ctx.GetClock().Now()))
		assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusApproved), mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).Status)
	})

	t.Run("ShouldNotRefreshWithinTheRefreshInterval", func(t *testing.T) {
		userSession := newTestOIDCUserSession(1)
		userSession.Groups = []string{"admins"}
		userSession.RefreshTTL = time.Now().Add(time.Minute)

		mock, userCode, consent := setup(t, &userSession, policyAdmins)
		defer mock.Close()

		mock.UserProviderMock.EXPECT().
			GetDetailsExtended(testUsername).
			Return(&authentication.UserDetailsExtended{UserDetails: details("admins")}, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Return(nil)

		assert.Equal(t, http.StatusOK, put(t, mock, userCode, consent))
	})
}

func newTestOIDCDeviceConsentSession(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, subject uuid.UUID) *model.OAuth2ConsentSession {
	t.Helper()

	device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

	request, err := device.ToRequest(mock.Ctx, oidc.NewSession(), mock.Ctx.Providers.OpenIDConnect.Store)
	require.NoError(t, err)

	consent, err := model.NewOAuth2ConsentSession(mock.Ctx.GetClock().Now().Add(time.Minute), subject, request)
	require.NoError(t, err)

	consent.ID = 1

	oidc.ConsentGrant(consent, true, nil)

	consent.SetRespondedAt(mock.Ctx.GetClock().Now(), 0)
	consent.Authorized = true

	device.ChallengeID = uuid.NullUUID{UUID: consent.ChallengeID, Valid: true}
	device.Subject = sql.NullString{String: subject.String(), Valid: true}
	device.Status = int(oauthelia2.DeviceAuthorizeStatusClaimed)

	require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))

	return consent
}

func mustLoadTestOIDCDeviceCodeSession(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string) *model.OAuth2DeviceCodeSession {
	t.Helper()

	signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(mock.Ctx, userCode)
	require.NoError(t, err)

	device, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(mock.Ctx, signature)
	require.NoError(t, err)

	return device
}

const testOIDCDeviceAuthorizationEndpoint = "https://login.example.com:8080/api/oidc/device-authorization"

func newTestOIDCDeviceCodeClient(t *testing.T) schema.IdentityProvidersOpenIDConnectClient {
	t.Helper()

	return schema.IdentityProvidersOpenIDConnectClient{
		ID:                      testOIDCDeviceCodeID,
		Secret:                  mustDecodeTestSecret(t, testOIDCClientSecretDigest),
		AuthorizationPolicy:     "one_factor",
		GrantTypes:              []string{oidc.GrantTypeDeviceCode},
		ResponseTypes:           []string{oidc.ResponseTypeAuthorizationCodeFlow},
		Scopes:                  []string{oidc.ScopeOpenID, oidc.ScopeProfile},
		RedirectURIs:            []string{testOIDCRedirectURI},
		TokenEndpointAuthMethod: oidc.ClientAuthMethodClientSecretPost,
	}
}
