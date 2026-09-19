// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/mocks"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

func TestProvisionSessionCollector(t *testing.T) {
	testCases := []struct {
		Name       string
		Repository session.Repository
		Expected   bool
	}{
		{"ShouldProvisionWithFrequency", &mockGarbageCollector{frequency: time.Minute}, true},
		{"ShouldNotProvisionWithoutFrequency", &mockGarbageCollector{}, false},
		{"ShouldNotProvisionWithoutRepository", nil, false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := newMockServiceCtx()
			ctx.providers.SessionRepository = tc.Repository

			service, err := ProvisionSessionCollector(ctx)
			require.NoError(t, err)

			if tc.Expected {
				require.NotNil(t, service)
				assert.Equal(t, serviceTypeCollector, service.ServiceType())
				assert.Equal(t, "session", service.ServiceName())
				assert.NotNil(t, service.Log())
			} else {
				assert.Nil(t, service)
			}
		})
	}
}

func TestProvisionOAuth2SessionIDCollector(t *testing.T) {
	ctrl := gomock.NewController(t)
	mockStorage := mocks.NewMockStorage(ctrl)

	config := &schema.IdentityProvidersOpenIDConnect{}

	testCases := []struct {
		Name       string
		OIDC       *schema.IdentityProvidersOpenIDConnect
		Repository session.Repository
		Storage    storage.Provider
		Expected   bool
	}{
		{"ShouldProvisionWithRepositoryAndStorage", config, &testSessionRepository{}, mockStorage, true},
		{"ShouldNotProvisionWithoutRepository", config, nil, mockStorage, false},
		{"ShouldNotProvisionWithoutStorage", config, &testSessionRepository{}, nil, false},
		{"ShouldNotProvisionWithoutOpenIDConnect", nil, &testSessionRepository{}, mockStorage, false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := newMockServiceCtx()
			ctx.config.IdentityProviders.OIDC = tc.OIDC
			ctx.providers.SessionRepository = tc.Repository
			ctx.providers.StorageProvider = tc.Storage

			service, err := ProvisionOAuth2SessionIDCollector(ctx)
			require.NoError(t, err)

			if tc.Expected {
				require.NotNil(t, service)
				assert.Equal(t, serviceTypeCollector, service.ServiceType())
				assert.Equal(t, "oauth2_session_id", service.ServiceName())
				assert.NotNil(t, service.Log())
			} else {
				assert.Nil(t, service)
			}
		})
	}
}

func TestSessionIDLookup_GetByPublicID(t *testing.T) {
	testCases := []struct {
		Name     string
		Record   session.Record
		Err      error
		Expected bool
	}{
		{"ShouldReportFoundWhenRecordExists", session.NewRecord("signature", []byte("data")), nil, true},
		{"ShouldReportNotFoundWhenRecordAbsent", nil, nil, false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			lookup := sessionIDLookup{repository: &testSessionRepository{record: tc.Record, err: tc.Err}}

			found, err := lookup.GetByPublicID(context.Background(), "https://example.com", "pid")

			require.NoError(t, err)
			assert.Equal(t, tc.Expected, found)
		})
	}

	t.Run("ShouldPropagateRepositoryError", func(t *testing.T) {
		expected := errors.New("connection refused")
		lookup := sessionIDLookup{repository: &testSessionRepository{err: expected}}

		found, err := lookup.GetByPublicID(context.Background(), "https://example.com", "pid")

		assert.ErrorIs(t, err, expected)
		assert.False(t, found)
	})
}

type testSessionRepository struct {
	session.Repository

	record session.Record
	err    error
}

func (r *testSessionRepository) GetByPublicID(_ context.Context, _, _ string) (record session.Record, err error) {
	return r.record, r.err
}

func TestNewCollector(t *testing.T) {
	testCases := []struct {
		Name      string
		Frequency time.Duration
		Expected  bool
	}{
		{"ShouldCreateWithPositiveFrequency", time.Minute, true},
		{"ShouldNotCreateWithZeroFrequency", 0, false},
		{"ShouldNotCreateWithNegativeFrequency", -time.Minute, false},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			ctx := newMockServiceCtx()

			service := NewCollector("session", &mockGarbageCollector{frequency: tc.Frequency}, ctx)

			if tc.Expected {
				require.NotNil(t, service)
				assert.Equal(t, tc.Frequency, service.frequency)
			} else {
				assert.Nil(t, service)
			}
		})
	}
}

func TestCollector_Collect(t *testing.T) {
	testCases := []struct {
		Name     string
		Err      error
		Expected int
	}{
		{"ShouldCollect", nil, 1},
		{"ShouldHandleCollectionError", errors.New("collection failed"), 1},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			collector := &mockGarbageCollector{frequency: time.Minute, err: tc.Err}

			NewCollector("session", collector, newMockServiceCtx()).collect()

			assert.Equal(t, tc.Expected, collector.Calls())
		})
	}
}

func TestCollector_CollectRecoversFromPanic(t *testing.T) {
	testCases := []struct {
		Name  string
		Panic any
	}{
		{"ShouldRecoverFromErrorPanic", errors.New("collection panicked")},
		{"ShouldRecoverFromStringPanic", "collection panicked"},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			collector := &mockGarbageCollector{frequency: time.Minute, panic: tc.Panic}

			service := NewCollector("session", collector, newMockServiceCtx())

			assert.NotPanics(t, func() {
				service.collect()
			})

			assert.Equal(t, 1, collector.Calls())
		})
	}
}

func TestCollector_RunContinuesAfterPanic(t *testing.T) {
	collector := &mockGarbageCollector{frequency: time.Millisecond, panic: errors.New("collection panicked"), notify: make(chan struct{}, 1)}

	service := NewCollector("session", collector, newMockServiceCtx())

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	for i := 0; i < 2; i++ {
		select {
		case <-collector.notify:
		case <-time.After(time.Second * 5):
			t.Fatal("timeout waiting for collection")
		}
	}

	service.Shutdown()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("timeout waiting for shutdown")
	}
}

func TestCollector_CollectSkipsWhenContextCancelled(t *testing.T) {
	collector := &mockGarbageCollector{frequency: time.Minute}

	cctx, cancel := context.WithCancel(context.Background())

	ctx := newMockServiceCtx()
	ctx.ctx = cctx

	service := NewCollector("session", collector, ctx)

	cancel()

	service.collect()

	assert.Equal(t, 0, collector.Calls())
}

func TestCollector_RunCollectsOnFrequency(t *testing.T) {
	collector := &mockGarbageCollector{frequency: time.Millisecond, notify: make(chan struct{}, 1)}

	service := NewCollector("session", collector, newMockServiceCtx())

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	select {
	case <-collector.notify:
	case <-time.After(time.Second * 5):
		t.Fatal("timeout waiting for collection")
	}

	service.Shutdown()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("timeout waiting for shutdown")
	}
}

func TestCollector_RunStopsOnContextCancellation(t *testing.T) {
	cctx, cancel := context.WithCancel(context.Background())

	ctx := newMockServiceCtx()
	ctx.ctx = cctx

	service := NewCollector("session", &mockGarbageCollector{frequency: time.Minute}, ctx)

	done := make(chan error, 1)

	go func() {
		done <- service.Run()
	}()

	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(time.Second * 5):
		t.Fatal("timeout waiting for shutdown")
	}
}

func TestCollector_ShutdownIsIdempotent(t *testing.T) {
	service := NewCollector("session", &mockGarbageCollector{frequency: time.Minute}, newMockServiceCtx())

	assert.NotPanics(t, func() {
		service.Shutdown()
		service.Shutdown()
	})
}

type mockGarbageCollector struct {
	session.Repository

	mu        sync.Mutex
	frequency time.Duration
	err       error
	panic     any
	calls     int
	notify    chan struct{}
}

func (m *mockGarbageCollector) GarbageCollectionFrequency(_ context.Context) (frequency time.Duration) {
	return m.frequency
}

func (m *mockGarbageCollector) GarbageCollection(_ context.Context) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.calls++

	if m.notify != nil {
		select {
		case m.notify <- struct{}{}:
		default:
		}
	}

	if m.panic != nil {
		panic(m.panic)
	}

	return m.err
}

func (m *mockGarbageCollector) Calls() (calls int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.calls
}
