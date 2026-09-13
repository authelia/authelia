// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/authelia/authelia/v4/internal/random"
)

const (
	// RecoveryCodeLength is the number of random characters in a recovery code, not counting the visual hyphen.
	RecoveryCodeLength = 10

	// RecoveryCodeBatchSize is the number of recovery codes generated per request.
	RecoveryCodeBatchSize = 10
)

// NewRecoveryCode returns a new RecoveryCode for the given username with a freshly generated plaintext value.
// The plaintext is held transiently on the returned struct's Plaintext field for the caller to display once
// to the user; only the HMAC of the normalized plaintext is ever persisted.
func NewRecoveryCode(ctx Context, username string) (code *RecoveryCode, err error) {
	src := ctx.GetRandom()

	var raw []byte

	if raw, err = src.BytesCustomErr(RecoveryCodeLength, []byte(random.CharSetUnambiguousUpper)); err != nil {
		return nil, fmt.Errorf("failed to generate recovery code random bytes: %w", err)
	}

	mid := RecoveryCodeLength / 2

	return &RecoveryCode{
		Username:  username,
		CreatedAt: ctx.GetClock().Now(),
		Plaintext: string(raw[:mid]) + "-" + string(raw[mid:]),
	}, nil
}

// NormalizeRecoveryCode returns the canonical form of a user-supplied recovery code: uppercased, with whitespace,
// hyphens, and underscores stripped. The HMAC signature is computed over this canonical form so that user-friendly
// formatting variations all match the same stored row.
func NormalizeRecoveryCode(input string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\n', '\r', '-', '_':
			return -1
		}

		return unicode.ToUpper(r)
	}, input)
}

// RecoveryCode represents a single user recovery code stored in the database. Only the HMAC signature of the
// normalized plaintext is persisted; the raw code is shown to the user once at generation time and then discarded.
type RecoveryCode struct {
	ID         int          `db:"id"`
	Username   string       `db:"username"`
	Signature  string       `db:"signature"`
	CreatedAt  time.Time    `db:"created_at"`
	ConsumedAt sql.NullTime `db:"consumed_at"`
	ConsumedIP NullIP       `db:"consumed_ip"`
	RevokedAt  sql.NullTime `db:"revoked_at"`
	RevokedIP  NullIP       `db:"revoked_ip"`

	// Plaintext is the raw code value, populated only transiently after generation for display to the user.
	// It is never read from or written to the database; the storage layer derives Signature from it before insert.
	Plaintext string `db:"-"`
}

// Consume marks the recovery code as used by stamping the consumption time and remote IP from ctx.
func (c *RecoveryCode) Consume(ctx Context) {
	c.ConsumedAt = sql.NullTime{Valid: true, Time: ctx.GetClock().Now()}
	c.ConsumedIP = NewNullIP(ctx.RemoteIP())
}

// Revoke marks the recovery code as revoked by stamping the revocation time and remote IP from ctx.
// Used when the user regenerates their codes; existing rows are revoked rather than deleted to preserve audit history.
func (c *RecoveryCode) Revoke(ctx Context) {
	c.RevokedAt = sql.NullTime{Valid: true, Time: ctx.GetClock().Now()}
	c.RevokedIP = NewNullIP(ctx.RemoteIP())
}
