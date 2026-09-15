// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestRequireCSRFToken(t *testing.T) {
	validToken := func(t *testing.T, mock *mocks.MockAutheliaCtx) string {
		provider, err := mock.Ctx.GetSessionProvider()
		require.NoError(t, err)

		token, err := provider.CSRFToken(mock.Ctx)
		require.NoError(t, err)
		require.NotEmpty(t, token)

		return token
	}

	testCases := []struct {
		name     string
		method   string
		session  bool
		token    func(t *testing.T, mock *mocks.MockAutheliaCtx) string
		expected int
	}{
		{
			"ShouldAllowSafeMethodWithoutToken",
			fasthttp.MethodGet,
			true,
			nil,
			fasthttp.StatusOK,
		},
		{
			"ShouldAllowUnsafeMethodWithoutSessionCookie",
			fasthttp.MethodPost,
			false,
			nil,
			fasthttp.StatusOK,
		},
		{
			"ShouldAllowUnsafeMethodWithValidToken",
			fasthttp.MethodPost,
			true,
			validToken,
			fasthttp.StatusOK,
		},
		{
			"ShouldAllowDeleteMethodWithValidToken",
			fasthttp.MethodDelete,
			true,
			validToken,
			fasthttp.StatusOK,
		},
		{
			"ShouldRejectUnsafeMethodWithoutToken",
			fasthttp.MethodPost,
			true,
			nil,
			fasthttp.StatusForbidden,
		},
		{
			"ShouldRejectUnsafeMethodWithInvalidToken",
			fasthttp.MethodPost,
			true,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) string {
				return "0000000000000000000000000000000000000000000000000000000000000000"
			},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldRejectUnsafeMethodWithTokenOfPreviousSessionIdentifier",
			fasthttp.MethodPost,
			true,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) string {
				token := validToken(t, mock)

				require.NoError(t, mock.Ctx.RegenerateSession())

				return token
			},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldRejectUnsafeMethodWithTokenOfPreviousCSRFSecret",
			fasthttp.MethodPost,
			true,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) string {
				token := validToken(t, mock)

				provider, err := mock.Ctx.GetSessionProvider()
				require.NoError(t, err)

				require.NoError(t, provider.RegenerateCSRFToken(mock.Ctx))

				return token
			},
			fasthttp.StatusForbidden,
		},
		{
			"ShouldAllowUnsafeMethodWithUnknownSessionCookie",
			fasthttp.MethodPost,
			false,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) string {
				mock.Ctx.Request.Header.SetCookie("authelia_session", "an-unknown-session-cookie-value")

				return ""
			},
			fasthttp.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Request.Header.SetMethod(tc.method)

			if tc.session {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				userSession.Username = "john"

				require.NoError(t, mock.Ctx.SaveSession(&userSession))
			}

			if tc.token != nil {
				mock.Ctx.Request.Header.Set("X-CSRF-Token", tc.token(t, mock))
			}

			called := false

			middlewares.RequireCSRFToken(func(ctx *middlewares.AutheliaCtx) {
				called = true
			})(mock.Ctx)

			assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected == fasthttp.StatusOK, called)
		})
	}
}

func TestRequireCSRFTokenShouldRejectWhenSessionProviderIsUnavailable(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Request.Header.SetMethod(fasthttp.MethodPost)
	mock.Ctx.Providers.Session = nil

	called := false

	middlewares.RequireCSRFToken(func(ctx *middlewares.AutheliaCtx) {
		called = true
	})(mock.Ctx)

	assert.False(t, called)
	assert.Equal(t, fasthttp.StatusForbidden, mock.Ctx.Response.StatusCode())

	mock.AssertLastLogMessage(t, "Error occurred retrieving the session provider to verify the CSRF token", "unable to retrieve session cookie domain provider: no session provider is configured")
}
