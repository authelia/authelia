package models

import (
	"time"
)

// AuthenticationAttempt represent an authentication attempt.
type AuthenticationAttempt struct {
	Username   string    `db:"username"`
	Successful bool      `db:"successful"`
	Time       time.Time `db:"time"`
}
