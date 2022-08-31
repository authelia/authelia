package schema

import (
	"time"
)

// TLSConfig is a representation of the TLS configuration.
type TLSConfig struct {
	MinimumVersion string `koanf:"minimum_version"`
	SkipVerify     bool   `koanf:"skip_verify"`
	ServerName     string `koanf:"server_name"`
}

// ServerTimeouts represents server timeout configurations.
type ServerTimeouts struct {
	Read  time.Duration `koanf:"read"`
	Write time.Duration `koanf:"write"`
	Idle  time.Duration `koanf:"idle"`
}

// ServerBuffers represents server buffer configurations.
type ServerBuffers struct {
	Read  int `koanf:"read"`
	Write int `koanf:"write"`
}
