package storage

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestSchemaMigrateBackfillsRefreshTokenAccessSignature(t *testing.T) {
	testCases := []struct {
		name      string
		signature string
		expected  string
	}{
		{"ShouldPairTheOnlyUnrevokedAccessToken", "rt_a2", "at_a2"},
		{"ShouldNotPairWhenAmbiguous", "rt_b1", ""},
		{"ShouldNotPairWhenNoUnrevokedAccessToken", "rt_c1", ""},
		{"ShouldNotPairARevokedRefreshToken", "rt_a1", ""},
	}

	config := &schema.Configuration{
		Storage: schema.Storage{
			EncryptionKey: "authelia-test-key-not-a-secret-authelia-test-key-not-a-secret",
			Local:         &schema.StorageLocal{Path: filepath.Join(t.TempDir(), "db.sqlite3")},
		},
	}

	ctx := context.Background()

	provider, err := NewSQLiteProvider(config)

	require.NoError(t, err)

	t.Cleanup(func() {
		_ = provider.Close()
	})

	require.NoError(t, provider.SchemaMigrate(ctx, true, 26))

	seedAccess := `INSERT INTO oauth2_access_token_session (challenge_id, request_id, client_id, signature, subject, requested_scopes, granted_scopes, active, revoked, form_data, session_data) VALUES (?, ?, 'app', ?, 's', '', '', TRUE, ?, '', '');`
	seedRefresh := `INSERT INTO oauth2_refresh_token_session (challenge_id, request_id, client_id, signature, subject, requested_scopes, granted_scopes, active, revoked, form_data, session_data) VALUES (?, ?, 'app', ?, 's', '', '', TRUE, ?, '', '');`

	for _, seed := range []struct {
		query     string
		challenge string
		requestID string
		signature string
		revoked   bool
	}{
		{seedAccess, "c", "a", "at_a1", true},
		{seedRefresh, "c", "a", "rt_a1", true},
		{seedAccess, "c", "a", "at_a2", false},
		{seedRefresh, "c", "a", "rt_a2", false},
		{seedAccess, "c", "b", "at_b1", false},
		{seedAccess, "c", "b", "at_b2", false},
		{seedRefresh, "c", "b", "rt_b1", false},
		{seedAccess, "c", "c", "at_c1", true},
		{seedRefresh, "c", "c", "rt_c1", false},
	} {
		_, err = provider.db.ExecContext(ctx, seed.query, seed.challenge, seed.requestID, seed.signature, seed.revoked)

		require.NoError(t, err)
	}

	require.NoError(t, provider.SchemaMigrate(ctx, true, 27))

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var actual string

			require.NoError(t, provider.db.GetContext(ctx, &actual, "SELECT access_signature FROM oauth2_refresh_token_session WHERE signature = ?;", tc.signature))

			assert.Equal(t, tc.expected, actual)
		})
	}
}
