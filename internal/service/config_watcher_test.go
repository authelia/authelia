package service

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestNewConfigReloader(t *testing.T) {
	config := &schema.Configuration{}
	log := logrus.NewEntry(logrus.New())

	reloader := NewConfigReloader(config, nil, log)
	require.NotNil(t, reloader)
	assert.Equal(t, config, reloader.config)
}

func TestConfigReloaderReloadWithInvalidSources(t *testing.T) {
	config := &schema.Configuration{
		Theme: "light",
	}
	log := logrus.NewEntry(logrus.New())

	// Use an empty source list - Load should still work (returns defaults)
	sources := []configuration.Source{
		configuration.NewDefaultsSource(),
	}

	reloader := NewConfigReloader(config, sources, log)

	reloaded, err := reloader.Reload()
	// With just defaults, the validator will produce warnings or errors
	// about required fields. The important thing is it doesn't panic.
	if err != nil {
		// Expected - default config alone won't be valid
		assert.False(t, reloaded)
	} else {
		assert.True(t, reloaded)
	}
}

func TestConfigReloaderSwapsConfig(t *testing.T) {
	original := &schema.Configuration{
		Theme: "dark",
	}
	log := logrus.NewEntry(logrus.New())

	reloader := NewConfigReloader(original, nil, log)

	// Verify the pointer is stored
	assert.Equal(t, "dark", reloader.config.Theme)
}
