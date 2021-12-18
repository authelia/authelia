package validator

import (
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidatePasswordPolicy validates and update Password Policy configuration.
func ValidatePasswordPolicy(configuration *schema.PasswordPolicyConfiguration, validator *schema.StructValidator) {
	// if configuration.Mode == "" {
	// 	configuration.Mode = schema.DefaultPasswordPolicyConfiguration.Mode
	// }

	switch configuration.Mode {
	case "":
		configuration.MinScore = 0
		configuration.MinLength = 0
		configuration.RequireLowercase = false
		configuration.RequireUppercase = false
		configuration.RequireSpecial = false
		configuration.RequireNumber = false
	case "zxcvbn":
		if configuration.MinScore == 0 {
			configuration.MinScore = schema.DefaultPasswordPolicyConfiguration.MinScore
		} else if configuration.MinScore < 0 || configuration.MinScore > 4 {
			validator.Push(fmt.Errorf("password_policy: min_score must be between 0 and 4"))
		}

		configuration.MinLength = 0
		configuration.RequireLowercase = false
		configuration.RequireUppercase = false
		configuration.RequireSpecial = false
		configuration.RequireNumber = false
	case "classic":
		if configuration.MinLength == 0 {
			configuration.MinLength = schema.DefaultPasswordPolicyConfiguration.MinLength
		} else if configuration.MinLength < 0 {
			validator.Push(fmt.Errorf("password_policy: min_length must be >= 0"))
		}

		configuration.MinScore = 0
	}
}
