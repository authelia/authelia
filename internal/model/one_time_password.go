package model

import (
	"database/sql"
	"net"
	"time"

	"github.com/google/uuid"
)

const (
	OTPIntentElevateUserSession = "eus"
)

// NewOneTimePassword returns a new OneTimePassword.
func NewOneTimePassword(publicID uuid.UUID, username, intent string, iat, exp time.Time, ip net.IP, value []byte) (otp OneTimePassword) {
	return OneTimePassword{
		PublicID:  publicID,
		IssuedAt:  iat,
		ExpiresAt: exp,
		Username:  username,
		Intent:    intent,
		IssuedIP:  NewIP(ip),
		Password:  value,
	}
}

// OneTimePassword represents special one time passwords stored in the database.
type OneTimePassword struct {
	ID         int          `db:"id"`
	PublicID   uuid.UUID    `db:"public_id"`
	Signature  string       `db:"signature"`
	IssuedAt   time.Time    `db:"iat"`
	IssuedIP   IP           `db:"issued_ip"`
	ExpiresAt  time.Time    `db:"exp"`
	Username   string       `db:"username"`
	Intent     string       `db:"intent"`
	Consumed   sql.NullTime `db:"consumed"`
	ConsumedIP NullIP       `db:"consumed_ip"`
	Revoked    sql.NullTime `db:"revoked"`
	RevokedIP  NullIP       `db:"revoked_ip"`
	Password   []byte       `db:"password"`
}
