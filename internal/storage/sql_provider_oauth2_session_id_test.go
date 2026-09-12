// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/model"
)

func TestSQLProviderOAuth2SessionID(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	t.Run("ShouldBeIdempotent", func(t *testing.T) {
		first, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "", "public-id")
		require.NoError(t, err)
		require.NotNil(t, first)

		second, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "", "public-id")
		require.NoError(t, err)
		require.NotNil(t, second)

		assert.Equal(t, first.SessionID, second.SessionID)
		assert.Equal(t, uuid.Version(4), first.SessionID.Version())
	})

	t.Run("ShouldSeparateSectors", func(t *testing.T) {
		a, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://a.example.com", "sector-public-id")
		require.NoError(t, err)

		b, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://b.example.com", "sector-public-id")
		require.NoError(t, err)

		assert.NotEqual(t, a.SessionID, b.SessionID)
	})

	t.Run("ShouldRoundTripBySessionID", func(t *testing.T) {
		created, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "", "roundtrip-public-id")
		require.NoError(t, err)

		loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer", created.SessionID.String())
		require.NoError(t, err)
		require.NotNil(t, loaded)

		assert.Equal(t, "roundtrip-public-id", loaded.PublicID)
	})

	t.Run("ShouldReturnNilWhenAbsent", func(t *testing.T) {
		loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer", uuid.Must(uuid.NewRandom()).String())

		assert.NoError(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("ShouldSeparateIssuers", func(t *testing.T) {
		a, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer-a", "", "shared-public-id")
		require.NoError(t, err)

		loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer-b", a.SessionID.String())

		assert.NoError(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("ShouldDeleteEverySectorByPublicID", func(t *testing.T) {
		a, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://a.example.com", "delete-public-id")
		require.NoError(t, err)

		b, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://b.example.com", "delete-public-id")
		require.NoError(t, err)

		require.NoError(t, provider.DeleteOAuth2SessionIDByPublicID(ctx, "issuer", "delete-public-id"))

		for _, record := range []*model.OAuth2SessionID{a, b} {
			loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer", record.SessionID.String())
			require.NoError(t, err)
			assert.Nil(t, loaded)
		}
	})

	t.Run("ShouldDeleteOnlyMatchingIssuer", func(t *testing.T) {
		created, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer-delete-single", "", "delete-single-public-id")
		require.NoError(t, err)

		require.NoError(t, provider.DeleteOAuth2SessionID(ctx, "issuer-delete-single-other", created.SessionID.String()))

		loaded, err := provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer-delete-single", created.SessionID.String())
		require.NoError(t, err)
		require.NotNil(t, loaded)

		require.NoError(t, provider.DeleteOAuth2SessionID(ctx, "issuer-delete-single", created.SessionID.String()))

		loaded, err = provider.LoadOAuth2SessionIDBySessionID(ctx, "issuer-delete-single", created.SessionID.String())
		require.NoError(t, err)
		assert.Nil(t, loaded)
	})

	t.Run("ShouldPageOldestByAfterAndLimit", func(t *testing.T) {
		paging := newTestSQLiteProvider(t)
		require.NoError(t, paging.StartupCheck())

		first, err := paging.GetOrCreateOAuth2SessionID(ctx, "issuer-paging", "", "paging-public-id-1")
		require.NoError(t, err)

		_, err = paging.GetOrCreateOAuth2SessionID(ctx, "issuer-paging", "", "paging-public-id-2")
		require.NoError(t, err)

		_, err = paging.GetOrCreateOAuth2SessionID(ctx, "issuer-paging", "", "paging-public-id-3")
		require.NoError(t, err)

		// The creation path doesn't populate ID, so read it back to learn the real database id.
		firstLoaded, err := paging.LoadOAuth2SessionIDBySessionID(ctx, "issuer-paging", first.SessionID.String())
		require.NoError(t, err)
		require.NotNil(t, firstLoaded)

		page, err := paging.LoadOAuth2SessionIDsOldest(ctx, 0, 1)
		require.NoError(t, err)
		require.Len(t, page, 1)
		assert.Equal(t, firstLoaded.ID, page[0].ID)
		assert.Equal(t, "paging-public-id-1", page[0].PublicID)

		next, err := paging.LoadOAuth2SessionIDsOldest(ctx, page[0].ID, 1)
		require.NoError(t, err)
		require.Len(t, next, 1)
		assert.NotEqual(t, page[0].ID, next[0].ID)
		assert.Equal(t, "paging-public-id-2", next[0].PublicID)
	})
}

func TestSQLProviderOAuth2SessionIDClient(t *testing.T) {
	provider := newTestSQLiteProvider(t)
	require.NoError(t, provider.StartupCheck())

	ctx := context.Background()

	clientIDs := func(records []model.OAuth2SessionIDClient) (ids []string) {
		for _, record := range records {
			ids = append(ids, record.ClientID)
		}

		return ids
	}

	t.Run("ShouldRecordEachClientOnce", func(t *testing.T) {
		record, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "", "clients-public-id")
		require.NoError(t, err)

		sid := record.SessionID.String()

		require.NoError(t, provider.SaveOAuth2SessionIDClient(ctx, "issuer", "clients-public-id", sid, "client-a"))
		require.NoError(t, provider.SaveOAuth2SessionIDClient(ctx, "issuer", "clients-public-id", sid, "client-b"))
		require.NoError(t, provider.SaveOAuth2SessionIDClient(ctx, "issuer", "clients-public-id", sid, "client-a"))

		records, err := provider.LoadOAuth2SessionIDClientsByPublicID(ctx, "issuer", "clients-public-id")
		require.NoError(t, err)

		assert.Equal(t, []string{"client-a", "client-b"}, clientIDs(records))

		for _, r := range records {
			assert.Equal(t, record.SessionID, r.SessionID)
			assert.Equal(t, "issuer", r.Issuer)
			assert.Equal(t, "clients-public-id", r.PublicID)
		}
	})

	t.Run("ShouldRemoveClientsWithSessionID", func(t *testing.T) {
		a, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://a.example.com", "delete-public-id")
		require.NoError(t, err)

		b, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "https://b.example.com", "delete-public-id")
		require.NoError(t, err)

		require.NoError(t, provider.SaveOAuth2SessionIDClient(ctx, "issuer", "delete-public-id", a.SessionID.String(), "client-a"))
		require.NoError(t, provider.SaveOAuth2SessionIDClient(ctx, "issuer", "delete-public-id", b.SessionID.String(), "client-b"))

		require.NoError(t, provider.DeleteOAuth2SessionID(ctx, "issuer", a.SessionID.String()))

		records, err := provider.LoadOAuth2SessionIDClientsByPublicID(ctx, "issuer", "delete-public-id")
		require.NoError(t, err)

		assert.Equal(t, []string{"client-b"}, clientIDs(records))
	})

	t.Run("ShouldRemoveClientsWithPublicID", func(t *testing.T) {
		record, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "", "logout-public-id")
		require.NoError(t, err)

		other, err := provider.GetOrCreateOAuth2SessionID(ctx, "issuer", "", "other-public-id")
		require.NoError(t, err)

		require.NoError(t, provider.SaveOAuth2SessionIDClient(ctx, "issuer", "logout-public-id", record.SessionID.String(), "client-a"))
		require.NoError(t, provider.SaveOAuth2SessionIDClient(ctx, "issuer", "other-public-id", other.SessionID.String(), "client-a"))

		require.NoError(t, provider.DeleteOAuth2SessionIDByPublicID(ctx, "issuer", "logout-public-id"))

		records, err := provider.LoadOAuth2SessionIDClientsByPublicID(ctx, "issuer", "logout-public-id")
		require.NoError(t, err)
		assert.Empty(t, records)

		records, err = provider.LoadOAuth2SessionIDClientsByPublicID(ctx, "issuer", "other-public-id")
		require.NoError(t, err)
		assert.Equal(t, []string{"client-a"}, clientIDs(records))
	})
}
