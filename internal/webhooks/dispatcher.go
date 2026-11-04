// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/metrics"
	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/utils"
)

// Dispatcher routes events to the destinations which subscribe to their type. It implements events.Emitter.
type Dispatcher struct {
	source       string
	version      string
	destinations []*destination
	routes       map[string][]*destination
	log          *logrus.Entry
	store        Store
	fallback     []byte
	recheck      time.Duration
	unanswered   time.Duration
	done         chan struct{}

	mutex     sync.RWMutex
	accepting bool
	started   bool
	stopped   bool
	wg        sync.WaitGroup
}

// NewDispatcher returns a Dispatcher for the configured destinations. The routing table is computed once so that
// matching at emit time is a map lookup rather than a glob walk.
func NewDispatcher(config *schema.Configuration, caCertPool *x509.CertPool, log *logrus.Entry) (dispatcher *Dispatcher) {
	dispatcher = &Dispatcher{
		source:  sourceOf(config),
		version: utils.Version(),
		routes:  map[string][]*destination{},
		log:     log,
		recheck: pendingRecheckInterval,
		done:    make(chan struct{}),

		unanswered: unansweredRecheckInterval,
	}

	dispatcher.fallback = make([]byte, callbackSecretLength)

	if _, err := rand.Read(dispatcher.fallback); err != nil {
		dispatcher.fallback = nil
	}

	for _, dc := range config.Webhooks.Destinations {
		d := newDestination(dc, dispatcher.source, dispatcher.version, caCertPool, log)

		dispatcher.destinations = append(dispatcher.destinations, d)

		for _, selector := range dc.Events {
			for _, name := range events.Match(selector) {
				if slices.Contains(dispatcher.routes[name], d) {
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

// SetStore sets the storage used to record callback confirmations. It is set after construction because it is
// optional, and without it a destination which confirms using the callback has to confirm again after a restart.
func (dispatcher *Dispatcher) SetStore(store Store) {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	dispatcher.store = store

	for _, d := range dispatcher.destinations {
		d.revoke = func() {
			dispatcher.forget(d)
		}

		d.regrant = func() (rate int, ok bool) {
			dispatcher.mutex.RLock()

			origin := d.origin

			dispatcher.mutex.RUnlock()

			saved, ok := dispatcher.recall(d, origin)

			return saved.Rate, ok
		}
	}
}

// Validate performs the CloudEvents abuse protection handshake against every destination which has not disabled it.
//
// It must run before Start. A destination which does not confirm that it accepts deliveries from this origin is marked
// refused and receives no events, while every other destination is unaffected. Failures are returned so the caller may
// log them and never prevent startup, because webhook availability must not stop Authelia running.
//
// Events are accepted from the moment the handshakes begin and are held in the destination buffers until the worker of
// the destination starts, so the events which occur while a slow receiver is still answering are delivered rather than
// lost. Each worker starts as soon as the handshake of its own destination settles, so one slow receiver does not delay
// deliveries to the others. It is a no-op once the dispatcher has been shut down.
func (dispatcher *Dispatcher) Validate() (err error) {
	dispatcher.mutex.Lock()

	if dispatcher.stopped {
		dispatcher.mutex.Unlock()

		return nil
	}

	dispatcher.accepting = true

	dispatcher.mutex.Unlock()

	dispatcher.prepare()

	derived := originFromSource(dispatcher.source)

	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		errs []error
	)

	for _, d := range dispatcher.destinations {
		if d.config.Validation.Disable {
			dispatcher.launch(d)

			continue
		}

		wg.Add(1)

		go func(d *destination) {
			defer wg.Done()

			defer dispatcher.launch(d)

			origin, e := originOf(d.config.Validation, derived)
			if e == nil {
				var rate int

				if rate, e = d.validate(origin); e == nil {
					dispatcher.accept(d, origin, rate)

					dispatcher.log.WithFields(map[string]any{"destination": d.config.Name, "origin": origin, "rate": rate}).
						Debug("Webhook destination confirmed it accepts deliveries from this origin")

					return
				}

				if held, report := dispatcher.hold(d, origin, e); held {
					if report != nil {
						mu.Lock()

						errs = append(errs, report)

						mu.Unlock()
					}

					return
				}
			}

			dispatcher.refuse(d)

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

func (dispatcher *Dispatcher) accept(d *destination, origin string, rate int) {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	d.origin, d.settled = origin, true

	d.setRate(rate)
}

func (dispatcher *Dispatcher) hold(d *destination, origin string, err error) (held bool, report error) {
	var refusal *refusalError

	if d.callback == "" || errors.As(err, &refusal) {
		return false, nil
	}

	fields := map[string]any{"destination": d.config.Name, "origin": origin}

	// A 410 withdraws a confirmation recorded from the callback. Every other answer without a WebHook-Allowed-Origin
	// header leaves it standing, because the receivers which confirm by callback are often those which do not implement
	// the handshake at all, and the specification says the handshake cannot rely on status codes.
	retired := errors.Is(err, errHandshakeRetired)

	if retired {
		dispatcher.withdraw(d)
	} else if saved, ok := dispatcher.recall(d, origin); ok {
		dispatcher.accept(d, origin, saved.Rate)

		dispatcher.log.WithFields(fields).WithField("confirmed", saved.Confirmed.Format(time.RFC3339)).
			Info("Webhook destination previously confirmed it accepts deliveries using the callback and will receive events")

		return true, nil
	}

	var permanent *permanentError

	switch {
	case !dispatcher.pend(d, origin):
		dispatcher.log.WithFields(fields).Debug("Webhook destination confirmed it accepts deliveries from this origin using the callback")
	case retired:
		dispatcher.log.WithFields(fields).Error("Webhook destination answered the handshake with the status code 410 which indicates it has been retired and it will receive no events until it confirms using the callback it was given")

		return true, fmt.Errorf("destination '%s': %w", d.config.Name, err)
	case errors.Is(err, errHandshakePending):
		dispatcher.log.WithFields(fields).Warn("Webhook destination withheld its confirmation and will receive no events until it confirms using the callback it was given")
	case !errors.As(err, &permanent):
		dispatcher.markUnanswered(d)

		dispatcher.log.WithError(err).WithFields(fields).WithField("recheck_in", dispatcher.unanswered.String()).Error("Webhook destination did not answer the handshake and will receive no events until it answers a repeated handshake or confirms using the callback")

		return true, fmt.Errorf("destination '%s': %w", d.config.Name, err)
	default:
		dispatcher.log.WithError(err).WithFields(fields).Error("Webhook destination did not answer the handshake and will receive no events until it confirms using the callback it was given")

		return true, fmt.Errorf("destination '%s': %w", d.config.Name, err)
	}

	return true, nil
}

// withdraw forgets the recorded confirmation of a destination which answered the handshake with 410, unless the
// callback confirmed it while the handshake was in flight, which is the more recent statement of consent.
func (dispatcher *Dispatcher) withdraw(d *destination) {
	dispatcher.mutex.RLock()

	granted := d.granted

	dispatcher.mutex.RUnlock()

	if !granted {
		dispatcher.forget(d)
	}
}

func (dispatcher *Dispatcher) remember(d *destination, origin string, rate int) {
	if dispatcher.store == nil {
		return
	}

	value, err := json.Marshal(grant{Name: d.config.Name, Address: grantAddress(d), Origin: origin, Rate: rate, Confirmed: time.Now().UTC()})
	if err != nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grantTimeout)

	defer cancel()

	if err = dispatcher.store.SaveCachedData(ctx, model.CachedData{Name: grantName(d.config.Name), Encrypted: true, Value: value}); err != nil {
		dispatcher.log.WithError(err).WithField("destination", d.config.Name).
			Warn("Webhook destination confirmation could not be saved and will have to be repeated after a restart")
	}
}

// recall loads the recorded confirmation of a destination. A confirmation only applies to the name, address, and origin
// it was given for, since consent is given to one sender for one target.
func (dispatcher *Dispatcher) recall(d *destination, origin string) (saved grant, ok bool) {
	if dispatcher.store == nil {
		return saved, false
	}

	ctx, cancel := context.WithTimeout(context.Background(), grantTimeout)

	defer cancel()

	data, err := dispatcher.store.LoadCachedData(ctx, grantName(d.config.Name))
	if err != nil {
		dispatcher.log.WithError(err).WithField("destination", d.config.Name).
			Warn("Webhook destination confirmation could not be loaded")

		return grant{}, false
	}

	if data == nil {
		return grant{}, false
	}

	if err = json.Unmarshal(data.Value, &saved); err != nil {
		return grant{}, false
	}

	if saved.Name != d.config.Name || saved.Address != grantAddress(d) || saved.Origin == "" || !strings.EqualFold(saved.Origin, origin) {
		return grant{}, false
	}

	saved.Rate = max(saved.Rate, 0)

	return saved, true
}

func (dispatcher *Dispatcher) forget(d *destination) {
	if dispatcher.store == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), grantTimeout)

	defer cancel()

	if err := dispatcher.store.DeleteCachedData(ctx, grantName(d.config.Name)); err != nil {
		dispatcher.log.WithError(err).WithField("destination", d.config.Name).
			Warn("Webhook destination confirmation could not be removed")
	}
}

func grantAddress(d *destination) string {
	sum := sha256.Sum256([]byte(d.config.Address.String()))

	return hex.EncodeToString(sum[:])
}

func grantName(destination string) string {
	sum := sha256.Sum256([]byte(destination))

	return (grantNamePrefix + hex.EncodeToString(sum[:]))[:grantNameLength]
}

func (dispatcher *Dispatcher) pend(d *destination, origin string) (pending bool) {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	d.origin, d.settled = origin, true

	if d.granted {
		d.setRate(callbackRate(d.allowed, d.config.Validation))

		return false
	}

	d.pending = true

	d.discard()

	return true
}

// Confirm grants a destination which withheld its confirmation during the handshake permission to receive events. It
// implements events.Confirmer. The key is compared in constant time, and a destination which is unknown, was refused,
// or presents the wrong key is not confirmed.
func (dispatcher *Dispatcher) Confirm(name, key, rate string) (confirmed bool) {
	var d *destination

	for _, candidate := range dispatcher.destinations {
		if candidate.config.Name == name {
			d = candidate

			break
		}
	}

	if d == nil || !dispatcher.verify(d, key) {
		return false
	}

	permitted := callbackRate(rate, d.config.Validation)

	dispatcher.mutex.Lock()

	refused, origin, retired := d.refused, d.origin, false

	switch {
	case refused:
		break
	case d.pending:
		d.pending = false

		d.setRate(permitted)

		dispatcher.log.WithField("destination", d.config.Name).
			Info("Webhook destination confirmed it accepts deliveries using the callback and will now receive events")
	case !d.settled:
		d.granted, d.allowed = true, rate
	case d.gone.Load():
		retired = true

		d.setRate(permitted)
	}

	dispatcher.mutex.Unlock()

	if refused {
		return false
	}

	if retired && d.reinstate() {
		dispatcher.log.WithField("destination", d.config.Name).
			Info("Webhook destination which was retired confirmed it accepts deliveries using the callback and will receive events again")
	}

	// The confirmation is recorded even for a destination which is already receiving events, so the other instances
	// sharing the storage backend can reinstate it if they have retired it.
	dispatcher.remember(d, origin, permitted)

	return true
}

func (dispatcher *Dispatcher) refuse(d *destination) {
	dispatcher.mutex.Lock()

	d.refused, d.settled = true, true

	d.discard()

	dispatcher.mutex.Unlock()

	dispatcher.forget(d)
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

	watch := false

	for _, d := range dispatcher.destinations {
		watch = watch || d.pending

		dispatcher.launchLocked(d)
	}

	if watch && dispatcher.store != nil {
		dispatcher.wg.Add(1)

		go func() {
			defer dispatcher.wg.Done()

			dispatcher.watch()
		}()
	}

	dispatcher.accepting = true
	dispatcher.started = true
}

func (dispatcher *Dispatcher) watch() {
	ticker := time.NewTicker(dispatcher.recheck)

	defer ticker.Stop()

	for {
		select {
		case <-dispatcher.done:
			return
		case <-ticker.C:
		}

		if !dispatcher.promote() {
			return
		}
	}
}

// launch starts the worker of a destination whose handshake has settled, along with the repeated handshake of a
// destination which did not answer it. It does nothing once the dispatcher has been shut down or the worker is running.
func (dispatcher *Dispatcher) launch(d *destination) {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	dispatcher.launchLocked(d)
}

func (dispatcher *Dispatcher) launchLocked(d *destination) {
	if dispatcher.stopped || d.launched {
		return
	}

	d.launched = true

	dispatcher.wg.Add(1)

	go func() {
		defer dispatcher.wg.Done()

		d.run()
	}()

	if d.unanswered {
		dispatcher.wg.Add(1)

		go func() {
			defer dispatcher.wg.Done()

			dispatcher.watchUnanswered(d)
		}()
	}
}

func (dispatcher *Dispatcher) markUnanswered(d *destination) {
	dispatcher.mutex.Lock()

	d.unanswered = true

	dispatcher.mutex.Unlock()
}

// watchUnanswered repeats the handshake for a destination which never answered it, which is the only way out for a
// receiver that was unavailable when Authelia started because it never saw the callback.
func (dispatcher *Dispatcher) watchUnanswered(d *destination) {
	ticker := time.NewTicker(dispatcher.unanswered)

	defer ticker.Stop()

	for {
		select {
		case <-dispatcher.done:
			return
		case <-ticker.C:
		}

		if !dispatcher.repeatHandshake(d) {
			return
		}
	}
}

func (dispatcher *Dispatcher) repeatHandshake(d *destination) (again bool) {
	dispatcher.mutex.RLock()

	waiting, origin := d.pending && d.unanswered, d.origin

	dispatcher.mutex.RUnlock()

	if !waiting {
		return false
	}

	if d.paused() > 0 {
		return true
	}

	rate, err := d.validate(origin)
	if err == nil {
		dispatcher.mutex.Lock()

		d.unanswered = false

		if d.pending {
			d.pending = false

			d.setRate(rate)
		}

		dispatcher.mutex.Unlock()

		d.log.Info("Webhook destination answered a repeated handshake and will now receive events")

		return false
	}

	if d.stopping() {
		return false
	}

	var (
		refusal   *refusalError
		permanent *permanentError
	)

	switch {
	case errors.As(err, &refusal):
		dispatcher.refuse(d)

		d.log.WithError(err).Error("Webhook destination did not confirm it accepts deliveries from this origin and will receive no events")

		return false
	case errors.As(err, &permanent):
		dispatcher.mutex.Lock()

		d.unanswered = false

		dispatcher.mutex.Unlock()

		d.log.WithError(err).Warn("Webhook destination answered a repeated handshake without confirming and will receive no events until it confirms using the callback it was given")

		return false
	default:
		d.log.WithError(err).WithField("recheck_in", dispatcher.unanswered.String()).Warn("Webhook destination still did not answer the handshake")

		return true
	}
}

func (dispatcher *Dispatcher) prepare() {
	dispatcher.mutex.Lock()
	defer dispatcher.mutex.Unlock()

	derived := originFromSource(dispatcher.source)

	for _, d := range dispatcher.destinations {
		d.callback = callbackOf(dispatcher.source, d.config.Name, dispatcher.callbackKey(d))

		// The origin is known before the handshake so a callback which arrives during it records the grant against the
		// origin it was given for. A destination which disables validation has no handshake for the origin to match, so
		// it is not sent one, and its origin options are neither checked nor used.
		if d.config.Validation.Disable {
			continue
		}

		if origin, err := originOf(d.config.Validation, derived); err == nil {
			d.origin = origin
		}
	}
}

func (dispatcher *Dispatcher) callbackKey(d *destination) (key string) {
	if d.config.Validation.Disable || d.config.Address == nil {
		return ""
	}

	values := [][]byte{[]byte(callbackKeyContext), {0}, []byte(d.config.Name), {0}, []byte(d.config.Address.String())}

	if dispatcher.store != nil {
		if key = dispatcher.store.WebhookCallbackSignature(values...); key != "" {
			return key
		}
	}

	if len(dispatcher.fallback) == 0 {
		return ""
	}

	mac := hmac.New(sha256.New, dispatcher.fallback)

	for _, value := range values {
		mac.Write(value)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

func (dispatcher *Dispatcher) verify(d *destination, key string) (verified bool) {
	expected := dispatcher.callbackKey(d)

	return expected != "" && subtle.ConstantTimeCompare([]byte(expected), []byte(key)) == 1
}

func (dispatcher *Dispatcher) promote() (remaining bool) {
	dispatcher.mutex.RLock()

	var (
		pending []*destination
		origins []string
	)

	for _, d := range dispatcher.destinations {
		if d.pending {
			pending, origins = append(pending, d), append(origins, d.origin)
		}
	}

	dispatcher.mutex.RUnlock()

	for i, d := range pending {
		saved, ok := dispatcher.recall(d, origins[i])
		if !ok {
			remaining = true

			continue
		}

		dispatcher.mutex.Lock()

		promoted := d.pending

		if promoted {
			d.pending = false

			d.setRate(saved.Rate)
		}

		dispatcher.mutex.Unlock()

		if promoted {
			dispatcher.log.WithField("destination", d.config.Name).
				Info("Webhook destination confirmed it accepts deliveries using the callback on another instance and will now receive events")
		}
	}

	return remaining
}

// Emit queues an event for every destination which subscribes to its type. It never blocks.
func (dispatcher *Dispatcher) Emit(_ context.Context, event *events.Event) {
	if event == nil {
		return
	}

	dispatcher.mutex.RLock()

	defer dispatcher.mutex.RUnlock()

	if !dispatcher.accepting {
		return
	}

	for _, d := range dispatcher.routes[event.Type] {
		if d.refused || d.pending || d.gone.Load() {
			continue
		}

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

	started, stopped := dispatcher.started, dispatcher.stopped

	dispatcher.accepting = false
	dispatcher.started = false
	dispatcher.stopped = true

	dispatcher.mutex.Unlock()

	if stopped {
		return
	}

	close(dispatcher.done)

	if !started {
		for _, d := range dispatcher.destinations {
			if undelivered := d.halt(); undelivered > 0 {
				dispatcher.log.WithFields(map[string]any{"destination": d.config.Name, "undelivered": undelivered}).
					Error("Webhook events were dropped during shutdown")
			}
		}

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
// returned so the caller may log them, and never prevent startup. It is a no-op once the dispatcher has been shut down.
func (dispatcher *Dispatcher) StartupCheck() (err error) {
	dispatcher.mutex.RLock()

	stopped := dispatcher.stopped

	dispatcher.mutex.RUnlock()

	if stopped {
		return nil
	}

	for _, d := range dispatcher.destinations {
		dispatcher.mutex.RLock()

		skip := d.refused || d.pending

		dispatcher.mutex.RUnlock()

		if skip {
			continue
		}

		if remaining := d.paused(); remaining > 0 {
			dispatcher.log.WithFields(map[string]any{"destination": d.config.Name, "retry_in": remaining.String()}).
				Warn("Webhook destination startup check was skipped because the destination asked not to be contacted until its Retry-After elapses")

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
