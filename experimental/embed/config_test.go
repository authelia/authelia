package embed

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/authelia/authelia/v4/internal/configuration"
)

func TestNewConfiguration(t *testing.T) {
	testCases := []struct {
		name     string
		paths    []string
		filters  []configuration.BytesFilter
		keys     []string
		warnings []string
		errors   []string
		err      string
	}{
		{
			name:  "ShouldHandleWebAuthn",
			paths: []string{"../../internal/configuration/test_resources/config.webauthn.yml"},
			keys: []string{
				"regulation.max_retries",
				"server.endpoints.rate_limits.reset_password_finish.enable",
				"server.endpoints.rate_limits.reset_password_start.enable",
				"server.endpoints.rate_limits.second_factor_duo.enable",
				"server.endpoints.rate_limits.second_factor_totp.enable",
				"server.endpoints.rate_limits.session_elevation_finish.enable",
				"server.endpoints.rate_limits.session_elevation_start.enable",
				"webauthn.metadata.cache_policy",
				"webauthn.selection_criteria.attachment",
				"webauthn.selection_criteria.discoverability",
				"webauthn.selection_criteria.user_verification",
			},
			warnings: nil,
			errors: []string{
				"identity_validation: reset_password: option 'jwt_secret' is required when the reset password functionality isn't disabled",
				"authentication_backend: you must ensure either the 'file' or 'ldap' authentication backend is configured",
				"access_control: 'default_policy' option 'deny' is invalid: when no rules are specified it must be 'two_factor' or 'one_factor'",
				"session: option 'cookies' is required",
				"storage: option 'encryption_key' is required",
				"storage: configuration for a 'local', 'mysql' or 'postgres' database must be provided",
				"notifier: you must ensure either the 'smtp' or 'filesystem' notifier is configured",
			},
		},
		{
			name:  "ShouldHandleConfigWithDefinitions",
			paths: []string{"../../internal/configuration/test_resources/config_with_definitions.yml"},
			keys: []string{
				"access_control.default_policy",
				"access_control.networks",
				"access_control.networks[].name",
				"access_control.networks[].networks",
				"access_control.rules",
				"access_control.rules[].domain",
				"access_control.rules[].networks",
				"access_control.rules[].policy",
				"access_control.rules[].resources",
				"access_control.rules[].subject",
				"authentication_backend.ldap.additional_groups_dn",
				"authentication_backend.ldap.additional_users_dn",
				"authentication_backend.ldap.address",
				"authentication_backend.ldap.attributes.group_name",
				"authentication_backend.ldap.attributes.mail",
				"authentication_backend.ldap.attributes.username",
				"authentication_backend.ldap.base_dn",
				"authentication_backend.ldap.groups_filter",
				"authentication_backend.ldap.tls.private_key",
				"authentication_backend.ldap.user",
				"authentication_backend.ldap.users_filter",
				"authentication_backend.refresh_interval",
				"definitions.network.lan",
				"definitions.user_attributes.example.expression",
				"duo_api.hostname",
				"duo_api.integration_key",
				"log.level",
				"notifier.smtp.address",
				"notifier.smtp.disable_require_tls",
				"notifier.smtp.sender",
				"notifier.smtp.username",
				"regulation.ban_time",
				"regulation.find_time",
				"regulation.max_retries",
				"server.address",
				"server.endpoints.authz.auth-request.authn_strategies",
				"server.endpoints.authz.auth-request.authn_strategies[].name",
				"server.endpoints.authz.auth-request.implementation",
				"server.endpoints.authz.ext-authz.authn_strategies",
				"server.endpoints.authz.ext-authz.authn_strategies[].name",
				"server.endpoints.authz.ext-authz.implementation",
				"server.endpoints.authz.forward-auth.authn_strategies",
				"server.endpoints.authz.forward-auth.authn_strategies[].name",
				"server.endpoints.authz.forward-auth.implementation",
				"server.endpoints.authz.legacy.implementation",
				"server.endpoints.rate_limits.reset_password_finish.enable",
				"server.endpoints.rate_limits.reset_password_start.enable",
				"server.endpoints.rate_limits.second_factor_duo.enable",
				"server.endpoints.rate_limits.second_factor_totp.enable",
				"server.endpoints.rate_limits.session_elevation_finish.enable",
				"server.endpoints.rate_limits.session_elevation_start.enable",
				"session.cookies",
				"session.cookies[].authelia_url",
				"session.cookies[].default_redirection_url",
				"session.cookies[].domain",
				"session.expiration",
				"session.inactivity",
				"session.name",
				"session.redis.high_availability.sentinel_name",
				"session.redis.host",
				"session.redis.port",
				"storage.mysql.address",
				"storage.mysql.database",
				"storage.mysql.username",
				"totp.issuer",
				"webauthn.metadata.cache_policy",
				"webauthn.selection_criteria.discoverability",
				"webauthn.selection_criteria.user_verification",
			},
			warnings: nil,
			errors: []string{
				"duo_api: option 'secret_key' is required when duo is enabled but it's absent",
				"identity_validation: reset_password: option 'jwt_secret' is required when the reset password functionality isn't disabled",
				"authentication_backend: ldap: option 'password' is required",
				"session: option 'secret' is required when using the 'redis' provider",
				"storage: option 'encryption_key' is required",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("Individual", func(t *testing.T) {
				keys, config, val, err := NewConfiguration(tc.paths, tc.filters)

				assert.Equal(t, tc.keys, keys)

				if tc.err != "" {
					assert.EqualError(t, err, tc.err)
					assert.Nil(t, config)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, config)
					ValidateConfigurationKeys(keys, val)
					ValidateConfiguration(config, val)
				}

				require.Len(t, val.Warnings(), len(tc.warnings))
				require.Len(t, val.Errors(), len(tc.errors))

				for i, err := range val.Warnings() {
					assert.EqualError(t, err, tc.warnings[i])
				}

				for i, err := range val.Errors() {
					assert.EqualError(t, err, tc.errors[i])
				}
			})
			t.Run("Combined", func(t *testing.T) {
				keys, config, val, err := NewConfiguration(tc.paths, tc.filters)

				assert.Equal(t, tc.keys, keys)

				if tc.err != "" {
					assert.EqualError(t, err, tc.err)
					assert.Nil(t, config)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, config)
					ValidateConfigurationAndKeys(config, keys, val)
				}

				require.Len(t, val.Warnings(), len(tc.warnings))
				require.Len(t, val.Errors(), len(tc.errors))

				for i, err := range val.Warnings() {
					assert.EqualError(t, err, tc.warnings[i])
				}

				for i, err := range val.Errors() {
					assert.EqualError(t, err, tc.errors[i])
				}
			})
		})
	}
}

func TestNewNamedConfigFileFilters(t *testing.T) {
	filters, err := NewNamedConfigFileFilters("abc")
	assert.Nil(t, filters)
	assert.EqualError(t, err, "error occurred loading filters: invalid filter named 'abc'")

	filters, err = NewNamedConfigFileFilters("template")
	assert.NotNil(t, filters)
	assert.NoError(t, err)
}
