package session

import (
	"context"
	"time"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/storage"
)

// SQLSessionProvider implements fasthttp/session/v2.Provider using the storage.SessionProvider SQL backend.
type SQLSessionProvider struct {
	storage storage.SessionProvider
}

// NewSQLSessionProvider creates a new SQL-backed session provider.
func NewSQLSessionProvider(storage storage.SessionProvider) *SQLSessionProvider {
	return &SQLSessionProvider{storage: storage}
}

// Get retrieves session data by ID.
func (p *SQLSessionProvider) Get(id []byte) ([]byte, error) {
	data, err := p.storage.LoadSession(context.Background(), string(id))
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Save stores session data with an expiration duration.
func (p *SQLSessionProvider) Save(id, data []byte, expiration time.Duration) error {
	now := time.Now()

	return p.storage.SaveSession(context.Background(), string(id), data, now, now.Add(expiration))
}

// Destroy removes a session by ID.
func (p *SQLSessionProvider) Destroy(id []byte) error {
	return p.storage.DeleteSession(context.Background(), string(id))
}

// Regenerate creates a new session from an existing one.
// Non-atomic: Load old → Save new → Delete old. Worst case is a brief stale session that GC cleans up.
func (p *SQLSessionProvider) Regenerate(id, newID []byte, expiration time.Duration) error {
	ctx := context.Background()

	data, err := p.storage.LoadSession(ctx, string(id))
	if err != nil {
		return err
	}

	// Nothing to regenerate if old session doesn't exist.
	if data == nil {
		return nil
	}

	now := time.Now()

	if err = p.storage.SaveSession(ctx, string(newID), data, now, now.Add(expiration)); err != nil {
		return err
	}

	if err = p.storage.DeleteSession(ctx, string(id)); err != nil {
		logging.Logger().WithError(err).Warn("Failed to delete old session during regeneration")
	}

	return nil
}

// Count returns the number of non-expired sessions.
func (p *SQLSessionProvider) Count() int {
	count, err := p.storage.CountSessions(context.Background())
	if err != nil {
		logging.Logger().WithError(err).Warn("Failed to count sessions")
		return 0
	}

	return count
}

// NeedGC returns true to let the fasthttp/session framework handle GC scheduling.
func (p *SQLSessionProvider) NeedGC() bool {
	return true
}

// GC removes expired sessions from the database.
func (p *SQLSessionProvider) GC() error {
	return p.storage.DeleteExpiredSessions(context.Background())
}
