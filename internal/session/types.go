package session

import (
	"context"

	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/redis"
	"github.com/sirupsen/logrus"
	"github.com/tstranex/u2f"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/authorization"
	"github.com/authelia/authelia/v4/internal/logging"
)

// ProviderConfig is the configuration used to create the session provider.
type ProviderConfig struct {
	config              session.Config
	redisConfig         *redis.Config
	redisSentinelConfig *redis.FailoverConfig
	providerName        string
}

// U2FRegistration is a serializable version of a U2F registration.
type U2FRegistration struct {
	KeyHandle []byte
	PublicKey []byte
}

// UserSession is the structure representing the session of a user.
type UserSession struct {
	Username string

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

	// This string is set to the username after identity verification and checked
	// while doing the query actually updating the password.
	PasswordResetUsername *string
}

// Identity identity of the user who is being verified.
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

func newRedisLogger() *redisLogger {
	return &redisLogger{logger: logging.Logger()}
}

type redisLogger struct {
	logger *logrus.Logger
}

func (l *redisLogger) Printf(_ context.Context, format string, v ...interface{}) {
	l.logger.Tracef(format, v...)
}
