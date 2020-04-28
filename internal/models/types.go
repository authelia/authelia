package models

import "time"

// AuthenticationAttempt represent an authentication attempt.
type AuthenticationAttempt struct {
	// The user who tried to authenticate.
	Username string
	// Successful true if the attempt was successful.
	Successful bool
	// The time of the attempt.
	Time time.Time
}
