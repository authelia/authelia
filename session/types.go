package session

import (
	"github.com/clems4ever/authelia/authentication"
	"github.com/fasthttp/session"
	"github.com/tstranex/u2f"
)

// ProviderConfig is the configuration used to create the session provider.
type ProviderConfig struct {
	config         *session.Config
	providerName   string
	providerConfig session.ProviderConfig
}

// UserSession is the structure representing the session of a user.
type UserSession struct {
	Username string
	// TODO(c.michaud): move groups out of the session.
	Groups []string
	Emails []string

	KeepMeLoggedIn      bool
	AuthenticationLevel authentication.Level
	LastActivity        int64

	// The challenge generated in first step of U2F registration (after identity verification) or authentication.
	// This is used reused in the second phase to check that the challenge has been completed.
	U2FChallenge *u2f.Challenge
	// The registration representing a U2F device in DB set after identity verification.
	// This is used in second phase of a U2F authentication.
	U2FRegistration *u2f.Registration

	// This boolean is set to true after identity verification and checked
	// while doing the query actually updating the password.
	PasswordResetUsername *string
}

// Identity identity of the user who is being verified.
type Identity struct {
	Username string
	Email    string
}
