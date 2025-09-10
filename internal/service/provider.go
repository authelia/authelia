package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
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
	Reload() (reloaded bool, err error, reason string)
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

type Context interface {
	GetLogger() *logrus.Entry
	GetProviders() middlewares.Providers
	GetConfiguration() *schema.Configuration

	context.Context
}
