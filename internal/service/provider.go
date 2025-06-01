// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"github.com/fsnotify/fsnotify"
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

// FileWatcherAction represents an action to perform when a FileWatcher is triggered. A bubble return value of true
// causes the returned error to terminate the service and be returned to the caller rather than just being logged.
type FileWatcherAction func(log *logrus.Entry, event fsnotify.Event) (bubble bool, err error)

// Provisioner is a function which provisions a Provider.
type Provisioner func(ctx Context) (provider Provider, err error)

// GetProvisioners returns every Provisioner in the order they should be provisioned.
func GetProvisioners() []Provisioner {
	return []Provisioner{
		ProvisionServer,
		ProvisionServerMetrics,
		ProvisionUsersFileWatcher,
		ProvisionConfigFileWatcher,
		ProvisionLoggingSignal,
		ProvisionApplicationReloadSignal,
		ProvisionGarbageCollector,
	}
}
