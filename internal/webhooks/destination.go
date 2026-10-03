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
	"net/url"
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/metrics"
)

type destination struct {
	config  schema.WebhookDestination
	address string
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

	callback string
	key      string

	refused bool
	pending bool
	settled bool
	granted bool
	allowed string
	gone    atomic.Bool
	recheck time.Duration
	revival *time.Timer
	revoke  func()

	pace   time.Duration
	last   time.Time
	resume atomic.Int64

	mutex   sync.RWMutex
	metrics metrics.Recorder
}

func newDestination(config schema.WebhookDestination, source, version string, caCertPool *x509.CertPool, log *logrus.Entry) (d *destination) {
	ctx, cancel := context.WithCancel(context.Background())

	return &destination{
		config:  config,
		address: addressOf(config),
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
		recheck:    goneRecheckInterval,
	}
}

func addressOf(config schema.WebhookDestination) string {
	if config.Address == nil {
		return ""
	}

	if config.Authentication.Bearer == nil || config.Authentication.Bearer.Method != schema.WebhookAuthenticationMethodQuery {
		return config.Address.String()
	}

	address := *config.Address

	parameter := schema.WebhookQueryParameterAccessToken + "=" + url.QueryEscape(config.Authentication.Bearer.Token)

	if address.RawQuery == "" {
		address.RawQuery = parameter
	} else {
		address.RawQuery += "&" + parameter
	}

	return address.String()
}

func callbackOf(source, name, key string) string {
	if source == "" || key == "" {
		return ""
	}

	callback, err := url.Parse(source)
	if err != nil || callback.Scheme != "https" {
		return ""
	}

	callback = callback.JoinPath(PathConfirm)

	query := url.Values{}

	query.Set(QueryConfirmID, name)
	query.Set(QueryConfirmKey, key)

	callback.RawQuery = query.Encode()

	return callback.String()
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
	if wait := time.Until(time.Unix(0, d.resume.Load())); wait > 0 && !d.backoff(wait) {
		return false
	}

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
			d.abandonBatch(batch)

			return
		}

		select {
		case <-d.quit:
			d.abandonBatch(batch)

			return
		case <-d.finish:
			d.sendBatch(batch)
			d.flush()

			return
		case <-wait:
			d.sendBatch(batch)

			batch, wait = d.newBatch(), nil
		case event := <-d.ch:
			if !d.batching() || slices.Contains(d.immediates, event.Type) {
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
				if d.batching() && !slices.Contains(d.immediates, event.Type) {
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
	d.halt()

	<-d.done

	return len(d.ch) + d.abandoned
}

func (d *destination) halt() (buffered int) {
	d.quitOnce.Do(func() {
		close(d.quit)

		d.cancel()
	})

	d.mutex.Lock()

	if d.revival != nil {
		d.revival.Stop()
	}

	d.mutex.Unlock()

	return len(d.ch)
}

func (d *destination) retire() {
	if !d.gone.CompareAndSwap(false, true) {
		return
	}

	d.log.WithField("recheck_in", d.recheck.String()).
		Error("Webhook destination responded with the status code 410 which indicates it has been retired and it will receive no events until it confirms it accepts deliveries again")

	if d.revoke != nil {
		d.revoke()
	}

	d.scheduleRevival()
}

func (d *destination) scheduleRevival() {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.stopping() {
		return
	}

	d.revival = time.AfterFunc(d.recheck, d.revive)
}

func (d *destination) revive() {
	if d.stopping() {
		return
	}

	if !d.config.Validation.Disable {
		if _, err := d.validate(d.origin); err != nil {
			if !d.stopping() {
				d.log.WithError(err).WithField("recheck_in", d.recheck.String()).
					Warn("Webhook destination which was retired did not confirm it accepts deliveries and will still receive no events")
			}

			d.scheduleRevival()

			return
		}
	}

	d.gone.Store(false)

	d.log.Info("Webhook destination which was retired will receive events again")
}

func (d *destination) discard() {
	for {
		select {
		case <-d.ch:
		default:
			return
		}
	}
}

func (d *destination) batching() bool {
	return d.config.Batch.Size > 0
}

func (d *destination) send(dl delivery) {
	interval := d.config.Retry.InitialInterval

	for attempt := 1; attempt <= d.config.Retry.Attempts; attempt++ {
		if d.gone.Load() {
			d.recordAll(dl, outcomeDropped)

			return
		}

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

		if after, ok := retryAfterOf(err); ok {
			d.postpone(after)
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

		wait := retryWait(interval, d.config.Retry.MaximumInterval, err)

		d.recordAll(dl, outcomeRetried)

		d.log.WithError(err).WithFields(d.fieldsWith(dl, map[string]any{"attempt": attempt, "retry_in": wait.String()})).Warn("Webhook delivery failed and will be retried")

		if !d.backoff(wait) {
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
	case d.config.Authentication.Basic != nil:
		req.SetBasicAuth(d.config.Authentication.Basic.Username, d.config.Authentication.Basic.Password)
	case d.config.Authentication.Bearer == nil:
		break
	case d.config.Authentication.Bearer.Method == schema.WebhookAuthenticationMethodQuery:
		req.Header.Set("Cache-Control", "no-store")
	default:
		req.Header.Set("Authorization", d.config.Authentication.Bearer.Scheme+" "+d.config.Authentication.Bearer.Token)
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

func (d *destination) abandonBatch(batch []*events.Event) {
	d.abandonAll(delivery{events: batch, batched: true})
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
	var (
		body    []byte
		headers map[string]string
	)

	if dl.batched {
		body, err = events.MarshalBatch(dl.events, d.source, d.version, !d.config.DisableRedaction)
	} else {
		body, headers, err = events.MarshalWithHeaders(dl.events[0], d.source, d.version, !d.config.DisableRedaction)
	}

	if err != nil {
		return &permanentError{err: fmt.Errorf("error marshaling event: %w", err)}
	}

	req, err := http.NewRequestWithContext(d.ctx, http.MethodPost, d.address, bytes.NewReader(body))
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

		for name, value := range headers {
			req.Header.Set(name, value)
		}
	}

	if signature := Sign(d.config.Signature.Secret, d.config.Signature.Algorithm, signedAt, body); signature != "" {
		req.Header.Set(events.HeaderSignature, signature)
	}

	d.setHeaders(req)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request: %w", transportError(err))
	}

	defer resp.Body.Close()

	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, responseBodyLimit))

	after, retryable := retryableStatus(resp)

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case retryable:
		return &retryableError{status: resp.StatusCode, after: after}
	case resp.StatusCode == http.StatusGone:
		d.retire()

		return &permanentError{err: fmt.Errorf("received status code %d", resp.StatusCode)}
	default:
		return &permanentError{err: fmt.Errorf("received status code %d", resp.StatusCode)}
	}
}

func (d *destination) deliver(event *events.Event) (err error) {
	err = d.attempt(delivery{events: []*events.Event{event}}, 1)

	if after, ok := retryAfterOf(err); ok {
		d.postpone(after)
	}

	return err
}

func (d *destination) postpone(after time.Duration) {
	d.resume.Store(time.Now().Add(after).UnixNano())
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

func retryableStatus(resp *http.Response) (after time.Duration, retryable bool) {
	switch {
	case resp.StatusCode == http.StatusTooManyRequests, resp.StatusCode == http.StatusServiceUnavailable:
		return parseRetryAfter(resp.Header.Get("Retry-After")), true
	case resp.StatusCode >= 500:
		return 0, true
	default:
		return 0, false
	}
}

func transportError(err error) error {
	var e *url.Error

	if errors.As(err, &e) {
		return e.Err
	}

	return err
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

func retryWait(interval, maximum time.Duration, err error) (wait time.Duration) {
	var ok bool

	if wait, ok = retryAfterOf(err); ok {
		return wait
	}

	if wait = jitter(interval); wait > maximum {
		wait = maximum
	}

	return wait
}

func jitter(interval time.Duration) (jittered time.Duration) {
	if interval <= 0 {
		return 0
	}

	//nolint:gosec // Jitter does not require a cryptographically secure source.
	return interval - time.Duration(rand.Int63n(int64(interval)/5+1))
}
