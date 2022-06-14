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
	authDuration     *prometheus.HistogramVec
	reqDuration      *prometheus.HistogramVec
	reqCounter       *prometheus.CounterVec
	reqVerifyCounter *prometheus.CounterVec
	auth1FACounter   *prometheus.CounterVec
	auth2FACounter   *prometheus.CounterVec
}

// RecordRequest takes the statusCode string, requestMethod string, and the elapsed time.Duration to record the request and request duration metrics.
func (r *Prometheus) RecordRequest(statusCode, requestMethod string, elapsed time.Duration) {
	r.reqCounter.WithLabelValues(statusCode, requestMethod).Inc()
	r.reqDuration.WithLabelValues(statusCode).Observe(elapsed.Seconds())
}

// RecordVerifyRequest takes the statusCode string to record the verify endpoint request metrics.
func (r *Prometheus) RecordVerifyRequest(statusCode string) {
	r.reqVerifyCounter.WithLabelValues(statusCode).Inc()
}

// RecordAuthentication takes the success and regulated booleans and a method string to record the authentication metrics.
func (r *Prometheus) RecordAuthentication(success, banned bool, authType string) {
	switch authType {
	case "1fa", "":
		r.auth1FACounter.WithLabelValues(strconv.FormatBool(success), strconv.FormatBool(banned)).Inc()
	default:
		r.auth2FACounter.WithLabelValues(strconv.FormatBool(success), strconv.FormatBool(banned), authType).Inc()
	}
}

// RecordAuthenticationDuration takes the statusCode string, requestMethod string, and the elapsed time.Duration to record the request and request duration metrics.
func (r *Prometheus) RecordAuthenticationDuration(success bool, elapsed time.Duration) {
	r.authDuration.WithLabelValues(strconv.FormatBool(success)).Observe(elapsed.Seconds())
}

func (r *Prometheus) register() {
	r.authDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "authelia",
			Name:      "authentication_duration",
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

	r.reqCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "request",
			Help:      "The number of HTTP requests processed.",
		},
		[]string{"code", "method"},
	)

	r.reqVerifyCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "verify_request",
			Help:      "The number of verify requests processed.",
		},
		[]string{"code"},
	)

	r.auth1FACounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "authentication_first_factor",
			Help:      "The number of 1FA authentications processed.",
		},
		[]string{"success", "banned"},
	)

	r.auth2FACounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "authentication_second_factor",
			Help:      "The number of 2FA authentications processed.",
		},
		[]string{"success", "banned", "type"},
	)
}
