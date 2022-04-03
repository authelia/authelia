package middlewares

import (
	"regexp"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// NewPasswordPolicyProvider returns a new password policy provider.
func NewPasswordPolicyProvider(config schema.PasswordPolicyConfiguration) (provider PasswordPolicyProvider) {
	if !config.Standard.Enabled {
		return provider
	}

	provider.min, provider.max = config.Standard.MinLength, config.Standard.MaxLength

	if config.Standard.RequireLowercase {
		provider.patterns = append(provider.patterns, *regexp.MustCompile(`[a-z]+`))
	}

	if config.Standard.RequireUppercase {
		provider.patterns = append(provider.patterns, *regexp.MustCompile(`[A-Z]+`))
	}

	if config.Standard.RequireNumber {
		provider.patterns = append(provider.patterns, *regexp.MustCompile(`[0-9]+`))
	}

	if config.Standard.RequireSpecial {
		provider.patterns = append(provider.patterns, *regexp.MustCompile(`[^a-zA-Z0-9]+`))
	}

	return provider
}

// PasswordPolicyProvider handles password policy checking.
type PasswordPolicyProvider struct {
	patterns []regexp.Regexp
	min, max int
}

// Check checks the password against the policy.
func (p PasswordPolicyProvider) Check(password string) (err error) {
	patterns := len(p.patterns)

	if (p.min > 0 && len(password) < p.min) || (p.max > 0 && len(password) > p.max) {
		return errPasswordPolicyNoMet
	}

	if patterns == 0 {
		return nil
	}

	for i := 0; i < patterns; i++ {
		if !p.patterns[i].MatchString(password) {
			return errPasswordPolicyNoMet
		}
	}

	return nil
}
