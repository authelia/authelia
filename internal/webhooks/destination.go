// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"bytes"
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/metrics"
)

const (
	responseBodyLimit = 4096
)

// Delivery outcomes recorded against the webhook delivery metric.
const (
	outcomeDelivered = "delivered"
	outcomeRetried   = "retried"
	outcomeDropped   = "dropped"
)

type destination struct {
	config  schema.WebhookDestination
	source  string
	version string
	origin  string

	ctx    context.Context
	client *http.Client
	log    *logrus.Entry

	ch   chan *events.Event
	done chan struct{}

	cancel context.CancelFunc

	finish     chan struct{}
	finishOnce sync.Once

	quit     chan struct{}
	quitOnce sync.Once

	abandoned int

	immediates []string

	refused bool

	pace time.Duration
	last time.Time

	mutex   sync.RWMutex
	metrics metrics.Recorder
}

func newDestination(config schema.WebhookDestination, source, version string, caCertPool *x509.CertPool, log *logrus.Entry) (d *destination) {
	ctx, cancel := context.WithCancel(context.Background())

	return &destination{
		config:  config,
		source:  source,
		version: version,
		client:  NewClient(&config, caCertPool),
		ch:      make(chan *events.Event, config.BufferSize),
		done:    make(chan struct{}),
		log:     log.WithFields(map[string]any{"destination": config.Name}),
		ctx:     ctx,
		cancel:  cancel,
		finish:  make(chan struct{}),
		quit:    make(chan struct{}),

		immediates: immediatesOf(config),
	}
}

func immediatesOf(config schema.WebhookDestination) (names []string) {
	if config.Batch.Size <= 0 {
		return nil
	}

	for _, selector := range config.Batch.Immediate {
		names = append(names, events.Match(selector)...)
	}

	return names
}

func (d *destination) enqueue(event *events.Event) (queued bool) {
	if d.refused {
		return false
	}

	select {
	case d.ch <- event:
		return true
	default:
		return false
	}
}

func (d *destination) throttle() (proceed bool) {
	if d.pace <= 0 {
		return !d.stopping()
	}

	if !d.last.IsZero() {
		if wait := d.pace - time.Since(d.last); wait > 0 && !d.backoff(wait) {
			return false
		}
	}

	d.last = time.Now()

	return true
}

func (d *destination) run() {
	defer close(d.done)

	var (
		timer *time.Timer
		wait  <-chan time.Time
	)

	batch := d.newBatch()

	defer func() {
		if timer != nil {
			timer.Stop()
		}
	}()

	for {
		if d.stopping() {
			return
		}

		select {
		case <-d.quit:
			return
		case <-d.finish:
			d.sendBatch(batch)
			d.flush()

			return
		case <-wait:
			d.sendBatch(batch)

			batch, wait = d.newBatch(), nil
		case event := <-d.ch:
			if !d.batching() || d.immediate(event) {
				d.send(delivery{events: []*events.Event{event}})

				continue
			}

			batch = append(batch, event)

			if len(batch) >= d.config.Batch.Size {
				d.sendBatch(batch)

				batch, wait = d.newBatch(), nil

				continue
			}

			if wait == nil {
				if timer == nil {
					timer = time.NewTimer(d.config.Batch.MaxWait)
				} else {
					timer.Reset(d.config.Batch.MaxWait)
				}

				wait = timer.C
			}
		}
	}
}

func (d *destination) newBatch() []*events.Event {
	return make([]*events.Event, 0, max(d.config.Batch.Size, 1))
}

func (d *destination) sendBatch(batch []*events.Event) {
	if len(batch) == 0 {
		return
	}

	d.send(delivery{events: batch, batched: true})
}

func (d *destination) flush() {
	for {
		if d.stopping() {
			return
		}

		batch := d.newBatch()

		for len(batch) < cap(batch) {
			select {
			case event := <-d.ch:
				if d.batching() && !d.immediate(event) {
					batch = append(batch, event)

					continue
				}

				d.sendBatch(batch)
				d.send(delivery{events: []*events.Event{event}})

				batch = d.newBatch()
			default:
				d.sendBatch(batch)

				return
			}
		}

		d.sendBatch(batch)
	}
}

func (d *destination) stopping() (stopping bool) {
	select {
	case <-d.quit:
		return true
	default:
		return false
	}
}

func (d *destination) beginDrain() {
	d.finishOnce.Do(func() {
		close(d.finish)
	})
}

func (d *destination) await(timeout time.Duration) {
	if timeout <= 0 {
		return
	}

	timer := time.NewTimer(timeout)

	defer timer.Stop()

	select {
	case <-d.done:
	case <-timer.C:
	}
}

func (d *destination) stop() (undelivered int) {
	d.quitOnce.Do(func() {
		close(d.quit)

		d.cancel()
	})

	<-d.done

	return len(d.ch) + d.abandoned
}

type delivery struct {
	events  []*events.Event
	batched bool
}

func (d *destination) batching() bool {
	return d.config.Batch.Size > 0
}

func (d *destination) immediate(event *events.Event) bool {
	for _, name := range d.immediates {
		if name == event.Type {
			return true
		}
	}

	return false
}

func (d *destination) send(dl delivery) {
	interval := d.config.Retry.InitialInterval

	for attempt := 1; attempt <= d.config.Retry.Attempts; attempt++ {
		if !d.throttle() {
			d.abandonAll(dl)

			return
		}

		err := d.attempt(dl, attempt)
		if err == nil {
			d.recordAll(dl, outcomeDelivered)

			return
		}

		if d.stopping() {
			d.abandonAll(dl)

			return
		}

		var permanent *permanentError

		if errors.As(err, &permanent) {
			d.recordAll(dl, outcomeDropped)

			d.log.WithError(err).WithFields(d.fields(dl)).Error("Webhook delivery failed permanently and the event was dropped")

			return
		}

		if attempt == d.config.Retry.Attempts {
			d.recordAll(dl, outcomeDropped)

			d.log.WithError(err).WithFields(d.fields(dl)).Error("Webhook delivery failed after the final attempt and the event was dropped")

			return
		}

		wait := interval

		if retryAfter, ok := retryAfterOf(err); ok {
			wait = retryAfter
		}

		if wait > d.config.Retry.MaximumInterval {
			wait = d.config.Retry.MaximumInterval
		}

		d.recordAll(dl, outcomeRetried)

		d.log.WithError(err).WithFields(d.fieldsWith(dl, map[string]any{"attempt": attempt, "retry_in": wait.String()})).Warn("Webhook delivery failed and will be retried")

		if !d.backoff(jitter(wait)) {
			d.abandonAll(dl)

			return
		}

		if interval *= 2; interval > d.config.Retry.MaximumInterval {
			interval = d.config.Retry.MaximumInterval
		}
	}
}

func (d *destination) setHeaders(req *http.Request) {
	for name, value := range d.config.Headers {
		req.Header.Set(name, value)
	}

	switch {
	case d.config.Authentication.Bearer != nil:
		value := d.config.Authentication.Bearer.Token

		if d.config.Authentication.Bearer.Scheme != "" {
			value = d.config.Authentication.Bearer.Scheme + " " + value
		}

		req.Header.Set(d.config.Authentication.Bearer.Header, value)
	case d.config.Authentication.Basic != nil:
		req.SetBasicAuth(d.config.Authentication.Basic.Username, d.config.Authentication.Basic.Password)
	}
}

func (d *destination) backoff(wait time.Duration) (waited bool) {
	if wait <= 0 {
		return !d.stopping()
	}

	timer := time.NewTimer(wait)

	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-d.quit:
		return false
	}
}

func (d *destination) abandon(event *events.Event) {
	d.abandoned++

	d.record(event, outcomeDropped)

	d.log.WithField("event", event.Type).Error("Webhook delivery was abandoned because the shutdown grace period elapsed and the event was dropped")
}

func (d *destination) abandonAll(dl delivery) {
	for _, event := range dl.events {
		d.abandon(event)
	}
}

func (d *destination) recordAll(dl delivery, outcome string) {
	for _, event := range dl.events {
		d.record(event, outcome)
	}
}

func (d *destination) fields(dl delivery) map[string]any {
	if !dl.batched {
		return map[string]any{"event": dl.events[0].Type}
	}

	return map[string]any{"events": len(dl.events), "batched": true}
}

func (d *destination) fieldsWith(dl delivery, extra map[string]any) map[string]any {
	fields := d.fields(dl)

	for k, v := range extra {
		fields[k] = v
	}

	return fields
}

func (d *destination) setMetrics(recorder metrics.Recorder) {
	d.mutex.Lock()

	defer d.mutex.Unlock()

	d.metrics = recorder
}

func (d *destination) record(event *events.Event, outcome string) {
	d.mutex.RLock()

	recorder := d.metrics

	d.mutex.RUnlock()

	if recorder == nil {
		return
	}

	recorder.RecordWebhookDelivery(d.config.Name, event.Type, outcome)
}

func (d *destination) attempt(dl delivery, attempt int) (err error) {
	var body []byte

	if dl.batched {
		body, err = events.MarshalBatch(dl.events, d.source, d.version, !d.config.DisableRedaction)
	} else {
		body, err = events.Marshal(dl.events[0], d.source, d.version, !d.config.DisableRedaction)
	}

	if err != nil {
		return &permanentError{err: fmt.Errorf("error marshaling event: %w", err)}
	}

	req, err := http.NewRequestWithContext(d.ctx, http.MethodPost, d.config.Address.String(), bytes.NewReader(body))
	if err != nil {
		return &permanentError{err: fmt.Errorf("error building request: %w", err)}
	}

	signedAt := time.Now().Unix()

	req.Header.Set("User-Agent", "Authelia/"+d.version)
	req.Header.Set(events.HeaderContractVersion, events.ContractVersion)
	req.Header.Set(events.HeaderDeliveryAttempt, strconv.Itoa(attempt))

	if d.origin != "" {
		req.Header.Set(events.HeaderWebhookRequestOrigin, d.origin)
	}

	if dl.batched {
		req.Header.Set("Content-Type", events.ContentTypeBatchHeader)
		req.Header.Set(events.HeaderEventCount, strconv.Itoa(len(dl.events)))
	} else {
		req.Header.Set("Content-Type", events.ContentTypeHeader)
		req.Header.Set(events.HeaderEventType, dl.events[0].Type)
		req.Header.Set(events.HeaderEventID, dl.events[0].ID)
	}

	if signature := Sign(d.config.Signature.Secret, d.config.Signature.Algorithm, signedAt, body); signature != "" {
		req.Header.Set(events.HeaderSignature, signature)
	}

	d.setHeaders(req)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request: %w", err)
	}

	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, responseBodyLimit))

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusTooManyRequests:
		return &retryableError{status: resp.StatusCode, after: parseRetryAfter(resp.Header.Get("Retry-After"))}
	case resp.StatusCode >= 500:
		return &retryableError{status: resp.StatusCode}
	default:
		return &permanentError{err: fmt.Errorf("received status code %d", resp.StatusCode)}
	}
}

func (d *destination) deliver(event *events.Event) (err error) {
	return d.attempt(delivery{events: []*events.Event{event}}, 1)
}

type permanentError struct {
	err error
}

func (e *permanentError) Error() string {
	return e.err.Error()
}

func (e *permanentError) Unwrap() error {
	return e.err
}

type retryableError struct {
	status int
	after  time.Duration
}

func (e *retryableError) Error() string {
	return fmt.Sprintf("received status code %d", e.status)
}

func retryAfterOf(err error) (after time.Duration, ok bool) {
	var retryable *retryableError

	if errors.As(err, &retryable) && retryable.after > 0 {
		return retryable.after, true
	}

	return 0, false
}

func parseRetryAfter(value string) (after time.Duration) {
	if value == "" {
		return 0
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			return 0
		}

		return time.Duration(seconds) * time.Second
	}

	date, err := http.ParseTime(value)
	if err != nil {
		return 0
	}

	if after = time.Until(date); after < 0 {
		return 0
	}

	return after
}

func jitter(interval time.Duration) (jittered time.Duration) {
	if interval <= 0 {
		return 0
	}

	//nolint:gosec // Jitter does not require a cryptographically secure source.
	return interval - time.Duration(rand.Int63n(int64(interval)/5+1))
}
