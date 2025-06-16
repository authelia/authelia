package service

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/logging"
	"github.com/authelia/authelia/v4/internal/metrics"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/templates"
)

func TestNewMainServer(t *testing.T) {
	tx, err := templates.New(templates.Config{})
	require.NoError(t, err)

	address, err := schema.NewAddress("tcp://:9091")
	require.NoError(t, err)

	config := &schema.Configuration{
		Server: schema.Server{
			Address: &schema.AddressTCP{Address: *address},
		},
	}

	ctx := &testContext{
		Context: t.Context(),
		config:  config,
		providers: middlewares.Providers{
			Templates: tx,
		},
		logger: logrus.NewEntry(logging.Logger()),
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
	tx, err := templates.New(templates.Config{})
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

	ctx := &testContext{
		Context: context.Background(),
		config:  config,
		providers: middlewares.Providers{
			Templates: tx,
			Metrics:   metrics.NewPrometheus(),
		},
		logger: logrus.NewEntry(logging.Logger()),
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
