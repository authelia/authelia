package schema

import (
	"time"
)

// TLS is a representation of the TLS configuration.
type TLS struct {
	MinimumVersion TLSVersion `koanf:"minimum_version" yaml:"minimum_version,omitempty" toml:"minimum_version,omitempty" json:"minimum_version,omitempty" jsonschema:"default=TLS1.2,title=Minimum Version" jsonschema_description:"The minimum TLS version accepted."`
	MaximumVersion TLSVersion `koanf:"maximum_version" yaml:"maximum_version,omitempty" toml:"maximum_version,omitempty" json:"maximum_version,omitempty" jsonschema:"default=TLS1.3,title=Maximum Version" jsonschema_description:"The maximum TLS version accepted."`

	SkipVerify bool   `koanf:"skip_verify" yaml:"skip_verify" toml:"skip_verify" json:"skip_verify" jsonschema:"default=false,title=Skip Verify" jsonschema_description:"Disable all verification of the TLS properties."`
	ServerName string `koanf:"server_name" yaml:"server_name,omitempty" toml:"server_name,omitempty" json:"server_name,omitempty" jsonschema:"format=hostname,title=Server Name" jsonschema_description:"The expected server name to match the certificate against."`

	PrivateKey       CryptographicPrivateKey `koanf:"private_key" yaml:"private_key,omitempty" toml:"private_key,omitempty" json:"private_key,omitempty" jsonschema:"title=Private Key" jsonschema_description:"The private key."`
	CertificateChain X509CertificateChain    `koanf:"certificate_chain" yaml:"certificate_chain,omitempty" toml:"certificate_chain,omitempty" json:"certificate_chain,omitempty" jsonschema:"title=Certificate Chain" jsonschema_description:"The certificate chain."`
}

// ServerTimeouts represents server timeout configurations.
type ServerTimeouts struct {
	Read  time.Duration `koanf:"read" yaml:"read,omitempty" toml:"read,omitempty" json:"read,omitempty" jsonschema:"default=6 seconds,title=Read" jsonschema_description:"The read timeout."`
	Write time.Duration `koanf:"write" yaml:"write,omitempty" toml:"write,omitempty" json:"write,omitempty" jsonschema:"default=6 seconds,title=Write" jsonschema_description:"The write timeout."`
	Idle  time.Duration `koanf:"idle" yaml:"idle,omitempty" toml:"idle,omitempty" json:"idle,omitempty" jsonschema:"default=30 seconds,title=Idle" jsonschema_description:"The idle timeout."`
}

// ServerBuffers represents server buffer configurations.
type ServerBuffers struct {
	Read  int `koanf:"read" yaml:"read" toml:"read" json:"read" jsonschema:"default=4096,title=Read" jsonschema_description:"The read buffer size."`
	Write int `koanf:"write" yaml:"write" toml:"write" json:"write" jsonschema:"default=4096,title=Write" jsonschema_description:"The write buffer size."`
}

// JWK represents a JWK.
type JWK struct {
	KeyID            string               `koanf:"key_id" yaml:"key_id,omitempty" toml:"key_id,omitempty" json:"key_id,omitempty" jsonschema:"maxLength=100,title=Key ID" jsonschema_description:"The ID of this JWK."`
	Use              string               `koanf:"use" yaml:"use,omitempty" toml:"use,omitempty" json:"use,omitempty" jsonschema:"default=sig,enum=sig,title=Use" jsonschema_description:"The Use of this JWK."`
	Algorithm        string               `koanf:"algorithm" yaml:"algorithm,omitempty" toml:"algorithm,omitempty" json:"algorithm,omitempty" jsonschema:"enum=HS256,enum=HS384,enum=HS512,enum=RS256,enum=RS384,enum=RS512,enum=ES256,enum=ES384,enum=ES512,enum=PS256,enum=PS384,enum=PS512,title=Algorithm" jsonschema_description:"The Algorithm of this JWK."`
	Key              CryptographicKey     `koanf:"key" yaml:"key,omitempty" toml:"key,omitempty" json:"key,omitempty" jsonschema_description:"The Private/Public key material of this JWK in Base64 PEM format."`
	CertificateChain X509CertificateChain `koanf:"certificate_chain" yaml:"certificate_chain,omitempty" toml:"certificate_chain,omitempty" json:"certificate_chain,omitempty" jsonschema:"title=Certificate Chain" jsonschema_description:"The optional associated certificate which matches the Key public key portion for this JWK."`
}
