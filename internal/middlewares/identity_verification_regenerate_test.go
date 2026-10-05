// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestShouldFailIfSessionCannotBeRegenerated(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Providers.Session = nil

	middlewares.IdentityVerificationStart(newArgs(defaultRetriever), nil)(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

	mock.AssertLastLogMessage(t, "Error occurred regenerating user session", "unable to regenerate user session: unable to retrieve session cookie domain provider: no session provider is configured")
}
