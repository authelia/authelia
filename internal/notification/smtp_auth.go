package notification

import (
	"fmt"
	"strings"

	"github.com/wneessen/go-mail"
	"github.com/wneessen/go-mail/smtp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewOpportunisticSMTPAuth is an opportunistic smtp.Auth implementation.
func NewOpportunisticSMTPAuth(config *schema.NotifierSMTP, preference ...mail.SMTPAuthType) smtp.Auth {
	if len(config.Username)+len(config.Password) == 0 {
		return nil
	}

	return &OpportunisticSMTPAuth{
		username:          config.Username,
		password:          config.Password,
		host:              config.Address.Hostname(),
		satPreference:     preference,
		disableRequireTLS: config.DisableRequireTLS,
	}
}

// OpportunisticSMTPAuth is an opportunistic smtp.Auth implementation.
type OpportunisticSMTPAuth struct {
	username, password, host string

	satPreference     []mail.SMTPAuthType
	sa                smtp.Auth
	disableRequireTLS bool
}

// Start begins an authentication with a server.
// It returns the name of the authentication protocol
// and optionally data to include in the initial AUTH message
// sent to the server.
// If it returns a non-nil error, the SMTP client aborts
// the authentication attempt and closes the connection.
func (a *OpportunisticSMTPAuth) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	a.setPreferred(server)

	if a.sa == nil {
		a.set(server)
	}

	if a.sa == nil {
		return "", nil, fmt.Errorf("unsupported SMTP AUTH types: %s", strings.Join(server.Auth, ", "))
	}

	return a.sa.Start(server)
}

func (a *OpportunisticSMTPAuth) setPreferred(server *smtp.ServerInfo) {
	for _, pref := range a.satPreference {
		if utils.IsStringInSlice(string(pref), server.Auth) {
			switch pref {
			case mail.SMTPAuthSCRAMSHA256:
				a.sa = smtp.ScramSHA256Auth(a.username, a.password)
			case mail.SMTPAuthSCRAMSHA1:
				a.sa = smtp.ScramSHA1Auth(a.username, a.password)
			case mail.SMTPAuthCramMD5:
				a.sa = smtp.CRAMMD5Auth(a.username, a.password)
			case mail.SMTPAuthPlain:
				a.sa = smtp.PlainAuth("", a.username, a.password, a.host, a.disableRequireTLS)
			case mail.SMTPAuthLogin:
				a.sa = smtp.LoginAuth(a.username, a.password, a.host, a.disableRequireTLS)
			}

			if a.sa != nil {
				break
			}
		}
	}
}

func (a *OpportunisticSMTPAuth) set(server *smtp.ServerInfo) {
	switch {
	case utils.IsStringInSlice(string(mail.SMTPAuthSCRAMSHA256), server.Auth):
		a.sa = smtp.ScramSHA256Auth(a.username, a.password)
	case utils.IsStringInSlice(string(mail.SMTPAuthSCRAMSHA1), server.Auth):
		a.sa = smtp.ScramSHA1Auth(a.username, a.password)
	case utils.IsStringInSlice(string(mail.SMTPAuthCramMD5), server.Auth):
		a.sa = smtp.CRAMMD5Auth(a.username, a.password)
	case utils.IsStringInSlice(string(mail.SMTPAuthPlain), server.Auth):
		a.sa = smtp.PlainAuth("", a.username, a.password, a.host, a.disableRequireTLS)
	case utils.IsStringInSlice(string(mail.SMTPAuthLogin), server.Auth):
		a.sa = smtp.LoginAuth(a.username, a.password, a.host, a.disableRequireTLS)
	}
}

// Next continues the authentication. The server has just sent
// the fromServer data. If more is true, the server expects a
// response, which Next should return as toServer; otherwise
// Next should return toServer == nil.
// If Next returns a non-nil error, the SMTP client aborts
// the authentication attempt and closes the connection.
func (a *OpportunisticSMTPAuth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	return a.sa.Next(fromServer, more)
}
