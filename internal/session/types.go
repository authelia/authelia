package session

import (
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
	CookieDomain string

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

	// WebAuthn holds the session registration data for this session.
	WebAuthn *WebAuthn

	// This boolean is set to true after identity verification and checked
	// while doing the query actually updating the password.
	PasswordResetUsername *string

	RefreshTTL time.Time
}

// WebAuthn holds the standard webauthn session data plus some extra.
type WebAuthn struct {
	*webauthn.SessionData
	Description string
}

// Identity of the user who is being verified.
type Identity struct {
	Username    string
	Email       string
	DisplayName string
}
