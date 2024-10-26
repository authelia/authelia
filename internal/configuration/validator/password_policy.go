package validator

import (
	"errors"
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

// ValidatePasswordPolicy validates and updates the Password Policy configuration.
func ValidatePasswordPolicy(config *schema.PasswordPolicy, validator *schema.StructValidator) {
	if !utils.IsBoolCountLessThanN(1, true, config.Standard.Enabled, config.ZXCVBN.Enabled) {
		validator.Push(errors.New(errPasswordPolicyMultipleDefined))
	}

	if config.Standard.Enabled {
		if config.Standard.MinLength == 0 {
			config.Standard.MinLength = schema.DefaultPasswordPolicyConfiguration.Standard.MinLength
		} else if config.Standard.MinLength < 0 {
			validator.Push(fmt.Errorf(errFmtPasswordPolicyStandardMinLengthNotGreaterThanZero, config.Standard.MinLength))
		}

		if config.Standard.MaxLength == 0 {
			config.Standard.MaxLength = schema.DefaultPasswordPolicyConfiguration.Standard.MaxLength
		}
	}

	if config.ZXCVBN.Enabled {
		switch {
		case config.ZXCVBN.MinScore == 0:
			config.ZXCVBN.MinScore = schema.DefaultPasswordPolicyConfiguration.ZXCVBN.MinScore
		case config.ZXCVBN.MinScore < 0, config.ZXCVBN.MinScore > 4:
			validator.Push(fmt.Errorf(errFmtPasswordPolicyZXCVBNMinScoreInvalid, config.ZXCVBN.MinScore))
		}
	}
}
