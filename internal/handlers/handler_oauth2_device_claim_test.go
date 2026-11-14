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

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
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
			err: "the user code has already been entered by another user agent: sql: no rows in result set",
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
			mock := newTestOIDCDeviceCodeConsentMock(t, 1, nil)
			defer mock.Close()

			userCode := mustGetTestOIDCUserCode(t, mock)

			if tc.setup != nil {
				tc.setup(t, mock, userCode)
			}

			if tc.err == "" {
				mustExchangeTestOIDCUserCode(t, mock, userCode)

				return
			}

			mock.Ctx.Response.Reset()
			mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"user_code":"%s"}`, userCode))

			OAuth2ConsentUserCodePOST(mock.Ctx)

			mock.Assert200KO(t, messageOperationFailed)

			AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred exchanging the user code for a consent session during the Consent Flow stage of the Device Authorization Flow", tc.err)
		})
	}
}

func TestOAuth2ConsentUserCodePOSTShouldOnlyAcceptUserCodeOnce(t *testing.T) {
	mock := newTestOIDCDeviceCodeConsentMock(t, 1, nil)
	defer mock.Close()

	userCode := mustGetTestOIDCUserCode(t, mock)

	flowID := mustExchangeTestOIDCUserCode(t, mock, userCode)

	device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

	assert.Equal(t, model.OAuth2DeviceCodeStatusClaimed, device.Status)
	assert.Equal(t, flowID, device.ChallengeID.UUID)

	mock.Ctx.Request.Header.SetCookie(session.AuthorizationBindingCookieName("authelia_session"), testOAuth2AuthorizationBindingAnotherUserAgent)
	mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"user_code":"%s"}`, userCode))

	OAuth2ConsentUserCodePOST(mock.Ctx)

	mock.Assert200KO(t, messageOperationFailed)

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred exchanging the user code for a consent session during the Consent Flow stage of the Device Authorization Flow", "the user code has already been entered by another user agent: the consent session is not bound to the user agent")
}

func TestOAuth2DeviceAuthorizationShouldCompleteTheUserCodeFlow(t *testing.T) {
	testCases := []struct {
		name    string
		consent bool
		status  int
	}{
		{"ShouldApprove", true, http.StatusOK},
		{"ShouldNotApproveAfterRejection", false, http.StatusBadRequest},
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

			require.Contains(t, string(mock.Ctx.Response.Body()), `"status":"OK"`)

			setupTestOIDCSessionStore(t, mock)

			mock.Ctx.Response.Reset()

			rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
				oidc.FormParameterUserCode: []string{userCode},
				oidc.FormParameterFlowID:   []string{flowID.String()},
			})

			OAuth2DeviceAuthorizationPUT(mock.Ctx, rw, r)

			assert.Equal(t, tc.status, rw.Code)

			device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

			if tc.consent {
				assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusApproved), device.Status)

				consent, err := mock.Ctx.Providers.StorageProvider.LoadOAuth2ConsentSessionByChallengeID(mock.Ctx, flowID)
				require.NoError(t, err)

				assert.True(t, consent.Granted)
			} else {
				assert.Equal(t, model.OAuth2DeviceCodeStatusDenied, device.Status)
			}
		})
	}
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
