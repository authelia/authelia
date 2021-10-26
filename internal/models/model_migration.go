package models

import (
	"time"
)

type Migration struct {
	ID      int       `db:"id"`
	Time    time.Time `db:"time"`
	Prior   int       `db:"prior"`
	Current int       `db:"current"`
	Version string    `db:"version"`
}
