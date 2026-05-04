// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
)

func newRecoveryCodeUserCtx(t *testing.T, mock *mocks.MockAutheliaCtx) {
	t.Helper()

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	us.Username = testUsername
	us.AuthenticationMethodRefs.UsernameAndPassword = true
	us.FirstFactorAuthnTimestamp = time.Now().Unix()
	require.NoError(t, mock.Ctx.SaveSession(us))

	mock.UserProviderMock.EXPECT().
		GetDetails(testUsername).
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
