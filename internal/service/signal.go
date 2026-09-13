// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/logging"
)

// ProvisionLoggingSignal returns a Provider which reopens the log file when the relevant signal is received.
func ProvisionLoggingSignal(ctx Context) (service Provider, err error) {
	config := ctx.GetConfiguration()

	if config == nil || len(config.Log.FilePath) == 0 {
		return nil, nil
	}

	return &Signal{
		name:    "log-reload",
		signals: []os.Signal{syscall.SIGUSR1},
		action: func() (bubble bool, err error) {
			return false, logging.Reopen()
		},
		log:    ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeSignal, serviceTypeSignal: "log-reload"}),
		notify: make(chan os.Signal, 1),
		quit:   make(chan struct{}),
	}, nil
}

// ProvisionApplicationReloadSignal returns a Provider which performs an effective application reload when the relevant
// signal is received.
func ProvisionApplicationReloadSignal(ctx Context) (service Provider, err error) {
	return &Signal{
		name:    "application-reload",
		signals: []os.Signal{syscall.SIGHUP},
		action: func() (bubble bool, err error) {
			return true, ErrApplicationReload
		},
		log:    ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeSignal, serviceTypeSignal: "application-reload"}),
		notify: make(chan os.Signal, 1),
		quit:   make(chan struct{}),
	}, nil
}

// Signal is a Service which performs actions on signals.
type Signal struct {
	name    string
	signals []os.Signal
	action  func() (bubble bool, err error)
	log     *logrus.Entry

	notify chan os.Signal
	quit   chan struct{}
	stop   sync.Once
}

// ServiceType returns the service type for this service, which is always 'signal'.
func (service *Signal) ServiceType() string {
	return serviceTypeSignal
}

// ServiceName returns the individual name for this service.
func (service *Signal) ServiceName() string {
	return service.name
}

// Run the ServerService.
func (service *Signal) Run() (err error) {
	signal.Notify(service.notify, service.signals...)

	defer signal.Stop(service.notify)

	for {
		select {
		case s := <-service.notify:
			log := service.log.WithFields(map[string]any{"signal-received": s.String()})

			var bubble bool

			switch bubble, err = service.action(); {
			case err != nil && bubble:
				return err
			case err != nil:
				log.WithError(err).Error("Error occurred executing service action.")
			default:
				log.Debug("Successfully executed service action.")
			}
		case <-service.quit:
			return nil
		}
	}
}

// Shutdown the ServerService.
func (service *Signal) Shutdown() {
	service.stop.Do(func() {
		signal.Stop(service.notify)

		close(service.quit)
	})
}

// Log returns the *logrus.Entry of the ServerService.
func (service *Signal) Log() *logrus.Entry {
	return service.log
}
