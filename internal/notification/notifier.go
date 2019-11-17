package notification

// Notifier interface for sending the identity verification link.
type Notifier interface {
	Send(to string, subject string, link string) error
}
