package templates

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsSecretEnvKey(t *testing.T) {
	testCases := []struct {
		name     string
		have     []string
		expected bool
	}{
		{"ShouldReturnFalseForKeysWithoutPrefix", []string{"A_KEY", "A_SECRET", "A_PASSWORD", "NOT_AUTHELIA_A_PASSWORD"}, false},
		{"ShouldReturnFalseForKeysWithoutSuffix", []string{"AUTHELIA_EXAMPLE", "X_AUTHELIA_EXAMPLE", "X_AUTHELIA_PASSWORD_NOT"}, false},
		{"ShouldReturnTrueForSecretKeys", []string{"AUTHELIA_JWT_SECRET", "AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET", "AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_CERTIFICATE_CHAIN", "X_AUTHELIA_JWT_SECRET", "X_AUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET", "X_AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_CERTIFICATE_CHAIN"}, true},
		{"ShouldReturnTrueForSecretKeysEvenWithMixedCase", []string{"aUTHELIA_JWT_SECRET", "aUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET", "aUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_CERTIFICATE_CHAIN", "X_aUTHELIA_JWT_SECREt", "X_aUTHELIA_IDENTITY_PROVIDERS_OIDC_HMAC_SECRET", "x_AUTHELIA_IDENTITY_PROVIDERS_OIDC_ISSUER_CERTIFICATE_CHAIN"}, true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			for _, env := range tc.have {
				t.Run(env, func(t *testing.T) {
					assert.Equal(t, tc.expected, isSecretEnvKey(env))
				})
			}
		})
	}
}
