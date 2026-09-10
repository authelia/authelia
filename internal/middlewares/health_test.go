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
	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/notification"
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

type fakeNotifier struct {
	notification.Notifier

	err error
}

func (f *fakeNotifier) StartupCheck() (err error) { return f.err }

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
			[]string{schema.ProviderNameSession},
			false,
			[]HealthCheck{
				{Name: schema.ProviderNameSession, Err: ErrHealthCheckProviderNotConfigured},
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

func TestHealthChecksOK(t *testing.T) {
	testCases := []struct {
		name     string
		have     []HealthCheck
		expected bool
	}{
		{"ShouldBeOKWhenEmpty", nil, true},
		{"ShouldBeOKWhenEveryCheckPassed", []HealthCheck{{Name: "storage"}, {Name: "user"}}, true},
		{"ShouldNotBeOKWhenAnyCheckFailed", []HealthCheck{{Name: "storage"}, {Name: "user", Err: errors.New("bad")}}, false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, HealthChecksOK(tc.have))
		})
	}
}
