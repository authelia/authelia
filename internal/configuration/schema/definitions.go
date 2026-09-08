package schema

import "net"

// Definitions represents the definitions which can be referenced elsewhere in the configuration.
type Definitions struct {
	Network        map[string][]*net.IPNet  `koanf:"network" yaml:"network,omitempty" toml:"network,omitempty" json:"network,omitempty" jsonschema:"title=Network Definitions" jsonschema_description:"Networks CIDR ranges that can be utilized elsewhere in the configuration."`
	UserAttributes map[string]UserAttribute `koanf:"user_attributes" yaml:"user_attributes,omitempty" toml:"user_attributes,omitempty" json:"user_attributes,omitempty" jsonschema:"title=User Attributes" jsonschema_description:"User attributes derived from other attributes."`
}

// UserAttribute represents a user attribute definition.
type UserAttribute struct {
	Expression string `koanf:"expression" yaml:"expression,omitempty" toml:"expression,omitempty" json:"expression,omitempty" jsonschema:"required,title=Expression" jsonschema_description:"Expression to derive the user attribute using the common expression language."`
}
