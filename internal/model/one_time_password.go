package model

import (
	"database/sql"
	"net"
	"time"
)

const (
	OTPIntentElevateUserSession = "eus"
)

// NewOneTimePassword returns a new OneTimePassword.
func NewOneTimePassword(username, intent string, iat, exp time.Time, ip net.IP, value []byte) (otp OneTimePassword) {
	return OneTimePassword{
		IssuedAt:  iat,
		ExpiresAt: exp,
		Username:  username,
		Intent:    intent,
		IssuedIP:  NewIP(ip),
		Password:  value,
	}
}

type OneTimePassword struct {
	ID         int          `db:"id"`
	Signature  string       `db:"signature"`
	IssuedAt   time.Time    `db:"iat"`
	IssuedIP   IP           `db:"issued_ip"`
	ExpiresAt  time.Time    `db:"exp"`
	Username   string       `db:"username"`
	Intent     string       `db:"intent"`
	Consumed   sql.NullTime `db:"consumed"`
	ConsumedIP NullIP       `db:"consumed_ip"`
	Password   []byte       `db:"password"`
}
