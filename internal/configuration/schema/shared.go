package schema

import (
	"crypto/tls"
	"time"
)

// TLSConfig is a representation of the TLS configuration.
type TLSConfig struct {
	ServerName string `koanf:"server_name"`

	SkipVerify bool `koanf:"skip_verify"`

	MinimumVersion TLSVersion `koanf:"minimum_version"`
	MaximumVersion TLSVersion `koanf:"maximum_version"`

	ClientAuthKeyPair *X509KeyPair `koanf:"client_auth_keypair"`
}

// Config returns the schema.TLSConfig as a *tls.Config.
func (c *TLSConfig) Config() *tls.Config {
	config := &tls.Config{
		ServerName: c.ServerName,

		InsecureSkipVerify: c.SkipVerify, //nolint:gosec // Informed choice by user. Off by default.

		MinVersion: c.MinimumVersion.Version(),
		MaxVersion: c.MaximumVersion.Version(),
	}

	if c.ClientAuthKeyPair != nil {
		config.Certificates = []tls.Certificate{c.ClientAuthKeyPair.Certificate()}
	}

	return config
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
