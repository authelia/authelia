package middlewares_test

import (
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

func handleSetStatus(code int) middlewares.RequestHandler {
	return func(ctx *middlewares.AutheliaCtx) {
		if err := ctx.ReplyJSON(middlewares.ErrorResponse{Status: "OK", Message: "Endpoint Response"}, 0); err != nil {
			ctx.Logger.Error(err)
		}

		ctx.SetStatusCode(code)
		ctx.Response.Header.Set("X-Testing-Success", "yes")
	}
}

func TestProtectedEndpointRequiredLevel(t *testing.T) {
	testCases := []struct {
		name                          string
		level                         authentication.Level
		have                          session.UserSession
		expected, status              int
		expectedAuth, expectedElevate bool
	}{
		{
			name:     "1FAWithAuthenticatedUser2FAShould200OK",
			level:    authentication.OneFactor,
			expected: fasthttp.StatusOK,
			status:   fasthttp.StatusOK,
			have: session.UserSession{
				Username:            "john",
				DisplayName:         "John Wick",
				Emails:              []string{"john.wick@notmessingaround.com"},
				AuthenticationLevel: authentication.TwoFactor,
			},
		},
		{
			name:     "1FAWithAuthenticatedUser1FAShould200OK",
			level:    authentication.OneFactor,
			expected: fasthttp.StatusOK,
			status:   fasthttp.StatusOK,
			have: session.UserSession{
				Username:            "john",
				DisplayName:         "John Wick",
				Emails:              []string{"john.wick@notmessingaround.com"},
				AuthenticationLevel: authentication.OneFactor,
			},
		},
		{
			name:     "1FAWithAuthenticatedUser2FAShould301Found",
			level:    authentication.OneFactor,
			expected: fasthttp.StatusFound,
			status:   fasthttp.StatusFound,
			have: session.UserSession{
				Username:            "john",
				DisplayName:         "John Wick",
				Emails:              []string{"john.wick@notmessingaround.com"},
				AuthenticationLevel: authentication.TwoFactor,
			},
		},
		{
			name:     "1FAWithAuthenticatedUser1FAShould301Found",
			level:    authentication.OneFactor,
			expected: fasthttp.StatusFound,
			status:   fasthttp.StatusFound,
			have: session.UserSession{
				Username:            "john",
				DisplayName:         "John Wick",
				Emails:              []string{"john.wick@notmessingaround.com"},
				AuthenticationLevel: authentication.OneFactor,
			},
		},
		{
			name:     "1FAWithNotAuthenticatedUserShould401Unauthenticated",
			level:    authentication.OneFactor,
			expected: fasthttp.StatusUnauthorized,
			status:   fasthttp.StatusOK,
			have: session.UserSession{
				AuthenticationLevel: authentication.NotAuthenticated,
			},
		},
		{
			name:     "2FAWithNotAuthenticatedUserShould401Unauthenticated",
			level:    authentication.OneFactor,
			expected: fasthttp.StatusUnauthorized,
			status:   fasthttp.StatusFound,
			have: session.UserSession{
				AuthenticationLevel: authentication.NotAuthenticated,
			},
		},
		{
			name:         "2FAWithNotAuthenticatedUserShould401Unauthenticated",
			level:        authentication.TwoFactor,
			expected:     fasthttp.StatusForbidden,
			expectedAuth: true,
			status:       fasthttp.StatusFound,
			have: session.UserSession{
				Username:            "john",
				DisplayName:         "John Wick",
				Emails:              []string{"john.wick@notmessingaround.com"},
				AuthenticationLevel: authentication.OneFactor,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			err := mock.Ctx.SaveSession(tc.have)

			require.NoError(t, err)

			var h middlewares.RequestHandler

			switch tc.level {
			case authentication.OneFactor:
				h = middlewares.Require1FA(handleSetStatus(tc.status))
			default:
				h = middlewares.Require2FA(handleSetStatus(tc.status))
			}

			h(mock.Ctx)

			assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())

			if tc.expected == tc.status {
				assert.Equal(t, `{"status":"OK","message":"Endpoint Response"}`, string(mock.Ctx.Response.Body()))
				assert.Equal(t, []byte("yes"), mock.Ctx.Response.Header.Peek("X-Testing-Success"))
			} else {
				assert.Equal(t, fmt.Sprintf(`{"status":"KO","message":"%s","authentication":%t,"elevation":%t}`, fasthttp.StatusMessage(tc.expected), tc.expectedAuth, tc.expectedElevate), string(mock.Ctx.Response.Body()))
				assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek("X-Testing-Success"))
			}
		})
	}
}

func TestProtectedEndpointOTP(t *testing.T) {
	testCases := []struct {
		name                 string
		characters           int
		emailexp, sessionexp time.Duration
		skip2fa              bool
		have                 session.UserSession
		ip                   net.IP
		time                 time.Time
		expected, status     int
	}{
		{
			name:       "ReturnUnauthorizedForAnonymous",
			characters: 10,
			emailexp:   time.Minute,
			sessionexp: time.Minute,
			skip2fa:    false,
			expected:   fasthttp.StatusUnauthorized,
			status:     fasthttp.StatusFound,
			have: session.UserSession{
				AuthenticationLevel: authentication.NotAuthenticated,
			},
		},
		{
			name:       "Return200OKWhen2FASkipAndUserIs2FAd",
			characters: 10,
			emailexp:   time.Minute,
			sessionexp: time.Minute,
			skip2fa:    true,
			expected:   fasthttp.StatusOK,
			status:     fasthttp.StatusOK,
			have: session.UserSession{
				Username:            "john",
				DisplayName:         "John Wick",
				Emails:              []string{"john.wick@notmessingaround.com"},
				AuthenticationLevel: authentication.TwoFactor,
			},
		},
		{
			name:       "HandleEscalationEmailWhen2FASkipAndUserIs1FAd",
			characters: 10,
			emailexp:   time.Minute,
			sessionexp: time.Minute,
			skip2fa:    true,
			expected:   fasthttp.StatusForbidden,
			status:     fasthttp.StatusOK,
			have: session.UserSession{
				Username:            "john",
				DisplayName:         "John Wick",
				Emails:              []string{"john.wick@notmessingaround.com"},
				AuthenticationLevel: authentication.OneFactor,
				Elevations:          session.Elevations{User: nil},
			},
		},
		{
			name:       "HandleEscalationEmailWhenUserIs2FAd",
			characters: 10,
			emailexp:   time.Minute,
			sessionexp: time.Minute,
			skip2fa:    false,
			expected:   fasthttp.StatusForbidden,
			status:     fasthttp.StatusOK,
			have: session.UserSession{
				Username:            "john",
				DisplayName:         "John Wick",
				Emails:              []string{"john.wick@notmessingaround.com"},
				AuthenticationLevel: authentication.TwoFactor,
				Elevations:          session.Elevations{User: nil},
			},
		},
		{
			name:       "Return200OKWhenUserIsEscalated",
			characters: 10,
			emailexp:   time.Minute,
			sessionexp: time.Minute,
			skip2fa:    false,
			expected:   fasthttp.StatusOK,
			status:     fasthttp.StatusOK,
			ip:         net.ParseIP("192.168.0.1"),
			time:       time.Unix(1671322337, 0),
			have: session.UserSession{
				Username:            "john",
				DisplayName:         "John Wick",
				Emails:              []string{"john.wick@notmessingaround.com"},
				AuthenticationLevel: authentication.TwoFactor,
				Elevations: session.Elevations{
					User: &session.Elevation{
						RemoteIP: net.ParseIP("192.168.0.1"),
						Expires:  time.Unix(1671322347, 0),
					},
				},
			},
		},
		{
			name:       "Return403ForbiddenWhenUserIsEscalatedButInvalidIP",
			characters: 10,
			emailexp:   time.Minute,
			sessionexp: time.Minute,
			skip2fa:    false,
			expected:   fasthttp.StatusForbidden,
			status:     fasthttp.StatusOK,
			ip:         net.ParseIP("192.168.0.2"),
			time:       time.Unix(1671322337, 0),
			have: session.UserSession{
				Username:            "john",
				DisplayName:         "John Wick",
				Emails:              []string{"john.wick@notmessingaround.com"},
				AuthenticationLevel: authentication.TwoFactor,
				Elevations: session.Elevations{
					User: &session.Elevation{
						RemoteIP: net.ParseIP("192.168.0.1"),
						Expires:  time.Unix(1671322347, 0),
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, tc.ip.String())

			err := mock.Ctx.SaveSession(tc.have)

			require.NoError(t, err)

			if !tc.time.IsZero() {
				mock.Clock.Set(tc.time)
				mock.Ctx.Clock = &mock.Clock
			}

			h := middlewares.ProtectedEndpoint(middlewares.NewOTPEscalationProtectedEndpointHandler(middlewares.OTPEscalationProtectedEndpointConfig{
				Characters:                 tc.characters,
				EmailValidityDuration:      tc.emailexp,
				EscalationValidityDuration: tc.sessionexp,
				Skip2FA:                    tc.skip2fa,
			}))(handleSetStatus(tc.status))

			h(mock.Ctx)

			switch {
			case tc.have.IsAnonymous():
				assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())
				assert.Equal(t, fmt.Sprintf(`{"status":"KO","message":"%s","authentication":false,"elevation":false}`, fasthttp.StatusMessage(tc.expected)), string(mock.Ctx.Response.Body()))
				assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek("X-Testing-Success"))
			case tc.skip2fa && tc.have.AuthenticationLevel == authentication.TwoFactor:
				assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())
				assert.Equal(t, `{"status":"OK","message":"Endpoint Response"}`, string(mock.Ctx.Response.Body()))
				assert.Equal(t, []byte("yes"), mock.Ctx.Response.Header.Peek("X-Testing-Success"))
			case tc.have.Elevations.User == nil || mock.Ctx.Clock.Now().After(tc.have.Elevations.User.Expires) || !tc.ip.Equal(tc.have.Elevations.User.RemoteIP):
				assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())
				assert.Equal(t, `{"status":"KO","message":"Forbidden","authentication":false,"elevation":true}`, string(mock.Ctx.Response.Body()))
				assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek("X-Testing-Success"))
			default:
				assert.Equal(t, tc.expected, mock.Ctx.Response.StatusCode())
				assert.Equal(t, `{"status":"OK","message":"Endpoint Response"}`, string(mock.Ctx.Response.Body()))
				assert.Equal(t, []byte("yes"), mock.Ctx.Response.Header.Peek("X-Testing-Success"))
			}
		})
	}
}
