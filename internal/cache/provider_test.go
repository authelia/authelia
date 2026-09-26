// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/session"
)

func TestSessionRepositoryShouldDelegateToProvider(t *testing.T) {
	ctx := context.Background()

	repository := NewSessionRepository(NewMemory())

	require.NoError(t, repository.StartupCheck())

	assert.Equal(t, sessionGarbageCollectionFrequency, repository.GarbageCollectionFrequency(ctx))

	require.NoError(t, repository.Save(ctx, "example.com", "id", "pid", "john", time.Hour, []byte("first")))

	record, err := repository.Get(ctx, "example.com", "id")
	require.NoError(t, err)
	assert.Equal(t, session.NewRecord("id", []byte("first")), record)

	require.NoError(t, repository.SaveData(ctx, "example.com", "id", "pid", "john", time.Hour, []byte("second")))

	record, err = repository.GetByPublicID(ctx, "example.com", "pid")
	require.NoError(t, err)
	assert.Equal(t, session.NewRecord("id", []byte("second")), record)

	ids, err := repository.GetIDsByUsername(ctx, "example.com", "john")
	require.NoError(t, err)
	assert.Equal(t, []string{"id"}, ids)

	require.NoError(t, repository.ChangeID(ctx, "example.com", "id", "new", "pid", "john", time.Hour, []byte("resealed")))

	record, err = repository.Get(ctx, "example.com", "id")
	require.NoError(t, err)
	assert.Nil(t, record)

	record, err = repository.Get(ctx, "example.com", "new")
	require.NoError(t, err)
	assert.Equal(t, session.NewRecord("new", []byte("resealed")), record)

	require.NoError(t, repository.GarbageCollection(ctx))

	require.NoError(t, repository.Delete(ctx, "example.com", "new", "pid", "john"))

	record, err = repository.Get(ctx, "example.com", "new")
	require.NoError(t, err)
	assert.Nil(t, record)
}

func TestMemory_StartupCheckAndGarbageCollectionFrequency(t *testing.T) {
	provider := NewMemory()

	assert.NoError(t, provider.StartupCheck())
	assert.Equal(t, sessionGarbageCollectionFrequency, provider.SessionGarbageCollectionFrequency(context.Background()))
}

func TestMemory_SessionDeleteUnknownSession(t *testing.T) {
	testCases := []struct {
		Name     string
		Username string
		Existing bool
	}{
		{"ShouldIgnoreAnonymousSession", "", false},
		{"ShouldIgnoreUsernameWithoutLookup", "john", false},
		{"ShouldKeepOtherSessionsOfUsername", "john", true},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			provider := NewMemory()

			if tc.Existing {
				require.NoError(t, provider.SessionSave(ctx, "example.com", "other", "otherpid", tc.Username, time.Hour, []byte("data")))
			}

			require.NoError(t, provider.SessionDelete(ctx, "example.com", "unknown", "pid", tc.Username))

			ids, err := provider.SessionGetIDsByUsername(ctx, "example.com", tc.Username)
			require.NoError(t, err)

			if tc.Existing {
				assert.Equal(t, []string{"other"}, ids)
			} else {
				assert.Empty(t, ids)
			}
		})
	}
}
