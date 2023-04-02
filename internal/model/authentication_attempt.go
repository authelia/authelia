// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package model

import (
	"time"
)

// AuthenticationAttempt represents an authentication attempt row in the database.
type AuthenticationAttempt struct {
	ID            int       `db:"id"`
	Time          time.Time `db:"time"`
	Successful    bool      `db:"successful"`
	Banned        bool      `db:"banned"`
	Username      string    `db:"username"`
	Type          string    `db:"auth_type"`
	RemoteIP      NullIP    `db:"remote_ip"`
	RequestURI    string    `db:"request_uri"`
	RequestMethod string    `db:"request_method"`
}
