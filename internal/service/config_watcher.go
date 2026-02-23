package service

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ConfigReloader implements ReloadableProvider for configuration hot-reloading.
// When triggered, it re-reads config from file sources, validates it, and
// atomically swaps the active configuration if the new config is valid.
type ConfigReloader struct {
	config  *schema.Configuration
	sources []configuration.Source
	log     *logrus.Entry
}

// NewConfigReloader creates a new ConfigReloader.
func NewConfigReloader(config *schema.Configuration, sources []configuration.Source, log *logrus.Entry) *ConfigReloader {
	return &ConfigReloader{
		config:  config,
		sources: sources,
		log:     log,
	}
}

// Reload re-reads and validates the configuration. If valid, the active
// configuration is updated in-place. Returns (true, nil) on success,
// (false, err) if validation fails.
func (r *ConfigReloader) Reload() (reloaded bool, err error) {
	val := schema.NewStructValidator()

	var newConfig *schema.Configuration

	if _, newConfig, err = configuration.Load(val, r.sources...); err != nil {
		return false, fmt.Errorf("failed to load configuration: %w", err)
	}

	if val.HasErrors() {
		errs := val.Errors()
		for _, e := range errs {
			r.log.WithError(e).Error("Configuration validation error")
		}

		return false, fmt.Errorf("configuration has %d validation error(s), keeping current configuration", len(errs))
	}

	if val.HasWarnings() {
		for _, w := range val.Warnings() {
			r.log.WithError(w).Warn("Configuration validation warning")
		}
	}

	// Atomically swap the configuration by copying all fields.
	*r.config = *newConfig

	r.log.Info("Configuration reloaded successfully")

	return true, nil
}

// ProvisionConfigWatcher creates a file watcher service for configuration files.
func ProvisionConfigWatcher(ctx Context) (service Provider, err error) {
	config := ctx.GetConfiguration()
	log := ctx.GetLogger()

	// Try to get config files from the context.
	ctxFiles, ok := ctx.(interface{ GetConfigurationFiles() []string })
	if !ok {
		return nil, nil
	}

	files := ctxFiles.GetConfigurationFiles()
	if len(files) == 0 {
		return nil, nil
	}

	// Get config sources from context.
	ctxSources, ok := ctx.(interface{ GetConfigurationSources() []configuration.Source })
	if !ok {
		return nil, nil
	}

	sources := ctxSources.GetConfigurationSources()

	reloader := NewConfigReloader(config, sources, log.WithField(logFieldService, "config-reloader"))

	// Watch the first config file for changes.
	if service, err = NewFileWatcher("configuration", files[0], reloader, log); err != nil {
		return nil, err
	}

	return service, nil
}
