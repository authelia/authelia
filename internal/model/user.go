package model

// "github.com/authelia/authelia/v4/internal/configuration/schema".

// User represents the user authentication details. used for authentication provider.
type User struct {
	ID int `db:"id"`

	// The user's username.
	Username string `db:"username"`

	// the user's password.
	Password []byte `db:"password"`

	// The users display name.
	DisplayName string `db:"display_name"`

	// The email for the user.
	Email string `db:"email"`

	// // The groups list for the user.
	Groups []string `db:"-"`

	// True if the user is disabled.
	Disabled bool `db:"disabled"`
}
