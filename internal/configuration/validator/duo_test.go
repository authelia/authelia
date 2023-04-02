// SPDX-FileCopyrightText: 2019 Authelia
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

func TestValidateDuo(t *testing.T) {
	testCases := []struct {
		desc     string
		have     *schema.Configuration
		expected schema.DuoAPIConfiguration
		errs     []string
	}{
		{
			desc:     "ShouldDisableDuo",
			have:     &schema.Configuration{},
			expected: schema.DuoAPIConfiguration{Disable: true},
		},
		{
			desc: "ShouldNotDisableDuo",
			have: &schema.Configuration{DuoAPI: schema.DuoAPIConfiguration{
				Hostname:       "test",
				IntegrationKey: "test",
				SecretKey:      "test",
			}},
			expected: schema.DuoAPIConfiguration{
				Hostname:       "test",
				IntegrationKey: "test",
				SecretKey:      "test",
			},
		},
		{
			desc: "ShouldDetectMissingSecretKey",
			have: &schema.Configuration{DuoAPI: schema.DuoAPIConfiguration{
				Hostname:       "test",
				IntegrationKey: "test",
			}},
			expected: schema.DuoAPIConfiguration{
				Hostname:       "test",
				IntegrationKey: "test",
			},
			errs: []string{
				"duo_api: option 'secret_key' is required when duo is enabled but it is missing",
			},
		},
		{
			desc: "ShouldDetectMissingIntegrationKey",
			have: &schema.Configuration{DuoAPI: schema.DuoAPIConfiguration{
				Hostname:  "test",
				SecretKey: "test",
			}},
			expected: schema.DuoAPIConfiguration{
				Hostname:  "test",
				SecretKey: "test",
			},
			errs: []string{
				"duo_api: option 'integration_key' is required when duo is enabled but it is missing",
			},
		},
		{
			desc: "ShouldDetectMissingHostname",
			have: &schema.Configuration{DuoAPI: schema.DuoAPIConfiguration{
				IntegrationKey: "test",
				SecretKey:      "test",
			}},
			expected: schema.DuoAPIConfiguration{
				IntegrationKey: "test",
				SecretKey:      "test",
			},
			errs: []string{
				"duo_api: option 'hostname' is required when duo is enabled but it is missing",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			val := schema.NewStructValidator()

			ValidateDuo(tc.have, val)

			assert.Equal(t, tc.expected.Disable, tc.have.DuoAPI.Disable)
			assert.Equal(t, tc.expected.Hostname, tc.have.DuoAPI.Hostname)
			assert.Equal(t, tc.expected.IntegrationKey, tc.have.DuoAPI.IntegrationKey)
			assert.Equal(t, tc.expected.SecretKey, tc.have.DuoAPI.SecretKey)
			assert.Equal(t, tc.expected.EnableSelfEnrollment, tc.have.DuoAPI.EnableSelfEnrollment)

			require.Len(t, val.Errors(), len(tc.errs))

			if len(tc.errs) != 0 {
				for i, err := range tc.errs {
					t.Run(fmt.Sprintf("Err%d", i+1), func(t *testing.T) {
						assert.EqualError(t, val.Errors()[i], err)
					})
				}
			}
		})
	}
}
