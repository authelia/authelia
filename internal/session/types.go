package session

import (
	"context"
	"time"

	"github.com/fasthttp/session/v2"
	"github.com/fasthttp/session/v2/providers/redis"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/sirupsen/logrus"

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

	// Webauthn holds the session registration data for this session.
	Webauthn *webauthn.SessionData

	// Represent an OIDC workflow session initiated by the client if not null.
	OIDCWorkflowSession *OIDCWorkflowSession

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
