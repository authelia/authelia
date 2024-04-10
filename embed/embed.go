package embed

import (
	"github.com/fasthttp/router"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/server"
)

// New returns a new embedded Authelia process.
func New() {
	config := Configuration{}

	x := schema.Configuration(config)

	r := router.New()

	server.RegisterOpenIDConnectRoutes(r, &x, middlewares.Providers{})
}

type embed struct {
	config    Configuration
	providers middlewares.Providers
}
