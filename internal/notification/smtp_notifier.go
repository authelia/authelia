package notification

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/templates"
	"github.com/authelia/authelia/v4/internal/utils"
)

// SMTPNotifier a notifier to send emails to SMTP servers.
type SMTPNotifier struct {
	config    *schema.SMTPNotifierConfiguration
	client    *smtp.Client
	tlsConfig *tls.Config
	log       *logrus.Logger
	templates *templates.Provider
}

// NewSMTPNotifier creates a SMTPNotifier using the notifier configuration.
func NewSMTPNotifier(configuration *schema.SMTPNotifierConfiguration, certPool *x509.CertPool, templateProvider *templates.Provider) *SMTPNotifier {
	notifier := &SMTPNotifier{
		config:    configuration,
		tlsConfig: utils.NewTLSConfig(configuration.TLS, tls.VersionTLS12, certPool),
		log:       logging.Logger(),
		templates: templateProvider,
	}

	return notifier
}

// Do startTLS if available (some servers only provide the auth extension after, and encryption is preferred).
func (n *SMTPNotifier) startTLS() error {
	// Only start if not already encrypted.
	if _, ok := n.client.TLSConnectionState(); ok {
		n.log.Debugf("Notifier SMTP connection is already encrypted, skipping STARTTLS")
		return nil
	}

	switch ok, _ := n.client.Extension("STARTTLS"); ok {
	case true:
		n.log.Debugf("Notifier SMTP server supports STARTTLS (disableVerifyCert: %t, ServerName: %s), attempting", n.tlsConfig.InsecureSkipVerify, n.tlsConfig.ServerName)

		if err := n.client.StartTLS(n.tlsConfig); err != nil {
			return err
		}

		n.log.Debug("Notifier SMTP STARTTLS completed without error")
	default:
		switch n.config.DisableRequireTLS {
		case true:
			n.log.Warn("Notifier SMTP server does not support STARTTLS and SMTP configuration is set to disable the TLS requirement (only useful for unauthenticated emails over plain text)")
		default:
			return errors.New("Notifier SMTP server does not support TLS and it is required by default (see documentation if you want to disable this highly recommended requirement)")
		}
	}

	return nil
}

// Attempt Authentication.
func (n *SMTPNotifier) auth() error {
	// Attempt AUTH if password is specified only.
	if n.config.Password != "" {
		_, ok := n.client.TLSConnectionState()
		if !ok {
			return errors.New("Notifier SMTP client does not support authentication over plain text and the connection is currently plain text")
		}

		// Check the server supports AUTH, and get the mechanisms.
		ok, m := n.client.Extension("AUTH")
		if ok {
			var auth smtp.Auth

			n.log.Debugf("Notifier SMTP server supports authentication with the following mechanisms: %s", m)
			mechanisms := strings.Split(m, " ")

			// Adaptively select the AUTH mechanism to use based on what the server advertised.
			if utils.IsStringInSlice("PLAIN", mechanisms) {
				auth = smtp.PlainAuth("", n.config.Username, n.config.Password, n.config.Host)

				n.log.Debug("Notifier SMTP client attempting AUTH PLAIN with server")
			} else if utils.IsStringInSlice("LOGIN", mechanisms) {
				auth = newLoginAuth(n.config.Username, n.config.Password, n.config.Host)

				n.log.Debug("Notifier SMTP client attempting AUTH LOGIN with server")
			}

			// Throw error since AUTH extension is not supported.
			if auth == nil {
				return fmt.Errorf("notifier SMTP server does not advertise a AUTH mechanism that are supported by Authelia (PLAIN or LOGIN are supported, but server advertised %s mechanisms)", m)
			}

			// Authenticate.
			if err := n.client.Auth(auth); err != nil {
				return err
			}

			n.log.Debug("Notifier SMTP client authenticated successfully with the server")

			return nil
		}

		return errors.New("Notifier SMTP server does not advertise the AUTH extension but config requires AUTH (password specified), either disable AUTH, or use an SMTP host that supports AUTH PLAIN or AUTH LOGIN")
	}

	n.log.Debug("Notifier SMTP config has no password specified so authentication is being skipped")

	return nil
}

func (n *SMTPNotifier) compose(recipient mail.Address, subject, body, htmlBody string) (err error) {
	n.log.Debugf("Notifier SMTP client attempting to send email body to %s", recipient.String())

	if !n.config.DisableRequireTLS {
		_, ok := n.client.TLSConnectionState()
		if !ok {
			return errors.New("Notifier SMTP client can't send an email over plain text connection")
		}
	}

	var wc io.WriteCloser

	if wc, err = n.client.Data(); err != nil {
		n.log.Debugf("Notifier SMTP client error while obtaining WriteCloser: %v", err)
		return err
	}

	muuid, err := uuid.NewRandom()
	if err != nil {
		return err
	}

	values := templates.EmailEnvelopeValues{
		UUID:     muuid.String(),
		From:     n.config.Sender.String(),
		To:       recipient.String(),
		Subject:  subject,
		Date:     time.Now(),
		Boundary: utils.RandomString(30, utils.AlphaNumericCharacters, true),
		Body: templates.EmailEnvelopeBodyValues{
			PlainText: body,
			HTML:      htmlBody,
		},
	}

	if err = n.templates.ExecuteEmailEnvelope(wc, values); err != nil {
		n.log.Debugf("Notifier SMTP client error while sending email body over WriteCloser: %s", err)

		return err
	}

	if err = wc.Close(); err != nil {
		n.log.Debugf("Notifier SMTP client error while closing the WriteCloser: %s", err)
		return err
	}

	return nil
}

// Dial the SMTP server with the SMTPNotifier config.
func (n *SMTPNotifier) dial() (err error) {
	var (
		client *smtp.Client
		conn   net.Conn
		dialer = &net.Dialer{Timeout: n.config.Timeout}
	)

	n.log.Debugf("Notifier SMTP client attempting connection to %s:%d", n.config.Host, n.config.Port)

	if n.config.Port == 465 {
		n.log.Infof("Notifier SMTP client using submissions port 465. Make sure the mail server you are connecting to is configured for submissions and not SMTPS.")

		conn, err = tls.DialWithDialer(dialer, "tcp", fmt.Sprintf("%s:%d", n.config.Host, n.config.Port), n.tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", fmt.Sprintf("%s:%d", n.config.Host, n.config.Port))
	}

	switch {
	case err == nil:
		break
	case errors.Is(err, io.EOF):
		return fmt.Errorf("received %w error: this error often occurs due to network errors such as a firewall, network policies, or closed ports which may be due to smtp service not running or an incorrect port specified in configuration", err)
	default:
		return err
	}

	client, err = smtp.NewClient(conn, n.config.Host)
	if err != nil {
		return err
	}

	n.client = client

	n.log.Debug("Notifier SMTP client connected successfully")

	return nil
}

// Closes the connection properly.
func (n *SMTPNotifier) cleanup() {
	if err := n.client.Quit(); err != nil {
		n.log.Warnf("Notifier SMTP client encountered error during cleanup: %v", err)
	}
}

// StartupCheck implements the startup check provider interface.
func (n *SMTPNotifier) StartupCheck() (err error) {
	if err = n.dial(); err != nil {
		return fmt.Errorf("error dialing the smtp server: %w", err)
	}

	defer n.cleanup()

	if err = n.client.Hello(n.config.Identifier); err != nil {
		return fmt.Errorf("error performing HELO/EHLO with the smtp server: %w", err)
	}

	if err = n.startTLS(); err != nil {
		return fmt.Errorf("error performing STARTTLS with the smtp server: %w", err)
	}

	if err = n.auth(); err != nil {
		return fmt.Errorf("error performing AUTH with the smtp server: %w", err)
	}

	if err = n.client.Mail(n.config.Sender.Address); err != nil {
		return fmt.Errorf("error performing MAIL FROM with the smtp server: %w", err)
	}

	if err = n.client.Rcpt(n.config.StartupCheckAddress); err != nil {
		return fmt.Errorf("error performing RCPT with the smtp server: %w", err)
	}

	return n.client.Reset()
}

// Send is used to send an email to a recipient.
func (n *SMTPNotifier) Send(recipient mail.Address, title, body, htmlBody string) error {
	subject := strings.ReplaceAll(n.config.Subject, "{title}", title)

	var err error

	if err = n.dial(); err != nil {
		return fmt.Errorf("error dialing the smtp server: %w", err)
	}

	// Always execute QUIT at the end once we're connected.
	defer n.cleanup()

	if err = n.client.Hello(n.config.Identifier); err != nil {
		return fmt.Errorf("error performing HELO/EHLO with the smtp server: %w", err)
	}

	if err = n.startTLS(); err != nil {
		return fmt.Errorf("error performing STARTTLS with the smtp server: %w", err)
	}

	if err = n.auth(); err != nil {
		return fmt.Errorf("error performing AUTH with the smtp server: %w", err)
	}

	if err = n.client.Mail(n.config.Sender.Address); err != nil {
		n.log.Debugf("Notifier SMTP failed while sending MAIL FROM (using sender) with error: %s", err)

		return fmt.Errorf("error performing MAIL FROM with the smtp server: %w", err)
	}

	if err = n.client.Rcpt(n.config.StartupCheckAddress); err != nil {
		n.log.Debugf("Notifier SMTP failed while sending RCPT TO (using recipient) with error: %s", err)

		return fmt.Errorf("error performing RCPT with the smtp server: %w", err)
	}

	// Compose and send the email body to the server.
	if err = n.compose(recipient, subject, body, htmlBody); err != nil {
		return err
	}

	n.log.Debug("Notifier SMTP client successfully sent email")

	return nil
}
