package schema

import (
	"time"
)

// ServerConfiguration represents the configuration of the http server.
type ServerConfiguration struct {
	Host               string `koanf:"host"`
	Port               int    `koanf:"port"`
	Path               string `koanf:"path"`
	AssetPath          string `koanf:"asset_path"`
	DisableHealthcheck bool   `koanf:"disable_healthcheck"`

	TLS       ServerTLS       `koanf:"tls"`
	Headers   ServerHeaders   `koanf:"headers"`
	Endpoints ServerEndpoints `koanf:"endpoints"`

	Buffers  ServerBuffers  `koanf:"buffers"`
	Timeouts ServerTimeouts `koanf:"timeouts"`
}

// ServerEndpoints is the endpoints configuration for the HTTP server.
type ServerEndpoints struct {
	EnablePprof   bool `koanf:"enable_pprof"`
	EnableExpvars bool `koanf:"enable_expvars"`

	Authz map[string]ServerAuthzEndpoint `koanf:"authz"`
}

// ServerAuthzEndpoint is the Authz endpoints configuration for the HTTP server.
type ServerAuthzEndpoint struct {
	Implementation string `koanf:"implementation"`

	AuthnStrategies []ServerAuthnStrategyAuthzEndpoint `koanf:"authn_strategies"`
}

// ServerAuthnStrategyAuthzEndpoint is the Authz endpoints configuration for the HTTP server.
type ServerAuthnStrategyAuthzEndpoint struct {
	Name string `koanf:"name"`
}

// ServerTLS represents the configuration of the http servers TLS options.
type ServerTLS struct {
	Certificate        string   `koanf:"certificate"`
	Key                string   `koanf:"key"`
	ClientCertificates []string `koanf:"client_certificates"`
}

// ServerHeaders represents the customization of the http server headers.
type ServerHeaders struct {
	CSPTemplate string `koanf:"csp_template"`
}

// DefaultServerConfiguration represents the default values of the ServerConfiguration.
var DefaultServerConfiguration = ServerConfiguration{
	Host: "0.0.0.0",
	Port: 9091,
	Buffers: ServerBuffers{
		Read:  4096,
		Write: 4096,
	},
	Timeouts: ServerTimeouts{
		Read:  time.Second * 6,
		Write: time.Second * 6,
		Idle:  time.Second * 30,
	},
	Endpoints: ServerEndpoints{
		Authz: map[string]ServerAuthzEndpoint{
			"legacy": {
				Implementation: "Legacy",
			},
			"auth-request": {
				Implementation: "AuthRequest",
				AuthnStrategies: []ServerAuthnStrategyAuthzEndpoint{
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
				AuthnStrategies: []ServerAuthnStrategyAuthzEndpoint{
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
				AuthnStrategies: []ServerAuthnStrategyAuthzEndpoint{
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
