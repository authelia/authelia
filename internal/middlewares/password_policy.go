package middlewares

import (
	"regexp"

	"github.com/trustelem/zxcvbn"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// PasswordPolicyProvider represents an implementation of a password policy provider.
type PasswordPolicyProvider interface {
	Check(password string) (err error)
}

// NewPasswordPolicyProvider returns a new password policy provider.
func NewPasswordPolicyProvider(config schema.PasswordPolicy) (provider PasswordPolicyProvider) {
	if !config.Standard.Enabled && !config.ZXCVBN.Enabled {
		return &StandardPasswordPolicyProvider{}
	}

	if config.Standard.Enabled {
		p := &StandardPasswordPolicyProvider{}

		p.min, p.max = config.Standard.MinLength, config.Standard.MaxLength

		if config.Standard.RequireLowercase {
			p.patterns = append(p.patterns, *regexp.MustCompile(`[a-z]+`))
		}

		if config.Standard.RequireUppercase {
			p.patterns = append(p.patterns, *regexp.MustCompile(`[A-Z]+`))
		}

		if config.Standard.RequireNumber {
			p.patterns = append(p.patterns, *regexp.MustCompile(`[0-9]+`))
		}

		if config.Standard.RequireSpecial {
			p.patterns = append(p.patterns, *regexp.MustCompile(`[^a-zA-Z0-9]+`))
		}

		return p
	}

	if config.ZXCVBN.Enabled {
		return &ZXCVBNPasswordPolicyProvider{minScore: config.ZXCVBN.MinScore}
	}

	return &StandardPasswordPolicyProvider{}
}

// ZXCVBNPasswordPolicyProvider handles zxcvbn password policy checking.
type ZXCVBNPasswordPolicyProvider struct {
	minScore int
}

// Check checks the password against the policy.
func (p ZXCVBNPasswordPolicyProvider) Check(password string) (err error) {
	result := zxcvbn.PasswordStrength(password, nil)

	if result.Score < p.minScore {
		return errPasswordPolicyNoMet
	}

	return nil
}

// StandardPasswordPolicyProvider handles standard password policy checking.
type StandardPasswordPolicyProvider struct {
	patterns []regexp.Regexp
	min, max int
}

// Check checks the password against the policy.
func (p StandardPasswordPolicyProvider) Check(password string) (err error) {
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
