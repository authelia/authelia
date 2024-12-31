package model

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// User represents the user authentication details. used for authentication provider.
type User struct {
	// The user's username.
	Username string `db:"username" json:"username" valid:"required"`

	Password *schema.PasswordDigest `db:"password" json:"password"`

	// The users display name.
	DisplayName string `db:"display_name" json:"display_name"`

	// The email for the user.
	Email string `db:"email" json:"email"`

	// The groups list for the user.
	Groups []string `db:"groups" json:"groups"`

	// True if the user is disabled.
	Disabled bool `db:"disabled" json:"disabled"`
}
