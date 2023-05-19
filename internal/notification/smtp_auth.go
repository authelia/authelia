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
func NewOpportunisticSMTPAuth(config *schema.SMTPNotifierConfiguration) *OpportunisticSMTPAuth {
	if config.Username == "" && config.Password == "" {
		return nil
	}

	return &OpportunisticSMTPAuth{
		username: config.Username,
		password: config.Password,
		host:     config.Address.Hostname(),
	}
}

// OpportunisticSMTPAuth is an opportunistic smtp.Auth implementation.
type OpportunisticSMTPAuth struct {
	username, password, host string

	satPreference []mail.SMTPAuthType
	sa            smtp.Auth
}

// Start begins an authentication with a server.
// It returns the name of the authentication protocol
// and optionally data to include in the initial AUTH message
// sent to the server.
// If it returns a non-nil error, the SMTP client aborts
// the authentication attempt and closes the connection.
func (a *OpportunisticSMTPAuth) Start(server *smtp.ServerInfo) (proto string, toServer []byte, err error) {
	for _, pref := range a.satPreference {
		if utils.IsStringInSlice(string(pref), server.Auth) {
			switch pref {
			case mail.SMTPAuthPlain:
				a.sa = smtp.PlainAuth("", a.username, a.password, a.host)
			case mail.SMTPAuthLogin:
				a.sa = smtp.LoginAuth(a.username, a.password, a.host)
			case mail.SMTPAuthCramMD5:
				a.sa = smtp.CRAMMD5Auth(a.username, a.password)
			}

			break
		}
	}

	if a.sa == nil {
		for _, sa := range server.Auth {
			switch mail.SMTPAuthType(sa) {
			case mail.SMTPAuthPlain:
				a.sa = smtp.PlainAuth("", a.username, a.password, a.host)
			case mail.SMTPAuthLogin:
				a.sa = smtp.LoginAuth(a.username, a.password, a.host)
			case mail.SMTPAuthCramMD5:
				a.sa = smtp.CRAMMD5Auth(a.username, a.password)
			}
		}
	}

	if a.sa == nil {
		return "", nil, fmt.Errorf("unsupported SMTP AUTH types: %s", strings.Join(server.Auth, ", "))
	}

	return a.sa.Start(server)
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
