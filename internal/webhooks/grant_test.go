// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

// grantServer answers the handshake with a configurable status and no WebHook-Allowed-Origin header, as a receiver which
// confirms by callback does, and answers deliveries with a configurable status.
type grantServer struct {
	handshake atomic.Int32
	delivery  atomic.Int32
	delivered atomic.Int32
}

func newGrantServer(t *testing.T) (*grantServer, *httptest.Server) {
	t.Helper()

	gs := &grantServer{}

	gs.handshake.Store(http.StatusOK)
	gs.delivery.Store(http.StatusOK)

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			rw.WriteHeader(int(gs.handshake.Load()))

			return
		}

		status := int(gs.delivery.Load())

		if status == http.StatusOK {
			gs.delivered.Add(1)
		}

		rw.WriteHeader(status)
	}))

	t.Cleanup(server.Close)

	return gs, server
}

// confirmOnce runs a first start which the receiver confirms using the callback, leaving the grant in the store.
func confirmOnce(t *testing.T, store *testStore, address string) {
	t.Helper()

	first := newValidationDispatcher(t, address, nil)
	first.SetStore(store)

	require.NoError(t, first.Validate())
	require.True(t, first.destinations[0].pending)
	require.True(t, first.Confirm("admin-api", first.destinations[0].key, ""))
	require.Len(t, store.grants(), 1)

	first.Shutdown()
}

func TestDispatcherShouldForgetAConfirmationWhenTheHandshakeIsAnsweredWithGone(t *testing.T) {
	gs, server := newGrantServer(t)

	store := newTestStore()

	confirmOnce(t, store, server.URL)

	gs.handshake.Store(http.StatusGone)

	second := newValidationDispatcher(t, server.URL, nil)
	second.SetStore(store)

	defer second.Shutdown()

	require.ErrorIs(t, second.Validate(), errHandshakeRetired)

	assert.True(t, second.destinations[0].pending, "a retired receiver is held until it confirms again")
	assert.False(t, second.destinations[0].refused, "a retired receiver may still confirm again using the callback")
	assert.Empty(t, store.grants())

	second.Start()

	second.Emit(context.Background(), testEvent())

	time.Sleep(time.Millisecond * 200)

	assert.Equal(t, int32(0), gs.delivered.Load())

	require.True(t, second.Confirm("admin-api", second.destinations[0].key, ""))
	assert.Len(t, store.grants(), 1)
}

func TestDispatcherShouldKeepAConfirmationWhenTheHandshakeIsNotAnsweredWithGone(t *testing.T) {
	testCases := []struct {
		name   string
		status int32
	}{
		{"ShouldKeepItForAWithheldConfirmation", http.StatusOK},
		{"ShouldKeepItForAReceiverWithoutOptions", http.StatusMethodNotAllowed},
		{"ShouldKeepItForForbidden", http.StatusForbidden},
		{"ShouldKeepItForNotFound", http.StatusNotFound},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gs, server := newGrantServer(t)

			store := newTestStore()

			confirmOnce(t, store, server.URL)

			gs.handshake.Store(tc.status)

			second := newValidationDispatcher(t, server.URL, nil)
			second.SetStore(store)

			defer second.Shutdown()

			require.NoError(t, second.Validate())

			assert.False(t, second.destinations[0].pending)
			assert.Len(t, store.grants(), 1)

			second.Start()

			second.Emit(context.Background(), testEvent())

			assert.Eventually(t, func() bool { return gs.delivered.Load() == 1 }, time.Second*5, time.Millisecond*10)
		})
	}
}

func TestDispatcherShouldNotReuseAConfirmationForAnotherOrigin(t *testing.T) {
	_, server := newGrantServer(t)

	store := newTestStore()

	confirmOnce(t, store, server.URL)

	second := newValidationDispatcher(t, server.URL, func(config *schema.WebhookDestination) {
		config.Validation.Origin, config.Validation.ForceOrigin = "other.example.com", true
	})
	second.SetStore(store)

	defer second.Shutdown()

	require.NoError(t, second.Validate())

	assert.True(t, second.destinations[0].pending)
	assert.True(t, second.promote(), "the grant must not promote a destination with another origin either")
	assert.True(t, second.destinations[0].pending)
}

func TestDispatcherShouldRecordTheOriginOfAConfirmation(t *testing.T) {
	_, server := newGrantServer(t)

	store := newTestStore()

	confirmOnce(t, store, server.URL)

	data, err := store.LoadCachedData(context.Background(), grantName("admin-api"))

	require.NoError(t, err)
	require.NotNil(t, data)

	saved := grant{}

	require.NoError(t, json.Unmarshal(data.Value, &saved))

	assert.Equal(t, "auth.example.com", saved.Origin)
}

func TestDispatcherShouldNotReuseAConfirmationWithoutAnOrigin(t *testing.T) {
	_, server := newGrantServer(t)

	store := newTestStore()

	dispatcher := newValidationDispatcher(t, server.URL, nil)
	dispatcher.SetStore(store)

	defer dispatcher.Shutdown()

	value, err := json.Marshal(grant{Name: "admin-api", Address: dispatcher.destinations[0].config.Address.String(), Confirmed: time.Now()})

	require.NoError(t, err)

	require.NoError(t, store.SaveCachedData(context.Background(), model.CachedData{Name: grantName("admin-api"), Value: value}))

	require.NoError(t, dispatcher.Validate())

	assert.True(t, dispatcher.destinations[0].pending)
}

func TestDispatcherShouldKeepAConfirmationGivenDuringAHandshakeAnsweredWithGone(t *testing.T) {
	entered, release := make(chan struct{}), make(chan struct{})

	server := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodOptions {
			close(entered)

			<-release

			rw.WriteHeader(http.StatusGone)

			return
		}

		rw.WriteHeader(http.StatusOK)
	}))

	defer server.Close()

	store := newTestStore()

	dispatcher := newValidationDispatcher(t, server.URL, nil)
	dispatcher.SetStore(store)

	defer dispatcher.Shutdown()

	validated := make(chan error, 1)

	go func() { validated <- dispatcher.Validate() }()

	<-entered

	require.True(t, dispatcher.Confirm("admin-api", dispatcher.destinations[0].key, ""))

	close(release)

	require.NoError(t, <-validated)

	assert.False(t, dispatcher.destinations[0].pending)
	assert.Len(t, store.grants(), 1, "the callback is the more recent statement of consent than the 410")
}

func TestDispatcherConfirmShouldReinstateARetiredDestination(t *testing.T) {
	gs, server := newGrantServer(t)

	store := newTestStore()

	dispatcher := newValidationDispatcher(t, server.URL, nil)
	dispatcher.SetStore(store)

	defer dispatcher.Shutdown()

	require.NoError(t, dispatcher.Validate())
	require.True(t, dispatcher.Confirm("admin-api", dispatcher.destinations[0].key, ""))

	dispatcher.Start()

	gs.delivery.Store(http.StatusGone)

	dispatcher.Emit(context.Background(), testEvent())

	require.Eventually(t, func() bool { return dispatcher.destinations[0].gone.Load() }, time.Second*5, time.Millisecond*10)
	require.Eventually(t, func() bool { return len(store.grants()) == 0 }, time.Second*5, time.Millisecond*10)

	gs.delivery.Store(http.StatusOK)

	require.True(t, dispatcher.Confirm("admin-api", dispatcher.destinations[0].key, "60"))

	assert.False(t, dispatcher.destinations[0].gone.Load())
	assert.Len(t, store.grants(), 1)

	d := dispatcher.destinations[0]

	d.mutex.Lock()
	stopped := d.revival == nil || !d.revival.Stop()
	d.mutex.Unlock()

	assert.True(t, stopped, "the scheduled recheck must be cancelled")

	dispatcher.Emit(context.Background(), testEvent())

	assert.Eventually(t, func() bool { return gs.delivered.Load() == 1 }, time.Second*5, time.Millisecond*10)
}

func TestDispatcherShouldReinstateARetiredDestinationConfirmedOnAnotherInstance(t *testing.T) {
	gs, server := newGrantServer(t)

	store := newTestStore()

	retired := newValidationDispatcher(t, server.URL, nil)
	retired.SetStore(store)

	defer retired.Shutdown()

	other := newValidationDispatcher(t, server.URL, nil)
	other.SetStore(store)

	defer other.Shutdown()

	require.NoError(t, retired.Validate())
	require.NoError(t, other.Validate())
	require.True(t, retired.Confirm("admin-api", retired.destinations[0].key, ""))
	require.True(t, other.Confirm("admin-api", other.destinations[0].key, ""))

	retired.Start()

	gs.delivery.Store(http.StatusGone)

	retired.Emit(context.Background(), testEvent())

	require.Eventually(t, func() bool { return retired.destinations[0].gone.Load() }, time.Second*5, time.Millisecond*10)
	require.Eventually(t, func() bool { return len(store.grants()) == 0 }, time.Second*5, time.Millisecond*10)

	d := retired.destinations[0]

	d.revive()

	assert.True(t, d.gone.Load(), "without a new confirmation the withheld handshake must not reinstate it")

	gs.delivery.Store(http.StatusOK)

	require.True(t, other.Confirm("admin-api", other.destinations[0].key, ""))
	require.Len(t, store.grants(), 1)

	d.revive()

	assert.False(t, d.gone.Load())
}

func TestDestinationReviveShouldWithdrawAConfirmationWhenTheHandshakeIsAnsweredWithGone(t *testing.T) {
	gs, server := newGrantServer(t)

	store := newTestStore()

	dispatcher := newValidationDispatcher(t, server.URL, nil)
	dispatcher.SetStore(store)

	defer dispatcher.Shutdown()

	require.NoError(t, dispatcher.Validate())
	require.True(t, dispatcher.Confirm("admin-api", dispatcher.destinations[0].key, ""))

	d := dispatcher.destinations[0]

	d.gone.Store(true)

	gs.handshake.Store(http.StatusGone)

	d.revive()

	assert.True(t, d.gone.Load())
	assert.Empty(t, store.grants())
}
