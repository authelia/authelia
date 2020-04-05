package notification

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
)

// SMTPNotifier a notifier to send emails to SMTP servers.
type SMTPNotifier struct {
	username          string
	password          string
	sender            string
	host              string
	port              int
	trustedCert       string
	disableVerifyCert bool
	disableRequireTLS bool
	address           string
	client            *smtp.Client
	tlsConfig         *tls.Config
}

// NewSMTPNotifier create an SMTPNotifier targeting a given address.
func NewSMTPNotifier(configuration schema.SMTPNotifierConfiguration) *SMTPNotifier {
	notifier := &SMTPNotifier{
		username:          configuration.Username,
		password:          configuration.Password,
		sender:            configuration.Sender,
		host:              configuration.Host,
		port:              configuration.Port,
		trustedCert:       configuration.TrustedCert,
		disableVerifyCert: configuration.DisableVerifyCert,
		disableRequireTLS: configuration.DisableRequireTLS,
		address:           fmt.Sprintf("%s:%d", configuration.Host, configuration.Port),
	}
	notifier.initializeTLSConfig()
	return notifier
}

func (n *SMTPNotifier) initializeTLSConfig() {
	// Do not allow users to disable verification of certs if they have also set a trusted cert that was loaded.
	// The second part of this check happens in the Configure Cert Pool code block.
	log.Debug("Notifier SMTP client initializing TLS configuration")
	insecureSkipVerify := false
	if n.disableVerifyCert {
		insecureSkipVerify = true
	}

	//Configure Cert Pool.
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
						insecureSkipVerify = false
					}
				}
			}
		} else {
			log.Warnf("Notifier SMTP failed to load cert from file (file does not exist) with error: %s", err)
		}
	}
	n.tlsConfig = &tls.Config{
		InsecureSkipVerify: insecureSkipVerify,
		ServerName:         n.host,
		RootCAs:            certPool,
	}
}

// Do startTLS if available (some servers only provide the auth extension after, and encryption is preferred).
func (n *SMTPNotifier) startTLS() (bool, error) {
	// Only start if not already encrypted.
	if _, ok := n.client.TLSConnectionState(); ok {
		log.Debugf("Notifier SMTP connection is already encrypted, skipping STARTTLS")
		return ok, nil
	}

	ok, _ := n.client.Extension("STARTTLS")
	if ok {
		log.Debugf("Notifier SMTP server supports STARTTLS (disableVerifyCert: %t, ServerName: %s), attempting", n.tlsConfig.InsecureSkipVerify, n.tlsConfig.ServerName)

		err := n.client.StartTLS(n.tlsConfig)
		if err != nil {
			return ok, err
		} else {
			log.Debug("Notifier SMTP STARTTLS completed without error")
		}
	} else if n.disableRequireTLS {
		log.Warn("Notifier SMTP server does not support STARTTLS and SMTP configuration is set to disable the TLS requirement (only useful for unauthenticated emails over plain text)")
	} else {
		return ok, errors.New("Notifier SMTP server does not support TLS and it is required by default (see documentation if you want to disable this highly recommended requirement)")
	}
	return ok, nil
}

// Attempt Authentication.
func (n *SMTPNotifier) auth() (bool, error) {
	// Attempt AUTH if password is specified only.
	if n.password != "" {
		_, ok := n.client.TLSConnectionState()
		if !ok {
			return false, errors.New("Notifier SMTP client does not support authentication over plain text and the connection is currently plain text")
		}

		// Check the server supports AUTH, and get the mechanisms.
		ok, m := n.client.Extension("AUTH")
		if ok {
			log.Debugf("Notifier SMTP server supports authentication with the following mechanisms: %s", m)
			mechanisms := strings.Split(m, " ")
			var auth smtp.Auth

			// Adaptively select the AUTH mechanism to use based on what the server advertised
			if utils.IsStringInSlice("PLAIN", mechanisms) {
				auth = smtp.PlainAuth("", n.username, n.password, n.host)
				log.Debug("Notifier SMTP client attempting AUTH PLAIN with server")
			} else if utils.IsStringInSlice("LOGIN", mechanisms) {
				auth = newLoginAuth(n.username, n.password, n.host)
				log.Debug("Notifier SMTP client attempting AUTH LOGIN with server")
			}

			// Throw error since AUTH extension is not supported.
			if auth == nil {
				return false, fmt.Errorf("notifier SMTP server does not advertise a AUTH mechanism that are supported by Authelia (PLAIN or LOGIN are supported, but server advertised %s mechanisms)", m)
			}

			// Authenticate.
			err := n.client.Auth(auth)
			if err != nil {
				return false, err
			} else {
				log.Debug("Notifier SMTP client authenticated successfully with the server")
				return true, nil
			}
		} else {
			return false, errors.New("Notifier SMTP server does not advertise the AUTH extension but config requires AUTH (password specified), either disable AUTH, or use an SMTP host that supports AUTH PLAIN or AUTH LOGIN")
		}
	} else {
		log.Debug("Notifier SMTP config has no password specified so authentication is being skipped")
		return false, nil
	}
}

func (n *SMTPNotifier) compose(recipient, subject, body string) error {
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

	msg := "From: " + n.sender + "\n" +
		"To: " + recipient + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		body

	_, err = fmt.Fprintf(wc, msg)
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

// Send an email
func (n *SMTPNotifier) Send(recipient, subject, body string) error {
	if err := n.dial(); err != nil {
		return err
	}

	// Always execute QUIT at the end once we're connected.
	defer n.cleanup()

	// Start TLS and then Authenticate.
	if _, err := n.startTLS(); err != nil {
		return err
	}
	if _, err := n.auth(); err != nil {
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
	if err := n.compose(recipient, subject, body); err != nil {
		return err
	}

	log.Debug("Notifier SMTP client successfully sent email")
	return nil
}
