package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateTelemetry validates the telemetry configuration.
func ValidateTelemetry(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Telemetry.Metrics.Address == nil {
		config.Telemetry.Metrics.Address = schema.DefaultTelemetryConfig.Metrics.Address
	}

	switch config.Telemetry.Metrics.Address.Scheme {
	case "tcp":
		break
	default:
		validator.Push(fmt.Errorf(errFmtTelemetryMetricsScheme, config.Telemetry.Metrics.Address.Scheme))
	}

	if config.Telemetry.Metrics.Address.Port == 0 {
		config.Telemetry.Metrics.Address.Port = schema.DefaultTelemetryConfig.Metrics.Address.Port
	}

	if config.Telemetry.Metrics.Buffers.Read <= 0 {
		config.Telemetry.Metrics.Buffers.Read = schema.DefaultTelemetryConfig.Metrics.Buffers.Read
	}

	if config.Telemetry.Metrics.Buffers.Write <= 0 {
		config.Telemetry.Metrics.Buffers.Write = schema.DefaultTelemetryConfig.Metrics.Buffers.Write
	}

	if config.Telemetry.Metrics.Timeouts.Read <= 0 {
		config.Telemetry.Metrics.Timeouts.Read = schema.DefaultTelemetryConfig.Metrics.Timeouts.Read
	}

	if config.Telemetry.Metrics.Timeouts.Write <= 0 {
		config.Telemetry.Metrics.Timeouts.Write = schema.DefaultTelemetryConfig.Metrics.Timeouts.Write
	}

	if config.Telemetry.Metrics.Timeouts.Idle <= 0 {
		config.Telemetry.Metrics.Timeouts.Idle = schema.DefaultTelemetryConfig.Metrics.Timeouts.Idle
	}
}
