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

func NilHandler(ctx *middlewares.AutheliaCtx) {
	ctx.SetContentTypeTextPlain()
	ctx.Response.SetBodyString("Example Nil")
}
