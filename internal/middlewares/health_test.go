// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/authentication"
	"github.com/authelia/authelia/v4/internal/cache"
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/notification"
	"github.com/authelia/authelia/v4/internal/session"
	"github.com/authelia/authelia/v4/internal/storage"
)

type fakeStorage struct {
	storage.Provider

	err error
}

func (f *fakeStorage) StartupCheck() (err error) { return f.err }

type fakeUser struct {
	authentication.UserProvider

	err error
}

func (f *fakeUser) StartupCheck() (err error) { return f.err }

type fakeCache struct {
	cache.Provider

	err error
}

func (f *fakeCache) StartupCheck() (err error) { return f.err }

type fakeSessionRepository struct {
	session.Repository
}

func (f *fakeSessionRepository) StartupCheck() (err error) { return nil }

type fakeNotifier struct {
	notification.Notifier

	err error
}

func (f *fakeNotifier) StartupCheck() (err error) { return f.err }

// The session repository is backed by the storage provider unless session.storage is 'cache', so the cache provider
// must be probed directly for an outage of the cache backend to be reported.
func TestProvidersHealthChecksShouldProbeTheCacheRatherThanTheSessionRepository(t *testing.T) {
	errBroken := errors.New("could not reach the redis server")

	providers := Providers{
		Cache:             &fakeCache{err: errBroken},
		SessionRepository: &fakeSessionRepository{},
	}

	checks := providers.HealthChecks(&clock.Real{}, []string{schema.ProviderNameCache})

	require.Len(t, checks, 1)
	assert.Equal(t, schema.ProviderNameCache, checks[0].Name)
	assert.EqualError(t, checks[0].Err, errBroken.Error())
}

func TestProvidersHealthChecks(t *testing.T) {
	errBroken := errors.New("could not reach the server")

	testCases := []struct {
		name     string
		have     []string
		broken   bool
		expected []HealthCheck
	}{
		{
			"ShouldReturnNothingWhenNoProvidersAreNamed",
			nil,
			false,
			[]HealthCheck{},
		},
		{
			"ShouldProbeEachNamedProviderInOrder",
			[]string{schema.ProviderNameStorage, schema.ProviderNameUser},
			false,
			[]HealthCheck{
				{Name: schema.ProviderNameStorage},
				{Name: schema.ProviderNameUser},
			},
		},
		{
			"ShouldReportTheProviderError",
			[]string{schema.ProviderNameNotification},
			true,
			[]HealthCheck{
				{Name: schema.ProviderNameNotification, Err: errBroken},
			},
		},
		{
			"ShouldReportAnUnknownProviderRatherThanSkipIt",
			[]string{"nonexistent"},
			false,
			[]HealthCheck{
				{Name: "nonexistent", Err: ErrHealthCheckProviderUnknown},
			},
		},
		{
			"ShouldReportAProviderWhichIsNotConfigured",
			[]string{schema.ProviderNameCache},
			false,
			[]HealthCheck{
				{Name: schema.ProviderNameCache, Err: ErrHealthCheckProviderNotConfigured},
			},
		},
		{
			"ShouldKeepProbingAfterAFailure",
			[]string{schema.ProviderNameNotification, schema.ProviderNameStorage},
			true,
			[]HealthCheck{
				{Name: schema.ProviderNameNotification, Err: errBroken},
				{Name: schema.ProviderNameStorage},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var notifierErr error

			if tc.broken {
				notifierErr = errBroken
			}

			providers := Providers{
				StorageProvider: &fakeStorage{},
				UserProvider:    &fakeUser{},
				Notifier:        &fakeNotifier{err: notifierErr},
			}

			checks := providers.HealthChecks(&clock.Real{}, tc.have)

			require.Len(t, checks, len(tc.expected))

			for i, expected := range tc.expected {
				assert.Equal(t, expected.Name, checks[i].Name)

				if expected.Err == nil {
					assert.NoError(t, checks[i].Err)
				} else {
					assert.EqualError(t, checks[i].Err, expected.Err.Error())
				}
			}
		})
	}
}
