package service

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/systemd"
)

func ProvisionLoggingSignal(ctx Context) (service Provider, err error) {
	config := ctx.GetConfiguration()

	var action func() (err error)

	if config != nil && len(config.Log.FilePath) != 0 {
		action = logging.Reopen
	}

	var notifier *systemd.Notifier

	if notifier, err = systemd.NewNotifier(); err != nil {
		return nil, err
	}

	return &Signal{
		name:     serviceNameReload,
		signals:  []os.Signal{syscall.SIGHUP},
		action:   action,
		notifier: notifier,
		log:      ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeSignal, serviceTypeSignal: serviceNameReload}),
		notify:   make(chan os.Signal, 1),
		quit:     make(chan struct{}),
	}, nil
}

// Signal is a Service which performs actions on signals.
type Signal struct {
	name     string
	signals  []os.Signal
	action   func() (err error)
	notifier *systemd.Notifier
	log      *logrus.Entry

	notify chan os.Signal
	quit   chan struct{}
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

	for {
		select {
		case s := <-service.notify:
			service.handle(s)
		case <-service.quit:
			return
		}
	}
}

// Shutdown the ServerService.
func (service *Signal) Shutdown() {
	signal.Stop(service.notify)

	if err := service.notifier.Close(); err != nil {
		service.log.WithError(err).Error("Error occurred during shutdown")
	}

	service.quit <- struct{}{}
}

// Log returns the *logrus.Entry of the ServerService.
func (service *Signal) Log() *logrus.Entry {
	return service.log
}

func (service *Signal) handle(s os.Signal) {
	if err := service.notifier.Reloading(); err != nil {
		service.log.WithError(err).Error("Error occurred notifying the service manager of the reload.")
	}

	if service.action != nil {
		if err := service.action(); err != nil {
			service.log.WithError(err).Error("Error occurred executing service action.")
		} else {
			service.log.WithFields(map[string]any{"signal-received": s.String()}).Debug("Successfully executed service action.")
		}
	}

	if err := service.notifier.Ready(statusReady); err != nil {
		service.log.WithError(err).Error("Error occurred notifying the service manager the reload is complete.")
	}
}
