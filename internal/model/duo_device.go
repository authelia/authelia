package model

// DuoDevice represents a DUO Device.
type DuoDevice struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Device   string `db:"device"`
	Method   string `db:"method"`
}
