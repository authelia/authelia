package notification

// Notifier interface for sending the identity verification link.
type Notifier interface {
	Send(recipient, subject, body, htmlBody string) error
	StartupCheck() (bool, error)
}
