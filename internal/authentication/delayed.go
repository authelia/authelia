package authentication

import (
	"time"

	"github.com/authelia/authelia/v4/internal/metrics"
)

// NewMiddleware creates a new Middleware UserProvider.
func NewMiddleware(provider UserProvider) *Middleware {
	return &Middleware{
		UserProvider: provider,
	}
}

// Middleware is a UserProvider which pre/post handles methods for another UserProvider
type Middleware struct {
	UserProvider

	delayer Delayer

	recorder metrics.Recorder
}

// ConfigureMetrics sets the metrics.Recorder for this Middleware.
func (m *Middleware) ConfigureMetrics(recorder metrics.Recorder) {
	if recorder == nil {
		return
	}

	m.recorder = recorder
}

// ConfigureDelayer sets the Delayer for this Middleware.
func (m *Middleware) ConfigureDelayer(delayer Delayer) {
	m.delayer = delayer
}

// CheckUserPassword adjusts the CheckUserPassword method of the underlying UserProvider.
func (m *Middleware) CheckUserPassword(username, password string) (valid bool, err error) {
	before := time.Now()

	valid, err = m.UserProvider.CheckUserPassword(username, password)

	success := valid && err == nil

	elapsed := time.Since(before)

	if m.recorder != nil {
		m.recorder.RecordAuthenticationDuration(success, elapsed)
	}

	if m.delayer != nil {
		m.delayer.Delay(success, elapsed)
	}

	return valid, err
}
