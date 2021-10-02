package models

// AuthenticationAttempt represents an authentication attempt.
type AuthenticationAttempt struct {
	Username   string `db:"username"`
	Successful bool   `db:"successful"`
	Time       Time   `db:"time"`
}
