package clock

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// NewFixed returns a new clock with an initial time.
func NewFixed(t time.Time) *Fixed {
	return &Fixed{now: t}
}

// Fixed implementation of clock.Provider for tests.
type Fixed struct {
	now time.Time
}

// Now return the stored clock.
func (c *Fixed) Now() time.Time {
	return c.now
}

// GetJWTWithTimeFuncOption returns the WithTimeFunc jwt.ParserOption.
func (c *Fixed) GetJWTWithTimeFuncOption() (option jwt.ParserOption) {
	return jwt.WithTimeFunc(c.Now)
}

// After return a channel receiving the time after duration has elapsed.
func (c *Fixed) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

func (c *Fixed) AfterFunc(d time.Duration, f func()) *time.Timer {
	return time.AfterFunc(d, f)
}

// Set the time of the clock.
func (c *Fixed) Set(now time.Time) {
	c.now = now
}

var (
	_ Provider = (*Fixed)(nil)
)
