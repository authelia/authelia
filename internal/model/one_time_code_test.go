package model

import (
	"database/sql"
	"net"
	"testing"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/random"
	"github.com/stretchr/testify/assert"
)

func TestNewOneTimeCode(t *testing.T) {
	testCases := []struct {
		name       string
		username   string
		characters int
		duration   time.Duration
		expected   *OneTimeCode
		err        string
	}{
		{
			"Success",
			"username",
			1,
			time.Hour,
			&OneTimeCode{
				Username:  "username",
				ExpiresAt: time.Unix(1000000000, 0).Add(time.Hour),
				Intent:    OTCIntentUserSessionElevation,
				IssuedAt:  time.Unix(1000000000, 0),
				IssuedIP:  NewIP(net.ParseIP("127.0.0.1")),
			},
			"",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := &TestContext{
				ip:     net.ParseIP("127.0.0.1"),
				clock:  clock.NewFixed(time.Unix(1000000000, 0)),
				random: random.NewMathematical(),
			}

			actual, err := NewOneTimeCode(ctx, tc.username, tc.characters, tc.duration)

			if tc.err != "" {
				assert.EqualError(t, err, tc.err)
				assert.Nil(t, actual)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expected.ID, actual.ID)
				assert.Equal(t, tc.expected.Username, actual.Username)
				assert.Equal(t, tc.expected.ConsumedAt, actual.ConsumedAt)
				assert.Equal(t, tc.expected.ConsumedIP, actual.ConsumedIP)
				assert.Equal(t, tc.expected.ExpiresAt, actual.ExpiresAt)
				assert.Equal(t, tc.expected.Intent, actual.Intent)
				assert.Equal(t, tc.expected.IssuedAt, actual.IssuedAt)
				assert.Equal(t, tc.expected.IssuedIP, actual.IssuedIP)
				assert.Equal(t, tc.expected.RevokedAt, actual.RevokedAt)
				assert.Equal(t, tc.expected.RevokedIP, actual.RevokedIP)

				actual.Consume(ctx)

				assert.Equal(t, sql.NullTime{Time: ctx.clock.Now(), Valid: true}, actual.ConsumedAt)
				assert.Equal(t, NewNullIP(net.ParseIP("127.0.0.1")), actual.ConsumedIP)
			}
		})
	}
}

/*
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
*/
