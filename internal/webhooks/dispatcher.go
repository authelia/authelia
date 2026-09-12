// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/metrics"
	"github.com/authelia/authelia/v4/internal/utils"
)

const (
	drainTimeout    = time.Second * 5
	validateTimeout = time.Second * 30
)

// Dispatcher routes events to the destinations which subscribe to their type. It implements events.Emitter.
type Dispatcher struct {
	source       string
	version      string
	destinations []*destination
	routes       map[string][]*destination
	log          *logrus.Entry

	mutex   sync.RWMutex
	started bool
	stopped bool
	wg      sync.WaitGroup
}

// NewDispatcher returns a Dispatcher for the configured destinations. The routing table is computed once so that
// matching at emit time is a map lookup rather than a glob walk.
func NewDispatcher(config *schema.Configuration, caCertPool *x509.CertPool, log *logrus.Entry) (dispatcher *Dispatcher) {
	dispatcher = &Dispatcher{
		source:  sourceOf(config),
		version: utils.Version(),
		routes:  map[string][]*destination{},
		log:     log,
	}

	for _, dc := range config.Webhooks.Destinations {
		d := newDestination(dc, dispatcher.source, dispatcher.version, caCertPool, log)

		dispatcher.destinations = append(dispatcher.destinations, d)

		for _, selector := range dc.Events {
			for _, name := range events.Match(selector) {
				if contains(dispatcher.routes[name], d) {
					continue
				}

				dispatcher.routes[name] = append(dispatcher.routes[name], d)
			}
		}
	}

	return dispatcher
}

// Source returns the CloudEvents source attribute used for every event.
func (dispatcher *Dispatcher) Source() string {
	return dispatcher.source
}

// SetMetrics sets the metrics recorder. It is set after construction because the metrics provider is built later
// during provider initialization.
func (dispatcher *Dispatcher) SetMetrics(recorder metrics.Recorder) {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	for _, d := range dispatcher.destinations {
		d.setMetrics(recorder)
	}
}

// Validate performs the CloudEvents abuse protection handshake against every destination which has not disabled it.
//
// It must run before Start. A destination which does not confirm that it accepts deliveries from this origin is marked
// refused and receives no events, while every other destination is unaffected. Failures are returned so the caller may
// log them and never prevent startup, because webhook availability must not stop Authelia running.
func (dispatcher *Dispatcher) Validate() (err error) {
	derived := originFromSource(dispatcher.source)

	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		errs []error
	)

	// The handshakes run concurrently so that one receiver which backs off does not add its delay to every other
	// destination's. Each goroutine touches only its own destination, and this waits for all of them before the
	// workers are started, so the verdicts need no further synchronization.
	for _, d := range dispatcher.destinations {
		if d.config.Validation.Disable {
			if origin, e := originOf(d.config.Validation, derived); e == nil {
				d.origin = origin
			}

			continue
		}

		wg.Add(1)

		go func(d *destination) {
			defer wg.Done()

			origin, e := originOf(d.config.Validation, derived)
			if e == nil {
				var rate int

				if rate, e = d.validate(origin); e == nil {
					d.origin, d.pace = origin, interval(rate)

					dispatcher.log.WithFields(map[string]any{"destination": d.config.Name, "origin": origin, "rate": rate}).
						Debug("Webhook destination confirmed it accepts deliveries from this origin")

					return
				}
			}

			d.refused = true

			e = fmt.Errorf("destination '%s': %w", d.config.Name, e)

			dispatcher.log.WithError(e).WithField("destination", d.config.Name).
				Error("Webhook destination did not confirm it accepts deliveries from this origin and will receive no events")

			mu.Lock()

			errs = append(errs, e)

			mu.Unlock()
		}(d)
	}

	wg.Wait()

	return errors.Join(errs...)
}

// Start launches one worker per destination. It is a no-op once the dispatcher has been shut down: the workers of a
// stopped dispatcher have returned for good, and relaunching them over the signaling channels shutdown closed would
// panic a goroutine which has nothing to recover it.
func (dispatcher *Dispatcher) Start() {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	if dispatcher.started || dispatcher.stopped {
		return
	}

	for _, d := range dispatcher.destinations {
		dispatcher.wg.Add(1)

		go func(d *destination) {
			defer dispatcher.wg.Done()

			d.run()
		}(d)
	}

	dispatcher.started = true
}

// Emit queues an event for every destination which subscribes to its type. It never blocks.
func (dispatcher *Dispatcher) Emit(_ context.Context, event *events.Event) {
	if event == nil {
		return
	}

	dispatcher.mutex.RLock()

	started := dispatcher.started

	dispatcher.mutex.RUnlock()

	if !started {
		return
	}

	for _, d := range dispatcher.routes[event.Type] {
		if d.enqueue(event) {
			continue
		}

		d.record(event, outcomeDropped)

		dispatcher.log.WithFields(map[string]any{"destination": d.config.Name, "event": event.Type}).
			Error("Webhook event was dropped because the destination buffer is full")
	}
}

// Shutdown stops accepting events and drains every destination for a bounded grace period, logging the number of
// events which were not delivered.
//
// The grace period is a single deadline shared by every destination rather than a fresh one for each, so the total
// time Shutdown takes is bounded by drainTimeout however many destinations are configured. Once it expires every
// worker is stopped: a pending retry backoff is abandoned and the request in flight is cancelled, so no receiver can
// hold the process open beyond it. Shutdown of the rest of Authelia, including closing the storage and user
// provider connections, waits on this.
func (dispatcher *Dispatcher) Shutdown() {
	dispatcher.mutex.Lock()

	// The stopped state is terminal and is recorded even when the dispatcher never started, so that a Start which
	// races or follows a shutdown cannot launch workers nothing will ever stop. Provisioning can fail before Run
	// reaches Start, and shutdown runs for every service regardless.
	started := dispatcher.started

	dispatcher.started = false
	dispatcher.stopped = true

	dispatcher.mutex.Unlock()

	if !started {
		return
	}

	deadline := time.Now().Add(drainTimeout)

	for _, d := range dispatcher.destinations {
		d.beginDrain()
	}

	for _, d := range dispatcher.destinations {
		d.await(time.Until(deadline))
	}

	for _, d := range dispatcher.destinations {
		if undelivered := d.stop(); undelivered > 0 {
			dispatcher.log.WithFields(map[string]any{"destination": d.config.Name, "undelivered": undelivered}).
				Error("Webhook events were dropped during shutdown")
		}
	}

	dispatcher.wg.Wait()
}

// StartupCheck performs a single delivery attempt of a synthetic event against every destination. Failures are
// returned so the caller may log them, and never prevent startup.
func (dispatcher *Dispatcher) StartupCheck() (err error) {
	for _, d := range dispatcher.destinations {
		if d.refused {
			continue
		}

		event := events.NewEvent(&events.DataStartupCheck{Probe: true})

		if e := d.deliver(event); e != nil {
			err = fmt.Errorf("destination '%s': %w", d.config.Name, e)

			dispatcher.log.WithError(err).Warn("Webhook destination startup check failed")
		}
	}

	return err
}

func contains(destinations []*destination, d *destination) bool {
	for _, existing := range destinations {
		if existing == d {
			return true
		}
	}

	return false
}

func sourceOf(config *schema.Configuration) (source string) {
	if len(config.Session.Cookies) == 0 {
		return ""
	}

	cookie := config.Session.Cookies[0]

	if cookie.AutheliaURL != nil && cookie.AutheliaURL.String() != "" {
		return cookie.AutheliaURL.String()
	}

	if cookie.Domain != "" {
		return "https://" + cookie.Domain
	}

	return ""
}
