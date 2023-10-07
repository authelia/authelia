package validator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidateTelemetry(t *testing.T) {
	mustParseAddress := func(a string) *schema.AddressTCP {
		addr, err := schema.NewAddress(a)
		if err != nil {
			panic(err)
		}

		return &schema.AddressTCP{Address: *addr}
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
			&schema.Configuration{Telemetry: schema.Telemetry{Metrics: schema.TelemetryMetrics{Address: mustParseAddress("tcp://0.0.0.0")}}},
			&schema.Configuration{Telemetry: schema.Telemetry{
				Metrics: schema.TelemetryMetrics{
					Address: mustParseAddress("tcp://0.0.0.0:9959/metrics"),
					Buffers: schema.ServerBuffers{
						Read:  4096,
						Write: 4096,
					},
					Timeouts: schema.ServerTimeouts{
						Read:  time.Second * 6,
						Write: time.Second * 6,
						Idle:  time.Second * 30,
					},
				},
			}},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultPortAlt",
			&schema.Configuration{Telemetry: schema.Telemetry{Metrics: schema.TelemetryMetrics{Address: mustParseAddress("tcp://:0/metrics")}}},
			&schema.Configuration{Telemetry: schema.DefaultTelemetryConfig},
			nil,
			nil,
		},
		{
			"ShouldSetDefaultPortWithCustomIP",
			&schema.Configuration{Telemetry: schema.Telemetry{Metrics: schema.TelemetryMetrics{Address: mustParseAddress("tcp://127.0.0.1")}}},
			&schema.Configuration{Telemetry: schema.Telemetry{Metrics: schema.TelemetryMetrics{Address: mustParseAddress("tcp://127.0.0.1:9959/metrics")}}},
			nil,
			nil,
		},
		{
			"ShouldNotValidateUDP",
			&schema.Configuration{Telemetry: schema.Telemetry{Metrics: schema.TelemetryMetrics{Address: mustParseAddress("udp://0.0.0.0")}}},
			&schema.Configuration{Telemetry: schema.Telemetry{Metrics: schema.TelemetryMetrics{Address: mustParseAddress("udp://0.0.0.0:9959/metrics")}}},
			nil,
			[]string{"telemetry: metrics: option 'address' with value 'udp://0.0.0.0:0' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'udp'"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()

			ValidateTelemetry(tc.have, validator)

			assert.Equal(t, tc.expected.Telemetry.Metrics.Enabled, tc.have.Telemetry.Metrics.Enabled)
			assert.Equal(t, tc.expected.Telemetry.Metrics.Address, tc.have.Telemetry.Metrics.Address)

			lenWrns := len(tc.expectedWrns)
			wrns := validator.Warnings()

			if lenWrns == 0 {
				assert.Len(t, wrns, 0)
			} else {
				require.Len(t, wrns, lenWrns)

				for i, expectedWrn := range tc.expectedWrns {
					assert.EqualError(t, wrns[i], expectedWrn)
				}
			}

			lenErrs := len(tc.expectedErrs)
			errs := validator.Errors()

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

func TestValidateTelemetryShouldCorrectlyIdentifyValidAddressSchemes(t *testing.T) {
	testCases := []struct {
		have     string
		expected string
	}{
		{schema.AddressSchemeTCP, ""},
		{schema.AddressSchemeTCP4, ""},
		{schema.AddressSchemeTCP6, ""},
		{schema.AddressSchemeUDP, "telemetry: metrics: option 'address' with value 'udp://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'udp'"},
		{schema.AddressSchemeUDP4, "telemetry: metrics: option 'address' with value 'udp4://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'udp4'"},
		{schema.AddressSchemeUDP6, "telemetry: metrics: option 'address' with value 'udp6://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'udp6'"},
		{schema.AddressSchemeUnix, ""},
		{"http", "telemetry: metrics: option 'address' with value 'http://:9091' is invalid: scheme must be one of 'tcp', 'tcp4', 'tcp6', or 'unix' but is configured as 'http'"},
	}

	have := &schema.Configuration{}

	validator := schema.NewStructValidator()

	for _, tc := range testCases {
		t.Run(tc.have, func(t *testing.T) {
			validator.Clear()

			switch tc.have {
			case schema.AddressSchemeUnix:
				have.Telemetry.Metrics.Address = &schema.AddressTCP{Address: schema.NewAddressUnix("/path/to/authelia.sock")}
			default:
				have.Telemetry.Metrics.Address = &schema.AddressTCP{Address: schema.NewAddressFromNetworkValues(tc.have, "", 9091)}
			}

			ValidateTelemetry(have, validator)

			assert.Len(t, validator.Warnings(), 0)

			if tc.expected == "" {
				assert.Len(t, validator.Errors(), 0)
			} else {
				require.Len(t, validator.Errors(), 1)
				assert.EqualError(t, validator.Errors()[0], tc.expected)
			}
		})
	}
}
