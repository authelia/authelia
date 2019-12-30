package notification

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/authelia/authelia/internal/configuration/schema"
	"github.com/authelia/authelia/internal/utils"
	log "github.com/sirupsen/logrus"
)

// SMTPNotifier a notifier to send emails to SMTP servers.
type SMTPNotifier struct {
	username string
	password string
	sender   string
	host     string
	port     int
	secure   bool
	address  string
}

// NewSMTPNotifier create an SMTPNotifier targeting a given address.
func NewSMTPNotifier(configuration schema.SMTPNotifierConfiguration) *SMTPNotifier {
	return &SMTPNotifier{
		username: configuration.Username,
		password: configuration.Password,
		sender:   configuration.Sender,
		host:     configuration.Host,
		port:     configuration.Port,
		secure:   configuration.Secure,
		address:  fmt.Sprintf("%s:%d", configuration.Host, configuration.Port),
	}
}

// Send send a identity verification link to a user.
func (n *SMTPNotifier) Send(recipient string, subject string, body string) error {
	msg := "From: " + n.sender + "\n" +
		"To: " + recipient + "\n" +
		"Subject: " + subject + "\n" +
		"MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n" +
		body

	c, err := smtp.Dial(n.address)

	if err != nil {
		return err
	}

	// Do StartTLS if available (some servers only provide the auth extnesion after, and encryption is preferred)
	starttls, _ := c.Extension("STARTTLS")
	if starttls {
		tlsconfig := &tls.Config{
			InsecureSkipVerify: !n.secure,
			ServerName:         n.host,
		}
		log.Debugf("SMTP server supports STARTTLS (InsecureSkipVerify: %t, ServerName: %s), attempting", tlsconfig.InsecureSkipVerify, tlsconfig.ServerName)
		err := c.StartTLS(tlsconfig)
		if err != nil {
			return err
		} else {
			log.Debug("SMTP STARTTLS completed without error")
		}
	} else {
		log.Debug("SMTP server does not support STARTTLS, skipping")
	}

	// Attempt AUTH if password is specified only
	if n.password != "" {
		if !starttls {
			log.Warn("Authentication is being attempted over an insecure connection. Using a SMTP server that supports STARTTLS is recommended, especially if the server is not on your local network (username and pasword are being transmitted in plain-text).")
		}

		// Check the server supports AUTH, and get the mechanisms
		authExtension, m := c.Extension("AUTH")
		if authExtension {
			log.Debugf("Config has SMTP password and server supports AUTH with the following mechanisms: %s.", m)
			mechanisms := strings.Split(m, " ")
			var auth smtp.Auth

			// Adaptively select the AUTH mechanism to use based on what the server advertised
			if utils.IsStringInSlice("PLAIN", mechanisms) {
				auth = smtp.PlainAuth("", n.username, n.password, n.host)
				log.Debug("SMTP server supports AUTH PLAIN, attempting...")
			} else if utils.IsStringInSlice("LOGIN", mechanisms) {
				auth = LoginAuth(n.username, n.password)
				log.Debug("SMTP server supports AUTH LOGIN, attempting...")
			}

			// Throw error since AUTH extension is not supported
			if auth == nil {
				return fmt.Errorf("SMTP server does not advertise a AUTH mechanism that Authelia supports (PLAIN or LOGIN). Advertised mechanisms: %s.", m)
			}

			// Authenticate
			err := c.Auth(auth)
			if err != nil {
				return err
			} else {
				log.Debug("SMTP AUTH completed successfully.")
			}
		} else {
			return errors.New("SMTP server does not advertise the AUTH extension but a password was specified. Either disable auth (don't specify a password/comment the password), or specify an SMTP host and port that supports AUTH PLAIN or AUTH LOGIN.")
		}
	} else {
		log.Debug("SMTP config has no password specified for use with AUTH, skipping.")
	}

	// Set the sender and recipient first
	if err := c.Mail(n.sender); err != nil {
		return err
	}

	if err := c.Rcpt(recipient); err != nil {
		return err
	}

	// Send the email body.
	wc, err := c.Data()
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(wc, msg)
	if err != nil {
		return err
	}
	err = wc.Close()
	if err != nil {
		return err
	}

	// Send the QUIT command and close the connection.
	err = c.Quit()
	if err != nil {
		return err
	}
	return nil
}
