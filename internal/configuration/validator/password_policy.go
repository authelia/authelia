package validator

import (
	"errors"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidatePasswordPolicy validates and update Password Policy configuration.
func ValidatePasswordPolicy(configuration *schema.PasswordPolicyConfiguration, validator *schema.StructValidator) {
	if !utils.IsBoolCountLessThanN(1, true, configuration.Standard.Enabled, configuration.Zxcvbn.Enabled) {
		validator.Push(errors.New("password_policy:only one password policy can be enabled at a time"))
	}

	if configuration.Standard.Enabled {
		if configuration.Standard.MinLength == 0 {
			configuration.Standard.MinLength = schema.DefaultPasswordPolicyConfiguration.Standard.MinLength
		} else if configuration.Standard.MinLength < 0 {
			validator.Push(errors.New("password_policy: min_length must be > 0"))
		}

		if configuration.Standard.MaxLength == 0 {
			configuration.Standard.MaxLength = schema.DefaultPasswordPolicyConfiguration.Standard.MaxLength
		}
	} else if configuration.Zxcvbn.Enabled {
		if configuration.Zxcvbn.MinScore == 0 {
			configuration.Zxcvbn.MinScore = schema.DefaultPasswordPolicyConfiguration.Zxcvbn.MinScore
		} else if configuration.Zxcvbn.MinScore < 0 || configuration.Zxcvbn.MinScore > 4 {
			validator.Push(errors.New("min_score must be between 0 and 4"))
		}
	}
}
