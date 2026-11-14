// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/storage"
)

func newRecoveryCodeUserCtx(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	us.Username = testUsername
	us.AuthenticationMethodRefs.UsernameAndPassword = true
	us.FirstFactorAuthnTimestamp = time.Now().Unix()
	require.NoError(t, mock.Ctx.SaveSession(&us))

	mock.UserProviderMock.EXPECT().
		GetDetails(testUsername).
		AnyTimes().
		Return(&authentication.UserDetails{Username: testUsername, DisplayName: "John", Emails: []string{"john@example.com"}}, nil)

	mock.UserProviderMock.EXPECT().
		GetDetailsCached(testUsername).
		AnyTimes().
		Return(&authentication.UserDetails{Username: testUsername, DisplayName: "John", Emails: []string{"john@example.com"}}, nil)
}

func TestRecoveryCodePOST_SuccessBumpsTwoFactor_NotElevation(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	newRecoveryCodeUserCtx(t, mock)

	stored := &model.RecoveryCode{
		ID:        42,
		Username:  testUsername,
		Signature: "deadbeef",
		CreatedAt: time.Unix(1700000000, 0),
	}

	mock.StorageMock.EXPECT().
		LoadRecoveryCode(gomock.Any(), gomock.Eq(testUsername), gomock.Any()).
		Return(stored, nil)

	mock.StorageMock.EXPECT().
		ConsumeRecoveryCode(gomock.Any(), gomock.Eq(42), gomock.Any()).
		Return(nil)

	mock.StorageMock.EXPECT().
		AppendAuthenticationLog(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil)

	mock.StorageMock.EXPECT().
		CountUnusedRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername)).
		AnyTimes().
		Return(7, nil)

	mock.NotifierMock.EXPECT().
		Send(gomock.Any(), gomock.Any(), gomock.Eq("Recovery code used"), gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil)

	mock.Ctx.Request.SetBody([]byte(`{"code":"XYZ-1234"}`))

	RecoveryCodePOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	assert.True(t, us.AuthenticationMethodRefs.RecoveryCode, "AMR should record recovery code")
	assert.NotZero(t, us.SecondFactorAuthnTimestamp, "TwoFactor timestamp should be stamped")
	assert.Nil(t, us.Elevations.User, "Recovery code success must NOT mint an elevated session")
}

func TestRecoveryCodePOST_RejectsAlreadyConsumedCode(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	newRecoveryCodeUserCtx(t, mock)

	consumed := &model.RecoveryCode{
		ID:         99,
		Username:   testUsername,
		Signature:  "deadbeef",
		CreatedAt:  time.Unix(1700000000, 0),
		ConsumedAt: sql.NullTime{Valid: true, Time: time.Unix(1700001000, 0)},
	}

	mock.StorageMock.EXPECT().
		LoadRecoveryCode(gomock.Any(), gomock.Eq(testUsername), gomock.Any()).
		Return(consumed, nil)

	mock.StorageMock.EXPECT().
		AppendAuthenticationLog(gomock.Any(), gomock.Any()).
		AnyTimes().
		Return(nil)

	mock.Ctx.Request.SetBody([]byte(`{"code":"XYZ-1234"}`))

	RecoveryCodePOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
}

func TestRecoveryCodePOST_RejectsAnonymous(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	mock.Ctx.Request.SetBody([]byte(`{"code":"XYZ-1234"}`))

	RecoveryCodePOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())
}

func TestRecoveryCodePOST_RejectsWithoutKnowledgeFactor(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	us.Username = testUsername
	us.AuthenticationMethodRefs.WebAuthn = true
	us.AuthenticationMethodRefs.WebAuthnHardware = true
	us.FirstFactorAuthnTimestamp = time.Now().Unix()
	require.NoError(t, mock.Ctx.SaveSession(&us))

	mock.Ctx.Request.SetBody([]byte(`{"code":"XYZ-1234"}`))

	RecoveryCodePOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())

	us, err = mock.Ctx.GetSession()
	require.NoError(t, err)

	assert.False(t, us.AuthenticationMethodRefs.RecoveryCode)
	assert.Equal(t, int64(0), us.SecondFactorAuthnTimestamp)
}

func TestRecoveryCodePOST_RejectsWhenConsumeFails(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	newRecoveryCodeUserCtx(t, mock)

	stored := &model.RecoveryCode{
		ID:        42,
		Username:  testUsername,
		Signature: "deadbeef",
		CreatedAt: time.Unix(1700000000, 0),
	}

	mock.StorageMock.EXPECT().
		LoadRecoveryCode(gomock.Any(), gomock.Eq(testUsername), gomock.Any()).
		Return(stored, nil)

	mock.StorageMock.EXPECT().
		ConsumeRecoveryCode(gomock.Any(), gomock.Eq(42), gomock.Any()).
		Return(storage.ErrRecoveryCodeNotActive)

	mock.Ctx.Request.SetBody([]byte(`{"code":"XYZ-1234"}`))

	RecoveryCodePOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	assert.False(t, us.AuthenticationMethodRefs.RecoveryCode)
	assert.Zero(t, us.SecondFactorAuthnTimestamp)
}

func TestRecoveryCodePOST(t *testing.T) {
	stored := func() *model.RecoveryCode {
		return &model.RecoveryCode{
			ID:        42,
			Username:  testUsername,
			Signature: "deadbeef",
			CreatedAt: time.Unix(1700000000, 0),
		}
	}

	expectSuccessfulConsume := func(mock *mocks.MockAutheliaCtx) {
		mock.StorageMock.EXPECT().
			LoadRecoveryCode(gomock.Any(), gomock.Eq(testUsername), gomock.Eq("XYZ-1234")).
			Return(stored(), nil)

		mock.StorageMock.EXPECT().
			ConsumeRecoveryCode(gomock.Any(), gomock.Eq(42), gomock.Any()).
			Return(nil)

		mock.StorageMock.EXPECT().
			AppendAuthenticationLog(gomock.Any(), gomock.Any()).
			Return(nil)
	}

	testCases := []struct {
		name     string
		body     string
		setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleSessionBackendError",
			`{"code":"XYZ-1234"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				repository := setupTestFailingSessionRepository(t, mock)
				repository.errGet = errTestSessionBackend

				mock.Ctx.Request.Header.SetCookie("authelia_session", newTestSessionCookie("an-identifier-which-cannot-be-retrieved"))
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a recovery code authentication: error occurred retrieving the user session data", "error occurred getting session from backend: backend unavailable")
			},
		},
		{
			"ShouldHandleBodyParseError",
			`{`,
			newRecoveryCodeUserCtx,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a recovery code authentication for user 'john': error parsing the request body", regexp.MustCompile(`.+`))
			},
		},
		{
			"ShouldHandleEmptyCodeAfterNormalization",
			`{"code":" - _ "}`,
			newRecoveryCodeUserCtx,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.AssertKO(t, messageMFAValidationFailed, fasthttp.StatusBadRequest)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a recovery code authentication for user 'john': empty code after normalization", "")
			},
		},
		{
			"ShouldHandleStorageLoadError",
			`{"code":"XYZ-1234"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				newRecoveryCodeUserCtx(t, mock)

				mock.StorageMock.EXPECT().
					LoadRecoveryCode(gomock.Any(), gomock.Eq(testUsername), gomock.Eq("XYZ-1234")).
					Return(nil, errors.New("query failed"))
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a recovery code authentication for user 'john': error occurred retrieving the recovery code from the storage backend", "query failed")
			},
		},
		{
			"ShouldRejectUnknownCode",
			`{"code":"XYZ-1234"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				newRecoveryCodeUserCtx(t, mock)

				mock.StorageMock.EXPECT().
					LoadRecoveryCode(gomock.Any(), gomock.Eq(testUsername), gomock.Eq("XYZ-1234")).
					Return(nil, nil)

				mock.StorageMock.EXPECT().
					AppendAuthenticationLog(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a recovery code authentication for user 'john'", "the user input did not match any active recovery code")
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Unsuccessful Recovery authentication attempt by user 'john'", "")
			},
		},
		{
			"ShouldRejectRevokedCode",
			`{"code":"XYZ-1234"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				newRecoveryCodeUserCtx(t, mock)

				code := stored()
				code.RevokedAt = sql.NullTime{Valid: true, Time: time.Unix(1700001000, 0)}

				mock.StorageMock.EXPECT().
					LoadRecoveryCode(gomock.Any(), gomock.Eq(testUsername), gomock.Eq("XYZ-1234")).
					Return(code, nil)

				mock.StorageMock.EXPECT().
					AppendAuthenticationLog(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.LogEntryN(1), "Error occurred validating a recovery code authentication for user 'john'", "the recovery code has been revoked")
			},
		},
		{
			"ShouldHandleUserDetailsError",
			`{"code":"XYZ-1234"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				us, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				us.Username = testUsername
				us.AuthenticationMethodRefs.UsernameAndPassword = true
				require.NoError(t, mock.Ctx.SaveSession(&us))

				mock.StorageMock.EXPECT().
					LoadRecoveryCode(gomock.Any(), gomock.Eq(testUsername), gomock.Eq("XYZ-1234")).
					Return(stored(), nil)

				mock.UserProviderMock.EXPECT().
					GetDetails(testUsername).
					Return(nil, errors.New("ldap unavailable"))
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a recovery code authentication for user 'john': error occurred fetching user details for user", "ldap unavailable")

				us, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.False(t, us.AuthenticationMethodRefs.RecoveryCode)
			},
		},
		{
			"ShouldHandleRegenerateSessionError",
			`{"code":"XYZ-1234"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				repository := setupTestFailingSessionRepository(t, mock)

				newRecoveryCodeUserCtx(t, mock)

				repository.errChangeID = errTestSessionBackend

				expectSuccessfulConsume(mock)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a recovery code authentication for user 'john': error regenerating the user session", "error occurred changing session ID: backend unavailable")
			},
		},
		{
			"ShouldHandleSaveSessionError",
			`{"code":"XYZ-1234"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				repository := setupTestFailingSessionRepository(t, mock)

				newRecoveryCodeUserCtx(t, mock)

				repository.errSave = []error{errTestSessionBackend}

				expectSuccessfulConsume(mock)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert403KO(t, messageMFAValidationFailed)

				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred validating a recovery code authentication for user 'john': error occurred saving the user session data", regexp.MustCompile(`backend unavailable`))
			},
		},
		{
			"ShouldWarnWhenRunningLowOnCodes",
			`{"code":"XYZ-1234"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				newRecoveryCodeUserCtx(t, mock)

				expectSuccessfulConsume(mock)

				mock.StorageMock.EXPECT().
					CountUnusedRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername)).
					Return(2, nil)

				mock.NotifierMock.EXPECT().
					Send(gomock.Any(), gomock.Any(), gomock.Eq("Recovery code used"), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

				mock.AssertLogEntryAdvanced(t, 0, logrus.WarnLevel, "User has signed in with a recovery code and is running low on remaining codes", map[string]any{"username": testUsername, "remaining": regexp.MustCompile(`^2$`)})

				us, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, us.AuthenticationMethodRefs.RecoveryCode)
			},
		},
		{
			"ShouldHandleUnknownFlow",
			`{"code":"XYZ-1234","flow":"unknown","flowID":"abc"}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				newRecoveryCodeUserCtx(t, mock)

				expectSuccessfulConsume(mock)

				mock.StorageMock.EXPECT().
					CountUnusedRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername)).
					Return(0, errors.New("count failed"))

				mock.NotifierMock.EXPECT().
					Send(gomock.Any(), gomock.Any(), gomock.Eq("Recovery code used"), gomock.Any(), gomock.Any()).
					Return(nil)
			},
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.Assert200KO(t, messageAuthenticationFailed)

				mock.AssertLogEntryAdvanced(t, 0, logrus.ErrorLevel, "Failed to find flow handler for the given flow parameters", map[string]any{"flow_id": "abc", "flow": "unknown"})

				for _, entry := range mock.Hook.AllEntries() {
					assert.NotEqual(t, logrus.WarnLevel, entry.Level)
				}

				us, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.True(t, us.AuthenticationMethodRefs.RecoveryCode)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "login.example.com:8080")

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			mock.Ctx.Request.SetBodyString(tc.body)

			RecoveryCodePOST(mock.Ctx)

			tc.expected(t, mock)
		})
	}
}
