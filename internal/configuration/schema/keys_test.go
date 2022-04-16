package schema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var ValidKeys = []string{
	// Root Keys.
	"certificates_directory",
	"theme",
	"default_redirection_url",
	"jwt_secret",

	// Log keys.
	"log.level",
	"log.format",
	"log.file_path",
	"log.keep_stdout",

	// Server Keys.
	"server.host",
	"server.port",
	"server.read_buffer_size",
	"server.write_buffer_size",
	"server.path",
	"server.asset_path",
	"server.enable_pprof",
	"server.enable_expvars",
	"server.disable_healthcheck",
	"server.tls.key",
	"server.tls.certificate",
	"server.tls.client_certificates",
	"server.headers.csp_template",

	// TOTP Keys.
	"totp.disable",
	"totp.issuer",
	"totp.algorithm",
	"totp.digits",
	"totp.period",
	"totp.skew",
	"totp.secret_size",

	// Webauthn Keys.
	"webauthn.disable",
	"webauthn.display_name",
	"webauthn.attestation_conveyance_preference",
	"webauthn.user_verification",
	"webauthn.timeout",

	// DUO API Keys.
	"duo_api.disable",
	"duo_api.hostname",
	"duo_api.enable_self_enrollment",
	"duo_api.secret_key",
	"duo_api.integration_key",

	// Access Control Keys.
	"access_control.default_policy",
	"access_control.networks",
	"access_control.networks[].name",
	"access_control.networks[].networks",
	"access_control.rules",
	"access_control.rules[].domain",
	"access_control.rules[].domain_regex",
	"access_control.rules[].methods",
	"access_control.rules[].networks",
	"access_control.rules[].subject",
	"access_control.rules[].policy",
	"access_control.rules[].resources",

	// Session Keys.
	"session.name",
	"session.domain",
	"session.secret",
	"session.same_site",
	"session.expiration",
	"session.inactivity",
	"session.remember_me_duration",

	// Redis Session Keys.
	"session.redis.host",
	"session.redis.port",
	"session.redis.username",
	"session.redis.password",
	"session.redis.database_index",
	"session.redis.maximum_active_connections",
	"session.redis.minimum_idle_connections",
	"session.redis.tls.minimum_version",
	"session.redis.tls.skip_verify",
	"session.redis.tls.server_name",
	"session.redis.high_availability.sentinel_name",
	"session.redis.high_availability.sentinel_username",
	"session.redis.high_availability.sentinel_password",
	"session.redis.high_availability.nodes",
	"session.redis.high_availability.nodes[].host",
	"session.redis.high_availability.nodes[].port",
	"session.redis.high_availability.route_by_latency",
	"session.redis.high_availability.route_randomly",

	// Storage Keys.
	"storage.encryption_key",

	// Local Storage Keys.
	"storage.local.path",

	// MySQL Storage Keys.
	"storage.mysql.host",
	"storage.mysql.port",
	"storage.mysql.database",
	"storage.mysql.username",
	"storage.mysql.password",
	"storage.mysql.timeout",

	// PostgreSQL Storage Keys.
	"storage.postgres.host",
	"storage.postgres.port",
	"storage.postgres.database",
	"storage.postgres.username",
	"storage.postgres.password",
	"storage.postgres.timeout",
	"storage.postgres.schema",
	"storage.postgres.ssl.mode",
	"storage.postgres.ssl.root_certificate",
	"storage.postgres.ssl.certificate",
	"storage.postgres.ssl.key",

	"storage.postgres.sslmode", // Deprecated. TODO: Remove in v4.36.0.

	// FileSystem Notifier Keys.
	"notifier.filesystem.filename",
	"notifier.disable_startup_check",

	// SMTP Notifier Keys.
	"notifier.smtp.host",
	"notifier.smtp.port",
	"notifier.smtp.timeout",
	"notifier.smtp.username",
	"notifier.smtp.password",
	"notifier.smtp.identifier",
	"notifier.smtp.sender",
	"notifier.smtp.subject",
	"notifier.smtp.startup_check_address",
	"notifier.smtp.disable_require_tls",
	"notifier.smtp.disable_html_emails",
	"notifier.smtp.tls.minimum_version",
	"notifier.smtp.tls.skip_verify",
	"notifier.smtp.tls.server_name",
	"notifier.template_path",

	// Regulation Keys.
	"regulation.max_retries",
	"regulation.find_time",
	"regulation.ban_time",

	// Authentication Backend Keys.
	"authentication_backend.disable_reset_password",
	"authentication_backend.password_reset.custom_url",
	"authentication_backend.refresh_interval",

	// LDAP Authentication Backend Keys.
	"authentication_backend.ldap.implementation",
	"authentication_backend.ldap.url",
	"authentication_backend.ldap.timeout",
	"authentication_backend.ldap.base_dn",
	"authentication_backend.ldap.username_attribute",
	"authentication_backend.ldap.additional_users_dn",
	"authentication_backend.ldap.users_filter",
	"authentication_backend.ldap.additional_groups_dn",
	"authentication_backend.ldap.groups_filter",
	"authentication_backend.ldap.group_name_attribute",
	"authentication_backend.ldap.mail_attribute",
	"authentication_backend.ldap.display_name_attribute",
	"authentication_backend.ldap.user",
	"authentication_backend.ldap.password",
	"authentication_backend.ldap.start_tls",
	"authentication_backend.ldap.tls.minimum_version",
	"authentication_backend.ldap.tls.skip_verify",
	"authentication_backend.ldap.tls.server_name",

	// File Authentication Backend Keys.
	"authentication_backend.file.path",
	"authentication_backend.file.password.algorithm",
	"authentication_backend.file.password.iterations",
	"authentication_backend.file.password.key_length",
	"authentication_backend.file.password.salt_length",
	"authentication_backend.file.password.memory",
	"authentication_backend.file.password.parallelism",

	// Identity Provider Keys.
	"identity_providers.oidc.hmac_secret",
	"identity_providers.oidc.issuer_private_key",
	"identity_providers.oidc.id_token_lifespan",
	"identity_providers.oidc.access_token_lifespan",
	"identity_providers.oidc.refresh_token_lifespan",
	"identity_providers.oidc.authorize_code_lifespan",
	"identity_providers.oidc.enforce_pkce",
	"identity_providers.oidc.enable_pkce_plain_challenge",
	"identity_providers.oidc.enable_client_debug_messages",
	"identity_providers.oidc.minimum_parameter_entropy",
	"identity_providers.oidc.cors.endpoints",
	"identity_providers.oidc.cors.allowed_origins",
	"identity_providers.oidc.cors.allowed_origins_from_client_redirect_uris",
	"identity_providers.oidc.clients",
	"identity_providers.oidc.clients[].id",
	"identity_providers.oidc.clients[].description",
	"identity_providers.oidc.clients[].secret",
	"identity_providers.oidc.clients[].sector_identifier",
	"identity_providers.oidc.clients[].public",
	"identity_providers.oidc.clients[].redirect_uris",
	"identity_providers.oidc.clients[].authorization_policy",
	"identity_providers.oidc.clients[].pre_configured_consent_duration",
	"identity_providers.oidc.clients[].scopes",
	"identity_providers.oidc.clients[].audience",
	"identity_providers.oidc.clients[].grant_types",
	"identity_providers.oidc.clients[].response_types",
	"identity_providers.oidc.clients[].response_modes",
	"identity_providers.oidc.clients[].userinfo_signing_algorithm",

	// NTP keys.
	"ntp.address",
	"ntp.version",
	"ntp.max_desync",
	"ntp.disable_startup_check",
	"ntp.disable_failure",

	// Password Policy keys.
	"password_policy.standard.enabled",
	"password_policy.standard.min_length",
	"password_policy.standard.max_length",
	"password_policy.standard.require_uppercase",
	"password_policy.standard.require_lowercase",
	"password_policy.standard.require_number",
	"password_policy.standard.require_special",
	"password_policy.zxcvbn.enabled",
	"password_policy.zxcvbn.min_score",
}

func TestOldKeys(t *testing.T) {
	for _, key := range ValidKeys {
		assert.Contains(t, Keys, key)
	}

	for _, key := range Keys {
		assert.Contains(t, ValidKeys, key)
	}
}

func TestDuplicates(t *testing.T) {
	assert.Equal(t, len(Keys), len(ValidKeys))
}
