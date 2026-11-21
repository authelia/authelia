// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/model"
)

// GetOrCreateOAuth2SessionID returns the session id mapping for the issuer, sector and public identifier, creating it
// when absent. It is safe against a concurrent creator, which the unique index rejects, by re-selecting on conflict.
func (p *SQLProvider) GetOrCreateOAuth2SessionID(ctx context.Context, issuer, sectorID, publicID string) (record *model.OAuth2SessionID, err error) {
	if record, err = p.selectOAuth2SessionIDBySector(ctx, issuer, sectorID, publicID); err != nil {
		return nil, err
	} else if record != nil {
		return record, nil
	}

	var sid uuid.UUID

	if sid, err = uuid.NewRandom(); err != nil {
		return nil, fmt.Errorf("error generating session id: %w", err)
	}

	created := time.Now()

	if _, errInsert := p.conn(ctx).ExecContext(ctx, p.sqlInsertOAuth2SessionID, issuer, sectorID, publicID, sid.String(), created); errInsert != nil {
		if record, err = p.selectOAuth2SessionIDBySector(ctx, issuer, sectorID, publicID); err != nil {
			return nil, err
		} else if record != nil {
			return record, nil
		}

		return nil, fmt.Errorf("error inserting oauth2 session id: %w", errInsert)
	}

	return &model.OAuth2SessionID{Issuer: issuer, SectorID: sectorID, PublicID: publicID, SessionID: sid, CreatedAt: created}, nil
}

func (p *SQLProvider) selectOAuth2SessionIDBySector(ctx context.Context, issuer, sectorID, publicID string) (record *model.OAuth2SessionID, err error) {
	record = &model.OAuth2SessionID{}

	if err = p.conn(ctx).GetContext(ctx, record, p.sqlSelectOAuth2SessionIDBySector, issuer, sectorID, publicID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting oauth2 session id by sector: %w", err)
	}

	return record, nil
}

// LoadOAuth2SessionIDBySessionID returns the mapping for the given 'sid', or a nil record without an error when there
// is no such mapping.
func (p *SQLProvider) LoadOAuth2SessionIDBySessionID(ctx context.Context, issuer, sid string) (record *model.OAuth2SessionID, err error) {
	record = &model.OAuth2SessionID{}

	if err = p.conn(ctx).GetContext(ctx, record, p.sqlSelectOAuth2SessionIDBySessionID, issuer, sid); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting oauth2 session id: %w", err)
	}

	return record, nil
}

// LoadOAuth2SessionIDsOldest pages the mappings by ascending id, which the liveness sweep walks.
func (p *SQLProvider) LoadOAuth2SessionIDsOldest(ctx context.Context, after, limit int) (records []model.OAuth2SessionID, err error) {
	if err = p.conn(ctx).SelectContext(ctx, &records, p.sqlSelectOAuth2SessionIDsOldest, after, limit); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 session ids: %w", err)
	}

	return records, nil
}

// SaveOAuth2SessionIDClient records a client as participating in the session identified by the 'sid' with the subject
// identifier it was issued. The mapping of the session is read with a shared lock, so a logout which ends the session
// either waits for the client to be recorded, which makes it a client the logout notifies, or ends the session first,
// in which case ErrOAuth2SessionEnded is returned. Recording the same client for the same 'sid' more than once is not
// an error.
func (p *SQLProvider) SaveOAuth2SessionIDClient(ctx context.Context, issuer, publicID, sid, clientID, subject string) (err error) {
	return p.transaction(ctx, func(ctx context.Context) (err error) {
		var found bool

		if found, err = p.lockOAuth2SessionIDs(ctx, sid, p.sqlOAuth2Logout.selectSessionIDIDByIssuerSessionID, sid, issuer); err != nil {
			return err
		}

		if !found {
			return ErrOAuth2SessionEnded
		}

		if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.insertSessionIDClient, issuer, publicID, sid, clientID, subject, time.Now()); err != nil {
			return fmt.Errorf("error inserting oauth2 session id client: %w", err)
		}

		return nil
	})
}
