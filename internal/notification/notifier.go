package notification

import (
	"net/mail"

	"github.com/authelia/authelia/v4/internal/model"
)

// Notifier interface for sending the identity verification link.
type Notifier interface {
	model.StartupCheck

	Send(recipient mail.Address, subject, body, htmlBody string) (err error)
}
