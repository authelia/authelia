package service

import (
	"time"

	"github.com/sirupsen/logrus"
)

// ProvisionOAuth2DPoPGarbageCollector provisions the GarbageCollector which removes the expired rows of the OAuth2.0
// DPoP tables. Both tables gain a row per request which is only meaningful until the value it records expires, so
// without this the tables grow for the lifetime of the deployment.
func ProvisionOAuth2DPoPGarbageCollector(ctx Context) (service Provider, err error) {
	config := ctx.GetConfiguration()

	if config == nil || config.IdentityProviders.OIDC == nil || !config.IdentityProviders.OIDC.DPoP.Enabled {
		return nil, nil
	}

	return NewGarbageCollector("oauth2-dpop", intervalGarbageCollectionOAuth2DPoP, ctx, collectOAuth2DPoP), nil
}

// NewGarbageCollector creates a new GarbageCollector with the appropriate logger etc.
func NewGarbageCollector(name string, interval time.Duration, ctx Context, collect func(ctx Context) (err error)) (service *GarbageCollector) {
	return &GarbageCollector{
		name:     name,
		interval: interval,
		ctx:      ctx,
		collect:  collect,
		log:      ctx.GetLogger().WithFields(map[string]any{logFieldService: serviceTypeGarbageCollector, serviceTypeGarbageCollector: name}),
		quit:     make(chan struct{}),
	}
}

// GarbageCollector is a Provider which periodically performs a collection of data which is no longer required.
type GarbageCollector struct {
	name     string
	interval time.Duration

	ctx     Context
	collect func(ctx Context) (err error)

	log  *logrus.Entry
	quit chan struct{}
}

// ServiceType returns the service type for this service, which is always 'gc'.
func (service *GarbageCollector) ServiceType() string {
	return serviceTypeGarbageCollector
}

// ServiceName returns the individual name for this service.
func (service *GarbageCollector) ServiceName() string {
	return service.name
}

// Run the GarbageCollector.
func (service *GarbageCollector) Run() (err error) {
	ticker := time.NewTicker(service.interval)

	defer ticker.Stop()

	service.log.WithField(logFieldInterval, service.interval.String()).Debug("Performing garbage collection on an interval")

	service.run()

	for {
		select {
		case <-ticker.C:
			service.run()
		case <-service.quit:
			return nil
		}
	}
}

// Shutdown the GarbageCollector.
func (service *GarbageCollector) Shutdown() {
	close(service.quit)
}

// Log returns the *logrus.Entry of the GarbageCollector.
func (service *GarbageCollector) Log() *logrus.Entry {
	return service.log
}

func (service *GarbageCollector) run() {
	defer func() {
		if r := recover(); r != nil {
			service.log.WithError(recoverErr(r)).Error("Critical error caught (recovered)")
		}
	}()

	if err := service.collect(service.ctx); err != nil {
		if service.ctx.Err() != nil {
			return
		}

		service.log.WithError(err).Error("Error occurred performing garbage collection")

		return
	}

	service.log.Trace("Garbage collection completed successfully")
}

func collectOAuth2DPoP(ctx Context) (err error) {
	provider, now := ctx.GetProviders().StorageProvider, time.Now()

	if err = provider.DeleteExpiredOAuth2DPoPProofs(ctx, now); err != nil {
		return err
	}

	return provider.DeleteExpiredOAuth2DPoPNonces(ctx, now)
}
