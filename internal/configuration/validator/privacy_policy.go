package validator

import (
	"errors"
	"fmt"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

// ValidatePrivacyPolicy validates and updates the Privacy Policy configuration.
func ValidatePrivacyPolicy(config *schema.PrivacyPolicy, validator *schema.StructValidator) {
	if !config.Enabled {
		return
	}

	switch config.PolicyURL {
	case nil:
		validator.Push(errors.New(errPrivacyPolicyEnabledWithoutURL))
	default:
		if config.PolicyURL.Scheme != schemeHTTPS {
			validator.Push(fmt.Errorf(errFmtPrivacyPolicyURLNotHTTPS, config.PolicyURL.Scheme))
		}
	}
}
