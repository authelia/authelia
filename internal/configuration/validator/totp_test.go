package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidateTOTP(t *testing.T) {
	testCases := []struct {
		desc     string
		have     schema.TOTPConfiguration
		expected schema.TOTPConfiguration
		errs     []string
		warns    []string
	}{
		{
			desc:     "ShouldSetDefaultTOTPValues",
			expected: schema.DefaultTOTPConfiguration,
		},
		{
			desc:     "ShouldNotSetDefaultTOTPValuesWhenDisabled",
			have:     schema.TOTPConfiguration{Disable: true},
			expected: schema.TOTPConfiguration{Disable: true},
		},
		{
			desc: "ShouldNormalizeTOTPAlgorithm",
			have: schema.TOTPConfiguration{
				DefaultAlgorithm: digestSHA1,
				DefaultDigits:    6,
				DefaultPeriod:    30,
				SecretSize:       32,
				Skew:             schema.DefaultTOTPConfiguration.Skew,
				Issuer:           "abc",
			},
			expected: schema.TOTPConfiguration{
				DefaultAlgorithm: "SHA1",
				DefaultDigits:    6,
				DefaultPeriod:    30,
				SecretSize:       32,
				Skew:             schema.DefaultTOTPConfiguration.Skew,
				Issuer:           "abc",
			},
		},
		{
			desc: "ShouldRaiseErrorWhenInvalidTOTPAlgorithm",
			have: schema.TOTPConfiguration{
				DefaultAlgorithm: "sha3",
				DefaultDigits:    6,
				DefaultPeriod:    30,
				SecretSize:       32,
				Skew:             schema.DefaultTOTPConfiguration.Skew,
				Issuer:           "abc",
			},
			errs: []string{
				"totp: option 'algorithm' must be one of 'SHA1', 'SHA256', or 'SHA512' but it's configured as 'SHA3'",
			},
		},
		{
			desc: "ShouldRaiseErrorWhenInvalidTOTPValue",
			have: schema.TOTPConfiguration{
				DefaultAlgorithm: "sha3",
				DefaultPeriod:    5,
				DefaultDigits:    20,
				SecretSize:       10,
				Skew:             schema.DefaultTOTPConfiguration.Skew,
				Issuer:           "abc",
			},
			errs: []string{
				"totp: option 'algorithm' must be one of 'SHA1', 'SHA256', or 'SHA512' but it's configured as 'SHA3'",
				"totp: option 'period' option must be 15 or more but it's configured as '5'",
				"totp: option 'digits' must be 6 or 8 but it's configured as '20'",
				"totp: option 'secret_size' must be 20 or higher but it's configured as '10'",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			validator := schema.NewStructValidator()
			config := &schema.Configuration{TOTP: tc.have}

			ValidateTOTP(config, validator)

			errs := validator.Errors()
			warns := validator.Warnings()

			if len(tc.errs) == 0 {
				assert.Len(t, errs, 0)
				assert.Len(t, warns, 0)
				assert.Equal(t, tc.expected.Disable, config.TOTP.Disable)
				assert.Equal(t, tc.expected.Issuer, config.TOTP.Issuer)
				assert.Equal(t, tc.expected.DefaultAlgorithm, config.TOTP.DefaultAlgorithm)
				assert.Equal(t, tc.expected.Skew, config.TOTP.Skew)
				assert.Equal(t, tc.expected.DefaultPeriod, config.TOTP.DefaultPeriod)
				assert.Equal(t, tc.expected.SecretSize, config.TOTP.SecretSize)
			} else {
				expectedErrs := len(tc.errs)

				require.Len(t, errs, expectedErrs)

				for i := 0; i < expectedErrs; i++ {
					t.Run(fmt.Sprintf("Err%d", i+1), func(t *testing.T) {
						assert.EqualError(t, errs[i], tc.errs[i])
					})
				}
			}

			expectedWarns := len(tc.warns)
			require.Len(t, warns, expectedWarns)

			for i := 0; i < expectedWarns; i++ {
				t.Run(fmt.Sprintf("Err%d", i+1), func(t *testing.T) {
					assert.EqualError(t, warns[i], tc.warns[i])
				})
			}
		})
	}
}
