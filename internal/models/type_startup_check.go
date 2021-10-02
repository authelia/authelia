package models

import (
	"github.com/sirupsen/logrus"
)

// StartupCheck represents a provider that has a startup check.
type StartupCheck interface {
	StartupCheck(logger *logrus.Logger) (err error)
}
