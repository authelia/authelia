// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package suites

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConformanceRunner_AwaitPlaceholderSettledReturnsAsSoonAsTheStatusChanges(t *testing.T) {
	client, calls, closeServer := conformanceSettleServer(t, 2)
	defer closeServer()

	runner := NewConformanceRunner(client, nil, "plan1", nil)

	start := time.Now()

	status, err := runner.awaitPlaceholderSettled(context.Background(), "m1", time.Millisecond*20, time.Second*2)
	require.NoError(t, err)

	assert.Equal(t, "FINISHED", status)
	assert.Equal(t, int32(3), calls.Load(), "polls until it changes and then stops")
	assert.Less(t, time.Since(start), time.Second, "does not wait out the whole window once it has an answer")
}

func TestConformanceRunner_AwaitPlaceholderSettledChecksBeforeWaitingAtAll(t *testing.T) {
	client, calls, closeServer := conformanceSettleServer(t, 0)
	defer closeServer()

	runner := NewConformanceRunner(client, nil, "plan1", nil)

	start := time.Now()

	status, err := runner.awaitPlaceholderSettled(context.Background(), "m1", time.Second*5, time.Second*30)
	require.NoError(t, err)

	assert.Equal(t, "FINISHED", status)
	assert.Equal(t, int32(1), calls.Load())
	assert.Less(t, time.Since(start), time.Second, "a module which already moved on is not made to wait for the interval")
}

func TestConformanceRunner_AwaitPlaceholderSettledGivesUpAtTheTimeoutAndReportsWhatItSaw(t *testing.T) {
	client, calls, closeServer := conformanceSettleServer(t, 1000)
	defer closeServer()

	runner := NewConformanceRunner(client, nil, "plan1", nil)

	status, err := runner.awaitPlaceholderSettled(context.Background(), "m1", time.Millisecond*20, time.Millisecond*100)
	require.NoError(t, err)

	assert.Equal(t, conformanceStatusWaiting, status, "reports the status it last saw rather than erroring")
	assert.Greater(t, calls.Load(), int32(2), "polled repeatedly across the window")
}

func TestConformanceRunner_AwaitPlaceholderSettledStopsWhenTheContextExpires(t *testing.T) {
	client, _, closeServer := conformanceSettleServer(t, 1000)
	defer closeServer()

	runner := NewConformanceRunner(client, nil, "plan1", nil)

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
	defer cancel()

	_, err := runner.awaitPlaceholderSettled(ctx, "m1", time.Millisecond*20, time.Minute)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "m1")
	assert.Contains(t, err.Error(), "evidence image")
}

func TestConformanceUploadSettleMatchesTheIntendedCadence(t *testing.T) {
	assert.Equal(t, time.Second, conformanceUploadSettleInterval)
	assert.Equal(t, time.Second*10, conformanceUploadSettleTimeout)
}

func conformanceSettleServer(t *testing.T, waits int32) (client *ConformanceClient, calls *atomic.Int32, close func()) {
	t.Helper()

	calls = &atomic.Int32{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/info/m1", r.URL.Path)

		status := "FINISHED"

		if calls.Add(1) <= waits {
			status = "WAITING"
		}

		_, _ = fmt.Fprintf(w, `{"_id":"m1","status":%q,"result":"REVIEW"}`, status)
	}))

	client, err := NewConformanceClient(server.URL)
	require.NoError(t, err)

	return client, calls, server.Close
}
