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
		expected schema.DuoAPI
		errs     []string
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
