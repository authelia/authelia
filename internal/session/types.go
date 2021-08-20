package session

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/tstranex/u2f"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
)

// Store interface implements the session.Provider with auxiliary methods.
type Store interface {
	Get(id []byte) (data []byte, err error)
	Save(id, data []byte, expiration time.Duration) (err error)
	Destroy(id []byte) (err error)
	Regenerate(id, newID []byte, expiration time.Duration) (err error)
	Count() (count int)
	NeedGC() (needsGC bool)
	GC() (err error)
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

	FirstFactorAuthnTimestamp  int64
	SecondFactorAuthnTimestamp int64

	// The challenge generated in first step of U2F registration (after identity verification) or authentication.
	// This is used reused in the second phase to check that the challenge has been completed.
	U2FChallenge *u2f.Challenge
	// The registration representing a U2F device in DB set after identity verification.
	// This is used in second phase of a U2F authentication.
	U2FRegistration *U2FRegistration

	// Represent an OIDC workflow session initiated by the client if not null.
	OIDCWorkflowSession *OIDCWorkflowSession

	// This boolean is set to true after identity verification and checked
	// while doing the query actually updating the password.
	PasswordResetUsername *string

	RefreshTTL time.Time
}

// Identity of the user who is being verified.
type Identity struct {
	Username string
	Email    string
}

// OIDCWorkflowSession represent an OIDC workflow session.
type OIDCWorkflowSession struct {
	ClientID                   string
	RequestedScopes            []string
	GrantedScopes              []string
	RequestedAudience          []string
	GrantedAudience            []string
	TargetURI                  string
	AuthURI                    string
	RequiredAuthorizationLevel authorization.Level
	CreatedTimestamp           int64
}

type redisLogger struct {
	logger *logrus.Logger
}

func (r *redisLogger) Printf(_ context.Context, format string, v ...interface{}) {
	if r.logger == nil {
		return
	}

	r.logger.Infof(format, v...)
}
