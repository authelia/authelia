package schema

import (
	"net/url"
	"time"
)

// NTP represents the configuration related to ntp server.
type NTP struct {
	Address             *AddressUDP   `koanf:"address" yaml:"address,omitempty" toml:"address,omitempty" json:"address,omitempty" jsonschema:"title=NTP Address" jsonschema_description:"The remote address of the NTP server."`
	Version             int           `koanf:"version" yaml:"version" toml:"version" json:"version" jsonschema:"enum=3,enum=4,title=NTP Version" jsonschema_description:"The NTP Version to use."`
	MaximumDesync       time.Duration `koanf:"max_desync" yaml:"max_desync,omitempty" toml:"max_desync,omitempty" json:"max_desync,omitempty" jsonschema:"default=3 seconds,title=Maximum Desync" jsonschema_description:"The maximum amount of time that the server can be out of sync."`
	DisableStartupCheck bool          `koanf:"disable_startup_check" yaml:"disable_startup_check" toml:"disable_startup_check" json:"disable_startup_check" jsonschema:"default=false,title=Disable Startup Check" jsonschema_description:"Disables the NTP Startup Check entirely."`
	DisableFailure      bool          `koanf:"disable_failure" yaml:"disable_failure" toml:"disable_failure" json:"disable_failure" jsonschema:"default=false,title=Disable Failure" jsonschema_description:"Disables complete failure whe the Startup Check fails and instead just logs the error."`
}

// DefaultNTPConfiguration represents default configuration parameters for the NTP server.
var DefaultNTPConfiguration = NTP{
	Address:       &AddressUDP{Address{valid: true, socket: false, port: 123, url: &url.URL{Scheme: AddressSchemeUDP, Host: "time.cloudflare.com:123"}}},
	Version:       4,
	MaximumDesync: time.Second * 3,
}
