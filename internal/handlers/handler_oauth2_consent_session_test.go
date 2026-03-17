// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestOAuth2ConsentShouldHandleUserDetailsErrors(t *testing.T) {
	setupDevice := func(t *testing.T) (mock *mocks.MockAutheliaCtx, userCode string) {
		t.Helper()

		mock = mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))

		mock.UserProviderMock.EXPECT().
			GetDetails(testUsername).
			AnyTimes().
			Return(nil, fmt.Errorf("failed to lookup user"))

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCDeviceCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)
		setupTestOIDCDeviceCodeStore(t, mock)

		userCode = mustGetTestOIDCUserCode(t, mock)

		mock.Ctx.Response.Reset()

		return mock, userCode
	}

	t.Run("ShouldHandleAuthorizationFlow", func(t *testing.T) {
		mock := mocks.NewMockAutheliaCtxWithUserSession(t, newTestOIDCUserSession(1))
		defer mock.Close()

		mock.UserProviderMock.EXPECT().
			GetDetails(testUsername).
			Return(nil, fmt.Errorf("failed to lookup user"))

		config := newTestOIDCConfig(t)
		config.Clients = []schema.IdentityProvidersOpenIDConnectClient{newTestOIDCAuthorizationCodeClient(t)}

		setupTestOIDCProvider(t, mock, config)

		consent := newTestOIDCConsentSession(t, mock, uuid.Must(uuid.NewRandom()))

		mock.StorageMock.EXPECT().
			LoadOAuth2ConsentSessionByChallengeID(gomock.Any(), consent.ChallengeID).
			Return(consent, nil)

		mock.Ctx.Request.SetRequestURI(fmt.Sprintf("/api/oidc/consent?flow_id=%s", consent.ChallengeID))

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred performing consent during the Consent Flow stage of the Authorization Flow as an error occurred retrieving user details", "failed to lookup user")
	})

	t.Run("ShouldHandleDeviceAuthorizationFlowGET", func(t *testing.T) {
		mock, userCode := setupDevice(t)
		defer mock.Close()

		mock.Ctx.Request.SetRequestURI("/api/oidc/consent?user_code=" + userCode)

		OAuth2ConsentGET(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Device Authorization Flow failed to retrieve user details", "failed to lookup user")
	})

	t.Run("ShouldHandleDeviceAuthorizationFlowPOST", func(t *testing.T) {
		mock, userCode := setupDevice(t)
		defer mock.Close()

		mock.Ctx.Request.SetBodyString(fmt.Sprintf(`{"subflow":"device_authorization","client_id":"%s","consent":true,"user_code":"%s"}`, testOIDCDeviceCodeID, userCode))

		OAuth2ConsentPOST(mock.Ctx)

		mock.Assert200KO(t, messageOperationFailed)

		AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred fetching user details during the Consent Flow stage of the Device Authorization Flow", "failed to lookup user")
	})
}

func TestHandleOAuth2AuthorizationConsentModeExplicitWithIDShouldRedirect(t *testing.T) {
	clientTest := &oidc.RegisteredClient{
		ID:            testValue,
		ConsentPolicy: oidc.ClientConsentPolicy{Mode: oidc.ClientConsentModeExplicit},
	}

	challenge := uuid.MustParse("11303e1f-f8af-436a-9a72-c7361bfc9f37")
	sub := uuid.MustParse("e79b6494-8852-4439-860c-159f2cba83dc")

	testCases := []struct {
		name    string
		consent func(mock *mocks.MockAutheliaCtx) *model.OAuth2ConsentSession
	}{
		{
			"ShouldRedirectWhenConsentSessionHasNoSubject",
			func(mock *mocks.MockAutheliaCtx) *model.OAuth2ConsentSession {
				return &model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: testValue}
			},
		},
		{
			"ShouldRedirectWhenConsentSessionIsNotAuthorized",
			func(mock *mocks.MockAutheliaCtx) *model.OAuth2ConsentSession {
				return &model.OAuth2ConsentSession{ID: 44, ChallengeID: challenge, ClientID: testValue, Subject: uuid.NullUUID{UUID: sub, Valid: true}, ExpiresAt: mock.Ctx.Providers.Clock.Now().Add(time.Second * 10)}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Clock.Set(time.Unix(1000000, 0))

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID: "abc",
							},
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, mock.StorageMock, mock.Ctx.Providers.Templates)

			mock.StorageMock.EXPECT().
				LoadOAuth2ConsentSessionByChallengeID(gomock.Eq(mock.Ctx), gomock.Eq(challenge)).
				Return(tc.consent(mock), nil)

			issuer, err := url.Parse("https://auth.example.com")

			require.NoError(t, err)

			requester := &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client:      clientTest,
					RequestedAt: time.Unix(1000000, 0),
				},
			}

			rw := httptest.NewRecorder()

			userSession := session.UserSession{Username: testValue, FirstFactorAuthnTimestamp: 1000000, SecondFactorAuthnTimestamp: 1000000}

			consent, handled := handleOAuth2AuthorizationConsentModeExplicitWithID(mock.Ctx, issuer, clientTest, userSession, &authentication.UserDetailsExtended{UserDetails: &authentication.UserDetails{}}, sub, challenge, rw, httptest.NewRequest("GET", "https://example.com", nil), requester)

			assert.True(t, handled)
			assert.Nil(t, consent)

			require.Equal(t, http.StatusSeeOther, rw.Code)

			location, err := url.Parse(rw.Header().Get("Location"))

			require.NoError(t, err)

			assert.Equal(t, challenge.String(), location.Query().Get(queryArgFlowID))
		})
	}
}

func TestHandleOAuth2AuthorizationConsentSessionUpdates(t *testing.T) {
	clientTest := &oidc.RegisteredClient{
		ID: testValue,
	}

	testCases := []struct {
		name      string
		elapsed   time.Duration
		errSave   []error
		handled   bool
		anonymous bool
		expected  *regexp.Regexp
	}{
		{"ShouldResetInactiveSession", time.Hour, nil, false, true, nil},
		{"ShouldHandleResetInactiveSessionSaveError", time.Hour, []error{errTestSessionBackend}, true, true, regexp.MustCompile(`had an error while saving updated session$`)},
		{"ShouldHandleRefreshedSessionSaveError", time.Minute, []error{errTestSessionBackend}, true, false, regexp.MustCompile(`had an error while saving updated session$`)},
		{"ShouldNotSaveCurrentSession", 0, []error{errTestSessionBackend}, false, false, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)

			defer mock.Close()

			mock.Ctx.Providers.Clock = &mock.Clock
			mock.Clock.Set(time.Unix(1000000, 0))

			mock.Ctx.Configuration.Session.Cookies[0].Inactivity = time.Minute * 5

			repository := setupTestFailingSessionRepository(t, mock)

			config := &schema.Configuration{
				IdentityProviders: schema.IdentityProviders{
					OIDC: &schema.IdentityProvidersOpenIDConnect{
						Clients: []schema.IdentityProvidersOpenIDConnectClient{
							{
								ID: "abc",
							},
						},
					},
				},
			}

			mock.Ctx.Providers.OpenIDConnect = oidc.NewOpenIDConnectProvider(config, mock.StorageMock, mock.Ctx.Providers.Templates)

			provider, err := mock.Ctx.GetSessionProvider()

			require.NoError(t, err)

			userSession := session.NewUserSession(testUsername)
			userSession.CookieDomain = "example.com"
			userSession.SetOneFactorPassword(mock.Clock.Now(), false)
			userSession.LastActivity = mock.Clock.Now().Add(-tc.elapsed).Unix()

			repository.errSave = tc.errSave

			requester := &oauthelia2.AuthorizeRequest{
				Request: oauthelia2.Request{
					Client: clientTest,
				},
			}

			handled := handleOAuth2AuthorizationConsentSessionUpdates(mock.Ctx, provider, &userSession, clientTest, oidc.ClientAuthorizationPolicy{Name: "one_factor"}, httptest.NewRecorder(), requester)

			assert.Equal(t, tc.handled, handled)
			assert.Equal(t, tc.anonymous, userSession.IsAnonymous())

			if tc.expected != nil {
				assert.Regexp(t, tc.expected, mock.Hook.LastEntry().Message)
			}
		})
	}
}
