// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

import "net"

// DynamicConfigurationSource represents the static configuration which describes where the dynamic configuration is
// loaded from and how changes to it are detected.
type DynamicConfigurationSource struct {
	Paths []string `koanf:"paths" yaml:"paths,omitempty" toml:"paths,omitempty" json:"paths,omitempty" jsonschema:"title=Paths" jsonschema_description:"The paths to the files or directories which contain the dynamic configuration."`
	Watch bool     `koanf:"watch" yaml:"watch" toml:"watch" json:"watch" jsonschema:"default=false,title=Watch" jsonschema_description:"Automatically reloads the dynamic configuration when the files change."`
}

// DynamicConfiguration represents the configuration which can be reloaded at runtime without restarting the
// application. Only the sections listed here are supported in the files referenced by DynamicConfigurationSource.
type DynamicConfiguration struct {
	Definitions   DynamicDefinitions `koanf:"definitions" yaml:"definitions,omitempty" toml:"definitions,omitempty" json:"definitions,omitempty" jsonschema:"title=Definitions" jsonschema_description:"Definitions for items reused elsewhere in the dynamic configuration."`
	AccessControl AccessControl      `koanf:"access_control" yaml:"access_control,omitempty" toml:"access_control,omitempty" json:"access_control,omitempty" jsonschema:"title=Access Control" jsonschema_description:"Access Control Configuration."`
}

// DynamicDefinitions represents the definitions which can be used in the dynamic configuration.
type DynamicDefinitions struct {
	Network map[string][]*net.IPNet `koanf:"network" yaml:"network,omitempty" toml:"network,omitempty" json:"network,omitempty" jsonschema:"title=Network Definitions" jsonschema_description:"Networks CIDR ranges that can be utilized elsewhere in the dynamic configuration."`
}
