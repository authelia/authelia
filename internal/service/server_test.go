package service

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valyala/fasthttp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/metrics"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestNewMainServer(t *testing.T) {
	var err error

	providers := middlewares.NewProvidersBasic()

	providers.Templates, err = templates.New(templates.Config{})
	require.NoError(t, err)

	address, err := schema.NewAddress("tcp://:9091")
	require.NoError(t, err)

	config := &schema.Configuration{
		Server: schema.Server{
			Address: &schema.AddressTCP{Address: *address},
		},
	}

	ctx := &testCtx{
		Context:       context.Background(),
		Configuration: config,
		Providers:     providers,
		Logger:        logrus.NewEntry(logging.Logger()),
	}

	server, err := ProvisionServer(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, server)

	go func() {
		require.NoError(t, server.Run())
	}()

	// Give the service a moment to start.
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "main", server.ServiceName())
	assert.Equal(t, "server", server.ServiceType())
	assert.NotNil(t, server.Log())

	server.Shutdown()
}

func TestNewMetricsServer(t *testing.T) {
	var err error

	providers := middlewares.NewProvidersBasic()

	providers.Templates, err = templates.New(templates.Config{})
	require.NoError(t, err)

	providers.Metrics, err = metrics.NewPrometheus()
	require.NoError(t, err)

	address, err := schema.NewAddress("tcp://:9891/metrics")
	require.NoError(t, err)

	config := &schema.Configuration{
		Telemetry: schema.Telemetry{
			Metrics: schema.TelemetryMetrics{
				Enabled: true,
				Address: &schema.AddressTCP{Address: *address},
			},
		},
	}

	ctx := &testCtx{
		Context:       context.Background(),
		Configuration: config,
		Providers:     providers,
		Logger:        logrus.NewEntry(logging.Logger()),
	}

	server, err := ProvisionServerMetrics(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, server)

	go func() {
		require.NoError(t, server.Run())
	}()

	// Give the service a moment to start.
	time.Sleep(100 * time.Millisecond)

	assert.Equal(t, "metrics", server.ServiceName())
	assert.Equal(t, "server", server.ServiceType())
	assert.NotNil(t, server.Log())

	server.Shutdown()
}

func TestProvisionServerShouldReturnErrorWhenTheListenerCannotBeCreated(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = listener.Close()
	})

	providers := middlewares.NewProvidersBasic()

	providers.Templates, err = templates.New(templates.Config{})
	require.NoError(t, err)

	address, err := schema.NewAddress(fmt.Sprintf("tcp://%s", listener.Addr().String()))
	require.NoError(t, err)

	ctx := &testCtx{
		Context: context.Background(),
		Configuration: &schema.Configuration{
			Server: schema.Server{
				Address: &schema.AddressTCP{Address: *address},
			},
		},
		Providers: providers,
		Logger:    logrus.NewEntry(logging.Logger()),
	}

	service, err := ProvisionServer(ctx)

	assert.Nil(t, service)
	assert.ErrorContains(t, err, fmt.Sprintf("error occurred initializing main server listener for address 'tcp://%s':", listener.Addr().String()))
}

func TestProvisionServerMetricsShouldReturnNilWhenMetricsAreDisabled(t *testing.T) {
	ctx := &testCtx{
		Context:       context.Background(),
		Configuration: &schema.Configuration{},
		Providers:     middlewares.NewProvidersBasic(),
		Logger:        logrus.NewEntry(logging.Logger()),
	}

	service, err := ProvisionServerMetrics(ctx)

	assert.NoError(t, err)
	assert.Nil(t, service)
}

func TestServerRunShouldReturnErrorWhenServingFails(t *testing.T) {
	listener := &testListener{err: errors.New("failed to accept connections")}

	service := NewBaseServer("test", &fasthttp.Server{}, listener, []string{"/"}, true, logrus.NewEntry(logging.Logger()))

	assert.EqualError(t, service.Run(), "failed to accept connections")
}

func TestServerRunShouldRecoverFromPanics(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = listener.Close()
	})

	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)

	service := NewBaseServer("test", nil, listener, []string{"/"}, false, logrus.NewEntry(logger))

	assert.NoError(t, service.Run())
	assert.True(t, testLogHasEntry(hook, logrus.ErrorLevel, "Critical error caught (recovered)"))
}

type testListener struct {
	err error
}

func (l *testListener) Accept() (net.Conn, error) {
	return nil, l.err
}

func (l *testListener) Close() error {
	return nil
}

func (l *testListener) Addr() net.Addr {
	return &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9091}
}

type testCtx struct {
	Configuration *schema.Configuration
	Providers     middlewares.Providers
	Logger        *logrus.Entry

	context.Context
}

func (c *testCtx) GetConfiguration() *schema.Configuration {
	return c.Configuration
}

func (c *testCtx) GetProviders() middlewares.Providers {
	return c.Providers
}

func (c *testCtx) GetLogger() *logrus.Entry {
	return c.Logger
}
