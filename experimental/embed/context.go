package embed

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/authelia/authelia/v4/experimental/embed/provider"
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// New creates a new instance of the embedded Authelia context. This can later be used with embed.ServiceRunAll.
func New(paths []string, filterNames []string) (ctx Context, val *schema.StructValidator, err error) {
	if len(paths) == 0 {
		return nil, nil, fmt.Errorf("no paths provided")
	}

	filters, err := NewNamedConfigFileFilters(filterNames...)
	if err != nil {
		return nil, nil, err
	}

	keys, config, val, err := NewConfiguration(paths, filters)
	if err != nil {
		return nil, val, err
	}

	ValidateConfigurationAndKeys(config, keys, val)

	if val.HasErrors() {
		return nil, val, fmt.Errorf("configuration validation errors")
	}

	providers, warns, errs := provider.New(config, nil, nil)

	for _, warn := range warns {
		val.PushWarning(warn)
	}

	for _, err = range errs {
		val.Push(err)
	}

	if val.HasErrors() {
		return nil, val, fmt.Errorf("provider validation errors")
	}

	ctx = &ctxEmbed{
		Configuration:      (*Configuration)(config),
		ConfigurationPaths: paths,
		Logger:             logrus.NewEntry(logrus.StandardLogger()),
		Providers:          Providers(providers),
	}

	return ctx, nil, err
}

// Context is an interface used in various areas of Authelia to simplify access to important elements like the
// configuration, providers, and logger.
type Context interface {
	GetLogger() *logrus.Entry
	GetProviders() middlewares.Providers
	GetConfiguration() *schema.Configuration
	GetConfigurationPaths() (paths []string)

	context.Context
}

type ctxEmbed struct {
	Configuration      *Configuration
	ConfigurationPaths []string

	Providers Providers
	Logger    *logrus.Entry

	context.Context
}

func (c *ctxEmbed) GetConfiguration() *schema.Configuration {
	return c.Configuration.ToInternal()
}

func (c *ctxEmbed) GetConfigurationPaths() (paths []string) {
	return c.ConfigurationPaths
}

func (c *ctxEmbed) GetProviders() middlewares.Providers {
	return c.Providers.ToInternal()
}

func (c *ctxEmbed) GetLogger() *logrus.Entry {
	return c.Logger
}

var (
	_ middlewares.Context = (*ctxEmbed)(nil)
	_ Context             = (*ctxEmbed)(nil)
)
