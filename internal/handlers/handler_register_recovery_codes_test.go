// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/metrics"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestRecoveryCodesGenerationPOST_SurfacesSMTPFailure(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	newRecoveryCodeUserCtx(t, mock)

	mock.StorageMock.EXPECT().
		ReplaceRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername), gomock.Len(model.RecoveryCodeBatchSize), gomock.Any()).
		Return(nil)

	mock.NotifierMock.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Eq("Recovery codes generated"), gomock.Any(), gomock.Any()).
		Return(errors.New("smtp temporarily unavailable"))

	RecoveryCodesGenerationPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	var resp struct {
		Status string `json:"status"`
		Data   struct {
			Codes            []string `json:"codes"`
			NotificationSent bool     `json:"notification_sent"`
		} `json:"data"`
	}

	require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), &resp))

	assert.Equal(t, "OK", resp.Status)
	assert.Len(t, resp.Data.Codes, model.RecoveryCodeBatchSize)
	assert.False(t, resp.Data.NotificationSent, "notification_sent must be false when SMTP delivery fails")
}

func TestRecoveryCodesGenerationPOST_RejectsAnonymous(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	RecoveryCodesGenerationPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
}

func TestRecoveryCodesGenerationPOST_ReplaceFailureReturnsNoCodes(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	newRecoveryCodeUserCtx(t, mock)

	mock.StorageMock.EXPECT().
		ReplaceRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername), gomock.Len(model.RecoveryCodeBatchSize), gomock.Any()).
		Return(errors.New("insert failed"))

	RecoveryCodesGenerationPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusInternalServerError, mock.Ctx.Response.StatusCode())
	assert.NotContains(t, string(mock.Ctx.Response.Body()), `"codes"`)
}

func TestRecoveryCodesGenerationPOST(t *testing.T) {
	testCases := []struct {
		name     string
		setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleSessionBackendError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				repository := setupTestFailingSessionRepository(t, mock)
				repository.errGet = errTestSessionBackend

				mock.Ctx.Request.Header.SetCookie("authelia_session", newTestSessionCookie("an-identifier-which-cannot-be-retrieved"))
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating recovery codes: error occurred retrieving the user session data", "error occurred getting session from backend: backend unavailable")
			},
		},
		{
			"ShouldHandleAnonymousUser",
			nil,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating recovery codes", "user is anonymous")
			},
		},
		{
			"ShouldHandleRandomError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				newRecoveryCodeUserCtx(t, mock)

				mock.Ctx.Providers.Random = mock.RandomMock

				mock.RandomMock.EXPECT().
					BytesCustomErr(model.RecoveryCodeLength, gomock.Any()).
					Return(nil, errors.New("entropy exhausted"))
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert500KO(t, messageOperationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred generating recovery codes for user 'john': error generating code 0", "failed to generate recovery code random bytes: entropy exhausted")
			},
		},
		{
			"ShouldGenerateCodesAndRecordMetrics",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				newRecoveryCodeUserCtx(t, mock)

				prom, err := metrics.NewPrometheus()
				require.NoError(t, err)

				mock.Ctx.Providers.Metrics = prom

				mock.StorageMock.EXPECT().
					ReplaceRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername), gomock.Len(model.RecoveryCodeBatchSize), gomock.Any()).
					Return(nil)

				mock.NotifierMock.EXPECT().
					Send(gomock.Any(), gomock.Any(), gomock.Eq("Recovery codes generated"), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

				var resp struct {
					Codes            []string `json:"codes"`
					NotificationSent bool     `json:"notification_sent"`
				}

				mock.GetResponseData(t, &resp)

				assert.True(t, resp.NotificationSent)
				require.Len(t, resp.Codes, model.RecoveryCodeBatchSize)

				seen := map[string]struct{}{}

				for _, code := range resp.Codes {
					assert.Len(t, code, model.RecoveryCodeLength+model.RecoveryCodeLength/model.RecoveryCodeGroupSize-1)

					seen[code] = struct{}{}
				}

				assert.Len(t, seen, model.RecoveryCodeBatchSize)

				families, err := mock.Ctx.Providers.Metrics.GetGatherer().Gather()
				require.NoError(t, err)

				var value float64

				for _, family := range families {
					if family.GetName() == "authelia_recovery_codes_generated_total" {
						require.Len(t, family.GetMetric(), 1)

						value = family.GetMetric()[0].GetCounter().GetValue()
					}
				}

				assert.Equal(t, float64(model.RecoveryCodeBatchSize), value)
			},
		},
		{
			"ShouldReportNotificationNotSentWhenUserHasNoEmail",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				require.NoError(t, mock.Ctx.SaveSession(&us))

				mock.UserProviderMock.EXPECT().
					GetDetails(testUsername).
					Return(&authentication.UserDetails{Username: testUsername, DisplayName: "John"}, nil)

				mock.StorageMock.EXPECT().
					ReplaceRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername), gomock.Len(model.RecoveryCodeBatchSize), gomock.Any()).
					Return(nil)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

				var resp struct {
					Codes            []string `json:"codes"`
					NotificationSent bool     `json:"notification_sent"`
				}

				mock.GetResponseData(t, &resp)

				assert.False(t, resp.NotificationSent)
				assert.Len(t, resp.Codes, model.RecoveryCodeBatchSize)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to send recovery-codes-generated notification email for user 'john'", "user has no email address")
			},
		},
		{
			"ShouldReportNotificationNotSentWhenUserDetailsLookupFails",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				require.NoError(t, mock.Ctx.SaveSession(&us))

				mock.UserProviderMock.EXPECT().
					GetDetails(testUsername).
					Return(nil, errors.New("ldap unavailable"))

				mock.StorageMock.EXPECT().
					ReplaceRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername), gomock.Len(model.RecoveryCodeBatchSize), gomock.Any()).
					Return(nil)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

				var resp struct {
					NotificationSent bool `json:"notification_sent"`
				}

				mock.GetResponseData(t, &resp)

				assert.False(t, resp.NotificationSent)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Failed to send recovery-codes-generated notification email for user 'john'", "failed to look up user details: ldap unavailable")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			RecoveryCodesGenerationPOST(mock.Ctx)

			tc.expected(t, mock)
		})
	}
}

func TestRecoveryCodesStatusGET(t *testing.T) {
	generated := time.Unix(1700000000, 0).UTC()
	used := time.Unix(1700005000, 0).UTC()

	testCases := []struct {
		name     string
		setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleSessionBackendError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				repository := setupTestFailingSessionRepository(t, mock)
				repository.errGet = errTestSessionBackend

				mock.Ctx.Request.Header.SetCookie("authelia_session", newTestSessionCookie("an-identifier-which-cannot-be-retrieved"))
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred retrieving recovery codes status: error occurred retrieving the user session data", "error occurred getting session from backend: backend unavailable")
			},
		},
		{
			"ShouldHandleAnonymousUser",
			nil,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred retrieving recovery codes status", "user is anonymous")
			},
		},
		{
			"ShouldHandleStorageError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				newRecoveryCodeUserCtx(t, mock)

				mock.StorageMock.EXPECT().
					LoadRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername)).
					Return(nil, errors.New("query failed"))
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert500KO(t, messageOperationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred retrieving recovery codes for user 'john'", "query failed")
			},
		},
		{
			"ShouldReturnEmptyStatusWithNoCodes",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				newRecoveryCodeUserCtx(t, mock)

				mock.StorageMock.EXPECT().
					LoadRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername)).
					Return(nil, nil)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
				assert.JSONEq(t, `{"status":"OK","data":{"codes_remaining":0,"codes_total":0}}`, string(mock.Ctx.Response.Body()))
			},
		},
		{
			"ShouldReturnStatus",
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				newRecoveryCodeUserCtx(t, mock)

				mock.StorageMock.EXPECT().
					LoadRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername)).
					Return([]model.RecoveryCode{
						{ID: 1, CreatedAt: generated},
						{ID: 2, CreatedAt: generated, ConsumedAt: sql.NullTime{Valid: true, Time: used}},
						{ID: 3, CreatedAt: generated},
					}, nil)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
				assert.JSONEq(t, `{"status":"OK","data":{"codes_remaining":2,"codes_total":3,"generated_at":"2023-11-14T22:13:20Z","last_used_at":"2023-11-14T23:36:40Z"}}`, string(mock.Ctx.Response.Body()))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			RecoveryCodesStatusGET(mock.Ctx)

			tc.expected(t, mock)
		})
	}
}

func TestComputeRecoveryCodesStatus(t *testing.T) {
	older := time.Unix(1600000000, 0)
	latest := time.Unix(1700000000, 0)
	usedOld := time.Unix(1650000000, 0)
	usedLatest := time.Unix(1700005000, 0)
	revoked := sql.NullTime{Valid: true, Time: time.Unix(1700000000, 0)}

	testCases := []struct {
		name     string
		have     []model.RecoveryCode
		expected recoveryCodesStatus
	}{
		{
			"ShouldReturnZeroForNil",
			nil,
			recoveryCodesStatus{},
		},
		{
			"ShouldReturnZeroWhenAllRevoked",
			[]model.RecoveryCode{
				{CreatedAt: older, RevokedAt: revoked},
				{CreatedAt: older, RevokedAt: revoked, ConsumedAt: sql.NullTime{Valid: true, Time: usedOld}},
			},
			recoveryCodesStatus{},
		},
		{
			"ShouldCountUnusedBatchWithoutLastUsed",
			[]model.RecoveryCode{
				{CreatedAt: latest},
				{CreatedAt: latest},
			},
			recoveryCodesStatus{CodesRemaining: 2, CodesTotal: 2, GeneratedAt: &latest},
		},
		{
			"ShouldCountOnlyLatestBatchAndLatestUse",
			[]model.RecoveryCode{
				{CreatedAt: older, RevokedAt: revoked, ConsumedAt: sql.NullTime{Valid: true, Time: usedLatest.Add(time.Hour)}},
				{CreatedAt: older, ConsumedAt: sql.NullTime{Valid: true, Time: usedOld}},
				{CreatedAt: older},
				{CreatedAt: latest, ConsumedAt: sql.NullTime{Valid: true, Time: usedLatest}},
				{CreatedAt: latest, ConsumedAt: sql.NullTime{Valid: true, Time: usedOld}},
				{CreatedAt: latest},
			},
			recoveryCodesStatus{CodesRemaining: 1, CodesTotal: 3, GeneratedAt: &latest, LastUsedAt: &usedLatest},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, computeRecoveryCodesStatus(tc.have))
		})
	}
}
