// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestLogoutPOSTShouldDeliverBackChannelLogoutBeforeRemovingSessionIDs(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	rp := newTestBackChannelLogoutRP(t)

	sector := &url.URL{Scheme: "https", Host: "sector.example.com"}

	userSession, issuer := setupTestBackChannelLogout(t, mock,
		schema.IdentityProvidersOpenIDConnectClient{ID: "notified", BackChannelLogoutURI: rp.URL},
		schema.IdentityProvidersOpenIDConnectClient{ID: "notified-sector", BackChannelLogoutURI: rp.URL, SectorIdentifierURI: sector},
		schema.IdentityProvidersOpenIDConnectClient{ID: "no-uri"},
	)

	sid, sidSector := uuid.Must(uuid.NewRandom()), uuid.Must(uuid.NewRandom())

	gomock.InOrder(
		mock.StorageMock.EXPECT().
			LoadOAuth2SessionIDClientsByPublicID(gomock.Any(), issuer, userSession.PublicID).
			Return([]model.OAuth2SessionIDClient{
				{Issuer: issuer, PublicID: userSession.PublicID, SessionID: sid, ClientID: "notified"},
				{Issuer: issuer, PublicID: userSession.PublicID, SessionID: sidSector, ClientID: "notified-sector"},
				{Issuer: issuer, PublicID: userSession.PublicID, SessionID: sid, ClientID: "no-uri"},
				{Issuer: issuer, PublicID: userSession.PublicID, SessionID: sid, ClientID: "removed"},
			}, nil),
		mock.StorageMock.EXPECT().
			LoadOAuth2SessionIDsByPublicID(gomock.Any(), issuer, userSession.PublicID).
			Return([]model.OAuth2SessionID{
				{Issuer: issuer, PublicID: userSession.PublicID, SessionID: sid},
				{Issuer: issuer, PublicID: userSession.PublicID, SessionID: sidSector},
			}, nil),
		mock.StorageMock.EXPECT().
			RevokeOAuth2SessionsBySessionID(gomock.Any(), sid.String()).
			DoAndReturn(func(_ any, _ string) error {
				// The Logout Tokens must have been delivered before the OAuth 2.0 sessions are revoked.
				assert.Len(t, rp.Received(), 2)

				return nil
			}),
		mock.StorageMock.EXPECT().
			RevokeOAuth2SessionsBySessionID(gomock.Any(), sidSector.String()).
			Return(nil),
		mock.StorageMock.EXPECT().
			DeleteOAuth2SessionIDByPublicID(gomock.Any(), issuer, userSession.PublicID).
			Return(nil),
	)

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())

	received := rp.Received()

	require.Len(t, received, 2)

	subject, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, "", testUsername)
	require.NoError(t, err)

	subjectSector, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, sector.String(), testUsername)
	require.NoError(t, err)

	byAudience := map[string]map[string]any{}

	for _, claims := range received {
		aud, ok := claims[oidc.ClaimAudience].([]any)
		require.True(t, ok)
		require.Len(t, aud, 1)

		byAudience[aud[0].(string)] = claims
	}

	require.Contains(t, byAudience, "notified")
	require.Contains(t, byAudience, "notified-sector")

	issuerURL, err := mock.Ctx.IssuerURL()
	require.NoError(t, err)

	for _, claims := range received {
		assert.Equal(t, issuerURL.String(), claims[oidc.ClaimIssuer])
	}

	assert.Equal(t, sid.String(), byAudience["notified"][oidc.ClaimSessionID])
	assert.Equal(t, subject.String(), byAudience["notified"][oidc.ClaimSubject])
	assert.Equal(t, sidSector.String(), byAudience["notified-sector"][oidc.ClaimSessionID])
	assert.Equal(t, subjectSector.String(), byAudience["notified-sector"][oidc.ClaimSubject])
}

func TestLogoutPOSTShouldRemoveSessionIDsWhenParticipantsFailToLoad(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	userSession, issuer := setupTestBackChannelLogout(t, mock)

	gomock.InOrder(
		mock.StorageMock.EXPECT().
			LoadOAuth2SessionIDClientsByPublicID(gomock.Any(), issuer, userSession.PublicID).
			Return(nil, errors.New("connection refused")),
		mock.StorageMock.EXPECT().
			LoadOAuth2SessionIDsByPublicID(gomock.Any(), issuer, userSession.PublicID).
			Return(nil, nil),
		mock.StorageMock.EXPECT().
			DeleteOAuth2SessionIDByPublicID(gomock.Any(), issuer, userSession.PublicID).
			Return(nil),
	)

	mock.Ctx.Request.SetBodyString(`{}`)

	LogoutPOST(mock.Ctx)

	assert.Equal(t, http.StatusOK, mock.Ctx.Response.StatusCode())
	assert.Equal(t, `{"status":"OK","data":{"safeTargetURL":false}}`, string(mock.Ctx.Response.Body()))
}

func TestOIDCBackChannelLogoutSessionShouldSkipWithoutPublicID(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	setupTestOIDCProvider(t, mock, nil)

	oidcBackChannelLogoutSession(mock.Ctx, "issuer", testUsername, "")
}

func TestOIDCBackChannelLogoutGroups(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	one := &url.URL{Scheme: "https", Host: "one.example.com"}
	two := &url.URL{Scheme: "https", Host: "two.example.com"}

	config := newTestOIDCConfig(t)
	config.Clients = []schema.IdentityProvidersOpenIDConnectClient{
		{ID: "a", BackChannelLogoutURI: "https://a.example.com/logout"},
		{ID: "b", BackChannelLogoutURI: "https://b.example.com/logout"},
		{ID: "c", BackChannelLogoutURI: "https://c.example.com/logout", SectorIdentifierURI: one},
		{ID: "d", BackChannelLogoutURI: "https://d.example.com/logout", SectorIdentifierURI: two},
		{ID: "e", BackChannelLogoutURI: "https://e.example.com/logout", SectorIdentifierURI: one},
		{ID: "no-uri"},
	}

	setupTestOIDCProvider(t, mock, config)

	sid, sidOne, sidTwo := uuid.Must(uuid.NewRandom()), uuid.Must(uuid.NewRandom()), uuid.Must(uuid.NewRandom())

	groups := oidcBackChannelLogoutGroups(mock.Ctx, testUsername, []model.OAuth2SessionIDClient{
		{SessionID: sid, ClientID: "a"},
		{SessionID: sidOne, ClientID: "c"},
		{SessionID: sid, ClientID: "no-uri"},
		{SessionID: sid, ClientID: "b"},
		{SessionID: sidTwo, ClientID: "d"},
		{SessionID: sid, ClientID: "removed"},
		{SessionID: sidOne, ClientID: "e"},
	})

	type group struct {
		sid, sectorID string
		clients       []string
	}

	actual := make([]group, len(groups))

	for i, g := range groups {
		actual[i] = group{sid: g.sid, sectorID: g.sectorID}

		for _, client := range g.clients {
			actual[i].clients = append(actual[i].clients, client.GetID())
		}
	}

	assert.Equal(t, []group{
		{sid: sid.String(), sectorID: "", clients: []string{"a", "b"}},
		{sid: sidOne.String(), sectorID: one.String(), clients: []string{"c", "e"}},
		{sid: sidTwo.String(), sectorID: two.String(), clients: []string{"d"}},
	}, actual)
}

type testBackChannelLogoutRP struct {
	*httptest.Server

	mu     sync.Mutex
	claims []map[string]any
}

func newTestBackChannelLogoutRP(t *testing.T) (rp *testBackChannelLogoutRP) {
	t.Helper()

	rp = &testBackChannelLogoutRP{}

	rp.Server = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			rw.WriteHeader(http.StatusBadRequest)

			return
		}

		parts := strings.Split(r.PostFormValue("logout_token"), ".")
		if len(parts) != 3 {
			rw.WriteHeader(http.StatusBadRequest)

			return
		}

		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)

			return
		}

		claims := map[string]any{}

		if err = json.Unmarshal(payload, &claims); err != nil {
			rw.WriteHeader(http.StatusBadRequest)

			return
		}

		rp.mu.Lock()
		rp.claims = append(rp.claims, claims)
		rp.mu.Unlock()

		rw.WriteHeader(http.StatusOK)
	}))

	t.Cleanup(rp.Close)

	return rp
}

func (rp *testBackChannelLogoutRP) Received() []map[string]any {
	rp.mu.Lock()
	defer rp.mu.Unlock()

	return append([]map[string]any(nil), rp.claims...)
}

func setupTestBackChannelLogout(t *testing.T, mock *mocks.MockAutheliaCtx, clients ...schema.IdentityProvidersOpenIDConnectClient) (userSession session.UserSession, issuer string) {
	t.Helper()

	config := newTestOIDCConfig(t)
	config.Clients = clients

	setupTestOIDCProvider(t, mock, config)
	setupTestOIDCSubjectStore(t, mock)

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	userSession.Username = testUsername

	require.NoError(t, mock.Ctx.SaveSession(&userSession))
	require.NotEmpty(t, userSession.PublicID)

	provider, err := mock.Ctx.GetSessionProvider()
	require.NoError(t, err)

	return userSession, provider.GetIssuer()
}
