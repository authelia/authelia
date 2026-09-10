// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package middlewares

import (
	"errors"
	"time"

	"github.com/authelia/authelia/v4/internal/clock"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/model"
)

var (
	// ErrHealthCheckProviderUnknown is returned when a health check names a provider which does not exist.
	ErrHealthCheckProviderUnknown = errors.New("unknown provider")

	// ErrHealthCheckProviderNotConfigured is returned when a health check names a provider which exists but which
	// this instance has not configured.
	ErrHealthCheckProviderNotConfigured = errors.New("provider is not configured")
)

// HealthCheck is the outcome of probing a single provider.
type HealthCheck struct {
	Name string
	Err  error
	Took time.Duration
}

// HealthChecks probes each named provider in the order given and returns one result per name. A provider which
// fails does not stop the others being probed: the caller wants the whole picture, not the first problem.
func (p *Providers) HealthChecks(clk clock.Provider, names []string) (checks []HealthCheck) {
	checks = make([]HealthCheck, 0, len(names))

	for _, name := range names {
		provider, err := p.healthCheckProvider(name)
		if err != nil {
			checks = append(checks, HealthCheck{Name: name, Err: err})

			continue
		}

		start := clk.Now()
		err = provider.StartupCheck()

		checks = append(checks, HealthCheck{Name: name, Err: err, Took: clk.Now().Sub(start)})
	}

	return checks
}

func (p *Providers) healthCheckProvider(name string) (provider model.StartupCheck, err error) {
	switch name {
	case schema.ProviderNameStorage:
		if p.StorageProvider == nil {
			return nil, ErrHealthCheckProviderNotConfigured
		}

		return p.StorageProvider, nil
	case schema.ProviderNameSession:
		if p.SessionProvider == nil {
			return nil, ErrHealthCheckProviderNotConfigured
		}

		return p.SessionProvider, nil
	case schema.ProviderNameUser:
		if p.UserProvider == nil {
			return nil, ErrHealthCheckProviderNotConfigured
		}

		return p.UserProvider, nil
	case schema.ProviderNameNotification:
		if p.Notifier == nil {
			return nil, ErrHealthCheckProviderNotConfigured
		}

		return p.Notifier, nil
	case schema.ProviderNameNTP:
		if p.NTP == nil {
			return nil, ErrHealthCheckProviderNotConfigured
		}

		return p.NTP, nil
	case schema.ProviderNameExpressions:
		if p.UserAttributeResolver == nil {
			return nil, ErrHealthCheckProviderNotConfigured
		}

		return p.UserAttributeResolver, nil
	case schema.ProviderNameWebAuthnMetaData:
		if p.MetaDataService == nil {
			return nil, ErrHealthCheckProviderNotConfigured
		}

		return p.MetaDataService, nil
	default:
		return nil, ErrHealthCheckProviderUnknown
	}
}

// HealthChecksOK reports whether every check passed.
func HealthChecksOK(checks []HealthCheck) (ok bool) {
	for _, check := range checks {
		if check.Err != nil {
			return false
		}
	}

	return true
}
