package schema

import (
	"net/url"
	"time"
)

// Session represents the configuration related to user sessions.
type Session struct {
	SessionCookieCommon `koanf:",squash" yaml:",inline"`

	Secret string `koanf:"secret" yaml:"secret,omitempty" toml:"secret,omitempty" json:"secret,omitempty" jsonschema:"title=Secret" jsonschema_description:"Secret used to encrypt the session data."`

	Storage string `koanf:"storage" yaml:"storage,omitempty" toml:"storage,omitempty" json:"storage,omitempty" jsonschema:"title=Storage,enum=internal,enum=cache" jsonschema_description:"The storage provider to use for session storage."`

	Cookies []SessionCookie `koanf:"cookies" yaml:"cookies,omitempty" toml:"cookies,omitempty" json:"cookies,omitempty" jsonschema:"title=Cookies" jsonschema_description:"List of cookie domain configurations."`

	// Deprecated: Use the session cookies option with the same name instead.
	Domain string `koanf:"domain" yaml:"domain,omitempty" toml:"domain,omitempty" json:"domain,omitempty" jsonschema:"deprecated,title=Domain"`
}

// SessionCookieCommon represents the session cookie configuration options shared by every session cookie domain.
type SessionCookieCommon struct {
	Name       string        `koanf:"name" yaml:"name,omitempty" toml:"name,omitempty" json:"name,omitempty" jsonschema:"default=authelia_session,title=Name" jsonschema_description:"The session cookie name."`
	SameSite   string        `koanf:"same_site" yaml:"same_site,omitempty" toml:"same_site,omitempty" json:"same_site,omitempty" jsonschema:"default=lax,enum=lax,enum=strict,enum=none,title=Same Site" jsonschema_description:"The session cookie same site value."`
	Expiration time.Duration `koanf:"expiration" yaml:"expiration,omitempty" toml:"expiration,omitempty" json:"expiration,omitempty" jsonschema:"default=1 hour,title=Expiration" jsonschema_description:"The session cookie expiration when remember me is not checked."`
	Inactivity time.Duration `koanf:"inactivity" yaml:"inactivity,omitempty" toml:"inactivity,omitempty" json:"inactivity,omitempty" jsonschema:"default=5 minutes,title=Inactivity" jsonschema_description:"The session inactivity timeout."`
	RememberMe time.Duration `koanf:"remember_me" yaml:"remember_me,omitempty" toml:"remember_me,omitempty" json:"remember_me,omitempty" jsonschema:"default=30 days,title=Remember Me" jsonschema_description:"The session cookie expiration when remember me is checked."`

	DisableRememberMe bool `yaml:"-" toml:"-" json:"-"`
}

// SessionCookie represents the configuration for a cookie domain.
type SessionCookie struct {
	SessionCookieCommon `koanf:",squash" yaml:",inline"`

	Domain                string   `koanf:"domain" yaml:"domain,omitempty" toml:"domain,omitempty" json:"domain,omitempty" jsonschema:"format=hostname,title=Domain" jsonschema_description:"The domain for this session cookie configuration."`
	AutheliaURL           *url.URL `koanf:"authelia_url" yaml:"authelia_url,omitempty" toml:"authelia_url,omitempty" json:"authelia_url,omitempty" jsonschema:"format=uri,title=Authelia URL" jsonschema_description:"The Root Authelia URL to redirect users to for this session cookie configuration."`
	DefaultRedirectionURL *url.URL `koanf:"default_redirection_url" yaml:"default_redirection_url,omitempty" toml:"default_redirection_url,omitempty" json:"default_redirection_url,omitempty" jsonschema:"format=uri,title=Default Redirection URL" jsonschema_description:"The default redirection URL for this session cookie configuration."`

	Legacy bool `yaml:"-" toml:"-" json:"-"`
}

// DefaultSessionConfiguration is the default session configuration.
var DefaultSessionConfiguration = Session{
	Storage: "internal",
	SessionCookieCommon: SessionCookieCommon{
		Name:       "authelia_session",
		Expiration: time.Hour,
		Inactivity: time.Minute * 5,
		RememberMe: time.Hour * 24 * 30,
		SameSite:   "lax",
	},
}
