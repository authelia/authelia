// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2ConsentDeviceAuthorizationUserCodeEntry(t *testing.T) {
	testCases := []struct {
		name  string
		setup func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string)
		err   string
	}{
		{
			name: "ShouldAllowNewUserCode",
		},
		{
			name: "ShouldRejectUserCodeEnteredBefore",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string) {
				mustClaimTestOIDCDeviceCodeSession(t, mock, userCode, newTestOIDCConsentSession(t, mock, uuid.Nil))
			},
			err: "the user code has already been entered",
		},
		{
			name: "ShouldRejectExpiredUserCode",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string) {
				future := &clock.Fixed{}
				future.Set(time.Now().Add(time.Hour * 24))

				mock.Ctx.Providers.Clock = future
			},
			err: "the user code has expired",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
			defer mock.Close()

			setupTestOIDCUserDetails(t, mock)

			config := newTestOIDCConfig(t)
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

			setupTestOIDCProvider(t, mock, config)
			setupTestOIDCDeviceCodeStore(t, mock)

			userCode := mustGetTestOIDCUserCode(t, mock)

			if tc.setup != nil {
				tc.setup(t, mock, userCode)
			}

			mock.Ctx.Response.Reset()
			mock.Ctx.Request.SetRequestURI("/api/oidc/consent?" + url.Values{oidc.FormParameterUserCode: []string{userCode}}.Encode())

			OAuth2ConsentGET(mock.Ctx)

			if tc.err == "" {
				assert.Contains(t, string(mock.Ctx.Response.Body()), `"status":"OK"`)
			} else {
				mock.Assert200KO(t, messageOperationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Flow failed to retrieve Consent Flow data as the user code can't be entered", tc.err)
			}
		})
	}
}

func TestOAuth2ConsentDeviceAuthorizationPOSTShouldOnlyAcceptUserCodeOnce(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
	defer mock.Close()

	setupTestOIDCUserDetails(t, mock)

	config := newTestOIDCConfig(t)
	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCDeviceCodeStore(t, mock)
	setupTestOIDCConsentStore(t, mock)
	setupTestOIDCSubjectStore(t, mock)

	userCode := mustGetTestOIDCUserCode(t, mock)

	body := fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":true,"user_code":"%s"}`, testOIDCDeviceCodeID, userCode)

	mock.Ctx.Response.Reset()
	mock.Ctx.Request.SetBodyString(body)

	OAuth2ConsentPOST(mock.Ctx)

	response := oidc.ConsentPostResponseBody{}

	mock.GetResponseData(t, &response)

	require.NotEmpty(t, response.FlowID)

	signature, err := mock.Ctx.Providers.OpenIDConnect.Strategy.Core.RFC8628UserCodeSignature(mock.Ctx, userCode)
	require.NoError(t, err)

	device, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2DeviceCodeSessionByUserCode(mock.Ctx, signature)
	require.NoError(t, err)

	assert.Equal(t, model.OAuth2DeviceCodeStatusClaimed, device.Status)
	assert.Equal(t, response.FlowID, device.ChallengeID.UUID.String())

	mock.Ctx.Response.Reset()
	mock.Ctx.Request.SetBodyString(body)

	OAuth2ConsentPOST(mock.Ctx)

	mock.Assert200KO(t, messageOperationFailed)

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred trying to advance the Consent Flow stage of the Device Authorization Flow as the user code can't be entered", "the user code has already been entered")
}

func TestOAuth2DeviceAuthorizationPUTShouldMatchConsentToDevice(t *testing.T) {
	testCases := []struct {
		name   string
		setup  func(t *testing.T, mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) (userCode string)
		err    string
		status int
	}{
		{
			name:   "ShouldApproveDeviceClaimedByTheConsentSession",
			status: http.StatusOK,
		},
		{
			name: "ShouldRejectConsentSessionForAnotherDevice",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) (userCode string) {
				userCode = mustGetTestOIDCUserCode(t, mock)

				mustClaimTestOIDCDeviceCodeSession(t, mock, userCode, newTestOIDCConsentSession(t, mock, consent.Subject.UUID))

				return userCode
			},
			err: "the device code session was claimed by another consent session",
		},
		{
			name: "ShouldRejectUnclaimedDevice",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) (userCode string) {
				return mustGetTestOIDCUserCode(t, mock)
			},
			err: "the device code session has the status 'New' rather than 'Claimed'",
		},
		{
			name: "ShouldRejectConsentSessionForAnotherClient",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) (userCode string) {
				consent.ClientID = testOIDCAuthorizationCodeID

				return ""
			},
			err: "the consent session belongs to client 'authorization-code' rather than client 'device-code'",
		},
		{
			name: "ShouldRejectUnauthorizedConsentSession",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) (userCode string) {
				consent.Authorized = false

				return ""
			},
			err: "the consent session has not been authorized",
		},
		{
			name: "ShouldRejectUsedConsentSession",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) (userCode string) {
				consent.Granted = true

				return ""
			},
			err: "the consent session has already been used or has expired",
		},
		{
			name: "ShouldRejectExpiredConsentSession",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, consent *model.OAuth2ConsentSession) (userCode string) {
				consent.ExpiresAt = mock.Ctx.GetClock().Now().Add(-time.Minute)

				return ""
			},
			err: "the consent session has already been used or has expired",
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

			consent := newTestOIDCConsentSession(t, mock, mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID))
			consent.ClientID = testOIDCDeviceCodeID
			consent.GrantedScopes = model.StringSlicePipeDelimited{oidc.ScopeOpenID}

			mustClaimTestOIDCDeviceCodeSession(t, mock, userCode, consent)

			if tc.setup != nil {
				if other := tc.setup(t, mock, consent); other != "" {
					userCode = other
				}
			}

			mock.StorageMock.EXPECT().
				LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
				AnyTimes().
				Return(consent, nil)

			if tc.err == "" {
				mock.StorageMock.EXPECT().
					SaveOAuth2ConsentSessionGranted(gomock.Any(), consent.ID).
					Return(nil)
			}

			rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
				oidc.FormParameterUserCode: []string{userCode},
				oidc.FormParameterFlowID:   []string{consent.ChallengeID.String()},
			})

			OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

			if tc.err == "" {
				assert.Equal(t, tc.status, rw.Code)
			} else {
				assert.NotEqual(t, http.StatusOK, rw.Code)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to match the consent session to the device code session during the User Authorization Flow", tc.err)
			}
		})
	}
}
