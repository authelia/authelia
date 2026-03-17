// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/regulation"
)

func TestFirstFactorPasswordPOSTSessionErrors(t *testing.T) {
	testCases := []struct {
		name           string
		setup          func(t *testing.T, mock *mocks.MockAutheliaCtx, repository *failingSessionRepository)
		expectedStatus int
		expectedf      func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldContinueWhenDestroyFails",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, repository *failingSessionRepository) {
				mock.Ctx.Request.Header.SetCookie("authelia_session", "an-identifier-which-was-never-saved")

				repository.errDelete = errTestSessionBackend
			},
			fasthttp.StatusOK,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, `{"status":"OK"}`, string(mock.Ctx.Response.Body()))
			},
		},
		{
			"ShouldHandleResetSaveError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, repository *failingSessionRepository) {
				repository.errSave = []error{errTestSessionBackend}
			},
			fasthttp.StatusUnauthorized,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), fmt.Sprintf(logFmtErrSessionReset, regulation.AuthType1FA, testValue), "error occurred saving session to registry: backend unavailable")
			},
		},
		{
			"ShouldHandleRegenerateError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, repository *failingSessionRepository) {
				repository.errChangeID = errTestSessionBackend
			},
			fasthttp.StatusUnauthorized,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), fmt.Sprintf(logFmtErrSessionRegenerate, regulation.AuthType1FA, testValue), "error occurred changing session ID: backend unavailable")
			},
		},
		{
			"ShouldHandleProfileSaveError",
			func(t *testing.T, mock *mocks.MockAutheliaCtx, repository *failingSessionRepository) {
				repository.errSave = []error{nil, errTestSessionBackend}
			},
			fasthttp.StatusUnauthorized,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), fmt.Sprintf(logFmtErrSessionSave, "updated profile", regulation.AuthType1FA, logFmtActionAuthentication, testValue), "error occurred saving session to registry: backend unavailable")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			repository := setupTestFailingSessionRepository(t, mock)

			mock.UserProviderMock.
				EXPECT().
				GetDetails(gomock.Eq(testValue)).
				Return(&authentication.UserDetails{Username: testValue, Emails: []string{"test@example.com"}, Groups: []string{"dev", "admins"}}, nil)

			mock.StorageMock.
				EXPECT().
				LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
				Return(nil, nil)

			mock.StorageMock.
				EXPECT().
				LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq(testValue)).
				Return(nil, nil)

			mock.UserProviderMock.
				EXPECT().
				CheckUserPassword(gomock.Eq(testValue), gomock.Eq("hello")).
				Return(true, nil)

			mock.StorageMock.
				EXPECT().
				AppendAuthenticationLog(mock.Ctx, gomock.Any()).
				Return(nil)

			mock.Ctx.Request.SetBodyString(`{"username":"test","password":"hello","requestMethod":"GET","keepMeLoggedIn":false}`)

			tc.setup(t, mock, repository)

			FirstFactorPasswordPOST(nil)(mock.Ctx)

			assert.Equal(t, tc.expectedStatus, mock.Ctx.Response.StatusCode())

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestFirstFactorReauthenticatePOSTShouldHandleFlow(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	us, err := mock.Ctx.GetSession()

	require.NoError(t, err)

	us.Username = testUsername
	us.AuthenticationMethodRefs.UsernameAndPassword = true

	require.NoError(t, mock.Ctx.SaveSession(&us))

	mock.StorageMock.
		EXPECT().
		LoadBannedIP(gomock.Eq(mock.Ctx), gomock.Eq(model.NewIP(mock.Ctx.RemoteIP()))).
		Return(nil, nil)

	mock.StorageMock.
		EXPECT().
		LoadBannedUser(gomock.Eq(mock.Ctx), gomock.Eq(testUsername)).
		Return(nil, nil)

	mock.UserProviderMock.
		EXPECT().
		CheckUserPassword(gomock.Eq(testUsername), gomock.Eq("hello")).
		Return(true, nil)

	mock.StorageMock.
		EXPECT().
		AppendAuthenticationLog(mock.Ctx, gomock.Any()).
		Return(nil)

	mock.UserProviderMock.
		EXPECT().
		GetDetails(gomock.Eq(testUsername)).
		Return(&authentication.UserDetails{Username: testUsername, Emails: []string{testEmail}}, nil)

	mock.Ctx.Request.SetBodyString(`{"password":"hello","flow":"not-a-flow"}`)

	FirstFactorReauthenticatePOST(nil)(mock.Ctx)

	assert.NotEqual(t, fasthttp.StatusUnauthorized, mock.Ctx.Response.StatusCode())
}

func TestFirstFactorPasskeyPOSTSessionErrors(t *testing.T) {
	t.Run("ShouldHandleSessionBackendError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "login.example.com:8080")

		repository := setupTestFailingSessionRepository(t, mock)
		repository.errGet = errTestSessionBackend

		mock.Ctx.Request.Header.SetCookie("authelia_session", "an-identifier-which-cannot-be-retrieved")

		FirstFactorPasskeyPOST(mock.Ctx)

		assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), fmt.Sprintf(logFmtErrPasskeyAuthenticationChallengeValidate, errStrUserSessionData), "error occurred getting session from backend: backend unavailable")
	})

	t.Run("ShouldLogSessionSaveError", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)

		defer mock.Close()

		mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedHost, "login.example.com:8080")

		repository := setupTestFailingSessionRepository(t, mock)
		repository.errSave = []error{errTestSessionBackend}

		mock.StorageMock.
			EXPECT().
			AppendAuthenticationLog(mock.Ctx, gomock.Any()).
			Return(nil)

		mock.Ctx.Request.SetBodyString(`{`)

		FirstFactorPasskeyPOST(mock.Ctx)

		assert.Equal(t, fasthttp.StatusBadRequest, mock.Ctx.Response.StatusCode())

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), fmt.Sprintf(logFmtErrPasskeyAuthenticationChallengeValidateUser, "", errStrUserSessionDataSave), "error occurred saving session to registry: backend unavailable")
	})
}
