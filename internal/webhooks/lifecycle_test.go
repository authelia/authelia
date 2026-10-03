// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
)

func TestDestinationReviveShouldApplyTheRateOfTheRecheck(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set(events.HeaderWebhookAllowedOrigin, r.Header.Get(events.HeaderWebhookRequestOrigin))
		rw.Header().Set(events.HeaderWebhookAllowedRate, "120")
		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	d := newTestDestination(t, server.URL, nil)

	defer d.halt()

	d.origin = "auth.example.com"
	d.setRate(1)
	d.gone.Store(true)

	d.revive()

	assert.False(t, d.gone.Load())
	assert.Equal(t, time.Millisecond*500, d.spacing())
}

func TestDestinationReviveShouldApplyTheRateOfAConfirmationFromAnotherInstance(t *testing.T) {
	_, server := newGrantServer(t)

	store := newTestStore()

	dispatcher := newValidationDispatcher(t, server.URL, nil)
	dispatcher.SetStore(store)

	defer dispatcher.Shutdown()

	require.NoError(t, dispatcher.Validate())
	require.True(t, dispatcher.Confirm("admin-api", dispatcher.callbackKey(dispatcher.destinations[0]), "120"))

	d := dispatcher.destinations[0]

	d.setRate(1)
	d.gone.Store(true)

	d.revive()

	assert.False(t, d.gone.Load())
	assert.Equal(t, time.Millisecond*500, d.spacing())
}

func TestDestinationValidateShouldNotLetARequestInFlightOverrunTheBudget(t *testing.T) {
	release := make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()
	defer close(release)

	d := newTestDestination(t, server.URL, func(config *schema.WebhookDestination) {
		config.Timeout = time.Second * 30
	})

	defer d.halt()

	d.budget = time.Millisecond * 200

	start := time.Now()

	_, err := d.validate("auth.example.com")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "the handshake budget of 200ms elapsed")
	assert.Less(t, time.Since(start), time.Second*5, "the budget must cancel the request rather than wait for its timeout")
}

func TestDispatcherShouldDeliverToADestinationWithoutWaitingForASlowHandshake(t *testing.T) {
	var delivered atomic.Int32

	fast := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			rw.Header().Set(events.HeaderWebhookAllowedOrigin, "*")
		} else {
			delivered.Add(1)
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer fast.Close()

	entered, release := make(chan struct{}), make(chan struct{})

	slow := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			close(entered)

			<-release

			rw.Header().Set(events.HeaderWebhookAllowedOrigin, "*")
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer slow.Close()

	dispatcher := newDispatcherForAddresses(t, fast.URL, slow.URL)

	defer dispatcher.Shutdown()

	validated := make(chan error, 1)

	go func() { validated <- dispatcher.Validate() }()

	<-entered

	dispatcher.Emit(context.Background(), testEvent())

	assert.Eventually(t, func() bool { return delivered.Load() == 1 }, time.Second*5, time.Millisecond*10,
		"the destination which answered must receive events while the other is still answering")

	close(release)

	require.NoError(t, <-validated)

	dispatcher.Start()

	assert.True(t, dispatcher.destinations[1].launched)
}

func TestDispatcherShouldNotSendTheOriginWhenValidationIsDisabled(t *testing.T) {
	var (
		origin    atomic.Value
		delivered atomic.Int32
	)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		origin.Store(r.Header.Get(events.HeaderWebhookRequestOrigin))
		delivered.Add(1)

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	dispatcher := newValidationDispatcher(t, server.URL, func(config *schema.WebhookDestination) {
		config.Validation.Disable = true
		config.Validation.Origin = "not a valid origin"
		config.Validation.ForceOrigin = true
	})

	defer dispatcher.Shutdown()

	require.NoError(t, dispatcher.Validate())

	dispatcher.Start()

	dispatcher.Emit(context.Background(), testEvent())

	require.Eventually(t, func() bool { return delivered.Load() == 1 }, time.Second*5, time.Millisecond*10)

	assert.Equal(t, "", origin.Load())
}

func newDispatcherForAddresses(t *testing.T, addresses ...string) *Dispatcher {
	t.Helper()

	config := &schema.Configuration{}
	config.Session.Cookies = []schema.SessionCookie{{
		Domain:      "example.com",
		AutheliaURL: &url.URL{Scheme: "https", Host: "auth.example.com"},
	}}

	for i, address := range addresses {
		parsed, err := url.Parse(address)

		require.NoError(t, err)

		config.Webhooks.Destinations = append(config.Webhooks.Destinations, schema.WebhookDestination{
			Name: "destination-" + string(rune('a'+i)), Address: parsed, BufferSize: 8, Timeout: time.Second * 5,
			Events: []string{events.TypeUserPasswordChanged},
			Retry:  schema.WebhookRetry{Attempts: 1, InitialInterval: time.Millisecond, MaximumInterval: time.Millisecond},
			TLS:    &schema.TLS{},
		})
	}

	return NewDispatcher(config, nil, logrus.NewEntry(logrus.New()))
}
