package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
)

const (
	// sessionFileExtension is the file extension for session files.
	sessionFileExtension = ".session"

	// sessionFileTempExtension is the extension for temporary files during atomic writes.
	sessionFileTempExtension = ".tmp"

	// orphanedTempFileMaxAge is the maximum age of temporary files before GC removes them.
	// This handles cleanup of temp files left behind after crashes.
	orphanedTempFileMaxAge = time.Minute
)

// sessionFileData represents the JSON structure of a session file.
type sessionFileData struct {
	// CreatedAt is the Unix nanosecond timestamp when the session was first created.
	CreatedAt int64 `json:"created_at"`

	// LastActiveAt is the Unix nanosecond timestamp when the session was last saved.
	LastActiveAt int64 `json:"last_active_at"`

	// ExpiresAt is the Unix nanosecond timestamp when the session expires.
	ExpiresAt int64 `json:"expires_at"`

	// ExpirationDuration is the original expiration duration in nanoseconds.
	// Stored for debugging and auditing purposes; ExpiresAt is used for expiration checks.
	ExpirationDuration int64 `json:"expiration_duration"`

	// Data is the base64-encoded encrypted session data.
	Data string `json:"data"`
}

// FileProvider implements session.Provider for file-based storage.
//
// This provider is designed for single-instance deployments. While multiple instances
// can share the same directory, there is no cross-process locking, so concurrent
// access to the same session from different processes may result in race conditions.
// For multi-instance deployments, use the Redis provider instead.
//
// The provider uses atomic writes (temp file + rename) to prevent torn writes,
// which is safe on local filesystems. NFS and other network filesystems may not
// guarantee atomic rename behavior.
type FileProvider struct {
	config    schema.SessionFile
	stopCh    chan struct{}
	closeOnce sync.Once
}

// NewFileProvider creates a new file-based session provider.
func NewFileProvider(config schema.SessionFile) (*FileProvider, error) {
	if err := os.MkdirAll(config.Path, 0700); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	info, err := os.Stat(config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to stat session directory: %w", err)
	}

	perm := info.Mode().Perm()
	if perm&0077 != 0 {
		logging.Logger().WithFields(logrus.Fields{
			"path":        config.Path,
			"permissions": fmt.Sprintf("%04o", perm),
		}).Warn("Session directory has overly permissive permissions, should be 0700")
	}

	// Defensive check for cleanup interval.
	if config.CleanupInterval <= 0 {
		config.CleanupInterval = schema.DefaultFileConfiguration.CleanupInterval
	}

	p := &FileProvider{
		config: config,
		stopCh: make(chan struct{}),
	}

	go p.cleanupLoop()

	return p, nil
}

// sessionFilePath returns the file path for a session ID.
func (p *FileProvider) sessionFilePath(id []byte) string {
	hash := sha256.Sum256(id)
	filename := hex.EncodeToString(hash[:]) + sessionFileExtension

	return filepath.Join(p.config.Path, filename)
}

// generateTempFilePath generates a unique temporary file path.
// Uses random bytes to prevent collisions across multiple processes.
func (p *FileProvider) generateTempFilePath(basePath string) string {
	randomBytes := make([]byte, 8)

	_, _ = rand.Read(randomBytes)

	return fmt.Sprintf("%s.%d.%s%s",
		basePath,
		time.Now().UnixNano(),
		hex.EncodeToString(randomBytes),
		sessionFileTempExtension,
	)
}

// readSessionFile reads and parses a session file.
func (p *FileProvider) readSessionFile(path string) (*sessionFileData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var fileData sessionFileData
	if err = json.Unmarshal(data, &fileData); err != nil {
		return nil, fmt.Errorf("failed to parse session file: %w", err)
	}

	return &fileData, nil
}

// writeSessionFile writes session data to a file atomically.
func (p *FileProvider) writeSessionFile(path string, fileData *sessionFileData) error {
	data, err := json.Marshal(fileData)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	tmpPath := p.generateTempFilePath(path)

	if err = os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	if err = os.Rename(tmpPath, path); err != nil {
		os.Remove(tmpPath)

		return fmt.Errorf("failed to rename session file: %w", err)
	}

	return nil
}

// Get retrieves session data.
func (p *FileProvider) Get(id []byte) ([]byte, error) {
	path := p.sessionFilePath(id)

	fileData, err := p.readSessionFile(path)
	if os.IsNotExist(err) {
		return nil, nil
	}

	if err != nil {
		// Remove corrupt session files so they don't cause permanent errors.
		logging.Logger().WithError(err).WithField("path", path).Warn("Removing corrupt session file")
		os.Remove(path)

		return nil, nil
	}

	// Check if session has expired.
	if time.Now().UnixNano() > fileData.ExpiresAt {
		os.Remove(path)

		return nil, nil
	}

	data, err := base64.StdEncoding.DecodeString(fileData.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to decode session data: %w", err)
	}

	return data, nil
}

// Save stores session data.
func (p *FileProvider) Save(id, data []byte, expiration time.Duration) error {
	path := p.sessionFilePath(id)
	now := time.Now().UnixNano()

	// Try to read existing file to preserve created_at.
	createdAt := now

	if existingData, err := p.readSessionFile(path); err == nil {
		createdAt = existingData.CreatedAt
	}

	fileData := &sessionFileData{
		CreatedAt:          createdAt,
		LastActiveAt:       now,
		ExpiresAt:          now + expiration.Nanoseconds(),
		ExpirationDuration: expiration.Nanoseconds(),
		Data:               base64.StdEncoding.EncodeToString(data),
	}

	return p.writeSessionFile(path, fileData)
}

// Regenerate creates a new session ID, moving data from old to new.
func (p *FileProvider) Regenerate(id, newID []byte, expiration time.Duration) error {
	oldPath := p.sessionFilePath(id)
	newPath := p.sessionFilePath(newID)

	// Read existing session data.
	existingData, err := p.readSessionFile(oldPath)
	if os.IsNotExist(err) {
		return nil
	}

	if err != nil {
		// Remove corrupt session files so they don't cause permanent errors.
		logging.Logger().WithError(err).WithField("path", oldPath).Warn("Removing corrupt session file during regeneration")
		os.Remove(oldPath)

		return nil
	}

	now := time.Now().UnixNano()

	// Don't regenerate expired sessions.
	if now > existingData.ExpiresAt {
		os.Remove(oldPath)

		return nil
	}

	// Create new session file preserving created_at.
	newFileData := &sessionFileData{
		CreatedAt:          existingData.CreatedAt,
		LastActiveAt:       now,
		ExpiresAt:          now + expiration.Nanoseconds(),
		ExpirationDuration: expiration.Nanoseconds(),
		Data:               existingData.Data,
	}

	// Write new session file atomically.
	if err = p.writeSessionFile(newPath, newFileData); err != nil {
		return err
	}

	// Remove old session file. Ignore errors since the new session is already created.
	os.Remove(oldPath)

	return nil
}

// Destroy removes a session.
func (p *FileProvider) Destroy(id []byte) error {
	path := p.sessionFilePath(id)

	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}

	return err
}

// Count returns the number of sessions.
// Note: This performs a directory scan and may be slow with many sessions.
func (p *FileProvider) Count() int {
	entries, err := os.ReadDir(p.config.Path)
	if err != nil {
		return 0
	}

	count := 0

	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == sessionFileExtension {
			count++
		}
	}

	return count
}

// NeedGC returns false because we handle GC internally via cleanupLoop.
// Note: This only affects the fasthttp/session library's automatic GC scheduling.
// The GC() method can still be called directly if needed.
func (p *FileProvider) NeedGC() bool {
	return false
}

// GC performs garbage collection of expired sessions and orphaned temp files.
func (p *FileProvider) GC() error {
	entries, err := os.ReadDir(p.config.Path)
	if err != nil {
		return err
	}

	log := logging.Logger()
	now := time.Now()
	nowNano := now.UnixNano()
	expiredCount := 0
	tempCleanedCount := 0

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		path := filepath.Join(p.config.Path, name)

		// Clean up orphaned temp files.
		if strings.HasSuffix(name, sessionFileTempExtension) {
			info, err := e.Info()
			if err != nil {
				log.WithError(err).WithField("file", name).Debug("Failed to get temp file info during GC")

				continue
			}

			if now.Sub(info.ModTime()) > orphanedTempFileMaxAge {
				if err = os.Remove(path); err == nil {
					tempCleanedCount++

					log.WithField("file", name).Debug("Cleaned up orphaned temp file")
				} else {
					log.WithError(err).WithField("file", name).Debug("Failed to remove orphaned temp file")
				}
			}

			continue
		}

		// Skip non-session files.
		if filepath.Ext(name) != sessionFileExtension {
			continue
		}

		fileData, err := p.readSessionFile(path)
		if err != nil {
			log.WithError(err).WithField("file", name).Warn("Removing corrupt session file during GC")

			if removeErr := os.Remove(path); removeErr != nil {
				log.WithError(removeErr).WithField("file", name).Warn("Failed to remove corrupt session file during GC")
			}

			continue
		}

		if nowNano > fileData.ExpiresAt {
			if err = os.Remove(path); err == nil {
				expiredCount++

				log.WithFields(logrus.Fields{
					"file":       name,
					"expired_at": time.Unix(0, fileData.ExpiresAt).Format(time.RFC3339),
				}).Debug("Cleaned up expired session file")
			} else {
				log.WithError(err).WithField("file", name).Debug("Failed to remove expired session file")
			}
		}
	}

	if expiredCount > 0 || tempCleanedCount > 0 {
		log.WithFields(logrus.Fields{
			"expired_sessions": expiredCount,
			"orphaned_temps":   tempCleanedCount,
		}).Debug("Session GC completed")
	}

	return nil
}

// Close stops the cleanup goroutine.
func (p *FileProvider) Close() error {
	p.closeOnce.Do(func() {
		close(p.stopCh)
	})

	return nil
}

func (p *FileProvider) cleanupLoop() {
	ticker := time.NewTicker(p.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.GC(); err != nil {
				logging.Logger().WithError(err).Warn("Session file cleanup failed")
			}
		case <-p.stopCh:
			return
		}
	}
}
