package schema

// PasswordPolicyConfiguration represents the password policy configuration.
type PasswordPolicyConfiguration struct {
	// Possible Modes: 'none', 'zxcvbn', 'classic', 'ldap'
	// 	none: doesn't provide any password policy
	// 	zxcvbn: uses zxcvbn to get the password strength. this option enables the MinScore param
	// 	classic: uses classic rules (i.e. Require LowerCases, Require Uppercases, etc)
	// 	ldap: uses classic rules, but the rules are fetched from ldap
	Mode      string `koanf:"mode"`
	MinLength int    `koanf:"min_length"`
	// MinScore set the minimal acceptable score for mode 'zxcvbn'
	MinScore         int  `koanf:"min_score"`
	RequireUppercase bool `koanf:"require_uppercase"`
	RequireLowercase bool `koanf:"require_lowercase"`
	RequireNumber    bool `koanf:"require_number"`
	RequireSpecial   bool `koanf:"require_special"`
}

// DefaultPasswordPolicyConfiguration is the default password policy configuration.
var DefaultPasswordPolicyConfiguration = PasswordPolicyConfiguration{
	Mode:      "classic",
	MinLength: 0,
	MinScore:  0,
}
