// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package model

// U2FDevice represents a users U2F device row in the database.
type U2FDevice struct {
	ID          int    `db:"id"`
	Username    string `db:"username"`
	Description string `db:"description"`
	KeyHandle   []byte `db:"key_handle"`
	PublicKey   []byte `db:"public_key"`
}
