package models

import (
	"time"
)

// AuthenticationAttempt represents an authentication attempt.
type AuthenticationAttempt struct {
	ID            int       `db:"id"`
	Time          time.Time `db:"time"`
	Successful    bool      `db:"successful"`
	Username      string    `db:"username"`
	Type          string    `db:"auth_type"`
	RemoteIP      IPAddress `db:"remote_ip"`
	RequestURI    string    `db:"request_uri"`
	RequestMethod string    `db:"request_method"`
}
