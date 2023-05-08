package schema

import (
	"net/url"
	"time"
)

// TelemetryConfig represents the telemetry config.
type TelemetryConfig struct {
	Metrics TelemetryMetricsConfig `koanf:"metrics"`
}

// TelemetryMetricsConfig represents the telemetry metrics config.
type TelemetryMetricsConfig struct {
	Enabled bool        `koanf:"enabled"`
	Address *AddressTCP `koanf:"address"`
	UMask   *int        `koanf:"umask"`

	Buffers  ServerBuffers  `koanf:"buffers"`
	Timeouts ServerTimeouts `koanf:"timeouts"`
}

// DefaultTelemetryConfig is the default telemetry configuration.
var DefaultTelemetryConfig = TelemetryConfig{
	Metrics: TelemetryMetricsConfig{
		Address: &AddressTCP{Address{true, false, 9959, &url.URL{Scheme: AddressSchemeTCP, Host: ":9959"}}},
		Buffers: ServerBuffers{
			Read:  4096,
			Write: 4096,
		},
		Timeouts: ServerTimeouts{
			Read:  time.Second * 6,
			Write: time.Second * 6,
			Idle:  time.Second * 30,
		},
	},
}
