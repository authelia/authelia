package clock

import (
	"time"
)

// Provider is an interface for a clock.
type Provider interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
	AfterFunc(d time.Duration, f func()) *time.Timer
}
