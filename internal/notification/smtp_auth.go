package notification

import (
	"fmt"
	"net/smtp"
	"strings"

	gomail "github.com/wneessen/go-mail"
	"github.com/wneessen/go-mail/auth"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewOpportunisticSMTPAuth is an opportunistic smtp.Auth implementation.
func NewOpportunisticSMTPAuth(config *schema.SMTPNotifierConfiguration) *OpportunisticSMTPAuth {
	return &OpportunisticSMTPAuth{
		username: config.Username,
		password: config.Password,
		host:     config.Host,
	}
}

// OpportunisticSMTPAuth is an opportunistic smtp.Auth implementation.
type OpportunisticSMTPAuth struct {
	username, password, host string

	satPreference []gomail.SMTPAuthType
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
			case gomail.SMTPAuthPlain:
				a.sa = smtp.PlainAuth("", a.username, a.password, a.host)
			case gomail.SMTPAuthLogin:
				a.sa = auth.LoginAuth(a.username, a.password, a.host)
			case gomail.SMTPAuthCramMD5:
				a.sa = smtp.CRAMMD5Auth(a.username, a.password)
			}

			break
		}
	}

	if a.sa == nil {
		for _, sa := range server.Auth {
			switch gomail.SMTPAuthType(sa) {
			case gomail.SMTPAuthPlain:
				a.sa = smtp.PlainAuth("", a.username, a.password, a.host)
			case gomail.SMTPAuthLogin:
				a.sa = auth.LoginAuth(a.username, a.password, a.host)
			case gomail.SMTPAuthCramMD5:
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
