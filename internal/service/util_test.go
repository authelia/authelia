package service

import (
	"context"
	"errors"
	"os"
	"sync"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestRunShouldShutdownProvisionedServicesOnProvisionerError(t *testing.T) {
	service := newTestServiceProvider("provisioned")

	ctx, cancel := newTestServiceCtx()

	defer cancel()

	err := Run(ctx, testProvisionerOf(service), testProvisionerError(errors.New("failed to provision")))

	assert.EqualError(t, err, "error occurred provisioning services: failed to provision")
	assert.Equal(t, int32(1), service.shutdowns.Load())
}

func TestRunShouldShutdownServicesOnContextCancellation(t *testing.T) {
	service := newTestServiceProvider("example")

	ctx, cancel := newTestServiceCtx()

	defer cancel()

	testWithProviderMocks(t, ctx, nil, nil)

	errCh := make(chan error, 1)

	go func() {
		errCh <- Run(ctx, testProvisionerOf(service), testProvisionerNil())
	}()

	<-service.runs

	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second * 10):
		t.Fatal("run did not return within timeout")
	}

	assert.Equal(t, int32(1), service.shutdowns.Load())
}

func TestRunShouldShutdownWhenAServiceReturnsAnError(t *testing.T) {
	failing := newTestServiceProvider("failing")
	failing.err = errors.New("failed to run")

	service := newTestServiceProvider("example")

	ctx, cancel := newTestServiceCtx()

	defer cancel()

	testWithProviderMocks(t, ctx, nil, nil)

	assert.NoError(t, Run(ctx, testProvisionerOf(failing), testProvisionerOf(service)))

	assert.Equal(t, int32(1), failing.shutdowns.Load())
	assert.Equal(t, int32(1), service.shutdowns.Load())
}

func TestRunShouldLogErrorsClosingProviders(t *testing.T) {
	service := newTestServiceProvider("example")

	ctx, cancel := newTestServiceCtx()

	defer cancel()

	testWithProviderMocks(t, ctx, errors.New("failed to close user provider"), errors.New("failed to close storage provider"))

	errCh := make(chan error, 1)

	go func() {
		errCh <- Run(ctx, testProvisionerOf(service))
	}()

	<-service.runs

	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second * 10):
		t.Fatal("run did not return within timeout")
	}
}

func TestRunShouldShutdownOnProcessSignal(t *testing.T) {
	service := newTestServiceProvider("example")

	ctx, cancel := newTestServiceCtx()

	defer cancel()

	testWithProviderMocks(t, ctx, nil, nil)

	errCh := make(chan error, 1)

	go func() {
		errCh <- Run(ctx, testProvisionerOf(service))
	}()

	<-service.runs

	p, err := os.FindProcess(os.Getpid())
	require.NoError(t, err)
	require.NoError(t, p.Signal(syscall.SIGTERM))

	select {
	case err = <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second * 10):
		t.Fatal("run did not return within timeout")
	}

	assert.Equal(t, int32(1), service.shutdowns.Load())
}

func TestRunShouldProvisionServicesWithARuntimeContext(t *testing.T) {
	service := newTestServiceProvider("example")

	ctx, cancel := newTestServiceCtx()

	defer cancel()

	testWithProviderMocks(t, ctx, nil, nil)

	var rctx Context

	var provisioner Provisioner = func(provisioned Context) (Provider, error) {
		rctx = provisioned

		return service, nil
	}

	errCh := make(chan error, 1)

	go func() {
		errCh <- Run(ctx, provisioner)
	}()

	<-service.runs

	require.NotNil(t, rctx)

	assert.Equal(t, ctx.GetLogger(), rctx.GetLogger())
	assert.Equal(t, ctx.GetConfiguration(), rctx.GetConfiguration())
	assert.Equal(t, ctx.GetProviders(), rctx.GetProviders())
	assert.Nil(t, rctx.Value("example"))

	deadline, ok := rctx.Deadline()

	assert.Zero(t, deadline)
	assert.False(t, ok)
	assert.NoError(t, rctx.Err())

	cancel()

	select {
	case err := <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second * 10):
		t.Fatal("run did not return within timeout")
	}

	assert.ErrorIs(t, rctx.Err(), context.Canceled)

	select {
	case <-rctx.Done():
	default:
		t.Fatal("the runtime context was not done")
	}
}

func TestRunAllShouldRunTheDefaultProvisioners(t *testing.T) {
	providers := middlewares.NewProvidersBasic()

	tmpl, err := templates.New(templates.Config{})
	require.NoError(t, err)

	providers.Templates = tmpl

	address, err := schema.NewAddress("tcp://127.0.0.1:0")
	require.NoError(t, err)

	cctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	ctx := &testCtx{
		Context: cctx,
		Configuration: &schema.Configuration{
			Server: schema.Server{
				Address: &schema.AddressTCP{Address: *address},
			},
		},
		Providers: providers,
		Logger:    logrus.NewEntry(logging.Logger()),
	}

	testWithProviderMocks(t, ctx, nil, nil)

	errCh := make(chan error, 1)

	go func() {
		errCh <- RunAll(ctx)
	}()

	time.Sleep(time.Millisecond * 100)

	cancel()

	select {
	case err = <-errCh:
		assert.NoError(t, err)
	case <-time.After(time.Second * 20):
		t.Fatal("run did not return within timeout")
	}
}

func TestConnectionType(t *testing.T) {
	testCases := []struct {
		name     string
		isTLS    bool
		expected string
	}{
		{
			name:     "ShouldReturnTLS",
			isTLS:    true,
			expected: "TLS",
		},
		{
			name:     "ShouldReturnNonTLS",
			isTLS:    false,
			expected: "non-TLS",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, connectionType(tc.isTLS))
		})
	}
}

func TestRecoverErr(t *testing.T) {
	testCases := []struct {
		name     string
		have     any
		expected string
	}{
		{
			name:     "ShouldHandleNil",
			have:     nil,
			expected: "",
		},
		{
			name:     "ShouldHandleString",
			have:     "a bad thing happened",
			expected: "recovered panic: a bad thing happened",
		},
		{
			name:     "ShouldHandleError",
			have:     errors.New("a bad thing happened"),
			expected: "recovered panic: a bad thing happened",
		},
		{
			name:     "ShouldHandleUnknown",
			have:     123,
			expected: "recovered panic with unknown type: 123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := recoverErr(tc.have)

			if tc.expected == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.expected)
			}
		})
	}
}

func TestRecoverErrShouldWrapErrors(t *testing.T) {
	err := errors.New("a bad thing happened")

	assert.ErrorIs(t, recoverErr(err), err)
}

func testWithProviderMocks(t *testing.T, ctx *testCtx, errUser, errStorage error) {
	t.Helper()

	ctrl := gomock.NewController(t)

	user := mocks.NewMockUserProvider(ctrl)
	user.EXPECT().Close().Return(errUser)

	storage := mocks.NewMockStorage(ctrl)
	storage.EXPECT().Close().Return(errStorage)

	ctx.Providers.UserProvider = user
	ctx.Providers.StorageProvider = storage
}

func testProvisionerNil() Provisioner {
	return func(ctx Context) (Provider, error) {
		return nil, nil
	}
}

func newTestServiceCtx() (ctx *testCtx, cancel context.CancelFunc) {
	cctx, cancel := context.WithCancel(context.Background())

	return &testCtx{
		Context:       cctx,
		Configuration: nil,
		Providers:     middlewares.NewProvidersBasic(),
		Logger:        logrus.NewEntry(logging.Logger()),
	}, cancel
}

func testProvisionerOf(service Provider) Provisioner {
	return func(ctx Context) (Provider, error) {
		return service, nil
	}
}

func testProvisionerError(err error) Provisioner {
	return func(ctx Context) (Provider, error) {
		return nil, err
	}
}

func newTestServiceProvider(name string) *testServiceProvider {
	return &testServiceProvider{
		name: name,
		log:  logrus.NewEntry(logging.Logger()).WithField(logFieldService, "test"),
		quit: make(chan struct{}),
		runs: make(chan struct{}, 1),
	}
}

type testServiceProvider struct {
	name string
	log  *logrus.Entry
	err  error

	quit      chan struct{}
	runs      chan struct{}
	stop      sync.Once
	shutdowns atomic.Int32
}

func (s *testServiceProvider) ServiceType() string {
	return "test"
}

func (s *testServiceProvider) ServiceName() string {
	return s.name
}

func (s *testServiceProvider) Log() *logrus.Entry {
	return s.log
}

func (s *testServiceProvider) Run() (err error) {
	s.runs <- struct{}{}

	if s.err != nil {
		return s.err
	}

	<-s.quit

	return nil
}

func (s *testServiceProvider) Shutdown() {
	s.shutdowns.Add(1)

	s.stop.Do(func() {
		close(s.quit)
	})
}
