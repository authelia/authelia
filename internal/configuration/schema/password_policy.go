// SPDX-FileCopyrightText: 2019 Authelia
//
// SPDX-License-Identifier: Apache-2.0

package schema

// PasswordPolicyStandardParams represents the configuration related to standard parameters of password policy.
type PasswordPolicyStandardParams struct {
	Enabled          bool `koanf:"enabled"`
	MinLength        int  `koanf:"min_length"`
	MaxLength        int  `koanf:"max_length"`
	RequireUppercase bool `koanf:"require_uppercase"`
	RequireLowercase bool `koanf:"require_lowercase"`
	RequireNumber    bool `koanf:"require_number"`
	RequireSpecial   bool `koanf:"require_special"`
}

// PasswordPolicyZXCVBNParams represents the configuration related to ZXCVBN parameters of password policy.
type PasswordPolicyZXCVBNParams struct {
	Enabled  bool `koanf:"enabled"`
	MinScore int  `koanf:"min_score"`
}

// PasswordPolicyConfiguration represents the configuration related to password policy.
type PasswordPolicyConfiguration struct {
	Standard PasswordPolicyStandardParams `koanf:"standard"`
	ZXCVBN   PasswordPolicyZXCVBNParams   `koanf:"zxcvbn"`
}

// DefaultPasswordPolicyConfiguration is the default password policy configuration.
var DefaultPasswordPolicyConfiguration = PasswordPolicyConfiguration{
	Standard: PasswordPolicyStandardParams{
		Enabled:   false,
		MinLength: 8,
		MaxLength: 0,
	},
	ZXCVBN: PasswordPolicyZXCVBNParams{
		Enabled:  false,
		MinScore: 3,
	},
}
