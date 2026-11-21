// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
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

			us, issuer := mustSaveTestUserSession(t, mock, testUsername)

			mock.StorageMock.EXPECT().EndOAuth2Session(gomock.Any(), issuer, us.PublicID, gomock.Any()).Return(nil, nil)

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
		logout   func(now time.Time) *model.OAuth2SessionLogout
		flowID   string
		expected string
	}{
		{
			"ShouldReportNothingPending",
			func(now time.Time) *model.OAuth2SessionLogout { return nil },
			testLogoutFlowID,
			`{"status":"OK","data":{"pending":false}}`,
		},
		{
			"ShouldReportPendingWithClient",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, Subject: testLogoutSubject, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			testLogoutFlowID,
			`{"status":"OK","data":{"pending":true,"clientID":"end-session-client","clientName":"End Session Client"}}`,
		},
		{
			"ShouldReportPendingWithoutClientWhenNoTrustedHint",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			testLogoutFlowID,
			`{"status":"OK","data":{"pending":true}}`,
		},
		{
			"ShouldReportPendingWithoutClient",
			testLogoutPendingFn(model.OAuth2SessionLogout{}, time.Minute),
			testLogoutFlowID,
			`{"status":"OK","data":{"pending":true}}`,
		},
		{
			"ShouldNotReportExpired",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID}, -time.Second),
			testLogoutFlowID,
			`{"status":"OK","data":{"pending":false}}`,
		},
		{
			"ShouldNotReportWithoutFlowID",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID}, time.Minute),
			"",
			`{"status":"OK","data":{"pending":false}}`,
		},
		{
			"ShouldNotReportWithMismatchedFlowID",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID}, time.Minute),
			"0b1d2c3e-4f5a-4b6c-8d7e-9f0a1b2c3d4e",
			`{"status":"OK","data":{"pending":false}}`,
		},
		{
			"ShouldNotReportStoredLogoutWithoutFlowID",
			func(now time.Time) *model.OAuth2SessionLogout {
				return &model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, ExpiresAt: now.Add(time.Minute)}
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

			mock.StorageMock.EXPECT().
				LoadUserOpaqueIdentifier(gomock.Any(), uuid.MustParse(testLogoutSubject)).
				AnyTimes().
				Return(&model.UserOpaqueIdentifier{Username: testUsername}, nil)

			_, _ = setTestLogoutPending(t, mock, tc.logout(mock.Ctx.GetClock().Now()))

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
		logout   func(now time.Time) *model.OAuth2SessionLogout
		body     string
		expected string
	}{
		{
			"ShouldRedirectToTheStoredURIWithState",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout, State: "state abc"}, time.Minute),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"https://app.example.com/logged-out?state=state+abc"}}`,
		},
		{
			"ShouldAppendStateToTheStoredURIsExistingQuery",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: "https://app.example.com/logged-out?b=2&a=1", State: "abc"}, time.Minute),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"https://app.example.com/logged-out?b=2\u0026a=1\u0026state=abc"}}`,
		},
		{
			"ShouldKeepTheStoredURIExactlyAsRegistered",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: "HTTPS://App.example.com/logged-out?", State: "abc"}, time.Minute),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"HTTPS://App.example.com/logged-out?\u0026state=abc"}}`,
		},
		{
			"ShouldRedirectToTheStoredURIWithoutState",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"https://app.example.com/logged-out"}}`,
		},
		{
			"ShouldNotRedirectWithoutAStoredURI",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID}, time.Minute),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
		{
			"ShouldNotRedirectWhenExpired",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, -time.Second),
			withFlow,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
		{
			"ShouldNotRedirectAnOrdinaryLogoutWithoutFlowID",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			`{}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
		{
			"ShouldNotRedirectWithMismatchedFlowID",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			`{"flowID":"0b1d2c3e-4f5a-4b6c-8d7e-9f0a1b2c3d4e"}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
		{
			"ShouldNotRedirectStoredLogoutWithoutFlowID",
			func(now time.Time) *model.OAuth2SessionLogout {
				return &model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout, ExpiresAt: now.Add(time.Minute)}
			},
			`{"flowID":""}`,
			`{"status":"OK","data":{"safeTargetURL":false}}`,
		},
		{
			"ShouldNotAcceptAnOffDomainTargetURLEvenIfRegistered",
			func(now time.Time) *model.OAuth2SessionLogout { return nil },
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

			us, _ := setTestLogoutPending(t, mock, tc.logout(mock.Ctx.GetClock().Now()))

			provider, err := mock.Ctx.GetSessionProvider()
			require.NoError(t, err)

			mock.StorageMock.EXPECT().EndOAuth2Session(gomock.Any(), provider.GetIssuer(), us.PublicID, gomock.Any()).Return(nil, nil)

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
		logout    func(now time.Time) *model.OAuth2SessionLogout
		flowID    string
		cancelled bool
	}{
		{
			"ShouldCancelThePendingLogoutAndKeepTheSession",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			testLogoutFlowID,
			true,
		},
		{
			"ShouldNotCancelWithoutFlowID",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			"",
			false,
		},
		{
			"ShouldNotCancelWithMismatchedFlowID",
			testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute),
			"0b1d2c3e-4f5a-4b6c-8d7e-9f0a1b2c3d4e",
			false,
		},
		{
			"ShouldSucceedWithNothingPending",
			func(now time.Time) *model.OAuth2SessionLogout { return nil },
			testLogoutFlowID,
			true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			_, store := setTestLogoutPending(t, mock, tc.logout(mock.Ctx.GetClock().Now()))

			if tc.flowID != "" {
				mock.Ctx.Request.URI().QueryArgs().Set(queryArgFlowID, tc.flowID)
			}

			LogoutDELETE(mock.Ctx)

			mock.Assert200OK(t, nil)

			us, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			assert.Equal(t, testUsername, us.Username, "the user must remain logged in")

			assert.Equal(t, tc.cancelled, slices.Contains(store.deleted, uuid.MustParse(testLogoutFlowID)))
		})
	}
}

func TestLogoutPOSTShouldLogAndContinueWhenEndingTheSessionFails(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{}

	us, issuer := mustSaveTestUserSession(t, mock, testUsername)

	mock.StorageMock.EXPECT().EndOAuth2Session(gomock.Any(), issuer, us.PublicID, gomock.Any()).Times(3).Return(nil, errors.New("connection refused"))

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred ending the OpenID Connect sessions during logout: they'll be ended once it's determined the session no longer exists", "connection refused")
}

func TestLogoutPOSTShouldNotRemoveSessionIDsWithoutOpenIDConnect(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.IdentityProviders.OIDC = nil

	us, _ := mustSaveTestUserSession(t, mock, testUsername)
	require.NotEmpty(t, us.PublicID)

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))
	assert.True(t, strings.HasPrefix(string(mock.Ctx.Response.Header.PeekCookie("authelia_session")), "authelia_session=;"))
}

func TestLogoutPOSTShouldLogoutOnBodyParseError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{}

	us, issuer := mustSaveTestUserSession(t, mock, testUsername)

	mock.StorageMock.EXPECT().EndOAuth2Session(gomock.Any(), issuer, us.PublicID, gomock.Any()).Return(nil, nil)

	mock.Ctx.Request.SetBodyString("not a valid json")

	LogoutPOST(mock.Ctx)

	assert.Equal(t, fasthttp.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))
	assert.True(t, strings.HasPrefix(string(mock.Ctx.Response.Header.PeekCookie("authelia_session")), "authelia_session=;"))

	AssertLogEntryMessageAndError(t, mock.Hook.AllEntries()[0], "Error occurred parsing the logout request body", "unable to parse body: invalid character 'o' in literal null (expecting 'u')")
}

func TestLogoutPOSTShouldHandleSessionDestroyError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred destroying the user session during logout", "unable to destroy user session: unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")

	mock.Assert200KO(t, messageOperationFailed)
}

func TestLogoutPOSTShouldNotRemoveSessionIDsWhenTheSessionSurvives(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Configuration.IdentityProviders.OIDC = &schema.IdentityProvidersOpenIDConnect{}

	repository := setupTestFailingSessionRepository(t, mock)

	us, _ := mustSaveTestUserSession(t, mock, testUsername)
	require.NotEmpty(t, us.PublicID)

	repository.errDelete = errTestSessionBackend

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	mock.Assert200KO(t, messageOperationFailed)

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred destroying the user session during logout", "error occurred deleting session from backend: "+errTestSessionBackend.Error())

	us, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	assert.Equal(t, testUsername, us.Username)
}

func TestOpenIDConnectStateDoesNotSurviveLogout(t *testing.T) {
	testCases := []struct {
		Name string
	}{
		{Name: "ShouldClearFlowAndPendingOnLogout"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.Username = "john"
			userSession.ExternalIdentity = &session.ExternalIdentityFlow{Provider: "example"}
			userSession.ExternalIdentityPending = &session.ExternalIdentityPending{Provider: "example"}

			require.NoError(t, mock.Ctx.SaveSession(&userSession))

			userSession, err = mock.Ctx.GetSession()
			require.NoError(t, err)
			require.NotNil(t, userSession.ExternalIdentity)
			require.NotNil(t, userSession.ExternalIdentityPending)
			assert.Equal(t, "example", userSession.ExternalIdentity.Provider)
			assert.Equal(t, "example", userSession.ExternalIdentityPending.Provider)

			LogoutPOST(mock.Ctx)

			userSession, err = mock.Ctx.GetSession()
			require.NoError(t, err)

			assert.Nil(t, userSession.ExternalIdentity)
			assert.Nil(t, userSession.ExternalIdentityPending)
		})
	}
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

	s.mock.StorageMock.EXPECT().EndOAuth2Session(gomock.Any(), provider.GetIssuer(), userSession.PublicID, gomock.Any()).Return(nil, nil)

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

func TestLogoutPOSTShouldRetainThePendingLogoutWhenTheSessionSurvives(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

	repository := setupTestFailingSessionRepository(t, mock)

	us, store := setTestLogoutPending(t, mock, testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute)(mock.Ctx.GetClock().Now()))

	repository.errDelete = errTestSessionBackend

	mock.Ctx.Request.SetBodyString(`{"flowID":"` + testLogoutFlowID + `"}`)

	LogoutPOST(mock.Ctx)

	mock.Assert200KO(t, messageOperationFailed)
	assert.Empty(t, store.deleted)

	repository.errDelete = nil

	provider, err := mock.Ctx.GetSessionProvider()
	require.NoError(t, err)

	mock.StorageMock.EXPECT().EndOAuth2Session(gomock.Any(), provider.GetIssuer(), us.PublicID, gomock.Any()).Return(nil, nil)

	mock.Ctx.Response.Reset()
	mock.Ctx.Request.SetBodyString(`{"flowID":"` + testLogoutFlowID + `"}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"https://app.example.com/logged-out"}}`, string(mock.Ctx.Response.Body()))
	assert.Equal(t, []uuid.UUID{uuid.MustParse(testLogoutFlowID)}, store.deleted)
}

func TestLogoutPOSTShouldConsumeThePendingLogout(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	setupTestOIDCProvider(t, mock, newTestOIDCEndSessionConfig(t))

	us, store := setTestLogoutPending(t, mock, testLogoutPendingFn(model.OAuth2SessionLogout{ClientID: testOIDCEndSessionClientID, RedirectURI: testOIDCEndSessionPostLogout}, time.Minute)(mock.Ctx.GetClock().Now()))

	provider, err := mock.Ctx.GetSessionProvider()
	require.NoError(t, err)

	mock.StorageMock.EXPECT().EndOAuth2Session(gomock.Any(), provider.GetIssuer(), us.PublicID, gomock.Any()).Return(nil, nil)

	mock.Ctx.Request.SetBodyString(`{"flowID":"` + testLogoutFlowID + `"}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"https://app.example.com/logged-out"}}`, string(mock.Ctx.Response.Body()))
	assert.Equal(t, []uuid.UUID{uuid.MustParse(testLogoutFlowID)}, store.deleted)

	mock.Ctx.Response.Reset()
	mock.Ctx.Request.SetBodyString(`{"flowID":"` + testLogoutFlowID + `"}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))
}

const (
	testLogoutFlowID  = "8c6f6e2a-5b3e-4f0a-9d1c-2e7b3a4f5c6d"
	testLogoutSubject = "3f1c2b7e-8a4d-4c9e-b5f6-1d2e3f4a5b6c"
)

type testLogoutStore struct {
	deleted []uuid.UUID
}

func setTestLogoutPending(t *testing.T, mock *mocks.MockAutheliaCtx, logout *model.OAuth2SessionLogout) (us session.UserSession, store *testLogoutStore) {
	t.Helper()

	us, _ = mustSaveTestUserSession(t, mock, testUsername)

	store = &testLogoutStore{}

	mock.StorageMock.EXPECT().
		LoadOAuth2SessionLogout(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, flowID uuid.UUID) (*model.OAuth2SessionLogout, error) {
			if logout == nil || logout.FlowID != flowID || slices.Contains(store.deleted, flowID) {
				return nil, nil
			}

			value := *logout

			return &value, nil
		})

	mock.StorageMock.EXPECT().
		DeleteOAuth2SessionLogout(gomock.Any(), gomock.Any()).
		AnyTimes().
		DoAndReturn(func(_ context.Context, flowID uuid.UUID) (bool, error) {
			removed := logout != nil && logout.FlowID == flowID && !slices.Contains(store.deleted, flowID)

			store.deleted = append(store.deleted, flowID)

			return removed, nil
		})

	return us, store
}

func testLogoutPendingFn(logout model.OAuth2SessionLogout, lifespan time.Duration) func(now time.Time) *model.OAuth2SessionLogout {
	return func(now time.Time) *model.OAuth2SessionLogout {
		if logout.FlowID == uuid.Nil {
			logout.FlowID = uuid.MustParse(testLogoutFlowID)
		}

		logout.ExpiresAt = now.Add(lifespan)

		return &logout
	}
}
