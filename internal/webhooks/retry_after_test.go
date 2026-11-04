// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	logtest "github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
)

func TestShouldReportWhetherRetryAfterWasLimited(t *testing.T) {
	testCases := []struct {
		name    string
		have    string
		limited bool
	}{
		{"ShouldNotLimitEmpty", "", false},
		{"ShouldNotLimitTheLimit", "3600", false},
		{"ShouldLimitLongDeltaSeconds", "3601", true},
		{"ShouldLimitDeltaSecondsWhichOverflowADuration", "99999999999", true},
		{"ShouldNotLimitANearDate", time.Now().Add(time.Minute).UTC().Format(http.TimeFormat), false},
		{"ShouldLimitADistantDate", time.Now().Add(time.Hour * 48).UTC().Format(http.TimeFormat), true},
		{"ShouldNotLimitGarbage", "soon", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, limited := parseRetryAfter(tc.have)

			assert.Equal(t, tc.limited, limited)
		})
	}
}

func TestDestinationPostponeShouldKeepTheLaterDeadline(t *testing.T) {
	d := newTestDestination(t, "https://example.com", nil)

	d.postpone(time.Hour)
	d.postpone(time.Second)

	assert.Greater(t, d.paused(), time.Minute*59, "a shorter Retry-After must not cut a longer one short")

	d.postpone(time.Hour * 2)

	assert.Greater(t, d.paused(), time.Minute*119, "a longer Retry-After must extend the pause")
}

func TestDestinationShouldWarnWhenRetryAfterIsLimited(t *testing.T) {
	d := newTestDestination(t, "https://example.com", nil)

	logger, hook := logtest.NewNullLogger()

	d.log = logrus.NewEntry(logger)

	d.observe(&retryableError{status: http.StatusTooManyRequests, after: time.Second})

	assert.Empty(t, hook.AllEntries())

	d.observe(&retryableError{status: http.StatusTooManyRequests, after: retryAfterLimit, limited: true, header: "86400"})

	require.Len(t, hook.AllEntries(), 1)
	assert.Equal(t, logrus.WarnLevel, hook.LastEntry().Level)
	assert.Equal(t, "86400", hook.LastEntry().Data["retry_after"])
}

func TestDestinationHandshakeShouldRecordRetryAfter(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Retry-After", "30")
		rw.WriteHeader(http.StatusTooManyRequests)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	_, err := d.validateAttempt(context.Background(), "auth.example.com", 1)

	require.Error(t, err)
	assert.InDelta(t, time.Second*30, d.paused(), float64(time.Second*2))
}

func TestDestinationReviveShouldNotHandshakeWhilePaused(t *testing.T) {
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		rw.Header().Set(events.HeaderWebhookAllowedOrigin, r.Header.Get(events.HeaderWebhookRequestOrigin))
		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	defer d.halt()

	d.recheck = time.Hour
	d.origin = "auth.example.com"
	d.gone.Store(true)
	d.postpone(time.Hour)

	d.revive()

	assert.Equal(t, int32(0), requests.Load())
	assert.True(t, d.gone.Load())

	d.mutex.Lock()
	scheduled := d.revival != nil
	d.mutex.Unlock()

	assert.True(t, scheduled, "the revival must be rescheduled rather than abandoned")
}

func TestDispatcherStartupCheckShouldSkipAPausedDestination(t *testing.T) {
	var requests atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	dispatcher := newValidationDispatcher(t, server.URL, func(config *schema.WebhookDestination) {
		config.Validation.Disable = true
	})

	defer dispatcher.Shutdown()

	require.NoError(t, dispatcher.Validate())

	dispatcher.Start()

	dispatcher.destinations[0].postpone(time.Hour)

	assert.NoError(t, dispatcher.StartupCheck())
	assert.Equal(t, int32(0), requests.Load())
}

// A receiver which confirmed by callback on an earlier start, and which answers this start's handshake with a
// Retry-After, is accepted from its stored grant. Neither the startup check probe nor an event may reach it before the
// Retry-After elapses.
func TestDispatcherShouldHonorAHandshakeRetryAfterForADestinationAcceptedFromItsGrant(t *testing.T) {
	var (
		limited    atomic.Bool
		posts      atomic.Int32
		limitedAt  atomic.Int64
		deliveryAt atomic.Int64
	)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			posts.Add(1)
			deliveryAt.CompareAndSwap(0, time.Now().UnixNano())

			rw.WriteHeader(http.StatusOK)

			return
		}

		if limited.Load() {
			limitedAt.CompareAndSwap(0, time.Now().UnixNano())

			rw.Header().Set("Retry-After", "1")
			rw.WriteHeader(http.StatusTooManyRequests)

			return
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	store := newTestStore()

	first := newValidationDispatcher(t, server.URL, nil)
	first.SetStore(store)

	require.NoError(t, first.Validate())
	require.True(t, first.Confirm("admin-api", first.callbackKey(first.destinations[0]), ""))

	first.Shutdown()

	limited.Store(true)

	second := newValidationDispatcher(t, server.URL, nil)
	second.SetStore(store)

	defer second.Shutdown()

	require.NoError(t, second.Validate())
	require.False(t, second.destinations[0].pending, "the stored grant must accept the destination")

	second.Start()

	require.NoError(t, second.StartupCheck())

	second.Emit(context.Background(), testEvent())

	require.Eventually(t, func() bool { return deliveryAt.Load() != 0 }, time.Second*5, time.Millisecond*10)

	assert.Equal(t, int32(1), posts.Load(), "the startup check probe must have been skipped")
	assert.GreaterOrEqual(t, time.Duration(deliveryAt.Load()-limitedAt.Load()), time.Second, "the event must wait for the Retry-After of the handshake")
}
