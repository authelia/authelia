// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"net"
	"time"

	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/authorization"
)

// IMPORTANT NOTE: Updating this file requires running the code generator.

//go:generate msgp

//msgp:replace authorization.AuthenticationMethodsReferences with:MessagePackAMR
//msgp:replace net.IP with:[]byte

// UserSession is the structure representing the session of a user.
type UserSession struct {
	CookieDomain string `msg:"d"`
	PublicID     string `msg:"p"`

	// Username for this session.
	//
	// SECURITY NOTE: This value MUST NOT be changed directly except within test files, and should instead be changed
	// by destroying the old session, and creating a new one.
	Username string `msg:"u,omitempty"`

	KeepMeLoggedIn bool  `msg:"r"`
	LastActivity   int64 `msg:"act"`

	FirstFactorAuthnTimestamp  int64 `msg:"ffa,omitempty"`
	SecondFactorAuthnTimestamp int64 `msg:"mfa,omitempty"`

	AuthenticationMethodRefs authorization.AuthenticationMethodsReferences `msg:"amr"`

	// WebAuthn holds the session registration data for this session.
	WebAuthn *WebAuthn `msg:"wa,omitempty"`
	TOTP     *TOTP     `msg:"otp,omitempty"`

	// PasswordResetUsername is given a value when a session exists where the password is being reset.
	PasswordResetUsername *string `msg:"pru,omitempty"`

	RefreshTTL time.Time `msg:"ttl"`

	Elevations Elevations `msg:"e"`

	// OpenIDConnectLogout is the validated OpenID Connect 1.0 RP-Initiated Logout request awaiting the confirmation of
	// the End-User.
	OpenIDConnectLogout *OpenIDConnectLogout `msg:"oidcl,omitempty"`
}

// OpenIDConnectLogout is a validated OpenID Connect 1.0 RP-Initiated Logout request. It is kept in the session rather
// than handed to the front-end so the destination of the logout can't be altered between validation and use.
type OpenIDConnectLogout struct {
	// FlowID identifies the redirect from the end session endpoint which this request belongs to. The request is only
	// presented to the End-User, or honored, when the logout page was reached through that redirect, so a logout the
	// End-User starts themselves is never mistaken for one a Relying Party requested.
	FlowID string `msg:"fid"`

	// ClientID is the identifier of the Relying Party which requested the logout.
	ClientID string `msg:"cid"`

	// RedirectURI is the validated post_logout_redirect_uri, if the request carried one.
	RedirectURI string `msg:"uri,omitempty"`

	// State is the state parameter to return to the RedirectURI, if the request carried one.
	State string `msg:"s,omitempty"`

	// Subject is the 'sub' claim of the id_token_hint, if the request carried one. It's the pairwise subject for the
	// sector of the Relying Party which requested the logout.
	Subject string `msg:"sub,omitempty"`

	// SessionID is the 'sid' claim of the id_token_hint, if the request carried one, which identifies the session the
	// Relying Party requested be ended.
	SessionID string `msg:"sid,omitempty"`

	// Expires is when the request lapses if it hasn't been confirmed or cancelled.
	Expires time.Time `msg:"exp"`
}

// WebAuthn holds the standard WebAuthn session data plus some extra.
type WebAuthn struct {
	*webauthn.SessionData

	Description string `json:"description,omitempty" msg:"desc,omitempty"`
}

// Elevations describes various session elevations.
type Elevations struct {
	User *Elevation `msg:"u,omitempty"`
}

// Elevation is an individual elevation.
type Elevation struct {
	ID       int       `msg:"id"`
	RemoteIP net.IP    `msg:"ip"`
	Expires  time.Time `msg:"exp"`
}

// TOTP holds the TOTP registration session data.
type TOTP struct {
	Issuer    string `msg:"iss,omitempty"`
	Algorithm string `msg:"alg"`
	Digits    uint32 `msg:"n"`
	Period    uint   `msg:"t"`
	Secret    string `msg:"k"`

	Expires time.Time `msg:"exp"`
}

// MessagePackAMR is the MessagePack representation of the authentication method references of a session.
type MessagePackAMR struct {
	KnowledgeBasedAuthentication bool     `msg:"kba"`
	UsernameAndPassword          bool     `msg:"pwd"`
	TOTP                         bool     `msg:"otp"`
	Duo                          bool     `msg:"duo"`
	WebAuthn                     bool     `msg:"wa"`
	WebAuthnHardware             bool     `msg:"hwk"`
	WebAuthnSoftware             bool     `msg:"swk"`
	WebAuthnUserPresence         bool     `msg:"up"`
	WebAuthnUserVerified         bool     `msg:"uv"`
	Extra                        []string `msg:"extra,omitempty"`
}
