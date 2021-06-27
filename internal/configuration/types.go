package configuration

import (
	"github.com/knadh/koanf"

	"github.com/authelia/authelia/internal/configuration/schema"
)

// Provider holds the koanf.Koanf instance and the schema.Configuration.
type Provider struct {
	*koanf.Koanf
	*schema.StructValidator

	configuration *schema.Configuration
}

// Source is a configuration source.
type Source struct {
	Name     string
	Provider koanf.Provider
	Parser   koanf.Parser
}
