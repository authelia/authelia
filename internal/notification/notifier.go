package notification

import (
	"github.com/sirupsen/logrus"
)

// Notifier interface for sending the identity verification link.
type Notifier interface {
	Send(recipient, subject, body, htmlBody string) (err error)
	StartupCheck(logger *logrus.Logger) (err error)
}
