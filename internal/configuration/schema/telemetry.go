package schema

import (
	"net"
	"time"
)

// TelemetryConfig represents the telemetry config.
type TelemetryConfig struct {
	Metrics TelemetryMetricsConfig `koanf:"metrics"`
}

// TelemetryMetricsConfig represents the telemetry metrics config.
type TelemetryMetricsConfig struct {
	Enabled bool     `koanf:"enabled"`
	Address *Address `koanf:"address"`

	Buffers  ServerBuffers  `koanf:"buffers"`
	Timeouts ServerTimeouts `koanf:"timeouts"`
}

// DefaultTelemetryConfig is the default telemetry configuration.
var DefaultTelemetryConfig = TelemetryConfig{
	Metrics: TelemetryMetricsConfig{
		Address: &Address{true, "tcp", net.ParseIP("0.0.0.0"), 9959},
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
