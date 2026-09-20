// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package storage

import (
	"context"
	"fmt"

	"github.com/authelia/authelia/v4/internal/model"
)

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

// RevokeOAuth2SessionsByClientIDAndSubject revokes every OAuth 2.0 session of the client for the subject, except the
// refresh token sessions which were granted offline access.
func (p *SQLProvider) RevokeOAuth2SessionsByClientIDAndSubject(ctx context.Context, clientID, subject string) (err error) {
	for _, q := range p.sqlOAuth2Logout.revokeByClientIDAndSubject {
		if _, err = p.conn(ctx).ExecContext(ctx, q.query, clientID, subject); err != nil {
			return fmt.Errorf("error revoking oauth2 sessions from table '%s' with client id '%s' and subject '%s': %w", q.table, clientID, subject, err)
		}
	}

	return p.revokeOAuth2RefreshTokenSessionsOnline(ctx, fmt.Sprintf("client id '%s' and subject '%s'", clientID, subject), p.sqlOAuth2Logout.selectRefreshByClientIDAndSubject, clientID, subject)
}

// HasOAuth2SessionsByClientIDAndSubject returns true when the client holds any OpenID Connect, refresh token, or access
// token session for the subject which hasn't been revoked.
func (p *SQLProvider) HasOAuth2SessionsByClientIDAndSubject(ctx context.Context, clientID, subject string) (has bool, err error) {
	var count int

	for _, q := range p.sqlOAuth2Logout.existsByClientIDAndSubject {
		if err = p.conn(ctx).GetContext(ctx, &count, q.query, clientID, subject); err != nil {
			return false, fmt.Errorf("error selecting oauth2 sessions from table '%s' with client id '%s' and subject '%s': %w", q.table, clientID, subject, err)
		}

		if count != 0 {
			return true, nil
		}
	}

	return false, nil
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
	revokeBySessionID          []sqlOAuth2LogoutQuery
	revokeByClientIDAndSubject []sqlOAuth2LogoutQuery
	existsByClientIDAndSubject []sqlOAuth2LogoutQuery

	selectRefreshBySessionID          string
	selectRefreshByClientIDAndSubject string
	selectSessionIDsByPublicID        string
}

func newSQLOAuth2LogoutQueries(rebind func(query string) string) (queries sqlOAuth2LogoutQueries) {
	for _, table := range []string{tableOAuth2AccessTokenSession, tableOAuth2AuthorizeCodeSession, tableOAuth2DeviceCodeSession, tableOAuth2OpenIDConnectSession, tableOAuth2PARContext, tableOAuth2PKCERequestSession} {
		queries.revokeBySessionID = append(queries.revokeBySessionID, sqlOAuth2LogoutQuery{table, rebind(fmt.Sprintf(queryFmtRevokeOAuth2SessionBySessionID, table))})
	}

	for _, table := range []string{tableOAuth2AccessTokenSession, tableOAuth2AuthorizeCodeSession, tableOAuth2DeviceCodeSession, tableOAuth2OpenIDConnectSession, tableOAuth2PKCERequestSession} {
		queries.revokeByClientIDAndSubject = append(queries.revokeByClientIDAndSubject, sqlOAuth2LogoutQuery{table, rebind(fmt.Sprintf(queryFmtRevokeOAuth2SessionByClientIDAndSubject, table))})
	}

	for _, table := range []string{tableOAuth2OpenIDConnectSession, tableOAuth2RefreshTokenSession, tableOAuth2AccessTokenSession} {
		queries.existsByClientIDAndSubject = append(queries.existsByClientIDAndSubject, sqlOAuth2LogoutQuery{table, rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionExistsByClientIDAndSubject, table))})
	}

	queries.selectRefreshBySessionID = rebind(fmt.Sprintf(queryFmtSelectOAuth2RefreshTokenSessionScopesBySessionID, tableOAuth2RefreshTokenSession))
	queries.selectRefreshByClientIDAndSubject = rebind(fmt.Sprintf(queryFmtSelectOAuth2RefreshTokenSessionScopesByClientIDAndSubject, tableOAuth2RefreshTokenSession))
	queries.selectSessionIDsByPublicID = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionIDsByPublicID, tableOAuth2SessionID))

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
