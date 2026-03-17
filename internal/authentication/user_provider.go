// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package authentication

import (
	"github.com/authelia/authelia/v4/internal/model"
)

// UserProvider is the interface for interacting with the authentication backends.
type UserProvider interface {
	model.StartupCheck

	// CheckUserPassword is used to check if a password matches for a specific user.
	CheckUserPassword(username string, password string) (valid bool, err error)

	// GetDetails is used to get a user's information directly from the backend. It must be used whenever the user is
	// authenticating, so that a user who was deleted, disabled, or otherwise changed is never authenticated with stale
	// information.
	GetDetails(username string) (details *UserDetails, err error)

	// GetDetailsCached is used to get a user's information from the cache if available. It's faster than GetDetails and
	// should be used for requests by a user who has already authenticated, as every authentication refreshes the cache.
	GetDetailsCached(username string) (details *UserDetails, err error)

	// GetDetailsExtended is used to get a user's information similar to GetDetails although with an extended suite of
	// attributes. It must be used whenever the user is authenticating.
	GetDetailsExtended(username string) (details *UserDetailsExtended, err error)

	// GetDetailsExtendedCached is used to get a user's extended information from the cache if available. It's faster
	// than GetDetailsExtended and should be used for requests by a user who has already authenticated.
	GetDetailsExtendedCached(username string) (details *UserDetailsExtended, err error)

	// UpdatePassword is used to change a user's password without verifying their old password.
	UpdatePassword(username string, newPassword string) (err error)

	// ChangePassword is used to change a user's password but requires their old password to be successfully verified.
	ChangePassword(username string, oldPassword string, newPassword string) (err error)

	// Close is used to close the provider and clean up any resources.
	Close() (err error)
}
