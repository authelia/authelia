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

	// ExternalIdentity holds the in-flight external identity authorization state.
	ExternalIdentity *ExternalIdentityFlow `msg:"eid,omitempty"`

	// ExternalIdentityPending holds the validated claims of an unlinked external identity awaiting the users decision.
	ExternalIdentityPending *ExternalIdentityPending `msg:"eidp,omitempty"`
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
	FederatedIdentity            bool     `msg:"fed"`
	Extra                        []string `msg:"extra,omitempty"`
}

// ExternalIdentityFlow holds the in-flight external identity authorization state. It never leaves the
// server-side session store.
type ExternalIdentityFlow struct {
	Provider       string    `msg:"p"`
	State          string    `msg:"s"`
	Nonce          string    `msg:"n"`
	CodeVerifier   string    `msg:"cv"`
	Handle         string    `msg:"h"`
	Language       string    `msg:"l"`
	TargetURL      string    `msg:"t"`
	RequestMethod  string    `msg:"m"`
	KeepMeLoggedIn bool      `msg:"r"`
	Expires        time.Time `msg:"exp"`

	// FlowID, Flow, SubFlow, and UserCode identify the flow the user is resumed into once they have signed in, exactly
	// as the parameters of the same names do for a password sign in.
	FlowID   string `msg:"fid,omitempty"`
	Flow     string `msg:"f,omitempty"`
	SubFlow  string `msg:"sf,omitempty"`
	UserCode string `msg:"uc,omitempty"`
}

// ExternalIdentityPending holds validated claims from an unlinked external identity awaiting the user's decision. It
// holds no tokens.
type ExternalIdentityPending struct {
	Provider       string    `msg:"p"`
	Type           string    `msg:"typ"`
	Issuer         string    `msg:"iss"`
	Subject        string    `msg:"sub"`
	RemoteUsername string    `msg:"u"`
	DisplayName    string    `msg:"dn"`
	Email          string    `msg:"e"`
	Expires        time.Time `msg:"exp"`
}
