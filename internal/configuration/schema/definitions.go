// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import "net"

// Definitions represents the definitions which can be referenced elsewhere in the configuration.
type Definitions struct {
	Network        map[string][]*net.IPNet  `koanf:"network" yaml:"network,omitempty" toml:"network,omitempty" json:"network,omitempty" jsonschema:"title=Network Definitions" jsonschema_description:"Networks CIDR ranges that can be utilized elsewhere in the configuration."`
	UserAttributes map[string]UserAttribute `koanf:"user_attributes" yaml:"user_attributes,omitempty" toml:"user_attributes,omitempty" json:"user_attributes,omitempty" jsonschema:"title=User Attributes" jsonschema_description:"User attributes derived from other attributes."`
	URIParsers     map[string]URIParser     `koanf:"uri_parsers" yaml:"uri_parsers,omitempty" toml:"uri_parsers,omitempty" json:"uri_parsers,omitempty" jsonschema:"title=URI Parsers" jsonschema_description:"URI parsers which can be used by the authz endpoints."`
}

// URIParser represents a URI parser definition.
type URIParser struct {
	Allow []string `koanf:"allow" yaml:"allow,omitempty" toml:"allow,omitempty" json:"allow,omitempty" jsonschema:"title=Allow" jsonschema_description:"The URI ambiguity classes to allow which are otherwise rejected by default."`

	Reject []string `koanf:"reject" yaml:"reject,omitempty" toml:"reject,omitempty" json:"reject,omitempty" jsonschema:"title=Reject" jsonschema_description:"The URI ambiguity classes to reject which are otherwise allowed by default."`
}

// UserAttribute represents a user attribute definition.
type UserAttribute struct {
	Expression string `koanf:"expression" yaml:"expression,omitempty" toml:"expression,omitempty" json:"expression,omitempty" jsonschema:"required,title=Expression" jsonschema_description:"Expression to derive the user attribute using the common expression language."`
}
