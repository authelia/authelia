package session

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestFileProvider_SaveAndGet(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	id := []byte("test-session-id")
	data := []byte("test-data")

	err = provider.Save(id, data, time.Hour)
	require.NoError(t, err)

	retrieved, err := provider.Get(id)
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)
}

func TestFileProvider_GetNonExistent(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	retrieved, err := provider.Get([]byte("non-existent"))
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestFileProvider_Destroy(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	id := []byte("test-session-id")
	data := []byte("test-data")

	err = provider.Save(id, data, time.Hour)
	require.NoError(t, err)

	err = provider.Destroy(id)
	require.NoError(t, err)

	retrieved, err := provider.Get(id)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestFileProvider_DestroyNonExistent(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	err = provider.Destroy([]byte("non-existent"))
	require.NoError(t, err)
}

func TestFileProvider_Regenerate(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	oldID := []byte("old-session-id")
	newID := []byte("new-session-id")
	data := []byte("test-data")

	err = provider.Save(oldID, data, time.Hour)
	require.NoError(t, err)

	err = provider.Regenerate(oldID, newID, time.Hour)
	require.NoError(t, err)

	oldData, err := provider.Get(oldID)
	require.NoError(t, err)
	assert.Nil(t, oldData)

	newData, err := provider.Get(newID)
	require.NoError(t, err)
	assert.Equal(t, data, newData)
}

func TestFileProvider_RegeneratePreservesCreatedAt(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	oldID := []byte("old-session-id")
	newID := []byte("new-session-id")
	data := []byte("test-data")

	err = provider.Save(oldID, data, time.Hour)
	require.NoError(t, err)

	// Read original created_at.
	oldPath := provider.sessionFilePath(oldID)
	oldFileData, err := provider.readSessionFile(oldPath)
	require.NoError(t, err)

	originalCreatedAt := oldFileData.CreatedAt

	// Wait a bit to ensure time difference.
	time.Sleep(10 * time.Millisecond)

	err = provider.Regenerate(oldID, newID, time.Hour)
	require.NoError(t, err)

	// Check new session has same created_at.
	newPath := provider.sessionFilePath(newID)
	newFileData, err := provider.readSessionFile(newPath)
	require.NoError(t, err)

	assert.Equal(t, originalCreatedAt, newFileData.CreatedAt)
	assert.Greater(t, newFileData.LastActiveAt, originalCreatedAt)
}

func TestFileProvider_RegenerateExpiredSession(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	oldID := []byte("expired-session")
	newID := []byte("new-session")

	err = provider.Save(oldID, []byte("data"), time.Hour)
	require.NoError(t, err)

	// Manually expire the session.
	oldPath := provider.sessionFilePath(oldID)
	fileData, err := provider.readSessionFile(oldPath)
	require.NoError(t, err)

	fileData.ExpiresAt = time.Now().Add(-time.Hour).UnixNano()
	rawData, err := json.Marshal(fileData)
	require.NoError(t, err)

	err = os.WriteFile(oldPath, rawData, 0600)
	require.NoError(t, err)

	// Regenerate should not carry forward the expired session.
	err = provider.Regenerate(oldID, newID, time.Hour)
	require.NoError(t, err)

	// Old file should be removed.
	_, err = os.Stat(oldPath)
	assert.True(t, os.IsNotExist(err))

	// New session should not exist.
	retrieved, err := provider.Get(newID)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestFileProvider_RegenerateNonExistent(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	err = provider.Regenerate([]byte("non-existent"), []byte("new-id"), time.Hour)
	require.NoError(t, err)
}

func TestFileProvider_Count(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	assert.Equal(t, 0, provider.Count())

	for i := 0; i < 5; i++ {
		id := []byte(fmt.Sprintf("session-%d", i))
		err = provider.Save(id, []byte("data"), time.Hour)
		require.NoError(t, err)
	}

	assert.Equal(t, 5, provider.Count())
}

func TestFileProvider_NeedGC(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	// Should return false since we handle GC internally.
	assert.False(t, provider.NeedGC())
}

func TestFileProvider_GCRemovesExpiredSessions(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	id := []byte("expired-session")
	err = provider.Save(id, []byte("data"), time.Hour)
	require.NoError(t, err)

	// Manually modify the expires_at to be in the past.
	path := provider.sessionFilePath(id)
	fileData, err := provider.readSessionFile(path)
	require.NoError(t, err)

	fileData.ExpiresAt = time.Now().Add(-time.Hour).UnixNano()
	data, err := json.Marshal(fileData)
	require.NoError(t, err)
	err = os.WriteFile(path, data, 0600)
	require.NoError(t, err)

	// Run GC.
	err = provider.GC()
	require.NoError(t, err)

	// Session should be gone.
	retrieved, err := provider.Get(id)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestFileProvider_GCKeepsActiveSessions(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	id := []byte("active-session")
	data := []byte("test-data")
	err = provider.Save(id, data, time.Hour)
	require.NoError(t, err)

	// Run GC.
	err = provider.GC()
	require.NoError(t, err)

	// Session should still exist.
	retrieved, err := provider.Get(id)
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)
}

func TestFileProvider_GetRemovesCorruptSessionFile(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	// Create a session so we can get its file path.
	id := []byte("corrupt-session")

	err = provider.Save(id, []byte("data"), time.Hour)
	require.NoError(t, err)

	// Overwrite with invalid JSON.
	path := provider.sessionFilePath(id)
	err = os.WriteFile(path, []byte("not valid json{{{"), 0600)
	require.NoError(t, err)

	// Get should return nil (not an error) and remove the corrupt file.
	retrieved, err := provider.Get(id)
	require.NoError(t, err)
	assert.Nil(t, retrieved)

	// File should be removed.
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestFileProvider_GetRemovesExpiredSessions(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	id := []byte("expired-session")
	err = provider.Save(id, []byte("data"), time.Hour)
	require.NoError(t, err)

	// Manually modify the expires_at to be in the past.
	path := provider.sessionFilePath(id)
	fileData, err := provider.readSessionFile(path)
	require.NoError(t, err)

	fileData.ExpiresAt = time.Now().Add(-time.Hour).UnixNano()
	data, err := json.Marshal(fileData)
	require.NoError(t, err)
	err = os.WriteFile(path, data, 0600)
	require.NoError(t, err)

	// Get should return nil and remove the file.
	retrieved, err := provider.Get(id)
	require.NoError(t, err)
	assert.Nil(t, retrieved)

	// File should be removed.
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

func TestFileProvider_CreateDirectory(t *testing.T) {
	dir := t.TempDir()
	sessionDir := filepath.Join(dir, "sessions", "nested")

	config := schema.SessionFile{
		Path:            sessionDir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	info, err := os.Stat(sessionDir)
	require.NoError(t, err)
	assert.True(t, info.IsDir())
}

func TestFileProvider_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	id := []byte("test-session")
	data1 := []byte("initial-data")
	data2 := []byte("updated-data")

	err = provider.Save(id, data1, time.Hour)
	require.NoError(t, err)

	err = provider.Save(id, data2, time.Hour)
	require.NoError(t, err)

	retrieved, err := provider.Get(id)
	require.NoError(t, err)
	assert.Equal(t, data2, retrieved)
}

func TestFileProvider_SavePreservesCreatedAt(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	id := []byte("test-session")

	err = provider.Save(id, []byte("data1"), time.Hour)
	require.NoError(t, err)

	// Read original created_at.
	path := provider.sessionFilePath(id)
	fileData1, err := provider.readSessionFile(path)
	require.NoError(t, err)

	originalCreatedAt := fileData1.CreatedAt

	// Wait a bit.
	time.Sleep(10 * time.Millisecond)

	// Save again.
	err = provider.Save(id, []byte("data2"), time.Hour)
	require.NoError(t, err)

	// Check created_at is preserved.
	fileData2, err := provider.readSessionFile(path)
	require.NoError(t, err)

	assert.Equal(t, originalCreatedAt, fileData2.CreatedAt)
	assert.Greater(t, fileData2.LastActiveAt, fileData1.LastActiveAt)
}

func TestFileProvider_JSONFormat(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	id := []byte("test-session")
	data := []byte("test-data")
	expiration := time.Hour

	err = provider.Save(id, data, expiration)
	require.NoError(t, err)

	// Read raw file and verify JSON structure.
	path := provider.sessionFilePath(id)
	rawData, err := os.ReadFile(path)
	require.NoError(t, err)

	var fileData sessionFileData

	err = json.Unmarshal(rawData, &fileData)
	require.NoError(t, err)

	assert.Greater(t, fileData.CreatedAt, int64(0))
	assert.Greater(t, fileData.LastActiveAt, int64(0))
	assert.Greater(t, fileData.ExpiresAt, fileData.LastActiveAt)
	assert.Equal(t, expiration.Nanoseconds(), fileData.ExpirationDuration)

	// Verify data is base64 encoded.
	decodedData, err := base64.StdEncoding.DecodeString(fileData.Data)
	require.NoError(t, err)
	assert.Equal(t, data, decodedData)
}

func TestFileProvider_CloseIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	// Close multiple times should not panic.
	err = provider.Close()
	require.NoError(t, err)

	err = provider.Close()
	require.NoError(t, err)

	err = provider.Close()
	require.NoError(t, err)
}

func TestFileProvider_ConcurrentAccess(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	const (
		numGoroutines = 10
		numOperations = 100
	)

	var wg sync.WaitGroup

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < numOperations; j++ {
				id := []byte(fmt.Sprintf("session-%d-%d", goroutineID, j))
				data := []byte(fmt.Sprintf("data-%d-%d", goroutineID, j))

				// Save.
				err := provider.Save(id, data, time.Hour)
				assert.NoError(t, err)

				// Get.
				retrieved, err := provider.Get(id)
				assert.NoError(t, err)
				assert.Equal(t, data, retrieved)

				// Destroy half.
				if j%2 == 0 {
					err = provider.Destroy(id)
					assert.NoError(t, err)
				}
			}
		}(i)
	}

	wg.Wait()

	// Verify count is approximately half.
	count := provider.Count()
	expectedCount := numGoroutines * numOperations / 2
	assert.Equal(t, expectedCount, count)
}

func TestFileProvider_ExpirationEnforced(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	// Create session with very short expiration.
	id := []byte("short-lived-session")
	data := []byte("test-data")
	err = provider.Save(id, data, 50*time.Millisecond)
	require.NoError(t, err)

	// Should be retrievable immediately.
	retrieved, err := provider.Get(id)
	require.NoError(t, err)
	assert.Equal(t, data, retrieved)

	// Wait for expiration.
	time.Sleep(100 * time.Millisecond)

	// Should return nil now.
	retrieved, err = provider.Get(id)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestFileProvider_GCCleansOrphanedTempFiles(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	// Create an orphaned temp file with old modification time.
	tempPath := filepath.Join(dir, "test.1234567890.abcd1234.tmp")
	err = os.WriteFile(tempPath, []byte("orphaned"), 0600)
	require.NoError(t, err)

	// Set modification time to 2 minutes ago.
	oldTime := time.Now().Add(-2 * time.Minute)
	err = os.Chtimes(tempPath, oldTime, oldTime)
	require.NoError(t, err)

	// Run GC.
	err = provider.GC()
	require.NoError(t, err)

	// Temp file should be removed.
	_, err = os.Stat(tempPath)
	assert.True(t, os.IsNotExist(err))
}

func TestFileProvider_GCKeepsRecentTempFiles(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	// Create a recent temp file.
	tempPath := filepath.Join(dir, "test.1234567890.abcd1234.tmp")
	err = os.WriteFile(tempPath, []byte("recent"), 0600)
	require.NoError(t, err)

	// Run GC.
	err = provider.GC()
	require.NoError(t, err)

	// Temp file should still exist.
	_, err = os.Stat(tempPath)
	assert.NoError(t, err)
}

func TestFileProvider_RegenerateRemovesCorruptSessionFile(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	oldID := []byte("corrupt-session")
	newID := []byte("new-session")

	err = provider.Save(oldID, []byte("data"), time.Hour)
	require.NoError(t, err)

	// Overwrite with invalid JSON.
	oldPath := provider.sessionFilePath(oldID)
	err = os.WriteFile(oldPath, []byte("not valid json{{{"), 0600)
	require.NoError(t, err)

	// Regenerate should return nil (not an error) and remove the corrupt file.
	err = provider.Regenerate(oldID, newID, time.Hour)
	require.NoError(t, err)

	// Old corrupt file should be removed.
	_, err = os.Stat(oldPath)
	assert.True(t, os.IsNotExist(err))

	// New session should not exist.
	retrieved, err := provider.Get(newID)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}

func TestFileProvider_GCRemovesCorruptSessionFiles(t *testing.T) {
	dir := t.TempDir()
	config := schema.SessionFile{
		Path:            dir,
		CleanupInterval: time.Hour,
	}

	provider, err := NewFileProvider(config)
	require.NoError(t, err)

	defer provider.Close()

	// Create a session so we can get its file path.
	id := []byte("corrupt-gc-session")

	err = provider.Save(id, []byte("data"), time.Hour)
	require.NoError(t, err)

	// Overwrite with invalid JSON.
	path := provider.sessionFilePath(id)
	err = os.WriteFile(path, []byte("not valid json{{{"), 0600)
	require.NoError(t, err)

	// Run GC.
	err = provider.GC()
	require.NoError(t, err)

	// Corrupt file should be removed.
	_, err = os.Stat(path)
	assert.True(t, os.IsNotExist(err))
}

