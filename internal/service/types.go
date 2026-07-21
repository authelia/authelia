package service

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

type Context interface {
	GetLogger() *logrus.Entry
	GetProviders() middlewares.Providers
	GetConfiguration() *schema.Configuration

	context.Context
}

type errWatcher interface {
	error

	WatcherReloadErrorCritical() bool
}

// runContext wraps a service.Context, replacing the underlying context.Context methods (Deadline/Done/Err/Value) with
// those of the runtime context which is cancelled when shutdown is initiated. The accessor methods continue to delegate
// to the original service.Context.
type runContext struct {
	context.Context

	base Context
}

func (c *runContext) GetLogger() *logrus.Entry {
	return c.base.GetLogger()
}

func (c *runContext) GetProviders() middlewares.Providers {
	return c.base.GetProviders()
}

func (c *runContext) GetConfiguration() *schema.Configuration {
	return c.base.GetConfiguration()
}
