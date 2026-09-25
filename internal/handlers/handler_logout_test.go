// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestLogoutPOST(t *testing.T) {
	testCases := []struct {
		name      string
		have      string
		expected  string
		expectedf func(t *testing.T, mock *mocks.MockAutheliaCtx)
	}{
		{
			"ShouldHandleNoTargetURL",
			`{}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
			nil,
		},
		{
			"ShouldHandleSafeTargetURL",
			`{"targetURL":"https://www.example.com"}`,
			`{"status":"OK","data":{"safeTargetURL":true}}`,
			nil,
		},
		{
			"ShouldHandleUnsafeTargetURL",
			`{"targetURL":"https://www.notexample.com"}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
			nil,
		},
		{
			"ShouldHandleMalformedTargetURL",
			`{"targetURL":"https//www.example.com"}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
			nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Configuration.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{}

			us, err := mock.Ctx.GetSession()

			require.NoError(t, err)

			us.Username = testUsername

			require.NoError(t, mock.Ctx.SaveSession(&us))

			provider, err := mock.Ctx.GetSessionProvider()
			require.NoError(t, err)
			require.NotEmpty(t, us.PublicID)

			mock.StorageMock.EXPECT().DeleteOAuth2SessionIDByPublicID(gomock.Any(), provider.GetIssuer(), us.PublicID).Return(nil)

			mock.Ctx.Request.SetBodyString(tc.have)

			LogoutPOST(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))

			assert.True(t, strings.HasPrefix(string(mock.Ctx.Response.Header.PeekCookie("authelia_session")), "authelia_session=;"))

			if tc.expectedf != nil {
				tc.expectedf(t, mock)
			}
		})
	}
}

func TestLogoutGET(t *testing.T) {
	testCases := []struct {
		name     string
		logout   func(now time.Time) *session.OpenIDConnectLogout
		flowID   string
		expected string
	}{
		{
			"ShouldReportNothingPending",
			func(now time.Time) *session.OpenIDConnectLogout { return nil },
			testLogoutFlowID,
			`{"status":"OK","data":{"pending":false}}`,
		},
		{
			"ShouldReportPendingWithClient",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			testLogoutFlowID,
			`{"status":"OK","data":{"pending":true,"clientID":"end-session-client","clientName":"End Session Client"}}`,
		},
		{
			"ShouldReportPendingWithoutClient",
			testLogoutPendingFn(session.OpenIDConnectLogout{}, time.Minute),
			testLogoutFlowID,
			`{"status":"OK","data":{"pending":true}}`,
		},
		{
			"ShouldNotReportExpired",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID}, -time.Second),
			testLogoutFlowID,
			`{"status":"OK","data":{"pending":false}}`,
		},
		{
			"ShouldNotReportWithoutFlowID",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID}, time.Minute),
			"",
			`{"status":"OK","data":{"pending":false}}`,
		},
		{
			"ShouldNotReportWithMismatchedFlowID",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID}, time.Minute),
			"0b1d2c3e-4f5a-4b6c-8d7e-9f0a1b2c3d4e",
			`{"status":"OK","data":{"pending":false}}`,
		},
		{
			"ShouldNotReportStoredLogoutWithoutFlowID",
			func(now time.Time) *session.OpenIDConnectLogout {
				return &session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, Expires: now.Add(time.Minute)}
			},
			"",
			`{"status":"OK","data":{"pending":false}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			config := newTestOIDCEndSessionConfig(t)
			config.Clients[0].Name = "End Session Client"

			setupTestOIDCProvider(t, mock, config)

			setTestLogoutPending(t, mock, tc.logout(mock.Ctx.GetClock().Now()))

			if tc.flowID != "" {
				mock.Ctx.Request.URI().QueryArgs().Set(queryArgFlowID, tc.flowID)
			}

			LogoutGET(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))
		})
	}
}

func TestLogoutPOSTWithPendingLogout(t *testing.T) {
	withFlow := `{"flowID":"` + testLogoutFlowID + `"}`

	testCases := []struct {
		name     string
		logout   func(now time.Time) *session.OpenIDConnectLogout
		body     string
		expected string
	}{
		{
			"ShouldRedirectToTheStoredURIWithState",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout, State: "state abc"}, time.Minute),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"https://app.example.com/logged-out?state=state+abc"}}`,
		},
		{
			"ShouldAppendStateToTheStoredURIsExistingQuery",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: "https://app.example.com/logged-out?b=2&a=1", State: "abc"}, time.Minute),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"https://app.example.com/logged-out?b=2\u0026a=1\u0026state=abc"}}`,
		},
		{
			"ShouldKeepTheStoredURIExactlyAsRegistered",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: "HTTPS://App.example.com/logged-out?", State: "abc"}, time.Minute),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"HTTPS://App.example.com/logged-out?\u0026state=abc"}}`,
		},
		{
			"ShouldRedirectToTheStoredURIWithoutState",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"https://app.example.com/logged-out"}}`,
		},
		{
			"ShouldNotRedirectWithoutAStoredURI",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID}, time.Minute),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
		{
			"ShouldNotRedirectWhenExpired",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, -time.Second),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
		{
			"ShouldNotRedirectAnOrdinaryLogoutWithoutFlowID",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			`{}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
		{
			"ShouldNotRedirectWithMismatchedFlowID",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			`{"flowID":"0b1d2c3e-4f5a-4b6c-8d7e-9f0a1b2c3d4e"}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
		{
			"ShouldNotRedirectStoredLogoutWithoutFlowID",
			func(now time.Time) *session.OpenIDConnectLogout {
				return &session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout, Expires: now.Add(time.Minute)}
			},
			`{"flowID":""}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
		{
			"ShouldNotAcceptAnOffDomainTargetURLEvenIfRegistered",
			func(now time.Time) *session.OpenIDConnectLogout { return nil },
			`{"targetURL":"https://app.example.net/logged-out"}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			config := newTestOIDCEndSessionConfig(t)
			config.Clients[0].PostLogoutRedirectURIs = append(config.Clients[0].PostLogoutRedirectURIs, "https://app.example.net/logged-out")

			setupTestOIDCProvider(t, mock, config)

			us := setTestLogoutPending(t, mock, tc.logout(mock.Ctx.GetClock().Now()))

			provider, err := mock.Ctx.GetSessionProvider()
			require.NoError(t, err)

			mock.StorageMock.EXPECT().DeleteOAuth2SessionIDByPublicID(gomock.Any(), provider.GetIssuer(), us.PublicID).Return(nil)

			mock.Ctx.Request.SetBodyString(tc.body)

			LogoutPOST(mock.Ctx)

			assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
			assert.Equal(t, tc.expected, string(mock.Ctx.Response.Body()))
			assert.True(t, strings.HasPrefix(string(mock.Ctx.Response.Header.PeekCookie("authelia_session")), "authelia_session=;"), "the session is always destroyed")
		})
	}
}

func TestLogoutDELETE(t *testing.T) {
	testCases := []struct {
		name      string
		logout    func(now time.Time) *session.OpenIDConnectLogout
		flowID    string
		cancelled bool
	}{
		{
			"ShouldCancelThePendingLogoutAndKeepTheSession",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			testLogoutFlowID,
			true,
		},
		{
			"ShouldNotCancelWithoutFlowID",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			"",
			false,
		},
		{
			"ShouldNotCancelWithMismatchedFlowID",
			testLogoutPendingFn(session.OpenIDConnectLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			"0b1d2c3e-4f5a-4b6c-8d7e-9f0a1b2c3d4e",
			false,
		},
		{
			"ShouldSucceedWithNothingPending",
			func(now time.Time) *session.OpenIDConnectLogout { return nil },
			testLogoutFlowID,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			have := tc.logout(mock.Ctx.GetClock().Now())

			setTestLogoutPending(t, mock, have)

			if tc.flowID != "" {
				mock.Ctx.Request.URI().QueryArgs().Set(queryArgFlowID, tc.flowID)
			}

			LogoutDELETE(mock.Ctx)

			mock.Assert200OK(t, nil)

			us, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			assert.Equal(t, testUsername, us.Username, "the user must remain logged in")

			if tc.cancelled {
				assert.Nil(t, us.OpenIDConnectLogout)
			} else {
				require.NotNil(t, us.OpenIDConnectLogout, "a request belonging to another flow must not be cancelled")
				assert.Equal(t, have.FlowID, us.OpenIDConnectLogout.FlowID)
			}
		})
	}
}

func TestLogoutPOSTShouldLogAndContinueWhenSessionIDDeleteFails(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{}

	us, err := mock.Ctx.GetSession()

	require.NoError(t, err)

	us.Username = testUsername

	require.NoError(t, mock.Ctx.SaveSession(&us))

	provider, err := mock.Ctx.GetSessionProvider()
	require.NoError(t, err)
	require.NotEmpty(t, us.PublicID)

	mock.StorageMock.EXPECT().DeleteOAuth2SessionIDByPublicID(gomock.Any(), provider.GetIssuer(), us.PublicID).Return(errors.New("connection refused"))

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred removing the OpenID Connect session identifiers during logout", "connection refused")
}

func TestLogoutPOSTShouldNotRemoveSessionIDsWithoutOpenIDConnect(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.IdentityProviders.OIDC = nil

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	us.Username = testUsername

	require.NoError(t, mock.Ctx.SaveSession(&us))
	require.NotEmpty(t, us.PublicID)

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))
	assert.True(t, strings.HasPrefix(string(mock.Ctx.Response.Header.PeekCookie("authelia_session")), "authelia_session=;"))
}

func TestLogoutPOSTShouldNotLogoutOnBodyParseError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{}

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	us.Username = testUsername

	require.NoError(t, mock.Ctx.SaveSession(&us))

	mock.Ctx.Request.SetBodyString("not a valid json")

	LogoutPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"KO","message":"Operation failed."}`, string(mock.Ctx.Response.Body()))
	assert.False(t, strings.HasPrefix(string(mock.Ctx.Response.Header.PeekCookie("authelia_session")), "authelia_session=;"), "the session cookie must not be cleared")

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing the logout request body", "unable to parse body: invalid character 'o' in literal null (expecting 'u')")
}

func TestLogoutPOSTShouldHandleSessionDestroyError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred destroying the user session during logout", "unable to destroy user session: unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")

	mock.Assert200KO(t, "Operation failed.")
}

func TestLogoutPOSTShouldNotRemoveSessionIDsWhenTheSessionSurvives(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{}

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	mock.Assert200KO(t, "Operation failed.")
}

func TestRunLogoutSuite(t *testing.T) {
	s := new(LogoutSuite)
	suite.Run(t, s)
}

type LogoutSuite struct {
	suite.Suite

	mock *mocks.MockAutheliaCtx
}

func (s *LogoutSuite) SetupTest() {
	s.mock = mocks.NewMockAutheliaCtx(s.T())
	s.mock.Ctx.Configuration.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{}

	provider, err := s.mock.Ctx.GetSessionProvider()
	s.Assert().NoError(err)

	userSession, err := provider.Get(s.mock.Ctx)
	s.Assert().NoError(err)

	userSession.Username = testUsername
	s.Assert().NoError(provider.Save(s.mock.Ctx, userSession))
	s.Require().NotEmpty(userSession.PublicID)

	s.mock.StorageMock.EXPECT().DeleteOAuth2SessionIDByPublicID(gomock.Any(), provider.GetIssuer(), userSession.PublicID).Return(nil)

	s.mock.Ctx.Request.SetBodyString(`{}`)
}

func (s *LogoutSuite) TearDownTest() {
	s.mock.Close()
}

func (s *LogoutSuite) TestShouldDestroySession() {
	LogoutPOST(s.mock.Ctx)
	b := s.mock.Ctx.Response.Header.PeekCookie("authelia_session")

	assert.True(s.T(), strings.HasPrefix(string(b), "authelia_session=;"))
}

const testLogoutFlowID = "8c6f6e2a-5b3e-4f0a-9d1c-2e7b3a4f5c6d"

func setTestLogoutPending(t *testing.T, mock *mocks.MockAutheliaCtx, logout *session.OpenIDConnectLogout) session.UserSession {
	t.Helper()

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	us.Username = testUsername
	us.OpenIDConnectLogout = logout

	require.NoError(t, mock.Ctx.SaveSession(&us))

	return us
}

func testLogoutPendingFn(logout session.OpenIDConnectLogout, lifespan time.Duration) func(now time.Time) *session.OpenIDConnectLogout {
	return func(now time.Time) *session.OpenIDConnectLogout {
		if logout.FlowID == "" {
			logout.FlowID = testLogoutFlowID
		}

		logout.Expires = now.Add(lifespan)

		return &logout
	}
}
