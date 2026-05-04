// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestRecoveryCodesGenerationPOST_SurfacesSMTPFailure(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	newRecoveryCodeUserCtx(t, mock)

	mock.StorageMock.EXPECT().
		RevokeRecoveryCodesByUsername(gomock.Any(), gomock.Eq(testUsername), gomock.Any()).
		Return(nil)

	mock.StorageMock.EXPECT().
		SaveRecoveryCode(gomock.Any(), gomock.Any()).
		Times(model.RecoveryCodeBatchSize).
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
