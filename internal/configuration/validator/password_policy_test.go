package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidatePasswordPolicy(t *testing.T) {
	testCases := []struct {
		desc           string
		have, expected *schema.PasswordPolicy
		expectedErrs   []string
	}{
		{
			desc: "ShouldRaiseErrorsWhenMisconfigured",
			have: &schema.PasswordPolicy{
				Standard: schema.PasswordPolicyStandard{
					Enabled:   true,
					MinLength: -1,
				},
				ZXCVBN: schema.PasswordPolicyZXCVBN{
					Enabled: true,
				},
			},
			expected: &schema.PasswordPolicy{
				Standard: schema.PasswordPolicyStandard{
					Enabled:   true,
					MinLength: -1,
				},
				ZXCVBN: schema.PasswordPolicyZXCVBN{
					Enabled:  true,
					MinScore: 3,
				},
			},
			expectedErrs: []string{
				"password_policy: only a single password policy mechanism can be specified",
				"password_policy: standard: option 'min_length' must be greater than 0 but it's configured as -1",
			},
		},
		{
			desc: "ShouldNotRaiseErrorsStandard",
			have: &schema.PasswordPolicy{
				Standard: schema.PasswordPolicyStandard{
					Enabled:   true,
					MinLength: 8,
				},
			},
			expected: &schema.PasswordPolicy{
				Standard: schema.PasswordPolicyStandard{
					Enabled:   true,
					MinLength: 8,
				},
			},
		},
		{
			desc: "ShouldNotRaiseErrorsZXCVBN",
			have: &schema.PasswordPolicy{
				ZXCVBN: schema.PasswordPolicyZXCVBN{
					Enabled: true,
				},
			},
			expected: &schema.PasswordPolicy{
				ZXCVBN: schema.PasswordPolicyZXCVBN{
					Enabled:  true,
					MinScore: 3,
				},
			},
		},
		{
			desc: "ShouldSetDefaultstandard",
			have: &schema.PasswordPolicy{
				Standard: schema.PasswordPolicyStandard{
					Enabled:   true,
					MinLength: 0,
				},
			},
			expected: &schema.PasswordPolicy{
				Standard: schema.PasswordPolicyStandard{
					Enabled:   true,
					MinLength: 8,
				},
			},
		},
		{
			desc: "ShouldRaiseErrorsZXCVBNTooLow",
			have: &schema.PasswordPolicy{
				ZXCVBN: schema.PasswordPolicyZXCVBN{
					Enabled:  true,
					MinScore: -1,
				},
			},
			expected: &schema.PasswordPolicy{
				ZXCVBN: schema.PasswordPolicyZXCVBN{
					Enabled:  true,
					MinScore: -1,
				},
			},
			expectedErrs: []string{
				"password_policy: zxcvbn: option 'min_score' is invalid: must be between 1 and 4 but it's configured as -1",
			},
		},
		{
			desc: "ShouldRaiseErrorsZXCVBNTooHigh",
			have: &schema.PasswordPolicy{
				ZXCVBN: schema.PasswordPolicyZXCVBN{
					Enabled:  true,
					MinScore: 5,
				},
			},
			expected: &schema.PasswordPolicy{
				ZXCVBN: schema.PasswordPolicyZXCVBN{
					Enabled:  true,
					MinScore: 5,
				},
			},
			expectedErrs: []string{
				"password_policy: zxcvbn: option 'min_score' is invalid: must be between 1 and 4 but it's configured as 5",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			validator := &schema.StructValidator{}
			ValidatePasswordPolicy(tc.have, validator)

			assert.Len(t, validator.Warnings(), 0)

			assert.Equal(t, tc.expected.Standard.MaxLength, tc.have.Standard.MaxLength)
			assert.Equal(t, tc.expected.Standard.MinLength, tc.have.Standard.MinLength)
			assert.Equal(t, tc.expected.Standard.RequireNumber, tc.have.Standard.RequireNumber)
			assert.Equal(t, tc.expected.Standard.RequireSpecial, tc.have.Standard.RequireSpecial)
			assert.Equal(t, tc.expected.Standard.RequireUppercase, tc.have.Standard.RequireUppercase)
			assert.Equal(t, tc.expected.Standard.RequireLowercase, tc.have.Standard.RequireLowercase)
			assert.Equal(t, tc.expected.ZXCVBN.MinScore, tc.have.ZXCVBN.MinScore)

			errs := validator.Errors()
			require.Len(t, errs, len(tc.expectedErrs))

			for i := 0; i < len(errs); i++ {
				t.Run(fmt.Sprintf("Err%d", i+1), func(t *testing.T) {
					assert.EqualError(t, errs[i], tc.expectedErrs[i])
				})
			}
		})
	}
}
