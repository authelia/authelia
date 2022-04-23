package schema

import (
	"net"
)

// TelemetryConfig represents the telemetry config.
type TelemetryConfig struct {
	Metrics TelemetryMetricsConfig `koanf:"metrics"`
}

// TelemetryMetricsConfig represents the telemetry metrics config.
type TelemetryMetricsConfig struct {
	Enabled bool    `koanf:"enabled"`
	Address Address `koanf:"address"`
}

var DefaultTelemetryConfig = TelemetryConfig{
	Metrics: TelemetryMetricsConfig{
		Address: NewAddress("tcp", net.ParseIP("0.0.0.0"), 9961),
	},
}
