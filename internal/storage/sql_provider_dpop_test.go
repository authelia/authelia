package storage

import (
	"context"
	"database/sql"
	"path"
	"regexp"
	"strconv"
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

func TestSQLProviderDeleteExpiredOAuth2DPoPProofs(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	now := time.Unix(1700000000, 0).UTC()

	htu := "https://auth.example.com/api/oidc/token"

	t.Run("ShouldNotErrorWithoutRows", func(t *testing.T) {
		require.NoError(t, provider.DeleteExpiredOAuth2DPoPProofs(ctx, now))
	})

	t.Run("ShouldRetainUnexpiredRecords", func(t *testing.T) {
		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "jti", "POST", htu, now.Add(time.Minute), now)

		require.NoError(t, err)
		require.False(t, used)

		require.NoError(t, provider.DeleteExpiredOAuth2DPoPProofs(ctx, now))

		used, err = provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "jti", "POST", htu, now.Add(time.Minute), now)

		require.NoError(t, err)
		assert.True(t, used)
	})

	t.Run("ShouldRemoveExpiredRecords", func(t *testing.T) {
		require.NoError(t, provider.DeleteExpiredOAuth2DPoPProofs(ctx, now.Add(time.Minute)))

		used, err := provider.CheckAndSetOAuth2DPoPProofUsed(ctx, "jti", "POST", htu, now.Add(time.Minute), now)

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

func TestOAuth2DPoPProofSchemaColumnWidths(t *testing.T) {
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

			widths := map[string]int{}

			for _, match := range reOAuth2DPoPProofColumn.FindAllStringSubmatch(string(data), -1) {
				width, err := strconv.Atoi(match[2])

				require.NoError(t, err)

				widths[match[1]] = width
			}

			total := 0

			for _, column := range columnsOAuth2DPoPProofUniqueIndex {
				width, ok := widths[column]

				require.True(t, ok, column)

				total += width
			}

			assert.GreaterOrEqual(t, widths["htu"], 400)
			assert.LessOrEqual(t, total*bytesPerCharUTF8MB4, maxIndexKeyBytesInnoDB)
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

// migrationFileOAuth2DPoP is the file name of the migration which creates the DPoP tables in every dialect.
const migrationFileOAuth2DPoP = "V0029.OAuth2DPoP.up.sql"

// bytesPerCharUTF8MB4 is the worst case storage of a single character in the utf8mb4 character set the MySQL dialect
// declares its tables with, and maxIndexKeyBytesInnoDB is the InnoDB limit on the total length of an index key with the
// DYNAMIC row format. Together they bound the sum of the declared widths of the columns which participate in the
// oauth2_dpop_proof unique index, exceeding which makes the CREATE UNIQUE INDEX statement itself fail.
const (
	bytesPerCharUTF8MB4    = 4
	maxIndexKeyBytesInnoDB = 3072
)

var (
	columnsOAuth2DPoPProofUniqueIndex = []string{"jti", "htm", "htu"}

	reOAuth2DPoPProofColumn = regexp.MustCompile(`(?m)^\s+(jti|htm|htu) VARCHAR\((\d+)\) NOT NULL,$`)
)
