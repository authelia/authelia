package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// PasswordPolicyNone represents password policy disable.
const PasswordPolicyNone = "none"

// ValidatePasswordPolicy validates and update Password Policy configuration.
func ValidatePasswordPolicy(configuration *schema.PasswordPolicyConfiguration, validator *schema.StructValidator) {
	switch configuration.Mode {
	case "", PasswordPolicyNone:
		configuration.Mode = PasswordPolicyNone
		configuration.MinScore = 0
		configuration.MinLength = 0
		configuration.RequireLowercase = false
		configuration.RequireUppercase = false
		configuration.RequireNumber = false
		configuration.RequireSpecial = false
	case "zxcvbn":
		if configuration.MinScore == 0 {
			configuration.MinScore = schema.DefaultPasswordPolicyConfiguration.MinScore
		} else if configuration.MinScore < 0 || configuration.MinScore > 4 {
			validator.Push(fmt.Errorf("password_policy: min_score must be between 0 and 4"))
		}

		configuration.MinLength = 0
		configuration.RequireLowercase = false
		configuration.RequireUppercase = false
		configuration.RequireNumber = false
		configuration.RequireSpecial = false
	case "classic":
		if configuration.MinLength == 0 {
			configuration.MinLength = schema.DefaultPasswordPolicyConfiguration.MinLength
		} else if configuration.MinLength < 0 {
			validator.Push(fmt.Errorf("password_policy: min_length must be > 0"))
		}

		configuration.MinScore = 0
	}
}
