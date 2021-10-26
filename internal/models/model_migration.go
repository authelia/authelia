package models

type Migration struct {
	ID      int    `db:"id"`
	Time    Time   `db:"time"`
	Prior   int    `db:"prior"`
	Current int    `db:"current"`
	Version string `db:"version"`
}
