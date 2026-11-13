// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package oidc

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/model"
)

const (
	logoutTaskCollectionFrequency = time.Minute
	logoutTaskCollectionLimit     = 500

	logoutTaskCleanupTimeout = 10 * time.Second

	logoutTaskBackoffMin = time.Minute
	logoutTaskBackoffMax = time.Hour
)

// LogoutTaskStorage is the storage the LogoutTaskCollector needs to process the logout tasks.
type LogoutTaskStorage interface {
	LoadOAuth2SessionLogoutTasksDue(ctx context.Context, now time.Time, limit int) (tasks []model.OAuth2SessionLogoutTask, err error)
	LoadOAuth2SessionLogoutTasksBySessionID(ctx context.Context, sid, kind string) (tasks []model.OAuth2SessionLogoutTask, err error)
	ClaimOAuth2SessionLogoutTask(ctx context.Context, id, attempts int, next time.Time) (claimed bool, err error)
	RescheduleOAuth2SessionLogoutTask(ctx context.Context, id, attempts int, next time.Time) (rescheduled bool, err error)
	DeleteOAuth2SessionLogoutTask(ctx context.Context, id, attempts int) (deleted bool, err error)
	RevokeOAuth2RefreshTokenSessionsBySessionID(ctx context.Context, sid string) (err error)
}

// LogoutTokenDeliverer delivers the Logout Tokens of the tasks of the kind model.OAuth2SessionLogoutTaskKindBackChannel.
type LogoutTokenDeliverer interface {
	AcquireBackChannelLogoutSlot(ctx context.Context) (release func(), err error)
	DeliverBackChannelLogout(ctx context.Context, task model.OAuth2SessionLogoutTask) (done bool, err error)
}

// NewLogoutTaskCollector returns a new *LogoutTaskCollector using the given storage, clock, and deliverer. When the
// deliverer is nil the tasks of the kind model.OAuth2SessionLogoutTaskKindBackChannel are removed once they expire.
func NewLogoutTaskCollector(storage LogoutTaskStorage, clock clock.Provider, deliverer LogoutTokenDeliverer) *LogoutTaskCollector {
	return &LogoutTaskCollector{storage: storage, clock: clock, deliverer: deliverer}
}

// LogoutTaskCollector processes the logout tasks whose next attempt is due. Each attempt is claimed first, which
// schedules the next attempt with an increasing backoff, so a task whose attempt fails, or whose instance stops during
// the attempt, is attempted again, and a task is never attempted by more than one instance at a time. A task of the
// kind model.OAuth2SessionLogoutTaskKindRevoke revokes the online refresh tokens of its session, and is kept until it
// expires so its session remains ended until every authorization code and device code issued to it has expired. A task
// of the kind model.OAuth2SessionLogoutTaskKindBackChannel delivers a Logout Token to its client until it's delivered or
// it expires. It's claimed only once a delivery may be performed, so the time spent waiting for one doesn't count
// towards the time allowed for the delivery, and the deliveries are performed concurrently.
type LogoutTaskCollector struct {
	storage   LogoutTaskStorage
	clock     clock.Provider
	deliverer LogoutTokenDeliverer
}

// Collect processes up to limit logout tasks whose next attempt is due, returning the number processed successfully
// along with any errors.
func (c *LogoutTaskCollector) Collect(ctx context.Context, limit int) (processed int, err error) {
	var tasks []model.OAuth2SessionLogoutTask

	if tasks, err = c.storage.LoadOAuth2SessionLogoutTasksDue(ctx, c.clock.Now(), limit); err != nil {
		return 0, err
	}

	return c.processAll(ctx, tasks)
}

// DeliverSessions delivers the Logout Tokens of the tasks of the kind model.OAuth2SessionLogoutTaskKindBackChannel for
// the session identifiers without waiting for them to be due, which is how they're delivered when the End-User logs
// out. A task which isn't delivered is delivered by a later Collect.
func (c *LogoutTaskCollector) DeliverSessions(ctx context.Context, sids []string) (err error) {
	var (
		all  []model.OAuth2SessionLogoutTask
		errs []error
	)

	for _, sid := range sids {
		loaded, err := c.storage.LoadOAuth2SessionLogoutTasksBySessionID(ctx, sid, model.OAuth2SessionLogoutTaskKindBackChannel)
		if err != nil {
			errs = append(errs, err)

			continue
		}

		all = append(all, loaded...)
	}

	if _, err = c.processAll(ctx, all); err != nil {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

func (c *LogoutTaskCollector) processAll(ctx context.Context, tasks []model.OAuth2SessionLogoutTask) (processed int, err error) {
	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		errs []error
	)

	done := func(task model.OAuth2SessionLogoutTask, err error) {
		mu.Lock()
		defer mu.Unlock()

		if err != nil {
			errs = append(errs, fmt.Errorf("error processing oauth2 session logout task of kind '%s' with session id '%s': %w", task.Kind, task.SessionID, err))

			return
		}

		processed++
	}

	for _, task := range tasks {
		if task.Kind == model.OAuth2SessionLogoutTaskKindBackChannel && c.deliverer != nil {
			wg.Go(func() {
				done(task, c.deliver(ctx, task))
			})

			continue
		}

		done(task, c.process(ctx, task))
	}

	wg.Wait()

	return processed, errors.Join(errs...)
}

func (c *LogoutTaskCollector) claim(ctx context.Context, task model.OAuth2SessionLogoutTask) (now time.Time, attempts int, claimed bool, err error) {
	now = c.clock.Now()

	claimed, err = c.storage.ClaimOAuth2SessionLogoutTask(ctx, task.ID, task.Attempts, now.Add(logoutTaskBackoff(task.Attempts)))

	return now, task.Attempts + 1, claimed, err
}

func (c *LogoutTaskCollector) process(ctx context.Context, task model.OAuth2SessionLogoutTask) (err error) {
	now, attempts, claimed, err := c.claim(ctx, task)
	if err != nil || !claimed {
		return err
	}

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

func (c *LogoutTaskCollector) deliver(ctx context.Context, task model.OAuth2SessionLogoutTask) (err error) {
	var release func()

	if release, err = c.deliverer.AcquireBackChannelLogoutSlot(ctx); err != nil {
		return err
	}

	defer release()

	now, attempts, claimed, err := c.claim(ctx, task)
	if err != nil || !claimed {
		return err
	}

	if !now.Before(task.ExpiresAt) {
		if _, err = c.storage.DeleteOAuth2SessionLogoutTask(ctx, task.ID, attempts); err != nil {
			return err
		}

		return fmt.Errorf("the Logout Token for the client with id '%s' was not delivered before the task expired", task.ClientID)
	}

	done, errDeliver := c.deliverer.DeliverBackChannelLogout(ctx, task)

	if done {
		cleanup, cancel := withoutDeadline(ctx)

		defer cancel()

		cleanup, cancelTimeout := context.WithTimeout(cleanup, logoutTaskCleanupTimeout)

		defer cancelTimeout()

		if _, err = c.storage.DeleteOAuth2SessionLogoutTask(cleanup, task.ID, attempts); err != nil {
			return errors.Join(errDeliver, err)
		}
	}

	if errDeliver != nil {
		return fmt.Errorf("error delivering the Logout Token to the client with id '%s': %w", task.ClientID, errDeliver)
	}

	return nil
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
