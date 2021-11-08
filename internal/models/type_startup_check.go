package models

// StartupCheck represents a provider that has a startup check.
type StartupCheck interface {
	StartupCheck() (err error)
}
