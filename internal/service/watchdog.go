package service

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/systemd"
)

// ProvisionSystemdWatchdog returns a Provider which periodically notifies the systemd service manager that this
// process is still alive. A nil Provider is returned when the service manager has not requested a watchdog.
func ProvisionSystemdWatchdog(ctx Context) (service Provider, err error) {
	var interval time.Duration

	if interval, err = systemd.WatchdogInterval(); err != nil {
		return nil, err
	}

	if interval <= 0 {
		return nil, nil
	}

	var notifier *systemd.Notifier

	if notifier, err = systemd.NewNotifier(); err != nil {
		return nil, err
	}

	if notifier == nil {
		return nil, nil
	}

	return &Watchdog{
		interval: interval / 2,
		notifier: notifier,
		log:      ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeWatchdog, serviceTypeWatchdog: serviceNameSystemd}),
		quit:     make(chan struct{}),
	}, nil
}

// Watchdog is a Provider which keeps the systemd service manager watchdog alive.
type Watchdog struct {
	interval time.Duration
	notifier *systemd.Notifier

	log  *logrus.Entry
	once sync.Once
	quit chan struct{}
}

// ServiceType returns the service type for this service, which is always 'watchdog'.
func (service *Watchdog) ServiceType() string {
	return serviceTypeWatchdog
}

// ServiceName returns the individual name for this service.
func (service *Watchdog) ServiceName() string {
	return serviceNameSystemd
}

// Run the Watchdog.
func (service *Watchdog) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			service.log.WithError(recoverErr(r)).Error("Critical error caught (recovered)")
		}
	}()

	service.log.WithField(logFieldInterval, service.interval.String()).Info("Notifying the service manager watchdog")

	ticker := time.NewTicker(service.interval)

	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err = service.notifier.Watchdog(); err != nil {
				service.log.WithError(err).Error("Error occurred notifying the service manager watchdog")
			}
		case <-service.quit:
			return nil
		}
	}
}

// Shutdown the Watchdog.
func (service *Watchdog) Shutdown() {
	service.once.Do(func() {
		close(service.quit)
	})

	if err := service.notifier.Close(); err != nil {
		service.log.WithError(err).Error("Error occurred during shutdown")
	}
}

// Log returns the *logrus.Entry of the Watchdog.
func (service *Watchdog) Log() *logrus.Entry {
	return service.log
}
