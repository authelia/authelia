package middlewares

import (
	"context"
	"sync"
	"time"
)

// GarbageCollectorProvider represents the required methods to support periodic garbage collection of a backend. A
// backend which expires records itself, such as Redis, should return a frequency of zero to indicate no collection
// service is required.
type GarbageCollectorProvider interface {
	GarbageCollection(ctx context.Context) (err error)
	GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration)
}

// NewGarbageCollector returns a new *GarbageCollector.
func NewGarbageCollector() (collector *GarbageCollector) {
	return &GarbageCollector{}
}

// GarbageCollector is a registry of GarbageCollectorProvider implementations which allows their garbage collection to
// be centrally scheduled. The zero value is ready for use and a nil *GarbageCollector is a no-op which allows
// consumers to omit the collector entirely.
type GarbageCollector struct {
	mu        sync.Mutex
	providers []GarbageCollectorProvider
}

// Register the given GarbageCollectorProvider implementations with this collector.
func (c *GarbageCollector) Register(providers ...GarbageCollectorProvider) {
	if c == nil || len(providers) == 0 {
		return
	}

	c.mu.Lock()

	defer c.mu.Unlock()

	c.providers = append(c.providers, providers...)
}

// Providers returns a copy of the registered GarbageCollectorProvider implementations.
func (c *GarbageCollector) Providers() (providers []GarbageCollectorProvider) {
	if c == nil {
		return nil
	}

	c.mu.Lock()

	defer c.mu.Unlock()

	providers = make([]GarbageCollectorProvider, len(c.providers))

	copy(providers, c.providers)

	return providers
}

// Len returns the number of registered GarbageCollectorProvider implementations.
func (c *GarbageCollector) Len() (n int) {
	if c == nil {
		return 0
	}

	c.mu.Lock()

	defer c.mu.Unlock()

	return len(c.providers)
}
