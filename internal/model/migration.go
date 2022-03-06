package model

import (
	"time"
)

// Migration represents a migration row in the database.
type Migration struct {
	ID      int       `db:"id"`
	Applied time.Time `db:"applied"`
	Before  int       `db:"version_before"`
	After   int       `db:"version_after"`
	Version string    `db:"application_version"`
}
