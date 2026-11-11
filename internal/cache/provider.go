// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"github.com/authelia/authelia/v4/internal/session"
)

// The Provider is implemented assuming all required encryption and washing is done prior to the values making it to
// the functions.
type Provider interface {
	session.Backend
}

// NewSessionRepository returns a session.Repository backed by the given Provider.
func NewSessionRepository(provider Provider) session.Repository {
	return session.NewRepository(provider)
}
