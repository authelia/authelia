package handlers

import (
	"database/sql"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestUserOpenIDConnectLinksGET(t *testing.T) {
	testCases := []struct {
		Name    string
		Pending *session.OpenIDConnectPending
		Setup   func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Assert  func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			Name: "ShouldReturnLinksWithoutPending",
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinksByUsername(mock.Ctx, "john").
					Return([]model.OpenIDConnectLink{{ID: 1, Provider: "example", Issuer: "https://op.example.com", Subject: "abc123", Username: "john", RemoteUsername: sql.NullString{String: "john", Valid: true}}}, nil)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Contains(t, string(mock.Ctx.Response.Body()), `"subject":"abc123"`)
				assert.Contains(t, string(mock.Ctx.Response.Body()), `"provider":"example","provider_name":"Example"`)
				assert.NotContains(t, string(mock.Ctx.Response.Body()), `"pending"`)
				assert.NotContains(t, string(mock.Ctx.Response.Body()), `"username":`)
			},
		},
		{
			Name: "ShouldFallBackToTheProviderIDWhenTheProviderIsNoLongerConfigured",
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinksByUsername(mock.Ctx, "john").
					Return([]model.OpenIDConnectLink{{ID: 1, Provider: "removed", Issuer: "https://op.example.com", Subject: "abc123", Username: "john"}}, nil)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Contains(t, string(mock.Ctx.Response.Body()), `"provider":"removed","provider_name":"removed"`)
			},
		},
		{
			Name:    "ShouldIncludeUnexpiredPending",
			Pending: &session.OpenIDConnectPending{Provider: "example", Issuer: "https://op.example.com", Subject: "xyz789", RemoteUsername: "jane"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinksByUsername(mock.Ctx, "john").
					Return(nil, storage.ErrNoOpenIDConnectLink)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Contains(t, string(mock.Ctx.Response.Body()), `"subject":"xyz789"`)
				assert.Contains(t, string(mock.Ctx.Response.Body()), `"provider_name":"Example"`)
			},
		},
		{
			Name:    "ShouldOmitExpiredPending",
			Pending: &session.OpenIDConnectPending{Provider: "example", Issuer: "https://op.example.com", Subject: "xyz789", Expires: time.Unix(1, 0)},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					LoadOpenIDConnectLinksByUsername(mock.Ctx, "john").
					Return(nil, storage.ErrNoOpenIDConnectLink)
			},
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.NotContains(t, string(mock.Ctx.Response.Body()), `xyz789`)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.OpenIDConnectRelyingParty = newTestRelyingPartyProviders()

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.Username = "john"
			userSession.AuthenticationMethodRefs.UsernameAndPassword = true

			if tc.Pending != nil {
				if tc.Pending.Expires.IsZero() {
					tc.Pending.Expires = mock.Ctx.Providers.Clock.Now().Add(time.Minute * 15)
				}

				userSession.OpenIDConnectPending = tc.Pending
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			if tc.Setup != nil {
				tc.Setup(t, mock)
			}

			UserOpenIDConnectLinksGET(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())

			if tc.Assert != nil {
				tc.Assert(t, mock)
			}
		})
	}
}

func TestUserOpenIDConnectLinkPUT(t *testing.T) {
	testCases := []struct {
		Name     string
		Pending  *session.OpenIDConnectPending
		Setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Expected string
		Assert   func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			Name:    "ShouldSaveTheLinkAndClearPending",
			Pending: &session.OpenIDConnectPending{Provider: "example", Issuer: "https://op.example.com", Subject: "xyz789", RemoteUsername: "jane"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					SaveOpenIDConnectLink(mock.Ctx, gomock.Any()).
					Return(nil)
			},
			Expected: `{"status":"OK"}`,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Nil(t, userSession.OpenIDConnectPending)
			},
		},
		{
			Name:     "ShouldRejectWhenNothingPending",
			Pending:  nil,
			Setup:    nil,
			Expected: `{"status":"KO","message":"There is no external account awaiting a decision."}`,
			Assert:   nil,
		},
		{
			Name:     "ShouldRejectExpiredPending",
			Pending:  &session.OpenIDConnectPending{Provider: "example", Issuer: "https://op.example.com", Subject: "xyz789", Expires: time.Unix(1, 0)},
			Setup:    nil,
			Expected: `{"status":"KO","message":"There is no external account awaiting a decision."}`,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.Nil(t, userSession.OpenIDConnectPending)
			},
		},
		{
			Name:    "ShouldReportConflict",
			Pending: &session.OpenIDConnectPending{Provider: "example", Issuer: "https://op.example.com", Subject: "xyz789"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					SaveOpenIDConnectLink(mock.Ctx, gomock.Any()).
					Return(errors.New("error inserting OpenID Connect 1.0 link for user 'john': UNIQUE constraint failed: user_openid_connect_links.issuer, user_openid_connect_links.subject"))
			},
			Expected: `{"status":"KO","message":"That external account is already linked."}`,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.NotNil(t, userSession.OpenIDConnectPending)
			},
		},
		{
			Name:    "ShouldReportFailureOnOtherStorageError",
			Pending: &session.OpenIDConnectPending{Provider: "example", Issuer: "https://op.example.com", Subject: "xyz789"},
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					SaveOpenIDConnectLink(mock.Ctx, gomock.Any()).
					Return(errors.New("error inserting OpenID Connect 1.0 link for user 'john': disk I/O error"))
			},
			Expected: `{"status":"KO","message":"Unable to link the external account."}`,
			Assert:   nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.OpenIDConnectRelyingParty = newTestRelyingPartyProviders()

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.Username = "john"
			userSession.AuthenticationMethodRefs.UsernameAndPassword = true

			if tc.Pending != nil {
				if tc.Pending.Expires.IsZero() {
					tc.Pending.Expires = mock.Ctx.Providers.Clock.Now().Add(time.Minute * 15)
				}

				userSession.OpenIDConnectPending = tc.Pending
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			if tc.Setup != nil {
				tc.Setup(t, mock)
			}

			UserOpenIDConnectLinkPUT(mock.Ctx)

			assert.Equal(t, tc.Expected, string(mock.Ctx.Response.Body()))

			if tc.Assert != nil {
				tc.Assert(t, mock)
			}
		})
	}
}

func TestUserOpenIDConnectLinkPUTRequiresElevatedSession(t *testing.T) {
	testCases := []struct {
		Name     string
		Elevated bool
		Setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Expected int
		Assert   func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			Name:     "ShouldRejectAuthenticatedNonElevatedSession",
			Elevated: false,
			Setup:    nil,
			Expected: fasthttp.StatusForbidden,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Contains(t, string(mock.Ctx.Response.Body()), `"elevation":true`)

				userSession, err := mock.Ctx.GetSession()
				require.NoError(t, err)

				assert.NotNil(t, userSession.OpenIDConnectPending)
			},
		},
		{
			Name:     "ShouldAllowElevatedSession",
			Elevated: true,
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					SaveOpenIDConnectLink(mock.Ctx, gomock.Any()).
					Return(nil)
			},
			Expected: fasthttp.StatusOK,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, `{"status":"OK"}`, string(mock.Ctx.Response.Body()))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Ctx.Providers.OpenIDConnectRelyingParty = newTestRelyingPartyProviders()
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "127.0.0.1")

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.Username = "john"
			userSession.AuthenticationMethodRefs.UsernameAndPassword = true
			userSession.OpenIDConnectPending = &session.OpenIDConnectPending{
				Provider: "example",
				Issuer:   "https://op.example.com",
				Subject:  "xyz789",
				Expires:  mock.Ctx.Providers.Clock.Now().Add(time.Minute * 15),
			}

			if tc.Elevated {
				userSession.Elevations.User = &session.Elevation{
					ID:       1,
					Expires:  mock.Ctx.Providers.Clock.Now().Add(time.Minute),
					RemoteIP: net.ParseIP("127.0.0.1"),
				}
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			if tc.Setup != nil {
				tc.Setup(t, mock)
			}

			handler := middlewares.RequireElevated(UserOpenIDConnectLinkPUT)

			handler(mock.Ctx)

			assert.Equal(t, tc.Expected, mock.Ctx.Response.StatusCode())

			if tc.Assert != nil {
				tc.Assert(t, mock)
			}
		})
	}
}

func TestUserOpenIDConnectLinkPendingDELETE(t *testing.T) {
	testCases := []struct {
		Name    string
		Pending *session.OpenIDConnectPending
	}{
		{
			Name:    "ShouldClearThePendingProposal",
			Pending: &session.OpenIDConnectPending{Provider: "example", Issuer: "https://op.example.com", Subject: "xyz789"},
		},
		{
			Name:    "ShouldSucceedWhenNothingPending",
			Pending: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.Username = "john"
			userSession.AuthenticationMethodRefs.UsernameAndPassword = true

			if tc.Pending != nil {
				tc.Pending.Expires = mock.Ctx.Providers.Clock.Now().Add(time.Minute * 15)
				userSession.OpenIDConnectPending = tc.Pending
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			UserOpenIDConnectLinkPendingDELETE(mock.Ctx)

			assert.Equal(t, `{"status":"OK"}`, string(mock.Ctx.Response.Body()))

			userSession, err = mock.Ctx.GetSession()
			require.NoError(t, err)

			assert.Nil(t, userSession.OpenIDConnectPending)
		})
	}
}

func TestUserOpenIDConnectLinkDELETE(t *testing.T) {
	testCases := []struct {
		Name     string
		ID       string
		Setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Expected string
	}{
		{
			Name: "ShouldDeleteScopedToTheUser",
			ID:   "1",
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					DeleteOpenIDConnectLink(mock.Ctx, "john", 1).
					Return(nil)
			},
			Expected: `{"status":"OK"}`,
		},
		{
			Name:     "ShouldRejectNonNumericID",
			ID:       "abc",
			Setup:    nil,
			Expected: `{"status":"KO","message":"Unable to remove the external account link."}`,
		},
		{
			Name: "ShouldRejectWhenLinkBelongsToAnotherUser",
			ID:   "2",
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					DeleteOpenIDConnectLink(mock.Ctx, "john", 2).
					Return(storage.ErrNoOpenIDConnectLink)
			},
			Expected: `{"status":"KO","message":"Unable to remove the external account link."}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.Username = "john"
			userSession.AuthenticationMethodRefs.UsernameAndPassword = true

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			mock.Ctx.SetUserValue("linkID", tc.ID)

			if tc.Setup != nil {
				tc.Setup(t, mock)
			}

			UserOpenIDConnectLinkDELETE(mock.Ctx)

			assert.Equal(t, tc.Expected, string(mock.Ctx.Response.Body()))
		})
	}
}

func TestUserOpenIDConnectLinkDELETERequiresElevatedSession(t *testing.T) {
	testCases := []struct {
		Name     string
		Elevated bool
		Setup    func(t *testing.T, mock *mocks.MockAutheliaCtx)
		Expected int
		Assert   func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			Name:     "ShouldRejectAuthenticatedNonElevatedSession",
			Elevated: false,
			Setup:    nil,
			Expected: fasthttp.StatusForbidden,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Contains(t, string(mock.Ctx.Response.Body()), `"elevation":true`)
			},
		},
		{
			Name:     "ShouldAllowElevatedSession",
			Elevated: true,
			Setup: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().
					DeleteOpenIDConnectLink(mock.Ctx, "john", 1).
					Return(nil)
			},
			Expected: fasthttp.StatusOK,
			Assert: func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				assert.Equal(t, `{"status":"OK"}`, string(mock.Ctx.Response.Body()))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Ctx.Request.Header.Set(fasthttp.HeaderXForwardedFor, "127.0.0.1")

			mock.Ctx.SetUserValue("linkID", "1")

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.Username = "john"
			userSession.AuthenticationMethodRefs.UsernameAndPassword = true

			if tc.Elevated {
				userSession.Elevations.User = &session.Elevation{
					ID:       1,
					Expires:  mock.Ctx.Providers.Clock.Now().Add(time.Minute),
					RemoteIP: net.ParseIP("127.0.0.1"),
				}
			}

			require.NoError(t, mock.Ctx.SaveSession(userSession))

			if tc.Setup != nil {
				tc.Setup(t, mock)
			}

			handler := middlewares.RequireElevated(UserOpenIDConnectLinkDELETE)

			handler(mock.Ctx)

			assert.Equal(t, tc.Expected, mock.Ctx.Response.StatusCode())

			if tc.Assert != nil {
				tc.Assert(t, mock)
			}
		})
	}
}
