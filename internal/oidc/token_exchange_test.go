// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

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
}
