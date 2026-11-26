// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"
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
			"ShouldHandleBodyParseError",
			"not a valid json",
			`{"status":"OK","data":{"safeTargetURL":false}}`,
			func(t *testing.T, mock *mocks.MockAutheliaCtx) {
				AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred parsing the logout request body", "unable to parse body: invalid character 'o' in literal null (expecting 'u')")
			},
		},
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

			us, err := mock.Ctx.GetSession()

			require.NoError(t, err)

			us.Username = testUsername

			require.NoError(t, mock.Ctx.SaveSession(&us))

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

func TestLogoutPOSTShouldHandleSessionDestroyError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)

	defer mock.Close()

	mock.Ctx.Request.Header.Set("X-Original-URL", "https://auth.notexample.com")
	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred destroying the user session during logout", "unable to destroy user session: unable to retrieve session cookie domain provider: no configured session cookie domain matches the url 'https://auth.notexample.com'")
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
	provider, err := s.mock.Ctx.GetSessionProvider()
	s.Assert().NoError(err)

	userSession, err := provider.Get(s.mock.Ctx)
	s.Assert().NoError(err)

	userSession.Username = testUsername
	s.Assert().NoError(provider.Save(s.mock.Ctx, userSession))
}

func (s *LogoutSuite) TearDownTest() {
	s.mock.Close()
}

func (s *LogoutSuite) TestShouldDestroySession() {
	LogoutPOST(s.mock.Ctx)
	b := s.mock.Ctx.Response.Header.PeekCookie("authelia_session")

	assert.True(s.T(), strings.HasPrefix(string(b), "authelia_session=;"))
}

func TestLogoutPOSTClearConsentSubject(t *testing.T) {
	t.Run("ShouldClearSubjectOfOwnConsentSession", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		subject := uuid.Must(uuid.NewRandom())
		consent := newTestOIDCConsentSession(t, mock, subject)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			ClearOAuth2ConsentSessionSubject(gomock.Any(), consent.ChallengeID, subject).
			Return(true, nil)

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flowID":"%s"}`, consent.ChallengeID))

		LogoutPOST(mock.Ctx)

		assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))

		us, err := mock.Ctx.GetSession()
		require.NoError(t, err)
		assert.True(t, us.IsAnonymous())
	})

	t.Run("ShouldNotClearSubjectWhenConsentSessionHasNone", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Nil)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flowID":"%s"}`, consent.ChallengeID))

		LogoutPOST(mock.Ctx)

		assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))
	})

	t.Run("ShouldNotClearSubjectWhenAnonymous", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtx(t)
		defer mock.Close()

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flowID":"%s"}`, uuid.Must(uuid.NewRandom())))

		LogoutPOST(mock.Ctx)

		assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))
	})

	t.Run("ShouldNotClearSubjectWithMalformedFlowID", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		mock.Ctx.Request.SetBodyString(`{"flowID":"not-a-uuid"}`)

		LogoutPOST(mock.Ctx)

		assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))
	})

	t.Run("ShouldContinueLogoutWhenClearFails", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		subject := uuid.Must(uuid.NewRandom())
		consent := newTestOIDCConsentSession(t, mock, subject)

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.StorageMock.EXPECT().
			ClearOAuth2ConsentSessionSubject(gomock.Any(), consent.ChallengeID, subject).
			Return(false, fmt.Errorf("error in db"))

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"flowID":"%s"}`, consent.ChallengeID))

		LogoutPOST(mock.Ctx)

		assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))

		AssertLogEntryMessageAndError(t, MustGetLogLastSeq(t, mock.Hook, 0), "Error occurred clearing the consent session subject during logout", "error in db")
	})
}
