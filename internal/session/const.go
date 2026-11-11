// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"errors"
	"time"
)

// ErrRepositoryGet is returned when the Repository fails to retrieve a session, as distinct from a session which is
// absent or which can't be opened.
var ErrRepositoryGet = errors.New("error occurred getting session from backend")

// ErrSessionSuperseded is returned by a Repository which discarded a save because the session was moved to another
// identifier or destroyed after it was retrieved, such as when a request which loaded the session before it was
// regenerated or logged out completes afterwards.
var ErrSessionSuperseded = errors.New("session has been superseded by another session identifier or destroyed")

const (
	sessionIDLength = 32

	hkdfKeyInfoCodec = "authelia:kdf:session:codec:encryption_key:v1"

	cookieDeletionOffset = time.Hour * 24
)

var (
	expireUnlimited time.Time
)
