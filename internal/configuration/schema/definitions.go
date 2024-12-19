package schema

import "net"

type Definitions struct {
	Network map[string][]*net.IPNet `koanf:"network" json:"network" jsonschema:"title=Network Definitions" jsonschema_description:"Networks CIDR ranges that can be utilized elsewhere in the configuration."`
}
