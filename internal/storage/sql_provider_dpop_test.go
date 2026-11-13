// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"database/sql"
	"path"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestSQLProviderCheckAndSetOAuth2DPoPProofUsed(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	now := time.Unix(1700000000, 0).UTC()

	htu := "https://auth.example.com/api/oidc/token"
	jkt := "0ZcOCORZNYy-DWpqq30jZyJGHTN0d2HglBV3uiguA4I"

	t.Run("ShouldNotBeUsedOnFirstUse", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "POST", htu, now.Add(time.Minute)), now)

		require.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldBeUsedOnReplay", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "POST", htu, now.Add(time.Minute)), now)

		require.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("ShouldBeUsedOnReplayWithinWindow", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "POST", htu, now.Add(2*time.Minute)), now.Add(59*time.Second))

		require.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("ShouldNotBeUsedAfterExp", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "POST", htu, now.Add(3*time.Minute)), now.Add(2*time.Minute))

		require.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldBeUsedOnReplayAfterRefresh", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "POST", htu, now.Add(4*time.Minute)), now.Add(2*time.Minute))

		require.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("ShouldNotBeUsedForDistinctJTI", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("other", jkt, "POST", htu, now.Add(time.Minute)), now)

		require.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldNotBeUsedForSameJTIAtDistinctHTU", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "POST", "https://auth.example.com/api/oidc/introspection", now.Add(time.Minute)), now)

		require.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldNotBeUsedForSameJTIWithDistinctJKT", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", "other", "POST", htu, now.Add(time.Minute)), now)

		require.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldHandleLongHTU", func(t *testing.T) {
		long := htu + "/" + strings.Repeat("a", 4096)

		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "GET", long, now.Add(time.Minute)), now)

		require.NoError(t, err)
		assert.False(t, used)

		used, err = provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "GET", long, now.Add(time.Minute)), now)

		require.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("ShouldNotBeUsedForSameJTIWithDistinctHTM", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "GET", htu, now.Add(time.Minute)), now)

		require.NoError(t, err)
		assert.False(t, used)
	})
}

func TestSQLProviderDeleteExpiredOAuth2DPoPProofs(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	now := time.Unix(1700000000, 0).UTC()

	htu := "https://auth.example.com/api/oidc/token"
	jkt := "0ZcOCORZNYy-DWpqq30jZyJGHTN0d2HglBV3uiguA4I"

	t.Run("ShouldNotErrorWithoutRows", func(t *testing.T) {
		require.NoError(t, provider.DeleteExpiredOAuth2DPoPProofs(ctx, now))
	})

	t.Run("ShouldRetainUnexpiredRecords", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "POST", htu, now.Add(time.Minute)), now)

		require.NoError(t, err)
		require.False(t, used)

		require.NoError(t, provider.DeleteExpiredOAuth2DPoPProofs(ctx, now))

		used, err = provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "POST", htu, now.Add(time.Minute)), now)

		require.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("ShouldRemoveExpiredRecords", func(t *testing.T) {
		require.NoError(t, provider.DeleteExpiredOAuth2DPoPProofs(ctx, now.Add(time.Minute)))

		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, model.NewOAuth2DPoPProof("jti", jkt, "POST", htu, now.Add(time.Minute)), now)

		require.NoError(t, err)
		assert.False(t, used)
	})
}

func TestSQLProviderDeleteExpiredOAuth2DPoPNonces(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	exp := time.Unix(1700000000, 0).UTC()

	t.Run("ShouldNotErrorWithoutRows", func(t *testing.T) {
		require.NoError(t, provider.DeleteExpiredOAuth2DPoPNonces(ctx, exp))
	})

	t.Run("ShouldRetainUnexpiredNonces", func(t *testing.T) {
		require.NoError(t, provider.SaveOAuth2DPoPNonce(ctx, model.OAuth2DPoPNonce{Signature: "abc", ExpiresAt: exp}))
		require.NoError(t, provider.DeleteExpiredOAuth2DPoPNonces(ctx, exp.Add(-time.Second)))

		nonce, err := provider.LoadOAuth2DPoPNonce(ctx, "abc")

		require.NoError(t, err)
		require.NotNil(t, nonce)
	})

	t.Run("ShouldRemoveExpiredNonces", func(t *testing.T) {
		require.NoError(t, provider.DeleteExpiredOAuth2DPoPNonces(ctx, exp))

		nonce, err := provider.LoadOAuth2DPoPNonce(ctx, "abc")

		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, nonce)
	})
}

func TestOAuth2DPoPProofSchemaSignatureColumn(t *testing.T) {
	testCases := []struct {
		Name     string
		Provider string
	}{
		{"MySQL", providerMySQL},
		{"PostgreSQL", providerPostgres},
		{"SQLite", providerSQLite},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			data, err := migrationsFS.ReadFile(path.Join(pathMigrations, tc.Provider, migrationFileOAuth2DPoP))

			require.NoError(t, err)

			match := reOAuth2DPoPProofSignatureColumn.FindStringSubmatch(string(data))

			require.Len(t, match, 2)

			width, err := strconv.Atoi(match[1])

			require.NoError(t, err)

			assert.Equal(t, len(model.NewOAuth2DPoPProofSignature("jti", "jkt", "POST", "https://auth.example.com")), width)
		})
	}
}

func TestSQLProviderOAuth2DPoPNonce(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	exp := time.Unix(1700000000, 0).UTC()

	t.Run("ShouldReturnErrNoRowsForUnknownNonce", func(t *testing.T) {
		nonce, err := provider.LoadOAuth2DPoPNonce(ctx, "abc")

		assert.ErrorIs(t, err, sql.ErrNoRows)
		assert.Nil(t, nonce)
	})

	t.Run("ShouldSaveAndLoadNonce", func(t *testing.T) {
		require.NoError(t, provider.SaveOAuth2DPoPNonce(ctx, model.OAuth2DPoPNonce{Signature: "abc", ExpiresAt: exp}))

		nonce, err := provider.LoadOAuth2DPoPNonce(ctx, "abc")

		require.NoError(t, err)
		require.NotNil(t, nonce)
		assert.Equal(t, "abc", nonce.Signature)
		assert.Equal(t, exp, nonce.ExpiresAt.UTC())
	})

	t.Run("ShouldNotAllowDuplicateNonce", func(t *testing.T) {
		assert.EqualError(t, provider.SaveOAuth2DPoPNonce(ctx, model.OAuth2DPoPNonce{Signature: "abc", ExpiresAt: exp}), "error inserting oauth2 dpop nonce with signature 'abc': UNIQUE constraint failed: oauth2_dpop_nonce.signature")
	})
}

const (
	migrationFileOAuth2DPoP = "V0033.OAuth2DPoP.up.sql"
)

var (
	reOAuth2DPoPProofSignatureColumn = regexp.MustCompile(`(?s)CREATE TABLE IF NOT EXISTS oauth2_dpop_proof \(.*?\bsignature VARCHAR\((\d+)\) NOT NULL,`)
)
