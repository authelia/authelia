// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/events"
)

func TestDispatcherShouldRouteByEventType(t *testing.T) {
	var (
		mu    sync.Mutex
		first []string
		other []string
	)

	serverA := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		mu.Lock()

		first = append(first, r.Header.Get(events.HeaderEventType))
		mu.Unlock()

		rw.WriteHeader(http.StatusOK)
	}))

	defer serverA.Close()

	serverB := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		mu.Lock()

		other = append(other, r.Header.Get(events.HeaderEventType))
		mu.Unlock()

		rw.WriteHeader(http.StatusOK)
	}))

	defer serverB.Close()

	addressA, err := url.Parse(serverA.URL)
	require.NoError(t, err)

	addressB, err := url.Parse(serverB.URL)
	require.NoError(t, err)

	config := &schema.Configuration{}
	config.Session.Cookies = []schema.SessionCookie{{
		Domain:      "example.com",
		AutheliaURL: &url.URL{Scheme: "https", Host: "auth.example.com"},
	}}
	config.Webhooks.Destinations = []schema.WebhookDestination{
		{
			Name: "a", Address: addressA, BufferSize: 8, Timeout: time.Second * 5,
			Events: []string{events.TypeUserPasswordChanged},
			Retry:  schema.WebhookRetry{Attempts: 1, InitialInterval: time.Millisecond, MaximumInterval: time.Millisecond},
			TLS:    &schema.TLS{},
		},
		{
			Name: "b", Address: addressB, BufferSize: 8, Timeout: time.Second * 5,
			Events: []string{"security.*"},
			Retry:  schema.WebhookRetry{Attempts: 1, InitialInterval: time.Millisecond, MaximumInterval: time.Millisecond},
			TLS:    &schema.TLS{},
		},
	}

	dispatcher := NewDispatcher(config, nil, logrus.NewEntry(logrus.New()))

	dispatcher.Start()

	dispatcher.Emit(context.Background(), events.NewEvent(&events.DataUserPassword{Type: events.TypeUserPasswordChanged}))
	dispatcher.Emit(context.Background(), events.NewEvent(&events.DataBan{Type: events.TypeSecurityBanApplied, Target: "john", TargetType: events.TargetTypeUser}))

	dispatcher.Shutdown()

	mu.Lock()
	defer mu.Unlock()

	assert.Equal(t, []string{events.TypeUserPasswordChanged}, first)
	assert.Equal(t, []string{events.TypeSecurityBanApplied}, other)
}

func TestDispatcherSourceShouldPreferTheAutheliaURL(t *testing.T) {
	config := &schema.Configuration{}
	config.Session.Cookies = []schema.SessionCookie{{
		Domain:      "example.com",
		AutheliaURL: &url.URL{Scheme: "https", Host: "auth.example.com"},
	}}

	assert.Equal(t, "https://auth.example.com", NewDispatcher(config, nil, logrus.NewEntry(logrus.New())).Source())
}

func TestDispatcherSourceShouldFallBackToTheDomain(t *testing.T) {
	config := &schema.Configuration{}
	config.Session.Cookies = []schema.SessionCookie{{Domain: "example.com"}}

	assert.Equal(t, "https://example.com", NewDispatcher(config, nil, logrus.NewEntry(logrus.New())).Source())
}

func TestDispatcherShouldNotPanicWhenEmitRacesShutdown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	address, err := url.Parse(server.URL)
	require.NoError(t, err)

	for i := 0; i < 32; i++ {
		config := &schema.Configuration{}
		config.Webhooks.Destinations = []schema.WebhookDestination{
			{
				Name: "race", Address: address, BufferSize: 4, Timeout: time.Second * 5,
				Events: []string{"*"},
				Retry:  schema.WebhookRetry{Attempts: 1, InitialInterval: time.Millisecond, MaximumInterval: time.Millisecond},
				TLS:    &schema.TLS{},
			},
		}

		dispatcher := NewDispatcher(config, nil, logrus.NewEntry(logrus.New()))

		dispatcher.Start()

		wg := &sync.WaitGroup{}

		start, stop := make(chan struct{}), make(chan struct{})

		for j := 0; j < 32; j++ {
			wg.Add(1)

			go func() {
				defer wg.Done()

				<-start

				for {
					select {
					case <-stop:
						return
					default:
					}

					dispatcher.Emit(context.Background(), events.NewEvent(&events.DataUserPassword{Type: events.TypeUserPasswordChanged}))
				}
			}()
		}

		close(start)

		runtime.Gosched()

		dispatcher.Shutdown()

		close(stop)

		wg.Wait()
	}
}

func TestDispatcherShouldIgnoreEmitWhenNotStarted(t *testing.T) {
	config := &schema.Configuration{}

	dispatcher := NewDispatcher(config, nil, logrus.NewEntry(logrus.New()))

	assert.NotPanics(t, func() {
		dispatcher.Emit(context.Background(), events.NewEvent(&events.DataUserPassword{Type: events.TypeUserPasswordChanged}))
	})
}

func TestDispatcherShouldBoundShutdownWithASlowReceiver(t *testing.T) {
	var received atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		received.Add(1)

		time.Sleep(time.Millisecond * 300)

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	address, err := url.Parse(server.URL)
	require.NoError(t, err)

	config := &schema.Configuration{}
	config.Webhooks.Destinations = []schema.WebhookDestination{
		{
			Name: "slow", Address: address, BufferSize: 64, Timeout: time.Second * 30,
			Events: []string{"*"},
			Retry:  schema.WebhookRetry{Attempts: 3, InitialInterval: time.Second, MaximumInterval: time.Second},
			TLS:    &schema.TLS{},
		},
	}

	dispatcher := NewDispatcher(config, nil, logrus.NewEntry(logrus.New()))

	dispatcher.Start()

	for i := 0; i < 64; i++ {
		dispatcher.Emit(context.Background(), events.NewEvent(&events.DataUserPassword{Type: events.TypeUserPasswordChanged}))
	}

	start := time.Now()

	dispatcher.Shutdown()

	elapsed := time.Since(start)

	assert.Less(t, elapsed, drainTimeout+time.Second*2)
	assert.Less(t, received.Load(), int32(64))
}

func TestDispatcherShouldNotRestartAfterShutdown(t *testing.T) {
	var received atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		received.Add(1)

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	address, err := url.Parse(server.URL)
	require.NoError(t, err)

	config := &schema.Configuration{}
	config.Webhooks.Destinations = []schema.WebhookDestination{
		{
			Name: "restart", Address: address, BufferSize: 8, Timeout: time.Second * 5,
			Events: []string{"*"},
			Retry:  schema.WebhookRetry{Attempts: 1, InitialInterval: time.Millisecond, MaximumInterval: time.Millisecond},
			TLS:    &schema.TLS{},
		},
	}

	dispatcher := NewDispatcher(config, nil, logrus.NewEntry(logrus.New()))

	dispatcher.Start()
	dispatcher.Shutdown()

	assert.NotPanics(t, func() {
		dispatcher.Start()

		dispatcher.Emit(context.Background(), events.NewEvent(&events.DataUserPassword{Type: events.TypeUserPasswordChanged}))

		dispatcher.Shutdown()
	})

	assert.Equal(t, int32(0), received.Load())
}

func TestDispatcherShouldNotStartAfterAShutdownWhichPrecededIt(t *testing.T) {
	config := &schema.Configuration{}
	config.Session.Cookies = []schema.SessionCookie{{Domain: "example.com"}}
	config.Webhooks.Destinations = []schema.WebhookDestination{
		{
			Name: "a", Address: &url.URL{Scheme: "https", Host: "127.0.0.1:1"}, BufferSize: 8, Timeout: time.Second,
			Events: []string{events.TypeUserPasswordChanged},
			Retry:  schema.WebhookRetry{Attempts: 1, InitialInterval: time.Millisecond, MaximumInterval: time.Millisecond},
			TLS:    &schema.TLS{},
		},
	}

	dispatcher := NewDispatcher(config, nil, logrus.NewEntry(logrus.New()))

	runtime.GC()

	before := runtime.NumGoroutine()

	// Shutting down before ever starting must still make the stopped state terminal, otherwise this Start launches a
	// worker which nothing will ever stop. The provisioning failure path shuts every service down before Run is called.
	dispatcher.Shutdown()
	dispatcher.Start()

	assert.Equal(t, before, runtime.NumGoroutine(), "Start after Shutdown must not launch a worker")

	assert.NotPanics(t, func() {
		dispatcher.Emit(context.Background(), events.NewEvent(&events.DataUserPassword{Type: events.TypeUserPasswordChanged}))
		dispatcher.Shutdown()
	})
}
