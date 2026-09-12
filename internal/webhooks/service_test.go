// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package webhooks

import (
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func newTestService(t *testing.T) *Service {
	t.Helper()

	dispatcher := NewDispatcher(&schema.Configuration{}, nil, logrus.NewEntry(logrus.New()))

	return NewService("main", dispatcher, false, logrus.NewEntry(logrus.New()))
}

func TestServiceShutdownShouldBeIdempotent(t *testing.T) {
	service := newTestService(t)

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	assert.NotPanics(t, func() {
		service.Shutdown()
		service.Shutdown()
		service.Shutdown()
	}, "shutting a service down more than once must not close an already closed channel")

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("webhooks service did not exit after shutdown")
	}
}

func TestServiceShutdownShouldBeSafeConcurrently(t *testing.T) {
	service := newTestService(t)

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	wg := &sync.WaitGroup{}

	for range 16 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			service.Shutdown()
		}()
	}

	wg.Wait()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("webhooks service did not exit after shutdown")
	}
}
