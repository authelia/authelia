package schema

import (
	"time"
)

// TLSConfig is a representation of the TLS configuration.
type TLSConfig struct {
	MinimumVersion TLSVersion `koanf:"minimum_version"`
	MaximumVersion TLSVersion `koanf:"maximum_version"`

	SkipVerify bool   `koanf:"skip_verify"`
	ServerName string `koanf:"server_name"`

	PrivateKey       CryptographicPrivateKey `koanf:"private_key"`
	CertificateChain X509CertificateChain    `koanf:"certificate_chain"`
}

// TLSCertificateConfig is a representation of the TLS Certificate configuration.
type TLSCertificateConfig struct {
	Key              CryptographicPrivateKey `koanf:"key"`
	CertificateChain X509CertificateChain    `koanf:"certificate_chain"`
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

// JWK represents a JWK.
type JWK struct {
	KeyID            string               `koanf:"key_id"`
	Use              string               `koanf:"use"`
	Algorithm        string               `koanf:"algorithm"`
	Key              CryptographicKey     `koanf:"key"`
	CertificateChain X509CertificateChain `koanf:"certificate_chain"`
}
