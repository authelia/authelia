package session

import (
	"net"
	"time"

	"github.com/fasthttp/session/v2"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/oidc"
)

// ProviderConfig is the configuration used to create the session provider.
type ProviderConfig struct {
	config       session.Config
	providerName string
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

	AuthenticationMethodRefs oidc.AuthenticationMethodsReferences

	// Webauthn holds the session registration data for this session.
	Webauthn *webauthn.SessionData

	// This boolean is set to true after identity verification and checked
	// while doing the query actually updating the password.
	PasswordResetUsername *string

	RefreshTTL time.Time

	Elevations Elevations
}

type Elevations struct {
	User *Elevation
}

type Elevation struct {
	ID       int
	RemoteIP net.IP
	Expires  time.Time
}

// Identity identity of the user who is being verified.
type Identity struct {
	Username    string
	Email       string
	DisplayName string
}
