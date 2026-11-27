// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"net/url"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
	"github.com/authelia/authelia/v4/internal/webhooks"
)

func TestProvisionWebhooks(t *testing.T) {
	testCases := []struct {
		name     string
		emitter  events.Emitter
		expected bool
	}{
		{
			"ShouldProvisionWithDispatcher",
			webhooks.NewDispatcher(&schema.Configuration{}, nil, logrus.NewEntry(logrus.New())),
			true,
		},
		{
			"ShouldNotProvisionWithNoOpEmitter",
			events.NewNoOpEmitter(),
			false,
		},
		{
			"ShouldNotProvisionWithoutAnEmitter",
			nil,
			false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := newMockServiceCtx()
			ctx.providers.Events = tc.emitter

			service, err := ProvisionWebhooks(ctx)

			require.NoError(t, err)

			if !tc.expected {
				assert.Nil(t, service)

				return
			}

			require.NotNil(t, service)

			assert.Equal(t, serviceTypeWebhooks, service.ServiceType())
			assert.Equal(t, "main", service.ServiceName())
			assert.NotNil(t, service.Log())
		})
	}
}

func TestWebhooksServiceShouldRunUntilShutdown(t *testing.T) {
	ctx := newMockServiceCtx()
	ctx.providers.Events = webhooks.NewDispatcher(&schema.Configuration{}, nil, logrus.NewEntry(logrus.New()))

	service, err := ProvisionWebhooks(ctx)

	require.NoError(t, err)
	require.NotNil(t, service)

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	service.Shutdown()

	select {
	case err = <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("webhooks service did not exit after shutdown")
	}
}

func TestWebhooksServiceShouldNotFailStartupWhenTheStartupCheckFails(t *testing.T) {
	config := &schema.Configuration{}
	config.Webhooks.StartupCheck = true
	config.Webhooks.Destinations = []schema.WebhookDestination{
		{
			Name:       "unreachable",
			Address:    &url.URL{Scheme: "https", Host: "127.0.0.1:1"},
			Events:     []string{events.TypeUserPasswordChanged},
			Timeout:    time.Millisecond * 100,
			BufferSize: 1,
			Retry:      schema.WebhookRetry{Attempts: 1, InitialInterval: time.Millisecond, MaximumInterval: time.Millisecond},
			TLS:        &schema.TLS{},
		},
	}

	ctx := newMockServiceCtx()
	ctx.config = config
	ctx.providers.Events = webhooks.NewDispatcher(config, nil, logrus.NewEntry(logrus.New()))

	service, err := ProvisionWebhooks(ctx)

	require.NoError(t, err)
	require.NotNil(t, service)

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	service.Shutdown()

	select {
	case err = <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("webhooks service did not exit after shutdown")
	}
}
