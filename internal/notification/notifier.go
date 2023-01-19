package notification

import (
	"context"
	"net/mail"

	"github.com/authelia/authelia/v4/internal/model"
	"github.com/authelia/authelia/v4/internal/templates"
)

// Notifier interface for sending the identity verification link.
type Notifier interface {
	model.StartupCheck

	Send(ctx context.Context, recipient mail.Address, subject string, et *templates.EmailTemplate, data any) (err error)
}
