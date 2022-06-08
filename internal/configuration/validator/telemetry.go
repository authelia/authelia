package validator

import (
	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidateTelemetry validates the telemetry configuration.
func ValidateTelemetry(config *schema.Configuration, validator *schema.StructValidator) {
	if config.Telemetry.Metrics.Enabled {
		if config.Telemetry.Metrics.Address.String() == "" {
			config.Telemetry.Metrics.Address = schema.DefaultTelemetryConfig.Metrics.Address
		}
	}
}
