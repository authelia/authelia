package schema

type Listeners []Listener

type Listener struct {
	Address Address `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty" jsonschema:"title=Address" jsonschema_description:"Listener Address"`
	TLS     *TLS    `koanf:"tls" yaml:"tls,omitempty" toml:"tls,omitempty" json:"tls,omitempty"`
}
