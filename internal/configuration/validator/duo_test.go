// SPDX-FileCopyrightText: 2026 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidateDuoShouldRejectDisableCAPinningForDuoHostnames(t *testing.T) {
	testCases := []struct {
		name     string
		hostname string
	}{
		{"ShouldRejectPort", "api-123456789.duosecurity.com:443"},
		{"ShouldRejectFullyQualifiedPort", "api-123456789.duosecurity.com.:8443"},
		{"ShouldRejectUserInfo", "user@api-123456789.duofederal.com"},
		{"ShouldRejectSurroundingWhitespace", " api-123456789.duosecurity.com "},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &schema.Configuration{DuoAPI: schema.DuoAPI{
				Hostname:         tc.hostname,
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			}}

			validator := schema.NewStructValidator()

			ValidateDuo(config, validator)

			assert.Len(t, validator.Warnings(), 0)
			require.Len(t, validator.Errors(), 1)
			assert.EqualError(t, validator.Errors()[0], fmt.Sprintf("duo_api: option 'experimental_disable_ca_pinning' must be false when the option 'hostname' is a Duo hostname but it's configured as '%s'", tc.hostname))
		})
	}
}

func TestValidateDuo(t *testing.T) {
	testCases := []struct {
		desc     string
		have     *schema.Configuration
		expected schema.DuoAPI
		errs     []string
		warns    []string
	}{
		{
			desc:     "ShouldDisableDuo",
			have:     &schema.Configuration{},
			expected: schema.DuoAPI{Disable: true},
		},
		{
			desc:     "ShouldDisableDuoConfigured",
			have:     &schema.Configuration{DuoAPI: schema.DuoAPI{Disable: true, Hostname: "example.com"}},
			expected: schema.DuoAPI{Disable: true, Hostname: "example.com"},
		},
		{
			desc: "ShouldNotDisableDuo",
			have: &schema.Configuration{DuoAPI: schema.DuoAPI{
				Hostname:       "test",
				IntegrationKey: "test",
				SecretKey:      "test",
			}},
			expected: schema.DuoAPI{
				Hostname:       "test",
				IntegrationKey: "test",
				SecretKey:      "test",
			},
		},
		{
			desc: "ShouldDetectMissingSecretKey",
			have: &schema.Configuration{DuoAPI: schema.DuoAPI{
				Hostname:       "test",
				IntegrationKey: "test",
			}},
			expected: schema.DuoAPI{
				Hostname:       "test",
				IntegrationKey: "test",
			},
			errs: []string{
				"duo_api: option 'secret_key' is required when duo is enabled but it's absent",
			},
		},
		{
			desc: "ShouldDetectMissingIntegrationKey",
			have: &schema.Configuration{DuoAPI: schema.DuoAPI{
				Hostname:  "test",
				SecretKey: "test",
			}},
			expected: schema.DuoAPI{
				Hostname:  "test",
				SecretKey: "test",
			},
			errs: []string{
				"duo_api: option 'integration_key' is required when duo is enabled but it's absent",
			},
		},
		{
			desc: "ShouldDetectMissingHostname",
			have: &schema.Configuration{DuoAPI: schema.DuoAPI{
				IntegrationKey: "test",
				SecretKey:      "test",
			}},
			expected: schema.DuoAPI{
				IntegrationKey: "test",
				SecretKey:      "test",
			},
			errs: []string{
				"duo_api: option 'hostname' is required when duo is enabled but it's absent",
			},
		},
		{
			desc: "ShouldWarnDisableCAPinningThirdPartyHostname",
			have: &schema.Configuration{DuoAPI: schema.DuoAPI{
				Hostname:         "push.example.com",
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			}},
			expected: schema.DuoAPI{
				Hostname:         "push.example.com",
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			},
			warns: []string{
				"duo_api: option 'experimental_disable_ca_pinning' is enabled: this option is experimental, insecure, not recommended, and may be removed at any time",
			},
		},
		{
			desc: "ShouldErrorDisableCAPinningDuoSecurityHostname",
			have: &schema.Configuration{DuoAPI: schema.DuoAPI{
				Hostname:         "api-123456789.duosecurity.com",
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			}},
			expected: schema.DuoAPI{
				Hostname:         "api-123456789.duosecurity.com",
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			},
			errs: []string{
				"duo_api: option 'experimental_disable_ca_pinning' must be false when the option 'hostname' is a Duo hostname but it's configured as 'api-123456789.duosecurity.com'",
			},
		},
		{
			desc: "ShouldErrorDisableCAPinningDuoFederalHostnameMixedCaseFQDN",
			have: &schema.Configuration{DuoAPI: schema.DuoAPI{
				Hostname:         "API-123456789.DuoFederal.com.",
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			}},
			expected: schema.DuoAPI{
				Hostname:         "API-123456789.DuoFederal.com.",
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			},
			errs: []string{
				"duo_api: option 'experimental_disable_ca_pinning' must be false when the option 'hostname' is a Duo hostname but it's configured as 'API-123456789.DuoFederal.com.'",
			},
		},
		{
			desc: "ShouldErrorDisableCAPinningDuoApexHostname",
			have: &schema.Configuration{DuoAPI: schema.DuoAPI{
				Hostname:         "duosecurity.com",
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			}},
			expected: schema.DuoAPI{
				Hostname:         "duosecurity.com",
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			},
			errs: []string{
				"duo_api: option 'experimental_disable_ca_pinning' must be false when the option 'hostname' is a Duo hostname but it's configured as 'duosecurity.com'",
			},
		},
		{
			desc: "ShouldWarnDisableCAPinningLookalikeHostname",
			have: &schema.Configuration{DuoAPI: schema.DuoAPI{
				Hostname:         "duosecurity.com.example.com",
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			}},
			expected: schema.DuoAPI{
				Hostname:         "duosecurity.com.example.com",
				IntegrationKey:   "test",
				SecretKey:        "test",
				DisableCAPinning: true,
			},
			warns: []string{
				"duo_api: option 'experimental_disable_ca_pinning' is enabled: this option is experimental, insecure, not recommended, and may be removed at any time",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			validator := schema.NewStructValidator()

			ValidateDuo(tc.have, validator)

			assert.Equal(t, tc.expected.Disable, tc.have.DuoAPI.Disable)
			assert.Equal(t, tc.expected.Hostname, tc.have.DuoAPI.Hostname)
			assert.Equal(t, tc.expected.IntegrationKey, tc.have.DuoAPI.IntegrationKey)
			assert.Equal(t, tc.expected.SecretKey, tc.have.DuoAPI.SecretKey)
			assert.Equal(t, tc.expected.EnableSelfEnrollment, tc.have.DuoAPI.EnableSelfEnrollment)
			assert.Equal(t, tc.expected.DisableCAPinning, tc.have.DuoAPI.DisableCAPinning)

			require.Len(t, validator.Warnings(), len(tc.warns))

			for i, warn := range tc.warns {
				assert.EqualError(t, validator.Warnings()[i], warn)
			}

			require.Len(t, validator.Errors(), len(tc.errs))

			if len(tc.errs) != 0 {
				for i, err := range tc.errs {
					t.Run(fmt.Sprintf("Err%d", i+1), func(t *testing.T) {
						assert.EqualError(t, validator.Errors()[i], err)
					})
				}
			}
		})
	}
}
