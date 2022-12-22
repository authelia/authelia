package schema

import (
	"net/url"
)

type PrivacyPolicy struct {
	Enable                bool     `koanf:"enable"`
	RequireUserAcceptance bool     `koanf:"require_user_acceptance"`
	PolicyURL             *url.URL `koanf:"policy_url"`
}
