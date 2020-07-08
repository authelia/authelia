package session

import (
	"time"

	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/redis"
	"github.com/tstranex/u2f"

	"github.com/authelia/authelia/internal/authentication"
)

// ProviderConfig is the configuration used to create the session provider.
type ProviderConfig struct {
	config       session.Config
	redisConfig  *redis.Config
	providerName string
}

// U2FRegistration is a serializable version of a U2F registration.
type U2FRegistration struct {
	KeyHandle []byte
	PublicKey []byte
}

// UserSession is the structure representing the session of a user.
type UserSession struct {
	Username    string
	DisplayName string
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
	U2FRegistration *U2FRegistration

	// This boolean is set to true after identity verification and checked
	// while doing the query actually updating the password.
	PasswordResetUsername *string

	RefreshTTL time.Time
}

// Identity identity of the user who is being verified.
type Identity struct {
	Username string
	Email    string
}
