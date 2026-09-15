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

const (
	randomSessionChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_!#$%^*"

	hkdfKeyInfoCodec = "authelia:kdf:session:codec:encryption_key:v1"
)

const (
	// cookieDeletionOffset is how far in the past a deletion cookie expires, which must be sufficient to account for
	// clock skew between this server and the user agent.
	cookieDeletionOffset = time.Hour * 24
)

const (
	// CSRFCookieName is the name of the cookie which delivers the CSRF token for the current session to the user agent.
	CSRFCookieName = "authelia_csrf"

	// csrfTokenPrefix separates the data the CSRF token is derived from and the data the session signature is derived
	// from, as both are the session cookie value signed with the same key, and the latter is stored in the Repository.
	//
	//nolint:gosec
	csrfTokenPrefix = "authelia:csrf:"
)

var (
	expireUnlimited time.Time
)
