package models

import (
	"time"
)

// IdentityVerification represents an identity verification row in the database.
type IdentityVerification struct {
	ID      int       `db:"id"`
	Created time.Time `db:"created"`
	Token   string    `db:"token"`
}
