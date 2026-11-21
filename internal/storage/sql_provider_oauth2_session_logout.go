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

// SaveOAuth2SessionLogout saves a logout session awaiting the confirmation of the End-User, removing every logout
// session which expired before it was created.
func (p *SQLProvider) SaveOAuth2SessionLogout(ctx context.Context, session model.OAuth2SessionLogout) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.deleteLogoutSessionsExpired, session.CreatedAt); err != nil {
		return fmt.Errorf("error deleting expired oauth2 session logouts: %w", err)
	}

	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.insertLogoutSession,
		session.FlowID, session.ClientID, session.Subject, session.SessionID, session.RedirectURI, session.State, session.CreatedAt, session.ExpiresAt); err != nil {
		return fmt.Errorf("error inserting oauth2 session logout with flow id '%s': %w", session.FlowID, err)
	}

	return nil
}

// LoadOAuth2SessionLogout returns the logout session with the flow id, or a nil session without an error when there is
// no such session.
func (p *SQLProvider) LoadOAuth2SessionLogout(ctx context.Context, flowID uuid.UUID) (session *model.OAuth2SessionLogout, err error) {
	session = &model.OAuth2SessionLogout{}

	if err = p.conn(ctx).GetContext(ctx, session, p.sqlOAuth2Logout.selectLogoutSession, flowID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("error selecting oauth2 session logout with flow id '%s': %w", flowID, err)
	}

	return session, nil
}

// DeleteOAuth2SessionLogout deletes the logout session with the flow id, returning false when there was no such
// session, such as when another request consumed it.
func (p *SQLProvider) DeleteOAuth2SessionLogout(ctx context.Context, flowID uuid.UUID) (deleted bool, err error) {
	return p.updateSingle(ctx, fmt.Sprintf("deleting oauth2 session logout with flow id '%s'", flowID), p.sqlOAuth2Logout.deleteLogoutSession, flowID)
}

// EndOAuth2Session ends every OpenID Connect session of the session public identifier, returning their session
// identifiers. In one transaction it locks their mappings, records each ended session with a task of the kind
// model.OAuth2SessionLogoutTaskKindRevoke using the schedule, creates a task of the kind
// model.OAuth2SessionLogoutTaskKindBackChannel for each participating client when the schedule has a back-channel
// schedule, and removes the mappings and their participants. A concurrent call for the same sessions waits for the lock
// and then finds no mappings, so each session is only ended once.
func (p *SQLProvider) EndOAuth2Session(ctx context.Context, issuer, publicID string, schedule model.OAuth2SessionLogoutTaskSchedule) (sids []string, err error) {
	if err = p.transaction(ctx, func(ctx context.Context) (err error) {
		var records []model.OAuth2SessionID

		if err = p.conn(ctx).SelectContext(ctx, &records, p.sqlOAuth2Logout.selectSessionIDsByPublicIDLocked, issuer, publicID); err != nil {
			return fmt.Errorf("error selecting oauth2 session ids by public id: %w", err)
		}

		sids = make([]string, 0, len(records))

		for _, record := range records {
			sid := record.SessionID.String()

			if err = p.SaveOAuth2SessionLogoutTask(ctx, schedule.RevokeTask(sid)); err != nil {
				return err
			}

			if bc := schedule.BackChannel; bc != nil {
				if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.insertBackChannelTasks,
					model.OAuth2SessionLogoutTaskKindBackChannel, bc.IssuerURL, bc.NextAttemptAt, schedule.CreatedAt, bc.ExpiresAt, issuer, sid); err != nil {
					return fmt.Errorf("error inserting oauth2 session logout tasks of kind '%s' with session id '%s': %w", model.OAuth2SessionLogoutTaskKindBackChannel, sid, err)
				}
			}

			if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.deleteSessionIDClientsBySessionID, issuer, sid); err != nil {
				return fmt.Errorf("error deleting oauth2 session id clients with session id '%s': %w", sid, err)
			}

			if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.deleteSessionIDByID, record.ID); err != nil {
				return fmt.Errorf("error deleting oauth2 session id with session id '%s': %w", sid, err)
			}

			sids = append(sids, sid)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return sids, nil
}

// SaveOAuth2SessionLogoutTask saves a logout task unless a task of the same kind for the same session and client exists.
func (p *SQLProvider) SaveOAuth2SessionLogoutTask(ctx context.Context, task model.OAuth2SessionLogoutTask) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.insertLogoutTask,
		task.Kind, task.SessionID, task.ClientID, task.Subject, task.IssuerURL, task.NextAttemptAt, task.CreatedAt, task.ExpiresAt); err != nil {
		return fmt.Errorf("error inserting oauth2 session logout task of kind '%s' with session id '%s': %w", task.Kind, task.SessionID, err)
	}

	return nil
}

// IsOAuth2SessionEnded returns true if the OpenID Connect session with the session identifier has ended. Within a
// transaction the mapping of the session and the record are read with a shared lock. A logout which ends the session
// locks the mapping exclusively, so a concurrent logout either waits for the transaction to finish, or finishes first
// and its record is observed.
func (p *SQLProvider) IsOAuth2SessionEnded(ctx context.Context, sid string) (ended bool, err error) {
	query := p.sqlOAuth2Logout.selectSessionEnded

	if _, ok := ctx.Value(ctxKeyTransaction).(SQLXTx); ok {
		query = p.sqlOAuth2Logout.selectSessionEndedLocked

		if p.sqlOAuth2Logout.lockSessionIDByID != "" {
			if _, err = p.lockOAuth2SessionIDs(ctx, sid, p.sqlOAuth2Logout.selectSessionIDIDsBySessionID, sid); err != nil {
				return false, err
			}
		}
	}

	var id int

	if err = p.conn(ctx).GetContext(ctx, &id, query, sid, model.OAuth2SessionLogoutTaskKindRevoke); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}

		return false, fmt.Errorf("error selecting oauth2 session logout task with session id '%s': %w", sid, err)
	}

	return true, nil
}

func (p *SQLProvider) lockOAuth2SessionIDsByCode(ctx context.Context, query, signature string) (err error) {
	if _, ok := ctx.Value(ctxKeyTransaction).(SQLXTx); !ok || p.sqlOAuth2Logout.lockSessionIDByID == "" {
		return nil
	}

	var sids []sql.NullString

	if err = p.conn(ctx).SelectContext(ctx, &sids, query, signature); err != nil {
		return fmt.Errorf("error selecting the session id: %w", err)
	}

	for _, sid := range sids {
		if sid.Valid && sid.String != "" {
			if _, err = p.lockOAuth2SessionIDs(ctx, sid.String, p.sqlOAuth2Logout.selectSessionIDIDsBySessionID, sid.String); err != nil {
				return err
			}
		}
	}

	return nil
}

// LoadOAuth2SessionLogoutTasksBySessionID returns the logout tasks of the kind for the session identifier.
func (p *SQLProvider) LoadOAuth2SessionLogoutTasksBySessionID(ctx context.Context, sid, kind string) (tasks []model.OAuth2SessionLogoutTask, err error) {
	if err = p.conn(ctx).SelectContext(ctx, &tasks, p.sqlOAuth2Logout.selectTasksBySessionID, sid, kind); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 session logout tasks of kind '%s' with session id '%s': %w", kind, sid, err)
	}

	return tasks, nil
}

func (p *SQLProvider) lockOAuth2SessionIDs(ctx context.Context, sid, query string, args ...any) (found bool, err error) {
	var ids []int

	if err = p.conn(ctx).SelectContext(ctx, &ids, query, args...); err != nil {
		return false, fmt.Errorf("error selecting oauth2 session id with session id '%s': %w", sid, err)
	}

	if p.sqlOAuth2Logout.lockSessionIDByID == "" {
		return len(ids) != 0, nil
	}

	for _, id := range ids {
		var locked []int

		if err = p.conn(ctx).SelectContext(ctx, &locked, p.sqlOAuth2Logout.lockSessionIDByID, id); err != nil {
			return false, fmt.Errorf("error locking oauth2 session id with session id '%s': %w", sid, err)
		}

		found = found || len(locked) != 0
	}

	return found, nil
}

// LoadOAuth2SessionLogoutTasksDue returns the logout tasks whose next attempt is due at the time, earliest first.
func (p *SQLProvider) LoadOAuth2SessionLogoutTasksDue(ctx context.Context, now time.Time, limit int) (tasks []model.OAuth2SessionLogoutTask, err error) {
	if err = p.conn(ctx).SelectContext(ctx, &tasks, p.sqlOAuth2Logout.selectLogoutTasksDue, now, limit); err != nil {
		return nil, fmt.Errorf("error selecting oauth2 session logout tasks: %w", err)
	}

	return tasks, nil
}

// ClaimOAuth2SessionLogoutTask claims the attempt of a logout task loaded with the number of attempts, incrementing it and
// scheduling the next attempt for when this one is presumed to have failed. It returns false when another attempt
// claimed the task first. The incremented number of attempts identifies the claim to RescheduleOAuth2SessionLogoutTask and
// DeleteOAuth2SessionLogoutTask.
func (p *SQLProvider) ClaimOAuth2SessionLogoutTask(ctx context.Context, id, attempts int, next time.Time) (claimed bool, err error) {
	return p.updateSingle(ctx, "claiming oauth2 session logout task", p.sqlOAuth2Logout.claimLogoutTask, next, id, attempts)
}

// RescheduleOAuth2SessionLogoutTask schedules the next attempt of a claimed logout task. It returns false when the claim was
// superseded by another attempt.
func (p *SQLProvider) RescheduleOAuth2SessionLogoutTask(ctx context.Context, id, attempts int, next time.Time) (rescheduled bool, err error) {
	return p.updateSingle(ctx, "rescheduling oauth2 session logout task", p.sqlOAuth2Logout.rescheduleLogoutTask, next, id, attempts)
}

// DeleteOAuth2SessionLogoutTask deletes a claimed logout task. It returns false when the claim was superseded by another
// attempt.
func (p *SQLProvider) DeleteOAuth2SessionLogoutTask(ctx context.Context, id, attempts int) (deleted bool, err error) {
	return p.updateSingle(ctx, "deleting oauth2 session logout task", p.sqlOAuth2Logout.deleteLogoutTask, id, attempts)
}

func (p *SQLProvider) updateSingle(ctx context.Context, action, query string, args ...any) (updated bool, err error) {
	var result sql.Result

	if result, err = p.conn(ctx).ExecContext(ctx, query, args...); err != nil {
		return false, fmt.Errorf("error %s: %w", action, err)
	}

	switch err = checkSingleUpdateResult(result); {
	case err == nil:
		return true, nil
	case errors.Is(err, ErrNoRowsAffected):
		return false, nil
	default:
		return false, fmt.Errorf("error %s: %w", action, err)
	}
}

// RevokeOAuth2RefreshTokenSessionsBySessionID revokes the refresh token sessions issued with the 'sid' claim value
// which were not recorded as offline refresh tokens when issued. A refresh token grant can't use an online refresh token
// of an ended session, so once the session has ended the refresh token sessions revoked are the last which are issued
// except for those issued by a refresh token grant which was already in progress, which a later attempt revokes.
func (p *SQLProvider) RevokeOAuth2RefreshTokenSessionsBySessionID(ctx context.Context, sid string) (err error) {
	if _, err = p.conn(ctx).ExecContext(ctx, p.sqlOAuth2Logout.revokeRefreshOnlineBySessionID, sid); err != nil {
		return fmt.Errorf("error revoking oauth2 refresh token sessions with session id '%s': %w", sid, err)
	}

	return nil
}

func (p *SQLProvider) transaction(ctx context.Context, fn func(ctx context.Context) (err error)) (err error) {
	if _, ok := ctx.Value(ctxKeyTransaction).(SQLXTx); ok {
		return fn(ctx)
	}

	var tx context.Context

	if tx, err = p.BeginTX(ctx); err != nil {
		return fmt.Errorf("error beginning transaction: %w", err)
	}

	if err = fn(tx); err != nil {
		if errRollback := p.Rollback(tx); errRollback != nil {
			return fmt.Errorf("%w: error rolling back transaction: %v", err, errRollback)
		}

		return err
	}

	if err = p.Commit(tx); err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

type sqlOAuth2LogoutQueries struct {
	revokeRefreshOnlineBySessionID    string
	selectSessionIDsByPublicIDLocked  string
	deleteSessionIDByID               string
	deleteSessionIDClientsBySessionID string

	insertLogoutSession         string
	selectLogoutSession         string
	deleteLogoutSession         string
	deleteLogoutSessionsExpired string

	insertLogoutTask                   string
	selectSessionEnded                 string
	selectSessionEndedLocked           string
	selectSessionIDIDsBySessionID      string
	selectSessionIDIDByIssuerSessionID string
	insertSessionIDClient              string
	insertBackChannelTasks             string
	selectTasksBySessionID             string
	lockSessionIDByID                  string
	selectLogoutTasksDue               string
	claimLogoutTask                    string
	rescheduleLogoutTask               string
	deleteLogoutTask                   string

	deactivateAuthorizeCodeSession string
	deactivateDeviceCodeSession    string
	selectAuthorizeCodeSessionID   string
	selectDeviceCodeSessionID      string
}

func newSQLOAuth2LogoutQueries(name string, rebind func(query string) string) (queries sqlOAuth2LogoutQueries) {
	var lockUpdate, lockShare, insertTask, insertClient, insertBackChannel string

	switch name {
	case providerMySQL:
		lockUpdate, lockShare = " FOR UPDATE", " LOCK IN SHARE MODE"
		insertTask, insertClient, insertBackChannel = queryFmtInsertOAuth2SessionLogoutTaskMySQL, queryFmtInsertOAuth2SessionIDClientMySQL, queryFmtInsertOAuth2SessionLogoutTasksBackChannelMySQL
	case providerPostgres:
		lockUpdate, lockShare = " FOR UPDATE", " FOR SHARE"

		fallthrough
	default:
		insertTask, insertClient, insertBackChannel = queryFmtInsertOAuth2SessionLogoutTask, queryFmtInsertOAuth2SessionIDClient, queryFmtInsertOAuth2SessionLogoutTasksBackChannel
	}

	queries.revokeRefreshOnlineBySessionID = rebind(fmt.Sprintf(queryFmtRevokeOAuth2RefreshTokenSessionsOnlineBySessionID, tableOAuth2RefreshTokenSession))
	queries.selectSessionIDsByPublicIDLocked = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionIDsByPublicIDLock, tableOAuth2SessionID, lockUpdate))
	queries.deleteSessionIDByID = rebind(fmt.Sprintf(queryFmtDeleteOAuth2SessionIDByID, tableOAuth2SessionID))
	queries.deleteSessionIDClientsBySessionID = rebind(fmt.Sprintf(queryFmtDeleteOAuth2SessionIDClientsBySessionID, tableOAuth2SessionIDClient))

	queries.insertLogoutSession = rebind(fmt.Sprintf(queryFmtInsertOAuth2SessionLogout, tableOAuth2SessionLogout))
	queries.selectLogoutSession = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionLogout, tableOAuth2SessionLogout))
	queries.deleteLogoutSession = rebind(fmt.Sprintf(queryFmtDeleteOAuth2SessionLogout, tableOAuth2SessionLogout))
	queries.deleteLogoutSessionsExpired = rebind(fmt.Sprintf(queryFmtDeleteOAuth2SessionLogoutExpired, tableOAuth2SessionLogout))

	queries.insertLogoutTask = rebind(fmt.Sprintf(insertTask, tableOAuth2SessionLogoutTask))
	queries.selectSessionEnded = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionLogoutTaskSessionEnded, tableOAuth2SessionLogoutTask, ""))
	queries.selectSessionEndedLocked = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionLogoutTaskSessionEnded, tableOAuth2SessionLogoutTask, lockShare))

	queries.selectSessionIDIDsBySessionID = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionIDIDsBySessionID, tableOAuth2SessionID))
	queries.selectSessionIDIDByIssuerSessionID = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionIDIDByIssuerSessionID, tableOAuth2SessionID))
	queries.insertSessionIDClient = rebind(fmt.Sprintf(insertClient, tableOAuth2SessionIDClient))
	queries.insertBackChannelTasks = rebind(fmt.Sprintf(insertBackChannel, tableOAuth2SessionLogoutTask, tableOAuth2SessionIDClient))
	queries.selectTasksBySessionID = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionLogoutTasksBySessionID, tableOAuth2SessionLogoutTask))

	if lockShare != "" {
		queries.lockSessionIDByID = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionIDByIDLock, tableOAuth2SessionID, lockShare))
	}

	queries.selectLogoutTasksDue = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionLogoutTasksDue, tableOAuth2SessionLogoutTask))
	queries.claimLogoutTask = rebind(fmt.Sprintf(queryFmtClaimOAuth2SessionLogoutTask, tableOAuth2SessionLogoutTask))
	queries.rescheduleLogoutTask = rebind(fmt.Sprintf(queryFmtRescheduleOAuth2SessionLogoutTask, tableOAuth2SessionLogoutTask))
	queries.deleteLogoutTask = rebind(fmt.Sprintf(queryFmtDeleteOAuth2SessionLogoutTask, tableOAuth2SessionLogoutTask))

	queries.deactivateAuthorizeCodeSession = rebind(fmt.Sprintf(queryFmtDeactivateOAuth2SessionNotEnded, tableOAuth2AuthorizeCodeSession, tableOAuth2SessionLogoutTask, model.OAuth2SessionLogoutTaskKindRevoke))
	queries.selectAuthorizeCodeSessionID = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionSessionIDBySignature, tableOAuth2AuthorizeCodeSession))
	queries.selectDeviceCodeSessionID = rebind(fmt.Sprintf(queryFmtSelectOAuth2SessionSessionIDBySignature, tableOAuth2DeviceCodeSession))
	queries.deactivateDeviceCodeSession = rebind(fmt.Sprintf(queryFmtDeactivateOAuth2SessionNotEnded, tableOAuth2DeviceCodeSession, tableOAuth2SessionLogoutTask, model.OAuth2SessionLogoutTaskKindRevoke))

	return queries
}
