package storage

import (
	"context"
	"database/sql"
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

	t.Run("ShouldNotBeUsedOnFirstUse", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "jti", "POST", htu, now.Add(time.Minute), now)

		require.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldBeUsedOnReplay", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "jti", "POST", htu, now.Add(time.Minute), now)

		require.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("ShouldBeUsedOnReplayWithinWindow", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "jti", "POST", htu, now.Add(2*time.Minute), now.Add(59*time.Second))

		require.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("ShouldNotBeUsedAfterExp", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "jti", "POST", htu, now.Add(3*time.Minute), now.Add(2*time.Minute))

		require.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldBeUsedOnReplayAfterRefresh", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "jti", "POST", htu, now.Add(4*time.Minute), now.Add(2*time.Minute))

		require.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("ShouldNotBeUsedForDistinctJTI", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "other", "POST", htu, now.Add(time.Minute), now)

		require.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldNotBeUsedForSameJTIAtDistinctHTU", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "jti", "POST", "https://auth.example.com/api/oidc/introspection", now.Add(time.Minute), now)

		require.NoError(t, err)
		assert.False(t, used)
	})

	t.Run("ShouldNotBeUsedForSameJTIWithDistinctHTM", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "jti", "GET", htu, now.Add(time.Minute), now)

		require.NoError(t, err)
		assert.False(t, used)
	})
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
