package service

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/middlewares"
)

// ProvisionGarbageCollector provisions the service which periodically performs the garbage collection of every
// middlewares.GarbageCollectorProvider registered with the collector.
func ProvisionGarbageCollector(ctx Context) (service Provider, err error) {
	collector := ctx.GetProviders().GarbageCollector

	if collector == nil {
		return nil, nil
	}

	return NewGarbageCollector("main", collector, ctx, ctx.GetLogger()), nil
}

// NewGarbageCollector creates a new GarbageCollector with the appropriate logger etc.
func NewGarbageCollector(name string, collector *middlewares.GarbageCollector, ctx context.Context, log *logrus.Entry) (service *GarbageCollector) {
	cctx, cancel := context.WithCancel(ctx)

	return &GarbageCollector{
		name:      name,
		collector: collector,
		ctx:       cctx,
		cancel:    cancel,
		log:       log.WithFields(map[string]any{logFieldService: serviceTypeGC, serviceTypeGC: name}),
	}
}

// GarbageCollector is a Provider which schedules the garbage collection of every registered
// middlewares.GarbageCollectorProvider at the frequency that provider requires.
type GarbageCollector struct {
	name      string
	collector *middlewares.GarbageCollector
	ctx       context.Context
	cancel    context.CancelFunc
	log       *logrus.Entry
}

// ServiceType returns the service type for this service, which is always 'gc'.
func (service *GarbageCollector) ServiceType() string {
	return serviceTypeGC
}

// ServiceName returns the individual name for this service.
func (service *GarbageCollector) ServiceName() string {
	return service.name
}

// Run the GarbageCollector. The registered providers are read when the service starts, which is after every service
// has been provisioned, so the order in which the providers are provisioned is not relevant.
func (service *GarbageCollector) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			service.log.WithError(recoverErr(r)).Error("Critical error caught (recovered)")
		}
	}()

	wg := &sync.WaitGroup{}

	for i, provider := range service.collector.Providers() {
		frequency := provider.GarbageCollectionFrequency(service.ctx)

		log := service.log.WithFields(map[string]any{logFieldProvider: i + 1, logFieldFrequency: frequency.String()})

		if frequency <= 0 {
			log.Trace("Garbage collection skipped as the provider does not require it")

			continue
		}

		wg.Add(1)

		go func() {
			defer wg.Done()

			service.collect(provider, frequency, log)
		}()
	}

	wg.Wait()

	return nil
}

// Shutdown the GarbageCollector.
func (service *GarbageCollector) Shutdown() {
	service.cancel()
}

// Log returns the *logrus.Entry of the GarbageCollector.
func (service *GarbageCollector) Log() *logrus.Entry {
	return service.log
}

func (service *GarbageCollector) collect(provider middlewares.GarbageCollectorProvider, frequency time.Duration, log *logrus.Entry) {
	defer func() {
		if r := recover(); r != nil {
			log.WithError(recoverErr(r)).Error("Critical error caught (recovered)")
		}
	}()

	ticker := time.NewTicker(frequency)

	defer ticker.Stop()

	log.Trace("Garbage collection scheduled")

	for {
		select {
		case <-service.ctx.Done():
			return
		case <-ticker.C:
			log.Trace("Garbage collection initiated")

			if err := provider.GarbageCollection(service.ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					log.WithError(err).Trace("Garbage collection cancelled")

					continue
				}

				log.WithError(err).Error("Error occurred performing garbage collection")

				continue
			}

			log.Trace("Garbage collection complete")
		}
	}
}
