package service

import (
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func ProvisionSessionCollector(ctx Context) (service Provider, err error) {
	repository := ctx.GetProviders().SessionRepository

	if repository == nil {
		return nil, nil
	}

	collector := NewCollector("session", repository, ctx)

	// The nil check is performed on the concrete type as returning a nil *Collector directly would produce a non-nil
	// Provider interface.
	if collector == nil {
		return nil, nil
	}

	return collector, nil
}

// NewCollector creates a new Collector with the appropriate logger etc. A nil Collector is returned when the
// GarbageCollectorProvider reports a frequency of zero as it expires records itself.
func NewCollector(name string, collector GarbageCollectorProvider, ctx Context) (service *Collector) {
	frequency := collector.GarbageCollectionFrequency(ctx)

	if frequency <= 0 {
		return nil
	}

	return &Collector{
		name:      name,
		collector: collector,
		frequency: frequency,
		ctx:       ctx,
		log:       ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeCollector, serviceTypeCollector: name}),
		quit:      make(chan struct{}),
	}
}

// Collector is a Provider which periodically performs garbage collection on a GarbageCollectorProvider.
type Collector struct {
	name      string
	collector GarbageCollectorProvider
	frequency time.Duration
	ctx       Context

	log  *logrus.Entry
	once sync.Once
	quit chan struct{}
}

// ServiceType returns the service type for this service, which is always 'collector'.
func (service *Collector) ServiceType() string {
	return serviceTypeCollector
}

// ServiceName returns the individual name for this service.
func (service *Collector) ServiceName() string {
	return service.name
}

// Run the Collector.
func (service *Collector) Run() (err error) {
	defer func() {
		if r := recover(); r != nil {
			service.log.WithError(recoverErr(r)).Error("Critical error caught (recovered)")
		}
	}()

	ticker := time.NewTicker(service.frequency)

	defer ticker.Stop()

	service.log.WithField(logFieldFrequency, service.frequency.String()).Info("Collecting expired records on a frequency")

	for {
		select {
		case <-ticker.C:
			service.collect()
		case <-service.ctx.Done():
			return nil
		case <-service.quit:
			return nil
		}
	}
}

// Shutdown the Collector.
func (service *Collector) Shutdown() {
	service.once.Do(func() {
		close(service.quit)
	})
}

// Log returns the *logrus.Entry of the Collector.
func (service *Collector) Log() *logrus.Entry {
	return service.log
}

func (service *Collector) collect() {
	defer func() {
		if r := recover(); r != nil {
			service.log.WithError(recoverErr(r)).Error("Critical error caught (recovered) collecting expired records")
		}
	}()

	if service.ctx.Err() != nil {
		return
	}

	if err := service.collector.GarbageCollection(service.ctx); err != nil {
		if service.ctx.Err() == nil {
			service.log.WithError(err).Error("Error occurred collecting expired records")
		}

		return
	}

	service.log.Debug("Collection of expired records completed successfully")
}

var (
	_ Provider = (*Collector)(nil)
)
