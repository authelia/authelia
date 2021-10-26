package models

// TOTPConfiguration represents a users TOTP configuration.
type TOTPConfiguration struct {
	ID        int    `db:"id"`
	Username  string `db:"username"`
	Algorithm string `db:"algorithm"`
	Digits    int    `db:"digits"`
	Period    uint64 `db:"totp_period"`
	Secret    string `db:"secret"`
}
