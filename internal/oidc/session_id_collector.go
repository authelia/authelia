// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/model"
)

const (
	sessionIDCollectionFrequency = time.Hour
	sessionIDCollectionBatch     = 500
	sessionIDCollectionLimit     = 5000
)

// SessionIDStorage is the storage this collector needs to enumerate and remove session id mappings.
type SessionIDStorage interface {
	LoadOAuth2SessionIDsOldest(ctx context.Context, after, limit int) (records []model.OAuth2SessionID, err error)
	DeleteOAuth2SessionID(ctx context.Context, issuer, sid string) (err error)
}

// SessionIDLookup reports whether a session still exists. A nil record with a nil error means the session is
// definitively gone; a non-nil error means the backend could not answer and nothing may be deleted.
type SessionIDLookup interface {
	GetByPublicID(ctx context.Context, issuer, pid string) (found bool, err error)
}

// NewSessionIDCollector returns a new *SessionIDCollector using the given storage and lookup.
func NewSessionIDCollector(storage SessionIDStorage, lookup SessionIDLookup) *SessionIDCollector {
	return &SessionIDCollector{storage: storage, lookup: lookup}
}

// SessionIDCollector removes 'sid' mappings whose session no longer exists. The session store is authoritative, so a
// mapping is removed only when the store reports the session definitively absent. A pass aborts immediately on any
// lookup or storage error, deleting nothing further, and resumes from the point it stopped on the next pass rather
// than restarting from the beginning, since an unreachable backend must never be read as every session having ended,
// and must not be allowed to permanently stall the sweep at the head of the table by always re-scanning from zero.
type SessionIDCollector struct {
	storage SessionIDStorage
	lookup  SessionIDLookup

	// cursor is the exclusive lower bound the next Collect call resumes from. It is reset to zero once a pass reaches
	// the end of the table, which is what turns repeated calls into a sweep rather than a single one-shot pass.
	cursor int
}

// Collect walks the mappings oldest first in pages of pageSize, removing every one whose session is definitively
// absent, until it has examined at least maxRows rows or reached the end of the table, whichever happens first. It
// resumes from the cursor left by the previous call instead of starting over, so a bounded per-run cost still
// eventually reaches every row. Reaching the end of the table resets the cursor to zero so the next call starts a
// fresh sweep; stopping early, whether the row cap was reached or a lookup or storage error occurred, leaves the
// cursor at the point already reached so the next call resumes there rather than re-scanning the head of the table.
// Any lookup or storage error aborts the pass immediately, deleting nothing further in that page or beyond, and is
// returned to the caller along with however many deletes already completed.
func (c *SessionIDCollector) Collect(ctx context.Context, pageSize, maxRows int) (deleted int, err error) {
	after := c.cursor

	var processed int

	for {
		var records []model.OAuth2SessionID

		if records, err = c.storage.LoadOAuth2SessionIDsOldest(ctx, after, pageSize); err != nil {
			c.cursor = after

			return deleted, err
		}

		for _, record := range records {
			var found bool

			if found, err = c.lookup.GetByPublicID(ctx, record.Issuer, record.PublicID); err != nil {
				c.cursor = after

				return deleted, err
			}

			if found {
				continue
			}

			if err = c.storage.DeleteOAuth2SessionID(ctx, record.Issuer, record.SessionID.String()); err != nil {
				c.cursor = after

				return deleted, err
			}

			deleted++
		}

		processed += len(records)

		if len(records) < pageSize {
			// Reached the end of the table. Reset so the next call sweeps from the beginning again, otherwise a
			// collector which ever finished a full pass would never collect anything created afterwards.
			c.cursor = 0

			return deleted, nil
		}

		last := records[len(records)-1].ID

		// Ids are monotonic and the query is strictly greater than the cursor, so the last id of a full page must
		// advance the cursor. This guards against an infinite loop if that were ever violated, rather than hanging
		// the collector service.
		if last <= after {
			c.cursor = after

			return deleted, fmt.Errorf("session id collector cursor did not advance past %d", after)
		}

		after = last

		if processed >= maxRows {
			// Hit the per-run row cap with more of the table left to examine. Leave the cursor here so the next call
			// resumes past this point instead of re-scanning from the start.
			c.cursor = after

			return deleted, nil
		}
	}
}

// GarbageCollection implements the service.GarbageCollectorProvider interface.
func (c *SessionIDCollector) GarbageCollection(ctx context.Context) (err error) {
	_, err = c.Collect(ctx, sessionIDCollectionBatch, sessionIDCollectionLimit)

	return err
}

// GarbageCollectionFrequency implements the service.GarbageCollectorProvider interface. Orphaned mappings are inert,
// so this runs rarely.
func (c *SessionIDCollector) GarbageCollectionFrequency(_ context.Context) (frequency time.Duration) {
	return sessionIDCollectionFrequency
}
