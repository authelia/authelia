package schema

// PasswordPolicy represents the configuration related to password policy.
type PasswordPolicy struct {
	Standard PasswordPolicyStandard `koanf:"standard" yaml:"standard,omitempty" toml:"standard,omitempty" json:"standard,omitempty" jsonschema:"title=Standard" jsonschema_description:"The standard password policy engine."`
	ZXCVBN   PasswordPolicyZXCVBN   `koanf:"zxcvbn" yaml:"zxcvbn,omitempty" toml:"zxcvbn,omitempty" json:"zxcvbn,omitempty" jsonschema:"title=ZXCVBN" jsonschema_description:"The ZXCVBN password policy engine."`
}

// PasswordPolicyStandard represents the configuration related to standard parameters of password policy.
type PasswordPolicyStandard struct {
	Enabled          bool `koanf:"enabled" yaml:"enabled" toml:"enabled" json:"enabled" jsonschema:"default=false,title=Enabled" jsonschema_description:"Enables the standard password policy engine."`
	MinLength        int  `koanf:"min_length" yaml:"min_length" toml:"min_length" json:"min_length" jsonschema:"title=Minimum Length" jsonschema_description:"Minimum password length."`
	MaxLength        int  `koanf:"max_length" yaml:"max_length" toml:"max_length" json:"max_length" jsonschema:"default=8,title=Maximum Length" jsonschema_description:"Maximum password length."`
	RequireUppercase bool `koanf:"require_uppercase" yaml:"require_uppercase" toml:"require_uppercase" json:"require_uppercase" jsonschema:"default=false,title=Require Uppercase" jsonschema_description:"Require uppercase characters."`
	RequireLowercase bool `koanf:"require_lowercase" yaml:"require_lowercase" toml:"require_lowercase" json:"require_lowercase" jsonschema:"default=false,title=Require Lowercase" jsonschema_description:"Require lowercase characters."`
	RequireNumber    bool `koanf:"require_number" yaml:"require_number" toml:"require_number" json:"require_number" jsonschema:"default=false,title=Require Number" jsonschema_description:"Require numeric characters."`
	RequireSpecial   bool `koanf:"require_special" yaml:"require_special" toml:"require_special" json:"require_special" jsonschema:"default=false,title=Require Special" jsonschema_description:"Require symbolic characters."`
}

// PasswordPolicyZXCVBN represents the configuration related to ZXCVBN parameters of password policy.
type PasswordPolicyZXCVBN struct {
	Enabled  bool `koanf:"enabled" yaml:"enabled" toml:"enabled" json:"enabled" jsonschema:"default=false,title=Enabled" jsonschema_description:"Enables the ZXCVBN password policy engine."`
	MinScore int  `koanf:"min_score" yaml:"min_score" toml:"min_score" json:"min_score" jsonschema:"default=3,title=Minimum Score" jsonschema_description:"The minimum ZXCVBN score allowed."`
}

// DefaultPasswordPolicyConfiguration is the default password policy configuration.
var DefaultPasswordPolicyConfiguration = PasswordPolicy{
	Standard: PasswordPolicyStandard{
		MinLength: 8,
		MaxLength: 0,
	},
	ZXCVBN: PasswordPolicyZXCVBN{
		MinScore: 3,
	},
}
