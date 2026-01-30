package session

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSessionStorage implements storage.SessionProvider for testing.
type mockSessionStorage struct {
	sessions map[string][]byte
	saveErr  error
	loadErr  error
	delErr   error
	gcErr    error
	countErr error
	count    int

	lastSaveLastActiveAt time.Time
	lastSaveExpiresAt    time.Time
}

func newMockSessionStorage() *mockSessionStorage {
	return &mockSessionStorage{sessions: make(map[string][]byte)}
}

func (m *mockSessionStorage) SaveSession(_ context.Context, sessionID string, data []byte, lastActiveAt, expiresAt time.Time) error {
	if m.saveErr != nil {
		return m.saveErr
	}

	m.sessions[sessionID] = data
	m.lastSaveLastActiveAt = lastActiveAt
	m.lastSaveExpiresAt = expiresAt

	return nil
}

func (m *mockSessionStorage) LoadSession(_ context.Context, sessionID string) ([]byte, error) {
	if m.loadErr != nil {
		return nil, m.loadErr
	}

	return m.sessions[sessionID], nil
}

func (m *mockSessionStorage) DeleteSession(_ context.Context, sessionID string) error {
	if m.delErr != nil {
		return m.delErr
	}

	delete(m.sessions, sessionID)

	return nil
}

func (m *mockSessionStorage) DeleteExpiredSessions(_ context.Context) error {
	return m.gcErr
}

func (m *mockSessionStorage) CountSessions(_ context.Context) (int, error) {
	if m.countErr != nil {
		return 0, m.countErr
	}

	return m.count, nil
}

// deleteFailMock extends mockSessionStorage to fail only on DeleteSession.
type deleteFailMock struct {
	mockSessionStorage
}

func (m *deleteFailMock) DeleteSession(_ context.Context, _ string) error {
	return fmt.Errorf("delete failed")
}

func TestSQLSessionProvider_SaveAndGet(t *testing.T) {
	mock := newMockSessionStorage()
	provider := NewSQLSessionProvider(mock)

	id := []byte("test-session-id")
	data := []byte("encrypted-session-data")

	err := provider.Save(id, data, time.Hour)
	require.NoError(t, err)

	result, err := provider.Get(id)
	require.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestSQLSessionProvider_GetNonExistent(t *testing.T) {
	mock := newMockSessionStorage()
	provider := NewSQLSessionProvider(mock)

	result, err := provider.Get([]byte("nonexistent"))
	require.NoError(t, err)
	assert.Nil(t, result)
}

func TestSQLSessionProvider_Destroy(t *testing.T) {
	mock := newMockSessionStorage()
	provider := NewSQLSessionProvider(mock)

	mock.sessions["test-session-id"] = []byte("data")

	err := provider.Destroy([]byte("test-session-id"))
	require.NoError(t, err)
	assert.Empty(t, mock.sessions)
}

func TestSQLSessionProvider_Regenerate(t *testing.T) {
	mock := newMockSessionStorage()
	provider := NewSQLSessionProvider(mock)

	data := []byte("session-data")
	mock.sessions["old-session-id"] = data

	err := provider.Regenerate([]byte("old-session-id"), []byte("new-session-id"), time.Hour)
	require.NoError(t, err)

	assert.Equal(t, data, mock.sessions["new-session-id"])
	assert.Nil(t, mock.sessions["old-session-id"])
}

func TestSQLSessionProvider_RegenerateNonExistent(t *testing.T) {
	mock := newMockSessionStorage()
	provider := NewSQLSessionProvider(mock)

	err := provider.Regenerate([]byte("nonexistent"), []byte("new-id"), time.Hour)
	require.NoError(t, err)
	assert.Empty(t, mock.sessions)
}

func TestSQLSessionProvider_Count(t *testing.T) {
	mock := newMockSessionStorage()
	mock.count = 42
	provider := NewSQLSessionProvider(mock)

	assert.Equal(t, 42, provider.Count())
}

func TestSQLSessionProvider_CountError(t *testing.T) {
	mock := newMockSessionStorage()
	mock.countErr = fmt.Errorf("db error")
	provider := NewSQLSessionProvider(mock)

	assert.Equal(t, 0, provider.Count())
}

func TestSQLSessionProvider_GC(t *testing.T) {
	mock := newMockSessionStorage()
	provider := NewSQLSessionProvider(mock)

	err := provider.GC()
	require.NoError(t, err)
}

func TestSQLSessionProvider_NeedGC(t *testing.T) {
	mock := newMockSessionStorage()
	provider := NewSQLSessionProvider(mock)

	assert.True(t, provider.NeedGC())
}

func TestSQLSessionProvider_GetError(t *testing.T) {
	mock := newMockSessionStorage()
	mock.loadErr = fmt.Errorf("db connection lost")
	provider := NewSQLSessionProvider(mock)

	result, err := provider.Get([]byte("sid"))
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestSQLSessionProvider_SaveError(t *testing.T) {
	mock := newMockSessionStorage()
	mock.saveErr = fmt.Errorf("db full")
	provider := NewSQLSessionProvider(mock)

	err := provider.Save([]byte("sid"), []byte("data"), time.Hour)
	assert.Error(t, err)
}

func TestSQLSessionProvider_DestroyError(t *testing.T) {
	mock := newMockSessionStorage()
	mock.delErr = fmt.Errorf("db locked")
	provider := NewSQLSessionProvider(mock)

	err := provider.Destroy([]byte("sid"))
	assert.Error(t, err)
}

func TestSQLSessionProvider_GCError(t *testing.T) {
	mock := newMockSessionStorage()
	mock.gcErr = fmt.Errorf("db timeout")
	provider := NewSQLSessionProvider(mock)

	err := provider.GC()
	assert.Error(t, err)
}

func TestSQLSessionProvider_RegenerateLoadError(t *testing.T) {
	mock := newMockSessionStorage()
	mock.loadErr = fmt.Errorf("db read error")
	provider := NewSQLSessionProvider(mock)

	err := provider.Regenerate([]byte("old"), []byte("new"), time.Hour)
	assert.Error(t, err)
}

func TestSQLSessionProvider_RegenerateSaveError(t *testing.T) {
	mock := newMockSessionStorage()
	mock.sessions["old"] = []byte("data")
	mock.saveErr = fmt.Errorf("db write error")
	provider := NewSQLSessionProvider(mock)

	err := provider.Regenerate([]byte("old"), []byte("new"), time.Hour)
	assert.Error(t, err)
}

func TestSQLSessionProvider_RegenerateDeleteError(t *testing.T) {
	mock := newMockSessionStorage()
	mock.sessions["old"] = []byte("data")

	// Use a custom mock that fails on delete only after save succeeds.
	mock2 := &deleteFailMock{mockSessionStorage: *mock}
	provider2 := NewSQLSessionProvider(mock2)

	err := provider2.Regenerate([]byte("old"), []byte("new"), time.Hour)
	// Should succeed — delete error is logged but not propagated (matches file provider behavior).
	assert.NoError(t, err)
	// New session should exist.
	assert.NotNil(t, mock2.sessions["new"])
}

func TestSQLSessionProvider_SaveSetsExpiration(t *testing.T) {
	mock := newMockSessionStorage()
	provider := NewSQLSessionProvider(mock)

	data := []byte("data")
	expiration := 30 * time.Minute

	err := provider.Save([]byte("sid"), data, expiration)
	require.NoError(t, err)

	assert.WithinDuration(t, time.Now(), mock.lastSaveLastActiveAt, time.Second)
	assert.WithinDuration(t, time.Now().Add(expiration), mock.lastSaveExpiresAt, time.Second)
}
