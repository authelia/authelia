package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
)

func TestValidatePrivacyPolicy(t *testing.T) {
	testCases := []struct {
		name     string
		have     *schema.PrivacyPolicy
		expected string
	}{
		{"ShouldValidateDefaultConfig", &schema.PrivacyPolicy{}, ""},
		{"ShouldValidateValidEnabledPolicy", &schema.PrivacyPolicy{Enabled: true, PolicyURL: MustParseURL("https://example.com/privacy")}, ""},
		{"ShouldValidateValidEnabledPolicyWithUserAcceptance", &schema.PrivacyPolicy{Enabled: true, RequireUserAcceptance: true, PolicyURL: MustParseURL("https://example.com/privacy")}, ""},
		{"ShouldNotValidateOnInvalidScheme", &schema.PrivacyPolicy{Enabled: true, PolicyURL: MustParseURL("http://example.com/privacy")}, "privacy_policy: option 'policy_url' must have the 'https' scheme but it's configured as 'http'"},
		{"ShouldNotValidateOnMissingURL", &schema.PrivacyPolicy{Enabled: true}, "privacy_policy: option 'policy_url' must be provided when the option 'enabled' is true"},
	}

	validator := schema.NewStructValidator()

	for _, tc := range testCases {
		validator.Clear()

		t.Run(tc.name, func(t *testing.T) {
			ValidatePrivacyPolicy(tc.have, validator)

			assert.Len(t, validator.Warnings(), 0)

			if tc.expected == "" {
				assert.Len(t, validator.Errors(), 0)
			} else {
				assert.EqualError(t, validator.Errors()[0], tc.expected)
			}
		})
	}
}
