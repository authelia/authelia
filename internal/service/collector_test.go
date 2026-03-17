package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/session"
)

func TestProvisionSessionCollector(t *testing.T) {
	testCases := []struct {
		Name       string
		Repository session.Repository
		Expected   bool
	}{
		{"ShouldProvisionWithFrequency", &mockGarbageCollector{frequency: time.Minute}, true},
		{"ShouldNotProvisionWithoutFrequency", &mockGarbageCollector{}, false},
		{"ShouldNotProvisionWithoutRepository", nil, false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := newMockServiceCtx()
			ctx.providers.SessionRepository = tc.Repository

			service, err := ProvisionSessionCollector(ctx)
			require.NoError(t, err)

			if tc.Expected {
				require.NotNil(t, service)
				assert.Equal(t, serviceTypeCollector, service.ServiceType())
				assert.Equal(t, "session", service.ServiceName())
				assert.NotNil(t, service.Log())
			} else {
				assert.Nil(t, service)
			}
		})
	}
}

func TestNewCollector(t *testing.T) {
	testCases := []struct {
		Name      string
		Frequency time.Duration
		Expected  bool
	}{
		{"ShouldCreateWithPositiveFrequency", time.Minute, true},
		{"ShouldNotCreateWithZeroFrequency", 0, false},
		{"ShouldNotCreateWithNegativeFrequency", -time.Minute, false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := newMockServiceCtx()

			service := NewCollector("session", &mockGarbageCollector{frequency: tc.Frequency}, ctx)

			if tc.Expected {
				require.NotNil(t, service)
				assert.Equal(t, tc.Frequency, service.frequency)
			} else {
				assert.Nil(t, service)
			}
		})
	}
}

func TestCollector_Collect(t *testing.T) {
	testCases := []struct {
		Name     string
		Err      error
		Expected int
	}{
		{"ShouldCollect", nil, 1},
		{"ShouldHandleCollectionError", errors.New("collection failed"), 1},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			collector := &mockGarbageCollector{frequency: time.Minute, err: tc.Err}

			NewCollector("session", collector, newMockServiceCtx()).collect()

			assert.Equal(t, tc.Expected, collector.Calls())
		})
	}
}

func TestCollector_CollectSkipsWhenContextCancelled(t *testing.T) {
	collector := &mockGarbageCollector{frequency: time.Minute}

	cctx, cancel := context.WithCancel(context.Background())

	ctx := newMockServiceCtx()
	ctx.ctx = cctx

	service := NewCollector("session", collector, ctx)

	cancel()

	service.collect()

	assert.Equal(t, 0, collector.Calls())
}

func TestCollector_RunCollectsOnFrequency(t *testing.T) {
	collector := &mockGarbageCollector{frequency: time.Millisecond, notify: make(chan struct{}, 1)}

	service := NewCollector("session", collector, newMockServiceCtx())

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	select {
	case <-collector.notify:
	case <-time.After(time.Second * 5):
		t.Fatal("timeout waiting for collection")
	}

	service.Shutdown()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("timeout waiting for shutdown")
	}
}

func TestCollector_RunStopsOnContextCancellation(t *testing.T) {
	cctx, cancel := context.WithCancel(context.Background())

	ctx := newMockServiceCtx()
	ctx.ctx = cctx

	service := NewCollector("session", &mockGarbageCollector{frequency: time.Minute}, ctx)

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("timeout waiting for shutdown")
	}
}

func TestCollector_ShutdownIsIdempotent(t *testing.T) {
	service := NewCollector("session", &mockGarbageCollector{frequency: time.Minute}, newMockServiceCtx())

	assert.NotPanics(t, func() {
		service.Shutdown()
		service.Shutdown()
	})
}

type mockGarbageCollector struct {
	session.Repository

	mu        sync.Mutex
	frequency time.Duration
	err       error
	calls     int
	notify    chan struct{}
}

func (m *mockGarbageCollector) GarbageCollectionFrequency(_ context.Context) (frequency time.Duration) {
	return m.frequency
}

func (m *mockGarbageCollector) GarbageCollection(_ context.Context) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls++

	if m.notify != nil {
		select {
		case m.notify <- struct{}{}:
		default:
		}
	}

	return m.err
}

func (m *mockGarbageCollector) Calls() (calls int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.calls
}
