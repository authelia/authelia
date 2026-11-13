// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

const (
	sessionIDCollectionFrequency = 5 * time.Minute
	sessionIDCollectionBatch     = 500
	sessionIDCollectionLimit     = 50000

	// SessionIDCollectionConfirm is how long the collector waits before looking up a session which appeared to be
	// absent a second time, which avoids ending a session whose record is momentarily absent while its identifier is
	// being regenerated.
	SessionIDCollectionConfirm = 2 * time.Second
)

// SessionIDStorage is the storage this collector needs to enumerate and end OpenID Connect sessions.
type SessionIDStorage interface {
	LoadOAuth2SessionIDsOldest(ctx context.Context, after, limit int) (records []model.OAuth2SessionID, err error)
	EndOAuth2Session(ctx context.Context, issuer, publicID string, schedule model.OAuth2SessionLogoutTaskSchedule) (sids []string, err error)
	RevokeOAuth2RefreshTokenSessionsBySessionID(ctx context.Context, sid string) (err error)
}

// SessionIDLookup reports whether a session still exists. A false result with a nil error means the session is
// definitively gone; a non-nil error means the backend could not answer and nothing may be ended.
type SessionIDLookup interface {
	GetByPublicID(ctx context.Context, issuer, pid string) (found bool, err error)
}

// NewSessionIDCollector returns a new *SessionIDCollector using the given configuration, storage, and lookup, which
// waits for the confirm duration before looking up a session which appeared to be absent a second time.
func NewSessionIDCollector(config *schema.IdentityProvidersOpenIDConnect, storage SessionIDStorage, lookup SessionIDLookup, confirm time.Duration) *SessionIDCollector {
	return &SessionIDCollector{config: config, storage: storage, lookup: lookup, confirm: confirm}
}

// SessionIDCollector ends the OpenID Connect sessions whose session no longer exists, such as one which expired or
// was ended through inactivity, as if the End-User logged out. The session store is authoritative, so a session is
// ended only when the store reports it definitively absent on two lookups separated by the confirm duration. A pass
// aborts on any lookup error, since an unreachable backend must never be read as every session having ended, and
// resumes from the point it stopped on the next pass. A failure to end a session doesn't abort the pass, as the
// mapping remains and the session is ended by a later pass.
type SessionIDCollector struct {
	config  *schema.IdentityProvidersOpenIDConnect
	storage SessionIDStorage
	lookup  SessionIDLookup
	confirm time.Duration

	// cursor is the exclusive lower bound the next Collect call resumes from. It is reset to zero once a pass reaches
	// the end of the table, which is what turns repeated calls into a sweep rather than a single one-shot pass.
	cursor int
}

// Collect walks the mappings oldest first in pages of pageSize, ending the OpenID Connect sessions of every one whose
// session is definitively absent, until it has examined at least maxRows rows or reached the end of the table. It
// returns the number of OpenID Connect sessions ended along with any errors.
func (c *SessionIDCollector) Collect(ctx context.Context, pageSize, maxRows int) (ended int, err error) {
	after := c.cursor

	var (
		processed int
		errs      []error
	)

	for {
		var records []model.OAuth2SessionID

		if records, err = c.storage.LoadOAuth2SessionIDsOldest(ctx, after, pageSize); err != nil {
			c.cursor = after

			return ended, errors.Join(append(errs, err)...)
		}

		var n int

		if n, err = c.collectPage(ctx, records, &errs); err != nil {
			c.cursor = after

			return ended + n, errors.Join(append(errs, err)...)
		}

		ended += n
		processed += len(records)

		if len(records) < pageSize {
			c.cursor = 0

			return ended, errors.Join(errs...)
		}

		last := records[len(records)-1].ID

		if last <= after {
			c.cursor = after

			return ended, errors.Join(append(errs, fmt.Errorf("session id collector cursor did not advance past %d", after))...)
		}

		after = last

		if processed >= maxRows {
			c.cursor = after

			return ended, errors.Join(errs...)
		}
	}
}

func (c *SessionIDCollector) collectPage(ctx context.Context, records []model.OAuth2SessionID, errs *[]error) (ended int, err error) {
	candidates, err := c.absent(ctx, records)
	if err != nil || len(candidates) == 0 {
		return 0, err
	}

	if c.confirm > 0 {
		timer := time.NewTimer(c.confirm)

		select {
		case <-ctx.Done():
			timer.Stop()

			return 0, ctx.Err()
		case <-timer.C:
		}
	}

	if candidates, err = c.absent(ctx, candidates); err != nil {
		return 0, err
	}

	seen := map[[2]string]bool{}

	for _, record := range candidates {
		key := [2]string{record.Issuer, record.PublicID}

		if seen[key] {
			continue
		}

		seen[key] = true

		var sids []string

		if sids, err = c.storage.EndOAuth2Session(ctx, record.Issuer, record.PublicID, NewLogoutRevokeSchedule(c.config, time.Now())); err != nil {
			*errs = append(*errs, err)

			continue
		}

		for _, sid := range sids {
			if err = c.storage.RevokeOAuth2RefreshTokenSessionsBySessionID(ctx, sid); err != nil {
				*errs = append(*errs, err)
			}
		}

		ended += len(sids)
	}

	return ended, nil
}

func (c *SessionIDCollector) absent(ctx context.Context, records []model.OAuth2SessionID) (absent []model.OAuth2SessionID, err error) {
	for _, record := range records {
		var found bool

		if found, err = c.lookup.GetByPublicID(ctx, record.Issuer, record.PublicID); err != nil {
			return nil, err
		}

		if !found {
			absent = append(absent, record)
		}
	}

	return absent, nil
}

// GarbageCollection implements the service.GarbageCollectorProvider interface.
func (c *SessionIDCollector) GarbageCollection(ctx context.Context) (err error) {
	_, err = c.Collect(ctx, sessionIDCollectionBatch, sessionIDCollectionLimit)

	return err
}

// GarbageCollectionFrequency implements the service.GarbageCollectorProvider interface. It bounds how long the online
// refresh tokens of a session which ended without the End-User logging out remain usable.
func (c *SessionIDCollector) GarbageCollectionFrequency(_ context.Context) (frequency time.Duration) {
	return sessionIDCollectionFrequency
}
