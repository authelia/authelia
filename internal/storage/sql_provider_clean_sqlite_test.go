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

func TestSQLiteProviderStaleOAuth2ConsentSessionsShouldKeepValidTokens(t *testing.T) {
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

	token := func(table string, challengeID uuid.UUID, issued time.Time, active, revoked bool) {
		_, err := provider.db.ExecContext(ctx, `
			INSERT INTO `+table+` (challenge_id, request_id, client_id, signature, subject, requested_at, requested_scopes, granted_scopes, active, revoked, form_data, session_data)
			VALUES (?, ?, 'client', ?, ?, ?, 'openid', 'openid', ?, ?, '', ?);`,
			challengeID.String(), uuid.NewString(), uuid.NewString(), subject.String(), issued, active, revoked, []byte("{}"))
		require.NoError(t, err)
	}

	accessIssuedAfter, refreshIssuedAfter := now.Add(-time.Hour), now.Add(-time.Hour*24)

	consent("none")

	refreshLive := consent("refresh-live")
	token(tableOAuth2RefreshTokenSession, refreshLive, now.Add(-time.Hour*25), false, false)
	token(tableOAuth2RefreshTokenSession, refreshLive, now.Add(-time.Hour), true, false)

	refreshExpired := consent("refresh-expired")
	token(tableOAuth2RefreshTokenSession, refreshExpired, now.Add(-time.Hour*48), true, false)

	refreshRotated := consent("refresh-rotated")
	token(tableOAuth2RefreshTokenSession, refreshRotated, now.Add(-time.Hour), false, false)

	refreshRevoked := consent("refresh-revoked")
	token(tableOAuth2RefreshTokenSession, refreshRevoked, now.Add(-time.Hour), true, true)

	accessLive := consent("access-live")
	token(tableOAuth2AccessTokenSession, accessLive, now.Add(-time.Minute*10), true, false)

	accessExpired := consent("access-expired")
	token(tableOAuth2AccessTokenSession, accessExpired, now.Add(-time.Hour*2), true, false)

	accessRevoked := consent("access-revoked")
	token(tableOAuth2AccessTokenSession, accessRevoked, now.Add(-time.Minute*10), true, true)

	count, err := provider.CountStaleOAuth2ConsentSessions(ctx, now, accessIssuedAfter, time.Time{})
	require.NoError(t, err)
	assert.Equal(t, 5, count, "a refresh token which never expires should keep its consent session however old it is")

	count, err = provider.CountStaleOAuth2ConsentSessions(ctx, now, accessIssuedAfter, refreshIssuedAfter)
	require.NoError(t, err)
	assert.Equal(t, 6, count)

	deleted, err := provider.DeleteStaleOAuth2ConsentSessions(ctx, now, accessIssuedAfter, refreshIssuedAfter)
	require.NoError(t, err)
	assert.Equal(t, 6, deleted)

	var remaining []string

	require.NoError(t, provider.db.SelectContext(ctx, &remaining, `SELECT challenge_id FROM oauth2_consent_session ORDER BY id;`))
	assert.Equal(t, []string{refreshLive.String(), accessLive.String()}, remaining)
}
