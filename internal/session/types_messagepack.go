package session

import (
	"net"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/go-webauthn/webauthn/webauthn"

	"github.com/authelia/authelia/v4/internal/authorization"
)

//go:generate codecgen -o types_messagepack_gen.go types_messagepack.go

// UserSessionMessagePack is the MessagePack representation for UserSession.
type UserSessionMessagePack struct {
	Domain                     string                `json:"domain,omitempty"`
	Username                   string                `json:"username,omitempty"`
	KeepMeLoggedIn             bool                  `json:"remember,omitempty"`
	LastActivity               int64                 `json:"act,omitempty"`
	FirstFactorAuthnTimestamp  int64                 `json:"ffa,omitempty"`
	SecondFactorAuthnTimestamp int64                 `json:"mfa,omitempty"`
	AuthenticationMethodRefs   []string              `json:"amr,omitempty"`
	WebAuthn                   *WebAuthn             `json:"webauthn,omitempty"`
	TOTP                       *TOTPMessagePack      `json:"totp,omitempty"`
	PasswordResetUsername      *string               `json:"username_reset,omitempty"`
	RefreshTTL                 int64                 `json:"ttl,omitempty"`
	Elevations                 ElevationsMessagePack `json:"elevations,omitempty"`
}

// WebAuthnMessagePack is the MessagePack representation for WebAuthn.
type WebAuthnMessagePack struct {
	Challenge            string         `json:"chal"`
	RelyingPartyID       string         `json:"rpid"`
	UserID               []byte         `json:"uid,omitempty"`
	Description          string         `json:"desc,omitempty"`
	AllowedCredentialIDs [][]byte       `json:"allow,omitempty"`
	Expires              int64          `json:"exp"`
	UserVerification     string         `json:"uv"`
	Extensions           map[string]any `json:"ext,omitempty"`

	CredParams CredentialParametersMessagePack `json:"param,omitempty"`
}

type CredentialParametersMessagePack []CredentialParameterMessagePack

type CredentialParameterMessagePack struct {
	Type      string `json:"typ"`
	Algorithm int    `json:"alg"`
}

// ElevationsMessagePack is the MessagePack representation for Elevations.
type ElevationsMessagePack struct {
	User *ElevationMessagePack `json:"user,omitempty"`
}

// ElevationMessagePack is the MessagePack representation for Elevation.
type ElevationMessagePack struct {
	ID       int    `json:"id"`
	RemoteIP net.IP `json:"ip"`
	Expires  int64  `json:"exp"`
}

// TOTPMessagePack is the MessagePack representation for TOTP.
type TOTPMessagePack struct {
	Issuer    string `json:"iss,omitempty"`
	Algorithm string `json:"alg"`
	Digits    uint32 `json:"digits"`
	Period    uint   `json:"period"`
	Secret    string `json:"secret"`
	Expires   int64  `json:"exp"`
}

func (s *UserSession) ToMessagePack() (mp *UserSessionMessagePack) {
	if s == nil {
		return nil
	}

	return &UserSessionMessagePack{
		Domain:                     s.CookieDomain,
		Username:                   s.Username,
		KeepMeLoggedIn:             s.KeepMeLoggedIn,
		LastActivity:               s.LastActivity,
		FirstFactorAuthnTimestamp:  s.FirstFactorAuthnTimestamp,
		SecondFactorAuthnTimestamp: s.SecondFactorAuthnTimestamp,
		AuthenticationMethodRefs:   s.AuthenticationMethodRefs.MarshalRFC8176(),
		WebAuthn:                   s.WebAuthn,
		TOTP:                       s.TOTP.ToMessagePack(),
		PasswordResetUsername:      s.PasswordResetUsername,
		RefreshTTL:                 s.RefreshTTL.UnixMicro(),
		Elevations:                 s.Elevations.ToMessagePack(),
	}
}

func (s *UserSessionMessagePack) ToUserSession() *UserSession {
	if s == nil {
		return nil
	}

	return &UserSession{
		CookieDomain:               s.Domain,
		Username:                   s.Username,
		KeepMeLoggedIn:             s.KeepMeLoggedIn,
		LastActivity:               s.LastActivity,
		FirstFactorAuthnTimestamp:  s.FirstFactorAuthnTimestamp,
		SecondFactorAuthnTimestamp: s.SecondFactorAuthnTimestamp,
		AuthenticationMethodRefs:   authorization.NewAuthenticationMethodsReferencesFromClaim(s.AuthenticationMethodRefs),
		WebAuthn:                   s.WebAuthn,
		TOTP:                       s.TOTP.ToTOTP(),
		PasswordResetUsername:      s.PasswordResetUsername,
		RefreshTTL:                 time.UnixMicro(s.RefreshTTL).UTC(),
		Elevations:                 s.Elevations.ToElevations(),
	}
}

func (w *WebAuthn) ToMessagePack() *WebAuthnMessagePack {
	if w == nil {
		return nil
	}

	return &WebAuthnMessagePack{
		Challenge:            w.Challenge,
		RelyingPartyID:       w.RelyingPartyID,
		UserID:               w.UserID,
		Description:          w.Description,
		AllowedCredentialIDs: w.AllowedCredentialIDs,
		Expires:              w.Expires.UnixMicro(),
		UserVerification:     string(w.UserVerification),
		Extensions:           w.Extensions,
		CredParams:           NewCredentialParametersMessagePack(w.CredParams),
	}
}

func (w *WebAuthnMessagePack) ToWebAuthn() *WebAuthn {
	if w == nil {
		return nil
	}

	return &WebAuthn{
		SessionData: &webauthn.SessionData{
			Challenge:            w.Challenge,
			RelyingPartyID:       w.RelyingPartyID,
			UserID:               w.UserID,
			AllowedCredentialIDs: w.AllowedCredentialIDs,
			Expires:              time.UnixMicro(w.Expires).UTC(),
			UserVerification:     protocol.UserVerificationRequirement(w.UserVerification),
			Extensions:           w.Extensions,
			CredParams:           w.CredParams.ToCredentialParameters(),
		},
		Description: w.Description,
	}
}

func (c CredentialParametersMessagePack) ToCredentialParameters() (cp []protocol.CredentialParameter) {
	if c == nil {
		return nil
	}

	if len(c) == 0 {
		return []protocol.CredentialParameter{}
	}

	cp = make([]protocol.CredentialParameter, len(c))

	for i, param := range c {
		cp[i] = protocol.CredentialParameter{
			Type:      protocol.CredentialType(param.Type),
			Algorithm: webauthncose.COSEAlgorithmIdentifier(param.Algorithm),
		}
	}

	return cp
}

func NewCredentialParametersMessagePack(cp []protocol.CredentialParameter) (cpmp CredentialParametersMessagePack) {
	if cp == nil {
		return nil
	}

	if len(cp) == 0 {
		return CredentialParametersMessagePack{}
	}

	cpmp = make(CredentialParametersMessagePack, len(cp))

	for i, param := range cp {
		cpmp[i] = CredentialParameterMessagePack{
			Type:      string(param.Type),
			Algorithm: int(param.Algorithm),
		}
	}

	return cpmp
}

func (t *TOTP) ToMessagePack() *TOTPMessagePack {
	if t == nil {
		return nil
	}

	return &TOTPMessagePack{
		Issuer:    t.Issuer,
		Algorithm: t.Algorithm,
		Digits:    t.Digits,
		Period:    t.Period,
		Secret:    t.Secret,
		Expires:   t.Expires.UnixMicro(),
	}
}

func (t *TOTPMessagePack) ToTOTP() *TOTP {
	if t == nil {
		return nil
	}

	return &TOTP{
		Issuer:    t.Issuer,
		Algorithm: t.Algorithm,
		Digits:    t.Digits,
		Period:    t.Period,
		Secret:    t.Secret,
		Expires:   time.UnixMicro(t.Expires).UTC(),
	}
}

func (e Elevations) ToMessagePack() ElevationsMessagePack {
	return ElevationsMessagePack{
		User: e.User.ToMessagePack(),
	}
}

func (e ElevationsMessagePack) ToElevations() Elevations {
	return Elevations{
		User: e.User.ToElevation(),
	}
}

func (e *Elevation) ToMessagePack() *ElevationMessagePack {
	if e == nil {
		return nil
	}

	return &ElevationMessagePack{
		ID:       e.ID,
		RemoteIP: e.RemoteIP,
		Expires:  e.Expires.UnixMicro(),
	}
}

func (e *ElevationMessagePack) ToElevation() *Elevation {
	if e == nil {
		return nil
	}

	return &Elevation{
		ID:       e.ID,
		RemoteIP: e.RemoteIP,
		Expires:  time.UnixMicro(e.Expires).UTC(),
	}
}
