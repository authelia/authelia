package notification

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/mail"
	"net/smtp"
	"strings"

	"github.com/sirupsen/logrus"
	gomail "github.com/wneessen/go-mail"
	"github.com/wneessen/go-mail/auth"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewSMTPNotifier creates a SMTPNotifier using the notifier configuration.
func NewSMTPNotifier(config *schema.SMTPNotifierConfiguration, certPool *x509.CertPool) *SMTPNotifier {
	opts := []gomail.Option{
		gomail.WithPort(config.Port),
		gomail.WithTLSConfig(utils.NewTLSConfig(config.TLS, certPool)),
		gomail.WithHELO(config.Identifier),
	}

	switch {
	case config.DisableStartTLS:
		opts = append(opts, gomail.WithTLSPolicy(gomail.NoTLS))
	case config.DisableRequireTLS:
		opts = append(opts, gomail.WithTLSPolicy(gomail.TLSOpportunistic))
	default:
		opts = append(opts, gomail.WithTLSPolicy(gomail.TLSMandatory))
	}

	if config.Port == smtpPortSUBMISSIONS {
		opts = append(opts, gomail.WithSSL())
	}

	var domain string

	at := strings.LastIndex(config.Sender.Address, "@")

	if at >= 0 {
		domain = config.Sender.Address[at:]
	}

	return &SMTPNotifier{
		config: config,
		domain: domain,
		tls:    utils.NewTLSConfig(config.TLS, certPool),
		log:    logging.Logger(),
		opts:   opts,
	}
}

// SMTPNotifier a notifier to send emails to SMTP servers.
type SMTPNotifier struct {
	config *schema.SMTPNotifierConfiguration
	domain string
	tls    *tls.Config
	log    *logrus.Logger
	opts   []gomail.Option
}

// StartupCheck implements model.StartupCheck to perform startup check operations.
func (n *SMTPNotifier) StartupCheck() (err error) {
	var client *gomail.Client

	if client, err = gomail.NewClient(n.config.Host, n.opts...); err != nil {
		return fmt.Errorf("failed to establish client: %w", err)
	}

	ctx := context.Background()

	if err = client.DialWithContext(ctx); err != nil {
		return fmt.Errorf("failed to dial connection: %w", err)
	}

	if err = client.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}

	return nil
}

// Send a notification via the SMTPNotifier.
func (n *SMTPNotifier) Send(ctx context.Context, recipient mail.Address, subject string, et *templates.EmailTemplate, data any) (err error) {
	msg := gomail.NewMsg(
		gomail.WithMIMEVersion(gomail.Mime10),
		gomail.WithBoundary(utils.RandomString(30, utils.CharSetAlphaNumeric, true)),
	)

	if err = msg.From(n.config.Sender.String()); err != nil {
		return fmt.Errorf("notifier: smtp: failed to set from address: %w", err)
	}

	if err = msg.AddTo(recipient.String()); err != nil {
		return fmt.Errorf("notifier: smtp: failed to set to address: %w", err)
	}

	msg.Subject(strings.ReplaceAll(n.config.Subject, "{title}", subject))

	switch {
	case n.config.DisableHTMLEmails:
		if err = msg.SetBodyTextTemplate(et.Text, data); err != nil {
			return fmt.Errorf("notifier: smtp: failed to set body: text template errored: %w", err)
		}
	default:
		if err = msg.AddAlternativeHTMLTemplate(et.HTML, data); err != nil {
			return fmt.Errorf("notifier: smtp: failed to set body: html template errored: %w", err)
		}

		if err = msg.AddAlternativeTextTemplate(et.Text, data); err != nil {
			return fmt.Errorf("notifier: smtp: failed to set body: text template errored: %w", err)
		}
	}

	var client *gomail.Client

	n.log.Debugf("creating client with %d options: %+v", len(n.opts), n.opts)

	if client, err = gomail.NewClient(n.config.Host, n.opts...); err != nil {
		return fmt.Errorf("notifier: smtp: failed to establish client: %w", err)
	}

	client.SetSMTPAuthCustom(NewOpportunisticSMTPAuth(n.config))

	if err = client.DialWithContext(ctx); err != nil {
		return fmt.Errorf("notifier: smtp: failed to dial connection: %w", err)
	}

	if err = client.Send(msg); err != nil {
		return fmt.Errorf("notifier: smtp: failed to send message: %w", err)
	}

	if err = client.Close(); err != nil {
		return fmt.Errorf("notifier: smtp: failed to close connection: %w", err)
	}

	return nil
}

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
