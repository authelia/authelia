// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestSQLProviderOAuth2LogoutSession(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	newLogoutSession := func(expires time.Time) model.OAuth2LogoutSession {
		return model.OAuth2LogoutSession{
			FlowID:      uuid.Must(uuid.NewRandom()),
			ClientID:    "client",
			Subject:     "subject",
			SessionID:   "sid",
			RedirectURI: "https://app.example.com/logged-out",
			State:       "state",
			CreatedAt:   now,
			ExpiresAt:   expires,
		}
	}

	t.Run("ShouldRoundTrip", func(t *testing.T) {
		logout := newLogoutSession(now.Add(time.Hour))

		require.NoError(t, provider.SaveOAuth2LogoutSession(ctx, logout))

		loaded, err := provider.LoadOAuth2LogoutSession(ctx, logout.FlowID)
		require.NoError(t, err)
		require.NotNil(t, loaded)

		assert.Equal(t, logout.FlowID, loaded.FlowID)
		assert.Equal(t, "client", loaded.ClientID)
		assert.Equal(t, "subject", loaded.Subject)
		assert.Equal(t, "sid", loaded.SessionID)
		assert.Equal(t, "https://app.example.com/logged-out", loaded.RedirectURI)
		assert.Equal(t, "state", loaded.State)
		assert.True(t, logout.ExpiresAt.Equal(loaded.ExpiresAt))
	})

	t.Run("ShouldReturnNilWhenAbsent", func(t *testing.T) {
		loaded, err := provider.LoadOAuth2LogoutSession(ctx, uuid.Must(uuid.NewRandom()))
		require.NoError(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("ShouldDelete", func(t *testing.T) {
		logout := newLogoutSession(now.Add(time.Hour))

		require.NoError(t, provider.SaveOAuth2LogoutSession(ctx, logout))
		require.NoError(t, provider.DeleteOAuth2LogoutSession(ctx, logout.FlowID))

		loaded, err := provider.LoadOAuth2LogoutSession(ctx, logout.FlowID)
		require.NoError(t, err)
		assert.Nil(t, loaded)

		require.NoError(t, provider.DeleteOAuth2LogoutSession(ctx, logout.FlowID))
	})

	t.Run("ShouldPurgeExpiredOnSave", func(t *testing.T) {
		expired := newLogoutSession(now.Add(-time.Minute))

		require.NoError(t, provider.SaveOAuth2LogoutSession(ctx, expired))

		current := newLogoutSession(now.Add(time.Hour))

		require.NoError(t, provider.SaveOAuth2LogoutSession(ctx, current))

		loaded, err := provider.LoadOAuth2LogoutSession(ctx, expired.FlowID)
		require.NoError(t, err)
		assert.Nil(t, loaded)

		loaded, err = provider.LoadOAuth2LogoutSession(ctx, current.FlowID)
		require.NoError(t, err)
		assert.NotNil(t, loaded)
	})
}
