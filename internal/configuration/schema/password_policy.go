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

// PasswordPolicyZxcvbnParams represents the configuration related to zxcvbn parameters of password policy.
type PasswordPolicyZxcvbnParams struct {
	Enabled  bool `koanf:"enabled"`
	MinScore int  `koanf:"min_score"`
}

// PasswordPolicyConfiguration represents the configuration related to password policy.
type PasswordPolicyConfiguration struct {
	Standard PasswordPolicyStandardParams `koanf:"standard"`
	Zxcvbn   PasswordPolicyZxcvbnParams   `koanf:"zxcvbn"`
}

// DefaultPasswordPolicyConfiguration is the default password policy configuration.
var DefaultPasswordPolicyConfiguration = PasswordPolicyConfiguration{
	Standard: PasswordPolicyStandardParams{
		Enabled:   false,
		MinLength: 8,
		MaxLength: 0,
	},
	Zxcvbn: PasswordPolicyZxcvbnParams{
		Enabled:  false,
		MinScore: 0,
	},
}
