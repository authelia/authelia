package schema

import (
	"net/url"
)

// PrivacyPolicy is the privacy policy configuration.
type PrivacyPolicy struct {
	Enabled               bool     `koanf:"enabled" yaml:"enabled" toml:"enabled" json:"enabled" jsonschema:"default=false,title=Enabled" jsonschema_description:"Enables the Privacy Policy functionality."`
	RequireUserAcceptance bool     `koanf:"require_user_acceptance" yaml:"require_user_acceptance" toml:"require_user_acceptance" json:"require_user_acceptance" jsonschema:"default=false,title=Require User Acceptance" jsonschema_description:"Enables the requirement for users to accept the policy."`
	PolicyURL             *url.URL `koanf:"policy_url" yaml:"policy_url,omitempty" toml:"policy_url,omitempty" json:"policy_url,omitempty" jsonschema:"title=Policy URL" jsonschema_description:"The URL of the privacy policy."`
}
