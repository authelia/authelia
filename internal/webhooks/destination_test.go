// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
)

func TestDestinationShouldDeliverWithTheExpectedRequest(t *testing.T) {
	var (
		method, path, contentType, userAgent, eventType, eventID, attempt, signature, authorization, tenant string
		body                                                                                                []byte
	)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		method, path = r.Method, r.URL.Path
		contentType = r.Header.Get("Content-Type")
		userAgent = r.Header.Get("User-Agent")
		eventType = r.Header.Get(events.HeaderCloudEventsType)
		eventID = r.Header.Get(events.HeaderCloudEventsID)
		attempt = r.Header.Get(events.HeaderDeliveryAttempt)
		signature = r.Header.Get(events.HeaderSignature)
		authorization = r.Header.Get("Authorization")
		tenant = r.Header.Get("X-Tenant")
		body, _ = io.ReadAll(r.Body)

		rw.WriteHeader(http.StatusNoContent)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL+"/hooks", func(config *schema.WebhookDestination) {
		config.Signature = schema.WebhookSignature{Secret: "secret", Algorithm: events.SignatureAlgorithmSHA256}
		config.Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "abc", Scheme: "Bearer"}
		config.Headers = map[string]string{"X-Tenant": "acme"}
	})

	event := testEvent()

	require.NoError(t, d.deliver(event))

	assert.Equal(t, http.MethodPost, method)
	assert.Equal(t, "/hooks", path)
	assert.Equal(t, events.ContentTypeHeader, contentType)
	assert.Equal(t, "Authelia/4.39.0", userAgent)
	assert.Equal(t, events.TypeUserPasswordChanged, eventType)
	assert.Equal(t, event.ID, eventID)
	assert.Equal(t, "1", attempt)
	assert.Equal(t, "Bearer abc", authorization)
	assert.Equal(t, "acme", tenant)

	signedAt := signatureTimestamp(t, signature)

	assert.InDelta(t, time.Now().Unix(), signedAt, 60)
	assert.Equal(t, Sign("secret", events.SignatureAlgorithmSHA256, signedAt, body), signature)
	assert.Contains(t, string(body), `"specversion":"1.0"`)
}

func TestDestinationShouldUseBasicAuthentication(t *testing.T) {
	var username, password string

	var ok bool

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		username, password, ok = r.BasicAuth()

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Authentication.Basic = &schema.WebhookAuthenticationBasic{Username: "john", Password: "abc123"}
	})

	require.NoError(t, d.deliver(testEvent()))

	assert.True(t, ok)
	assert.Equal(t, "john", username)
	assert.Equal(t, "abc123", password)
}

func TestDestinationShouldPresentTheTokenInTheQueryWhenConfigured(t *testing.T) {
	var authorization, cache, token, tenant string

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		authorization = r.Header.Get("Authorization")
		cache = r.Header.Get("Cache-Control")
		token = r.URL.Query().Get("access_token")
		tenant = r.URL.Query().Get("tenant")

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL+"?tenant=acme", func(config *schema.WebhookDestination) {
		config.Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "a b&c", Method: "query"}
	})

	require.NoError(t, d.deliver(testEvent()))

	assert.Empty(t, authorization)
	assert.Equal(t, "no-store", cache)
	assert.Equal(t, "a b&c", token)
	assert.Equal(t, "acme", tenant)
}

func TestDestinationShouldNotLogTheTokenWhenTheRequestFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}))

	address := server.URL

	server.Close()

	d := newTestDestination(t, address, func(config *schema.WebhookDestination) {
		config.Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "insecure_token", Method: "query"}
	})

	err := d.deliver(testEvent())

	require.Error(t, err)
	assert.NotContains(t, err.Error(), "insecure_token")
	assert.NotContains(t, err.Error(), "access_token")
}

func TestDestinationShouldRetryUntilSuccess(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) < 3 {
			rw.WriteHeader(http.StatusInternalServerError)

			return
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	d.send(single(testEvent()))

	assert.Equal(t, int32(3), attempts.Load())
}

func TestDestinationShouldNotRetryPermanentFailures(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		attempts.Add(1)

		rw.WriteHeader(http.StatusBadRequest)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	d.send(single(testEvent()))

	assert.Equal(t, int32(1), attempts.Load())
}

func TestDestinationShouldRetryTooManyRequests(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			rw.Header().Set("Retry-After", "0")
			rw.WriteHeader(http.StatusTooManyRequests)

			return
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	d.send(single(testEvent()))

	assert.Equal(t, int32(2), attempts.Load())
}

func TestDestinationShouldReportAFullBuffer(t *testing.T) {
	d := newTestDestination(t, "https://example.com", func(config *schema.WebhookDestination) {
		config.BufferSize = 1
	})

	assert.True(t, d.enqueue(testEvent()))
	assert.False(t, d.enqueue(testEvent()))
}

func TestDestinationShouldStampTheSameIdentifierOnEveryAttempt(t *testing.T) {
	identifiers := map[string]int{}

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		identifiers[r.Header.Get(events.HeaderCloudEventsID)]++

		rw.WriteHeader(http.StatusServiceUnavailable)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	d.send(single(testEvent()))

	require.Len(t, identifiers, 1)

	for _, count := range identifiers {
		assert.Equal(t, 3, count)
	}
}

func TestDestinationShouldReportAnHonestUndeliveredCount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		time.Sleep(time.Millisecond * 40)

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	recorder := &testRecorder{}

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.BufferSize = 16
	})

	d.setMetrics(recorder)

	for i := 0; i < 16; i++ {
		require.True(t, d.enqueue(testEvent()))
	}

	go d.run()

	d.beginDrain()
	d.await(time.Millisecond * 120)

	undelivered := d.stop()

	assert.Positive(t, undelivered)
	assert.Positive(t, recorder.count(outcomeDelivered))
	assert.Equal(t, 16, recorder.count(outcomeDelivered)+undelivered)
}

func TestDestinationShouldAbandonABackoffWhenStopped(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		attempts.Add(1)

		rw.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Retry.InitialInterval = time.Minute
		config.Retry.MaximumInterval = time.Minute
	})

	require.True(t, d.enqueue(testEvent()))

	go d.run()

	require.Eventually(t, func() bool { return attempts.Load() > 0 }, time.Second*5, time.Millisecond)

	start := time.Now()

	undelivered := d.stop()

	assert.Less(t, time.Since(start), time.Second)
	assert.Equal(t, 1, undelivered)
	assert.Equal(t, int32(1), attempts.Load())
}

func TestDestinationShouldNotRetryBeforeRetryAfterElapses(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			rw.Header().Set("Retry-After", "1")
			rw.WriteHeader(http.StatusTooManyRequests)

			return
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Retry.MaximumInterval = time.Millisecond * 20
	})

	start := time.Now()

	d.send(single(testEvent()))

	assert.GreaterOrEqual(t, time.Since(start), time.Second, "the maximum interval must not shorten a Retry-After")
	assert.Equal(t, int32(2), attempts.Load())
}

func TestDestinationShouldHonorRetryAfterOnServiceUnavailable(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			rw.Header().Set("Retry-After", "1")
			rw.WriteHeader(http.StatusServiceUnavailable)

			return
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Retry.MaximumInterval = time.Millisecond * 20
	})

	start := time.Now()

	d.send(single(testEvent()))

	assert.GreaterOrEqual(t, time.Since(start), time.Second)
	assert.Equal(t, int32(2), attempts.Load())
}

func TestDestinationShouldHonorRetryAfterFromAFinalAttemptBeforeTheNextEvent(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			rw.Header().Set("Retry-After", "1")
			rw.WriteHeader(http.StatusTooManyRequests)

			return
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	recorder := &testRecorder{}

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Retry.Attempts = 1
	})

	d.setMetrics(recorder)

	start := time.Now()

	d.send(single(testEvent()))

	assert.Less(t, time.Since(start), time.Second)
	assert.Equal(t, 1, recorder.count(outcomeDropped))

	d.send(single(testEvent()))

	assert.GreaterOrEqual(t, time.Since(start), time.Second, "the next event must wait for the Retry-After of the previous one")
	assert.Equal(t, int32(2), attempts.Load())
	assert.Equal(t, 1, recorder.count(outcomeDelivered))
}

func TestDestinationShouldReviveAfterGoneWhenTheHandshakeSucceeds(t *testing.T) {
	var deliveries, handshakes atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			handshakes.Add(1)

			rw.Header().Set(events.HeaderWebhookAllowedOrigin, r.Header.Get(events.HeaderWebhookRequestOrigin))
			rw.WriteHeader(http.StatusOK)

			return
		}

		if deliveries.Add(1) == 1 {
			rw.WriteHeader(http.StatusGone)

			return
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	recorder := &testRecorder{}

	d := newTestDestination(t, server.URL, nil)

	d.origin, d.recheck = "auth.example.com", time.Millisecond*20

	d.setMetrics(recorder)

	d.send(single(testEvent()))

	require.Eventually(t, func() bool { return !d.gone.Load() }, time.Second*5, time.Millisecond*5)

	assert.Equal(t, int32(1), handshakes.Load())

	d.send(single(testEvent()))

	assert.Equal(t, int32(2), deliveries.Load())
	assert.Equal(t, 1, recorder.count(outcomeDelivered))

	d.halt()
}

func TestDestinationShouldStayGoneWhileTheHandshakeFails(t *testing.T) {
	var handshakes atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			handshakes.Add(1)
		}

		rw.WriteHeader(http.StatusGone)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	d.origin, d.recheck = "auth.example.com", time.Millisecond*20

	d.send(single(testEvent()))

	require.Eventually(t, func() bool { return handshakes.Load() >= 2 }, time.Second*5, time.Millisecond*5)

	assert.True(t, d.gone.Load())

	d.halt()

	settled := handshakes.Load()

	time.Sleep(time.Millisecond * 100)

	assert.LessOrEqual(t, handshakes.Load(), settled+1, "a halted destination must stop rechecking")
}

func TestDestinationShouldReviveAfterGoneWithoutAHandshakeWhenValidationIsDisabled(t *testing.T) {
	var handshakes atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			handshakes.Add(1)
		}

		rw.WriteHeader(http.StatusGone)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Validation.Disable = true
	})

	d.recheck = time.Millisecond * 20

	d.send(single(testEvent()))

	require.Eventually(t, func() bool { return !d.gone.Load() }, time.Second*5, time.Millisecond*5)

	assert.Equal(t, int32(0), handshakes.Load())

	d.halt()
}

func TestDestinationShouldHonorRetryAfterFromTheStartupCheckBeforeTheFirstEvent(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if attempts.Add(1) == 1 {
			rw.Header().Set("Retry-After", "1")
			rw.WriteHeader(http.StatusTooManyRequests)

			return
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	start := time.Now()

	require.Error(t, d.deliver(events.NewEvent(&events.DataStartupCheck{Probe: true})))

	assert.Less(t, time.Since(start), time.Second)

	d.send(single(testEvent()))

	assert.GreaterOrEqual(t, time.Since(start), time.Second, "the first event must wait for the Retry-After of the startup check")
	assert.Equal(t, int32(2), attempts.Load())
}

func TestDestinationShouldStopSendingAfterGone(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		attempts.Add(1)

		rw.WriteHeader(http.StatusGone)
	}))

	defer server.Close()

	recorder := &testRecorder{}

	d := newTestDestination(t, server.URL, nil)

	d.setMetrics(recorder)

	d.send(single(testEvent()))

	require.True(t, d.gone.Load())

	d.send(single(testEvent()))

	assert.Equal(t, int32(1), attempts.Load(), "a destination which answered 410 must receive no further requests")
	assert.Equal(t, 2, recorder.count(outcomeDropped))
}

func TestShouldNotShortenRetryAfterWithJitter(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		interval time.Duration
		maximum  time.Duration
		min, max time.Duration
	}{
		{"ShouldHonorRetryAfterInFull", &retryableError{after: time.Second * 10}, time.Second, time.Minute, time.Second * 10, time.Second * 10},
		{"ShouldNotClampRetryAfterToTheMaximum", &retryableError{after: time.Hour}, time.Second, time.Minute, time.Hour, time.Hour},
		{"ShouldJitterTheIntervalWithoutRetryAfter", &retryableError{}, time.Second * 10, time.Minute, time.Second * 8, time.Second * 10},
		{"ShouldJitterTheIntervalForOtherErrors", errors.New("failure"), time.Second * 10, time.Minute, time.Second * 8, time.Second * 10},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for i := 0; i < 100; i++ {
				wait := retryWait(tc.interval, tc.maximum, tc.err)

				assert.GreaterOrEqual(t, wait, tc.min)
				assert.LessOrEqual(t, wait, tc.max)
			}
		})
	}
}

func TestShouldParseRetryAfter(t *testing.T) {
	testCases := []struct {
		name     string
		have     string
		expected time.Duration
		delta    time.Duration
	}{
		{"ShouldHandleEmpty", "", 0, 0},
		{"ShouldHandleDeltaSeconds", "120", time.Minute * 2, 0},
		{"ShouldHandleZeroSeconds", "0", 0, 0},
		{"ShouldRejectNegativeSeconds", "-5", 0, 0},
		{"ShouldRejectGarbage", "soon", 0, 0},
		{"ShouldHandleIMFFixdate", time.Now().Add(time.Minute).UTC().Format(http.TimeFormat), time.Minute, time.Second * 5},
		{"ShouldHandleANSICDate", time.Now().Add(time.Minute).UTC().Format(time.ANSIC), time.Minute, time.Second * 5},
		{"ShouldHandleAPastDate", time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat), 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.InDelta(t, tc.expected, parseRetryAfter(tc.have), float64(tc.delta))
		})
	}
}

func TestDestinationShouldDeliverAFullBatchInOneRequest(t *testing.T) {
	capture := &batchCapture{}

	server := httptest.NewServer(capture.handler())

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Batch = schema.WebhookBatch{Size: 3, MaxWait: time.Hour}
	})

	go d.run()

	for range 3 {
		require.True(t, d.enqueue(testEvent()))
	}

	assert.Eventually(t, func() bool { n, _, _, _ := capture.snapshot(); return n == 1 }, time.Second*5, time.Millisecond*10)

	d.beginDrain()
	d.await(time.Second * 5)
	d.stop()

	_, bodies, types, counts := capture.snapshot()

	require.Len(t, bodies, 1)
	assert.Equal(t, events.ContentTypeBatchHeader, types[0])
	assert.Equal(t, "3", counts[0])

	decoded := []map[string]any{}

	require.NoError(t, json.Unmarshal(bodies[0], &decoded))
	assert.Len(t, decoded, 3)
	assert.Equal(t, events.TypeUserPasswordChanged, decoded[0]["type"])
}

func TestDestinationShouldDeliverAPartialBatchWhenTheWindowElapses(t *testing.T) {
	capture := &batchCapture{}

	server := httptest.NewServer(capture.handler())

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Batch = schema.WebhookBatch{Size: 100, MaxWait: time.Millisecond * 50}
	})

	go d.run()

	require.True(t, d.enqueue(testEvent()))
	require.True(t, d.enqueue(testEvent()))

	assert.Eventually(t, func() bool { n, _, _, _ := capture.snapshot(); return n == 1 }, time.Second*5, time.Millisecond*10)

	d.beginDrain()
	d.await(time.Second * 5)
	d.stop()

	_, bodies, types, counts := capture.snapshot()

	require.Len(t, bodies, 1)
	assert.Equal(t, events.ContentTypeBatchHeader, types[0])
	assert.Equal(t, "2", counts[0])
}

func TestDestinationShouldDeliverAnImmediateTypeOnItsOwn(t *testing.T) {
	capture := &batchCapture{}

	server := httptest.NewServer(capture.handler())

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Batch = schema.WebhookBatch{Size: 100, MaxWait: time.Hour, Immediate: []string{"com.authelia.security.*"}}
	})

	d.immediates = immediatesOf(d.config)

	go d.run()

	require.True(t, d.enqueue(testEvent()))
	require.True(t, d.enqueue(events.NewEvent(&events.DataBan{
		Type: events.TypeSecurityBanApplied, Target: "10.0.0.1", TargetType: events.TargetTypeIP,
	})))

	assert.Eventually(t, func() bool { n, _, _, _ := capture.snapshot(); return n == 1 }, time.Second*5, time.Millisecond*10)

	_, bodies, types, counts := capture.snapshot()

	require.Len(t, bodies, 1)
	assert.Equal(t, events.ContentTypeHeader, types[0], "an immediate event must use single event mode")
	assert.Equal(t, "", counts[0], "an immediate event must not carry a batch count")

	decoded := map[string]any{}

	require.NoError(t, json.Unmarshal(bodies[0], &decoded))
	assert.Equal(t, events.TypeSecurityBanApplied, decoded["type"])

	d.beginDrain()
	d.await(time.Second * 5)
	d.stop()
}

func TestDestinationShouldFlushAPartialBatchOnShutdown(t *testing.T) {
	capture := &batchCapture{}

	server := httptest.NewServer(capture.handler())

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Batch = schema.WebhookBatch{Size: 100, MaxWait: time.Hour}
	})

	go d.run()

	require.True(t, d.enqueue(testEvent()))

	assert.Eventually(t, func() bool { return len(d.ch) == 0 }, time.Second*5, time.Millisecond*10)

	d.beginDrain()
	d.await(time.Second * 5)
	d.stop()

	requests, bodies, types, _ := capture.snapshot()

	require.Equal(t, 1, requests, "the partial batch must be delivered within the grace period rather than dropped")
	assert.Equal(t, events.ContentTypeBatchHeader, types[0])

	decoded := []map[string]any{}

	require.NoError(t, json.Unmarshal(bodies[0], &decoded))
	assert.Len(t, decoded, 1)
}

func TestDestinationShouldCountAPartialBatchAbandonedOnStop(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		attempts.Add(1)

		rw.WriteHeader(http.StatusInternalServerError)
	}))

	defer server.Close()

	recorder := &testRecorder{}

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Retry.InitialInterval = time.Minute
		config.Retry.MaximumInterval = time.Minute
		config.Batch = schema.WebhookBatch{Size: 100, MaxWait: time.Hour, Immediate: []string{"com.authelia.security.*"}}
	})

	d.immediates = immediatesOf(d.config)

	d.setMetrics(recorder)

	require.True(t, d.enqueue(testEvent()))
	require.True(t, d.enqueue(testEvent()))
	require.True(t, d.enqueue(events.NewEvent(&events.DataBan{
		Type: events.TypeSecurityBanApplied, Target: "10.0.0.1", TargetType: events.TargetTypeIP,
	})))

	go d.run()

	require.Eventually(t, func() bool { return attempts.Load() > 0 }, time.Second*5, time.Millisecond)

	undelivered := d.stop()

	assert.Equal(t, 3, undelivered, "the batch held while the immediate delivery was retrying must be counted")
	assert.Equal(t, 3, recorder.count(outcomeDropped))
	assert.Equal(t, int32(1), attempts.Load())
}

func TestDestinationShouldSignTheBatchBody(t *testing.T) {
	var signature string

	capture := &batchCapture{}

	inner := capture.handler()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		signature = r.Header.Get(events.HeaderSignature)

		inner(rw, r)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Batch = schema.WebhookBatch{Size: 2, MaxWait: time.Hour}
		config.Signature = schema.WebhookSignature{Secret: "secret", Algorithm: events.SignatureAlgorithmSHA256}
	})

	go d.run()

	require.True(t, d.enqueue(testEvent()))
	require.True(t, d.enqueue(testEvent()))

	assert.Eventually(t, func() bool { n, _, _, _ := capture.snapshot(); return n == 1 }, time.Second*5, time.Millisecond*10)

	d.beginDrain()
	d.await(time.Second * 5)
	d.stop()

	_, bodies, _, _ := capture.snapshot()

	require.Len(t, bodies, 1)
	require.NotEmpty(t, signature)

	assert.Equal(t, Sign("secret", events.SignatureAlgorithmSHA256, signatureTimestamp(t, signature), bodies[0]), signature,
		"the signature must cover the batch body exactly as transmitted")
}

func TestDestinationShouldDeliverCloudEventsHeadersMatchingTheBody(t *testing.T) {
	var (
		headers http.Header
		body    []byte
	)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()
		body, _ = io.ReadAll(r.Body)

		rw.WriteHeader(http.StatusNoContent)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	d.version = "4.39.0 dev"

	event := testEvent()

	require.NoError(t, d.deliver(event))

	decoded := map[string]any{}

	require.NoError(t, json.Unmarshal(body, &decoded))

	testCases := []struct {
		name      string
		header    string
		attribute string
	}{
		{"ShouldMatchSpecVersion", events.HeaderCloudEventsSpecVersion, "specversion"},
		{"ShouldMatchID", events.HeaderCloudEventsID, "id"},
		{"ShouldMatchType", events.HeaderCloudEventsType, "type"},
		{"ShouldMatchSource", events.HeaderCloudEventsSource, "source"},
		{"ShouldMatchTime", events.HeaderCloudEventsTime, "time"},
		{"ShouldMatchDataSchema", events.HeaderCloudEventsDataSchema, "dataschema"},
		{"ShouldMatchAutheliaVersion", events.HeaderCloudEventsAutheliaVersion, "autheliaversion"},
		{"ShouldMatchSubject", events.HeaderCloudEventsSubject, "subject"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Len(t, headers.Values(tc.header), 1)

			have, err := url.PathUnescape(headers.Get(tc.header))

			require.NoError(t, err)
			assert.NotEmpty(t, have)
			assert.Equal(t, decoded[tc.attribute], have)
		})
	}

	assert.Equal(t, event.ID, headers.Get(events.HeaderCloudEventsID))
	assert.Equal(t, event.Type, headers.Get(events.HeaderCloudEventsType))
	assert.Equal(t, "4.39.0%20dev", headers.Get(events.HeaderCloudEventsAutheliaVersion))
	assert.Equal(t, events.ContentTypeHeader, headers.Get("Content-Type"))
	assert.Empty(t, headers.Values(events.HeaderPrefixCloudEvents+"datacontenttype"))
	assert.Len(t, cloudEventsHeaders(headers), len(testCases))
}

func TestDestinationShouldNotDeliverCloudEventsHeadersWithABatch(t *testing.T) {
	var headers http.Header

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()

		rw.WriteHeader(http.StatusNoContent)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	require.NoError(t, d.attempt(delivery{events: []*events.Event{testEvent(), testEvent()}, batched: true}, 1))

	require.NotNil(t, headers)
	assert.Equal(t, events.ContentTypeBatchHeader, headers.Get("Content-Type"))
	assert.Equal(t, "2", headers.Get(events.HeaderEventCount))
	assert.Empty(t, cloudEventsHeaders(headers))
}

func newTestDestination(t *testing.T, address string, modify func(config *schema.WebhookDestination)) *destination {
	t.Helper()

	target, err := url.Parse(address)
	require.NoError(t, err)

	config := schema.WebhookDestination{
		Name:       "test",
		Address:    target,
		Events:     []string{"*"},
		Timeout:    time.Second * 5,
		BufferSize: 8,
		Retry: schema.WebhookRetry{
			Attempts:        3,
			InitialInterval: time.Millisecond,
			MaximumInterval: time.Millisecond * 5,
		},
		TLS: &schema.TLS{},
	}

	if modify != nil {
		modify(&config)
	}

	return newDestination(config, "https://auth.example.com", "4.39.0", nil, logrus.NewEntry(logrus.New()))
}

func cloudEventsHeaders(headers http.Header) (names []string) {
	for name := range headers {
		if strings.HasPrefix(strings.ToLower(name), events.HeaderPrefixCloudEvents) {
			names = append(names, name)
		}
	}

	return names
}

func single(event *events.Event) delivery {
	return delivery{events: []*events.Event{event}}
}

func testEvent() *events.Event {
	return events.NewEvent(&events.DataUserPassword{
		Type:    events.TypeUserPasswordChanged,
		Subject: events.Subject{Username: "john"},
	})
}

func signatureTimestamp(t *testing.T, signature string) (timestamp int64) {
	t.Helper()

	element, _, ok := strings.Cut(signature, ",")
	require.True(t, ok)

	value, ok := strings.CutPrefix(element, "t=")
	require.True(t, ok)

	timestamp, err := strconv.ParseInt(value, 10, 64)
	require.NoError(t, err)

	return timestamp
}

type testRecorder struct {
	mutex    sync.Mutex
	outcomes map[string]int
}

func (r *testRecorder) RecordRequest(_, _ string, _ time.Duration) {}

func (r *testRecorder) RecordRequestOpenIDConnect(_, _ string, _ time.Duration) {}

func (r *testRecorder) RecordAuthz(_ string) {}

func (r *testRecorder) RecordAuthenticationDuration(_ bool, _ time.Duration) {}

func (r *testRecorder) RecordWebhookDelivery(_, _, outcome string) {
	r.mutex.Lock()

	defer r.mutex.Unlock()

	if r.outcomes == nil {
		r.outcomes = map[string]int{}
	}

	r.outcomes[outcome]++
}

func (r *testRecorder) count(outcome string) (count int) {
	r.mutex.Lock()

	defer r.mutex.Unlock()

	return r.outcomes[outcome]
}

type batchCapture struct {
	mu       sync.Mutex
	bodies   [][]byte
	types    []string
	counts   []string
	requests int
}

func (c *batchCapture) handler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)

		c.mu.Lock()

		c.bodies = append(c.bodies, body)
		c.types = append(c.types, r.Header.Get("Content-Type"))
		c.counts = append(c.counts, r.Header.Get(events.HeaderEventCount))
		c.requests++

		c.mu.Unlock()

		rw.WriteHeader(http.StatusOK)
	}
}

func (c *batchCapture) snapshot() (requests int, bodies [][]byte, types, counts []string) {
	c.mu.Lock()

	defer c.mu.Unlock()

	return c.requests, append([][]byte{}, c.bodies...), append([]string{}, c.types...), append([]string{}, c.counts...)
}
