// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"errors"
	"net/url"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	oauthelia2 "authelia.com/provider/oauth2"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
	"github.com/authelia/authelia/v4/internal/session"
)

func TestOIDCSessionIDShouldBeEmptyWithoutOpenIDScope(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{Username: "john", PublicID: "public-id"})
	defer mock.Close()

	client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{ID: "client-id"}, &schema.IdentityProvidersOpenIDConnect{}, nil)

	requester := oauthelia2.NewRequest()
	requester.GrantScope("profile")

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	sid, err := oidcSessionID(mock.Ctx, client, requester, &userSession)

	assert.NoError(t, err)
	assert.Empty(t, sid)
}

func TestOIDCSessionIDShouldMintForOpenIDScope(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{Username: "john", PublicID: "public-id"})
	defer mock.Close()

	client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{
		ID:                  "client-id",
		SectorIdentifierURI: &url.URL{Scheme: "https", Host: "sector.example.com"},
	}, &schema.IdentityProvidersOpenIDConnect{}, nil)

	requester := oauthelia2.NewRequest()
	requester.GrantScope(oidc.ScopeOpenID)

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	provider, err := mock.Ctx.GetSessionProvider()
	require.NoError(t, err)

	record := &model.OAuth2SessionID{SessionID: uuid.Must(uuid.NewRandom())}

	mock.StorageMock.EXPECT().
		GetOrCreateOAuth2SessionID(gomock.Eq(mock.Ctx), gomock.Eq(provider.GetIssuer()), gomock.Eq(client.GetSectorIdentifierURI()), gomock.Eq(userSession.PublicID)).
		Times(1).
		Return(record, nil)

	mock.StorageMock.EXPECT().
		SaveOAuth2SessionIDClient(gomock.Eq(mock.Ctx), gomock.Eq(provider.GetIssuer()), gomock.Eq(userSession.PublicID), gomock.Eq(record.SessionID.String()), gomock.Eq("client-id")).
		Times(1).
		Return(nil)

	sid, err := oidcSessionID(mock.Ctx, client, requester, &userSession)

	require.NoError(t, err)
	assert.Equal(t, record.SessionID.String(), sid)
	assert.NotEmpty(t, sid)
}

func TestOIDCSessionIDShouldPropagateParticipationStorageError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{Username: "john", PublicID: "public-id"})
	defer mock.Close()

	client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{ID: "client-id"}, &schema.IdentityProvidersOpenIDConnect{}, nil)

	requester := oauthelia2.NewRequest()
	requester.GrantScope(oidc.ScopeOpenID)

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	expected := errors.New("storage failure")

	mock.StorageMock.EXPECT().
		GetOrCreateOAuth2SessionID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1).
		Return(&model.OAuth2SessionID{SessionID: uuid.Must(uuid.NewRandom())}, nil)

	mock.StorageMock.EXPECT().
		SaveOAuth2SessionIDClient(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1).
		Return(expected)

	sid, err := oidcSessionID(mock.Ctx, client, requester, &userSession)

	assert.ErrorIs(t, err, expected)
	assert.Empty(t, sid)
}

func TestOIDCSessionIDShouldBeEmptyForAnonymousSession(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{ID: "client-id"}, &schema.IdentityProvidersOpenIDConnect{}, nil)

	requester := oauthelia2.NewRequest()
	requester.GrantScope(oidc.ScopeOpenID)

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	require.Empty(t, userSession.PublicID)

	sid, err := oidcSessionID(mock.Ctx, client, requester, &userSession)

	assert.NoError(t, err)
	assert.Empty(t, sid)
}

func TestOIDCSessionIDShouldPropagateStorageError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{Username: "john", PublicID: "public-id"})
	defer mock.Close()

	client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{ID: "client-id"}, &schema.IdentityProvidersOpenIDConnect{}, nil)

	requester := oauthelia2.NewRequest()
	requester.GrantScope(oidc.ScopeOpenID)

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	expected := errors.New("storage failure")

	mock.StorageMock.EXPECT().
		GetOrCreateOAuth2SessionID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1).
		Return(nil, expected)

	sid, err := oidcSessionID(mock.Ctx, client, requester, &userSession)

	assert.ErrorIs(t, err, expected)
	assert.Empty(t, sid)
}
