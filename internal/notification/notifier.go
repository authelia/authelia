package notification

import (
	"github.com/authelia/authelia/v4/internal/model"
)

// Notifier interface for sending the identity verification link.
type Notifier interface {
	model.StartupCheck

	Send(recipient, subject, body, htmlBody string) (err error)
}
