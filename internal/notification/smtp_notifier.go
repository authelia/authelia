package notification

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/utils"
)

// SMTPNotifier a notifier to send emails to SMTP servers.
type SMTPNotifier struct {
	configuration *schema.SMTPNotifierConfiguration
	client        *smtp.Client
	tlsConfig     *tls.Config
	log           *logrus.Logger
}

// NewSMTPNotifier creates a SMTPNotifier using the notifier configuration.
func NewSMTPNotifier(configuration *schema.SMTPNotifierConfiguration, certPool *x509.CertPool) *SMTPNotifier {
	notifier := &SMTPNotifier{
		configuration: configuration,
		tlsConfig:     utils.NewTLSConfig(configuration.TLS, tls.VersionTLS12, certPool),
		log:           logging.Logger(),
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
		switch n.configuration.DisableRequireTLS {
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
	if n.configuration.Password != "" {
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
				auth = smtp.PlainAuth("", n.configuration.Username, n.configuration.Password, n.configuration.Host)

				n.log.Debug("Notifier SMTP client attempting AUTH PLAIN with server")
			} else if utils.IsStringInSlice("LOGIN", mechanisms) {
				auth = newLoginAuth(n.configuration.Username, n.configuration.Password, n.configuration.Host)

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

func (n *SMTPNotifier) compose(recipient, subject, body, htmlBody string) error {
	n.log.Debugf("Notifier SMTP client attempting to send email body to %s", recipient)

	if !n.configuration.DisableRequireTLS {
		_, ok := n.client.TLSConnectionState()
		if !ok {
			return errors.New("Notifier SMTP client can't send an email over plain text connection")
		}
	}

	wc, err := n.client.Data()
	if err != nil {
		n.log.Debugf("Notifier SMTP client error while obtaining WriteCloser: %s", err)
		return err
	}

	boundary := utils.RandomString(30, utils.AlphaNumericCharacters, true)

	now := time.Now()

	msg := "Date:" + now.Format(rfc5322DateTimeLayout) + "\n" +
		"From: " + n.configuration.Sender.String() + "\n" +
		"To: " + recipient + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-version: 1.0\n" +
		"Content-Type: multipart/alternative; boundary=" + boundary + "\n\n" +
		"--" + boundary + "\n" +
		"Content-Type: text/plain; charset=\"UTF-8\"\n" +
		"Content-Transfer-Encoding: quoted-printable\n" +
		"Content-Disposition: inline\n\n" +
		body + "\n"

	if htmlBody != "" {
		msg += "--" + boundary + "\n" +
			"Content-Type: text/html; charset=\"UTF-8\"\n\n" +
			htmlBody + "\n"
	}

	msg += "--" + boundary + "--"

	_, err = fmt.Fprint(wc, msg)
	if err != nil {
		n.log.Debugf("Notifier SMTP client error while sending email body over WriteCloser: %s", err)
		return err
	}

	err = wc.Close()
	if err != nil {
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
		dialer = &net.Dialer{Timeout: n.configuration.Timeout}
	)

	n.log.Debugf("Notifier SMTP client attempting connection to %s:%d", n.configuration.Host, n.configuration.Port)

	if n.configuration.Port == 465 {
		n.log.Infof("Notifier SMTP client using submissions port 465. Make sure the mail server you are connecting to is configured for submissions and not SMTPS.")

		conn, err = tls.DialWithDialer(dialer, "tcp", fmt.Sprintf("%s:%d", n.configuration.Host, n.configuration.Port), n.tlsConfig)
	} else {
		conn, err = dialer.Dial("tcp", fmt.Sprintf("%s:%d", n.configuration.Host, n.configuration.Port))
	}

	switch {
	case err == nil:
		break
	case errors.Is(err, io.EOF):
		return fmt.Errorf("received %w error: this error often occurs due to network errors such as a firewall, network policies, or closed ports which may be due to smtp service not running or an incorrect port specified in configuration", err)
	default:
		return err
	}

	client, err = smtp.NewClient(conn, n.configuration.Host)
	if err != nil {
		return err
	}

	n.client = client

	n.log.Debug("Notifier SMTP client connected successfully")

	return nil
}

// Closes the connection properly.
func (n *SMTPNotifier) cleanup() {
	err := n.client.Quit()
	if err != nil {
		n.log.Warnf("Notifier SMTP client encountered error during cleanup: %s", err)
	}
}

// StartupCheck implements the startup check provider interface.
func (n *SMTPNotifier) StartupCheck() (err error) {
	if err = n.dial(); err != nil {
		return fmt.Errorf("error dialing the smtp server: %w", err)
	}

	defer n.cleanup()

	if err = n.client.Hello(n.configuration.Identifier); err != nil {
		return fmt.Errorf("error performing HELO/EHLO with the smtp server: %w", err)
	}

	if err = n.startTLS(); err != nil {
		return fmt.Errorf("error performing STARTTLS with the smtp server: %w", err)
	}

	if err = n.auth(); err != nil {
		return fmt.Errorf("error performing AUTH with the smtp server: %w", err)
	}

	if err = n.client.Mail(n.configuration.Sender.Address); err != nil {
		return fmt.Errorf("error performing MAIL FROM with the smtp server: %w", err)
	}

	if err = n.client.Rcpt(n.configuration.StartupCheckAddress); err != nil {
		return fmt.Errorf("error performing RCPT with the smtp server: %w", err)
	}

	return n.client.Reset()
}

// Send is used to send an email to a recipient.
func (n *SMTPNotifier) Send(recipient, title, body, htmlBody string) error {
	subject := strings.ReplaceAll(n.configuration.Subject, "{title}", title)

	var err error

	if err = n.dial(); err != nil {
		return fmt.Errorf("error dialing the smtp server: %w", err)
	}

	// Always execute QUIT at the end once we're connected.
	defer n.cleanup()

	if err = n.client.Hello(n.configuration.Identifier); err != nil {
		return fmt.Errorf("error performing HELO/EHLO with the smtp server: %w", err)
	}

	if err = n.startTLS(); err != nil {
		return fmt.Errorf("error performing STARTTLS with the smtp server: %w", err)
	}

	if err = n.auth(); err != nil {
		return fmt.Errorf("error performing AUTH with the smtp server: %w", err)
	}

	if err = n.client.Mail(n.configuration.Sender.Address); err != nil {
		n.log.Debugf("Notifier SMTP failed while sending MAIL FROM (using sender) with error: %s", err)

		return fmt.Errorf("error performing MAIL FROM with the smtp server: %w", err)
	}

	if err = n.client.Rcpt(n.configuration.StartupCheckAddress); err != nil {
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
