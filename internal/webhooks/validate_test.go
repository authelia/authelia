// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/model"
)

func TestOriginOf(t *testing.T) {
	testCases := []struct {
		name     string
		config   schema.WebhookValidation
		derived  string
		expected string
		err      string
	}{
		{"ShouldPreferTheDerivedOrigin", schema.WebhookValidation{Origin: "configured.example.com"}, "auth.example.com", "auth.example.com", ""},
		{"ShouldFallBackToTheConfiguredOrigin", schema.WebhookValidation{Origin: "configured.example.com"}, "", "configured.example.com", ""},
		{"ShouldForceTheConfiguredOrigin", schema.WebhookValidation{Origin: "configured.example.com", ForceOrigin: true}, "auth.example.com", "configured.example.com", ""},
		{"ShouldErrorWhenForcedWithoutAnOrigin", schema.WebhookValidation{ForceOrigin: true}, "auth.example.com", "", "option 'force_origin' is enabled but option 'origin' is not configured"},
		{"ShouldErrorWhenNeitherIsAvailable", schema.WebhookValidation{}, "", "", "no origin could be derived from the configuration and option 'origin' is not configured"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := originOf(tc.config, tc.derived)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Equal(t, "", actual)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestOriginFromSource(t *testing.T) {
	assert.Equal(t, "auth.example.com", originFromSource("https://auth.example.com"))
	assert.Equal(t, "auth.example.com", originFromSource("https://auth.example.com:9091/base"))
	assert.Equal(t, "", originFromSource(""))
}

func TestRequestedRate(t *testing.T) {
	for _, tc := range []struct {
		name     string
		config   schema.WebhookValidation
		expected string
		send     bool
	}{
		{"ShouldOmitWhenAbsent", schema.WebhookValidation{}, "", false},
		{"ShouldOmitWhenZero", schema.WebhookValidation{Rate: 0}, "", false},
		{"ShouldOmitWhenNegative", schema.WebhookValidation{Rate: -1}, "", false},
		{"ShouldSendAPositiveRate", schema.WebhookValidation{Rate: 120}, "120", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rate, send := requestedRate(tc.config)

			assert.Equal(t, tc.expected, rate)
			assert.Equal(t, tc.send, send)
		})
	}
}

func TestAllowedRate(t *testing.T) {
	for _, tc := range []struct {
		value    string
		expected int
		err      bool
	}{
		{"", 0, false},
		{"*", 0, false},
		{"120", 120, false},
		{"0", 0, true},
		{"-1", 0, true},
		{"many", 0, true},
	} {
		t.Run(tc.value, func(t *testing.T) {
			rate, err := allowedRate(tc.value)

			if tc.err {
				assert.Error(t, err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, rate)
		})
	}
}

func TestInterval(t *testing.T) {
	assert.Equal(t, time.Duration(0), interval(0))
	assert.Equal(t, time.Duration(0), interval(-1))
	assert.Equal(t, time.Millisecond*500, interval(120))
	assert.Equal(t, time.Second, interval(60))
}

func TestDestinationValidateShouldNegotiate(t *testing.T) {
	testCases := []struct {
		name        string
		requestRate int
		allowOrigin string
		allowRate   string
		status      int
		rate        int
		err         string
	}{
		{"ShouldAcceptAnExactOrigin", 0, "auth.example.com", "", http.StatusOK, 0, ""},
		{"ShouldAcceptAWildcardOrigin", 0, "*", "", http.StatusOK, 0, ""},
		{"ShouldAcceptRegardlessOfCase", 0, "AUTH.EXAMPLE.COM", "", http.StatusOK, 0, ""},
		{"ShouldReturnTheAllowedRate", 0, "auth.example.com", "120", http.StatusOK, 120, ""},
		{"ShouldTreatAWildcardRateAsUnrestricted", 0, "auth.example.com", "*", http.StatusOK, 0, ""},
		{"ShouldTreatAnAbsentRateAsUnrestrictedWhenNoneWasRequested", 0, "auth.example.com", "", http.StatusOK, 0, ""},
		{"ShouldHonorTheAllowedRateWhenOneWasRequested", 60, "auth.example.com", "30", http.StatusOK, 30, ""},
		{"ShouldWithholdConfirmationOnAMissingAllowedOrigin", 0, "", "", http.StatusOK, 0, "the destination withheld its confirmation and may grant it using the callback"},
		{"ShouldRejectAMismatchedOrigin", 0, "other.example.com", "", http.StatusOK, 0, "the response permitted the origin 'other.example.com' rather than 'auth.example.com'"},
		{"ShouldRejectANonSuccessStatusWithoutConsent", 0, "", "", http.StatusNotFound, 0, "received status code 404"},
		{"ShouldAcceptConsentWhateverTheStatus", 0, "auth.example.com", "", http.StatusMethodNotAllowed, 0, ""},
		{"ShouldRejectAnInvalidAllowedRate", 0, "auth.example.com", "soon", http.StatusOK, 0, "the response included an invalid WebHook-Allowed-Rate header with value 'soon'"},
		{"ShouldRejectAnAbsentRateWhenOneWasRequested", 60, "auth.example.com", "", http.StatusOK, 0, "the response did not include a WebHook-Allowed-Rate header for the requested rate of 60"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			capture := &optionsCapture{allowOrigin: tc.allowOrigin, allowRate: tc.allowRate, status: tc.status}

			server := httptest.NewServer(capture.handler())

			defer server.Close()

			d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
				config.Validation.Rate = tc.requestRate
				config.Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "abc", Scheme: "Bearer"}
			})

			d.callback = "https://auth.example.com/api/webhooks/confirm?id=test&key=abc"

			rate, err := d.validate("auth.example.com")

			assert.Equal(t, int32(1), capture.requests.Load())
			assert.Equal(t, "auth.example.com", capture.origin.Load())
			assert.Equal(t, "Bearer abc", capture.auth.Load(), "the handshake must carry the destination's credentials")

			if tc.requestRate > 0 {
				assert.Equal(t, "60", capture.rate.Load(), "a configured rate must be sent as a positive integer")
			} else {
				assert.Equal(t, "", capture.rate.Load(), "the wildcard is not valid for the request rate, so the header must be omitted")
			}

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.rate, rate)
		})
	}
}

func TestDestinationShouldSendTheNegotiatedOriginOnEveryDelivery(t *testing.T) {
	var (
		origins  []string
		attempts atomic.Int32
	)

	capture := &optionsCapture{allowOrigin: "*", status: http.StatusOK}

	inner := capture.handler()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			origins = append(origins, r.Header.Get(events.HeaderWebhookRequestOrigin))

			if attempts.Add(1) < 2 {
				rw.WriteHeader(http.StatusInternalServerError)

				return
			}

			rw.WriteHeader(http.StatusOK)

			return
		}

		inner(rw, r)
	}))

	defer server.Close()

	dispatcher := newValidationDispatcher(t, server.URL, func(config *schema.WebhookDestination) {
		config.Retry = schema.WebhookRetry{Attempts: 3, InitialInterval: time.Millisecond, MaximumInterval: time.Millisecond}
	})

	require.NoError(t, dispatcher.Validate())

	dispatcher.Start()

	dispatcher.Emit(t.Context(), events.NewEvent(&events.DataUserPassword{Type: events.TypeUserPasswordChanged}))

	dispatcher.Shutdown()

	require.Len(t, origins, 2)
	assert.Equal(t, []string{"auth.example.com", "auth.example.com"}, origins)
}

func TestDispatcherStartupCheckShouldSkipARefusedDestination(t *testing.T) {
	var probes atomic.Int32

	capture := &optionsCapture{status: http.StatusOK, allowOrigin: "other.example.com"}

	inner := capture.handler()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			probes.Add(1)
		}

		inner(rw, r)
	}))

	defer server.Close()

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	require.Error(t, dispatcher.Validate(), "the receiver permits another origin, so the destination is refused")
	require.True(t, dispatcher.destinations[0].refused)

	require.NoError(t, dispatcher.StartupCheck())
	assert.Equal(t, int32(0), probes.Load(),
		"a refused destination must not receive the startup probe, which would be the unsolicited payload the handshake prevents")
}

func TestDispatcherEmitShouldSkipARefusedDestinationWithoutRecordingADrop(t *testing.T) {
	capture := &optionsCapture{status: http.StatusOK, allowOrigin: "other.example.com"}

	server := httptest.NewServer(capture.handler())

	defer server.Close()

	recorder := &testRecorder{}

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	dispatcher.SetMetrics(recorder)

	require.Error(t, dispatcher.Validate(), "the receiver permits another origin, so the destination is refused")
	require.True(t, dispatcher.destinations[0].refused)

	dispatcher.Start()

	dispatcher.Emit(context.Background(), testEvent())

	dispatcher.Shutdown()

	assert.Equal(t, 0, recorder.count(outcomeDropped), "an event skipped for a refused destination is not a full buffer drop")
	assert.Empty(t, dispatcher.destinations[0].ch)
}

func TestDispatcherShouldDeliverEventsEmittedDuringTheHandshake(t *testing.T) {
	handshake := newBlockingHandshake(true)

	server := httptest.NewServer(handshake.handler())

	defer server.Close()
	defer handshake.unblock()

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	validated := make(chan error, 1)

	go func() { validated <- dispatcher.Validate() }()

	<-handshake.entered

	dispatcher.Emit(context.Background(), testEvent())
	dispatcher.Emit(context.Background(), testEvent())

	assert.Equal(t, int32(0), handshake.deliveries.Load(), "no event may be delivered before the handshake completes")

	handshake.unblock()

	require.NoError(t, <-validated)

	dispatcher.Start()

	assert.Eventually(t, func() bool { return handshake.deliveries.Load() == 2 }, time.Second*5, time.Millisecond*10)

	dispatcher.Shutdown()
}

func TestDispatcherShouldDiscardEventsBufferedForADestinationWhichIsRefused(t *testing.T) {
	handshake := newBlockingHandshake(false)

	server := httptest.NewServer(handshake.handler())

	defer server.Close()
	defer handshake.unblock()

	recorder := &testRecorder{}

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	dispatcher.SetMetrics(recorder)

	validated := make(chan error, 1)

	go func() { validated <- dispatcher.Validate() }()

	<-handshake.entered

	dispatcher.Emit(context.Background(), testEvent())

	handshake.unblock()

	require.Error(t, <-validated)

	assert.Empty(t, dispatcher.destinations[0].ch)

	dispatcher.Start()
	dispatcher.Emit(context.Background(), testEvent())
	dispatcher.Shutdown()

	assert.Equal(t, int32(0), handshake.deliveries.Load(), "a refused destination must never receive an event")
	assert.Equal(t, 0, recorder.count(outcomeDropped))
}

func TestServiceShutdownShouldInterruptTheHandshake(t *testing.T) {
	handshake := newBlockingHandshake(true)

	server := httptest.NewServer(handshake.handler())

	defer server.Close()
	defer handshake.unblock()

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	service := NewService("main", dispatcher, true, logrus.NewEntry(logrus.New()))

	done := make(chan error, 1)

	go func() { done <- service.Run() }()

	<-handshake.entered

	dispatcher.Emit(context.Background(), testEvent())

	start := time.Now()

	service.Shutdown()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("the service did not stop while the handshake was in progress")
	}

	assert.Less(t, time.Since(start), time.Second*2)
	assert.Equal(t, int32(0), handshake.deliveries.Load(), "neither the startup check nor an event may be sent after shutdown")
}

func TestDestinationValidateShouldRetryTransientHandshakeFailures(t *testing.T) {
	testCases := []struct {
		name       string
		statuses   []int
		retryAfter string
		attempts   int32
	}{
		{"ShouldRetryTooManyRequests", []int{http.StatusTooManyRequests}, "0", 2},
		{"ShouldRetryTooManyRequestsWithoutRetryAfter", []int{http.StatusTooManyRequests}, "", 2},
		{"ShouldRetryAServerError", []int{http.StatusInternalServerError}, "", 2},
		{"ShouldRetryAGatewayError", []int{http.StatusBadGateway, http.StatusServiceUnavailable}, "", 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flaky := &flakyOptions{statuses: tc.statuses, retryAfter: tc.retryAfter}

			server := httptest.NewServer(flaky.handler())

			defer server.Close()

			d := newTestDestination(t, server.URL, nil)

			rate, err := d.validate("auth.example.com")

			require.NoError(t, err)
			assert.Equal(t, 0, rate)
			assert.Equal(t, tc.attempts, flaky.attempts.Load())
			assert.Equal(t, "1", flaky.order()[0], "the attempt number must be stamped on the handshake")
		})
	}
}

func TestDestinationValidateShouldNotRetryAPermanentHandshakeFailure(t *testing.T) {
	testCases := []struct {
		name     string
		statuses []int
	}{
		{"ShouldNotRetryNotFound", []int{http.StatusNotFound}},
		{"ShouldNotRetryForbidden", []int{http.StatusForbidden}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flaky := &flakyOptions{statuses: append(tc.statuses, tc.statuses...)}

			server := httptest.NewServer(flaky.handler())

			defer server.Close()

			d := newTestDestination(t, server.URL, nil)

			_, err := d.validate("auth.example.com")

			require.Error(t, err)
			assert.Equal(t, int32(1), flaky.attempts.Load())
		})
	}
}

func TestDestinationValidateShouldNotRetryARefusedOrigin(t *testing.T) {
	capture := &optionsCapture{status: http.StatusOK}

	server := httptest.NewServer(capture.handler())

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	_, err := d.validate("auth.example.com")

	require.Error(t, err)
	assert.Equal(t, int32(1), capture.requests.Load())
}

func TestDestinationValidateShouldRetryATransportFailure(t *testing.T) {
	var attempts atomic.Int32

	server := httptest.NewUnstartedServer(nil)

	server.Config.Handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			rw.WriteHeader(http.StatusOK)

			return
		}

		if attempts.Add(1) == 1 {
			conn, _, err := rw.(http.Hijacker).Hijack()
			if err == nil {
				_ = conn.Close()
			}

			return
		}

		rw.Header().Set(events.HeaderWebhookAllowedOrigin, "*")
		rw.WriteHeader(http.StatusOK)
	})

	server.Start()

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	_, err := d.validate("auth.example.com")

	require.NoError(t, err, "a failure which produced no status code must be retried")
	assert.Equal(t, int32(2), attempts.Load())
}

func TestDestinationValidateShouldNotExceedItsBudget(t *testing.T) {
	flaky := &flakyOptions{statuses: []int{http.StatusTooManyRequests, http.StatusTooManyRequests}, retryAfter: "600"}

	server := httptest.NewServer(flaky.handler())

	defer server.Close()

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Retry = schema.WebhookRetry{Attempts: 3, InitialInterval: time.Minute, MaximumInterval: time.Hour}
	})

	start := time.Now()

	_, err := d.validate("auth.example.com")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "handshake budget")
	assert.Less(t, time.Since(start), time.Second*5, "the budget must stop the retry rather than honor a long Retry-After")
}

func TestDispatcherShouldHoldADestinationWhichWithholdsConfirmationUntilTheCallback(t *testing.T) {
	var (
		deliveries atomic.Int32
		callback   atomic.Value
	)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			callback.Store(r.Header.Get(events.HeaderWebhookRequestCallback))
		} else {
			deliveries.Add(1)
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	recorder := &testRecorder{}

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	dispatcher.SetMetrics(recorder)

	require.NoError(t, dispatcher.Validate())
	require.True(t, dispatcher.destinations[0].pending)
	require.False(t, dispatcher.destinations[0].refused)

	target, err := url.Parse(callback.Load().(string))

	require.NoError(t, err)
	assert.Equal(t, "https", target.Scheme)
	assert.Equal(t, "auth.example.com", target.Host)
	assert.Equal(t, PathConfirm, target.Path)
	assert.Equal(t, "admin-api", target.Query().Get(QueryConfirmID))
	assert.Len(t, target.Query().Get(QueryConfirmKey), 64)

	dispatcher.Start()

	require.NoError(t, dispatcher.StartupCheck())

	dispatcher.Emit(context.Background(), testEvent())

	assert.Equal(t, int32(0), deliveries.Load(), "a destination which has not confirmed must receive nothing")
	assert.Equal(t, 0, recorder.count(outcomeDropped))

	assert.False(t, dispatcher.Confirm("admin-api", "incorrect", ""))
	assert.False(t, dispatcher.Confirm("unknown", target.Query().Get(QueryConfirmKey), ""))
	assert.True(t, dispatcher.destinations[0].pending)

	assert.True(t, dispatcher.Confirm("admin-api", target.Query().Get(QueryConfirmKey), "120"))
	assert.False(t, dispatcher.destinations[0].pending)
	assert.Equal(t, time.Minute/120, dispatcher.destinations[0].pace)

	assert.True(t, dispatcher.Confirm("admin-api", target.Query().Get(QueryConfirmKey), ""), "a repeated confirmation is accepted")

	dispatcher.Emit(context.Background(), testEvent())

	assert.Eventually(t, func() bool { return deliveries.Load() == 1 }, time.Second*5, time.Millisecond*10)

	dispatcher.Shutdown()
}

func TestDispatcherShouldHoldADestinationWhichDoesNotSupportTheHandshake(t *testing.T) {
	var deliveries atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			rw.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		deliveries.Add(1)

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	err := dispatcher.Validate()

	require.EqualError(t, err, "destination 'admin-api': received status code 405")
	require.True(t, dispatcher.destinations[0].pending)
	require.False(t, dispatcher.destinations[0].refused)

	dispatcher.Start()

	dispatcher.Emit(context.Background(), testEvent())

	assert.Equal(t, int32(0), deliveries.Load())

	assert.True(t, dispatcher.Confirm("admin-api", dispatcher.destinations[0].key, ""))

	dispatcher.Emit(context.Background(), testEvent())

	assert.Eventually(t, func() bool { return deliveries.Load() == 1 }, time.Second*5, time.Millisecond*10)

	dispatcher.Shutdown()
}

func TestDispatcherShouldRememberACallbackConfirmationAcrossRestarts(t *testing.T) {
	var deliveries atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			deliveries.Add(1)
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	store := newTestStore()

	first := newValidationDispatcher(t, server.URL, nil)

	first.SetStore(store)

	require.NoError(t, first.Validate())
	require.True(t, first.destinations[0].pending)

	require.True(t, first.Confirm("admin-api", first.destinations[0].key, "120"))
	require.Len(t, store.grants(), 1)
	assert.LessOrEqual(t, len(store.grants()[0]), 20)

	first.Shutdown()

	second := newValidationDispatcher(t, server.URL, nil)

	second.SetStore(store)

	require.NoError(t, second.Validate())

	assert.False(t, second.destinations[0].pending)
	assert.False(t, second.destinations[0].refused)
	assert.Equal(t, time.Minute/120, second.destinations[0].pace)
	assert.Equal(t, first.destinations[0].key, second.destinations[0].key)

	second.Start()

	second.Emit(context.Background(), testEvent())

	assert.Eventually(t, func() bool { return deliveries.Load() == 1 }, time.Second*5, time.Millisecond*10)

	second.Shutdown()
}

func TestDispatcherShouldNotReuseAConfirmationForAnotherAddress(t *testing.T) {
	handler := http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	})

	original, moved := httptest.NewServer(handler), httptest.NewServer(handler)

	defer original.Close()
	defer moved.Close()

	store := newTestStore()

	first := newValidationDispatcher(t, original.URL, nil)

	first.SetStore(store)

	require.NoError(t, first.Validate())
	require.True(t, first.Confirm("admin-api", first.destinations[0].key, ""))

	second := newValidationDispatcher(t, moved.URL, nil)

	second.SetStore(store)

	require.NoError(t, second.Validate())

	assert.True(t, second.destinations[0].pending)
}

func TestDispatcherShouldForgetAConfirmationWhenTheDestinationRefuses(t *testing.T) {
	var refuse atomic.Bool

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if refuse.Load() {
			rw.Header().Set(events.HeaderWebhookAllowedOrigin, "other.example.com")
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	store := newTestStore()

	first := newValidationDispatcher(t, server.URL, nil)

	first.SetStore(store)

	require.NoError(t, first.Validate())
	require.True(t, first.Confirm("admin-api", first.destinations[0].key, ""))
	require.Len(t, store.grants(), 1)

	refuse.Store(true)

	second := newValidationDispatcher(t, server.URL, nil)

	second.SetStore(store)

	require.Error(t, second.Validate())

	assert.True(t, second.destinations[0].refused)
	assert.Empty(t, store.grants())
}

func TestDispatcherShouldForgetAConfirmationWhenTheDestinationIsGone(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			rw.WriteHeader(http.StatusOK)

			return
		}

		rw.WriteHeader(http.StatusGone)
	}))

	defer server.Close()

	store := newTestStore()

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	dispatcher.SetStore(store)

	require.NoError(t, dispatcher.Validate())
	require.True(t, dispatcher.Confirm("admin-api", dispatcher.destinations[0].key, ""))
	require.Len(t, store.grants(), 1)

	dispatcher.Start()

	dispatcher.Emit(context.Background(), testEvent())

	assert.Eventually(t, func() bool { return len(store.grants()) == 0 }, time.Second*5, time.Millisecond*10)
	assert.True(t, dispatcher.destinations[0].gone.Load())

	dispatcher.Shutdown()
}

func TestDispatcherShouldTolerateAStoreWhichFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	store := newTestStore()

	store.err = fmt.Errorf("the storage backend is unavailable")

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	dispatcher.SetStore(store)

	require.NoError(t, dispatcher.Validate())
	require.True(t, dispatcher.destinations[0].pending)

	assert.True(t, dispatcher.Confirm("admin-api", dispatcher.destinations[0].key, ""))
	assert.False(t, dispatcher.destinations[0].pending)
}

func TestDestinationValidateShouldNotCarryEventHeaders(t *testing.T) {
	var headers http.Header

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		headers = r.Header.Clone()

		rw.Header().Set(events.HeaderWebhookAllowedOrigin, r.Header.Get(events.HeaderWebhookRequestOrigin))
		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	_, err := d.validate("auth.example.com")

	require.NoError(t, err)
	require.NotNil(t, headers)

	assert.Empty(t, cloudEventsHeaders(headers))
	assert.Empty(t, headers.Get(events.HeaderEventCount))
	assert.Empty(t, headers.Get(events.HeaderSignature))
	assert.Empty(t, headers.Get("Content-Type"))
	assert.Equal(t, "1", headers.Get(events.HeaderDeliveryAttempt))
	assert.Equal(t, "auth.example.com", headers.Get(events.HeaderWebhookRequestOrigin))
}

func TestDispatcherConfirmShouldFallBackToTheRequestedRate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	dispatcher := newValidationDispatcher(t, server.URL, func(config *schema.WebhookDestination) {
		config.Validation.Rate = 60
	})

	require.NoError(t, dispatcher.Validate())
	require.True(t, dispatcher.destinations[0].pending)

	assert.True(t, dispatcher.Confirm("admin-api", dispatcher.destinations[0].key, ""))
	assert.Equal(t, time.Second, dispatcher.destinations[0].pace)
}

func TestDispatcherShouldAcceptACallbackWhichArrivesDuringTheHandshake(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			close(entered)

			<-release
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	validated := make(chan error, 1)

	go func() { validated <- dispatcher.Validate() }()

	<-entered

	assert.True(t, dispatcher.Confirm("admin-api", dispatcher.destinations[0].key, ""))

	close(release)

	require.NoError(t, <-validated)
	assert.False(t, dispatcher.destinations[0].pending)
	assert.False(t, dispatcher.destinations[0].refused)
}

func TestDispatcherConfirmShouldNotGrantARefusedDestination(t *testing.T) {
	capture := &optionsCapture{status: http.StatusOK, allowOrigin: "other.example.com"}

	server := httptest.NewServer(capture.handler())

	defer server.Close()

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	require.Error(t, dispatcher.Validate())

	assert.False(t, dispatcher.Confirm("admin-api", dispatcher.destinations[0].key, ""))
	assert.True(t, dispatcher.destinations[0].refused)
}

func TestDispatcherShouldNotOfferACallbackWhenValidationIsDisabled(t *testing.T) {
	dispatcher := newValidationDispatcher(t, "https://admin.example.com/hooks", func(config *schema.WebhookDestination) {
		config.Validation.Disable = true
	})

	dispatcher.SetStore(newTestStore())

	require.NoError(t, dispatcher.Validate())

	assert.Empty(t, dispatcher.destinations[0].key)
	assert.Empty(t, dispatcher.destinations[0].callback)
	assert.False(t, dispatcher.Confirm("admin-api", "", ""))
}

func TestDispatcherShouldShareTheCallbackKeyBetweenInstances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	store := newTestStore()

	first, second := newValidationDispatcher(t, server.URL, nil), newValidationDispatcher(t, server.URL, nil)

	first.SetStore(store)
	second.SetStore(store)

	require.NoError(t, first.Validate())
	require.NoError(t, second.Validate())

	require.NotEmpty(t, first.destinations[0].key)
	assert.Equal(t, first.destinations[0].key, second.destinations[0].key)
	assert.Equal(t, first.destinations[0].callback, second.destinations[0].callback)

	assert.True(t, second.Confirm("admin-api", first.destinations[0].key, ""))
	assert.False(t, second.destinations[0].pending)
}

func TestDispatcherShouldPickUpAConfirmationMadeOnAnotherInstance(t *testing.T) {
	var deliveries atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			deliveries.Add(1)
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	store := newTestStore()

	first, second := newValidationDispatcher(t, server.URL, nil), newValidationDispatcher(t, server.URL, nil)

	first.SetStore(store)
	second.SetStore(store)

	second.recheck = time.Millisecond * 20

	require.NoError(t, first.Validate())
	require.NoError(t, second.Validate())

	second.Start()

	require.True(t, first.Confirm("admin-api", first.destinations[0].key, "120"))

	require.Eventually(t, func() bool {
		second.mutex.RLock()
		defer second.mutex.RUnlock()

		return !second.destinations[0].pending
	}, time.Second*5, time.Millisecond*10)

	second.Emit(context.Background(), testEvent())

	assert.Eventually(t, func() bool { return deliveries.Load() == 1 }, time.Second*5, time.Millisecond*10)

	second.Shutdown()
}

func TestDispatcherShouldUseDifferentCallbackKeysWithoutAStore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	first, second := newValidationDispatcher(t, server.URL, nil), newValidationDispatcher(t, server.URL, nil)

	require.NoError(t, first.Validate())
	require.NoError(t, second.Validate())

	assert.NotEqual(t, first.destinations[0].key, second.destinations[0].key)
	assert.False(t, second.Confirm("admin-api", first.destinations[0].key, ""))
}

func TestDispatcherCallbackKeyShouldBeUniqueToTheDestinationAndAddress(t *testing.T) {
	key := func(store Store, name, address string, modify func(config *schema.WebhookDestination)) string {
		dispatcher := newValidationDispatcher(t, address, func(config *schema.WebhookDestination) {
			config.Name = name

			if modify != nil {
				modify(config)
			}
		})

		if store != nil {
			dispatcher.SetStore(store)
		}

		return dispatcher.callbackKey(dispatcher.destinations[0])
	}

	store := newTestStore()

	base := key(store, "admin-api", "https://admin.example.com/hooks", nil)

	assert.Len(t, base, 64)
	assert.Equal(t, base, key(store, "admin-api", "https://admin.example.com/hooks", nil))
	assert.NotEqual(t, base, key(store, "other-api", "https://admin.example.com/hooks", nil))
	assert.NotEqual(t, base, key(store, "admin-api", "https://other.example.com/hooks", nil))
	assert.NotEqual(t, base, key(newTestStore(), "admin-api", "https://admin.example.com/hooks", nil))
	assert.NotEqual(t, base, key(nil, "admin-api", "https://admin.example.com/hooks", nil))
	assert.Empty(t, key(store, "admin-api", "https://admin.example.com/hooks", func(config *schema.WebhookDestination) {
		config.Validation.Disable = true
	}))
}

type blockingHandshake struct {
	accept bool

	entered chan struct{}
	release chan struct{}

	enter, once sync.Once
	deliveries  atomic.Int32
}

func newBlockingHandshake(accept bool) *blockingHandshake {
	return &blockingHandshake{accept: accept, entered: make(chan struct{}), release: make(chan struct{})}
}

func (b *blockingHandshake) unblock() {
	b.once.Do(func() { close(b.release) })
}

func (b *blockingHandshake) handler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			b.deliveries.Add(1)

			rw.WriteHeader(http.StatusOK)

			return
		}

		b.enter.Do(func() { close(b.entered) })

		select {
		case <-b.release:
		case <-r.Context().Done():
			return
		}

		if b.accept {
			rw.Header().Set(events.HeaderWebhookAllowedOrigin, r.Header.Get(events.HeaderWebhookRequestOrigin))
		} else {
			rw.Header().Set(events.HeaderWebhookAllowedOrigin, "other.example.com")
		}

		rw.WriteHeader(http.StatusOK)
	}
}

func newValidationDispatcher(t *testing.T, address string, modify func(config *schema.WebhookDestination)) *Dispatcher {
	t.Helper()

	parsed, err := url.Parse(address)

	require.NoError(t, err)

	config := &schema.Configuration{}
	config.Session.Cookies = []schema.SessionCookie{{
		Domain:      "example.com",
		AutheliaURL: &url.URL{Scheme: "https", Host: "auth.example.com"},
	}}

	destination := schema.WebhookDestination{
		Name: "admin-api", Address: parsed, BufferSize: 8, Timeout: time.Second * 5,
		Events: []string{events.TypeUserPasswordChanged},
		Retry:  schema.WebhookRetry{Attempts: 1, InitialInterval: time.Millisecond, MaximumInterval: time.Millisecond},
		TLS:    &schema.TLS{},
	}

	if modify != nil {
		modify(&destination)
	}

	config.Webhooks.Destinations = []schema.WebhookDestination{destination}

	return NewDispatcher(config, nil, logrus.NewEntry(logrus.New()))
}

type optionsCapture struct {
	allowOrigin string
	allowRate   string
	status      int

	requests atomic.Int32
	origin   atomic.Value
	rate     atomic.Value
	auth     atomic.Value
}

func (c *optionsCapture) handler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			rw.WriteHeader(http.StatusOK)

			return
		}

		c.requests.Add(1)
		c.origin.Store(r.Header.Get(events.HeaderWebhookRequestOrigin))
		c.rate.Store(r.Header.Get(events.HeaderWebhookRequestRate))
		c.auth.Store(r.Header.Get("Authorization"))

		if c.allowOrigin != "" {
			rw.Header().Set(events.HeaderWebhookAllowedOrigin, c.allowOrigin)
		}

		if c.allowRate != "" {
			rw.Header().Set(events.HeaderWebhookAllowedRate, c.allowRate)
		}

		rw.WriteHeader(c.status)
	}
}

type flakyOptions struct {
	statuses   []int
	retryAfter string

	attempts atomic.Int32
	seen     []string
	mu       sync.Mutex
}

func (f *flakyOptions) handler() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodOptions {
			rw.WriteHeader(http.StatusOK)

			return
		}

		n := int(f.attempts.Add(1))

		f.mu.Lock()
		f.seen = append(f.seen, r.Header.Get(events.HeaderDeliveryAttempt))
		f.mu.Unlock()

		if n <= len(f.statuses) {
			if f.retryAfter != "" {
				rw.Header().Set("Retry-After", f.retryAfter)
			}

			rw.WriteHeader(f.statuses[n-1])

			return
		}

		rw.Header().Set(events.HeaderWebhookAllowedOrigin, "*")
		rw.WriteHeader(http.StatusOK)
	}
}

func (f *flakyOptions) order() []string {
	f.mu.Lock()

	defer f.mu.Unlock()

	return append([]string{}, f.seen...)
}

func newTestStore() *testStore {
	secret := make([]byte, 32)

	_, _ = rand.Read(secret)

	return &testStore{values: map[string][]byte{}, secret: secret}
}

type testStore struct {
	mutex  sync.Mutex
	values map[string][]byte
	secret []byte
	err    error
}

func (s *testStore) WebhookCallbackSignature(values ...[]byte) string {
	mac := hmac.New(sha256.New, s.secret)

	for _, value := range values {
		mac.Write(value)
	}

	return hex.EncodeToString(mac.Sum(nil))
}

func (s *testStore) LoadCachedData(_ context.Context, name string) (*model.CachedData, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.err != nil {
		return nil, s.err
	}

	value, ok := s.values[name]
	if !ok {
		return nil, nil
	}

	return &model.CachedData{Name: name, Value: value}, nil
}

func (s *testStore) SaveCachedData(_ context.Context, data model.CachedData) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.err != nil {
		return s.err
	}

	s.values[data.Name] = data.Value

	return nil
}

func (s *testStore) DeleteCachedData(_ context.Context, name string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.err != nil {
		return s.err
	}

	delete(s.values, name)

	return nil
}

func (s *testStore) grants() (names []string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for name := range s.values {
		names = append(names, name)
	}

	return names
}
