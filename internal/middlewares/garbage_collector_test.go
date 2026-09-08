package middlewares

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGarbageCollectorRegister(t *testing.T) {
	testCases := []struct {
		name      string
		collector *GarbageCollector
		providers []GarbageCollectorProvider
		expected  int
	}{
		{
			"ShouldRegisterSingleProvider",
			NewGarbageCollector(),
			[]GarbageCollectorProvider{&testGarbageCollectorProvider{}},
			1,
		},
		{
			"ShouldRegisterMultipleProviders",
			NewGarbageCollector(),
			[]GarbageCollectorProvider{&testGarbageCollectorProvider{}, &testGarbageCollectorProvider{}, &testGarbageCollectorProvider{}},
			3,
		},
		{
			"ShouldHandleNoProviders",
			NewGarbageCollector(),
			nil,
			0,
		},
		{
			"ShouldHandleNilCollector",
			nil,
			[]GarbageCollectorProvider{&testGarbageCollectorProvider{}},
			0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.collector.Register(tc.providers...)

			assert.Equal(t, tc.expected, tc.collector.Len())
			assert.Len(t, tc.collector.Providers(), tc.expected)
		})
	}
}

func TestGarbageCollectorProviders(t *testing.T) {
	collector := NewGarbageCollector()

	provider := &testGarbageCollectorProvider{}

	collector.Register(provider)

	providers := collector.Providers()

	require.Len(t, providers, 1)
	assert.Same(t, provider, providers[0])

	providers[0] = nil

	assert.Same(t, provider, collector.Providers()[0])
}

func TestGarbageCollectorNilIsNoOp(t *testing.T) {
	var collector *GarbageCollector

	assert.NotPanics(t, func() {
		collector.Register(&testGarbageCollectorProvider{})

		assert.Equal(t, 0, collector.Len())
		assert.Nil(t, collector.Providers())
	})
}

func TestGarbageCollectorConcurrentRegisterAndRead(t *testing.T) {
	collector := NewGarbageCollector()

	wg := &sync.WaitGroup{}

	for range 10 {
		wg.Add(2)

		go func() {
			defer wg.Done()

			collector.Register(&testGarbageCollectorProvider{})
		}()

		go func() {
			defer wg.Done()

			collector.Providers()
		}()
	}

	wg.Wait()

	assert.Equal(t, 10, collector.Len())
}

type testGarbageCollectorProvider struct {
	mu        sync.Mutex
	n         int
	err       error
	frequency time.Duration
}

func (t *testGarbageCollectorProvider) GarbageCollection(ctx context.Context) (err error) {
	t.mu.Lock()

	defer t.mu.Unlock()

	t.n++

	return t.err
}

func (t *testGarbageCollectorProvider) GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	return t.frequency
}

func (t *testGarbageCollectorProvider) Count() (n int) {
	t.mu.Lock()

	defer t.mu.Unlock()

	return t.n
}
