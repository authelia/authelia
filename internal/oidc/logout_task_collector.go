// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/model"
)

const (
	logoutTaskCollectionFrequency = time.Minute
	logoutTaskCollectionLimit     = 500

	logoutTaskBackoffMin = time.Minute
	logoutTaskBackoffMax = time.Hour
)

// LogoutTaskStorage is the storage the LogoutTaskCollector needs to process the logout tasks.
type LogoutTaskStorage interface {
	LoadOAuth2SessionLogoutTasksDue(ctx context.Context, now time.Time, limit int) (tasks []model.OAuth2SessionLogoutTask, err error)
	ClaimOAuth2SessionLogoutTask(ctx context.Context, id, attempts int, next time.Time) (claimed bool, err error)
	RescheduleOAuth2SessionLogoutTask(ctx context.Context, id, attempts int, next time.Time) (rescheduled bool, err error)
	DeleteOAuth2SessionLogoutTask(ctx context.Context, id, attempts int) (deleted bool, err error)
	RevokeOAuth2RefreshTokenSessionsBySessionID(ctx context.Context, sid string) (err error)
}

// NewLogoutTaskCollector returns a new *LogoutTaskCollector using the given storage and clock.
func NewLogoutTaskCollector(storage LogoutTaskStorage, clock clock.Provider) *LogoutTaskCollector {
	return &LogoutTaskCollector{storage: storage, clock: clock}
}

// LogoutTaskCollector processes the logout tasks whose next attempt is due. Each attempt is claimed first, which
// schedules the next attempt with an increasing backoff, so a task whose attempt fails, or whose instance stops during
// the attempt, is attempted again, and a task is never attempted by more than one instance at a time. A task of the
// kind model.OAuth2SessionLogoutTaskKindRevoke revokes the online refresh tokens of its session, and is kept until it expires
// so its session remains ended until every authorization code and device code issued to it has expired.
type LogoutTaskCollector struct {
	storage LogoutTaskStorage
	clock   clock.Provider
}

// Collect processes up to limit logout tasks whose next attempt is due, returning the number processed successfully
// along with any errors.
func (c *LogoutTaskCollector) Collect(ctx context.Context, limit int) (processed int, err error) {
	now := c.clock.Now()

	var tasks []model.OAuth2SessionLogoutTask

	if tasks, err = c.storage.LoadOAuth2SessionLogoutTasksDue(ctx, now, limit); err != nil {
		return 0, err
	}

	var errs []error

	for _, task := range tasks {
		if err = c.process(ctx, now, task); err != nil {
			errs = append(errs, fmt.Errorf("error processing oauth2 session logout task of kind '%s' with session id '%s': %w", task.Kind, task.SessionID, err))

			continue
		}

		processed++
	}

	return processed, errors.Join(errs...)
}

func (c *LogoutTaskCollector) process(ctx context.Context, now time.Time, task model.OAuth2SessionLogoutTask) (err error) {
	var claimed bool

	if claimed, err = c.storage.ClaimOAuth2SessionLogoutTask(ctx, task.ID, task.Attempts, now.Add(logoutTaskBackoff(task.Attempts))); err != nil || !claimed {
		return err
	}

	attempts := task.Attempts + 1

	if task.Kind == model.OAuth2SessionLogoutTaskKindRevoke {
		if err = c.storage.RevokeOAuth2RefreshTokenSessionsBySessionID(ctx, task.SessionID); err != nil {
			return err
		}
	}

	if now.Before(task.ExpiresAt) {
		_, err = c.storage.RescheduleOAuth2SessionLogoutTask(ctx, task.ID, attempts, task.ExpiresAt)
	} else {
		_, err = c.storage.DeleteOAuth2SessionLogoutTask(ctx, task.ID, attempts)
	}

	return err
}

func logoutTaskBackoff(attempts int) (backoff time.Duration) {
	backoff = logoutTaskBackoffMin

	for range attempts {
		if backoff >= logoutTaskBackoffMax {
			break
		}

		backoff *= 2
	}

	return min(backoff, logoutTaskBackoffMax)
}

// GarbageCollection implements the service.GarbageCollectorProvider interface.
func (c *LogoutTaskCollector) GarbageCollection(ctx context.Context) (err error) {
	_, err = c.Collect(ctx, logoutTaskCollectionLimit)

	return err
}

// GarbageCollectionFrequency implements the service.GarbageCollectorProvider interface.
func (c *LogoutTaskCollector) GarbageCollectionFrequency(_ context.Context) (frequency time.Duration) {
	return logoutTaskCollectionFrequency
}
