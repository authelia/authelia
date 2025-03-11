package embed

import (
	"context"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// Context is an interface used in various areas of Authelia to simplify access to important elements like the
// configuration, providers, and logger.
type Context interface {
	GetLogger() *logrus.Entry
	GetProviders() middlewares.Providers
	GetConfiguration() *schema.Configuration

	context.Context
}

type ctxEmbed struct {
	Configuration *Configuration
	Providers     Providers
	Logger        *logrus.Entry

	context.Context
}

func (c *ctxEmbed) GetConfiguration() *schema.Configuration {
	return c.Configuration.ToInternal()
}

func (c *ctxEmbed) GetProviders() middlewares.Providers {
	return c.Providers.ToInternal()
}

func (c *ctxEmbed) GetLogger() *logrus.Entry {
	return c.Logger
}

var (
	_ middlewares.Context = (*ctxEmbed)(nil)
)
