// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package dynamic

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// DefaultThrottleInterval is the default minimum interval between reloads triggered by file events.
const DefaultThrottleInterval = 2 * time.Second

// NewThrottle creates a Throttle around a reload function.
func NewThrottle(interval time.Duration, reload func() (bool, error), log *logrus.Entry) *Throttle {
	return &Throttle{interval: interval, reload: reload, log: log}
}

// Throttle collapses bursts of reload triggers, which are common when editors write a file several times or replace it
// via a rename. The first trigger runs immediately. Triggers within the interval of the previous run are not lost, they
// schedule a single trailing run at the end of the interval, which reads the files as they are at that point and so
// always sees the latest content. It implements service.ReloadableProvider.
type Throttle struct {
	interval time.Duration
	reload   func() (bool, error)
	log      *logrus.Entry

	mu      sync.Mutex
	last    time.Time
	timer   *time.Timer
	stopped bool
}

// Reload runs the reload now if outside the throttle interval, otherwise it ensures a trailing run is scheduled. A
// scheduled run reports its own result via the logger so the return for a deferred run is always (false, nil).
func (t *Throttle) Reload() (reloaded bool, err error) {
	t.mu.Lock()

	if t.stopped {
		t.mu.Unlock()

		return false, nil
	}

	now := time.Now()

	if t.timer == nil && (t.last.IsZero() || now.Sub(t.last) >= t.interval) {
		t.last = now

		t.mu.Unlock()

		return t.reload()
	}

	if t.timer == nil {
		t.timer = time.AfterFunc(t.interval-now.Sub(t.last), t.trailing)
	}

	t.mu.Unlock()

	return false, nil
}

func (t *Throttle) trailing() {
	t.mu.Lock()

	t.timer = nil

	if t.stopped {
		t.mu.Unlock()

		return
	}

	t.last = time.Now()

	t.mu.Unlock()

	reloaded, err := t.reload()

	switch {
	case err != nil:
		t.log.WithError(err).Error("Error occurred during reload")
	case reloaded:
		t.log.Info("Reloaded successfully")
	default:
		t.log.Debug("Reload was triggered but it was skipped")
	}
}

// Stop prevents any further reloads, including a scheduled trailing run.
func (t *Throttle) Stop() {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.stopped = true

	if t.timer != nil {
		t.timer.Stop()
		t.timer = nil
	}
}
