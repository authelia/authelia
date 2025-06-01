package service

import (
	"context"
	"testing"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// testContext implements service.Context for testing.
type testContext struct {
	context.Context

	config    *schema.Configuration
	logger    *logrus.Entry
	providers middlewares.Providers
	paths     []string
}

func (m *testContext) GetLogger() *logrus.Entry {
	return m.logger
}

func (m *testContext) GetProviders() middlewares.Providers {
	return m.providers
}

func (m *testContext) GetConfiguration() *schema.Configuration {
	return m.config
}

func (m *testContext) GetConfigurationPaths() []string {
	return m.paths
}

func newMockServiceCtx(t *testing.T) *testContext {
	logger := logrus.New()
	logger.SetLevel(logrus.TraceLevel)

	config := &schema.Configuration{}

	return &testContext{
		Context:   t.Context(),
		config:    config,
		logger:    logrus.NewEntry(logger),
		providers: middlewares.Providers{},
	}
}
