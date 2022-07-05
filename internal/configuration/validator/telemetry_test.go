package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidateTelemetry(t *testing.T) {
	mustParseAddress := func(a string) *schema.Address {
		addr, err := schema.NewAddressFromString(a)
		if err != nil {
			panic(err)
		}

		return addr
	}

	testCases := []struct {
		name                       string
		have                       *schema.Configuration
		expected                   *schema.Configuration
		expectedWrns, expectedErrs []string
	}{
		{
			"ShouldSetDefaults",
			&schema.Configuration{},
			&schema.Configuration{Telemetry: schema.DefaultTelemetryConfig},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultPort",
			&schema.Configuration{Telemetry: schema.TelemetryConfig{Metrics: schema.TelemetryMetricsConfig{Address: mustParseAddress("tcp://0.0.0.0")}}},
			&schema.Configuration{Telemetry: schema.DefaultTelemetryConfig},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultPortAlt",
			&schema.Configuration{Telemetry: schema.TelemetryConfig{Metrics: schema.TelemetryMetricsConfig{Address: mustParseAddress("tcp://0.0.0.0:0")}}},
			&schema.Configuration{Telemetry: schema.DefaultTelemetryConfig},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultPortWithCustomIP",
			&schema.Configuration{Telemetry: schema.TelemetryConfig{Metrics: schema.TelemetryMetricsConfig{Address: mustParseAddress("tcp://127.0.0.1")}}},
			&schema.Configuration{Telemetry: schema.TelemetryConfig{Metrics: schema.TelemetryMetricsConfig{Address: mustParseAddress("tcp://127.0.0.1:9959")}}},
			nil,
			nil,
		},
		{
			"ShouldNotValidateUDP",
			&schema.Configuration{Telemetry: schema.TelemetryConfig{Metrics: schema.TelemetryMetricsConfig{Address: mustParseAddress("udp://0.0.0.0")}}},
			&schema.Configuration{Telemetry: schema.TelemetryConfig{Metrics: schema.TelemetryMetricsConfig{Address: mustParseAddress("udp://0.0.0.0:9959")}}},
			nil,
			[]string{"telemetry: metrics: option 'address' must have a scheme 'tcp://' but it is configured as 'udp'"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			v := schema.NewStructValidator()

			ValidateTelemetry(tc.have, v)

			assert.Equal(t, tc.expected.Telemetry.Metrics.Enabled, tc.have.Telemetry.Metrics.Enabled)
			assert.Equal(t, tc.expected.Telemetry.Metrics.Address, tc.have.Telemetry.Metrics.Address)

			lenWrns := len(tc.expectedWrns)
			wrns := v.Warnings()

			if lenWrns == 0 {
				assert.Len(t, wrns, 0)
			} else {
				require.Len(t, wrns, lenWrns)

				for i, expectedWrn := range tc.expectedWrns {
					assert.EqualError(t, wrns[i], expectedWrn)
				}
			}

			lenErrs := len(tc.expectedErrs)
			errs := v.Errors()

			if lenErrs == 0 {
				assert.Len(t, errs, 0)
			} else {
				require.Len(t, errs, lenErrs)

				for i, expectedErr := range tc.expectedErrs {
					assert.EqualError(t, errs[i], expectedErr)
				}
			}
		})
	}
}
