package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// NewPrometheus returns a new Prometheus metrics recorder.
func NewPrometheus() (provider *Prometheus) {
	provider = &Prometheus{}

	provider.register()

	return provider
}

// Prometheus is a middleware for recording prometheus metrics.
type Prometheus struct {
	authnDuration   *prometheus.HistogramVec
	reqDuration     *prometheus.HistogramVec
	reqDurationOIDC *prometheus.HistogramVec
	reqCounter      *prometheus.CounterVec
	authzCounter    *prometheus.CounterVec
	authnCounter    *prometheus.CounterVec
	authn2FACounter *prometheus.CounterVec
}

// RecordRequest takes the statusCode string, requestMethod string, and the elapsed time.Duration to record the request and request duration metrics.
func (r *Prometheus) RecordRequest(statusCode, requestMethod string, elapsed time.Duration) {
	r.reqCounter.WithLabelValues(statusCode, requestMethod).Inc()
	r.reqDuration.WithLabelValues(statusCode).Observe(elapsed.Seconds())
}

// RecordRequestOpenIDConnect takes the statusCode string, requestMethod string, and the elapsed time.Duration to record the request and request duration metrics.
func (r *Prometheus) RecordRequestOpenIDConnect(endpoint, statusCode string, elapsed time.Duration) {
	r.reqDurationOIDC.WithLabelValues(endpoint, statusCode).Observe(elapsed.Seconds())
}

// RecordAuthz takes the statusCode string to record the verify endpoint request metrics.
func (r *Prometheus) RecordAuthz(statusCode string) {
	r.authzCounter.WithLabelValues(statusCode).Inc()
}

// RecordAuthn takes the success and regulated booleans and a method string to record the authentication metrics.
func (r *Prometheus) RecordAuthn(success, banned bool, authType string) {
	switch authType {
	case "1fa", "":
		r.authnCounter.WithLabelValues(strconv.FormatBool(success), strconv.FormatBool(banned)).Inc()
	default:
		r.authn2FACounter.WithLabelValues(strconv.FormatBool(success), strconv.FormatBool(banned), authType).Inc()
	}
}

// RecordAuthenticationDuration takes the statusCode string, requestMethod string, and the elapsed time.Duration to record the request and request duration metrics.
func (r *Prometheus) RecordAuthenticationDuration(success bool, elapsed time.Duration) {
	r.authnDuration.WithLabelValues(strconv.FormatBool(success)).Observe(elapsed.Seconds())
}

func (r *Prometheus) register() {
	r.authnDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "authelia",
			Name:      "authn_duration",
			Help:      "The time an authentication attempt takes in seconds.",
			Buckets:   []float64{.0005, .00075, .001, .005, .01, .025, .05, .075, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.8, 0.9, 1, 5, 10, 15, 30, 60},
		},
		[]string{"success"},
	)

	r.reqDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "authelia",
			Name:      "request_duration",
			Help:      "The time a HTTP request takes to process in seconds.",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 20, 30, 40, 50, 60},
		},
		[]string{"code"},
	)

	r.reqDurationOIDC = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "authelia",
			Name:      "request_duration_openid_connect",
			Help:      "The time a HTTP request takes to process in seconds for the OpenID Connect 1.0 endpoints.",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 20, 30, 40, 50, 60},
		},
		[]string{"endpoint", "code"},
	)

	r.reqCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "request",
			Help:      "The number of HTTP requests processed.",
		},
		[]string{"code", "method"},
	)

	r.authzCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "authz",
			Help:      "The number of authz requests processed.",
		},
		[]string{"code"},
	)

	r.authnCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "authn",
			Help:      "The number of 1FA authentications processed.",
		},
		[]string{"success", "banned"},
	)

	r.authn2FACounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "authn_second_factor",
			Help:      "The number of 2FA authentications processed.",
		},
		[]string{"success", "banned", "type"},
	)
}
