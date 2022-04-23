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

	return provider
}

// Prometheus is a middleware for recording prometheus metrics.
type Prometheus struct {
	reqDuration      *prometheus.HistogramVec
	reqCounter       *prometheus.CounterVec
	reqVerifyCounter *prometheus.CounterVec
	auth1FACounter   *prometheus.CounterVec
	auth2FACounter   *prometheus.CounterVec
}

// RecordRequest takes the statusCode string, requestMethod string, and the elapsed time.Duration to record the request and request duration metrics.
func (p *Prometheus) RecordRequest(statusCode, requestMethod string, elapsed time.Duration) {
	if p.reqCounter == nil || p.reqDuration == nil {
		return
	}

	p.reqCounter.WithLabelValues(statusCode, requestMethod).Inc()
	p.reqDuration.WithLabelValues(statusCode).Observe(elapsed.Seconds())
}

// RecordVerifyRequest takes the statusCode string to record the verify endpoint request metrics.
func (p *Prometheus) RecordVerifyRequest(statusCode string) {
	if p.reqVerifyCounter == nil {
		return
	}

	p.reqVerifyCounter.WithLabelValues(statusCode).Inc()
}

// RecordAuthentication takes the success and regulated booleans and a method string to record the authentication metrics.
func (p *Prometheus) RecordAuthentication(success, banned bool, authType string) {
	switch authType {
	case "1fa", "":
		if p.auth1FACounter == nil {
			return
		}

		p.auth1FACounter.WithLabelValues(strconv.FormatBool(success), strconv.FormatBool(banned)).Inc()
	default:
		if p.auth2FACounter == nil {
			return
		}

		p.auth2FACounter.WithLabelValues(strconv.FormatBool(success), strconv.FormatBool(banned), authType).Inc()
	}
}

// Start the metrics recording process.
func (p *Prometheus) Start() {
	p.register()
}

func (p *Prometheus) register() {
	p.reqDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Subsystem: "authelia",
			Name:      "request_duration_seconds",
			Help:      "The time a HTTP request takes to process in seconds.",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 15, 20, 30, 40, 50, 60},
		},
		[]string{"code"},
	)

	p.reqCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "requests_total",
			Help:      "The number of HTTP requests processed.",
		},
		[]string{"code", "method"},
	)

	p.reqVerifyCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "verify_requests_total",
			Help:      "The number of verify requests processed.",
		},
		[]string{"code"},
	)

	p.auth1FACounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "authentication_first_factor",
			Help:      "The number of 1FA authentications processed.",
		},
		[]string{"success", "regulated"},
	)

	p.auth2FACounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "authelia",
			Name:      "authentication_second_factor",
			Help:      "The number of 2FA authentications processed.",
		},
		[]string{"success", "banned", "method"},
	)
}
