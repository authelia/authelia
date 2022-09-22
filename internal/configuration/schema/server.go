package schema

import (
	"time"
)

// ServerConfiguration represents the configuration of the http server.
type ServerConfiguration struct {
	Host               string `koanf:"host"`
	Port               int    `koanf:"port"`
	Path               string `koanf:"path"`
	AssetPath          string `koanf:"asset_path"`
	EnablePprof        bool   `koanf:"enable_pprof"`
	EnableExpvars      bool   `koanf:"enable_expvars"`
	DisableHealthcheck bool   `koanf:"disable_healthcheck"`

	TLS     ServerTLSConfiguration     `koanf:"tls"`
	Headers ServerHeadersConfiguration `koanf:"headers"`

	Buffers  ServerBuffers  `koanf:"buffers"`
	Timeouts ServerTimeouts `koanf:"timeouts"`
}

// ServerTLSConfiguration represents the configuration of the http servers TLS options.
type ServerTLSConfiguration struct {
	Certificate        string   `koanf:"certificate"`
	Key                string   `koanf:"key"`
	ClientCertificates []string `koanf:"client_certificates"`
}

// ServerHeadersConfiguration represents the customization of the http server headers.
type ServerHeadersConfiguration struct {
	AllowedHosts []string `koanf:"allowed_hosts"`
	CSPTemplate  string   `koanf:"csp_template"`
}

// DefaultServerConfiguration represents the default values of the ServerConfiguration.
var DefaultServerConfiguration = ServerConfiguration{
	Host: "0.0.0.0",
	Port: 9091,
	Buffers: ServerBuffers{
		Read:  4096,
		Write: 4096,
	},
	Timeouts: ServerTimeouts{
		Read:  time.Second * 2,
		Write: time.Second * 2,
		Idle:  time.Second * 30,
	},
}
