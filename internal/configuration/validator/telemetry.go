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
}
