package schema

import (
	"crypto/tls"
	"net/url"
	"time"
)

// Session represents the configuration related to user sessions.
type Session struct {
	SessionCookieCommon `koanf:",squash"`

	Secret string `koanf:"secret" json:"secret" jsonschema:"title=Secret" jsonschema_description:"Secret used to encrypt the session data."`

	Cookies []SessionCookie `koanf:"cookies" json:"cookies" jsonschema:"title=Cookies" jsonschema_description:"List of cookie domain configurations."`

	Redis *SessionRedis `koanf:"redis" json:"redis" jsonschema:"title=Redis" jsonschema_description:"Redis Session Provider configuration."`

	// Deprecated: Use the session cookies option with the same name instead.
	Domain string `koanf:"domain" json:"domain" jsonschema:"deprecated,title=Domain"`
}

type SessionCookieCommon struct {
	Name       string        `koanf:"name" json:"name" jsonschema:"default=authelia_session" jsonschema_description:"The session cookie name."`
	SameSite   string        `koanf:"same_site" json:"same_site" jsonschema:"default=lax,enum=lax,enum=strict,enum=none" jsonschema_description:"The session cookie same site value."`
	Expiration time.Duration `koanf:"expiration" json:"expiration" jsonschema:"default=1 hour" jsonschema_description:"The session cookie expiration when remember me is not checked."`
	Inactivity time.Duration `koanf:"inactivity" json:"inactivity" jsonschema:"default=5 minutes" jsonschema_description:"The session inactivity timeout."`
	RememberMe time.Duration `koanf:"remember_me" json:"remember_me" jsonschema:"default=30 days" jsonschema_description:"The session cookie expiration when remember me is checked."`

	DisableRememberMe bool `json:"-"`
}

// SessionCookie represents the configuration for a cookie domain.
type SessionCookie struct {
	SessionCookieCommon `koanf:",squash"`

	Domain                string   `koanf:"domain" json:"domain" jsonschema:"format=hostname,title=Domain" jsonschema_description:"The domain for this session cookie configuration."`
	AutheliaURL           *url.URL `koanf:"authelia_url" json:"authelia_url" jsonschema:"format=uri,title=Authelia URL" jsonschema_description:"The Root Authelia URL to redirect users to for this session cookie configuration."`
	DefaultRedirectionURL *url.URL `koanf:"default_redirection_url" json:"default_redirection_url" jsonschema:"format=uri,title=Default Redirection URL" jsonschema_description:"The default redirection URL for this session cookie configuration."`

	Legacy bool `json:"-"`
}

// SessionRedis represents the configuration related to redis session store.
type SessionRedis struct {
	Host                     string `koanf:"host" json:"host" jsonschema:"title=Host" jsonschema_description:"The redis server host."`
	Port                     int    `koanf:"port" json:"port" jsonschema:"default=6379,title=Host" jsonschema_description:"The redis server port."`
	Username                 string `koanf:"username" json:"username" jsonschema:"title=Username" jsonschema_description:"The redis username."`
	Password                 string `koanf:"password" json:"password" jsonschema:"title=Password" jsonschema_description:"The redis password."`
	DatabaseIndex            int    `koanf:"database_index" json:"database_index" jsonschema:"default=0,title=Database Index" jsonschema_description:"The redis database index."`
	MaximumActiveConnections int    `koanf:"maximum_active_connections" json:"maximum_active_connections" jsonschema:"default=8,title=Maximum Active Connections" jsonschema_description:"The maximum connections that can be made to redis at one time."`
	MinimumIdleConnections   int    `koanf:"minimum_idle_connections" json:"minimum_idle_connections" jsonschema:"title=Minimum Idle Connections" jsonschema_description:"The minimum idle connections that should be open to redis."`
	TLS                      *TLS   `koanf:"tls" json:"tls"`

	HighAvailability *SessionRedisHighAvailability `koanf:"high_availability" json:"high_availability"`
}

// SessionRedisHighAvailability holds configuration variables for Redis Cluster/Sentinel.
type SessionRedisHighAvailability struct {
	SentinelName     string `koanf:"sentinel_name" json:"sentinel_name" jsonschema:"title=Sentinel Name" jsonschema_description:"The name of the sentinel instance."`
	SentinelUsername string `koanf:"sentinel_username" json:"sentinel_username" jsonschema:"title=Sentinel Username" jsonschema_description:"The username for the sentinel instance."`
	SentinelPassword string `koanf:"sentinel_password" json:"sentinel_password" jsonschema:"title=Sentinel Username" jsonschema_description:"The username for the sentinel instance."`
	RouteByLatency   bool   `koanf:"route_by_latency" json:"route_by_latency" jsonschema:"default=false,title=Route by Latency" jsonschema_description:"Uses the Route by Latency mode."`
	RouteRandomly    bool   `koanf:"route_randomly" json:"route_randomly" jsonschema:"default=false,title=Route Randomly" jsonschema_description:"Uses the Route Randomly mode."`

	Nodes []SessionRedisHighAvailabilityNode `koanf:"nodes" json:"nodes" jsonschema:"title=Nodes" jsonschema_description:"The pre-populated list of nodes for the sentinel instance."`
}

// SessionRedisHighAvailabilityNode Represents a Node.
type SessionRedisHighAvailabilityNode struct {
	Host string `koanf:"host" json:"host" jsonschema:"title=Host" jsonschema_description:"The redis sentinel node host."`
	Port int    `koanf:"port" json:"port" jsonschema:"default=26379,title=Port" jsonschema_description:"The redis sentinel node port."`
}

// DefaultSessionConfiguration is the default session configuration.
var DefaultSessionConfiguration = Session{
	SessionCookieCommon: SessionCookieCommon{
		Name:       "authelia_session",
		Expiration: time.Hour,
		Inactivity: time.Minute * 5,
		RememberMe: time.Hour * 24 * 30,
		SameSite:   "lax",
	},
}

// DefaultRedisConfiguration is the default redis configuration.
var DefaultRedisConfiguration = SessionRedis{
	Port:                     6379,
	MaximumActiveConnections: 8,
	TLS: &TLS{
		MinimumVersion: TLSVersion{Value: tls.VersionTLS12},
	},
}

// DefaultRedisHighAvailabilityConfiguration is the default redis configuration.
var DefaultRedisHighAvailabilityConfiguration = SessionRedis{
	Port:                     26379,
	MaximumActiveConnections: 8,
	TLS: &TLS{
		MinimumVersion: TLSVersion{Value: tls.VersionTLS12},
	},
}
