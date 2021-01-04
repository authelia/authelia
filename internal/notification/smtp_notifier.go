package notification

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/logging"
	"github.com/authelia/authelia/internal/utils"
)

// SMTPNotifier a notifier to send emails to SMTP servers.
type SMTPNotifier struct {
	username            string
	password            string
	sender              string
	identifier          string
	host                string
	port                int
	disableRequireTLS   bool
	address             string
	subject             string
	startupCheckAddress string
	client              *smtp.Client
	tlsConfig           *tls.Config
}

// NewSMTPNotifier creates a SMTPNotifier using the notifier configuration.
func NewSMTPNotifier(configuration schema.SMTPNotifierConfiguration, certPool *x509.CertPool) *SMTPNotifier {
	notifier := &SMTPNotifier{
		username:            configuration.Username,
		password:            configuration.Password,
		sender:              configuration.Sender,
		identifier:          configuration.Identifier,
		host:                configuration.Host,
		port:                configuration.Port,
		disableRequireTLS:   configuration.DisableRequireTLS,
		address:             fmt.Sprintf("%s:%d", configuration.Host, configuration.Port),
		subject:             configuration.Subject,
		startupCheckAddress: configuration.StartupCheckAddress,
		tlsConfig:           utils.NewTLSConfig(configuration.TLS, tls.VersionTLS12, certPool),
	}

	return notifier
}

// Do startTLS if available (some servers only provide the auth extension after, and encryption is preferred).
func (n *SMTPNotifier) startTLS() error {
	// Only start if not already encrypted
	if _, ok := n.client.TLSConnectionState(); ok {
		logging.Logger().Debugf("Notifier SMTP connection is already encrypted, skipping STARTTLS")
		return nil
	}

	switch ok, _ := n.client.Extension("STARTTLS"); ok {
	case true:
		logging.Logger().Debugf("Notifier SMTP server supports STARTTLS (disableVerifyCert: %t, ServerName: %s), attempting", n.tlsConfig.InsecureSkipVerify, n.tlsConfig.ServerName)

		if err := n.client.StartTLS(n.tlsConfig); err != nil {
			return err
		}

		logging.Logger().Debug("Notifier SMTP STARTTLS completed without error")
	default:
		switch n.disableRequireTLS {
		case true:
			logging.Logger().Warn("Notifier SMTP server does not support STARTTLS and SMTP configuration is set to disable the TLS requirement (only useful for unauthenticated emails over plain text)")
		default:
			return errors.New("Notifier SMTP server does not support TLS and it is required by default (see documentation if you want to disable this highly recommended requirement)")
		}
	}

	return nil
}

// Attempt Authentication.
func (n *SMTPNotifier) auth() error {
	// Attempt AUTH if password is specified only.
	if n.password != "" {
		_, ok := n.client.TLSConnectionState()
		if !ok {
			return errors.New("Notifier SMTP client does not support authentication over plain text and the connection is currently plain text")
		}

		// Check the server supports AUTH, and get the mechanisms.
		ok, m := n.client.Extension("AUTH")
		if ok {
			var auth smtp.Auth

			logging.Logger().Debugf("Notifier SMTP server supports authentication with the following mechanisms: %s", m)
			mechanisms := strings.Split(m, " ")

			// Adaptively select the AUTH mechanism to use based on what the server advertised.
			if utils.IsStringInSlice("PLAIN", mechanisms) {
				auth = smtp.PlainAuth("", n.username, n.password, n.host)

				logging.Logger().Debug("Notifier SMTP client attempting AUTH PLAIN with server")
			} else if utils.IsStringInSlice("LOGIN", mechanisms) {
				auth = newLoginAuth(n.username, n.password, n.host)

				logging.Logger().Debug("Notifier SMTP client attempting AUTH LOGIN with server")
			}

			// Throw error since AUTH extension is not supported.
			if auth == nil {
				return fmt.Errorf("notifier SMTP server does not advertise a AUTH mechanism that are supported by Authelia (PLAIN or LOGIN are supported, but server advertised %s mechanisms)", m)
			}

			// Authenticate.
			if err := n.client.Auth(auth); err != nil {
				return err
			}

			logging.Logger().Debug("Notifier SMTP client authenticated successfully with the server")

			return nil
		}

		return errors.New("Notifier SMTP server does not advertise the AUTH extension but config requires AUTH (password specified), either disable AUTH, or use an SMTP host that supports AUTH PLAIN or AUTH LOGIN")
	}

	logging.Logger().Debug("Notifier SMTP config has no password specified so authentication is being skipped")

	return nil
}

func (n *SMTPNotifier) compose(recipient, subject, body, htmlBody string) error {
	logging.Logger().Debugf("Notifier SMTP client attempting to send email body to %s", recipient)

	if !n.disableRequireTLS {
		_, ok := n.client.TLSConnectionState()
		if !ok {
			return errors.New("Notifier SMTP client can't send an email over plain text connection")
		}
	}

	wc, err := n.client.Data()
	if err != nil {
		logging.Logger().Debugf("Notifier SMTP client error while obtaining WriteCloser: %s", err)
		return err
	}

	boundary := utils.RandomString(30, utils.AlphaNumericCharacters)

	now := time.Now()

	msg := "Date:" + now.Format(rfc5322DateTimeLayout) + "\n" +
		"From: " + n.sender + "\n" +
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
		logging.Logger().Debugf("Notifier SMTP client error while sending email body over WriteCloser: %s", err)
		return err
	}

	err = wc.Close()
	if err != nil {
		logging.Logger().Debugf("Notifier SMTP client error while closing the WriteCloser: %s", err)
		return err
	}

	return nil
}

// Dial the SMTP server with the SMTPNotifier config.
func (n *SMTPNotifier) dial() error {
	logging.Logger().Debugf("Notifier SMTP client attempting connection to %s", n.address)

	if n.port == 465 {
		logging.Logger().Warnf("Notifier SMTP client configured to connect to a SMTPS server. It's highly recommended you use a non SMTPS port and STARTTLS instead of SMTPS, as the protocol is long deprecated.")

		conn, err := tls.Dial("tcp", n.address, n.tlsConfig)
		if err != nil {
			return err
		}

		client, err := smtp.NewClient(conn, n.host)
		if err != nil {
			return err
		}

		n.client = client
	} else {
		client, err := smtp.Dial(n.address)
		if err != nil {
			return err
		}

		n.client = client
	}

	logging.Logger().Debug("Notifier SMTP client connected successfully")

	return nil
}

// Closes the connection properly.
func (n *SMTPNotifier) cleanup() {
	err := n.client.Quit()
	if err != nil {
		logging.Logger().Warnf("Notifier SMTP client encountered error during cleanup: %s", err)
	}
}

// StartupCheck checks the server is functioning correctly and the configuration is correct.
func (n *SMTPNotifier) StartupCheck() (bool, error) {
	if err := n.dial(); err != nil {
		return false, err
	}

	defer n.cleanup()

	if err := n.client.Hello(n.identifier); err != nil {
		return false, err
	}

	if err := n.startTLS(); err != nil {
		return false, err
	}

	if err := n.auth(); err != nil {
		return false, err
	}

	if err := n.client.Mail(n.sender); err != nil {
		return false, err
	}

	if err := n.client.Rcpt(n.startupCheckAddress); err != nil {
		return false, err
	}

	if err := n.client.Reset(); err != nil {
		return false, err
	}

	return true, nil
}

// Send is used to send an email to a recipient.
func (n *SMTPNotifier) Send(recipient, title, body, htmlBody string) error {
	subject := strings.ReplaceAll(n.subject, "{title}", title)

	if err := n.dial(); err != nil {
		return err
	}

	// Always execute QUIT at the end once we're connected.
	defer n.cleanup()

	if err := n.client.Hello(n.identifier); err != nil {
		return err
	}

	// Start TLS and then Authenticate.
	if err := n.startTLS(); err != nil {
		return err
	}

	if err := n.auth(); err != nil {
		return err
	}

	// Set the sender and recipient first.
	if err := n.client.Mail(n.sender); err != nil {
		logging.Logger().Debugf("Notifier SMTP failed while sending MAIL FROM (using sender) with error: %s", err)
		return err
	}

	if err := n.client.Rcpt(recipient); err != nil {
		logging.Logger().Debugf("Notifier SMTP failed while sending RCPT TO (using recipient) with error: %s", err)
		return err
	}

	// Compose and send the email body to the server.
	if err := n.compose(recipient, subject, body, htmlBody); err != nil {
		return err
	}

	logging.Logger().Debug("Notifier SMTP client successfully sent email")

	return nil
}
