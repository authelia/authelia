// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

// Package dynamic implements the loading, validation and atomic application of the portions of the configuration which
// can be changed at runtime without restarting the application.
package dynamic

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/configuration/validator"
)

// Consumer is something which uses part of the dynamic configuration and can have it changed at runtime.
//
// Apply is only ever called with a configuration which has fully passed validation. Implementations must build all of
// the new state before making any change visible, and make that change atomically, so that in-flight work continues
// to use the previous state and new work uses the new state. If Apply returns an error it must not have changed
// anything.
type Consumer interface {
	// Name returns a short name for the Consumer used in log messages.
	Name() string

	// Apply switches the Consumer to the given configuration.
	Apply(config *schema.DynamicConfiguration) (err error)
}

// ReferencesLoader loads the data from outside the dynamic configuration that it is allowed to reference. It is called
// on every reload so changes to the static configuration which affect validation are observed.
type ReferencesLoader func() (refs validator.DynamicReferences, err error)

// NewManager creates a new Manager. The loader is used to load the references that the dynamic configuration is
// validated against, and the filters are applied to the files as they are with the static configuration.
func NewManager(paths []string, filters []configuration.BytesFilter, loader ReferencesLoader, log *logrus.Entry, consumers ...Consumer) *Manager {
	return &Manager{
		paths:     paths,
		filters:   filters,
		refs:      loader,
		log:       log,
		consumers: consumers,
	}
}

// Manager loads, validates and applies the dynamic configuration. It never applies a configuration which has errors, in
// which case the last good configuration remains in use.
type Manager struct {
	paths     []string
	filters   []configuration.BytesFilter
	refs      ReferencesLoader
	log       *logrus.Entry
	consumers []Consumer

	mu         sync.Mutex
	last       *schema.DynamicConfiguration
	generation uint64
}

// AddConsumer adds a Consumer to the Manager. It must be called before the first Reload.
func (m *Manager) AddConsumer(consumer Consumer) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.consumers = append(m.consumers, consumer)
}

// Paths returns the paths the dynamic configuration is loaded from.
func (m *Manager) Paths() []string {
	return m.paths
}

// Load loads and validates the dynamic configuration from disk without applying it. The returned warnings are
// non-fatal, and an error is returned if there are any validation errors.
func (m *Manager) Load() (config *schema.DynamicConfiguration, warnings []error, err error) {
	val := schema.NewStructValidator()

	var keys []string

	if config, keys, err = configuration.LoadDynamic(val, m.paths, m.filters); err != nil {
		return nil, nil, err
	}

	if val.HasErrors() {
		return nil, val.Warnings(), errors.Join(val.Errors()...)
	}

	validator.ValidateDynamicKeys(keys, configuration.GetMultiKeyMappedDeprecationKeys(), configuration.DefaultEnvPrefix, val)

	var refs validator.DynamicReferences

	if refs, err = m.refs(); err != nil {
		return nil, nil, fmt.Errorf("error loading the references for validation: %w", err)
	}

	validator.ValidateDynamicConfiguration(config, refs, val)

	if val.HasErrors() {
		return nil, val.Warnings(), errors.Join(val.Errors()...)
	}

	return config, val.Warnings(), nil
}

// SetCurrent records the configuration which was used at startup as the current one. It is not applied to the
// consumers as they are expected to have been created from it.
func (m *Manager) SetCurrent(config *schema.DynamicConfiguration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.last = config
	m.generation = 1
}

// Generation returns the current generation number, which starts at 1 and increases on every successful reload.
func (m *Manager) Generation() uint64 {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.generation
}

// Reload implements service.ReloadableProvider. It loads and validates the configuration, and if there are no errors
// and it differs from the one in use, switches every Consumer to it. On any error the previous configuration remains
// in effect. Concurrent calls are serialized.
func (m *Manager) Reload() (reloaded bool, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var (
		config   *schema.DynamicConfiguration
		warnings []error
	)

	if config, warnings, err = m.Load(); err != nil {
		return false, fmt.Errorf("the dynamic configuration was not reloaded and the previous configuration (generation %d) remains in use: %w", m.generation, err)
	}

	for _, warning := range warnings {
		m.log.WithError(warning).Warn("Dynamic configuration validation warning")
	}

	if reflect.DeepEqual(m.last, config) {
		return false, nil
	}

	for _, consumer := range m.consumers {
		if err = consumer.Apply(config); err != nil {
			return false, fmt.Errorf("the dynamic configuration was not reloaded as the '%s' consumer rejected it and the previous configuration (generation %d) remains in use: %w", consumer.Name(), m.generation, err)
		}
	}

	m.last = config
	m.generation++

	m.log.WithFields(logrus.Fields{"generation": m.generation, "rules": len(config.AccessControl.Rules)}).Info("Dynamic configuration applied")

	return true, nil
}
