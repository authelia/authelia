// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
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

func setupTestOIDCLogoutPending(t *testing.T, mock *mocks.MockAutheliaCtx, username string, pending model.OAuth2LogoutSession, clients ...schema.IdentityProvidersOpenIDConnectClient) (userSession session.UserSession, issuer string) {
	t.Helper()

	config := newTestOIDCConfig(t)
	config.Clients = clients

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSubjectStore(t, mock)

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	userSession.Username = username

	require.NoError(t, mock.Ctx.SaveSession(&userSession))

	pending.FlowID = uuid.MustParse(testLogoutFlowID)
	pending.ExpiresAt = mock.Ctx.GetClock().Now().Add(time.Hour)

	mock.StorageMock.EXPECT().LoadOAuth2LogoutSession(gomock.Any(), pending.FlowID).AnyTimes().Return(&pending, nil)
	mock.StorageMock.EXPECT().DeleteOAuth2LogoutSession(gomock.Any(), pending.FlowID).MaxTimes(1).Return(nil)

	provider, err := mock.Ctx.GetSessionProvider()
	require.NoError(t, err)

	mock.Ctx.Request.SetBodyString(`{"flowID":"` + testLogoutFlowID + `"}`)

	return userSession, provider.GetIssuer()
}

func TestLogoutPOSTShouldNotEndAnotherSessionFromAnAnonymousSession(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	clients := []schema.IdentityProvidersOpenIDConnectClient{{ID: "notified", BackChannelLogoutURI: rp.URL}}

	config := newTestOIDCConfig(t)
	config.Clients = clients

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSubjectStore(t, mock)

	subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", testUsername)
	require.NoError(t, err)

	sid := uuid.Must(uuid.NewRandom())

	_, issuer := setupTestOIDCLogoutPending(t, mock, "", model.OAuth2LogoutSession{ClientID: "notified", Subject: subject.String(), SessionID: sid.String(), RedirectURI: "https://app.example.com/logged-out"}, clients...)

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

	_, _ = setupTestOIDCLogoutPending(t, mock, "", model.OAuth2LogoutSession{ClientID: "requester", Subject: "pairwise-subject"},
		schema.IdentityProvidersOpenIDConnectClient{ID: "requester", BackChannelLogoutURI: rp.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "sibling", BackChannelLogoutURI: rp.URL},
	)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Empty(t, rp.Received())
}

func TestLogoutPOSTShouldEndTheEndedHintSessionOfTheAuthenticatedEndUser(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	clients := []schema.IdentityProvidersOpenIDConnectClient{
		{ID: "requester", BackChannelLogoutURI: rp.URL},
		{ID: "sibling", BackChannelLogoutURI: rp.URL},
	}

	config := newTestOIDCConfig(t)
	config.Clients = clients

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSubjectStore(t, mock)

	subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", testUsername)
	require.NoError(t, err)

	sid := uuid.Must(uuid.NewRandom())

	userSession, issuer := setupTestOIDCLogoutPending(t, mock, testUsername, model.OAuth2LogoutSession{ClientID: "requester", Subject: subject.String(), SessionID: sid.String()}, clients...)

	mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).AnyTimes().Return(nil, nil)
	mock.StorageMock.EXPECT().LoadOAuth2SessionIDClientsByPublicID(gomock.Any(), issuer, userSession.PublicID).Return(nil, nil)
	mock.StorageMock.EXPECT().LoadOAuth2SessionIDsByPublicID(gomock.Any(), issuer, userSession.PublicID).Return(nil, nil)
	mock.StorageMock.EXPECT().DeleteOAuth2SessionIDByPublicID(gomock.Any(), issuer, userSession.PublicID).Return(nil)
	mock.StorageMock.EXPECT().RevokeOAuth2SessionsBySessionID(gomock.Any(), sid.String()).Return(nil)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())

	requests := rp.Requests()

	require.Len(t, requests, 1)

	issuerURL, err := mock.Ctx.IssuerURL()
	require.NoError(t, err)

	assertTestBackChannelLogoutRequest(t, requests[0], issuerURL.String(), "requester")

	assert.Equal(t, sid.String(), requests[0].claims[oidc.ClaimSessionID])
	assert.Equal(t, subject.String(), requests[0].claims[oidc.ClaimSubject])
}

func TestLogoutPOSTShouldIgnoreTheEndedHintSessionOfAnotherEndUser(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	clients := []schema.IdentityProvidersOpenIDConnectClient{{ID: "requester", BackChannelLogoutURI: rp.URL}}

	config := newTestOIDCConfig(t)
	config.Clients = clients

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSubjectStore(t, mock)

	subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", "another-user")
	require.NoError(t, err)

	sid := uuid.Must(uuid.NewRandom())

	userSession, issuer := setupTestOIDCLogoutPending(t, mock, testUsername, model.OAuth2LogoutSession{ClientID: "requester", Subject: subject.String(), SessionID: sid.String()}, clients...)

	mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).AnyTimes().Return(nil, nil)
	mock.StorageMock.EXPECT().LoadOAuth2SessionIDClientsByPublicID(gomock.Any(), issuer, userSession.PublicID).Return(nil, nil)
	mock.StorageMock.EXPECT().LoadOAuth2SessionIDsByPublicID(gomock.Any(), issuer, userSession.PublicID).Return(nil, nil)
	mock.StorageMock.EXPECT().DeleteOAuth2SessionIDByPublicID(gomock.Any(), issuer, userSession.PublicID).Return(nil)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Empty(t, rp.Received())
}

func TestOIDCLogoutEndedSessionID(t *testing.T) {
	sid := uuid.Must(uuid.NewRandom())

	testCases := []struct {
		name     string
		pending  *model.OAuth2LogoutSession
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
			&model.OAuth2LogoutSession{Subject: "subject"},
			nil,
			"",
		},
		{
			"ShouldReturnTheUnknownSessionID",
			&model.OAuth2LogoutSession{SessionID: sid.String()},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(nil, nil)
			},
			sid.String(),
		},
		{
			"ShouldReturnNothingForTheCurrentSession",
			&model.OAuth2LogoutSession{SessionID: sid.String()},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(&model.OAuth2SessionID{PublicID: "current-public-id"}, nil)
			},
			"",
		},
		{
			"ShouldReturnNothingForAnotherLiveSession",
			&model.OAuth2LogoutSession{SessionID: sid.String()},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(&model.OAuth2SessionID{PublicID: "remote-public-id"}, nil)
			},
			"",
		},
		{
			"ShouldReturnNothingWhenTheSessionFailsToLoad",
			&model.OAuth2LogoutSession{SessionID: sid.String()},
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

	_, _ = setupTestOIDCLogoutPending(t, mock, testUsername, model.OAuth2LogoutSession{ClientID: "requester", Subject: "not-a-uuid"},
		schema.IdentityProvidersOpenIDConnectClient{ID: "requester"},
	)

	mock.Ctx.Request.URI().QueryArgs().Set(queryArgFlowID, testLogoutFlowID)

	LogoutGET(mock.Ctx)

	assert.Equal(t, `{"status":"KO","code":"operation_failed","message":"Operation failed."}`, string(mock.Ctx.Response.Body()))
}

func TestLogoutPOSTShouldDiscardAHintOfAnotherEndUser(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	userSession, issuer := setupTestOIDCLogoutPending(t, mock, testUsername, model.OAuth2LogoutSession{ClientID: "requester", Subject: "not-a-uuid", RedirectURI: "https://app.example.com/logged-out"},
		schema.IdentityProvidersOpenIDConnectClient{ID: "requester", BackChannelLogoutURI: rp.URL},
	)

	mock.StorageMock.EXPECT().LoadOAuth2SessionIDClientsByPublicID(gomock.Any(), issuer, userSession.PublicID).Return(nil, nil)
	mock.StorageMock.EXPECT().LoadOAuth2SessionIDsByPublicID(gomock.Any(), issuer, userSession.PublicID).Return(nil, nil)
	mock.StorageMock.EXPECT().DeleteOAuth2SessionIDByPublicID(gomock.Any(), issuer, userSession.PublicID).Return(nil)

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
		setup    func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2LogoutSession
		expected string
	}{
		{
			name: "ShouldAcceptAnyHintWhenNotAuthenticated",
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2LogoutSession {
				return &model.OAuth2LogoutSession{Subject: "someone", SessionID: sid.String()}
			},
		},
		{
			name:     "ShouldAcceptARequestWithoutAHint",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2LogoutSession {
				return &model.OAuth2LogoutSession{}
			},
		},
		{
			name:     "ShouldAcceptTheSessionOfTheLoggedInEndUser",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2LogoutSession {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(&model.OAuth2SessionID{PublicID: publicID}, nil)

				return &model.OAuth2LogoutSession{Subject: "ignored", SessionID: sid.String()}
			},
		},
		{
			name:     "ShouldDeclineAnotherSession",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2LogoutSession {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(&model.OAuth2SessionID{PublicID: "another-public-id"}, nil)

				return &model.OAuth2LogoutSession{SessionID: sid.String()}
			},
			expected: "the 'id_token_hint' identifies a session other than the one which is logged in",
		},
		{
			name:     "ShouldErrorWhenTheSessionFailsToLoad",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2LogoutSession {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(nil, errors.New("connection refused"))

				return &model.OAuth2LogoutSession{SessionID: sid.String()}
			},
			expected: "error loading the session identified by the 'id_token_hint': connection refused",
		},
		{
			name:     "ShouldAcceptTheSubjectOfTheLoggedInEndUserWhenTheSessionIsUnknown",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2LogoutSession {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(nil, nil)

				subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", testUsername)
				require.NoError(t, err)

				return &model.OAuth2LogoutSession{Subject: subject.String(), SessionID: sid.String()}
			},
		},
		{
			name:     "ShouldDeclineTheSubjectOfAnotherEndUser",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2LogoutSession {
				subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", "another-user")
				require.NoError(t, err)

				return &model.OAuth2LogoutSession{Subject: subject.String()}
			},
			expected: "the 'id_token_hint' was not issued to the currently authenticated End-User",
		},
		{
			name:     "ShouldDeclineAMalformedSubject",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) *model.OAuth2LogoutSession {
				return &model.OAuth2LogoutSession{Subject: "not-a-uuid"}
			},
			expected: "the 'id_token_hint' was not issued to the currently authenticated End-User",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			config := newTestOIDCConfig(t)
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
