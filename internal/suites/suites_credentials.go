package suites

import (
	"sync"
	"time"

	"github.com/go-rod/rod/lib/proto"

	"github.com/authelia/otp/totp"
)

// NewRodSuiteCredentials returns a new initialized RodSuiteCredentials.
func NewRodSuiteCredentials() *RodSuiteCredentials {
	return &RodSuiteCredentials{
		lock: &sync.Mutex{},
		totp: map[string]RodSuiteCredentialOneTimePassword{},
	}
}

// RodSuiteCredentials is a concurrency safe store for the credentials used by a suite.
type RodSuiteCredentials struct {
	lock     *sync.Mutex
	totp     map[string]RodSuiteCredentialOneTimePassword
	webauthn RodSuiteCredentialWebAuthn
}

// RodSuiteCredentialWebAuthn represents the virtual WebAuthn authenticator and its credentials.
type RodSuiteCredentialWebAuthn struct {
	AuthenticatorID proto.WebAuthnAuthenticatorID
	Credentials     []*proto.WebAuthnCredential
}

// GetOneTimePassword returns the one time password credential for the given username.
func (rsc *RodSuiteCredentials) GetOneTimePassword(username string) RodSuiteCredentialOneTimePassword {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	return rsc.totp[username]
}

// SetOneTimePassword sets the one time password credential for the given username.
func (rsc *RodSuiteCredentials) SetOneTimePassword(username string, credential RodSuiteCredentialOneTimePassword) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	credential.valid = true

	rsc.totp[username] = credential
}

// DeleteOneTimePassword deletes the one time password credential for the given username.
func (rsc *RodSuiteCredentials) DeleteOneTimePassword(username string) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	rsc.totp[username] = RodSuiteCredentialOneTimePassword{
		valid: false,
	}
}

// GetWebAuthnAuthenticatorID returns the virtual WebAuthn authenticator id.
func (rsc *RodSuiteCredentials) GetWebAuthnAuthenticatorID() (authenticatorID proto.WebAuthnAuthenticatorID) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	return rsc.webauthn.AuthenticatorID
}

// GetWebAuthnCredentials returns the credentials registered against the virtual WebAuthn authenticator.
func (rsc *RodSuiteCredentials) GetWebAuthnCredentials() (credentials []*proto.WebAuthnCredential) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	return rsc.webauthn.Credentials
}

// GetWebAuthnAuthenticator returns the virtual WebAuthn authenticator id and its credentials.
func (rsc *RodSuiteCredentials) GetWebAuthnAuthenticator() (authenticatorID proto.WebAuthnAuthenticatorID, credentials []*proto.WebAuthnCredential) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	return rsc.webauthn.AuthenticatorID, rsc.webauthn.Credentials
}

// SetWebAuthnAuthenticatorID sets the virtual WebAuthn authenticator id.
func (rsc *RodSuiteCredentials) SetWebAuthnAuthenticatorID(authenticatorID proto.WebAuthnAuthenticatorID) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	rsc.webauthn.AuthenticatorID = authenticatorID
}

// SetWebAuthnAuthenticatorCredentials sets the credentials registered against the virtual WebAuthn authenticator.
func (rsc *RodSuiteCredentials) SetWebAuthnAuthenticatorCredentials(credentials ...*proto.WebAuthnCredential) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	rsc.webauthn.Credentials = nil

	for _, credential := range credentials {
		if credential.RpID == "" {
			continue
		}

		rsc.webauthn.Credentials = append(rsc.webauthn.Credentials, credential)
	}
}

// DeleteWebAuthnAuthenticatorCredentials deletes the credentials registered against the virtual WebAuthn authenticator.
func (rsc *RodSuiteCredentials) DeleteWebAuthnAuthenticatorCredentials() {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	rsc.webauthn.Credentials = nil
}

// UpdateWebAuthnAuthenticator atomically updates the virtual WebAuthn authenticator id and its credentials using the given function.
func (rsc *RodSuiteCredentials) UpdateWebAuthnAuthenticator(funcUpdate func(authenticatorID proto.WebAuthnAuthenticatorID, credentials []*proto.WebAuthnCredential) (proto.WebAuthnAuthenticatorID, []*proto.WebAuthnCredential)) {
	if funcUpdate == nil {
		return
	}

	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	credentials := make([]*proto.WebAuthnCredential, len(rsc.webauthn.Credentials))

	if len(credentials) != 0 {
		copy(credentials, rsc.webauthn.Credentials)
	}

	rsc.webauthn.AuthenticatorID, rsc.webauthn.Credentials = funcUpdate(rsc.webauthn.AuthenticatorID, credentials)
}

// DeleteWebAuthnAuthenticator deletes the virtual WebAuthn authenticator and its credentials.
func (rsc *RodSuiteCredentials) DeleteWebAuthnAuthenticator() {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	rsc.webauthn = RodSuiteCredentialWebAuthn{}
}

// RodSuiteCredentialsProvider is the interface implemented by RodSuiteCredentials.
type RodSuiteCredentialsProvider interface {
	GetOneTimePassword(username string) RodSuiteCredentialOneTimePassword
	SetOneTimePassword(username string, credential RodSuiteCredentialOneTimePassword)
	DeleteOneTimePassword(username string)

	GetWebAuthnAuthenticatorID() (authenticatorID proto.WebAuthnAuthenticatorID)
	GetWebAuthnCredentials() (credentials []*proto.WebAuthnCredential)
	SetWebAuthnAuthenticatorID(authenticatorID proto.WebAuthnAuthenticatorID)
	SetWebAuthnAuthenticatorCredentials(credentials ...*proto.WebAuthnCredential)
	DeleteWebAuthnAuthenticatorCredentials()
	GetWebAuthnAuthenticator() (authenticatorID proto.WebAuthnAuthenticatorID, credentials []*proto.WebAuthnCredential)
	UpdateWebAuthnAuthenticator(funcUpdate func(authenticatorID proto.WebAuthnAuthenticatorID, credentials []*proto.WebAuthnCredential) (proto.WebAuthnAuthenticatorID, []*proto.WebAuthnCredential))
	DeleteWebAuthnAuthenticator()
}

// RodSuiteCredentialOneTimePassword represents a registered one time password credential.
type RodSuiteCredentialOneTimePassword struct {
	valid             bool
	Secret            string
	ValidationOptions totp.ValidateOpts
}

// Valid returns true if this credential has been registered.
func (otp *RodSuiteCredentialOneTimePassword) Valid() bool {
	return otp.valid
}

// Generate returns the passcode for this credential at the given time.
func (otp *RodSuiteCredentialOneTimePassword) Generate(at time.Time) (passcode string, err error) {
	return totp.GenerateCodeCustom(otp.Secret, at, otp.ValidationOptions)
}
