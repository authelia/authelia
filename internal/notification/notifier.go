package notification

import (
	"github.com/authelia/authelia/v4/internal/models"
)

// Notifier interface for sending the identity verification link.
type Notifier interface {
	models.StartupCheck

	Send(recipient, subject, body, htmlBody string) (err error)
}
