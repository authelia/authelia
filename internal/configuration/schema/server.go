package schema

import (
	"time"
)

// ServerConfig represents the configuration of the http server.
type ServerConfig struct {
	Host               string `koanf:"host"`
	Port               int    `koanf:"port"`
	Path               string `koanf:"path"`
	AssetPath          string `koanf:"asset_path"`
	DisableHealthcheck bool   `koanf:"disable_healthcheck"`

	TLS       ServerTLSConfig       `koanf:"tls"`
	Headers   ServerHeadersConfig   `koanf:"headers"`
	Endpoints ServerEndpointsConfig `koanf:"endpoints"`

	Buffers  ServerBuffers  `koanf:"buffers"`
	Timeouts ServerTimeouts `koanf:"timeouts"`
}

// ServerEndpointsConfig is the endpoints configuration for the HTTP server.
type ServerEndpointsConfig struct {
	EnablePprof   bool `koanf:"enable_pprof"`
	EnableExpvars bool `koanf:"enable_expvars"`

	Authz map[string]ServerAuthzEndpointConfig `koanf:"authz"`
}

// ServerAuthzEndpointConfig is the Authz endpoints configuration for the HTTP server.
type ServerAuthzEndpointConfig struct {
	Implementation string `koanf:"string"`

	AuthnStrategies []ServerAuthnStrategyAuthzEndpointConfig `koanf:"authn_strategies"`
}

// ServerAuthnStrategyAuthzEndpointConfig is the Authz endpoints configuration for the HTTP server.
type ServerAuthnStrategyAuthzEndpointConfig struct {
	Name string `koanf:"name"`
}

// ServerTLSConfig represents the configuration of the http servers TLS options.
type ServerTLSConfig struct {
	Certificate        string   `koanf:"certificate"`
	Key                string   `koanf:"key"`
	ClientCertificates []string `koanf:"client_certificates"`
}

// ServerHeadersConfig represents the customization of the http server headers.
type ServerHeadersConfig struct {
	CSPTemplate string `koanf:"csp_template"`
}

// DefaultServerConfig represents the default values of the ServerConfig.
var DefaultServerConfig = ServerConfig{
	Host: "0.0.0.0",
	Port: 9091,
	Buffers: ServerBuffers{
		Read:  4096,
		Write: 4096,
	},
	Timeouts: ServerTimeouts{
		Read:  time.Second * 2,
		Write: time.Second * 2,
		Idle:  time.Second * 30,
	},
	Endpoints: ServerEndpointsConfig{
		Authz: map[string]ServerAuthzEndpointConfig{
			"legacy": {
				Implementation: "Legacy",
			},
			"auth-request": {
				Implementation: "AuthRequest",
				AuthnStrategies: []ServerAuthnStrategyAuthzEndpointConfig{
					{
						Name: "HeaderAuthRequestProxyAuthorization",
					},
					{
						Name: "CookieSession",
					},
				},
			},
			"forward-auth": {
				Implementation: "ForwardAuth",
				AuthnStrategies: []ServerAuthnStrategyAuthzEndpointConfig{
					{
						Name: "HeaderProxyAuthorization",
					},
					{
						Name: "CookieSession",
					},
				},
			},
			"ext-authz": {
				Implementation: "ExtAuthz",
				AuthnStrategies: []ServerAuthnStrategyAuthzEndpointConfig{
					{
						Name: "HeaderProxyAuthorization",
					},
					{
						Name: "CookieSession",
					},
				},
			},
		},
	},
}
