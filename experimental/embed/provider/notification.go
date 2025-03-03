package provider

import (
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/notification"
)

// NewNotificationSMTP creates a new notification.Notifier using the *notification.SMTPNotifier given a valid
// configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewNotificationSMTP(config *schema.Configuration, caCertPool *x509.CertPool) notification.Notifier {
	return notification.NewSMTPNotifier(config.Notifier.SMTP, caCertPool)
}

// NewNotificationFile creates a new notification.Notifier using the *notification.FileNotifier given a valid
// configuration.
//
// Warning: This method may panic if the provided configuration isn't validated.
func NewNotificationFile(config *schema.Configuration, caCertPool *x509.CertPool) notification.Notifier {
	return notification.NewSMTPNotifier(config.Notifier.SMTP, caCertPool)
}
