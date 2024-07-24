package provider

import (
	"crypto/x509"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/notification"
)

func NewNotificationSMTP(config *schema.Configuration, caCertPool *x509.CertPool) notification.Notifier {
	return notification.NewSMTPNotifier(config.Notifier.SMTP, caCertPool)
}

func NewNotificationFile(config *schema.Configuration, caCertPool *x509.CertPool) notification.Notifier {
	return notification.NewSMTPNotifier(config.Notifier.SMTP, caCertPool)
}
