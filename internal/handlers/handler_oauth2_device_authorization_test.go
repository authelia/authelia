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

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
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
		assert.Equal(t, "no-store", rw.Header().Get(fasthttp.HeaderCacheControl))
		assert.Equal(t, "no-cache", rw.Header().Get(fasthttp.HeaderPragma))

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_request", response["error"])

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), regexpDeviceAuthorizationIssuerError, "missing required X-Forwarded-Host header")
	})

	t.Run("ShouldHandleInvalidClientBasicAuthWithChallenge", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPost, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterScope: []string{oidc.ScopeOpenID},
		})

		r.SetBasicAuth(testOIDCDeviceCodeID, "not-the-secret")

		OAuth2DeviceAuthorizationPOST(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusUnauthorized, rw.Code)
		assert.NotEmpty(t, rw.Header().Get(fasthttp.HeaderWWWAuthenticate))
		assert.Equal(t, "no-store", rw.Header().Get(fasthttp.HeaderCacheControl))

		response := getTestOAuth2ErrorResponse(t, rw)

		assert.Equal(t, "invalid_client", response["error"])
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

func TestOAuth2DeviceAuthorizationPOSTVerificationURIComplete(t *testing.T) {
	testCases := []struct {
		name     string
		enable   bool
		client   bool
		expected bool
	}{
		{
			"ShouldOmitVerificationURICompleteByDefault",
			false,
			false,
			false,
		},
		{
			"ShouldIncludeVerificationURICompleteWhenEnabled",
			true,
			false,
			true,
		},
		{
			"ShouldIncludeVerificationURICompleteWhenEnabledForClient",
			false,
			true,
			true,
		},
		{
			"ShouldIncludeVerificationURICompleteWhenEnabledForBoth",
			true,
			true,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			config := newTestOIDCConfig(t)
			config.EnableDeviceCodeVerificationURIComplete = tc.enable

			client := newTestOIDCDeviceCodeClient(t)
			client.DeviceCodeVerificationURIComplete = tc.client

			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

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

			assert.NotEmpty(t, response["verification_uri"])

			if tc.expected {
				assert.NotEmpty(t, response["verification_uri_complete"])
			} else {
				assert.NotContains(t, response, "verification_uri_complete")
			}
		})
	}
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

		flowID := uuid.Must(uuid.NewRandom())

		bindTestOIDCDeviceConsentSession(t, mock, userCode, flowID)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{flowID.String()},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error to validate the consent session during the User Authorization Flow", regexpAnyError)
	})

	t.Run("ShouldHandleSubjectMismatch", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		setupTestOIDCUserDetails(t, mock)

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCDeviceConsentSession(t, mock, uuid.Must(uuid.NewRandom()), userCode)

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

		setupTestOIDCUserDetails(t, mock)

		client := newTestOIDCDeviceCodeClient(t)
		client.AuthorizationPolicy = "two_factor"

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{client}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCDeviceConsentSession(t, mock, uuid.Must(uuid.NewRandom()), userCode)

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

		consent := newTestOIDCDeviceConsentSession(t, mock, subject, userCode)

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

	t.Run("ShouldApproveAfterConsentAndRejectFlowIDReuse", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)
		otherUserCode := mustGetTestOIDCUserCode(t, mock)

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":true,"user_code":"%s"}`, testOIDCDeviceCodeID, userCode))

		OAuth2ConsentPOST(mock.Ctx)

		body := oidc.ConsentPostResponseBody{}

		mock.GetResponseData(t, &body)

		require.NotEmpty(t, body.FlowID)

		rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{userCode},
			oidc.FormParameterFlowID:   []string{body.FlowID},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		require.Equal(t, http.StatusOK, rw.Code)

		rw, r = newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
			oidc.FormParameterUserCode: []string{otherUserCode},
			oidc.FormParameterFlowID:   []string{body.FlowID},
		})

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		assert.Equal(t, http.StatusBadRequest, rw.Code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error to validate the consent session during the User Authorization Flow", "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified. The provided flow id does not match the consent session for the device code session.")
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

		consent = newTestOIDCDeviceConsentSession(t, mock, subject, userCode)

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

		mock.UserProviderMock.EXPECT().
			GetDetailsExtended(gomock.Eq(testUsername)).
			Return(nil, fmt.Errorf("error in backend"))

		rw, r := newRequest(t, userCode, consent)

		OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed as an error occurred fetching the user details", regexpAnyError)
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

		consent := newTestOIDCDeviceConsentSession(t, mock, uuid.Must(uuid.NewRandom()), userCode)

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

		setupTestOIDCUserDetails(t, mock)

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		consent := newTestOIDCDeviceConsentSession(t, mock, uuid.Must(uuid.NewRandom()), userCode)

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

		consent := newTestOIDCDeviceConsentSession(t, mock, subject, userCode)

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

		consent := newTestOIDCDeviceConsentSession(t, mock, subject, userCode)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			AnyTimes().
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
			Return(nil)

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

func TestOAuth2DeviceAuthorizationPUTConsentValidation(t *testing.T) {
	testCases := []struct {
		name   string
		bind   func(consent *model.OAuth2ConsentSession) uuid.UUID
		setup  func(mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession)
		status int
		err    string
	}{
		{
			name: "ShouldRejectDeviceCodeWithoutFlowID",
			bind: func(consent *model.OAuth2ConsentSession) uuid.UUID {
				return uuid.Nil
			},
			status: http.StatusBadRequest,
			err:    "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified. The provided flow id does not match the consent session for the device code session.",
		},
		{
			name: "ShouldRejectFlowIDOfAnotherDeviceCode",
			bind: func(consent *model.OAuth2ConsentSession) uuid.UUID {
				return uuid.Must(uuid.NewRandom())
			},
			status: http.StatusBadRequest,
			err:    "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified. The provided flow id does not match the consent session for the device code session.",
		},
		{
			name: "ShouldRejectClientMismatch",
			setup: func(mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) {
				consent.ClientID = testOIDCAuthorizationCodeID
			},
			status: http.StatusBadRequest,
			err:    "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified. The requested client id 'device-code' does not match the requested client id 'authorization-code' from the consent session.",
		},
		{
			name: "ShouldRejectScopeMismatch",
			setup: func(mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) {
				consent.RequestedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID, oidc.ScopeProfile}
			},
			status: http.StatusBadRequest,
			err:    "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Make sure that the various parameters are correct, be aware of case sensitivity and trim your parameters. Make sure that the client you are using has exactly whitelisted the redirect_uri you specified. The requested scope 'openid' does not match the requested scope 'openid profile' from the consent session.",
		},
		{
			name: "ShouldRejectGranted",
			setup: func(mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) {
				consent.Granted = true
			},
			status: http.StatusBadRequest,
			err:    "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Could not perform consent. The consent session has already been granted, has expired, or otherwise does not appear to be valid for the authorization request. The consent session has already been granted or has expired.",
		},
		{
			name: "ShouldRejectExpired",
			setup: func(mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) {
				consent.ExpiresAt = mock.Ctx.GetClock().Now().Add(-time.Second)
			},
			status: http.StatusBadRequest,
			err:    "The request is missing a required parameter, includes an invalid parameter value, includes a parameter more than once, or is otherwise malformed. Could not perform consent. The consent session has already been granted, has expired, or otherwise does not appear to be valid for the authorization request. The consent session has already been granted or has expired.",
		},
		{
			name: "ShouldRejectNotResponded",
			setup: func(mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) {
				consent.RespondedAt = sql.NullTime{}
			},
			status: http.StatusForbidden,
			err:    "The resource owner or authorization server denied the request. Make sure that the request you are making is valid. Maybe the credential or request parameters you are using are limited in scope or otherwise restricted. The consent session was not authorized by the user.",
		},
		{
			name: "ShouldRejectDenied",
			setup: func(mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) {
				consent.Authorized = false
				consent.GrantedScopes = nil
			},
			status: http.StatusForbidden,
			err:    "The resource owner or authorization server denied the request. Make sure that the request you are making is valid. Maybe the credential or request parameters you are using are limited in scope or otherwise restricted. The consent session was not authorized by the user.",
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
			setupTestOIDCSubjectStore(t, mock)

			userCode := mustGetTestOIDCUserCode(t, mock)

			subject := mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID)

			consent := newTestOIDCDeviceConsentSession(t, mock, subject, userCode)

			if tc.bind != nil {
				bindTestOIDCDeviceConsentSession(t, mock, userCode, tc.bind(consent))
			}

			if tc.setup != nil {
				tc.setup(mock, consent)
			}

			mock.StorageMock.EXPECT().
				LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
				AnyTimes().
				Return(consent, nil)

			rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
				oidc.FormParameterUserCode: []string{userCode},
				oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
			})

			OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

			assert.Equal(t, tc.status, rw.Code)

			AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed with error to validate the consent session during the User Authorization Flow", tc.err)
		})
	}
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

func newTestOIDCDeviceConsentSession(t *testing.T, mock *mocks.MockAutheliaCtx, subject uuid.UUID, userCode string) *model.OAuth2ConsentSession {
	t.Helper()

	now := mock.Ctx.GetClock().Now()

	form := url.Values{
		oidc.FormParameterClientID: []string{testOIDCDeviceCodeID},
		oidc.FormParameterScope:    []string{oidc.ScopeOpenID},
	}

	consent := &model.OAuth2ConsentSession{
		ID:              1,
		ChallengeID:     uuid.Must(uuid.NewRandom()),
		ClientID:        testOIDCDeviceCodeID,
		Subject:         uuid.NullUUID{UUID: subject, Valid: subject != uuid.Nil},
		Authorized:      true,
		RequestedAt:     now.Add(-time.Minute),
		ExpiresAt:       now.Add(10 * time.Second),
		RespondedAt:     sql.NullTime{Time: now, Valid: true},
		Form:            form.Encode(),
		RequestedScopes: model.StringSlicePipeDelimited{oidc.ScopeOpenID},
		GrantedScopes:   model.StringSlicePipeDelimited{oidc.ScopeOpenID},
	}

	bindTestOIDCDeviceConsentSession(t, mock, userCode, consent.ChallengeID)

	return consent
}

func bindTestOIDCDeviceConsentSession(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, challengeID uuid.UUID) {
	t.Helper()

	signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(mock.Ctx, userCode)
	require.NoError(t, err)

	device, err := mock.StorageMock.LoadOAuth2DeviceCodeSessionByUserCode(mock.Ctx, signature)
	require.NoError(t, err)

	device.ChallengeID = uuid.NullUUID{UUID: challengeID, Valid: challengeID != uuid.Nil}

	require.NoError(t, mock.StorageMock.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))
}
