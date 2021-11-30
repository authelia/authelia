package models

// DUODevice represents a DUO Device.
type DUODevice struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Device   string `db:"device"`
	Method   string `db:"method"`
}
