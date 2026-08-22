package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/session"
)

// NewSessionRepository returns a session.Repository which is backed by the storage Provider.
func NewSessionRepository(provider Provider) session.Repository {
	return SessionRepository{provider: provider}
}

// SessionRepository adapts the storage Provider to the session.Repository interface.
type SessionRepository struct {
	provider Provider
}

// Get implements the session.Repository interface.
func (r SessionRepository) Get(ctx context.Context, issuer, id string) (record session.Record, err error) {
	return r.provider.SessionGet(ctx, issuer, id)
}

// GetByPublicID implements the session.Repository interface.
func (r SessionRepository) GetByPublicID(ctx context.Context, issuer, pid string) (record session.Record, err error) {
	return r.provider.SessionGetByPublicID(ctx, issuer, pid)
}

// GetIDsByUsername implements the session.Repository interface.
func (r SessionRepository) GetIDsByUsername(ctx context.Context, issuer, username string) (ids []string, err error) {
	return r.provider.SessionGetIDsByUsername(ctx, issuer, username)
}

// Save implements the session.Repository interface.
func (r SessionRepository) Save(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	return r.provider.SessionSave(ctx, issuer, id, pid, username, expiration, data)
}

// SaveData implements the session.Repository interface.
func (r SessionRepository) SaveData(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	return r.provider.SessionSaveData(ctx, issuer, id, pid, username, expiration, data)
}

// Delete implements the session.Repository interface.
func (r SessionRepository) Delete(ctx context.Context, issuer, id, pid, username string) (err error) {
	return r.provider.SessionDelete(ctx, issuer, id, pid, username)
}

// ChangeID implements the session.Repository interface.
func (r SessionRepository) ChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	return r.provider.SessionChangeID(ctx, issuer, oldID, id, pid, username, expiration, data)
}

// GarbageCollection implements the session.Repository interface.
func (r SessionRepository) GarbageCollection(ctx context.Context) (err error) {
	return r.provider.SessionGarbageCollection(ctx)
}

// GarbageCollectionFrequency implements the session.Repository interface.
func (r SessionRepository) GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	return r.provider.SessionGarbageCollectionFrequency(ctx)
}

var (
	_ session.Repository = (*SessionRepository)(nil)
)

// sessionRow is the shape a session is selected into, which is converted into a session.Record for the caller.
type sessionRow struct {
	Signature string `db:"signature"`
	Data      []byte `db:"data"`
}

// SessionGet returns the session for the given signature and issuer. A session which does not exist or which has
// expired returns an empty record and no error, matching the semantics the session provider expects for anonymous
// requests.
func (p *SQLProvider) SessionGet(ctx context.Context, issuer, id string) (record session.Record, err error) {
	var row sessionRow

	if err = p.db.GetContext(ctx, &row, p.sqlSelectSession, issuer, id, time.Now()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting session: %w", err)
	}

	return session.NewRecord(row.Signature, row.Data), nil
}

// SessionGetByPublicID returns the session for the given public id and issuer.
func (p *SQLProvider) SessionGetByPublicID(ctx context.Context, issuer, pid string) (record session.Record, err error) {
	var row sessionRow

	if err = p.db.GetContext(ctx, &row, p.sqlSelectSessionByPublicID, issuer, pid, time.Now()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting session by public id: %w", err)
	}

	return session.NewRecord(row.Signature, row.Data), nil
}

// SessionGetIDsByUsername returns the signatures of every unexpired session belonging to the given username and issuer.
func (p *SQLProvider) SessionGetIDsByUsername(ctx context.Context, issuer, username string) (ids []string, err error) {
	if err = p.db.SelectContext(ctx, &ids, p.sqlSelectSessionSignaturesByUsername, issuer, username, time.Now()); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting session ids by username: %w", err)
	}

	return ids, nil
}

// SessionSave persists a session, replacing any existing session with the same signature and issuer.
func (p *SQLProvider) SessionSave(ctx context.Context, issuer, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpsertSession, issuer, id, pid, username, p.sessionExpires(expiration), data); err != nil {
		return fmt.Errorf("error upserting session: %w", err)
	}

	return nil
}

// SessionSaveData updates the data and expiration of an existing session. The public id and username are unused as
// they are columns of the session row itself rather than separate records with their own expiry.
func (p *SQLProvider) SessionSaveData(ctx context.Context, issuer, id, _, _ string, expiration time.Duration, data []byte) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateSessionData, p.sessionExpires(expiration), data, issuer, id); err != nil {
		return fmt.Errorf("error updating session data: %w", err)
	}

	return nil
}

// SessionDelete removes a session.
func (p *SQLProvider) SessionDelete(ctx context.Context, issuer, id, pid, username string) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteSession, issuer, id); err != nil {
		return fmt.Errorf("error deleting session: %w", err)
	}

	return nil
}

// SessionChangeID changes the signature of an existing session, which is used to regenerate the session identifier
// without discarding the session itself. The data is updated alongside the signature as the caller reseals the session
// against the signature it is stored under, and the two must never disagree.
func (p *SQLProvider) SessionChangeID(ctx context.Context, issuer, oldID, id, pid, username string, expiration time.Duration, data []byte) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlUpdateSessionSignature, id, p.sessionExpires(expiration), data, issuer, oldID); err != nil {
		return fmt.Errorf("error updating session signature: %w", err)
	}

	return nil
}

// SessionGarbageCollection removes every expired session.
func (p *SQLProvider) SessionGarbageCollection(ctx context.Context) (err error) {
	if _, err = p.db.ExecContext(ctx, p.sqlDeleteSessionExpired, time.Now()); err != nil {
		return fmt.Errorf("error deleting expired sessions: %w", err)
	}

	return nil
}

// SessionGarbageCollectionFrequency returns the frequency expired sessions should be removed at.
func (p *SQLProvider) SessionGarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	return sessionGarbageCollectionFrequency
}

func (p *SQLProvider) sessionExpires(expiration time.Duration) (expires time.Time) {
	return time.Now().Add(expiration)
}
