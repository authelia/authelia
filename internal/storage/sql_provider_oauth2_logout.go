// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/model"
)

// SaveOAuth2LogoutSession saves a logout session awaiting the confirmation of the End-User, removing every logout
// session which expired before it was created.
func (p *SQLProvider) SaveOAuth2LogoutSession(ctx context.Context, session model.OAuth2LogoutSession) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.deleteLogoutSessionsExpired, session.CreatedAt); err != nil {
		return fmt.Errorf("error deleting expired oauth2 logout sessions: %w", err)
	}

	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.insertLogoutSession,
		session.FlowID, session.ClientID, session.Subject, session.SessionID, session.RedirectURI, session.State, session.CreatedAt, session.ExpiresAt); err != nil {
		return fmt.Errorf("error inserting oauth2 logout session with flow id '%s': %w", session.FlowID, err)
	}

	return nil
}

// LoadOAuth2LogoutSession returns the logout session with the flow id, or a nil session without an error when there is
// no such session.
func (p *SQLProvider) LoadOAuth2LogoutSession(ctx context.Context, flowID uuid.UUID) (session *model.OAuth2LogoutSession, err error) {
	session = &model.OAuth2LogoutSession{}

	if err = p.conn(ctx).GetContext(ctx, session, p.sqlOAuth2Logout.selectLogoutSession, flowID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting oauth2 logout session with flow id '%s': %w", flowID, err)
	}

	return session, nil
}

// DeleteOAuth2LogoutSession deletes the logout session with the flow id.
func (p *SQLProvider) DeleteOAuth2LogoutSession(ctx context.Context, flowID uuid.UUID) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.deleteLogoutSession, flowID); err != nil {
		return fmt.Errorf("error deleting oauth2 logout session with flow id '%s': %w", flowID, err)
	}

	return nil
}

// LoadOAuth2SessionIDsByPublicID returns every sector's session id mapping for the issuer and session public identifier.
func (p *SQLProvider) LoadOAuth2SessionIDsByPublicID(ctx context.Context, issuer, publicID string) (records []model.OAuth2SessionID, err error) {
	if err = p.conn(ctx).SelectContext(ctx, &records, p.sqlOAuth2Logout.selectSessionIDsByPublicID, issuer, publicID); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 session ids by public id: %w", err)
	}

	return records, nil
}

// RevokeOAuth2SessionsBySessionID revokes every OAuth 2.0 session issued with the 'sid' claim value, except the refresh
// token sessions which were granted offline access.
func (p *SQLProvider) RevokeOAuth2SessionsBySessionID(ctx context.Context, sid string) (err error) {
	for _, q := range p.sqlOAuth2Logout.revokeBySessionID {
		if _, err = p.conn(ctx).ExecContext(ctx, q.query, sid); err != nil {
			return fmt.Errorf("error revoking oauth2 sessions from table '%s' with session id '%s': %w", q.table, sid, err)
		}
	}

	return p.revokeOAuth2RefreshTokenSessionsOnline(ctx, fmt.Sprintf("session id '%s'", sid), p.sqlOAuth2Logout.selectRefreshBySessionID, sid)
}

func (p *SQLProvider) revokeOAuth2RefreshTokenSessionsOnline(ctx context.Context, description, query string, args ...any) (err error) {
	var sessions []oauth2LogoutRefreshTokenSession

	if err = p.conn(ctx).SelectContext(ctx, &sessions, query, args...); err != nil {
		return fmt.Errorf("error selecting oauth2 refresh token sessions with %s: %w", description, err)
	}

	for _, session := range sessions {
		if session.offline() {
			continue
		}

		if _, err = p.conn(ctx).ExecContext(ctx, p.sqlRevokeOAuth2RefreshTokenSession, session.Signature); err != nil {
			return fmt.Errorf("error revoking oauth2 refresh token session with %s: %w", description, err)
		}
	}

	return nil
}

var oauth2LogoutOfflineScopes = []string{"offline", "offline_access"}

type sqlOAuth2LogoutQuery struct {
	table string
	query string
}

type sqlOAuth2LogoutQueries struct {
	revokeBySessionID []sqlOAuth2LogoutQuery

	selectRefreshBySessionID   string
	selectSessionIDsByPublicID string

	insertLogoutSession         string
	selectLogoutSession         string
	deleteLogoutSession         string
	deleteLogoutSessionsExpired string
}

func newSQLOAuth2LogoutQueries(rebind func(query string) string) (queries sqlOAuth2LogoutQueries) {
	for _, table := range []string{tableOAuth2AccessTokenSession, tableOAuth2AuthorizeCodeSession, tableOAuth2DeviceCodeSession, tableOAuth2OpenIDConnectSession, tableOAuth2PARContext, tableOAuth2PKCERequestSession} {
		queries.revokeBySessionID = append(queries.revokeBySessionID, sqlOAuth2LogoutQuery{table, rebind(fmt.Sprintf(queryFmtRevokeOAuth2SessionBySessionID, table))})
	}

	queries.selectRefreshBySessionID = rebind(fmt.Sprintf(queryFmtSelectOAuth2RefreshTokenSessionScopesBySessionID, tableOAuth2RefreshTokenSession))
	queries.selectSessionIDsByPublicID = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionIDsByPublicID, tableOAuth2SessionID))

	queries.insertLogoutSession = rebind(fmt.Sprintf(queryFmtInsertOAuth2LogoutSession, tableOAuth2LogoutSession))
	queries.selectLogoutSession = rebind(fmt.Sprintf(queryFmtSelectOAuth2LogoutSession, tableOAuth2LogoutSession))
	queries.deleteLogoutSession = rebind(fmt.Sprintf(queryFmtDeleteOAuth2LogoutSession, tableOAuth2LogoutSession))
	queries.deleteLogoutSessionsExpired = rebind(fmt.Sprintf(queryFmtDeleteOAuth2LogoutSessionsExpired, tableOAuth2LogoutSession))

	return queries
}

type oauth2LogoutRefreshTokenSession struct {
	Signature     string                         `db:"signature"`
	GrantedScopes model.StringSlicePipeDelimited `db:"granted_scopes"`
}

func (s oauth2LogoutRefreshTokenSession) offline() bool {
	for _, scope := range s.GrantedScopes {
		for _, offline := range oauth2LogoutOfflineScopes {
			if scope == offline {
				return true
			}
		}
	}

	return false
}
