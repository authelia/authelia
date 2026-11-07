// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares_test

import (
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestRequireElevated(t *testing.T) {
	type elevation struct {
		id      int
		expires time.Duration
		ip      net.IP
	}

	type response struct {
		Status string                                `json:"status"`
		Data   middlewares.ElevatedForbiddenResponse `json:"data"`
	}

	testCases := []struct {
		name              string
		level             authentication.Level
		elevation         *elevation
		require2FA        bool
		skip2fA           bool
		setup             func(t *testing.T, mock *mocks.MockAutheliaCtx)
		expected          int
		expected1FA       bool
		expected2FA       bool
		expectedElevation bool
	}{
		{
			"ShouldPassAuthenticatedElevatedUser",
			authentication.OneFactor,
			&elevation{
				1, time.Minute, net.ParseIP("127.0.0.1"),
			},
			false,
			false,
			nil,
			fasthttp.StatusOK,
			false,
			false,
			false,
		},
		{
			"ShouldRequireElevation",
			authentication.OneFactor,
			nil,
			false,
			false,
			nil,
			fasthttp.StatusForbidden,
			false,
			false,
			true,
		},
		{
			"ShouldRequireElevationExpired",
			authentication.OneFactor,
			&elevation{
				1, time.Minute * -1, net.ParseIP("127.0.0.1"),
			},
			false,
			false,
			nil,
			fasthttp.StatusForbidden,
			false,
			false,
			true,
		},
		{
			"ShouldRequireElevationBadIP",
			authentication.OneFactor,
			&elevation{
				1, time.Minute, net.ParseIP("127.0.0.2"),
			},
			false,
			false,
			nil,
			fasthttp.StatusForbidden,
			false,
			false,
			true,
		},
		{
			"ShouldRequire2FAWhenElevated",
			authentication.OneFactor,
			&elevation{
				1, time.Minute, net.ParseIP("127.0.0.1"),
			},
			true,
			false,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).
					Return(model.UserInfo{HasWebAuthn: true}, nil)
			},
			fasthttp.StatusForbidden,
			false,
			true,
			false,
		},
		{
			"ShouldNotRequire2FAWhenNotSetup",
			authentication.OneFactor,
			&elevation{
				1, time.Minute, net.ParseIP("127.0.0.1"),
			},
			true,
			false,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).
					Return(model.UserInfo{}, nil)
			},
			fasthttp.StatusOK,
			false,
			false,
			false,
		},
		{
			"ShouldRequire2FAWhenError",
			authentication.OneFactor,
			&elevation{
				1, time.Minute, net.ParseIP("127.0.0.1"),
			},
			true,
			false,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).
					Return(model.UserInfo{}, errors.New("example"))
			},
			fasthttp.StatusForbidden,
			false,
			true,
			false,
		},
		{
			"ShouldPass2FAUser",
			authentication.TwoFactor,
			nil,
			false,
			true,
			nil,
			fasthttp.StatusOK,
			false,
			false,
			false,
		},
		{
			"ShouldRequireElevation1FAUser",
			authentication.OneFactor,
			nil,
			false,
			true,
			nil,
			fasthttp.StatusForbidden,
			false,
			false,
			true,
		},
		{
			"ShouldRequireAuthentication",
			authentication.NotAuthenticated,
			nil,
			false,
			true,
			nil,
			fasthttp.StatusForbidden,
			true,
			false,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.IdentityValidation.ElevatedSession = schema.IdentityValidationElevatedSession{
				CodeLifespan:        time.Minute,
				ElevationLifespan:   time.Minute,
				Characters:          8,
				RequireSecondFactor: tc.require2FA,
				SkipSecondFactor:    tc.skip2fA,
			}

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "127.0.0.1")

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			switch tc.level {
			case authentication.OneFactor:
				userSession.Username = john
				userSession.AuthenticationMethodRefs.UsernameAndPassword = true
				userSession.AuthenticationMethodRefs.WebAuthn = false
			case authentication.TwoFactor:
				userSession.Username = john
				userSession.AuthenticationMethodRefs.UsernameAndPassword = true
				userSession.AuthenticationMethodRefs.WebAuthn = true
			case authentication.NotAuthenticated:
				userSession.AuthenticationMethodRefs.UsernameAndPassword = false
				userSession.AuthenticationMethodRefs.WebAuthn = false
			}

			if tc.elevation != nil {
				userSession.Elevations.User = &session.Elevation{
					ID:       tc.elevation.id,
					Expires:  mock.Clock.Now().Add(tc.elevation.expires),
					RemoteIP: tc.elevation.ip,
				}
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			if tc.setup != nil {
				tc.setup(t, mock)
			}

			handler := middlewares.RequireElevated(NilHandler)

			handler(mock.Ctx)

			assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())

			if tc.expected == fasthttp.StatusOK {
				assert.Equal(t, "text/plain; charset=utf-8", string(mock.Ctx.Response.Header.Peek(fasthttp.HeaderContentType)))
				assert.Equal(t, "Example Nil", string(mock.Ctx.Response.Body()))
			} else {
				data := &response{}

				require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), data))

				assert.Equal(t, tc.expectedElevation, data.Data.Elevation)
				assert.Equal(t, tc.expected1FA, data.Data.FirstFactor)
				assert.Equal(t, tc.expected2FA, data.Data.SecondFactor)
			}
		})
	}
}

func TestRequireElevatedReauthentication(t *testing.T) {
	type response struct {
		Status string                                `json:"status"`
		Data   middlewares.ElevatedForbiddenResponse `json:"data"`
	}

	now := time.Unix(1700000000, 0)

	testCases := []struct {
		name                     string
		mode                     string
		level                    authentication.Level
		first                    int64
		second                   int64
		elevated                 bool
		skip2FA                  bool
		setup                    func(mock *mocks.MockAutheliaCtx)
		expected                 int
		expectedReauthentication bool
		expectedElevation        bool
	}{
		{
			name:                     "ShouldForbidStalePassword",
			mode:                     schema.ElevatedSessionReauthenticationPassword,
			level:                    authentication.OneFactor,
			first:                    now.Add(-time.Hour).Unix(),
			elevated:                 true,
			expected:                 fasthttp.StatusForbidden,
			expectedReauthentication: true,
		},
		{
			name:     "ShouldPassFreshPasswordElevated",
			mode:     schema.ElevatedSessionReauthenticationPassword,
			level:    authentication.OneFactor,
			first:    now.Add(-time.Minute).Unix(),
			elevated: true,
			expected: fasthttp.StatusOK,
		},
		{
			name:              "ShouldRequireElevationAfterFreshPassword",
			mode:              schema.ElevatedSessionReauthenticationPassword,
			level:             authentication.OneFactor,
			first:             now.Add(-time.Minute).Unix(),
			expected:          fasthttp.StatusForbidden,
			expectedElevation: true,
		},
		{
			name:    "ShouldEvaluateGateBeforeSkipSecondFactor",
			mode:    schema.ElevatedSessionReauthenticationSecondFactor,
			level:   authentication.TwoFactor,
			first:   now.Add(-time.Minute).Unix(),
			second:  now.Add(-time.Hour).Unix(),
			skip2FA: true,
			setup: func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{HasWebAuthn: true}, nil)
			},
			expected:                 fasthttp.StatusForbidden,
			expectedReauthentication: true,
		},
		{
			name:    "ShouldPassSkipSecondFactorWithFreshSecondFactor",
			mode:    schema.ElevatedSessionReauthenticationSecondFactor,
			level:   authentication.TwoFactor,
			second:  now.Add(-time.Minute).Unix(),
			skip2FA: true,
			setup: func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{HasWebAuthn: true}, nil)
			},
			expected: fasthttp.StatusOK,
		},
		{
			name:     "ShouldFailClosedOnStorageError",
			mode:     schema.ElevatedSessionReauthenticationSecondFactor,
			level:    authentication.OneFactor,
			second:   now.Add(-time.Minute).Unix(),
			elevated: true,
			setup: func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadUserInfo(mock.Ctx, john).Return(model.UserInfo{}, errors.New("bad storage"))
			},
			expected:                 fasthttp.StatusForbidden,
			expectedReauthentication: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.IdentityValidation.ElevatedSession = schema.IdentityValidationElevatedSession{
				CodeLifespan:             time.Minute,
				ElevationLifespan:        time.Minute,
				Characters:               8,
				SkipSecondFactor:         tc.skip2FA,
				RequireReauthentication:  tc.mode,
				ReauthenticationLifespan: time.Minute * 5,
			}

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Clock.Set(now)
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "127.0.0.1")

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.Username = john
			userSession.AuthenticationMethodRefs.UsernameAndPassword = true
			userSession.AuthenticationMethodRefs.WebAuthn = tc.level == authentication.TwoFactor
			userSession.FirstFactorAuthnTimestamp = tc.first
			userSession.SecondFactorAuthnTimestamp = tc.second
			userSession.SecondFactorPossessionAuthnTimestamp = tc.second

			if tc.elevated {
				userSession.Elevations.User = &session.Elevation{
					ID:       1,
					Expires:  now.Add(time.Minute),
					RemoteIP: net.ParseIP("127.0.0.1"),
				}
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			if tc.setup != nil {
				tc.setup(mock)
			}

			middlewares.RequireElevated(NilHandler)(mock.Ctx)

			assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())

			if tc.expected != fasthttp.StatusOK {
				data := &response{}

				require.NoError(t, json.Unmarshal(mock.Ctx.Response.Body(), data))

				assert.Equal(t, tc.expectedReauthentication, data.Data.Reauthentication)
				assert.Equal(t, tc.expectedElevation, data.Data.Elevation)
			}
		})
	}
}

func NilHandler(ctx *middlewares.AutheliaCtx) {
	ctx.SetContentTypeTextPlain()
	ctx.Response.SetBodyString("Example Nil")
}
