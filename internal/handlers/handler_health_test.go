// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

var okMessageBytes = []byte("{\"status\":\"OK\"}")

func TestHealthOk(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{
		Username:            "john",
		AuthenticationLevel: authentication.OneFactor,
	})
	defer mock.Close()

	HealthGET(mock.Ctx)

	assert.Equal(t, 200, mock.Ctx.Response.StatusCode())
	assert.Equal(t, okMessageBytes, mock.Ctx.Response.Body())
}
