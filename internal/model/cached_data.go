// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package model

import "time"

// CachedData represents a cached data row in the storage provider.
type CachedData struct {
	ID        int       `db:"id"`
	Created   time.Time `db:"created_at"`
	Updated   time.Time `db:"updated_at"`
	Name      string    `db:"name"`
	Encrypted bool      `db:"encrypted"`
	Value     []byte    `db:"value"`
}
