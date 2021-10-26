package models

import (
	"time"
)

type IdentityVerification struct {
	ID      int       `db:"id"`
	Created time.Time `db:"created"`
	Token   string    `db:"token"`
}
