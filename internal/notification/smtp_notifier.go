package notification

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// SMTPNotifier a notifier to send emails to SMTP servers.
type SMTPNotifier struct {
	username            string
	password            string
	sender              string
	host                string
	port                int
	trustedCert         string
	disableVerifyCert   bool
	disableRequireTLS   bool
	address             string
	subject             string
	startupCheckAddress string
	client              *smtp.Client
	tlsConfig           *tls.Config
}

// NewSMTPNotifier creates a SMTPNotifier using the notifier configuration.
func NewSMTPNotifier(configuration schema.SMTPNotifierConfiguration) *SMTPNotifier {
	notifier := &SMTPNotifier{
		username:            configuration.Username,
		password:            configuration.Password,
		sender:              configuration.Sender,
		host:                configuration.Host,
		port:                configuration.Port,
		trustedCert:         configuration.TrustedCert,
		disableVerifyCert:   configuration.DisableVerifyCert,
		disableRequireTLS:   configuration.DisableRequireTLS,
		address:             fmt.Sprintf("%s:%d", configuration.Host, configuration.Port),
		subject:             configuration.Subject,
		startupCheckAddress: configuration.StartupCheckAddress,
	}
	notifier.initializeTLSConfig()

	return notifier
}

func (n *SMTPNotifier) initializeTLSConfig() {
	// Do not allow users to disable verification of certs if they have also set a trusted cert that was loaded
	// The second part of this check happens in the Configure Cert Pool code block
	log.Debug("Notifier SMTP client initializing TLS configuration")

	//Configure Cert Pool
	certPool, err := x509.SystemCertPool()
	if err != nil || certPool == nil {
		certPool = x509.NewCertPool()
	}

	if n.trustedCert != "" {
		log.Debugf("Notifier SMTP client attempting to load certificate from %s", n.trustedCert)

		if exists, err := utils.FileExists(n.trustedCert); exists {
			pem, err := ioutil.ReadFile(n.trustedCert)
			if err != nil {
				log.Warnf("Notifier SMTP failed to load cert from file with error: %s", err)
			} else {
				if ok := certPool.AppendCertsFromPEM(pem); !ok {
					log.Warn("Notifier SMTP failed to import cert loaded from file")
				} else {
					log.Debug("Notifier SMTP successfully loaded certificate")
					if n.disableVerifyCert {
						log.Warn("Notifier SMTP when trusted_cert is specified we force disable_verify_cert to false, if you want to disable certificate validation please comment/delete trusted_cert from your config")
						n.disableVerifyCert = false
					}
				}
			}
		} else {
			log.Warnf("Notifier SMTP failed to load cert from file (file does not exist) with error: %s", err)
		}
	}

	n.tlsConfig = &tls.Config{
		InsecureSkipVerify: n.disableVerifyCert, //nolint:gosec // This is an intended config, we never default true, provide alternate options, and we constantly warn the user.
		ServerName:         n.host,
		RootCAs:            certPool,
	}
}

// Do startTLS if available (some servers only provide the auth extension after, and encryption is preferred).
func (n *SMTPNotifier) startTLS() error {
	// Only start if not already encrypted
	if _, ok := n.client.TLSConnectionState(); ok {
		log.Debugf("Notifier SMTP connection is already encrypted, skipping STARTTLS")
		return nil
	}

	switch ok, _ := n.client.Extension("STARTTLS"); ok {
	case true:
		log.Debugf("Notifier SMTP server supports STARTTLS (disableVerifyCert: %t, ServerName: %s), attempting", n.tlsConfig.InsecureSkipVerify, n.tlsConfig.ServerName)

		if err := n.client.StartTLS(n.tlsConfig); err != nil {
			return err
		}

		log.Debug("Notifier SMTP STARTTLS completed without error")
	default:
		switch n.disableRequireTLS {
		case true:
			log.Warn("Notifier SMTP server does not support STARTTLS and SMTP configuration is set to disable the TLS requirement (only useful for unauthenticated emails over plain text)")
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

			log.Debugf("Notifier SMTP server supports authentication with the following mechanisms: %s", m)
			mechanisms := strings.Split(m, " ")

			// Adaptively select the AUTH mechanism to use based on what the server advertised.
			if utils.IsStringInSlice("PLAIN", mechanisms) {
				auth = smtp.PlainAuth("", n.username, n.password, n.host)

				log.Debug("Notifier SMTP client attempting AUTH PLAIN with server")
			} else if utils.IsStringInSlice("LOGIN", mechanisms) {
				auth = newLoginAuth(n.username, n.password, n.host)

				log.Debug("Notifier SMTP client attempting AUTH LOGIN with server")
			}

			// Throw error since AUTH extension is not supported.
			if auth == nil {
				return fmt.Errorf("notifier SMTP server does not advertise a AUTH mechanism that are supported by Authelia (PLAIN or LOGIN are supported, but server advertised %s mechanisms)", m)
			}

			// Authenticate.
			if err := n.client.Auth(auth); err != nil {
				return err
			}

			log.Debug("Notifier SMTP client authenticated successfully with the server")

			return nil
		}

		return errors.New("Notifier SMTP server does not advertise the AUTH extension but config requires AUTH (password specified), either disable AUTH, or use an SMTP host that supports AUTH PLAIN or AUTH LOGIN")
	}

	log.Debug("Notifier SMTP config has no password specified so authentication is being skipped")

	return nil
}

func (n *SMTPNotifier) compose(recipient, subject, body, htmlBody string) error {
	log.Debugf("Notifier SMTP client attempting to send email body to %s", recipient)

	if !n.disableRequireTLS {
		_, ok := n.client.TLSConnectionState()
		if !ok {
			return errors.New("Notifier SMTP client can't send an email over plain text connection")
		}
	}

	wc, err := n.client.Data()
	if err != nil {
		log.Debugf("Notifier SMTP client error while obtaining WriteCloser: %s", err)
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
		log.Debugf("Notifier SMTP client error while sending email body over WriteCloser: %s", err)
		return err
	}

	err = wc.Close()
	if err != nil {
		log.Debugf("Notifier SMTP client error while closing the WriteCloser: %s", err)
		return err
	}

	return nil
}

// Dial the SMTP server with the SMTPNotifier config.
func (n *SMTPNotifier) dial() error {
	log.Debugf("Notifier SMTP client attempting connection to %s", n.address)

	if n.port == 465 {
		log.Warnf("Notifier SMTP client configured to connect to a SMTPS server. It's highly recommended you use a non SMTPS port and STARTTLS instead of SMTPS, as the protocol is long deprecated.")

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

	log.Debug("Notifier SMTP client connected successfully")

	return nil
}

// Closes the connection properly.
func (n *SMTPNotifier) cleanup() {
	err := n.client.Quit()
	if err != nil {
		log.Warnf("Notifier SMTP client encountered error during cleanup: %s", err)
	}
}

// StartupCheck checks the server is functioning correctly and the configuration is correct.
func (n *SMTPNotifier) StartupCheck() (bool, error) {
	if err := n.dial(); err != nil {
		return false, err
	}

	defer n.cleanup()

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

	// Start TLS and then Authenticate.
	if err := n.startTLS(); err != nil {
		return err
	}

	if err := n.auth(); err != nil {
		return err
	}

	// Set the sender and recipient first.
	if err := n.client.Mail(n.sender); err != nil {
		log.Debugf("Notifier SMTP failed while sending MAIL FROM (using sender) with error: %s", err)
		return err
	}

	if err := n.client.Rcpt(recipient); err != nil {
		log.Debugf("Notifier SMTP failed while sending RCPT TO (using recipient) with error: %s", err)
		return err
	}

	// Compose and send the email body to the server.
	if err := n.compose(recipient, subject, body, htmlBody); err != nil {
		return err
	}

	log.Debug("Notifier SMTP client successfully sent email")

	return nil
}
