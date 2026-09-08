package service

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
)

func TestProvisionGarbageCollector(t *testing.T) {
	testCases := []struct {
		name      string
		collector *middlewares.GarbageCollector
		expected  bool
	}{
		{
			"ShouldProvisionWithCollector",
			middlewares.NewGarbageCollector(),
			true,
		},
		{
			"ShouldNotProvisionWithoutCollector",
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newMockServiceCtx()
			ctx.providers.GarbageCollector = tc.collector

			service, err := ProvisionGarbageCollector(ctx)

			require.NoError(t, err)

			if !tc.expected {
				assert.Nil(t, service)

				return
			}

			require.NotNil(t, service)

			assert.Equal(t, serviceTypeGC, service.ServiceType())
			assert.Equal(t, "main", service.ServiceName())
			assert.NotNil(t, service.Log())
		})
	}
}

func TestGarbageCollectorServiceRun(t *testing.T) {
	testCases := []struct {
		name      string
		frequency time.Duration
		err       error
		collected bool
	}{
		{
			"ShouldCollectAtFrequency",
			time.Millisecond * 10,
			nil,
			true,
		},
		{
			"ShouldContinueAfterError",
			time.Millisecond * 10,
			errors.New("collection failed"),
			true,
		},
		{
			"ShouldNotCollectWithZeroFrequency",
			0,
			nil,
			false,
		},
		{
			"ShouldNotCollectWithNegativeFrequency",
			-time.Minute,
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := &testGarbageCollectorProvider{frequency: tc.frequency, err: tc.err}

			collector := middlewares.NewGarbageCollector()
			collector.Register(provider)

			service := NewGarbageCollector("main", collector, t.Context(), newTestLogger())

			done := make(chan error, 1)

			go func() {
				done <- service.Run()
			}()

			if tc.collected {
				assert.Eventually(t, func() bool { return provider.Count() > 1 }, time.Second, time.Millisecond*10)
			} else {
				time.Sleep(time.Millisecond * 50)

				assert.Equal(t, 0, provider.Count())
			}

			service.Shutdown()

			select {
			case err := <-done:
				assert.NoError(t, err)
			case <-time.After(time.Second):
				t.Fatal("garbage collector service did not exit after shutdown")
			}
		})
	}
}

func TestGarbageCollectorServiceCollectionErrorLogging(t *testing.T) {
	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			"ShouldLogGenuineError",
			errors.New("collection failed"),
			true,
		},
		{
			"ShouldLogWrappedGenuineError",
			fmt.Errorf("bucket: %w", errors.New("collection failed")),
			true,
		},
		{
			"ShouldNotLogCancellation",
			context.Canceled,
			false,
		},
		{
			"ShouldNotLogWrappedCancellation",
			fmt.Errorf("bucket: %w", context.Canceled),
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			provider := &testGarbageCollectorProvider{frequency: time.Millisecond * 10, err: tc.err}

			collector := middlewares.NewGarbageCollector()
			collector.Register(provider)

			log, hook := newTestLoggerHook()

			service := NewGarbageCollector("main", collector, t.Context(), log)

			done := make(chan error, 1)

			go func() {
				done <- service.Run()
			}()

			assert.Eventually(t, func() bool { return provider.Count() > 0 }, time.Second, time.Millisecond*10)

			service.Shutdown()

			select {
			case err := <-done:
				require.NoError(t, err)
			case <-time.After(time.Second):
				t.Fatal("garbage collector service did not exit after shutdown")
			}

			assert.Equal(t, tc.expected, hasCollectionErrorEntry(hook))
		})
	}
}

func TestGarbageCollectorServiceDoesNotLogErrorWhenCancelledMidCollection(t *testing.T) {
	started, release := make(chan struct{}, 1), make(chan struct{})

	provider := &testGarbageCollectorProvider{
		frequency: time.Millisecond * 10,
		collect: func(ctx context.Context) (err error) {
			select {
			case started <- struct{}{}:
			default:
			}

			<-release

			return ctx.Err()
		},
	}

	collector := middlewares.NewGarbageCollector()
	collector.Register(provider)

	log, hook := newTestLoggerHook()

	service := NewGarbageCollector("main", collector, t.Context(), log)

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("garbage collection did not start")
	}

	service.Shutdown()

	close(release)

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("garbage collector service did not exit after shutdown")
	}

	assert.False(t, hasCollectionErrorEntry(hook))
}

func TestGarbageCollectorServiceRunsEveryProvider(t *testing.T) {
	fast := &testGarbageCollectorProvider{frequency: time.Millisecond * 10}
	slow := &testGarbageCollectorProvider{frequency: time.Hour}
	none := &testGarbageCollectorProvider{frequency: 0}

	collector := middlewares.NewGarbageCollector()
	collector.Register(fast, slow, none)

	service := NewGarbageCollector("main", collector, t.Context(), newTestLogger())

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	assert.Eventually(t, func() bool { return fast.Count() > 1 }, time.Second, time.Millisecond*10)

	assert.Equal(t, 0, slow.Count())
	assert.Equal(t, 0, none.Count())

	service.Shutdown()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("garbage collector service did not exit after shutdown")
	}
}

func TestGarbageCollectorServiceCancelsProviderContext(t *testing.T) {
	provider := &testGarbageCollectorProvider{frequency: time.Millisecond * 10}

	collector := middlewares.NewGarbageCollector()
	collector.Register(provider)

	service := NewGarbageCollector("main", collector, t.Context(), newTestLogger())

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	assert.Eventually(t, func() bool { return provider.Count() > 0 }, time.Second, time.Millisecond*10)

	service.Shutdown()

	<-done

	assert.Error(t, provider.Context().Err())
}

func TestGarbageCollectorServiceRecoversFromCollectionPanic(t *testing.T) {
	provider := &testGarbageCollectorProvider{
		frequency: time.Millisecond * 10,
		collect: func(ctx context.Context) (err error) {
			panic("collection exploded")
		},
	}

	collector := middlewares.NewGarbageCollector()
	collector.Register(provider)

	log, hook := newTestLoggerHook()

	service := NewGarbageCollector("main", collector, t.Context(), log)

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("garbage collector service did not exit after the collection panicked")
	}

	service.Shutdown()

	entry := findEntry(hook, logrus.ErrorLevel, "Critical error caught (recovered)")

	require.NotNil(t, entry)
	require.NotNil(t, entry.Data[logrus.ErrorKey])
	assert.EqualError(t, entry.Data[logrus.ErrorKey].(error), "recovered panic: collection exploded")
}

func TestGarbageCollectorServiceRecoversFromFrequencyPanic(t *testing.T) {
	provider := &testGarbageCollectorProvider{panicFrequency: errors.New("frequency exploded")}

	collector := middlewares.NewGarbageCollector()
	collector.Register(provider)

	log, hook := newTestLoggerHook()

	service := NewGarbageCollector("main", collector, t.Context(), log)

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("garbage collector service did not exit after the frequency panicked")
	}

	service.Shutdown()

	assert.Equal(t, 0, provider.Count())

	entry := findEntry(hook, logrus.ErrorLevel, "Critical error caught (recovered)")

	require.NotNil(t, entry)
	require.NotNil(t, entry.Data[logrus.ErrorKey])
	assert.EqualError(t, entry.Data[logrus.ErrorKey].(error), "recovered panic: frequency exploded")
}

func TestGarbageCollectorServiceShutdownWithoutProviders(t *testing.T) {
	service := NewGarbageCollector("main", middlewares.NewGarbageCollector(), t.Context(), newTestLogger())

	assert.NoError(t, service.Run())

	service.Shutdown()
}

func TestGarbageCollectorServiceExitsOnContextCancel(t *testing.T) {
	provider := &testGarbageCollectorProvider{frequency: time.Millisecond * 10}

	collector := middlewares.NewGarbageCollector()
	collector.Register(provider)

	ctx, cancel := context.WithCancel(t.Context())

	service := NewGarbageCollector("main", collector, ctx, newTestLogger())

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	assert.Eventually(t, func() bool { return provider.Count() > 0 }, time.Second, time.Millisecond*10)

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("garbage collector service did not exit after the parent context was cancelled")
	}
}

func TestGarbageCollectorServiceShutdownIsIdempotent(t *testing.T) {
	collector := middlewares.NewGarbageCollector()
	collector.Register(&testGarbageCollectorProvider{frequency: time.Millisecond * 10})

	service := NewGarbageCollector("main", collector, t.Context(), newTestLogger())

	assert.NotPanics(t, service.Shutdown)

	assert.NoError(t, service.Run())

	assert.NotPanics(t, service.Shutdown)
	assert.NotPanics(t, service.Shutdown)
}

func TestGarbageCollectorServiceDoesNotLeakGoroutines(t *testing.T) {
	collector := middlewares.NewGarbageCollector()

	for range 10 {
		collector.Register(&testGarbageCollectorProvider{frequency: time.Millisecond * 10})
	}

	runtime.GC()

	before := runtime.NumGoroutine()

	service := NewGarbageCollector("main", collector, t.Context(), newTestLogger())

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	assert.Eventually(t, func() bool { return runtime.NumGoroutine() >= before+10 }, time.Second, time.Millisecond*10)

	service.Shutdown()

	require.NoError(t, <-done)

	var after int

	for range 20 {
		if after = runtime.NumGoroutine(); after <= before {
			break
		}

		time.Sleep(time.Millisecond * 25)
	}

	assert.LessOrEqual(t, after, before)
}

func TestRunShutsDownGarbageCollectorService(t *testing.T) {
	ctrl := gomock.NewController(t)

	defer ctrl.Finish()

	userProvider := mocks.NewMockUserProvider(ctrl)
	userProvider.EXPECT().Close().Return(nil)

	storageProvider := mocks.NewMockStorage(ctrl)
	storageProvider.EXPECT().Close().Return(nil)

	provider := &testGarbageCollectorProvider{frequency: time.Millisecond * 10}

	collector := middlewares.NewGarbageCollector()
	collector.Register(provider)

	ctx, cancel := context.WithCancel(t.Context())

	sctx := newMockServiceCtx()
	sctx.ctx = ctx
	sctx.providers.GarbageCollector = collector
	sctx.providers.UserProvider = userProvider
	sctx.providers.StorageProvider = storageProvider

	done := make(chan error, 1)

	go func() {
		done <- Run(sctx, ProvisionGarbageCollector)
	}()

	assert.Eventually(t, func() bool { return provider.Count() > 0 }, time.Second, time.Millisecond*10)

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("service.Run did not return after the context was cancelled")
	}

	n := provider.Count()

	assert.Error(t, provider.Context().Err())

	time.Sleep(time.Millisecond * 50)

	assert.Equal(t, n, provider.Count(), "garbage collection continued after shutdown completed")
}

func newTestLoggerHook() (log *logrus.Entry, hook *test.Hook) {
	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)

	return logrus.NewEntry(logger), hook
}

func hasCollectionErrorEntry(hook *test.Hook) (found bool) {
	return findEntry(hook, logrus.ErrorLevel, "Error occurred performing garbage collection") != nil
}

func findEntry(hook *test.Hook, level logrus.Level, message string) (entry *logrus.Entry) {
	for _, e := range hook.AllEntries() {
		if e.Level == level && e.Message == message {
			return e
		}
	}

	return nil
}

func newTestLogger() (log *logrus.Entry) {
	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)

	return logrus.NewEntry(logger)
}

type testGarbageCollectorProvider struct {
	mu             sync.Mutex
	n              int
	ctx            context.Context
	err            error
	collect        func(ctx context.Context) (err error)
	frequency      time.Duration
	panicFrequency any
}

func (p *testGarbageCollectorProvider) GarbageCollection(ctx context.Context) (err error) {
	p.mu.Lock()

	p.n++
	p.ctx = ctx

	collect, err := p.collect, p.err

	p.mu.Unlock()

	if collect != nil {
		return collect(ctx)
	}

	return err
}

func (p *testGarbageCollectorProvider) GarbageCollectionFrequency(ctx context.Context) (frequency time.Duration) {
	if p.panicFrequency != nil {
		panic(p.panicFrequency)
	}

	return p.frequency
}

func (p *testGarbageCollectorProvider) Count() (n int) {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.n
}

func (p *testGarbageCollectorProvider) Context() (ctx context.Context) {
	p.mu.Lock()

	defer p.mu.Unlock()

	return p.ctx
}
