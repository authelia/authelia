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
	Enabled bool     `koanf:"enabled"`
	Address *Address `koanf:"address"`
}

// DefaultTelemetryConfig is the default telemetry configuration.
var DefaultTelemetryConfig = TelemetryConfig{
	Metrics: TelemetryMetricsConfig{
		Address: &Address{true, "tcp", net.ParseIP("0.0.0.0"), 9959},
	},
}
