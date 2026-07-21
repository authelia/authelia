package service

import (
	"github.com/sirupsen/logrus"
)

// Provider represents the required methods to support handling a service.
type Provider interface {
	// ServiceType returns the type name for the Provider.
	ServiceType() string

	// ServiceName returns the individual name for the Provider.
	ServiceName() string

	// Run performs the running operations for the Provider.
	Run() (err error)

	// Shutdown perform the shutdown cleanup and termination operations for the Provider.
	Shutdown()

	// Log returns the logger configured for the service.
	Log() *logrus.Entry
}

// ReloadableProvider represents the required methods to support reloading a provider.
type ReloadableProvider interface {
	Reload() (reloaded bool, err error)
}

type Provisioner func(ctx Context) (provider Provider, err error)

func GetProvisioners() []Provisioner {
	return []Provisioner{
		ProvisionServer,
		ProvisionServerMetrics,
		ProvisionUsersFileWatcher,
		ProvisionLoggingSignal,
	}
}
