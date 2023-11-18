package suites

import (
	"sync"
	"time"

	"github.com/pquerna/otp/totp"
)

func NewRodSuiteCredentials() *RodSuiteCredentials {
	return &RodSuiteCredentials{
		lock: &sync.Mutex{},
		totp: map[string]RodSuiteCredentialOneTimePassword{},
	}
}

type RodSuiteCredentials struct {
	lock *sync.Mutex
	totp map[string]RodSuiteCredentialOneTimePassword
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

type RodSuiteCredentialsProvider interface {
	GetOneTimePassword(username string) RodSuiteCredentialOneTimePassword
	SetOneTimePassword(username string, credential RodSuiteCredentialOneTimePassword)
	DeleteOneTimePassword(username string)
}

type RodSuiteCredentialOneTimePassword struct {
	valid             bool
	Secret            string
	ValidationOptions totp.ValidateOpts
}

func (otp *RodSuiteCredentialOneTimePassword) Valid() bool {
	return otp.valid
}

func (otp *RodSuiteCredentialOneTimePassword) Generate(at time.Time) (passcode string, err error) {
	return totp.GenerateCodeCustom(otp.Secret, at, otp.ValidationOptions)
}
