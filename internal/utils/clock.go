package utils

import "time"

// Clock is an interface for a clock.
type Clock interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
}

// RealClock is the implementation of a clock for production code.
type RealClock struct{}

// Now return the current time.
func (RealClock) Now() time.Time {
	return time.Now()
}

// After return a channel receiving the time after the defined duration.
func (RealClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// TestingClock implementation of clock for tests.
type TestingClock struct {
	now time.Time
}

// Now return the stored clock.
func (dc *TestingClock) Now() time.Time {
	return dc.now
}

// After return a channel receiving the time after duration has elapsed.
func (dc *TestingClock) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// Set set the time of the clock.
func (dc *TestingClock) Set(now time.Time) {
	dc.now = now
}
