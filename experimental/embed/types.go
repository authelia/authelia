package embed

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

type (
	Configuration schema.Configuration
	Providers     middlewares.Providers
)

func (c *Configuration) ToInternal() *schema.Configuration {
	return (*schema.Configuration)(c)
}

func (p Providers) ToInternal() middlewares.Providers {
	return middlewares.Providers(p)
}
