package suites

import (
	"sync"
	"time"

	"github.com/authelia/otp/totp"
	"github.com/go-rod/rod/lib/proto"
)

func NewRodSuiteCredentials() *RodSuiteCredentials {
	return &RodSuiteCredentials{
		lock: &sync.Mutex{},
		totp: map[string]RodSuiteCredentialOneTimePassword{},
	}
}

type RodSuiteCredentials struct {
	lock     *sync.Mutex
	totp     map[string]RodSuiteCredentialOneTimePassword
	webauthn RodSuiteCredentialWebAuthn
}

type RodSuiteCredentialWebAuthn struct {
	AuthenticatorID proto.WebAuthnAuthenticatorID
	Credentials     []*proto.WebAuthnCredential
}

func (rsc *RodSuiteCredentials) GetOneTimePassword(username string) RodSuiteCredentialOneTimePassword {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	return rsc.totp[username]
}

func (rsc *RodSuiteCredentials) SetOneTimePassword(username string, credential RodSuiteCredentialOneTimePassword) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	credential.valid = true

	rsc.totp[username] = credential
}

func (rsc *RodSuiteCredentials) DeleteOneTimePassword(username string) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	rsc.totp[username] = RodSuiteCredentialOneTimePassword{
		valid: false,
	}
}

func (rsc *RodSuiteCredentials) GetWebAuthnAuthenticatorID() (authenticatorID proto.WebAuthnAuthenticatorID) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	return rsc.webauthn.AuthenticatorID
}

func (rsc *RodSuiteCredentials) GetWebAuthnCredentials() (credentials []*proto.WebAuthnCredential) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	return rsc.webauthn.Credentials
}

func (rsc *RodSuiteCredentials) GetWebAuthnAuthenticator() (authenticatorID proto.WebAuthnAuthenticatorID, credentials []*proto.WebAuthnCredential) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	return rsc.webauthn.AuthenticatorID, rsc.webauthn.Credentials
}

func (rsc *RodSuiteCredentials) SetWebAuthnAuthenticatorID(authenticatorID proto.WebAuthnAuthenticatorID) {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	rsc.webauthn.AuthenticatorID = authenticatorID
}

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

func (rsc *RodSuiteCredentials) DeleteWebAuthnAuthenticatorCredentials() {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	rsc.webauthn.Credentials = nil
}

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

func (rsc *RodSuiteCredentials) DeleteWebAuthnAuthenticator() {
	rsc.lock.Lock()

	defer rsc.lock.Unlock()

	rsc.webauthn = RodSuiteCredentialWebAuthn{}
}

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

type RodSuiteCredentialOneTimePassword struct {
	valid             bool
	Secret            string //nolint:gosec
	ValidationOptions totp.ValidateOpts
}

func (otp *RodSuiteCredentialOneTimePassword) Valid() bool {
	return otp.valid
}

func (otp *RodSuiteCredentialOneTimePassword) Generate(at time.Time) (passcode string, err error) {
	return totp.GenerateCodeCustom(otp.Secret, at, otp.ValidationOptions)
}
