package service

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/logging"
)

func ProvisionLoggingSignal(ctx Context) (service Provider, err error) {
	config := ctx.GetConfiguration()

	if config == nil || len(config.Log.FilePath) == 0 {
		return nil, nil
	}

	return &Signal{
		name:    "log-reload",
		signals: []os.Signal{syscall.SIGHUP},
		action: func() (bubble bool, err error) {
			return false, logging.Reopen()
		},
		log: ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeSignal, serviceTypeSignal: "log-reload"}),
	}, nil
}

// ProvisionApplicationReloadSignal creates a Signal service that performs an effective application reload.
func ProvisionApplicationReloadSignal(ctx Context) (service Provider, err error) {
	return &Signal{
		name:    "application-reload",
		signals: []os.Signal{syscall.SIGUSR1},
		action: func() (bubble bool, err error) {
			return true, ErrApplicationReload
		},
		log: ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeSignal, serviceTypeSignal: "application-reload"}),
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
			if bubble, err := service.action(); err != nil {
				if bubble {
					return err
				}

				service.log.WithError(err).Error("Error occurred executing service action")
			} else {
				service.log.WithFields(map[string]any{"signal-received": s.String()}).Debug("Successfully executed service action")
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
