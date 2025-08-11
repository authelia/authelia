package model

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/authelia/authelia/v4/internal/random"
)

const (
	// OTCIntentUserSessionElevation is the intent value for a one-time code indicating it's used for user session
	// elevation.
	OTCIntentUserSessionElevation = "use"
)

// NewOneTimeCode returns a new OneTimeCode.
func NewOneTimeCode(ctx Context, username string, characters int, duration time.Duration) (otp *OneTimeCode, err error) {
	var (
		publicID uuid.UUID
		code     []byte
	)

	src := ctx.GetRandom()

	if publicID, err = uuid.NewRandomFromReader(src); err != nil {
		return nil, fmt.Errorf("failed to generate public id: %w", err)
	}

	if code, err = src.BytesCustomErr(characters, []byte(random.CharSetUnambiguousUpper)); err != nil {
		return nil, fmt.Errorf("failed to generate random bytes: %w", err)
	}

	return &OneTimeCode{
		PublicID:  publicID,
		IssuedAt:  ctx.GetClock().Now(),
		IssuedIP:  NewIP(ctx.RemoteIP()),
		ExpiresAt: ctx.GetClock().Now().Add(duration),
		Username:  username,
		Intent:    OTCIntentUserSessionElevation,
		Code:      code,
	}, nil
}

// OneTimeCode represents special one-time codes stored in the database.
type OneTimeCode struct {
	ID         int          `db:"id"`
	PublicID   uuid.UUID    `db:"public_id"`
	Signature  string       `db:"signature"`
	IssuedAt   time.Time    `db:"issued"`
	IssuedIP   IP           `db:"issued_ip"`
	ExpiresAt  time.Time    `db:"expires"`
	Username   string       `db:"username"`
	Intent     string       `db:"intent"`
	ConsumedAt sql.NullTime `db:"consumed"`
	ConsumedIP NullIP       `db:"consumed_ip"`
	RevokedAt  sql.NullTime `db:"revoked"`
	RevokedIP  NullIP       `db:"revoked_ip"`
	Code       []byte       `db:"code"`
}

// Consume sets the values required to consume the one-time code.
func (otc *OneTimeCode) Consume(ctx Context) {
	otc.ConsumedAt = sql.NullTime{Valid: true, Time: ctx.GetClock().Now()}
	otc.ConsumedIP = NewNullIP(ctx.RemoteIP())
}
