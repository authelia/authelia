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
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestOIDCSessionID(t *testing.T) {
	failure := errors.New("storage failure")

	testCases := []struct {
		name        string
		scopes      []string
		participant bool
		subject     error
		have        error
		expected    error
	}{
		{name: "ShouldRecordTheParticipantWithTheOpenIDScope", scopes: []string{oidc.ScopeOpenID}, participant: true},
		{name: "ShouldNotRecordTheParticipantWithoutTheOpenIDScope", scopes: []string{oidc.ScopeProfile}},
		{name: "ShouldRecordTheParticipantWithoutSubjectWhenItCantBeDetermined", scopes: []string{oidc.ScopeOpenID}, participant: true, subject: errors.New("connection refused")},
		{name: "ShouldPropagateParticipationStorageError", scopes: []string{oidc.ScopeOpenID}, participant: true, have: failure, expected: failure},
		{name: "ShouldReportEndedSession", scopes: []string{oidc.ScopeOpenID}, participant: true, have: storage.ErrOAuth2SessionEnded, expected: errOIDCSessionEnded},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{Username: "john", PublicID: "public-id"})
			defer mock.Close()

			setupTestOIDCProvider(t, mock, newTestOIDCConfig(t))

			client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{
				ID:                  "client-id",
				SectorIdentifierURI: &url.URL{Scheme: "https", Host: "sector.example.com"},
			}, &schema.IdentityProvidersOpenIDConnect{}, nil)

			subject := ""

			if tc.subject == nil {
				setupTestOIDCSubjectStore(t, mock)

				identifier, err := mock.Ctx.Providers.OpenIDConnect.GetSubject(mock.Ctx, client.GetSectorIdentifierURI(), "john")
				require.NoError(t, err)

				subject = identifier.String()
			} else {
				mock.StorageMock.EXPECT().
					LoadUserOpaqueIdentifierBySignature(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					AnyTimes().
					Return(nil, tc.subject)
			}

			requester := oauthelia2.NewRequest()

			for _, scope := range tc.scopes {
				requester.GrantScope(scope)
			}

			userSession, err := mock.Ctx.GetSession()
			require.NoError(t, err)

			provider, err := mock.Ctx.GetSessionProvider()
			require.NoError(t, err)

			record := &model.OAuth2SessionID{SessionID: uuid.Must(uuid.NewRandom())}

			mock.StorageMock.EXPECT().
				GetOrCreateOAuth2SessionID(gomock.Eq(mock.Ctx), gomock.Eq(provider.GetIssuer()), gomock.Eq(client.GetSectorIdentifierURI()), gomock.Eq(userSession.PublicID)).
				Times(1).
				Return(record, nil)

			if tc.participant {
				mock.StorageMock.EXPECT().
					SaveOAuth2SessionIDClient(gomock.Eq(mock.Ctx), gomock.Eq(provider.GetIssuer()), gomock.Eq(userSession.PublicID), gomock.Eq(record.SessionID.String()), gomock.Eq("client-id"), gomock.Eq(subject)).
					Times(1).
					Return(tc.have)
			}

			sid, err := oidcSessionID(mock.Ctx, client, requester, &userSession)

			if tc.expected == nil {
				require.NoError(t, err)
				assert.Equal(t, record.SessionID.String(), sid)
			} else {
				assert.ErrorIs(t, err, tc.expected)
				assert.Empty(t, sid)
			}
		})
	}
}

func TestOIDCSessionIDShouldBeEmptyForAnonymousSession(t *testing.T) {
	mock := mocks.NewMockAutheliaCtx(t)
	defer mock.Close()

	client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{ID: "client-id"}, &schema.IdentityProvidersOpenIDConnect{}, nil)

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	require.Empty(t, userSession.PublicID)

	sid, err := oidcSessionID(mock.Ctx, client, oauthelia2.NewRequest(), &userSession)

	assert.NoError(t, err)
	assert.Empty(t, sid)
}

func TestOIDCSessionIDShouldPropagateStorageError(t *testing.T) {
	mock := mocks.NewMockAutheliaCtxWithUserSession(t, session.UserSession{Username: "john", PublicID: "public-id"})
	defer mock.Close()

	client := oidc.NewClient(schema.IdentityProvidersOpenIDConnectClient{ID: "client-id"}, &schema.IdentityProvidersOpenIDConnect{}, nil)

	userSession, err := mock.Ctx.GetSession()
	require.NoError(t, err)

	expected := errors.New("storage failure")

	mock.StorageMock.EXPECT().
		GetOrCreateOAuth2SessionID(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1).
		Return(nil, expected)

	sid, err := oidcSessionID(mock.Ctx, client, oauthelia2.NewRequest(), &userSession)

	assert.ErrorIs(t, err, expected)
	assert.Empty(t, sid)
}
