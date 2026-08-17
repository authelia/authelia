package cache

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/session"
)

func TestMemory_SessionGetExpiry(t *testing.T) {
	testCases := []struct {
		Name     string
		Expire   bool
		Expected session.Record
	}{
		{"ShouldReturnRecordWhenUnexpired", false, session.NewRecord("id", []byte("data"))},
		{"ShouldReturnNoRecordWhenExpired", true, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			provider := NewMemory()

			require.NoError(t, provider.SessionSave(ctx, "example.com", "id", "pid", "john", time.Hour, []byte("data")))

			if tc.Expire {
				provider.session[provider.key("example.com", "id")].expires = time.Now().Add(-time.Second)
			}

			record, err := provider.SessionGet(ctx, "example.com", "id")
			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, record)

			record, err = provider.SessionGetByPublicID(ctx, "example.com", "pid")
			assert.NoError(t, err)
			assert.Equal(t, tc.Expected, record)

			ids, err := provider.SessionGetIDsByUsername(ctx, "example.com", "john")
			assert.NoError(t, err)

			if tc.Expire {
				assert.Empty(t, ids)
			} else {
				assert.Equal(t, []string{"id"}, ids)
			}
		})
	}
}

func TestMemory_SessionNeverExpiresWithZeroExpiration(t *testing.T) {
	ctx := context.Background()
	provider := NewMemory()

	require.NoError(t, provider.SessionSave(ctx, "example.com", "id", "pid", "john", 0, []byte("data")))

	assert.True(t, provider.session[provider.key("example.com", "id")].expires.IsZero())

	record, err := provider.SessionGet(ctx, "example.com", "id")
	assert.NoError(t, err)
	assert.Equal(t, session.NewRecord("id", []byte("data")), record)
}

func TestMemory_SessionSaveDataRefreshesExpiry(t *testing.T) {
	ctx := context.Background()
	provider := NewMemory()

	require.NoError(t, provider.SessionSave(ctx, "example.com", "id", "pid", "john", time.Hour, []byte("data")))

	provider.session[provider.key("example.com", "id")].expires = time.Now().Add(-time.Second)

	require.NoError(t, provider.SessionSaveData(ctx, "example.com", "id", "pid", "john", time.Hour, []byte("updated")))

	record, err := provider.SessionGet(ctx, "example.com", "id")
	assert.NoError(t, err)
	assert.Equal(t, session.NewRecord("id", []byte("updated")), record)
}

func TestMemory_SessionSaveDataNotFound(t *testing.T) {
	provider := NewMemory()

	err := provider.SessionSaveData(context.Background(), "example.com", "id", "pid", "john", time.Hour, []byte("data"))
	assert.EqualError(t, err, "session not found: example.com:id")
}

func TestMemory_SessionUsernameLookup(t *testing.T) {
	testCases := []struct {
		Name     string
		Username string
		Saves    int
		Expected []string
	}{
		{"ShouldNotDuplicateIDsOnRepeatedSaves", "john", 5, []string{"id"}},
		{"ShouldNotIndexAnonymousSessions", "", 5, nil},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := context.Background()
			provider := NewMemory()

			for i := 0; i < tc.Saves; i++ {
				require.NoError(t, provider.SessionSave(ctx, "example.com", "id", "pid", tc.Username, time.Hour, []byte("data")))
			}

			if tc.Expected == nil {
				assert.Empty(t, provider.lookupUsername)
			} else {
				assert.Equal(t, tc.Expected, provider.lookupUsername[provider.key("example.com", tc.Username)])
			}
		})
	}
}

func TestMemory_SessionChangeID(t *testing.T) {
	ctx := context.Background()
	provider := NewMemory()

	require.NoError(t, provider.SessionSave(ctx, "example.com", "id", "pid", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionChangeID(ctx, "example.com", "id", "id2", "pid2", "john", time.Hour, []byte("resealed")))

	record, err := provider.SessionGet(ctx, "example.com", "id2")
	assert.NoError(t, err)
	assert.Equal(t, session.NewRecord("id2", []byte("resealed")), record)

	record, err = provider.SessionGet(ctx, "example.com", "id")
	assert.NoError(t, err)
	assert.Nil(t, record)

	record, err = provider.SessionGetByPublicID(ctx, "example.com", "pid2")
	assert.NoError(t, err)
	assert.Equal(t, session.NewRecord("id2", []byte("resealed")), record)

	record, err = provider.SessionGetByPublicID(ctx, "example.com", "pid")
	assert.NoError(t, err)
	assert.Nil(t, record)

	assert.Len(t, provider.lookupPublicID, 1)
	assert.Equal(t, []string{"id2"}, provider.lookupUsername[provider.key("example.com", "john")])
}

func TestMemory_SessionChangeIDMissingSession(t *testing.T) {
	provider := NewMemory()

	assert.NoError(t, provider.SessionChangeID(context.Background(), "example.com", "id", "id2", "pid2", "john", time.Hour, []byte("resealed")))
	assert.Empty(t, provider.session)
	assert.Empty(t, provider.lookupPublicID)
	assert.Empty(t, provider.lookupUsername)
}

func TestMemory_SessionDelete(t *testing.T) {
	ctx := context.Background()
	provider := NewMemory()

	require.NoError(t, provider.SessionSave(ctx, "example.com", "id", "pid", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionSave(ctx, "example.com", "id2", "pid2", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionDelete(ctx, "example.com", "id", "pid", "john"))

	assert.Equal(t, []string{"id2"}, provider.lookupUsername[provider.key("example.com", "john")])

	require.NoError(t, provider.SessionDelete(ctx, "example.com", "id2", "pid2", "john"))

	assert.Empty(t, provider.session)
	assert.Empty(t, provider.lookupPublicID)
	assert.Empty(t, provider.lookupUsername)
}

func TestMemory_SessionGarbageCollection(t *testing.T) {
	ctx := context.Background()
	provider := NewMemory()

	require.NoError(t, provider.SessionSave(ctx, "example.com", "id", "pid", "john", time.Hour, []byte("data")))
	require.NoError(t, provider.SessionSave(ctx, "example.com", "id2", "pid2", "jane", time.Hour, []byte("data")))

	provider.session[provider.key("example.com", "id")].expires = time.Now().Add(-time.Second)

	require.NoError(t, provider.SessionGarbageCollection(ctx))

	assert.Len(t, provider.session, 1)
	assert.Len(t, provider.lookupPublicID, 1)
	assert.NotContains(t, provider.lookupUsername, provider.key("example.com", "john"))
	assert.Equal(t, []string{"id2"}, provider.lookupUsername[provider.key("example.com", "jane")])
}

func TestMemory_SessionConcurrentAccess(t *testing.T) {
	ctx := context.Background()
	provider := NewMemory()

	wg := &sync.WaitGroup{}

	for i := 0; i < 8; i++ {
		id := fmt.Sprintf("id%d", i)

		require.NoError(t, provider.SessionSave(ctx, "example.com", id, fmt.Sprintf("pid%d", i), "john", time.Hour, []byte("data")))

		wg.Add(1)

		go func() {
			defer wg.Done()

			for n := 0; n < 50; n++ {
				_, _ = provider.SessionGet(ctx, "example.com", id)
				_, _ = provider.SessionGetIDsByUsername(ctx, "example.com", "john")
				_ = provider.SessionSaveData(ctx, "example.com", id, "pid", "john", time.Hour, []byte("updated"))
				_ = provider.SessionGarbageCollection(ctx)
			}
		}()
	}

	wg.Wait()

	ids, err := provider.SessionGetIDsByUsername(ctx, "example.com", "john")
	assert.NoError(t, err)
	assert.Len(t, ids, 8)
}
