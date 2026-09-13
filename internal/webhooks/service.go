// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"sync"

	"github.com/sirupsen/logrus"
)

// Service wraps a Dispatcher so it can participate in the service lifecycle.
type Service struct {
	name       string
	dispatcher *Dispatcher
	startup    bool
	done       chan struct{}
	log        *logrus.Entry
	shutdown   sync.Once
}

// NewService returns a Service for a Dispatcher.
func NewService(name string, dispatcher *Dispatcher, startup bool, log *logrus.Entry) (service *Service) {
	return &Service{
		name:       name,
		dispatcher: dispatcher,
		startup:    startup,
		done:       make(chan struct{}),
		log:        log,
	}
}

// ServiceType returns the service type for this service, which is always 'webhooks'.
func (service *Service) ServiceType() string {
	return "webhooks"
}

// ServiceName returns the individual name for this service.
func (service *Service) ServiceName() string {
	return service.name
}

// Log returns the logger configured for this service.
func (service *Service) Log() *logrus.Entry {
	return service.log
}

// Run starts the dispatcher workers and blocks until shutdown.
func (service *Service) Run() (err error) {
	if err = service.dispatcher.Validate(); err != nil {
		service.log.WithError(err).Error("One or more webhook destinations did not confirm they accept deliveries and will receive no events. Startup continues because webhook availability must not prevent authentication")
	}

	service.dispatcher.Start()

	if service.startup {
		if err = service.dispatcher.StartupCheck(); err != nil {
			service.log.WithError(err).Warn("One or more webhook destinations failed the startup check. Startup continues because webhook availability must not prevent authentication")
		}
	}

	<-service.done

	return nil
}

// Shutdown drains the dispatcher and stops the service.
func (service *Service) Shutdown() {
	service.shutdown.Do(func() {
		service.dispatcher.Shutdown()

		close(service.done)
	})
}
