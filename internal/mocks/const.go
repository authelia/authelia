// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package mocks

import (
	"github.com/authelia/authelia/v4/internal/session"
)

var testSessionHMACKeys = session.HMACKeys{
	Session:      []byte("cd21a70d4d24b23e0d1ae0a5cf6c4e1d3e2b8a5f7c9d0e1f2a3b4c5d6e7f8a9b"),
	SessionAuthz: []byte("7f3e9b1c5a2d8e4f6b0a9c3d7e1f5a2b8c4d0e6f3a9b1c5d7e2f8a4b0c6d3e9f"),
	SessionCSRF:  []byte("2b8e4a6c0d9f3e7b1a5c8d2e6f0a4b9c3d7e1f5a8b2c6d0e4f9a3b7c1d5e8f2a"),
}

// NewTestSessionHMACKeys returns the session HMAC keys the mocks use, for tests which build their own session provider.
func NewTestSessionHMACKeys() (keys session.HMACKeys) {
	return testSessionHMACKeys
}
