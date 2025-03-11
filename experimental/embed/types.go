package embed

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
)

// Configuration is a type alias for the internal schema.Configuration type. It allows manually configuring Authelia
// and transitioning to the internal implementation.
type Configuration schema.Configuration

// ToInternal converts this Configuration struct into a *schema.Configuration struct using a type cast.
func (c *Configuration) ToInternal() *schema.Configuration {
	return (*schema.Configuration)(c)
}

// Providers is a type alias for the internal middlewares.Providers type. It allows manually performing setup of the
// various Authelia providers and transitioning to the internal implementation.
type Providers middlewares.Providers

// ToInternal converts this Providers struct into a middlewares.Providers struct using a type cast.
func (p Providers) ToInternal() middlewares.Providers {
	return middlewares.Providers(p)
}
