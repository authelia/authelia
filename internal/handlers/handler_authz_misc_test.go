package handlers

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestAuthzImplementation(t *testing.T) {
	assert.Equal(t, "Legacy", AuthzImplLegacy.String())
	assert.Equal(t, "", AuthzImplementation(-1).String())
}

func TestFriendlyMethod(t *testing.T) {
	assert.Equal(t, "unknown", friendlyMethod(""))
	assert.Equal(t, "GET", friendlyMethod(fasthttp.MethodGet))
}

func TestCookieSessionAuthnStrategyFlags(t *testing.T) {
	strategy := NewCookieSessionAuthnStrategy(schema.NewRefreshIntervalDurationAlways())

	assert.False(t, strategy.CanHandleUnauthorized())
	assert.False(t, strategy.HeaderStrategy())

	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	strategy.HandleUnauthorized(mock.Ctx, &Authn{}, nil)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, []byte(nil), mock.Ctx.Response.Header.Peek(fasthttp.HeaderWWWAuthenticate))
}

func TestHandleVerifyGETAuthorizationBearerResolveUser(t *testing.T) {
	testCases := []struct {
		Name          string
		Username      string
		ClientID      string
		CCS           bool
		Level         authentication.Level
		Setup         func(mock *mocks.MockAutheliaCtx)
		ExpectDetails *authentication.UserDetails
		ExpectError   string
	}{
		{
			Name:     "ShouldReturnClientIDForClientCredentialsWithoutCallingGetDetails",
			Username: "",
			ClientID: "client-abc",
			CCS:      true,
			Level:    authentication.OneFactor,
			Setup: func(mock *mocks.MockAutheliaCtx) {
				mock.UserProviderMock.EXPECT().GetDetails(gomock.Any()).Times(0)
			},
			ExpectDetails: nil,
		},
		{
			Name:     "ShouldResolveDetailsForUserBoundToken",
			Username: "john",
			ClientID: "",
			CCS:      false,
			Level:    authentication.OneFactor,
			Setup: func(mock *mocks.MockAutheliaCtx) {
				mock.UserProviderMock.EXPECT().
					GetDetails(gomock.Eq("john")).
					Return(&authentication.UserDetails{Username: "john"}, nil)
			},
			ExpectDetails: &authentication.UserDetails{Username: "john"},
		},
		{
			Name:     "ShouldReturnErrorWhenGetDetailsFails",
			Username: "ghost",
			ClientID: "",
			CCS:      false,
			Level:    authentication.OneFactor,
			Setup: func(mock *mocks.MockAutheliaCtx) {
				mock.UserProviderMock.EXPECT().
					GetDetails(gomock.Eq("ghost")).
					Return(nil, fmt.Errorf("boom"))
			},
			ExpectError: "failed to retrieve user details for user ghost: boom",
		},
		{
			Name:     "ShouldReturnErrorWhenUserNotFound",
			Username: "missing",
			ClientID: "",
			CCS:      false,
			Level:    authentication.OneFactor,
			Setup: func(mock *mocks.MockAutheliaCtx) {
				mock.UserProviderMock.EXPECT().
					GetDetails(gomock.Eq("missing")).
					Return(nil, authentication.ErrUserNotFound)
			},
			ExpectError: "failed to retrieve user details for user missing: user not found",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			if tc.Setup != nil {
				tc.Setup(mock)
			}

			details, clientID, ccs, level, err := handleVerifyGETAuthorizationBearerResolveUser(mock.Ctx, tc.Username, tc.ClientID, tc.CCS, tc.Level)

			if tc.ExpectError != "" {
				require.EqualError(t, err, tc.ExpectError)
				assert.Nil(t, details)
				assert.Empty(t, clientID)
				assert.False(t, ccs)
				assert.Equal(t, authentication.NotAuthenticated, level)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.ExpectDetails, details)
			assert.Equal(t, tc.ClientID, clientID)
			assert.Equal(t, tc.CCS, ccs)
			assert.Equal(t, tc.Level, level)
		})
	}
}

func TestGenerateVerifySessionHasUpToDateProfileTraceLogs(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	generateVerifySessionHasUpToDateProfileTraceLogs(mock.Ctx, &session.UserSession{Username: "john", DisplayName: "example", Groups: []string{"abc"}, Emails: []string{"user@example.com", "test@example.com"}}, &authentication.UserDetails{Username: "john", Groups: []string{"123"}, DisplayName: "notexample", Emails: []string{"notuser@example.com"}})
	generateVerifySessionHasUpToDateProfileTraceLogs(mock.Ctx, &session.UserSession{Username: "john", DisplayName: "example"}, &authentication.UserDetails{Username: "john", DisplayName: "example"})
	generateVerifySessionHasUpToDateProfileTraceLogs(mock.Ctx, &session.UserSession{Username: "john", DisplayName: "example", Emails: []string{"abc@example.com"}}, &authentication.UserDetails{Username: "john", DisplayName: "example"})
	generateVerifySessionHasUpToDateProfileTraceLogs(mock.Ctx, &session.UserSession{Username: "john", DisplayName: "example"}, &authentication.UserDetails{Username: "john", DisplayName: "example", Emails: []string{"abc@example.com"}})
}
