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
	// unanswered is true while a pending destination has never answered the handshake, so it may not have seen the
	// callback and the handshake has to be repeated.
	unanswered bool
	// launched is true once the worker has been started, which happens as soon as the handshake settles.
	launched bool
	settled  bool
	granted  bool
	allowed  string
	gone     atomic.Bool
	recheck  time.Duration
	// budget limits each handshake, including its retries and any request in flight.
	budget  time.Duration
	revival *time.Timer
	revoke  func()
	regrant func() (rate int, ok bool)

	pace   atomic.Int64
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
		budget:     validateTimeout,
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

	pace := d.spacing()

	if pace <= 0 {
		return !d.stopping()
	}

	if !d.last.IsZero() {
		if wait := pace - time.Since(d.last); wait > 0 && !d.backoff(wait) {
			return false
		}
	}

	d.last = time.Now()

	return true
}

// setRate sets the interval between requests from the permitted rate in requests per minute. It is atomic because a
// confirmation or a recheck changes it while the worker is delivering.
func (d *destination) setRate(rate int) {
	d.pace.Store(int64(interval(rate)))
}

func (d *destination) spacing() time.Duration {
	return time.Duration(d.pace.Load())
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

	if d.config.Validation.Disable {
		d.log.Error("Webhook destination responded with the status code 410 which indicates it has been retired and it will receive no events until Authelia is restarted")

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
	if d.stopping() || !d.gone.Load() {
		return
	}

	if d.paused() > 0 {
		d.scheduleRevival()

		return
	}

	if !d.config.Validation.Disable {
		rate, err := d.validate(d.origin)
		if err != nil {
			var ok bool

			if rate, ok = d.reconfirmed(err); !ok {
				if !d.stopping() {
					d.log.WithError(err).WithField("recheck_in", d.recheck.String()).
						Warn("Webhook destination which was retired did not confirm it accepts deliveries and will still receive no events")
				}

				d.scheduleRevival()

				return
			}
		}

		// The receiver may permit a different rate than it did before it was retired.
		d.setRate(rate)
	}

	if d.gone.CompareAndSwap(true, false) {
		d.log.Info("Webhook destination which was retired will receive events again")
	}
}

// reconfirmed reports whether a retired destination whose handshake did not confirm it has since confirmed using the
// callback, and the rate that confirmation permits. The callback may have been received by another instance sharing
// the storage backend. The 410 which retired the destination removed its earlier confirmation, so any confirmation found
// was given afterwards. A handshake which is itself answered with 410 withdraws that confirmation again.
func (d *destination) reconfirmed(err error) (rate int, ok bool) {
	if errors.Is(err, errHandshakeRetired) {
		if d.revoke != nil {
			d.revoke()
		}

		return 0, false
	}

	if d.regrant == nil {
		return 0, false
	}

	return d.regrant()
}

// reinstate resumes deliveries to a retired destination which confirmed using the callback, and cancels the handshake
// which was scheduled to recheck it. It reports whether the destination was retired.
func (d *destination) reinstate() (reinstated bool) {
	d.mutex.Lock()

	if d.revival != nil {
		d.revival.Stop()
	}

	d.mutex.Unlock()

	return d.gone.CompareAndSwap(true, false)
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

	if failure := retryableStatus(resp); failure != nil {
		d.observe(failure)

		return failure
	}

	switch {
	case resp.StatusCode >= 200 && resp.StatusCode < 300:
		return nil
	case resp.StatusCode == http.StatusGone:
		d.retire()

		return &permanentError{err: fmt.Errorf("received status code %d", resp.StatusCode)}
	default:
		return &permanentError{err: fmt.Errorf("received status code %d", resp.StatusCode)}
	}
}

func (d *destination) deliver(event *events.Event) (err error) {
	return d.attempt(delivery{events: []*events.Event{event}}, 1)
}

// observe records the Retry-After of a retryable response, so every later request to the destination waits for it
// whichever path sends that request. It is called wherever a response is read, rather than by the callers, so no path
// can accept or contact a destination while forgetting a wait it was asked to observe.
func (d *destination) observe(failure *retryableError) {
	if failure.after <= 0 {
		return
	}

	d.postpone(failure.after)

	if failure.limited {
		d.log.WithFields(map[string]any{"retry_after": failure.header, "limit": retryAfterLimit.String()}).
			Warn("Webhook destination asked for a Retry-After longer than the limit and will be contacted again once the limit elapses")
	}
}

// postpone pauses every request to the destination for the duration. A pause already in force which ends later is
// kept, because concurrent requests may return their Retry-After headers in any order.
func (d *destination) postpone(after time.Duration) {
	until := time.Now().Add(after).UnixNano()

	for {
		current := d.resume.Load()

		if current >= until || d.resume.CompareAndSwap(current, until) {
			return
		}
	}
}

// paused returns how much longer the destination has asked not to be contacted.
func (d *destination) paused() (remaining time.Duration) {
	return max(time.Until(time.Unix(0, d.resume.Load())), 0)
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
	status  int
	after   time.Duration
	limited bool
	header  string
}

func (e *retryableError) Error() string {
	return fmt.Sprintf("received status code %d", e.status)
}

// retryableStatus returns the error for a response which is retried, or nil for any other response.
func retryableStatus(resp *http.Response) (failure *retryableError) {
	switch {
	case resp.StatusCode == http.StatusTooManyRequests, resp.StatusCode == http.StatusServiceUnavailable:
		header := resp.Header.Get("Retry-After")
		after, limited := parseRetryAfter(header)

		return &retryableError{status: resp.StatusCode, after: after, limited: limited, header: header}
	case resp.StatusCode >= 500:
		return &retryableError{status: resp.StatusCode}
	default:
		return nil
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

// parseRetryAfter returns the wait a Retry-After header asks for, and whether it was reduced to retryAfterLimit.
func parseRetryAfter(value string) (after time.Duration, limited bool) {
	if value == "" {
		return 0, false
	}

	if seconds, err := strconv.Atoi(value); err == nil {
		if seconds < 0 {
			return 0, false
		}

		if seconds > int(retryAfterLimit/time.Second) {
			return retryAfterLimit, true
		}

		return time.Duration(seconds) * time.Second, false
	}

	date, err := http.ParseTime(value)
	if err != nil {
		return 0, false
	}

	if after = time.Until(date); after < 0 {
		return 0, false
	}

	if after > retryAfterLimit {
		return retryAfterLimit, true
	}

	return after, false
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
