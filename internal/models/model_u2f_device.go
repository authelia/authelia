package models

// U2FDevice represents a users U2F device.
type U2FDevice struct {
	Username  string `db:"username"`
	KeyHandle []byte `db:"key_handle"`
	PublicKey []byte `db:"public_key"`
}
