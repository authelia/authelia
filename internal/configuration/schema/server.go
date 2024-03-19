package schema

import (
	"net/url"
	"time"
)

// Server represents the configuration of the http server.
type Server struct {
	Address            *AddressTCP `koanf:"address" json:"address" jsonschema:"default=tcp://:9091/,title=Address" jsonschema_description:"The address to listen on."`
	AssetPath          string      `koanf:"asset_path" json:"asset_path" jsonschema:"title=Asset Path" jsonschema_description:"The directory where the server asset overrides reside."`
	DisableHealthcheck bool        `koanf:"disable_healthcheck" json:"disable_healthcheck" jsonschema:"default=false,title=Disable Healthcheck" jsonschema_description:"Disables the healthcheck functionality."`

	TLS       ServerTLS       `koanf:"tls" json:"tls" jsonschema:"title=TLS" jsonschema_description:"The server TLS configuration."`
	Headers   ServerHeaders   `koanf:"headers" json:"headers" jsonschema:"title=Headers" jsonschema_description:"The server headers configuration."`
	Endpoints ServerEndpoints `koanf:"endpoints" json:"endpoints" jsonschema:"title=Endpoints" jsonschema_description:"The server endpoints configuration."`

	Buffers  ServerBuffers  `koanf:"buffers" json:"buffers" jsonschema:"title=Buffers" jsonschema_description:"The server buffers configuration."`
	Timeouts ServerTimeouts `koanf:"timeouts" json:"timeouts" jsonschema:"title=Timeouts" jsonschema_description:"The server timeouts configuration."`
}

// ServerEndpoints is the endpoints configuration for the HTTP server.
type ServerEndpoints struct {
	EnablePprof   bool `koanf:"enable_pprof" json:"enable_pprof" jsonschema:"default=false,title=Enable PProf" jsonschema_description:"Enables the developer specific pprof endpoints which should not be used in production and only used for debugging purposes."`
	EnableExpvars bool `koanf:"enable_expvars" json:"enable_expvars" jsonschema:"default=false,title=Enable ExpVars" jsonschema_description:"Enables the developer specific ExpVars endpoints which should not be used in production and only used for debugging purposes."`

	Authz map[string]ServerEndpointsAuthz `koanf:"authz" json:"authz" jsonschema:"title=Authz" jsonschema_description:"Configures the Authorization endpoints."`
}

// ServerEndpointsAuthz is the Authz endpoints configuration for the HTTP server.
type ServerEndpointsAuthz struct {
	Implementation string `koanf:"implementation" json:"implementation" jsonschema:"enum=ForwardAuth,enum=AuthRequest,enum=ExtAuthz,enum=Legacy,title=Implementation" jsonschema_description:"The specific Authorization implementation to use for this endpoint."`

	AuthnStrategies []ServerEndpointsAuthzAuthnStrategy `koanf:"authn_strategies" json:"authn_strategies" jsonschema:"title=Authn Strategies" jsonschema_description:"The specific Authorization strategies to use for this endpoint."`
}

// ServerEndpointsAuthzAuthnStrategy is the Authz endpoints configuration for the HTTP server.
type ServerEndpointsAuthzAuthnStrategy struct {
	Name    string   `koanf:"name" json:"name" jsonschema:"enum=HeaderAuthorization,enum=HeaderProxyAuthorization,enum=HeaderAuthRequestProxyAuthorization,enum=HeaderLegacy,enum=CookieSession,title=Name" jsonschema_description:"The name of the Authorization strategy to use."`
	Schemes []string `koanf:"schemes" json:"schemes" jsonschema:"enum=basic,enum=bearer,default=basic,title=Authorization Schemes" jsonschema_description:"The name of the authorization schemes to allow with the header strategies."`
}

// ServerTLS represents the configuration of the http servers TLS options.
type ServerTLS struct {
	Certificate        string   `koanf:"certificate" json:"certificate" jsonschema:"title=Certificate" jsonschema_description:"Path to the Certificate."`
	Key                string   `koanf:"key" json:"key" jsonschema:"title=Key" jsonschema_description:"Path to the Private Key."`
	ClientCertificates []string `koanf:"client_certificates" json:"client_certificates" jsonschema:"uniqueItems,title=Client Certificates" jsonschema_description:"Path to the Client Certificates to trust for mTLS."`
}

// ServerHeaders represents the customization of the http server headers.
type ServerHeaders struct {
	CSPTemplate CSPTemplate `koanf:"csp_template" json:"csp_template" jsonschema:"title=CSP Template" jsonschema_description:"The Content Security Policy template."`
}

// DefaultServerConfiguration represents the default values of the Server.
var DefaultServerConfiguration = Server{
	Address: &AddressTCP{Address{true, false, -1, 9091, &url.URL{Scheme: AddressSchemeTCP, Host: ":9091", Path: "/"}}},
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
		Authz: map[string]ServerEndpointsAuthz{
			AuthzEndpointNameLegacy: {
				Implementation: AuthzImplementationLegacy,
				AuthnStrategies: []ServerEndpointsAuthzAuthnStrategy{
					{
						Name: AuthzStrategyHeaderLegacy,
					},
					{
						Name: AuthzStrategyHeaderCookieSession,
					},
				},
			},
			AuthzEndpointNameAuthRequest: {
				Implementation: AuthzImplementationAuthRequest,
				AuthnStrategies: []ServerEndpointsAuthzAuthnStrategy{
					{
						Name:    AuthzStrategyHeaderAuthorization,
						Schemes: []string{SchemeBasic},
					},
					{
						Name: AuthzStrategyHeaderCookieSession,
					},
				},
			},
			AuthzEndpointNameExtAuthz: {
				Implementation: AuthzImplementationExtAuthz,
				AuthnStrategies: []ServerEndpointsAuthzAuthnStrategy{
					{
						Name:    AuthzStrategyHeaderAuthorization,
						Schemes: []string{SchemeBasic},
					},
					{
						Name: AuthzStrategyHeaderCookieSession,
					},
				},
			},
			AuthzEndpointNameForwardAuth: {
				Implementation: AuthzImplementationForwardAuth,
				AuthnStrategies: []ServerEndpointsAuthzAuthnStrategy{
					{
						Name:    AuthzStrategyHeaderAuthorization,
						Schemes: []string{SchemeBasic},
					},
					{
						Name: AuthzStrategyHeaderCookieSession,
					},
				},
			},
		},
	},
}
