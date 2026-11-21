// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/cache"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestLogoutPOSTShouldNotEndAnotherSessionFromAnAnonymousSession(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	clients := []schema.IdentityProvidersOpenIDConnectClient{{ID: "notified", BackChannelLogoutURI: rp.URL}}

	setupTestOIDCLogoutProvider(t, mock, clients...)

	subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", testUsername)
	require.NoError(t, err)

	sid := uuid.Must(uuid.NewRandom())

	_, issuer := setupTestOIDCLogoutPending(t, mock, "", model.OAuth2SessionLogout{ClientID: "notified", Subject: subject.String(), SessionID: sid.String(), RedirectURI: "https://app.example.com/logged-out"})

	repository := cache.NewSessionRepository(cache.NewMemory())
	require.NoError(t, repository.Save(mock.Ctx, issuer, "remote-signature", "remote-public-id", testUsername, time.Hour, []byte("data")))

	mock.Ctx.Providers.SessionRepository = repository

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false,"redirectURL":"https://app.example.com/logged-out"}}`, string(mock.Ctx.Response.Body()))
	assert.Empty(t, rp.Received())

	record, err := repository.GetByPublicID(mock.Ctx, issuer, "remote-public-id")
	require.NoError(t, err)

	assert.NotNil(t, record)
}

func TestLogoutPOSTShouldNotEndTheSubjectSessionsFromAnAnonymousSession(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	setupTestOIDCLogoutProvider(t, mock,
		schema.IdentityProvidersOpenIDConnectClient{ID: "requester", BackChannelLogoutURI: rp.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "sibling", BackChannelLogoutURI: rp.URL},
	)

	_, _ = setupTestOIDCLogoutPending(t, mock, "", model.OAuth2SessionLogout{ClientID: "requester", Subject: "pairwise-subject"})

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Empty(t, rp.Received())
}

func TestLogoutPOSTShouldEndTheEndedHintSessionOfTheAuthenticatedEndUser(t *testing.T) {
	testCases := []struct {
		name     string
		failures int
		recorded bool
	}{
		{name: "ShouldRecordTheEndedSession", recorded: true},
		{name: "ShouldRetryRecordingTheEndedSession", failures: 2, recorded: true},
		{name: "ShouldLogWhenTheEndedSessionCantBeRecorded", failures: 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			rp := newTestBackChannelLogoutRP(t)

			config := setupTestOIDCLogoutProvider(t, mock,
				schema.IdentityProvidersOpenIDConnectClient{ID: "requester", BackChannelLogoutURI: rp.URL},
				schema.IdentityProvidersOpenIDConnectClient{ID: "sibling", BackChannelLogoutURI: rp.URL},
			)

			subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", testUsername)
			require.NoError(t, err)

			sid := uuid.Must(uuid.NewRandom())

			userSession, issuer := setupTestOIDCLogoutPending(t, mock, testUsername, model.OAuth2SessionLogout{ClientID: "requester", Subject: subject.String(), SessionID: sid.String()})

			mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).AnyTimes().Return(nil, nil)

			store := setupTestBackChannelLogoutStore(mock, issuer, userSession.PublicID)
			store.failSaves = tc.failures
			store.failRevoke = !tc.recorded

			LogoutPOST(mock.Ctx)

			assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())

			requests := rp.Requests()

			var revoke []model.OAuth2SessionLogoutTask

			for _, task := range store.All() {
				if task.Kind == model.OAuth2SessionLogoutTaskKindRevoke {
					revoke = append(revoke, task)
				}
			}

			if tc.recorded {
				require.Len(t, revoke, 1)
				assert.Equal(t, sid.String(), revoke[0].SessionID)
				assert.Equal(t, revoke[0].CreatedAt.Add(oidc.LogoutRevokeGrace), revoke[0].NextAttemptAt)
				assert.Equal(t, revoke[0].CreatedAt.Add(oidc.LogoutRevokeHorizon(config)), revoke[0].ExpiresAt)
			} else {
				assert.Empty(t, revoke)
			}

			logged := map[string]bool{}

			for _, entry := range mock.Hook.AllEntries() {
				logged[entry.Message] = true
			}

			assert.Equal(t, !tc.recorded, logged["Error occurred ending the OpenID Connect session identified by the 'id_token_hint' during logout"])
			assert.False(t, logged["Error occurred revoking the refresh tokens of an ended OpenID Connect session during logout: they'll be revoked shortly"])
			assert.Equal(t, !tc.recorded, logged[fmt.Sprintf("Error occurred revoking the refresh tokens of an ended OpenID Connect session during logout: they won't be revoked as the end of the session with id '%s' could not be recorded, and its online refresh tokens remain usable until they expire", sid)])

			require.Len(t, requests, 1)

			issuerURL, err := mock.Ctx.IssuerURL()
			require.NoError(t, err)

			assertTestBackChannelLogoutRequest(t, requests[0], issuerURL.String(), "requester")

			assert.Equal(t, sid.String(), requests[0].claims[oidc.ClaimSessionID])
			assert.Equal(t, subject.String(), requests[0].claims[oidc.ClaimSubject])
		})
	}
}

func TestLogoutPOSTShouldIgnoreTheEndedHintSessionOfAnotherEndUser(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	clients := []schema.IdentityProvidersOpenIDConnectClient{{ID: "requester", BackChannelLogoutURI: rp.URL}}

	setupTestOIDCLogoutProvider(t, mock, clients...)

	subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", "another-user")
	require.NoError(t, err)

	sid := uuid.Must(uuid.NewRandom())

	userSession, issuer := setupTestOIDCLogoutPending(t, mock, testUsername, model.OAuth2SessionLogout{ClientID: "requester", Subject: subject.String(), SessionID: sid.String()})

	mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).AnyTimes().Return(nil, nil)
	mock.StorageMock.EXPECT().EndOAuth2Session(gomock.Any(), issuer, userSession.PublicID, gomock.Any()).Return(nil, nil)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Empty(t, rp.Received())
}

func TestOIDCLogoutEndedSessionID(t *testing.T) {
	sid := uuid.Must(uuid.NewRandom())

	testCases := []struct {
		name     string
		pending  *model.OAuth2SessionLogout
		setup    func(mock *mocks.MockAutheliaCtx)
		expected string
	}{
		{
			"ShouldReturnNothingWithoutAPendingLogout",
			nil,
			nil,
			"",
		},
		{
			"ShouldReturnNothingWithoutASessionID",
			&model.OAuth2SessionLogout{Subject: "subject"},
			nil,
			"",
		},
		{
			"ShouldReturnTheUnknownSessionID",
			&model.OAuth2SessionLogout{SessionID: sid.String()},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(nil, nil)
			},
			sid.String(),
		},
		{
			"ShouldReturnNothingForTheCurrentSession",
			&model.OAuth2SessionLogout{SessionID: sid.String()},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(&model.OAuth2SessionID{PublicID: "current-public-id"}, nil)
			},
			"",
		},
		{
			"ShouldReturnNothingForAnotherLiveSession",
			&model.OAuth2SessionLogout{SessionID: sid.String()},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(&model.OAuth2SessionID{PublicID: "remote-public-id"}, nil)
			},
			"",
		},
		{
			"ShouldReturnNothingWhenTheSessionFailsToLoad",
			&model.OAuth2SessionLogout{SessionID: sid.String()},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(nil, errors.New("connection refused"))
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			if tc.setup != nil {
				tc.setup(mock)
			}

			userSession := &session.UserSession{Username: testUsername, PublicID: "current-public-id"}

			assert.Equal(t, tc.expected, oidcLogoutEndedSessionID(mock.Ctx, "issuer", userSession, tc.pending))
		})
	}
}

func TestLogoutGETShouldDeclineAHintOfAnotherEndUser(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	setupTestOIDCLogoutProvider(t, mock,
		schema.IdentityProvidersOpenIDConnectClient{ID: "requester"},
	)

	_, _ = setupTestOIDCLogoutPending(t, mock, testUsername, model.OAuth2SessionLogout{ClientID: "requester", Subject: "not-a-uuid"})

	mock.Ctx.Request.URI().QueryArgs().Set(queryArgFlowID, testLogoutFlowID)

	LogoutGET(mock.Ctx)

	assert.Equal(t, `{"status":"KO","code":"operation_failed","message":"Operation failed."}`, string(mock.Ctx.Response.Body()))
}

func TestLogoutPOSTShouldDiscardAHintOfAnotherEndUser(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	setupTestOIDCLogoutProvider(t, mock,
		schema.IdentityProvidersOpenIDConnectClient{ID: "requester", BackChannelLogoutURI: rp.URL},
	)

	userSession, issuer := setupTestOIDCLogoutPending(t, mock, testUsername, model.OAuth2SessionLogout{ClientID: "requester", Subject: "not-a-uuid", RedirectURI: "https://app.example.com/logged-out"})

	mock.StorageMock.EXPECT().EndOAuth2Session(gomock.Any(), issuer, userSession.PublicID, gomock.Any()).Return(nil, nil)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))
	assert.True(t, strings.HasPrefix(string(mock.Ctx.Response.Header.PeekCookie("authelia_session")), "authelia_session=;"))
	assert.Empty(t, rp.Received())
}

func TestOIDCLogoutValidate(t *testing.T) {
	sid := uuid.Must(uuid.NewRandom())

	testCases := []struct {
		name     string
		username string
		setup    func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2SessionLogout
		expected string
	}{
		{
			name: "ShouldAcceptAnyHintWhenNotAuthenticated",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2SessionLogout {
				return &model.OAuth2SessionLogout{Subject: "someone", SessionID: sid.String()}
			},
		},
		{
			name:     "ShouldAcceptARequestWithoutAHint",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2SessionLogout {
				return &model.OAuth2SessionLogout{}
			},
		},
		{
			name:     "ShouldAcceptTheSessionOfTheLoggedInEndUser",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2SessionLogout {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(&model.OAuth2SessionID{PublicID: publicID}, nil)

				return &model.OAuth2SessionLogout{Subject: "ignored", SessionID: sid.String()}
			},
		},
		{
			name:     "ShouldDeclineAnotherSession",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2SessionLogout {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(&model.OAuth2SessionID{PublicID: "another-public-id"}, nil)

				return &model.OAuth2SessionLogout{SessionID: sid.String()}
			},
			expected: "the 'id_token_hint' identifies a session other than the one which is logged in",
		},
		{
			name:     "ShouldErrorWhenTheSessionFailsToLoad",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2SessionLogout {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(nil, errors.New("connection refused"))

				return &model.OAuth2SessionLogout{SessionID: sid.String()}
			},
			expected: "error loading the session identified by the 'id_token_hint': connection refused",
		},
		{
			name:     "ShouldAcceptTheSubjectOfTheLoggedInEndUserWhenTheSessionIsUnknown",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2SessionLogout {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(nil, nil)

				subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", testUsername)
				require.NoError(t, err)

				return &model.OAuth2SessionLogout{Subject: subject.String(), SessionID: sid.String()}
			},
		},
		{
			name:     "ShouldDeclineTheSubjectOfAnotherEndUser",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2SessionLogout {
				subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", "another-user")
				require.NoError(t, err)

				return &model.OAuth2SessionLogout{Subject: subject.String()}
			},
			expected: "the 'id_token_hint' was not issued to the currently authenticated End-User",
		},
		{
			name:     "ShouldDeclineAMalformedSubject",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2SessionLogout {
				return &model.OAuth2SessionLogout{Subject: "not-a-uuid"}
			},
			expected: "the 'id_token_hint' was not issued to the currently authenticated End-User",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			config := newTestOIDCConfig(t)
			config.SpecialUseAddressGuard.AllowedNetworks = newTestOIDCLoopbackNetworks(t)
			config.Clients = []schema.IdentityProvidersOpenIDConnectClient{{ID: "requester"}}

			setupTestOIDCProvider(t, mock, config)
			setupTestOIDCSubjectStore(t, mock)

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			userSession.Username = tc.username

			require.NoError(t, mock.Ctx.SaveSession(&userSession))

			provider, err := mock.Ctx.GetSessionProvider()
			require.NoError(t, err)

			pending := tc.setup(t, mock, provider.GetIssuer(), userSession.PublicID)

			err = oidcLogoutValidate(mock.Ctx, &userSession, pending)

			if tc.expected == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expected)
			}
		})
	}
}

func TestLogoutPOSTShouldNotRedirectWhenThePendingLogoutIsNotRemoved(t *testing.T) {
	testCases := []struct {
		name    string
		removed bool
		err     error
		log     string
	}{
		{"ShouldNotRedirectWhenConsumedByAnotherRequest", false, nil, "The pending logout was consumed by another request: the user is logged out without it"},
		{"ShouldNotRedirectWhenRemovingFails", false, errors.New("connection refused"), "Error occurred removing the pending logout: the user is logged out without it"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			mock.Ctx.Logger.Logger.SetLevel(logrus.DebugLevel)
			mock.Ctx.Configuration.IdentityProviders.OIDC = nil

			pending := model.OAuth2SessionLogout{
				FlowID:      uuid.MustParse(testLogoutFlowID),
				ClientID:    "requester",
				RedirectURI: "https://app.example.com/logged-out",
				ExpiresAt:   mock.Ctx.GetClock().Now().Add(time.Hour),
			}

			mock.StorageMock.EXPECT().LoadOAuth2SessionLogout(gomock.Any(), pending.FlowID).Return(&pending, nil)
			mock.StorageMock.EXPECT().DeleteOAuth2SessionLogout(gomock.Any(), pending.FlowID).Return(tc.removed, tc.err)

			mock.Ctx.Request.SetBodyString(`{"flowID":"` + testLogoutFlowID + `"}`)

			LogoutPOST(mock.Ctx)

			assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())
			assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))

			found := false

			for _, entry := range mock.Hook.AllEntries() {
				if entry.Message == tc.log {
					found = true
				}
			}

			assert.True(t, found)
		})
	}
}

func TestLogoutPOSTShouldRevokeTheEndedSessions(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	config := newTestOIDCConfig(t)
	config.SpecialUseAddressGuard.AllowedNetworks = newTestOIDCLoopbackNetworks(t)

	setupTestOIDCProvider(t, mock, config)

	userSession, issuer := mustSaveTestUserSession(t, mock, testUsername)

	a, b := uuid.Must(uuid.NewRandom()).String(), uuid.Must(uuid.NewRandom()).String()

	gomock.InOrder(
		mock.StorageMock.EXPECT().
			EndOAuth2Session(gomock.Any(), issuer, userSession.PublicID, gomock.Any()).
			DoAndReturn(func(_ context.Context, _, _ string, schedule model.OAuth2SessionLogoutTaskSchedule) ([]string, error) {
				assert.Equal(t, schedule.CreatedAt.Add(oidc.LogoutRevokeGrace), schedule.NextAttemptAt)
				assert.Equal(t, schedule.CreatedAt.Add(oidc.LogoutRevokeHorizon(config)), schedule.ExpiresAt)

				return []string{a, b}, nil
			}),
		mock.StorageMock.EXPECT().RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), a).Return(nil),
		mock.StorageMock.EXPECT().RevokeOAuth2RefreshTokenSessionsBySessionID(gomock.Any(), b).Return(errors.New("connection refused")),
	)

	mock.StorageMock.EXPECT().
		LoadOAuth2SessionLogoutTasksBySessionID(gomock.Any(), gomock.Any(), model.OAuth2SessionLogoutTaskKindBackChannel).
		Times(2).
		Return(nil, nil)

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	oidcBackChannelLogoutDeliveries.Wait()

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())

	AssertLogEntryMessageAndError(t, mock.Hook.LastEntry(), "Error occurred revoking the refresh tokens of an ended OpenID Connect session during logout: they'll be revoked shortly", "connection refused")
}

func setupTestOIDCLogoutProvider(t *testing.T, mock *mocks.MockAutheliaCtx, clients ...schema.IdentityProvidersOpenIDConnectClient) (config *schema.IdentityProvidersOpenIDConnect) {
	t.Helper()

	config = newTestOIDCConfig(t)
	config.SpecialUseAddressGuard.AllowedNetworks = newTestOIDCLoopbackNetworks(t)
	config.Clients = clients

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSubjectStore(t, mock)

	return config
}

func setupTestOIDCLogoutPending(t *testing.T, mock *mocks.MockAutheliaCtx, username string, pending model.OAuth2SessionLogout) (userSession session.UserSession, issuer string) {
	t.Helper()

	userSession, issuer = mustSaveTestUserSession(t, mock, username)

	pending.FlowID = uuid.MustParse(testLogoutFlowID)
	pending.ExpiresAt = mock.Ctx.GetClock().Now().Add(time.Hour)

	mock.StorageMock.EXPECT().LoadOAuth2SessionLogout(gomock.Any(), pending.FlowID).AnyTimes().Return(&pending, nil)
	mock.StorageMock.EXPECT().DeleteOAuth2SessionLogout(gomock.Any(), pending.FlowID).MaxTimes(1).Return(true, nil)

	mock.Ctx.Request.SetBodyString(`{"flowID":"` + testLogoutFlowID + `"}`)

	return userSession, issuer
}
