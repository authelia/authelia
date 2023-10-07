package clock

import (
	"time"
)

// New returns a new real clock.
func New() *Real {
	return &Real{}
}

// Real is the implementation of a clock.Provider for production.
type Real struct{}

// Now return the current time.
func (Real) Now() time.Time {
	return time.Now()
}

// After return a channel receiving the time after the defined duration.
func (Real) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (Real) AfterFunc(d time.Duration, f func()) *time.Timer {
	return time.AfterFunc(d, f)
}

var (
	_ Provider = (*Real)(nil)
)
