package clock

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Provider is an interface for a clock.
type Provider interface {
	Now() time.Time
	After(d time.Duration) <-chan time.Time
	AfterFunc(d time.Duration, f func()) *time.Timer
	GetJWTWithTimeFuncOption() (option jwt.ParserOption)
}
