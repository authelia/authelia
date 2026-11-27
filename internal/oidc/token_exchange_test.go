// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"net/url"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	oauthelia2 "authelia.com/provider/oauth2"
	"authelia.com/provider/oauth2/handler/openid"
	"authelia.com/provider/oauth2/token/jwt"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/oidc"
)

func TestHydrateTokenExchangeSessionWithRequestingClient(t *testing.T) {
	newSession := func() *oidc.Session {
		return &oidc.Session{
			ChallengeID: model.NullUUID(uuid.Must(uuid.NewRandom())),
			ClientID:    "subject-client",
			DefaultSession: &openid.DefaultSession{
				Claims: &jwt.IDTokenClaims{
					Subject:                         "abc",
					AuthorizedParty:                 "subject-client",
					Audience:                        []string{"subject-client"},
					Nonce:                           "abcdefghijklmnopqrstuvwxyz",
					ExpirationTime:                  jwt.NewNumericDate(time.Unix(1000, 0)),
					AccessTokenHash:                 "at-hash",
					CodeHash:                        "c-hash",
					StateHash:                       "s-hash",
					AuthenticationMethodsReferences: []string{"pwd"},
				},
			},
		}
	}

	t.Run("ShouldReattributeSessionToRequestingClient", func(t *testing.T) {
		session := newSession()

		oidc.HydrateTokenExchangeSessionWithRequestingClient(&oidc.RegisteredClient{ID: "exchange-client"}, session)

		assert.Equal(t, "exchange-client", session.ClientID)
		assert.False(t, session.ChallengeID.Valid)
		assert.Equal(t, "exchange-client", session.Claims.AuthorizedParty)
		assert.Nil(t, session.Claims.Audience)
		assert.Empty(t, session.Claims.Nonce)
		assert.Nil(t, session.Claims.ExpirationTime)
		assert.Empty(t, session.Claims.AccessTokenHash)
		assert.Empty(t, session.Claims.CodeHash)
		assert.Empty(t, session.Claims.StateHash)

		// The subject and the claims describing their authentication are what the exchange carries across.
		assert.Equal(t, "abc", session.Claims.Subject)
		assert.Equal(t, []string{"pwd"}, session.Claims.AuthenticationMethodsReferences)
	})

	t.Run("ShouldHandleNilValues", func(t *testing.T) {
		assert.NotPanics(t, func() {
			oidc.HydrateTokenExchangeSessionWithRequestingClient(nil, newSession())
			oidc.HydrateTokenExchangeSessionWithRequestingClient(&oidc.RegisteredClient{ID: "exchange-client"}, nil)
			oidc.HydrateTokenExchangeSessionWithRequestingClient(&oidc.RegisteredClient{ID: "exchange-client"}, &oidc.Session{})
			oidc.HydrateTokenExchangeSessionWithRequestingClient(&oidc.RegisteredClient{ID: "exchange-client"}, &oidc.Session{DefaultSession: &openid.DefaultSession{}})
		})

		session := &oidc.Session{DefaultSession: &openid.DefaultSession{}}

		oidc.HydrateTokenExchangeSessionWithRequestingClient(&oidc.RegisteredClient{ID: "exchange-client"}, session)

		assert.Equal(t, "exchange-client", session.ClientID)
	})

	t.Run("ShouldCarrySubjectIdentityFromSubjectRequester", func(t *testing.T) {
		session := newSession()

		session.SetSubjectRequester(&oauthelia2.Request{Session: &oidc.Session{
			ClientCredentials: true,
			DefaultSession:    &openid.DefaultSession{Username: "john"},
		}})

		oidc.HydrateTokenExchangeSessionWithRequestingClient(&oidc.RegisteredClient{ID: "exchange-client"}, session)

		assert.Equal(t, "john", session.Username)
		assert.True(t, session.ClientCredentials)
	})

	t.Run("ShouldClearSubjectOfClientCredentialsSubjectRequester", func(t *testing.T) {
		session := newSession()
		session.Subject = "exchange-client"

		session.SetSubjectRequester(&oauthelia2.Request{Session: &oidc.Session{
			ClientCredentials: true,
			DefaultSession:    &openid.DefaultSession{},
		}})

		oidc.HydrateTokenExchangeSessionWithRequestingClient(&oidc.RegisteredClient{ID: "exchange-client"}, session)

		assert.Empty(t, session.Subject)
		assert.Equal(t, "exchange-client", session.GetSubject())
	})

	t.Run("ShouldKeepSubjectOfUserSubjectRequester", func(t *testing.T) {
		session := newSession()
		session.Subject = "user-subject"

		session.SetSubjectRequester(&oauthelia2.Request{Session: &oidc.Session{
			DefaultSession: &openid.DefaultSession{Username: "john"},
		}})

		oidc.HydrateTokenExchangeSessionWithRequestingClient(&oidc.RegisteredClient{ID: "exchange-client"}, session)

		assert.Equal(t, "user-subject", session.Subject)
	})

	t.Run("ShouldNotCarryIdentityWithoutSubjectRequester", func(t *testing.T) {
		session := newSession()

		oidc.HydrateTokenExchangeSessionWithRequestingClient(&oidc.RegisteredClient{ID: "exchange-client"}, session)

		assert.Empty(t, session.Username)
		assert.False(t, session.ClientCredentials)

		clientID, stored := oidc.TokenExchangeSubjectClientID(session)

		assert.Empty(t, clientID)
		assert.False(t, stored)
		assert.False(t, oidc.TokenExchangeSubjectClientCredentials(session))
	})

	t.Run("ShouldReportSubjectRequesterDetails", func(t *testing.T) {
		session := newSession()

		session.SetSubjectRequester(&oauthelia2.Request{
			Client: &oidc.RegisteredClient{ID: "subject-client"},
			Session: &oidc.Session{
				ClientCredentials: true,
				DefaultSession:    &openid.DefaultSession{Username: "john"},
			},
		})

		clientID, stored := oidc.TokenExchangeSubjectClientID(session)

		assert.Equal(t, "subject-client", clientID)
		assert.True(t, stored)
		assert.True(t, oidc.TokenExchangeSubjectClientCredentials(session))
	})

	t.Run("ShouldStripMayActAndKeepAct", func(t *testing.T) {
		session := newSession()

		session.Extra = map[string]any{oidc.ClaimAuthorizedActor: map[string]any{"sub": "x"}, oidc.ClaimActor: map[string]any{"sub": "y"}}
		session.Claims.Extra = map[string]any{oidc.ClaimAuthorizedActor: map[string]any{"sub": "x"}, oidc.ClaimActor: map[string]any{"sub": "y"}}

		oidc.HydrateTokenExchangeSessionWithRequestingClient(&oidc.RegisteredClient{ID: "exchange-client"}, session)

		assert.NotContains(t, session.Extra, oidc.ClaimAuthorizedActor)
		assert.NotContains(t, session.Claims.Extra, oidc.ClaimAuthorizedActor)
		assert.Contains(t, session.Extra, oidc.ClaimActor)
		assert.Contains(t, session.Claims.Extra, oidc.ClaimActor)
	})
}

func TestIsTokenExchangeTxnTokenReplacement(t *testing.T) {
	testCases := []struct {
		name      string
		requester oauthelia2.Requester
		expected  bool
	}{
		{"ShouldHandleNil", nil, false},
		{"ShouldHandleNilForm", &oauthelia2.Request{}, false},
		{"ShouldDetectTxnToken", &oauthelia2.Request{Form: url.Values{oidc.FormParameterSubjectTokenType: []string{oidc.TokenTypeTxnToken}, oidc.FormParameterRequestedTokenType: []string{oidc.TokenTypeTxnToken}}}, true},
		{"ShouldNotDetectTxnTokenForOtherRequestedType", &oauthelia2.Request{Form: url.Values{oidc.FormParameterSubjectTokenType: []string{oidc.TokenTypeTxnToken}, oidc.FormParameterRequestedTokenType: []string{oidc.TokenTypeAccessToken}}}, false},
		{"ShouldNotDetectAccessToken", &oauthelia2.Request{Form: url.Values{oidc.FormParameterSubjectTokenType: []string{oidc.TokenTypeAccessToken}}}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, oidc.IsTokenExchangeTxnTokenReplacement(tc.requester))
		})
	}
}
