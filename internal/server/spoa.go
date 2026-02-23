package server

import (
	"crypto/tls"
	"net"

	"github.com/go-spop/spop/agent"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/handlers"
	"github.com/authelia/authelia/v4/internal/middlewares"
	"github.com/authelia/authelia/v4/internal/utils"
)

// NewSPOA Create Authelia's internal web server with the given configuration and providers.
func NewSPOA(config *schema.Configuration, providers middlewares.Providers) (server *agent.Agent, listener net.Listener, isTLS bool, err error) {
	if listener, err = config.Listener.Address.Listener(); err != nil {
		return nil, nil, false, err
	}

	if config.Listener.TLS != nil {
		tc := utils.NewTLSConfig(config.Listener.TLS, nil)

		listener = tls.NewListener(listener, tc)
	}

	bridge := middlewares.NewBridgeSPOPBuilder(*config, providers).Build()

	authz := handlers.NewAuthzBuilder().WithImplementationForwardAuth().Build()

	server = agent.New(bridge(func(ctx *middlewares.AutheliaSPOPCtx) {
		authz.Handler(ctx)
	}), nil)

	return server, listener, false, nil
}
