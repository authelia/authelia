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

// ErrNotPermittedOnAuthzEndpointRequest is returned when the session of a request to an authorization endpoint is
// saved, regenerated, or destroyed through the Strategy, as those endpoints only honor the authorization cookie.
var ErrNotPermittedOnAuthzEndpointRequest = errors.New("the session cookie is not honored on authorization endpoints")

// ErrNotPermittedOnPortalRequest is returned when the session of a request which isn't to an authorization endpoint
// is destroyed through the AuthzStrategy, as only those endpoints honor the authorization cookie.
var ErrNotPermittedOnPortalRequest = errors.New("the authorization cookie is only honored on authorization endpoints")

const (
	sessionIDLength = 32

	hkdfKeyInfoCodec = "authelia:kdf:session:codec:encryption_key:v1"

	cookieDeletionOffset = time.Hour * 24

	sessionCookieLifetimeMinimum = time.Hour * 24 * 30

	cookieNameSuffixAuthz = "_authz"
	authzValuePrefix      = "authelia:session:authz:v1:"

	issuerPrefixAuthz   = "authz:"
	cacheKeyPrefixAuthz = "authz:"

	authzIndexPublicIDLength = 36
)

var (
	expireUnlimited time.Time
)

// The names of the HMAC keys the session codec requires, which are the names they are stored under.
const (
	HMACKeyNameSession      = "session"
	HMACKeyNameSessionAuthz = "session_authz"
	HMACKeyNameSessionCSRF  = "session_csrf"
)
