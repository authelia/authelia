package services

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

func ProvisionLoggingSignal(config *schema.Configuration, providers middlewares.Providers, log *logrus.Logger) (service Provider, err error) {
	return &Signal{
		name:    "log-reload",
		signals: []os.Signal{syscall.SIGHUP},
		action:  logging.Reopen,
		log:     log.WithFields(map[string]any{logFieldService: serviceTypeSignal, serviceTypeSignal: "log-reload"}),
	}, nil
}

// Signal is a Service which performs actions on signals.
type Signal struct {
	name    string
	signals []os.Signal
	action  func() (err error)
	log     *logrus.Entry

	notify chan os.Signal
	quit   chan struct{}
}

// ServiceType returns the service type for this service, which is always 'server'.
func (service *Signal) ServiceType() string {
	return serviceTypeSignal
}

// ServiceName returns the individual name for this service.
func (service *Signal) ServiceName() string {
	return service.name
}

// Run the ServerService.
func (service *Signal) Run() (err error) {
	service.quit = make(chan struct{})

	service.notify = make(chan os.Signal, 1)

	signal.Notify(service.notify, service.signals...)

	for {
		select {
		case s := <-service.notify:
			if err = service.action(); err != nil {
				service.log.WithError(err).Error("Error occurred executing service action.")
			} else {
				service.log.WithFields(map[string]any{"signal-received": s.String()}).Debug("Successfully executed service action.")
			}
		case <-service.quit:
			return
		}
	}
}

// Shutdown the ServerService.
func (service *Signal) Shutdown() {
	signal.Stop(service.notify)

	service.quit <- struct{}{}
}

// Log returns the *logrus.Entry of the ServerService.
func (service *Signal) Log() *logrus.Entry {
	return service.log
}
