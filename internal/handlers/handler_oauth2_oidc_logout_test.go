// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/cache"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

// setupTestOIDCLogoutPending configures the OpenID Connect provider with the clients, and stores a pending
// RP-Initiated Logout in an anonymous session, as if the End-User reached the end session endpoint while logged out.
func setupTestOIDCLogoutPending(t *testing.T, mock *mocks.MockAutheliaCtx, pending session.OpenIDConnectLogout, clients ...schema.IdentityProvidersOpenIDConnectClient) (userSession session.UserSession, issuer string) {
	t.Helper()

	config := newTestOIDCConfig(t)
	config.Clients = clients

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSubjectStore(t, mock)

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	pending.FlowID = "flow-id"
	pending.Expires = mock.Ctx.GetClock().Now().Add(time.Hour)

	userSession.OpenIDConnectLogout = &pending

	require.NoError(t, mock.Ctx.SaveSession(&userSession))

	provider, err := mock.Ctx.GetSessionProvider()
	require.NoError(t, err)

	mock.Ctx.Request.SetBodyString(`{"flowID":"flow-id"}`)

	return userSession, provider.GetIssuer()
}

func TestLogoutPOSTShouldEndTheSessionIdentifiedByTheIDTokenHint(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	config := newTestOIDCConfig(t)
	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{{ID: "notified", BackChannelLogoutURI: rp.URL}}

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSubjectStore(t, mock)

	subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", testUsername)
	require.NoError(t, err)

	sid := uuid.Must(uuid.NewRandom())

	userSession, issuer := setupTestOIDCLogoutPending(t, mock, session.OpenIDConnectLogout{ClientID: "notified", Subject: subject.String(), SessionID: sid.String()}, config.Clients...)

	require.Empty(t, userSession.Username, "the End-User is not logged in to the browser which requested the logout")

	// The session the hint identifies exists in another browser.
	repository := cache.NewSessionRepository(cache.NewMemory())
	require.NoError(t, repository.Save(mock.Ctx, issuer, "remote-signature", "remote-public-id", testUsername, time.Hour, []byte("data")))

	mock.Ctx.Providers.SessionRepository = repository

	gomock.InOrder(
		mock.StorageMock.EXPECT().
			LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).
			Return(&model.OAuth2SessionID{Issuer: issuer, PublicID: "remote-public-id", SessionID: sid}, nil),
		mock.StorageMock.EXPECT().
			LoadOAuth2SessionIDClientsByPublicID(gomock.Any(), issuer, "remote-public-id").
			Return([]model.OAuth2SessionIDClient{{Issuer: issuer, PublicID: "remote-public-id", SessionID: sid, ClientID: "notified"}}, nil),
		mock.StorageMock.EXPECT().
			LoadOAuth2SessionIDsByPublicID(gomock.Any(), issuer, "remote-public-id").
			Return([]model.OAuth2SessionID{{Issuer: issuer, PublicID: "remote-public-id", SessionID: sid}}, nil),
		mock.StorageMock.EXPECT().
			RevokeOAuth2SessionsBySessionID(gomock.Any(), sid.String()).
			Return(nil),
		mock.StorageMock.EXPECT().
			DeleteOAuth2SessionIDByPublicID(gomock.Any(), issuer, "remote-public-id").
			Return(nil),
	)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())

	received := rp.Received()

	require.Len(t, received, 1)

	assert.Equal(t, sid.String(), received[0][oidc.ClaimSessionID])
	assert.Equal(t, subject.String(), received[0][oidc.ClaimSubject])

	issuerURL, err := mock.Ctx.IssuerURL()
	require.NoError(t, err)

	assertTestBackChannelLogoutRequest(t, rp.Requests()[0], issuerURL.String(), "notified")

	record, err := repository.GetByPublicID(mock.Ctx, issuer, "remote-public-id")
	require.NoError(t, err)

	assert.Nil(t, record, "the session identified by the hint must be destroyed")
}

func TestLogoutPOSTShouldEndTheSubjectSessionsOfTheSectorWithoutASessionID(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	sector := &url.URL{Scheme: "https", Host: "sector.example.com"}

	_, _ = setupTestOIDCLogoutPending(t, mock, session.OpenIDConnectLogout{ClientID: "requester", Subject: "pairwise-subject"},
		schema.IdentityProvidersOpenIDConnectClient{ID: "requester", BackChannelLogoutURI: rp.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "sibling", BackChannelLogoutURI: rp.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "idle", BackChannelLogoutURI: rp.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "requires-sid", BackChannelLogoutURI: rp.URL, BackChannelLogoutSessionRequired: true},
		schema.IdentityProvidersOpenIDConnectClient{ID: "other-sector", BackChannelLogoutURI: rp.URL, SectorIdentifierURI: sector},
	)

	mock.StorageMock.EXPECT().HasOAuth2SessionsByClientIDAndSubject(gomock.Any(), "requester", "pairwise-subject").Return(true, nil)
	mock.StorageMock.EXPECT().HasOAuth2SessionsByClientIDAndSubject(gomock.Any(), "sibling", "pairwise-subject").Return(true, nil)
	mock.StorageMock.EXPECT().HasOAuth2SessionsByClientIDAndSubject(gomock.Any(), "idle", "pairwise-subject").Return(false, nil)
	mock.StorageMock.EXPECT().HasOAuth2SessionsByClientIDAndSubject(gomock.Any(), "requires-sid", "pairwise-subject").Return(true, nil)

	mock.StorageMock.EXPECT().RevokeOAuth2SessionsByClientIDAndSubject(gomock.Any(), "requester", "pairwise-subject").Return(nil)
	mock.StorageMock.EXPECT().RevokeOAuth2SessionsByClientIDAndSubject(gomock.Any(), "sibling", "pairwise-subject").Return(nil)
	mock.StorageMock.EXPECT().RevokeOAuth2SessionsByClientIDAndSubject(gomock.Any(), "requires-sid", "pairwise-subject").Return(nil)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())

	received := rp.Received()

	require.Len(t, received, 2)

	for _, claims := range received {
		aud, _ := claims[oidc.ClaimAudience].([]any)

		require.Len(t, aud, 1)
		assert.NotEqual(t, "requires-sid", aud[0])
	}

	issuerURL, err := mock.Ctx.IssuerURL()
	require.NoError(t, err)

	for _, request := range rp.Requests() {
		// Section 2.4: without a 'sid' the Logout Token identifies the End-User alone, which logs out every session
		// they have at the Relying Party.
		assert.Equal(t, "pairwise-subject", request.claims[oidc.ClaimSubject])
		assert.NotContains(t, request.claims, oidc.ClaimSessionID)

		aud, _ := request.claims[oidc.ClaimAudience].([]any)

		require.Len(t, aud, 1)

		assertTestBackChannelLogoutRequest(t, request, issuerURL.String(), aud[0].(string))
	}
}

func TestOIDCLogoutResolveTarget(t *testing.T) {
	sid := uuid.Must(uuid.NewRandom())

	authenticated := &session.UserSession{Username: testUsername, PublicID: "current-public-id"}
	anonymous := &session.UserSession{PublicID: "anonymous-public-id"}

	testCases := []struct {
		name        string
		userSession *session.UserSession
		pending     *session.OpenIDConnectLogout
		setup       func(mock *mocks.MockAutheliaCtx)
		expected    oidcLogoutTarget
	}{
		{
			"ShouldUseTheCurrentSessionWithoutAPendingLogout",
			authenticated,
			nil,
			nil,
			oidcLogoutTarget{publicID: "current-public-id", username: testUsername, current: true},
		},
		{
			"ShouldUseNothingForAnAnonymousSessionWithoutAPendingLogout",
			anonymous,
			nil,
			nil,
			oidcLogoutTarget{},
		},
		{
			"ShouldUseTheCurrentSessionWhenTheHintHasNoSessionID",
			authenticated,
			&session.OpenIDConnectLogout{Subject: "subject"},
			nil,
			oidcLogoutTarget{publicID: "current-public-id", username: testUsername, current: true},
		},
		{
			"ShouldUseTheHintSessionForAnAnonymousSession",
			anonymous,
			&session.OpenIDConnectLogout{SessionID: sid.String(), Subject: "not-a-uuid"},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(&model.OAuth2SessionID{PublicID: "remote-public-id"}, nil)
			},
			oidcLogoutTarget{publicID: "remote-public-id"},
		},
		{
			"ShouldPreferTheCurrentSessionOverAnotherHintSession",
			authenticated,
			&session.OpenIDConnectLogout{SessionID: sid.String()},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(&model.OAuth2SessionID{PublicID: "remote-public-id"}, nil)
			},
			oidcLogoutTarget{publicID: "current-public-id", username: testUsername, current: true},
		},
		{
			"ShouldUseNothingForAnAnonymousSessionWithAnUnknownHintSession",
			anonymous,
			&session.OpenIDConnectLogout{SessionID: sid.String()},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(nil, nil)
			},
			oidcLogoutTarget{},
		},
		{
			"ShouldUseTheCurrentSessionWhenTheHintSessionFailsToLoad",
			authenticated,
			&session.OpenIDConnectLogout{SessionID: sid.String()},
			func(mock *mocks.MockAutheliaCtx) {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), "issuer", sid.String()).Return(nil, errors.New("connection refused"))
			},
			oidcLogoutTarget{publicID: "current-public-id", username: testUsername, current: true},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtx(t)
			defer mock.Close()

			if tc.setup != nil {
				tc.setup(mock)
			}

			assert.Equal(t, tc.expected, oidcLogoutResolveTarget(mock.Ctx, "issuer", tc.userSession, tc.pending))
		})
	}
}

func TestOIDCEndSessionValidateHint(t *testing.T) {
	sid := uuid.Must(uuid.NewRandom())

	testCases := []struct {
		name     string
		username string
		claims   jwt.MapClaims
		setup    func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) jwt.MapClaims
		expected string
	}{
		{
			name:   "ShouldAcceptAnyHintWhenNotAuthenticated",
			claims: jwt.MapClaims{oidc.ClaimSubject: "someone", oidc.ClaimSessionID: sid.String()},
		},
		{
			name:     "ShouldAcceptTheSessionOfTheLoggedInEndUser",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) jwt.MapClaims {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(&model.OAuth2SessionID{PublicID: publicID}, nil)

				return jwt.MapClaims{oidc.ClaimSubject: "ignored", oidc.ClaimSessionID: sid.String()}
			},
		},
		{
			name:     "ShouldDeclineAnotherSession",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) jwt.MapClaims {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(&model.OAuth2SessionID{PublicID: "another-public-id"}, nil)

				return jwt.MapClaims{oidc.ClaimSessionID: sid.String()}
			},
			expected: "The 'id_token_hint' identifies a session other than the one which is logged in.",
		},
		{
			name:     "ShouldAcceptTheSubjectOfTheLoggedInEndUserWhenTheSessionIsUnknown",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) jwt.MapClaims {
				mock.StorageMock.EXPECT().LoadOAuth2SessionIDBySessionID(gomock.Any(), issuer, sid.String()).Return(nil, nil)

				subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", testUsername)
				require.NoError(t, err)

				return jwt.MapClaims{oidc.ClaimSubject: subject.String(), oidc.ClaimSessionID: sid.String()}
			},
		},
		{
			name:     "ShouldDeclineTheSubjectOfAnotherEndUser",
			username: testUsername,
			setup: func(t *testing.T, mock *mocks.MockAutheliaCtx, issuer, publicID string) jwt.MapClaims {
				subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", "another-user")
				require.NoError(t, err)

				return jwt.MapClaims{oidc.ClaimSubject: subject.String()}
			},
			expected: "The 'id_token_hint' was not issued to the currently authenticated End-User.",
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

			if tc.username != "" {
				userSession.SetOneFactorPassword(mock.Ctx.GetClock().Now(), false)
			}

			require.NoError(t, mock.Ctx.SaveSession(&userSession))

			provider, err := mock.Ctx.GetSessionProvider()
			require.NoError(t, err)

			claims := tc.claims

			if tc.setup != nil {
				claims = tc.setup(t, mock, provider.GetIssuer(), userSession.PublicID)
			}

			client, err := mock.Ctx.Providers.OpenIDConnect.GetRegisteredClient(mock.Ctx, "requester")
			require.NoError(t, err)

			err = oidcEndSessionValidateHint(mock.Ctx, &oauthelia2.RPInitiatedLogoutRequest{IDTokenHintClaims: claims, Client: client})

			if tc.expected == "" {
				assert.NoError(t, err)
			} else {
				rfc := oauthelia2.ErrorToRFC6749Error(err)

				assert.Equal(t, oauthelia2.ErrInvalidRequest.ErrorField, rfc.ErrorField)
				assert.Equal(t, tc.expected, rfc.HintField)
			}
		})
	}
}
