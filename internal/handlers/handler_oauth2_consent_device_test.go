// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestOAuth2ConsentUserCodePOST(t *testing.T) {
	testCases := []struct {
		name  string
		level int
		body  func(userCode string) string
		setup func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string)
		msg   string
		err   any
	}{
		{
			name: "ShouldHandleMalformedBody",
			body: func(_ string) string { return "not json" },
			msg:  "Error occurred unmarshaling the user code request body during the Consent Flow stage of the Device Authorization Flow",
			err:  regexpAnyError,
		},
		{
			name: "ShouldHandleEmptyUserCode",
			body: func(_ string) string { return `{"user_code":""}` },
			msg:  "Error occurred exchanging the user code during the Consent Flow stage of the Device Authorization Flow as it is empty or too long",
		},
		{
			name: "ShouldHandleLongUserCode",
			body: func(_ string) string { return `{"user_code":"ABCDEFGHABCDEFGHABCDEFGHABCDEFGHA"}` },
			msg:  "Error occurred exchanging the user code during the Consent Flow stage of the Device Authorization Flow as it is empty or too long",
		},
		{
			name:  "ShouldHandleAnonymousUser",
			level: -1,
			msg:   "Error occurred fetching user session during the Consent Flow stage of the Device Authorization Flow as the user is anonymous",
		},
		{
			name: "ShouldHandleUnknownUserCode",
			body: func(_ string) string { return `{"user_code":"ABCDEFGH"}` },
			msg:  "Error occurred using the signature of the user code to retrieve the device code session during the Consent Flow stage of the Device Authorization Flow",
			err:  "sql: no rows in result set",
		},
		{
			name: "ShouldHandleRevokedDeviceCodeSession",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string) {
				mutateTestOIDCDeviceCodeSession(t, mock, userCode, func(device *model.OAuth2DeviceCodeSession) {
					device.Revoked = true
				})
			},
			msg: "Error occurred exchanging the user code for a consent session during the Consent Flow stage of the Device Authorization Flow",
			err: "the device code session is inactive or revoked",
		},
		{
			name: "ShouldHandleCorruptDeviceCodeSession",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string) {
				mutateTestOIDCDeviceCodeSession(t, mock, userCode, func(device *model.OAuth2DeviceCodeSession) {
					device.Session = []byte("not json")
				})
			},
			msg: "Error occurred exchanging the user code for a consent session during the Consent Flow stage of the Device Authorization Flow",
			err: "error occurred restoring the device code session data: invalid character 'o' in literal null (expecting 'u')",
		},
		{
			name: "ShouldHandleSubjectMismatch",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string) {
				mutateTestOIDCDeviceCodeSession(t, mock, userCode, func(device *model.OAuth2DeviceCodeSession) {
					device.Subject.String, device.Subject.Valid = uuid.Must(uuid.NewRandom()).String(), true
				})
			},
			msg: "Error occurred authorizing the user to enter the user code during the Consent Flow stage of the Device Authorization Flow",
			err: "the subject of the device code session does not match the subject of the user",
		},
		{
			name: "ShouldHandleUserCodeEnteredByAnotherUserAgent",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string) {
				mustClaimTestOIDCDeviceCodeSession(t, mock, userCode, newTestOIDCConsentSession(t, mock, uuid.Nil))
			},
			msg: "Error occurred exchanging the user code for a consent session during the Consent Flow stage of the Device Authorization Flow",
			err: "the user code has already been entered by another user agent: sql: no rows in result set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			level := tc.level
			if level == 0 {
				level = 1
			}

			mock := newTestOIDCDeviceCodeConsentMock(t, level, nil)
			defer mock.Close()

			userCode := mustGetTestOIDCUserCode(t, mock)

			if tc.setup != nil {
				tc.setup(t, mock, userCode)
			}

			body := fmt.Sprintf(`{"user_code":"%s"}`, userCode)

			if tc.body != nil {
				body = tc.body(userCode)
			}

			mock.Ctx.Response.Reset()
			mock.Ctx.Request.SetBodyString(body)

			OAuth2ConsentUserCodePOST(mock.Ctx)

			mock.Assert200KO(t, messageOperationFailed)

			AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), tc.msg, tc.err)

			assert.Nil(t, mock.Ctx.UserValue(middlewares.UserValueRateLimitExempt))
		})
	}
}

func TestOAuth2ConsentUserCodePOSTShouldHandleInsufficientAuthenticationLevel(t *testing.T) {
	client := newTestOIDCDeviceCodeClient(t)
	client.AuthorizationPolicy = "two_factor"

	mock := newTestOIDCDeviceCodeConsentMock(t, 1, &client)
	defer mock.Close()

	userCode := mustGetTestOIDCUserCode(t, mock)

	mock.Ctx.Response.Reset()
	mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"user_code":"%s"}`, userCode))

	OAuth2ConsentUserCodePOST(mock.Ctx)

	mock.Assert200KO(t, messageOperationFailed)

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred authorizing the user to enter the user code during the Consent Flow stage of the Device Authorization Flow", "the user is not sufficiently authenticated given the client authorization policy 'two_factor' with the authentication level 'one_factor'")
}

func TestOAuth2ConsentUserCodePOSTShouldExchangeUserCode(t *testing.T) {
	mock := newTestOIDCDeviceCodeConsentMock(t, 1, nil)
	defer mock.Close()

	userCode := mustGetTestOIDCUserCode(t, mock)

	flowID := mustExchangeTestOIDCUserCode(t, mock, userCode)

	assert.Equal(t, true, mock.Ctx.UserValue(middlewares.UserValueRateLimitExempt))

	consent, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(mock.Ctx, flowID)
	require.NoError(t, err)

	assert.Equal(t, testOIDCDeviceCodeID, consent.ClientID)
	assert.True(t, consent.Subject.Valid)
	assert.False(t, consent.Responded())
	assert.Equal(t, mocks.UserAgentBindingDigest, consent.Binding)

	device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

	expires, err := getOAuth2DeviceCodeSessionUserCodeExpiresAt(device)
	require.NoError(t, err)

	assert.True(t, expires.Equal(consent.ExpiresAt))
	assert.Equal(t, model.OAuth2DeviceCodeStatusClaimed, device.Status)
	assert.Equal(t, uuid.NullUUID{UUID: flowID, Valid: true}, device.ChallengeID)
	assert.True(t, device.Active)
}

func TestOAuth2ConsentUserCodePOSTShouldResumeExchangeForTheSameUserAgent(t *testing.T) {
	testCases := []struct {
		name    string
		binding bool
		setup   func(t *testing.T, mock *mocks.MockAutheliaCtx, flowID uuid.UUID)
		err     string
	}{
		{
			name:    "ShouldResume",
			binding: true,
		},
		{
			name: "ShouldNotResumeForAnotherUserAgent",
			err:  "the user code has already been entered by another user agent: the consent session is not bound to the user agent",
		},
		{
			name:    "ShouldNotResumeRespondedConsentSession",
			binding: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, flowID uuid.UUID) {
				consent, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(mock.Ctx, flowID)
				require.NoError(t, err)

				consent.SetRespondedAt(mock.Ctx.GetClock().Now(), 0)

				require.NoError(t, mock.Ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(mock.Ctx, consent, false))
			},
			err: "the consent session has already been responded to",
		},
		{
			name:    "ShouldNotResumeConsentSessionOfAnotherSubject",
			binding: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, flowID uuid.UUID) {
				consent, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(mock.Ctx, flowID)
				require.NoError(t, err)

				consent.Subject = uuid.NullUUID{UUID: uuid.Must(uuid.NewRandom()), Valid: true}

				require.NoError(t, mock.Ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(mock.Ctx, consent, false))
			},
			err: "the subject of the consent session does not match the subject of the user",
		},
		{
			name:    "ShouldNotResumeConsentSessionWithoutSubject",
			binding: true,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, flowID uuid.UUID) {
				consent, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(mock.Ctx, flowID)
				require.NoError(t, err)

				consent.Subject = uuid.NullUUID{}

				require.NoError(t, mock.Ctx.Providers.StorageProvider.SaveOAuth2ConsentSessionResponse(mock.Ctx, consent, false))
			},
			err: "the subject of the consent session does not match the subject of the user",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := newTestOIDCDeviceCodeConsentMock(t, 1, nil)
			defer mock.Close()

			userCode := mustGetTestOIDCUserCode(t, mock)

			flowID := mustExchangeTestOIDCUserCode(t, mock, userCode)

			if !tc.binding {
				mock.Ctx.Request.Header.SetCookie(session.UserAgentBindingCookieName("authelia_session"), testOAuth2UserAgentBindingAnotherUserAgent)
			}

			if tc.setup != nil {
				tc.setup(t, mock, flowID)
			}

			mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"user_code":"%s"}`, userCode))

			OAuth2ConsentUserCodePOST(mock.Ctx)

			if tc.err != "" {
				mock.Assert200KO(t, messageOperationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred exchanging the user code for a consent session during the Consent Flow stage of the Device Authorization Flow", tc.err)

				return
			}

			body := oidc.ConsentPostResponseBody{}

			require.Contains(t, string(mock.Ctx.Response.Body()), `"status":"OK"`)
			mock.GetResponseData(t, &body)

			assert.Equal(t, flowID.String(), body.FlowID)
			assert.Equal(t, mocks.UserAgentBinding, assertTestOAuth2UserAgentBindingCookie(t, mock))
		})
	}
}

func TestOAuth2ConsentUserCodePOSTShouldRollBackWhenTheClaimFails(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(mock *mocks.MockAutheliaCtx, tx context.Context)
		err   string
	}{
		{
			name: "ShouldHandleSaveConsentSessionError",
			setup: func(mock *mocks.MockAutheliaCtx, tx context.Context) {
				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(tx), gomock.Any()).
					Return(fmt.Errorf("error in db"))
			},
			err: "error occurred saving the consent session: error in db",
		},
		{
			name: "ShouldHandleClaimError",
			setup: func(mock *mocks.MockAutheliaCtx, tx context.Context) {
				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSession(gomock.Eq(tx), gomock.Any()).
					Return(nil)

				mock.StorageMock.EXPECT().
					ClaimOAuth2DeviceCodeSession(gomock.Eq(tx), gomock.Any()).
					Return(storage.ErrNoRowsAffected)
			},
			err: "error occurred claiming the device code session as the user code has already been entered or can no longer be used: no rows affected",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			defer mock.Close()

			tx := expectTestOIDCDeviceCodeConsentTX(mock, false)

			tc.setup(mock, tx)

			setupTestOIDCUserDetails(t, mock)

			config := newTestOIDCConfig(t)
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

			setupTestOIDCProvider(t, mock, config)
			setupTestOIDCDeviceCodeStore(t, mock)
			setupTestOIDCSubjectStore(t, mock)

			userCode := mustGetTestOIDCUserCode(t, mock)

			mock.Ctx.Response.Reset()
			mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"user_code":"%s"}`, userCode))

			OAuth2ConsentUserCodePOST(mock.Ctx)

			mock.Assert200KO(t, messageOperationFailed)

			AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred exchanging the user code for a consent session during the Consent Flow stage of the Device Authorization Flow", tc.err)
			assert.Nil(t, mock.Ctx.Response.Header.PeekCookie(session.UserAgentBindingCookieName("authelia_session")))

			device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

			assert.False(t, device.ChallengeID.Valid)
			assert.Equal(t, model.OAuth2DeviceCodeStatusNew, device.Status)
		})
	}
}

func TestOAuth2ConsentUserCodePOSTShouldCompleteEveryFlowOfTheUserAgent(t *testing.T) {
	mock := newTestOIDCDeviceCodeConsentMock(t, 1, nil)
	defer mock.Close()

	first := mustGetTestOIDCUserCode(t, mock)
	firstFlowID := mustExchangeTestOIDCUserCode(t, mock, first)

	second := mustGetTestOIDCUserCode(t, mock)
	secondFlowID := mustExchangeTestOIDCUserCode(t, mock, second)

	require.NotEqual(t, firstFlowID, secondFlowID)

	for _, flowID := range []uuid.UUID{firstFlowID, secondFlowID} {
		expectTestOIDCDeviceCodeConsentTX(mock, true)

		mock.Ctx.Response.Reset()
		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","flow_id":"%s","client_id":"%s","consent":true}`, flowID, testOIDCDeviceCodeID))

		OAuth2ConsentPOST(mock.Ctx)

		require.Contains(t, string(mock.Ctx.Response.Body()), `"status":"OK"`)
	}

	for _, userCode := range []string{first, second} {
		device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

		assert.Equal(t, model.OAuth2DeviceCodeStatusClaimed, device.Status)
		assert.True(t, device.Active)
	}

	for _, flowID := range []uuid.UUID{firstFlowID, secondFlowID} {
		consent, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(mock.Ctx, flowID)
		require.NoError(t, err)

		assert.Equal(t, mocks.UserAgentBindingDigest, consent.Binding)
		assert.True(t, consent.Responded())
		assert.True(t, consent.Authorized)
	}
}

func TestOAuth2ConsentGETDeviceAuthorization(t *testing.T) {
	testCases := []struct {
		name        string
		subflow     string
		exchanged   bool
		preConfig   bool
		expectError string
	}{
		{
			name:      "ShouldDisablePreConfigurationForTheDeviceAuthorizationSubFlow",
			subflow:   flowOpenIDConnectSubFlowNameDeviceAuthorization,
			exchanged: true,
		},
		{
			name:      "ShouldNotCheckTheDeviceWithoutTheSubFlow",
			exchanged: true,
			preConfig: true,
		},
		{
			name:        "ShouldRejectConsentSessionWithoutDeviceCodeSession",
			subflow:     flowOpenIDConnectSubFlowNameDeviceAuthorization,
			expectError: "error occurred loading the device code session: sql: no rows in result set",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			client := newTestOIDCDeviceCodeClient(t)
			duration := time.Hour

			client.ConsentMode = oidc.ClientConsentModePreConfigured.String()
			client.ConsentPreConfiguredDuration = &duration

			mock := newTestOIDCDeviceCodeConsentMock(t, 1, &client)
			defer mock.Close()

			var flowID uuid.UUID

			if tc.exchanged {
				flowID = mustExchangeTestOIDCUserCode(t, mock, mustGetTestOIDCUserCode(t, mock))
			} else {
				consent := newTestOIDCConsentSession(t, mock, mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID))
				consent.ClientID = testOIDCDeviceCodeID

				require.NoError(t, mock.Ctx.Providers.StorageProvider.SaveOAuth2ConsentSession(mock.Ctx, consent))

				flowID = consent.ChallengeID
			}

			uri := fmt.Sprintf("/api/oidc/consent?flow_id=%s", flowID)

			if tc.subflow != "" {
				uri += "&subflow=" + tc.subflow
			}

			mock.Ctx.Request.SetRequestURI(uri)

			OAuth2ConsentGET(mock.Ctx)

			if tc.expectError != "" {
				mock.Assert200KO(t, messageOperationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed providing consent flow information as the consent session has not claimed a device code session which is pending a decision", tc.expectError)

				return
			}

			body := oidc.ConsentGetResponseBody{}

			require.Contains(t, string(mock.Ctx.Response.Body()), `"status":"OK"`)
			mock.GetResponseData(t, &body)

			assert.Equal(t, testOIDCDeviceCodeID, body.ClientID)
			assert.Equal(t, tc.preConfig, body.PreConfiguration)
		})
	}
}

func TestOAuth2ConsentPOSTDeviceAuthorization(t *testing.T) {
	testCases := []struct {
		name  string
		body  func(flowID uuid.UUID) string
		setup func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, flowID uuid.UUID)
		msg   string
		err   any
	}{
		{
			name: "ShouldHandleMissingFlowID",
			body: func(_ uuid.UUID) string {
				return fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":true}`, testOIDCDeviceCodeID)
			},
			msg: "Request is missing the required field 'flow_id' from the JSON body during the Consent Flow stage of the Device Authorization Flow",
		},
		{
			name: "ShouldHandleMalformedFlowID",
			body: func(_ uuid.UUID) string {
				return fmt.Sprintf(`{"subflow":"device_authorization","flow_id":"abc","client_id":"%s","consent":true}`, testOIDCDeviceCodeID)
			},
			msg: "Error occurred parsing flow ID as a UUID during the Consent Flow stage of the Device Authorization Flow",
			err: "invalid UUID length: 3",
		},
		{
			name: "ShouldHandleClientIDMismatch",
			body: func(flowID uuid.UUID) string {
				return fmt.Sprintf(`{"subflow":"device_authorization","flow_id":"%s","client_id":"a-different-client","consent":true}`, flowID)
			},
			msg: "Error occurred during the Consent Flow stage of the Device Authorization Flow as the client id of the form and the client id of the consent session do not match",
		},
		{
			name: "ShouldHandleDeviceCodeSessionNoLongerPendingADecision",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, flowID uuid.UUID) {
				device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

				require.NoError(t, mock.Ctx.Providers.StorageProvider.DenyOAuth2DeviceCodeSession(mock.Ctx, device.Signature, flowID))
			},
			msg: "Error occurred during the Consent Flow stage of the Device Authorization Flow as the consent session has not claimed a device code session which is pending a decision",
			err: "the device code session is inactive or revoked",
		},
		{
			name: "ShouldHandleConsentSessionBoundToAnotherUserAgent",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, _ string, _ uuid.UUID) {
				mock.Ctx.Request.Header.SetCookie(session.UserAgentBindingCookieName("authelia_session"), testOAuth2UserAgentBindingAnotherUserAgent)
			},
			msg: "Error occurred fetching consent session during the Consent Flow stage of the Authorization Flow",
			err: "the consent session is not bound to the user agent",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := newTestOIDCDeviceCodeConsentMock(t, 1, nil)
			defer mock.Close()

			userCode := mustGetTestOIDCUserCode(t, mock)

			flowID := mustExchangeTestOIDCUserCode(t, mock, userCode)

			if tc.setup != nil {
				tc.setup(t, mock, userCode, flowID)
			}

			body := fmt.Sprintf(`{"subflow":"device_authorization","flow_id":"%s","client_id":"%s","consent":true}`, flowID, testOIDCDeviceCodeID)

			if tc.body != nil {
				body = tc.body(flowID)
			}

			mock.Ctx.Request.SetBodyString(body)

			OAuth2ConsentPOST(mock.Ctx)

			mock.Assert200KO(t, messageOperationFailed)

			AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), tc.msg, tc.err)
		})
	}
}

func TestOAuth2ConsentPOSTShouldRejectDeviceAuthorizationConsentSessionWithoutSubflow(t *testing.T) {
	testCases := []struct {
		name    string
		subflow string
	}{
		{"ShouldRejectWithoutSubflow", ""},
		{"ShouldRejectWithUnknownSubflow", `"subflow":"not-a-subflow",`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := newTestOIDCDeviceCodeConsentMock(t, 1, nil)
			defer mock.Close()

			userCode := mustGetTestOIDCUserCode(t, mock)

			flowID := mustExchangeTestOIDCUserCode(t, mock, userCode)

			mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{%s"flow_id":"%s","client_id":"%s","consent":true}`, tc.subflow, flowID, testOIDCDeviceCodeID))

			OAuth2ConsentPOST(mock.Ctx)

			mock.Assert200KO(t, messageOperationFailed)

			AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred performing consent during the Consent Flow stage of the Authorization Flow as the consent session belongs to the Device Authorization Flow", nil)

			consent, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(mock.Ctx, flowID)
			require.NoError(t, err)

			assert.False(t, consent.Responded())
			assert.False(t, consent.Authorized)

			device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

			assert.Equal(t, model.OAuth2DeviceCodeStatusClaimed, device.Status)
			assert.True(t, device.Active)
		})
	}
}

func TestOAuth2ConsentPOSTDeviceAuthorizationDecision(t *testing.T) {
	testCases := []struct {
		name    string
		consent bool
		status  int
		active  bool
	}{
		{"ShouldGrantConsent", true, model.OAuth2DeviceCodeStatusClaimed, true},
		{"ShouldRejectConsent", false, model.OAuth2DeviceCodeStatusDenied, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := newTestOIDCDeviceCodeConsentMock(t, 1, nil)
			defer mock.Close()

			userCode := mustGetTestOIDCUserCode(t, mock)

			flowID := mustExchangeTestOIDCUserCode(t, mock, userCode)

			expectTestOIDCDeviceCodeConsentTX(mock, true)

			mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","flow_id":"%s","client_id":"%s","consent":%t}`, flowID, testOIDCDeviceCodeID, tc.consent))

			OAuth2ConsentPOST(mock.Ctx)

			body := oidc.ConsentPostResponseBody{}

			require.Contains(t, string(mock.Ctx.Response.Body()), `"status":"OK"`)
			mock.GetResponseData(t, &body)

			assert.Equal(t, flowID.String(), body.FlowID)
			assert.Empty(t, body.RedirectURI)

			consent, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(mock.Ctx, flowID)
			require.NoError(t, err)

			assert.True(t, consent.Responded())
			assert.Equal(t, tc.consent, consent.Authorized)

			device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

			assert.Equal(t, tc.status, device.Status)
			assert.Equal(t, tc.active, device.Active)
			assert.Equal(t, uuid.NullUUID{UUID: flowID, Valid: true}, device.ChallengeID)
		})
	}
}

func TestOAuth2ConsentPOSTDeviceAuthorizationShouldRollBack(t *testing.T) {
	testCases := []struct {
		name    string
		consent bool
		setup   func(mock *mocks.MockAutheliaCtx)
		msg     string
	}{
		{
			name:    "ShouldHandleSaveConsentSessionResponseError",
			consent: true,
			setup: func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), true).
					Return(fmt.Errorf("error in db"))
			},
			msg: "Error occurred saving the consent session response to the database during the Consent Flow stage of the Device Authorization Flow",
		},
		{
			name: "ShouldHandleDenyError",
			setup: func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					DenyOAuth2DeviceCodeSession(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("error in db"))
			},
			msg: "Error occurred denying the device code session during the Consent Flow stage of the Device Authorization Flow as it is no longer pending a decision",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			defer mock.Close()

			expectTestOIDCDeviceCodeConsentTX(mock, true)
			expectTestOIDCDeviceCodeConsentTX(mock, false)

			tc.setup(mock)

			setupTestOIDCUserDetails(t, mock)

			config := newTestOIDCConfig(t)
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

			setupTestOIDCProvider(t, mock, config)
			setupTestOIDCDeviceCodeStore(t, mock)
			setupTestOIDCConsentStore(t, mock)
			setupTestOIDCSubjectStore(t, mock)

			userCode := mustGetTestOIDCUserCode(t, mock)

			flowID := mustExchangeTestOIDCUserCodeWithoutTX(t, mock, userCode)

			mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","flow_id":"%s","client_id":"%s","consent":%t}`, flowID, testOIDCDeviceCodeID, tc.consent))

			OAuth2ConsentPOST(mock.Ctx)

			mock.Assert200KO(t, messageOperationFailed)

			AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), tc.msg, "error in db")
		})
	}
}

func newTestOIDCDeviceCodeConsentMock(t *testing.T, level int, client *schema.IdentityProvidersOpenIDConnectClient) (mock *mocks.MockAutheliaCtx) {
	t.Helper()

	if level < 0 {
		mock = mocks.NewMockAutheliaCtx(t)
	} else {
		mock = mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(level))
	}

	setupTestOIDCUserDetails(t, mock)

	if client == nil {
		value := newTestOIDCDeviceCodeClient(t)

		client = &value
	}

	config := newTestOIDCConfig(t)
	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{*client}

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCDeviceCodeStore(t, mock)
	setupTestOIDCConsentStore(t, mock)
	setupTestOIDCSubjectStore(t, mock)

	return mock
}

func mustExchangeTestOIDCUserCode(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string) (flowID uuid.UUID) {
	t.Helper()

	expectTestOIDCDeviceCodeConsentTX(mock, true)

	return mustExchangeTestOIDCUserCodeWithoutTX(t, mock, userCode)
}

func mustExchangeTestOIDCUserCodeWithoutTX(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string) (flowID uuid.UUID) {
	t.Helper()

	mock.Ctx.Response.Reset()
	mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"user_code":"%s"}`, userCode))

	OAuth2ConsentUserCodePOST(mock.Ctx)

	body := oidc.ConsentPostResponseBody{}

	require.Contains(t, string(mock.Ctx.Response.Body()), `"status":"OK"`)
	mock.GetResponseData(t, &body)

	flowID = uuid.MustParse(body.FlowID)

	mock.Ctx.Request.Header.SetCookie(session.UserAgentBindingCookieName("authelia_session"), assertTestOAuth2UserAgentBindingCookie(t, mock))
	mock.Ctx.Response.Reset()

	return flowID
}

func mustLoadTestOIDCDeviceCodeSession(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string) (device *model.OAuth2DeviceCodeSession) {
	t.Helper()

	signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(mock.Ctx, userCode)
	require.NoError(t, err)

	device, err = mock.Ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(mock.Ctx, signature)
	require.NoError(t, err)

	return device
}

func mutateTestOIDCDeviceCodeSession(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, mutate func(device *model.OAuth2DeviceCodeSession)) {
	t.Helper()

	device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

	mutate(device)

	require.NoError(t, mock.Ctx.Providers.StorageProvider.UpdateOAuth2DeviceCodeSession(mock.Ctx, device))
}

type testOIDCDeviceCodeConsentTXKey struct{}

func expectTestOIDCDeviceCodeConsentTX(mock *mocks.MockAutheliaCtx, committed bool) (tx context.Context) {
	tx = context.WithValue(mock.Ctx, testOIDCDeviceCodeConsentTXKey{}, true)

	mock.StorageMock.EXPECT().
		BeginTX(gomock.Any()).
		Return(tx, nil)

	if committed {
		mock.StorageMock.EXPECT().
			Commit(gomock.Eq(tx)).
			Return(nil)
	} else {
		mock.StorageMock.EXPECT().
			Rollback(gomock.Eq(tx)).
			Return(nil)
	}

	return tx
}
