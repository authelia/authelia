// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package model

// DuoDevice represents a DUO Device.
type DuoDevice struct {
	ID       int    `db:"id"`
	Username string `db:"username"`
	Device   string `db:"device"`
	Method   string `db:"method"`
}
