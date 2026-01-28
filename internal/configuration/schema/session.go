package schema

import (
	"crypto/tls"
	"net/url"
	"time"
)

// Session represents the configuration related to user sessions.
type Session struct {
	SessionCookieCommon `koanf:",squash"`

	Secret string `koanf:"secret" yaml:"secret,omitempty" toml:"secret,omitempty" json:"secret,omitempty" jsonschema:"title=Secret" jsonschema_description:"Secret used to encrypt the session data."`

	Cookies []SessionCookie `koanf:"cookies" yaml:"cookies,omitempty" toml:"cookies,omitempty" json:"cookies,omitempty" jsonschema:"title=Cookies" jsonschema_description:"List of cookie domain configurations."`

	Redis *SessionRedis `koanf:"redis" yaml:"redis,omitempty" toml:"redis,omitempty" json:"redis,omitempty" jsonschema:"title=Redis" jsonschema_description:"Redis Session Provider configuration."`

	File *SessionFile `koanf:"file" yaml:"file,omitempty" toml:"file,omitempty" json:"file,omitempty" jsonschema:"title=File" jsonschema_description:"File Session Provider configuration."`

	// Deprecated: Use the session cookies option with the same name instead.
	Domain string `koanf:"domain" yaml:"domain,omitempty" toml:"domain,omitempty" json:"domain,omitempty" jsonschema:"deprecated,title=Domain"`
}

type SessionCookieCommon struct {
	Name       string        `koanf:"name" yaml:"name,omitempty" toml:"name,omitempty" json:"name,omitempty" jsonschema:"default=authelia_session" jsonschema_description:"The session cookie name."`
	SameSite   string        `koanf:"same_site" yaml:"same_site,omitempty" toml:"same_site,omitempty" json:"same_site,omitempty" jsonschema:"default=lax,enum=lax,enum=strict,enum=none" jsonschema_description:"The session cookie same site value."`
	Expiration time.Duration `koanf:"expiration" yaml:"expiration,omitempty" toml:"expiration,omitempty" json:"expiration,omitempty" jsonschema:"default=1 hour" jsonschema_description:"The session cookie expiration when remember me is not checked."`
	Inactivity time.Duration `koanf:"inactivity" yaml:"inactivity,omitempty" toml:"inactivity,omitempty" json:"inactivity,omitempty" jsonschema:"default=5 minutes" jsonschema_description:"The session inactivity timeout."`
	RememberMe time.Duration `koanf:"remember_me" yaml:"remember_me,omitempty" toml:"remember_me,omitempty" json:"remember_me,omitempty" jsonschema:"default=30 days" jsonschema_description:"The session cookie expiration when remember me is checked."`

	DisableRememberMe bool `json:"-"`
}

// SessionCookie represents the configuration for a cookie domain.
type SessionCookie struct {
	SessionCookieCommon `koanf:",squash"`

	Domain                string   `koanf:"domain" yaml:"domain,omitempty" toml:"domain,omitempty" json:"domain,omitempty" jsonschema:"format=hostname,title=Domain" jsonschema_description:"The domain for this session cookie configuration."`
	AutheliaURL           *url.URL `koanf:"authelia_url" yaml:"authelia_url,omitempty" toml:"authelia_url,omitempty" json:"authelia_url,omitempty" jsonschema:"format=uri,title=Authelia URL" jsonschema_description:"The Root Authelia URL to redirect users to for this session cookie configuration."`
	DefaultRedirectionURL *url.URL `koanf:"default_redirection_url" yaml:"default_redirection_url,omitempty" toml:"default_redirection_url,omitempty" json:"default_redirection_url,omitempty" jsonschema:"format=uri,title=Default Redirection URL" jsonschema_description:"The default redirection URL for this session cookie configuration."`

	Legacy bool `json:"-"`
}

// SessionRedis represents the configuration related to redis session store.
type SessionRedis struct {
	Host                     string        `koanf:"host" yaml:"host,omitempty" toml:"host,omitempty" json:"host,omitempty" jsonschema:"title=Host" jsonschema_description:"The redis server host."`
	Port                     int           `koanf:"port" yaml:"port" toml:"port" json:"port" jsonschema:"default=6379,title=Host" jsonschema_description:"The redis server port."`
	Timeout                  time.Duration `koanf:"timeout" yaml:"timeout,omitempty" toml:"timeout,omitempty" json:"timeout,omitempty" jsonschema:"default=5 seconds,title=Timeout" jsonschema_description:"The Redis server connection timeout."`
	MaxRetries               int           `koanf:"max_retries" yaml:"max_retries" toml:"max_retries" json:"max_retries" jsonschema:"default=3,title=Maximum Retries" jsonschema_description:"The maximum number of retries on a failed command."`
	Username                 string        `koanf:"username" yaml:"username,omitempty" toml:"username,omitempty" json:"username,omitempty" jsonschema:"title=Username" jsonschema_description:"The redis username."`
	Password                 string        `koanf:"password" yaml:"password,omitempty" toml:"password,omitempty" json:"password,omitempty" jsonschema:"title=Password" jsonschema_description:"The redis password."`
	DatabaseIndex            int           `koanf:"database_index" yaml:"database_index" toml:"database_index" json:"database_index" jsonschema:"default=0,title=Database Index" jsonschema_description:"The redis database index."`
	MaximumActiveConnections int           `koanf:"maximum_active_connections" yaml:"maximum_active_connections" toml:"maximum_active_connections" json:"maximum_active_connections" jsonschema:"default=8,title=Maximum Active Connections" jsonschema_description:"The maximum connections that can be made to redis at one time."`
	MinimumIdleConnections   int           `koanf:"minimum_idle_connections" yaml:"minimum_idle_connections" toml:"minimum_idle_connections" json:"minimum_idle_connections" jsonschema:"title=Minimum Idle Connections" jsonschema_description:"The minimum idle connections that should be open to redis."`
	TLS                      *TLS          `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty"`

	HighAvailability *SessionRedisHighAvailability `koanf:"high_availability" yaml:"high_availability,omitempty" toml:"high_availability,omitempty" json:"high_availability,omitempty"`
}

// SessionRedisHighAvailability holds configuration variables for Redis Cluster/Sentinel.
type SessionRedisHighAvailability struct {
	SentinelName     string `koanf:"sentinel_name" yaml:"sentinel_name,omitempty" toml:"sentinel_name,omitempty" json:"sentinel_name,omitempty" jsonschema:"title=Sentinel Name" jsonschema_description:"The name of the sentinel instance."`
	SentinelUsername string `koanf:"sentinel_username" yaml:"sentinel_username,omitempty" toml:"sentinel_username,omitempty" json:"sentinel_username,omitempty" jsonschema:"title=Sentinel Username" jsonschema_description:"The username for the sentinel instance."`
	SentinelPassword string `koanf:"sentinel_password" yaml:"sentinel_password,omitempty" toml:"sentinel_password,omitempty" json:"sentinel_password,omitempty" jsonschema:"title=Sentinel Username" jsonschema_description:"The username for the sentinel instance."`
	RouteByLatency   bool   `koanf:"route_by_latency" yaml:"route_by_latency" toml:"route_by_latency" json:"route_by_latency" jsonschema:"default=false,title=Route by Latency" jsonschema_description:"Uses the Route by Latency mode."`
	RouteRandomly    bool   `koanf:"route_randomly" yaml:"route_randomly" toml:"route_randomly" json:"route_randomly" jsonschema:"default=false,title=Route Randomly" jsonschema_description:"Uses the Route Randomly mode."`

	Nodes []SessionRedisHighAvailabilityNode `koanf:"nodes" yaml:"nodes,omitempty" toml:"nodes,omitempty" json:"nodes,omitempty" jsonschema:"title=Nodes" jsonschema_description:"The pre-populated list of nodes for the sentinel instance."`
}

// SessionRedisHighAvailabilityNode Represents a Node.
type SessionRedisHighAvailabilityNode struct {
	Host string `koanf:"host" yaml:"host,omitempty" toml:"host,omitempty" json:"host,omitempty" jsonschema:"title=Host" jsonschema_description:"The redis sentinel node host."`
	Port int    `koanf:"port" yaml:"port" toml:"port" json:"port" jsonschema:"default=26379,title=Port" jsonschema_description:"The redis sentinel node port."`
}

// SessionFile represents the configuration for file-based session storage.
type SessionFile struct {
	Path            string        `koanf:"path" yaml:"path,omitempty" toml:"path,omitempty" json:"path,omitempty" jsonschema:"title=Path" jsonschema_description:"The directory path for session files."`
	CleanupInterval time.Duration `koanf:"cleanup_interval" yaml:"cleanup_interval,omitempty" toml:"cleanup_interval,omitempty" json:"cleanup_interval,omitempty" jsonschema:"default=5 minutes,title=Cleanup Interval" jsonschema_description:"The interval between expired session cleanup runs."`
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
	Timeout:                  time.Second * 5,
	MaxRetries:               0,
	MaximumActiveConnections: 8,
	TLS: &TLS{
		MinimumVersion: TLSVersion{Value: tls.VersionTLS12},
	},
}

// DefaultRedisHighAvailabilityConfiguration is the default redis configuration.
var DefaultRedisHighAvailabilityConfiguration = SessionRedis{
	Port:                     26379,
	Timeout:                  time.Second * 5,
	MaxRetries:               0,
	MaximumActiveConnections: 8,
	TLS: &TLS{
		MinimumVersion: TLSVersion{Value: tls.VersionTLS12},
	},
}

// DefaultFileConfiguration is the default file session configuration.
var DefaultFileConfiguration = SessionFile{
	CleanupInterval: time.Minute * 5,
}
