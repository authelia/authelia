// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestSQLiteProviderStaleOAuth2ConsentSessionsShouldKeepLiveRefreshTokens(t *testing.T) {
	config := &schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local: &schema.StorageLocal{
				Path: filepath.Join(t.TempDir(), "db.sqlite3"),
			},
		},
	}

	provider, err := NewSQLiteProvider(config)
	require.NoError(t, err)
	require.NoError(t, provider.StartupCheck())

	defer provider.Close()

	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)
	subject := uuid.New()

	_, err = provider.db.ExecContext(ctx, `INSERT INTO user_opaque_identifier (service, sector_id, username, identifier) VALUES ('openid', '', 'john', ?);`, subject.String())
	require.NoError(t, err)

	consent := func(name string) uuid.UUID {
		challengeID := uuid.New()

		_, err := provider.db.ExecContext(ctx, `
			INSERT INTO oauth2_consent_session (challenge_id, client_id, subject, authorized, granted, requested_at, responded_at, expires_at, form_data, requested_scopes, granted_scopes)
			VALUES (?, ?, ?, TRUE, TRUE, ?, ?, ?, '', 'openid', 'openid');`,
			challengeID.String(), name, subject.String(), now.Add(-time.Hour*2), now.Add(-time.Hour*2), now.Add(-time.Hour))
		require.NoError(t, err)

		return challengeID
	}

	refresh := func(challengeID uuid.UUID, active, revoked bool) {
		_, err := provider.db.ExecContext(ctx, `
			INSERT INTO oauth2_refresh_token_session (challenge_id, request_id, client_id, signature, subject, requested_scopes, granted_scopes, active, revoked, form_data, session_data)
			VALUES (?, ?, 'client', ?, ?, 'openid', 'openid', ?, ?, '', ?);`,
			challengeID.String(), uuid.NewString(), uuid.NewString(), subject.String(), active, revoked, []byte("{}"))
		require.NoError(t, err)
	}

	consent("none")

	live := consent("live")
	refresh(live, false, false)
	refresh(live, true, false)

	rotated := consent("rotated")
	refresh(rotated, false, false)

	revoked := consent("revoked")
	refresh(revoked, true, true)

	count, err := provider.CountStaleOAuth2ConsentSessions(ctx, now)
	require.NoError(t, err)
	assert.Equal(t, 3, count)

	deleted, err := provider.DeleteStaleOAuth2ConsentSessions(ctx, now)
	require.NoError(t, err)
	assert.Equal(t, 3, deleted)

	var remaining []string

	require.NoError(t, provider.db.SelectContext(ctx, &remaining, `SELECT challenge_id FROM oauth2_consent_session;`))
	assert.Equal(t, []string{live.String()}, remaining)
}
