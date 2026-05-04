package validator

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration/schema"
	"github.com/authelia/authelia/v4/internal/utils"
)

func TestShouldValidateGoodKeys(t *testing.T) {
	configKeys := schema.Keys
	validator := schema.NewStructValidator()
	ValidateKeys(configKeys, nil, "AUTHELIA_", validator)

	require.Len(t, validator.Errors(), 0)
}

func TestShouldNotValidateBadKeys(t *testing.T) {
	configKeys := schema.Keys
	configKeys = append(configKeys, "bad_key")
	configKeys = append(configKeys, "totp.skewy")
	validator := schema.NewStructValidator()
	ValidateKeys(configKeys, nil, "AUTHELIA_", validator)

	errs := validator.Errors()
	require.Len(t, errs, 2)

	assert.EqualError(t, errs[0], "configuration key not expected: bad_key")
	assert.EqualError(t, errs[1], "configuration key not expected: totp.skewy")
}

func TestShouldNotValidateBadEnvKeys(t *testing.T) {
	configKeys := schema.Keys
	configKeys = append(configKeys, "AUTHELIA__BAD_ENV_KEY")
	configKeys = append(configKeys, "AUTHELIA_BAD_ENV_KEY")

	validator := schema.NewStructValidator()
	ValidateKeys(configKeys, nil, "AUTHELIA_", validator)

	warns := validator.Warnings()
	assert.Len(t, validator.Errors(), 0)
	require.Len(t, warns, 2)

	assert.EqualError(t, warns[0], "configuration environment variable not expected: AUTHELIA__BAD_ENV_KEY")
	assert.EqualError(t, warns[1], "configuration environment variable not expected: AUTHELIA_BAD_ENV_KEY")
}

func TestAllSpecificErrorKeys(t *testing.T) {
	var configKeys []string //nolint:prealloc // This is because the test is dynamic based on the keys that exist in the map.

	var uniqueValues []string

	// Setup configKeys and uniqueValues expected.
	for key, value := range specificErrorKeys {
		configKeys = append(configKeys, key)

		if !utils.IsStringInSlice(value, uniqueValues) {
			uniqueValues = append(uniqueValues, value)
		}
	}

	validator := schema.NewStructValidator()
	ValidateKeys(configKeys, nil, "AUTHELIA_", validator)

	errs := validator.Errors()

	// Check only unique errors are shown. Require because if we don't the next test panics.
	require.Len(t, errs, len(uniqueValues))

	// Dynamically check all specific errors.
	for i, value := range uniqueValues {
		assert.EqualError(t, errs[i], value)
	}
}

func TestSpecificErrorKeys(t *testing.T) {
	configKeys := []string{
		"notifier.smtp.trusted_cert",
		"google_analytics",
		"authentication_backend.file.password_options.algorithm",
		"authentication_backend.file.password_options.iterations", // This should not show another error since our target for the specific error is password_options.
		"authentication_backend.file.password_hashing.algorithm",
		"authentication_backend.file.hashing.algorithm",
	}

	validator := schema.NewStructValidator()
	ValidateKeys(configKeys, nil, "AUTHELIA_", validator)

	errs := validator.Errors()

	require.Len(t, errs, 5)

	assert.EqualError(t, errs[0], specificErrorKeys["notifier.smtp.trusted_cert"])
	assert.EqualError(t, errs[1], specificErrorKeys["google_analytics"])
	assert.EqualError(t, errs[2], specificErrorKeys["authentication_backend.file.password_options.iterations"])
	assert.EqualError(t, errs[3], specificErrorKeys["authentication_backend.file.password_hashing.algorithm"])
	assert.EqualError(t, errs[4], specificErrorKeys["authentication_backend.file.hashing.algorithm"])
}

func TestPatternKeys(t *testing.T) {
	configKeys := []string{
		"server.endpoints.authz.xx.implementation",
		"server.endpoints.authz.x.implementation",
	}

	validator := schema.NewStructValidator()
	ValidateKeys(configKeys, nil, "AUTHELIA_", validator)

	errs := validator.Errors()

	require.Len(t, errs, 0)
}

func TestPatternKeysWithDottedMapKeySegments(t *testing.T) {
	testCases := []struct {
		name string
		keys []string
	}{
		{
			"ScopeWithDot",
			[]string{
				"identity_providers.oidc.scopes.my.scope",
				"identity_providers.oidc.scopes.my.scope.claims",
			},
		},
		{
			"ClaimsPolicyWithDot",
			[]string{
				"identity_providers.oidc.claims_policies.my.policy",
				"identity_providers.oidc.claims_policies.my.policy.id_token",
				"identity_providers.oidc.claims_policies.my.policy.custom_claims",
			},
		},
		{
			"CustomClaimWithDot",
			[]string{
				"identity_providers.oidc.claims_policies.test.custom_claims.http://example.com/claim",
				"identity_providers.oidc.claims_policies.test.custom_claims.http://example.com/claim.name",
				"identity_providers.oidc.claims_policies.test.custom_claims.http://example.com/claim.attribute",
			},
		},
		{
			"AuthorizationPolicyWithDot",
			[]string{
				"identity_providers.oidc.authorization_policies.my.policy",
				"identity_providers.oidc.authorization_policies.my.policy.default_policy",
			},
		},
		{
			"CustomLifespanWithDot",
			[]string{
				"identity_providers.oidc.lifespans.custom.my.lifespan",
				"identity_providers.oidc.lifespans.custom.my.lifespan.access_token",
			},
		},
		{
			"ServerEndpointWithDot",
			[]string{
				"server.endpoints.authz.my.endpoint",
				"server.endpoints.authz.my.endpoint.implementation",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			ValidateKeys(tc.keys, nil, "AUTHELIA_", validator)

			assert.Len(t, validator.Errors(), 0, "keys with dots in map key segments should be accepted: %v", tc.keys)
			assert.Len(t, validator.Warnings(), 0)
		})
	}
}

func TestPatternKeysWithSpecialCharactersInMapKeySegments(t *testing.T) {
	testCases := []struct {
		name string
		keys []string
	}{
		{
			"Colon",
			[]string{
				"identity_providers.oidc.scopes.urn:example:scope",
				"identity_providers.oidc.scopes.urn:example:scope.claims",
			},
		},
		{
			"ColonInURI",
			[]string{
				"identity_providers.oidc.claims_policies.test.custom_claims.http://example.com:8080/claim",
				"identity_providers.oidc.claims_policies.test.custom_claims.http://example.com:8080/claim.name",
			},
		},
		{
			"MultipleDots",
			[]string{
				"identity_providers.oidc.scopes.org.example.scope.v2",
				"identity_providers.oidc.scopes.org.example.scope.v2.claims",
			},
		},
		{
			"Tilde",
			[]string{
				"identity_providers.oidc.scopes.scope~draft",
				"identity_providers.oidc.scopes.scope~draft.claims",
			},
		},
		{
			"Slash",
			[]string{
				"identity_providers.oidc.scopes.org/scope",
				"identity_providers.oidc.scopes.org/scope.claims",
			},
		},
		{
			"MixedSpecialChars",
			[]string{
				"identity_providers.oidc.scopes.urn:ietf:params:oauth:scope:example.read",
				"identity_providers.oidc.scopes.urn:ietf:params:oauth:scope:example.read.claims",
			},
		},
		{
			"URNStyle",
			[]string{
				"identity_providers.oidc.scopes.urn:authelia:scope:pam",
				"identity_providers.oidc.scopes.urn:authelia:scope:pam.claims",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			validator := schema.NewStructValidator()
			ValidateKeys(tc.keys, nil, "AUTHELIA_", validator)

			assert.Len(t, validator.Errors(), 0, "keys with special characters in map key segments should be accepted: %v", tc.keys)
			assert.Len(t, validator.Warnings(), 0)
		})
	}
}

func TestReplacedErrors(t *testing.T) {
	configKeys := []string{
		"authentication_backend.ldap.skip_verify",
		"authentication_backend.ldap.minimum_tls_version",
		"notifier.smtp.disable_verify_cert",
		"logs_file_path",
		"logs_level",
	}

	validator := schema.NewStructValidator()
	ValidateKeys(configKeys, nil, "AUTHELIA_", validator)

	warns := validator.Warnings()
	errs := validator.Errors()

	assert.Len(t, warns, 0)
	require.Len(t, errs, 5)

	assert.EqualError(t, errs[0], fmt.Sprintf(errFmtReplacedConfigurationKey, "authentication_backend.ldap.skip_verify", "authentication_backend.ldap.tls.skip_verify"))
	assert.EqualError(t, errs[1], fmt.Sprintf(errFmtReplacedConfigurationKey, "authentication_backend.ldap.minimum_tls_version", "authentication_backend.ldap.tls.minimum_version"))
	assert.EqualError(t, errs[2], fmt.Sprintf(errFmtReplacedConfigurationKey, "notifier.smtp.disable_verify_cert", "notifier.smtp.tls.skip_verify"))
	assert.EqualError(t, errs[3], fmt.Sprintf(errFmtReplacedConfigurationKey, "logs_file_path", "log.file_path"))
	assert.EqualError(t, errs[4], fmt.Sprintf(errFmtReplacedConfigurationKey, "logs_level", "log.level"))
}
