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

var (
	// ErrCSRFTokenNoSession is returned when the request has no established session, and therefore has no CSRF token.
	ErrCSRFTokenNoSession = errors.New("the request has no established session")

	// ErrCSRFTokenInvalid is returned when the CSRF token doesn't match the CSRF token of the established session.
	ErrCSRFTokenInvalid = errors.New("the CSRF token is invalid")
)

const (
	randomSessionChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_!#$%^*"

	hkdfKeyInfoCodec = "authelia:kdf:session:codec:encryption_key:v1"

	cookieDeletionOffset = time.Hour * 24
)

const (
	// CSRFCookieNameSuffix is appended to the session cookie name to derive the name of the cookie which delivers the
	// CSRF token for the current session to the user agent, so that the token cookie of each session cookie domain is
	// distinctly named.
	CSRFCookieNameSuffix = "_csrf"

	//nolint:gosec
	csrfTokenPrefix = "authelia:csrf:"
)

var (
	expireUnlimited time.Time
)
