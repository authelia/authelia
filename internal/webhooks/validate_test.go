// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
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
		{"ShouldRejectAMissingAllowedOrigin", 0, "", "", http.StatusOK, 0, "the response did not include a WebHook-Allowed-Origin header"},
		{"ShouldRejectAMismatchedOrigin", 0, "other.example.com", "", http.StatusOK, 0, "the response permitted the origin 'other.example.com' rather than 'auth.example.com'"},
		{"ShouldRejectANonSuccessStatus", 0, "auth.example.com", "", http.StatusNotFound, 0, "received status code 404"},
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
				config.Authentication.Bearer = &schema.WebhookAuthenticationBearer{Token: "abc", Header: "Authorization", Scheme: "Bearer"}
			})

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

	capture := &optionsCapture{status: http.StatusOK}

	inner := capture.handler()

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			probes.Add(1)
		}

		inner(rw, r)
	}))

	defer server.Close()

	dispatcher := newValidationDispatcher(t, server.URL, nil)

	require.Error(t, dispatcher.Validate(), "the receiver sends no allowed origin, so the destination is refused")
	require.True(t, dispatcher.destinations[0].refused)

	require.NoError(t, dispatcher.StartupCheck())
	assert.Equal(t, int32(0), probes.Load(),
		"a refused destination must not receive the startup probe, which would be the unsolicited payload the handshake prevents")
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
