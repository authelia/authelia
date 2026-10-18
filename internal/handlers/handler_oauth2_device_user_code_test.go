// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
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

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestOAuth2DeviceAuthorizationUserCodeRateLimit(t *testing.T) {
	setup := func(t *testing.T) (*mocks.MockAutheliaCtx, middlewares.AutheliaMiddleware) {
		t.Helper()

		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		limit := middlewares.NewRateLimiter(
			middlewares.WithRateLimitBuckets(middlewares.RateLimitBucketConfig{Period: time.Minute, Requests: 2}),
		).Middleware()

		return mock, limit
	}

	t.Run("ShouldLimitUserCodesWhichDoNotIdentifyAPendingSession", func(t *testing.T) {
		mock, limit := setup(t)
		defer mock.Close()

		getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJK")
		mock.Assert200KO(t, messageOperationFailed)

		getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJL")
		mock.Assert200KO(t, messageOperationFailed)

		getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJM")
		assert.Equal(t, fasthttp.StatusTooManyRequests, mock.Ctx.Response.StatusCode())
		assert.NotEmpty(t, mock.Ctx.Response.Header.Peek(fasthttp.HeaderRetryAfter))

		userCode := mustGetTestOIDCUserCode(t, mock)

		getTestOAuth2ConsentUserCode(mock, limit, userCode)
		assert.Equal(t, fasthttp.StatusTooManyRequests, mock.Ctx.Response.StatusCode())
	})

	t.Run("ShouldNotLimitUserCodesWhichIdentifyAPendingSession", func(t *testing.T) {
		mock, limit := setup(t)
		defer mock.Close()

		userCode := mustGetTestOIDCUserCode(t, mock)

		for range 5 {
			getTestOAuth2ConsentUserCode(mock, limit, userCode)

			body := map[string]any{}

			mock.GetResponseData(t, &body)

			assert.Equal(t, testOIDCDeviceCodeID, body["client_id"])
		}
	})

	t.Run("ShouldLimitClaimedUserCodes", func(t *testing.T) {
		mock, limit := setup(t)
		defer mock.Close()

		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)

		userCode := mustGetTestOIDCUserCode(t, mock)

		newTestOIDCDeviceConsentSession(t, mock, userCode, mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID))

		for range 2 {
			getTestOAuth2ConsentUserCode(mock, limit, userCode)
			mock.Assert200KO(t, messageOperationFailed)
		}

		getTestOAuth2ConsentUserCode(mock, limit, userCode)
		assert.Equal(t, fasthttp.StatusTooManyRequests, mock.Ctx.Response.StatusCode())
	})

	t.Run("ShouldNotLimitRequestsWithoutAUserCode", func(t *testing.T) {
		mock, limit := setup(t)
		defer mock.Close()

		setupTestOIDCConsentStore(t, mock)

		for range 5 {
			mock.Ctx.Request.SetRequestURI("/api/oidc/consent?flow_id=" + uuid.Must(uuid.NewRandom()).String())

			doTestOAuth2UserCodeRequest(mock, limit, OAuth2ConsentGET)
			mock.Assert200KO(t, messageOperationFailed)
		}

		getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJK")
		mock.Assert200KO(t, messageOperationFailed)
	})

	t.Run("ShouldLimitEachRemoteIP", func(t *testing.T) {
		mock, limit := setup(t)
		defer mock.Close()

		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "192.0.2.1")

		for range 2 {
			getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJK")
			mock.Assert200KO(t, messageOperationFailed)
		}

		getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJK")
		assert.Equal(t, fasthttp.StatusTooManyRequests, mock.Ctx.Response.StatusCode())

		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "192.0.2.2")

		getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJK")
		mock.Assert200KO(t, messageOperationFailed)
	})

	t.Run("ShouldShareTheLimitBetweenEndpoints", func(t *testing.T) {
		mock, limit := setup(t)
		defer mock.Close()

		getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJK")
		mock.Assert200KO(t, messageOperationFailed)

		postTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJL", true)
		mock.Assert200KO(t, messageOperationFailed)

		assert.Equal(t, http.StatusTooManyRequests, putTestOAuth2DeviceAuthorizationUserCode(t, mock, limit, "BCDFGHJN", uuid.Must(uuid.NewRandom()).String()))
		assert.NotEmpty(t, mock.Ctx.Response.Header.Peek(fasthttp.HeaderRetryAfter))
	})

	t.Run("ShouldNotCountTheFlowOfAValidUserCode", func(t *testing.T) {
		mock, limit := setup(t)
		defer mock.Close()

		setupTestOIDCConsentStore(t, mock)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJK")
		mock.Assert200KO(t, messageOperationFailed)

		userCode := mustGetTestOIDCUserCode(t, mock)

		getTestOAuth2ConsentUserCode(mock, limit, userCode)
		require.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

		flowID := postTestOAuth2ConsentUserCode(mock, limit, userCode, true)
		require.NotEmpty(t, flowID)

		code := putTestOAuth2DeviceAuthorizationUserCode(t, mock, limit, userCode, flowID)
		require.Equal(t, http.StatusOK, code)

		getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJL")
		mock.Assert200KO(t, messageOperationFailed)

		getTestOAuth2ConsentUserCode(mock, limit, "BCDFGHJM")
		assert.Equal(t, fasthttp.StatusTooManyRequests, mock.Ctx.Response.StatusCode())
	})

	t.Run("ShouldLimitUnclaimedUserCodesAtTheDeviceAuthorizationEndpoint", func(t *testing.T) {
		mock, limit := setup(t)
		defer mock.Close()

		setupTestOIDCConsentStore(t, mock)

		for range 2 {
			code := putTestOAuth2DeviceAuthorizationUserCode(t, mock, limit, mustGetTestOIDCUserCode(t, mock), uuid.Must(uuid.NewRandom()).String())
			assert.NotEqual(t, http.StatusOK, code)
			assert.NotEqual(t, http.StatusTooManyRequests, code)
		}

		code := putTestOAuth2DeviceAuthorizationUserCode(t, mock, limit, mustGetTestOIDCUserCode(t, mock), uuid.Must(uuid.NewRandom()).String())
		assert.Equal(t, http.StatusTooManyRequests, code)
	})
}

func TestOAuth2DeviceAuthorizationClaim(t *testing.T) {
	var limit middlewares.AutheliaMiddleware

	setup := func(t *testing.T) (mock *mocks.MockAutheliaCtx, userCode string) {
		t.Helper()

		mock = mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)
		setupTestOIDCSessionStore(t, mock)
		setupTestOIDCSubjectStore(t, mock)
		setupTestOIDCUserDetails(t, mock)

		return mock, mustGetTestOIDCUserCode(t, mock)
	}

	t.Run("ShouldClaimThenApprove", func(t *testing.T) {
		mock, userCode := setup(t)
		defer mock.Close()

		setupTestOIDCConsentStore(t, mock)

		flowID := postTestOAuth2ConsentUserCode(mock, limit, userCode, true)

		require.NotEmpty(t, flowID)

		device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

		assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusClaimed), device.Status)
		assert.Equal(t, flowID, device.ChallengeID.UUID.String())
		assert.Equal(t, mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID).String(), device.Subject.String)

		code := putTestOAuth2DeviceAuthorizationUserCode(t, mock, limit, userCode, flowID)
		require.Equal(t, http.StatusOK, code)

		assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusApproved), mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).Status)

		code = putTestOAuth2DeviceAuthorizationUserCode(t, mock, limit, userCode, flowID)
		assert.NotEqual(t, http.StatusOK, code)
	})

	t.Run("ShouldClaimAndDeny", func(t *testing.T) {
		mock, userCode := setup(t)
		defer mock.Close()

		setupTestOIDCConsentStore(t, mock)

		require.NotEmpty(t, postTestOAuth2ConsentUserCode(mock, limit, userCode, false))

		device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

		assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusDenied), device.Status)
		assert.False(t, device.Active)
		assert.True(t, device.ChallengeID.Valid)
	})

	t.Run("ShouldNotClaimTwice", func(t *testing.T) {
		mock, userCode := setup(t)
		defer mock.Close()

		setupTestOIDCConsentStore(t, mock)

		flowID := postTestOAuth2ConsentUserCode(mock, limit, userCode, true)
		require.NotEmpty(t, flowID)

		postTestOAuth2ConsentUserCode(mock, limit, userCode, true)
		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred during the Consent Flow stage of the Device Authorization Flow as the user code does not identify a pending device code session", "the device code session has been claimed by another request: the device code session is already bound to a consent session")

		assert.Equal(t, flowID, mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).ChallengeID.UUID.String())
	})

	t.Run("ShouldLoseTheClaimToAConcurrentRequest", func(t *testing.T) {
		mock, userCode := setup(t)
		defer mock.Close()

		other := &model.OAuth2ConsentSession{ChallengeID: uuid.Must(uuid.NewRandom()), Subject: uuid.NullUUID{UUID: uuid.Must(uuid.NewRandom()), Valid: true}}

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSession(gomock.Any(), gomock.Any()).
			DoAndReturn(func(_ context.Context, _ *model.OAuth2ConsentSession) error {
				claimTestOIDCDeviceCodeSession(t, mock, userCode, other)

				return nil
			})

		mock.StorageMock.EXPECT().
			SaveOAuth2ConsentSessionResponse(gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil)

		postTestOAuth2ConsentUserCode(mock, limit, userCode, true)
		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred claiming the device code session during the Consent Flow stage of the Device Authorization Flow", "the device code session has been claimed by another request: no rows affected")

		assert.Equal(t, other.ChallengeID, mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).ChallengeID.UUID)
	})

	t.Run("ShouldNotApproveUnclaimed", func(t *testing.T) {
		mock, userCode := setup(t)
		defer mock.Close()

		consent := newTestOIDCConsentSession(t, mock, mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID))
		consent.ClientID = testOIDCDeviceCodeID

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		code := putTestOAuth2DeviceAuthorizationUserCode(t, mock, limit, userCode, consent.ChallengeID.String())
		assert.NotEqual(t, http.StatusOK, code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to match the device code session to the consent session which claimed it during the User Authorization Flow", "the device code session has been claimed by another request: the consent session is not the one which claimed the device code session")

		assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusNew), mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).Status)
	})

	t.Run("ShouldNotApproveWithAnotherConsentSession", func(t *testing.T) {
		mock, userCode := setup(t)
		defer mock.Close()

		subject := mustGetTestOIDCSubject(t, mock, testOIDCDeviceCodeID)

		claimed := newTestOIDCConsentSession(t, mock, subject)
		claimed.ClientID = testOIDCDeviceCodeID

		claimTestOIDCDeviceCodeSession(t, mock, userCode, claimed)

		consent := newTestOIDCConsentSession(t, mock, subject)
		consent.ClientID = testOIDCDeviceCodeID

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		code := putTestOAuth2DeviceAuthorizationUserCode(t, mock, limit, userCode, consent.ChallengeID.String())
		assert.NotEqual(t, http.StatusOK, code)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Request failed to match the device code session to the consent session which claimed it during the User Authorization Flow", "the device code session has been claimed by another request: the consent session is not the one which claimed the device code session")

		assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusClaimed), mustLoadTestOIDCDeviceCodeSession(t, mock, userCode).Status)
	})

	t.Run("ShouldKeepTheClaimWhenPolled", func(t *testing.T) {
		mock, userCode := setup(t)
		defer mock.Close()

		setupTestOIDCConsentStore(t, mock)

		flowID := postTestOAuth2ConsentUserCode(mock, limit, userCode, true)

		device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

		device.ChallengeID, device.Subject, device.Status = uuid.NullUUID{}, sql.NullString{}, int(oauthelia2.DeviceAuthorizeStatusNew)

		request, err := device.ToRequest(mock.Ctx, oidc.NewSession(), mock.Ctx.Providers.OpenIDConnect.Store)
		require.NoError(t, err)

		request.SetLastChecked(time.Now())

		require.NoError(t, mock.Ctx.Providers.OpenIDConnect.UpdateDeviceCodeSession(mock.Ctx, device.Signature, request))

		device = mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)

		assert.Equal(t, int(oauthelia2.DeviceAuthorizeStatusClaimed), device.Status)
		assert.Equal(t, flowID, device.ChallengeID.UUID.String())
		assert.True(t, device.Subject.Valid)
	})
}

func doTestOAuth2UserCodeRequest(mock *mocks.MockAutheliaCtx, limit middlewares.AutheliaMiddleware, handler middlewares.RequestHandler) {
	mock.Ctx.Response.Reset()
	mock.Ctx.RemoveUserValue(middlewares.UserValueRateLimitExempt)

	if limit != nil {
		handler = limit(handler)
	}

	handler(mock.Ctx)
}

func getTestOAuth2ConsentUserCode(mock *mocks.MockAutheliaCtx, limit middlewares.AutheliaMiddleware, userCode string) {
	mock.Ctx.Request.SetRequestURI("/api/oidc/consent?user_code=" + userCode)

	doTestOAuth2UserCodeRequest(mock, limit, OAuth2ConsentGET)
}

func postTestOAuth2ConsentUserCode(mock *mocks.MockAutheliaCtx, limit middlewares.AutheliaMiddleware, userCode string, consent bool) (flowID string) {
	mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":%t,"user_code":"%s"}`, testOIDCDeviceCodeID, consent, userCode))

	doTestOAuth2UserCodeRequest(mock, limit, OAuth2ConsentPOST)

	body := struct {
		Data oidc.ConsentPostResponseBody `json:"data"`
	}{}

	_ = json.Unmarshal(mock.Ctx.Response.Body(), &body)

	return body.Data.FlowID
}

func putTestOAuth2DeviceAuthorizationUserCode(t *testing.T, mock *mocks.MockAutheliaCtx, limit middlewares.AutheliaMiddleware, userCode, flowID string) (code int) {
	t.Helper()

	rw, r := newTestOAuth2Request(t, fasthttp.MethodPut, testOIDCDeviceAuthorizationEndpoint, url.Values{
		oidc.FormParameterUserCode: []string{userCode},
		oidc.FormParameterFlowID:   []string{flowID},
	})

	handled := false

	doTestOAuth2UserCodeRequest(mock, limit, func(ctx *middlewares.AutheliaCtx) {
		handled = true

		OAuth2DeviceAuthorizationPUT(ctx, rw, r)
	})

	if !handled {
		return mock.Ctx.Response.StatusCode()
	}

	return rw.Code
}

func claimTestOIDCDeviceCodeSession(t *testing.T, mock *mocks.MockAutheliaCtx, userCode string, consent *model.OAuth2ConsentSession) {
	t.Helper()

	device := mustLoadTestOIDCDeviceCodeSession(t, mock, userCode)
	device.ChallengeID = uuid.NullUUID{UUID: consent.ChallengeID, Valid: true}
	device.Subject = sql.NullString{String: consent.Subject.UUID.String(), Valid: true}
	device.Status = int(oauthelia2.DeviceAuthorizeStatusClaimed)

	require.NoError(t, mock.Ctx.Providers.StorageProvider.ClaimOAuth2DeviceCodeSession(mock.Ctx, device))
}
