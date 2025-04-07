package schema

import (
	"net/url"
	"time"
)

// Server represents the configuration of the http server.
type Server struct {
	Address            *AddressTCP `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty" jsonschema:"default=tcp://:9091/,title=Address" jsonschema_description:"The address to listen on."`
	AssetPath          string      `koanf:"asset_path" yaml:"asset_path,omitempty" toml:"asset_path,omitempty" json:"asset_path,omitempty" jsonschema:"title=Asset Path" jsonschema_description:"The directory where the server asset overrides reside."`
	DisableHealthcheck bool        `koanf:"disable_healthcheck" yaml:"disable_healthcheck" toml:"disable_healthcheck" json:"disable_healthcheck" jsonschema:"default=false,title=Disable Healthcheck" jsonschema_description:"Disables the healthcheck functionality."`

	TLS       ServerTLS       `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty" jsonschema:"title=TLS" jsonschema_description:"The server TLS configuration."`
	Headers   ServerHeaders   `koanf:"headers" yaml:"headers,omitempty" toml:"headers,omitempty" json:"headers,omitempty" jsonschema:"title=Headers" jsonschema_description:"The server headers configuration."`
	Endpoints ServerEndpoints `koanf:"endpoints" yaml:"endpoints,omitempty" toml:"endpoints,omitempty" json:"endpoints,omitempty" jsonschema:"title=Endpoints" jsonschema_description:"The server endpoints configuration."`

	Buffers  ServerBuffers  `koanf:"buffers" yaml:"buffers,omitempty" toml:"buffers,omitempty" json:"buffers,omitempty" jsonschema:"title=Buffers" jsonschema_description:"The server buffers configuration."`
	Timeouts ServerTimeouts `koanf:"timeouts" yaml:"timeouts,omitempty" toml:"timeouts,omitempty" json:"timeouts,omitempty" jsonschema:"title=Timeouts" jsonschema_description:"The server timeouts configuration."`
}

// ServerEndpoints is the endpoints configuration for the HTTP server.
type ServerEndpoints struct {
	EnablePprof   bool `koanf:"enable_pprof" yaml:"enable_pprof" toml:"enable_pprof" json:"enable_pprof" jsonschema:"default=false,title=Enable PProf" jsonschema_description:"Enables the developer specific pprof endpoints which should not be used in production and only used for debugging purposes."`
	EnableExpvars bool `koanf:"enable_expvars" yaml:"enable_expvars" toml:"enable_expvars" json:"enable_expvars" jsonschema:"default=false,title=Enable ExpVars" jsonschema_description:"Enables the developer specific ExpVars endpoints which should not be used in production and only used for debugging purposes."`

	RateLimits ServerEndpointRateLimits `koanf:"rate_limits" yaml:"rate_limits,omitempty" toml:"rate_limits,omitempty" json:"rate_limits,omitempty"`

	Authz map[string]ServerEndpointsAuthz `koanf:"authz" yaml:"authz,omitempty" toml:"authz,omitempty" json:"authz,omitempty" jsonschema:"title=Authz" jsonschema_description:"Configures the Authorization endpoints."`
}

// ServerEndpointsAuthz is the Authz endpoints configuration for the HTTP server.
type ServerEndpointsAuthz struct {
	Implementation string `koanf:"implementation" yaml:"implementation,omitempty" toml:"implementation,omitempty" json:"implementation,omitempty" jsonschema:"enum=ForwardAuth,enum=AuthRequest,enum=ExtAuthz,enum=Legacy,title=Implementation" jsonschema_description:"The specific Authorization implementation to use for this endpoint."`

	AuthnStrategies []ServerEndpointsAuthzAuthnStrategy `koanf:"authn_strategies" yaml:"authn_strategies,omitempty" toml:"authn_strategies,omitempty" json:"authn_strategies,omitempty" jsonschema:"title=Authn Strategies" jsonschema_description:"The specific Authorization strategies to use for this endpoint."`
}

// ServerEndpointsAuthzAuthnStrategy is the Authz endpoints configuration for the HTTP server.
type ServerEndpointsAuthzAuthnStrategy struct {
	Name                     string        `koanf:"name" yaml:"name,omitempty" toml:"name,omitempty" json:"name,omitempty" jsonschema:"enum=HeaderAuthorization,enum=HeaderProxyAuthorization,enum=HeaderAuthRequestProxyAuthorization,enum=HeaderLegacy,enum=CookieSession,title=Name" jsonschema_description:"The name of the Authorization strategy to use."`
	Schemes                  []string      `koanf:"schemes" yaml:"schemes,omitempty" toml:"schemes,omitempty" json:"schemes,omitempty" jsonschema:"enum=basic,enum=bearer,default=basic,title=Authorization Schemes" jsonschema_description:"The name of the authorization schemes to allow with the header strategies."`
	SchemeBasicCacheLifespan time.Duration `koanf:"scheme_basic_cache_lifespan" yaml:"scheme_basic_cache_lifespan,omitempty" toml:"scheme_basic_cache_lifespan,omitempty" json:"scheme_basic_cache_lifespan,omitempty" jsonschema:"default=0,title=Scheme Basic Cache Lifespan" jsonschema_description:"The lifespan for cached basic scheme authorization attempts."`
}

// ServerTLS represents the configuration of the http servers TLS options.
type ServerTLS struct {
	Certificate        string   `koanf:"certificate" yaml:"certificate,omitempty" toml:"certificate,omitempty" json:"certificate,omitempty" jsonschema:"title=Certificate" jsonschema_description:"Path to the Certificate."`
	Key                string   `koanf:"key" yaml:"key,omitempty" toml:"key,omitempty" json:"key,omitempty" jsonschema:"title=Key" jsonschema_description:"Path to the Private Key."`
	ClientCertificates []string `koanf:"client_certificates" yaml:"client_certificates,omitempty" toml:"client_certificates,omitempty" json:"client_certificates,omitempty" jsonschema:"uniqueItems,title=Client Certificates" jsonschema_description:"Path to the Client Certificates to trust for mTLS."`
}

// ServerHeaders represents the customization of the http server headers.
type ServerHeaders struct {
	CSPTemplate CSPTemplate `koanf:"csp_template" yaml:"csp_template,omitempty" toml:"csp_template,omitempty" json:"csp_template,omitempty" jsonschema:"title=CSP Template" jsonschema_description:"The Content Security Policy template."`
}

type ServerEndpointRateLimits struct {
	ResetPasswordStart     ServerEndpointRateLimit `koanf:"reset_password_start" yaml:"reset_password_start,omitempty" toml:"reset_password_start,omitempty" json:"reset_password_start,omitempty"`
	ResetPasswordFinish    ServerEndpointRateLimit `koanf:"reset_password_finish" yaml:"reset_password_finish,omitempty" toml:"reset_password_finish,omitempty" json:"reset_password_finish,omitempty"`
	SecondFactorTOTP       ServerEndpointRateLimit `koanf:"second_factor_totp" yaml:"second_factor_totp,omitempty" toml:"second_factor_totp,omitempty" json:"second_factor_totp,omitempty"`
	SecondFactorDuo        ServerEndpointRateLimit `koanf:"second_factor_duo" yaml:"second_factor_duo,omitempty" toml:"second_factor_duo,omitempty" json:"second_factor_duo,omitempty"`
	SessionElevationStart  ServerEndpointRateLimit `koanf:"session_elevation_start" yaml:"session_elevation_start,omitempty" toml:"session_elevation_start,omitempty" json:"session_elevation_start,omitempty"`
	SessionElevationFinish ServerEndpointRateLimit `koanf:"session_elevation_finish" yaml:"session_elevation_finish,omitempty" toml:"session_elevation_finish,omitempty" json:"session_elevation_finish,omitempty"`
}

type ServerEndpointRateLimit struct {
	Enable  bool                            `koanf:"enable" yaml:"enable" toml:"enable" json:"enable"`
	Buckets []ServerEndpointRateLimitBucket `koanf:"buckets" yaml:"buckets,omitempty" toml:"buckets,omitempty" json:"buckets,omitempty"`
}

type ServerEndpointRateLimitBucket struct {
	Period   time.Duration `koanf:"period" yaml:"period,omitempty" toml:"period,omitempty" json:"period,omitempty" jsonschema:"" jsonschema_description:"The period of time this rate limit bucket applies to."`
	Requests int           `koanf:"requests" yaml:"requests" toml:"requests" json:"requests" jsonschema:"" jsonschema_description:"The number of requests allowed in this rate limit bucket for the configured period before the rate limit kicks in."`
}

// DefaultServerConfiguration represents the default values of the Server.
var DefaultServerConfiguration = Server{
	Address: &AddressTCP{Address{true, false, -1, 9091, nil, &url.URL{Scheme: AddressSchemeTCP, Host: ":9091", Path: "/"}}},
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
		RateLimits: ServerEndpointRateLimits{
			ResetPasswordStart: ServerEndpointRateLimit{
				Buckets: []ServerEndpointRateLimitBucket{
					{Period: 10 * time.Minute, Requests: 5},
					{Period: 15 * time.Minute, Requests: 10},
					{Period: 30 * time.Minute, Requests: 15},
				},
			},
			ResetPasswordFinish: ServerEndpointRateLimit{
				Buckets: []ServerEndpointRateLimitBucket{
					{Period: 1 * time.Minute, Requests: 10},
					{Period: 2 * time.Minute, Requests: 15},
				},
			},
			SecondFactorTOTP: ServerEndpointRateLimit{
				Buckets: []ServerEndpointRateLimitBucket{
					{Period: 1 * time.Minute, Requests: 30},
					{Period: 2 * time.Minute, Requests: 40},
					{Period: 10 * time.Minute, Requests: 50},
				},
			},
			SecondFactorDuo: ServerEndpointRateLimit{
				Buckets: []ServerEndpointRateLimitBucket{
					{Period: 1 * time.Minute, Requests: 10},
					{Period: 2 * time.Minute, Requests: 15},
				},
			},
			SessionElevationStart: ServerEndpointRateLimit{
				Buckets: []ServerEndpointRateLimitBucket{
					{Period: 1, Requests: 3},   // 3 requests per 1.0x of identity_validation.elevated_session.code_lifespan.
					{Period: 2, Requests: 5},   // 5 requests per 2.0x of identity_validation.elevated_session.code_lifespan.
					{Period: 12, Requests: 15}, // 15 requests per 12.0x of identity_validation.elevated_session.code_lifespan.
				},
			},
			SessionElevationFinish: ServerEndpointRateLimit{
				Buckets: []ServerEndpointRateLimitBucket{
					{Period: 1, Requests: 3},  // 3 requests per 1.0x of identity_validation.elevated_session.elevation_lifespan.
					{Period: 2, Requests: 5},  // 5 requests per 2.0x of identity_validation.elevated_session.elevation_lifespan.
					{Period: 6, Requests: 15}, // 15 requests per 6.0x of identity_validation.elevated_session.elevation_lifespan.
				},
			},
		},
	},
}
